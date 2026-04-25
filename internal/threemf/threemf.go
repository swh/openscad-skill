// Package threemf reads and writes the subset of 3MF / Bambu Studio project
// files that openscad-pack-3mf and openscad-build-template care about: zip
// containers around `Metadata/project_settings.config` (JSON), `Metadata/
// model_settings.config` (XML), `3D/3dmodel.model` (XML), `3D/Objects/
// object_1.model` (XML mesh), thumbnails, and assorted boilerplate.
//
// The package keeps every entry of the input archive in memory so callers
// can mutate, drop, or replace specific files and re-zip in one shot.
package threemf

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ThreeMF holds the unzipped contents of a 3MF archive. Files maps the
// in-zip path (e.g. "Metadata/project_settings.config") to its raw bytes.
type ThreeMF struct {
	Files map[string][]byte
	// Names preserves the original entry order so re-zipped archives stay
	// stable across calls.
	Names []string
}

// Open reads a 3MF (zip) file into memory.
func Open(path string) (*ThreeMF, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer zr.Close()

	t := &ThreeMF{Files: map[string][]byte{}}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f.Name, err)
		}
		t.Files[f.Name] = data
		t.Names = append(t.Names, f.Name)
	}
	return t, nil
}

// WriteTo writes the (possibly mutated) archive to disk. Entries are written
// sorted lexicographically so the output is stable across runs.
func (t *ThreeMF) WriteTo(out string) error {
	tmp := out + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)

	names := make([]string, 0, len(t.Files))
	for n := range t.Files {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			zw.Close()
			f.Close()
			os.Remove(tmp)
			return err
		}
		if _, err := w.Write(t.Files[name]); err != nil {
			zw.Close()
			f.Close()
			os.Remove(tmp)
			return err
		}
	}
	if err := zw.Close(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, out)
}

// ProjectSettings parses Metadata/project_settings.config. Returns a generic
// JSON map (Bambu's schema is heterogeneous: scalar strings + per-slot lists).
func (t *ThreeMF) ProjectSettings() (map[string]any, error) {
	raw, ok := t.Files["Metadata/project_settings.config"]
	if !ok {
		return nil, fmt.Errorf("Metadata/project_settings.config missing")
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse project_settings.config: %w", err)
	}
	return cfg, nil
}

// WriteProjectSettings re-encodes the map and stores it back in the archive.
// The encoding matches Python's `json.dumps(indent=4)` output (sorted keys,
// 4-space indent) for parity with the previous build-template implementation.
func (t *ThreeMF) WriteProjectSettings(cfg map[string]any) error {
	data, err := marshalIndent4Sorted(cfg)
	if err != nil {
		return err
	}
	// Trailing newline matches `json.dumps(...) + "\n"`.
	data = append(data, '\n')
	t.Files["Metadata/project_settings.config"] = data
	return nil
}

// DropThumbnails removes every PNG under Metadata/. Bambu Studio re-renders
// thumbnails on open, so leaving stale ones is worse than dropping them.
func (t *ThreeMF) DropThumbnails() {
	for n := range t.Files {
		if strings.HasPrefix(n, "Metadata/") && strings.HasSuffix(strings.ToLower(n), ".png") {
			delete(t.Files, n)
		}
	}
}

// reTransform matches `transform="..."` inside a tag. We use regex (not a
// proper XML parser) because we only mutate one attribute per file and
// preserving exact whitespace / namespace declarations is easier this way —
// matches the Python implementation.
var (
	reItemTransform     = regexp.MustCompile(`(<item[^>]*?\btransform=")[^"]*(")`)
	reAssembleTransform = regexp.MustCompile(`(<assemble_item[^>]*?\btransform=")[^"]*(")`)
	reSourceOffsetZ     = regexp.MustCompile(`(<metadata key="source_offset_z" value=")[^"]*(")`)
	reExtruderMeta      = regexp.MustCompile(`(<metadata key="extruder" value=")[^"]*(")`)
	reFaceCount         = regexp.MustCompile(`face_count="\d+"`)
	reVerticesBlock     = regexp.MustCompile(`(?s)<vertices\b[^>]*>.+?</vertices>`)
	reTrianglesBlock    = regexp.MustCompile(`(?s)<triangles\b[^>]*>.+?</triangles>`)
	reVertexCoord       = regexp.MustCompile(`x="([^"]+)"\s+y="([^"]+)"\s+z="([^"]+)"`)
	reTriangleTag       = regexp.MustCompile(`<triangle\b`)
)

// RewriteItemTransform replaces the placement transforms in 3D/3dmodel.model
// (build > item) and Metadata/model_settings.config (assemble_item) with the
// caller's tx/ty/tz translation, identity rotation. Source_offset_z is reset
// to 0 because the new transform already accounts for placement.
func (t *ThreeMF) RewriteItemTransform(tx, ty, tz float64) {
	formatted := fmt.Sprintf("1 0 0 0 1 0 0 0 1 %s %s %s",
		fmtFloat(tx), fmtFloat(ty), fmtFloat(tz))

	if data, ok := t.Files["3D/3dmodel.model"]; ok {
		t.Files["3D/3dmodel.model"] = reItemTransform.ReplaceAll(data,
			[]byte("${1}"+formatted+"${2}"))
	}
	if data, ok := t.Files["Metadata/model_settings.config"]; ok {
		data = reAssembleTransform.ReplaceAll(data,
			[]byte("${1}"+formatted+"${2}"))
		data = reSourceOffsetZ.ReplaceAll(data, []byte("${1}0${2}"))
		t.Files["Metadata/model_settings.config"] = data
	}
}

// ZeroPlacementTransforms is for build-template: collapse all transforms to
// identity so the next splice can compute its own placement from scratch.
func (t *ThreeMF) ZeroPlacementTransforms() {
	identity := []byte("${1}1 0 0 0 1 0 0 0 1 0 0 0${2}")

	if data, ok := t.Files["3D/3dmodel.model"]; ok {
		t.Files["3D/3dmodel.model"] = reItemTransform.ReplaceAll(data, identity)
	}
	if data, ok := t.Files["Metadata/model_settings.config"]; ok {
		data = reAssembleTransform.ReplaceAll(data, identity)
		data = reSourceOffsetZ.ReplaceAll(data, []byte("${1}0${2}"))
		t.Files["Metadata/model_settings.config"] = data
	}
}

// SetEveryObjectExtruder rewrites every <metadata key="extruder" value="N"/>
// in model_settings.config to the given (1-indexed) slot value. Used by
// pack-3mf when applying a filament so all objects point at slot 0.
func (t *ThreeMF) SetEveryObjectExtruder(slot int) {
	data, ok := t.Files["Metadata/model_settings.config"]
	if !ok {
		return
	}
	repl := []byte(fmt.Sprintf("${1}%d${2}", slot))
	t.Files["Metadata/model_settings.config"] = reExtruderMeta.ReplaceAll(data, repl)
}

// ReplaceMesh swaps the <vertices> and <triangles> contents of
// 3D/Objects/object_1.model with the supplied bodies. The face_count
// metadata in model_settings.config is also updated to match the new
// triangle count.
func (t *ThreeMF) ReplaceMesh(verticesXML, trianglesXML string) error {
	objKey := "3D/Objects/object_1.model"
	data, ok := t.Files[objKey]
	if !ok {
		return fmt.Errorf("template missing %s", objKey)
	}
	newVerts := []byte("<vertices>\n" + verticesXML + "\n   </vertices>")
	newTris := []byte("<triangles>\n" + trianglesXML + "\n   </triangles>")
	if !reVerticesBlock.Match(data) {
		return fmt.Errorf("%s missing <vertices> block", objKey)
	}
	if !reTrianglesBlock.Match(data) {
		return fmt.Errorf("%s missing <triangles> block", objKey)
	}
	data = reVerticesBlock.ReplaceAllLiteral(data, newVerts)
	data = reTrianglesBlock.ReplaceAllLiteral(data, newTris)
	t.Files[objKey] = data

	nTri := len(reTriangleTag.FindAll([]byte(trianglesXML), -1))
	if ms, ok := t.Files["Metadata/model_settings.config"]; ok {
		ms = reFaceCount.ReplaceAllLiteral(ms, []byte(fmt.Sprintf(`face_count="%d"`, nTri)))
		t.Files["Metadata/model_settings.config"] = ms
	}
	return nil
}

// ExtractMeshXML pulls the <vertices> and <triangles> *contents* out of an
// OpenSCAD-produced 3MF (whose mesh lives directly in 3D/3dmodel.model).
// Returns inner text bodies suitable for ReplaceMesh.
func ExtractMeshXML(in *ThreeMF) (verts, tris string, err error) {
	data, ok := in.Files["3D/3dmodel.model"]
	if !ok {
		return "", "", fmt.Errorf("input 3MF missing 3D/3dmodel.model")
	}
	vMatch := reVerticesBlock.Find(data)
	tMatch := reTrianglesBlock.Find(data)
	if vMatch == nil || tMatch == nil {
		return "", "", fmt.Errorf("input 3MF mesh has no <vertices>/<triangles>")
	}
	return innerOfBlock(vMatch, "vertices"), innerOfBlock(tMatch, "triangles"), nil
}

// PlateSize returns the printable plate dimensions (width, depth) in mm,
// derived from project_settings.config's printable_area corner list. Falls
// back to 256x256 when missing or malformed.
func (t *ThreeMF) PlateSize() (w, d float64) {
	cfg, err := t.ProjectSettings()
	if err != nil {
		return 256, 256
	}
	raw, ok := cfg["printable_area"].([]any)
	if !ok || len(raw) == 0 {
		return 256, 256
	}
	var minX, minY, maxX, maxY float64
	first := true
	for _, v := range raw {
		s, ok := v.(string)
		if !ok {
			continue
		}
		parts := strings.Split(s, "x")
		if len(parts) != 2 {
			continue
		}
		x, err1 := strconv.ParseFloat(parts[0], 64)
		y, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		if first {
			minX, minY, maxX, maxY = x, y, x, y
			first = false
			continue
		}
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	if first {
		return 256, 256
	}
	return maxX - minX, maxY - minY
}

// BBox returns the axis-aligned bounding box of vertices in the 3D/3dmodel.model
// file (typical OpenSCAD output). Useful only for plain meshes, not project
// archives whose mesh is in 3D/Objects/object_1.model.
func (t *ThreeMF) BBox() (min, max [3]float64, err error) {
	data, ok := t.Files["3D/3dmodel.model"]
	if !ok {
		return min, max, fmt.Errorf("3D/3dmodel.model missing")
	}
	v := reVerticesBlock.Find(data)
	if v == nil {
		return min, max, fmt.Errorf("no <vertices> in mesh")
	}
	matches := reVertexCoord.FindAllSubmatch(v, -1)
	if len(matches) == 0 {
		return min, max, fmt.Errorf("no vertex coordinates parsed")
	}
	first := true
	for _, m := range matches {
		x, _ := strconv.ParseFloat(string(m[1]), 64)
		y, _ := strconv.ParseFloat(string(m[2]), 64)
		z, _ := strconv.ParseFloat(string(m[3]), 64)
		if first {
			min[0], min[1], min[2] = x, y, z
			max[0], max[1], max[2] = x, y, z
			first = false
			continue
		}
		if x < min[0] {
			min[0] = x
		}
		if y < min[1] {
			min[1] = y
		}
		if z < min[2] {
			min[2] = z
		}
		if x > max[0] {
			max[0] = x
		}
		if y > max[1] {
			max[1] = y
		}
		if z > max[2] {
			max[2] = z
		}
	}
	return min, max, nil
}

// CountTriangles returns the number of <triangle ...> entries in the given
// mesh XML body (used for status output).
func CountTriangles(trianglesXML string) int {
	return len(reTriangleTag.FindAll([]byte(trianglesXML), -1))
}

func innerOfBlock(block []byte, tag string) string {
	open := []byte("<" + tag)
	close := []byte("</" + tag + ">")
	startIdx := bytes.Index(block, open)
	if startIdx < 0 {
		return ""
	}
	gtIdx := bytes.IndexByte(block[startIdx:], '>')
	if gtIdx < 0 {
		return ""
	}
	startIdx += gtIdx + 1
	endIdx := bytes.Index(block, close)
	if endIdx < 0 || endIdx <= startIdx {
		return ""
	}
	return strings.TrimSpace(string(block[startIdx:endIdx]))
}

// fmtFloat formats a number the way the Python tool did: %g rounded to 4 dp,
// with negative zero normalised to 0. Avoids decimal noise like "128.000000".
func fmtFloat(v float64) string {
	v = v + 0.0 // turns -0.0 into 0.0
	// Round to 4 dp.
	r := float64(int64(v*10000+sgn(v)*0.5)) / 10000
	s := strconv.FormatFloat(r, 'g', -1, 64)
	if s == "-0" {
		return "0"
	}
	return s
}

func sgn(x float64) float64 {
	if x < 0 {
		return -1
	}
	return 1
}

// marshalIndent4Sorted produces JSON output equivalent to Python's
// `json.dumps(d, indent=4)` — keys sorted alphabetically, 4-space indent,
// `": "` after keys, `,` after elements. Necessary for byte-identical
// build-template parity with the Python tool. HTML-character escaping
// (Go's default for `<`, `>`, `&`) is disabled to match Python.
func marshalIndent4Sorted(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeSorted(&buf, v, 0); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// encodeStringNoHTMLEscape writes a JSON-encoded string without Go's default
// `<`/`>`/`&` → `\uXXXX` substitution, matching Python's json.dumps.
func encodeStringNoHTMLEscape(buf *bytes.Buffer, s string) error {
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return err
	}
	// json.Encoder appends a newline; trim it.
	out := buf.Bytes()
	if n := len(out); n > 0 && out[n-1] == '\n' {
		buf.Truncate(n - 1)
	}
	return nil
}

func encodeSorted(buf *bytes.Buffer, v any, depth int) error {
	switch val := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case json.Number:
		buf.WriteString(string(val))
	case float64:
		buf.WriteString(strconv.FormatFloat(val, 'g', -1, 64))
	case string:
		if err := encodeStringNoHTMLEscape(buf, val); err != nil {
			return err
		}
	case []any:
		if len(val) == 0 {
			buf.WriteString("[]")
			return nil
		}
		buf.WriteString("[\n")
		indent := strings.Repeat("    ", depth+1)
		closeIndent := strings.Repeat("    ", depth)
		for i, item := range val {
			buf.WriteString(indent)
			if err := encodeSorted(buf, item, depth+1); err != nil {
				return err
			}
			if i < len(val)-1 {
				buf.WriteString(",")
			}
			buf.WriteString("\n")
		}
		buf.WriteString(closeIndent + "]")
	case map[string]any:
		if len(val) == 0 {
			buf.WriteString("{}")
			return nil
		}
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteString("{\n")
		indent := strings.Repeat("    ", depth+1)
		closeIndent := strings.Repeat("    ", depth)
		for i, k := range keys {
			buf.WriteString(indent)
			if err := encodeStringNoHTMLEscape(buf, k); err != nil {
				return err
			}
			buf.WriteString(": ")
			if err := encodeSorted(buf, val[k], depth+1); err != nil {
				return err
			}
			if i < len(keys)-1 {
				buf.WriteString(",")
			}
			buf.WriteString("\n")
		}
		buf.WriteString(closeIndent + "}")
	default:
		return fmt.Errorf("unsupported JSON type %T", v)
	}
	return nil
}

