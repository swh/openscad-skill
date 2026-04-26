// Package pack3mf is the entry-point logic for the openscad-pack-3mf binary.
// The binary is a tiny shim that calls Main(); the shim's main.go is
// generated into cmd/ at build time by the Makefile and removed afterwards,
// so cmd/ doesn't need to be tracked in git.
//
// AMS-aware filament selection: by default the tool syncs the printer's
// current AMS contents into every slot of the output project (via cached
// state in BambuStudio.conf, or live MQTT if creds are set), so the saved
// template's filaments don't end up stale. --ams-slot N and --filament-type
// TYPE additionally pick a specific AMS slot's filament for slot 0 (the
// active-object filament). --filament NAME skips the sync entirely on the
// assumption the user wants a specific profile that may not be in the AMS.
//
// Env vars (override CLAUDE.md):
//   BAMBU_ACCESS_CODE   8-digit LAN access code (recommended — keeps the
//                       secret out of any prompt context that loads CLAUDE.md)
//   BAMBU_SERIAL        printer serial number
//   BAMBU_HOST          printer LAN IP or hostname
package pack3mf

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/swh/openscad-skill/internal/bambu"
	"github.com/swh/openscad-skill/internal/claudemd"
	"github.com/swh/openscad-skill/internal/threemf"
)

const (
	defaultSlicer = "bambu-studio"
	mqttTimeout   = 3 * time.Second
)

type repeatString []string

func (r *repeatString) String() string     { return strings.Join(*r, ",") }
func (r *repeatString) Set(s string) error { *r = append(*r, s); return nil }

func defaultTemplatesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "openscad-bosl2", "templates")
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "openscad-pack-3mf: "+format+"\n", args...)
	os.Exit(1)
}

func usageError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "openscad-pack-3mf: "+format+"\n", args...)
	os.Exit(2)
}

func usage(code int) {
	fmt.Fprint(os.Stderr,
		`usage: openscad-pack-3mf [options] INPUT.{scad,3mf}

Splice OpenSCAD geometry into a slicer template 3MF.

Options:
  --slicer NAME           Resolves to ~/.config/openscad-bosl2/templates/<NAME>.3mf
                          (default: bambu-studio).
  -t, --template P        Explicit template path (overrides --slicer's lookup).
  -f, --filament NAME     Apply a Bambu Studio system filament profile by name
                          (e.g. "Bambu PETG HF @BBL X1C"). Skips AMS sync.
      --ams-slot N        Apply the filament currently in AMS slot N (0-indexed)
                          to the active object. The full AMS state is also
                          synced into every slot of the project.
      --filament-type T   Apply the first AMS-loaded filament whose type
                          matches T (e.g. PLA, PETG, PETG-CF). Same source
                          resolution as --ams-slot.
  -s, --set KEY=VALUE     Override a slicer setting on top of the template
                          (repeatable, e.g. --set sparse_infill_density=50%).
  -o, --output P          Output path (default: INPUT-packed.3mf).
  -h, --help              Show this help.

At most one of --filament / --ams-slot / --filament-type may be passed.

Env: BAMBU_ACCESS_CODE / BAMBU_SERIAL / BAMBU_HOST override CLAUDE.md creds.
`)
	os.Exit(code)
}

// Main is the entry point invoked by cmd/openscad-pack-3mf/main.go.
func Main() {
	fs := flag.NewFlagSet("openscad-pack-3mf", flag.ExitOnError)
	fs.Usage = func() { usage(2) }

	slicer := fs.String("slicer", defaultSlicer, "")
	template := fs.String("template", "", "")
	fs.StringVar(template, "t", "", "")
	filament := fs.String("filament", "", "")
	fs.StringVar(filament, "f", "", "")
	amsSlotS := fs.String("ams-slot", "", "")
	filamentType := fs.String("filament-type", "", "")
	output := fs.String("output", "", "")
	fs.StringVar(output, "o", "", "")
	help := fs.Bool("help", false, "")
	fs.BoolVar(help, "h", false, "")

	var setOverrides repeatString
	fs.Var(&setOverrides, "set", "")
	fs.Var(&setOverrides, "s", "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if *help {
		usage(0)
	}
	if fs.NArg() != 1 {
		usageError("exactly one input is required")
	}
	in := fs.Arg(0)

	amsSlot := -1
	if *amsSlotS != "" {
		n, err := strconv.Atoi(*amsSlotS)
		if err != nil {
			usageError("--ams-slot must be an integer, got %q", *amsSlotS)
		}
		amsSlot = n
	}

	chosen := 0
	if *filament != "" {
		chosen++
	}
	if amsSlot >= 0 {
		chosen++
	}
	if *filamentType != "" {
		chosen++
	}
	if chosen > 1 {
		usageError("pass at most one of --filament / --ams-slot / --filament-type")
	}

	overrides := map[string]string{}
	for _, s := range setOverrides {
		k, v, ok := strings.Cut(s, "=")
		if !ok || strings.TrimSpace(k) == "" {
			usageError("--set value must be 'KEY=VALUE', got %q", s)
		}
		overrides[strings.TrimSpace(k)] = v
	}

	tmplPath := *template
	if tmplPath == "" {
		tmplPath = filepath.Join(defaultTemplatesDir(), *slicer+".3mf")
	}
	if st, err := os.Stat(tmplPath); err != nil || st.IsDir() {
		die("template not found: %s\n  Save a slicer project as the template, or pass --template PATH.", tmplPath)
	}

	out := *output
	if out == "" {
		base := strings.TrimSuffix(in, filepath.Ext(in))
		out = base + "-packed.3mf"
	}

	if _, err := os.Stat(in); err != nil {
		die("input not found: %s", in)
	}

	tmpDir, err := os.MkdirTemp("", "openscad-pack-3mf-*")
	if err != nil {
		die("mkdir tmp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	rawIn := in
	if strings.EqualFold(filepath.Ext(in), ".scad") {
		rawIn = filepath.Join(tmpDir, "raw.3mf")
		if err := scadToThreeMF(in, rawIn); err != nil {
			die("%v", err)
		}
	} else if !strings.EqualFold(filepath.Ext(in), ".3mf") {
		die("unsupported input extension: %s (need .scad or .3mf)", filepath.Ext(in))
	}

	srcMesh, err := threemf.Open(rawIn)
	if err != nil {
		die("%v", err)
	}
	verticesXML, trianglesXML, err := threemf.ExtractMeshXML(srcMesh)
	if err != nil {
		die("%v", err)
	}
	mn, mx, err := srcMesh.BBox()
	if err != nil {
		die("%v", err)
	}

	tmpl, err := threemf.Open(tmplPath)
	if err != nil {
		die("%v", err)
	}

	var (
		amsSource        bambu.AMSSource
		amsSynced        bool
		resolvedFilament = *filament
		filamentSource   = "explicit"
	)

	var (
		idx       *bambu.FilamentIndex
		studioDir string
		printerID string
	)
	if *slicer == defaultSlicer {
		var err error
		studioDir, err = bambu.FindStudioDir()
		if err == nil {
			filDir, err := bambu.FindFilamentDir(studioDir)
			if err == nil {
				idx, err = bambu.LoadFilamentIndex(filDir)
				if err != nil {
					die("load filament index: %v", err)
				}
			}
		}
		tmplCfg, err := tmpl.ProjectSettings()
		if err != nil {
			die("%v", err)
		}
		printerID, _ = tmplCfg["printer_settings_id"].(string)
	}

	// Sync AMS state into all template slots (default behaviour). Skipped
	// for --filament NAME (the user wants a specific profile, not whatever
	// is loaded), and silently skipped if no AMS state is available at all.
	if *slicer == defaultSlicer && idx != nil && *filament == "" {
		state, src, err := getAMSState(studioDir)
		if err != nil {
			die("%v", err)
		}
		if len(state) > 0 {
			if err := bambu.ApplyAMSState(tmpl, idx, state, printerID); err != nil {
				die("sync AMS state: %v", err)
			}
			amsSource = src
			amsSynced = true

			// --ams-slot / --filament-type pick a specific slot's filament
			// for slot 0 too (so the active object uses it).
			if amsSlot >= 0 {
				resolvedFilament, err = bambu.PickFilamentNameBySlot(state, printerID, idx, amsSlot)
				filamentSource = fmt.Sprintf("AMS slot %d", amsSlot)
				if err != nil {
					die("%v", err)
				}
			} else if *filamentType != "" {
				resolvedFilament, err = bambu.PickFilamentNameByType(state, printerID, idx, *filamentType)
				filamentSource = fmt.Sprintf("AMS type %s", *filamentType)
				if err != nil {
					die("%v", err)
				}
			}
		} else if amsSlot >= 0 || *filamentType != "" {
			die("--ams-slot / --filament-type need AMS state but none is available (no MQTT creds and BambuStudio.conf has no ams_filament_ids)")
		}
	}

	// Apply specific filament to slot 0 (after AMS sync, so this overrides
	// slot 0 with the user's chosen filament).
	if resolvedFilament != "" && idx != nil {
		if err := bambu.ApplyFilament(tmpl, idx, resolvedFilament); err != nil {
			die("apply filament: %v", err)
		}
	}

	if len(overrides) > 0 {
		if *slicer != defaultSlicer {
			die("--set not yet implemented for slicer: %s", *slicer)
		}
		if err := bambu.ApplyOverrides(tmpl, overrides); err != nil {
			die("apply overrides: %v", err)
		}
	}

	plateW, plateD := tmpl.PlateSize()
	tx := plateW/2 - (mn[0]+mx[0])/2
	ty := plateD/2 - (mn[1]+mx[1])/2
	tz := -mn[2]

	tmpl.RewriteItemTransform(tx, ty, tz)
	if err := tmpl.ReplaceMesh(verticesXML, trianglesXML); err != nil {
		die("%v", err)
	}
	tmpl.DropThumbnails()
	tmpl.SetSourceTitle(filepath.Base(in))
	tmpl.SetModificationDate("")

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		die("mkdir %s: %v", filepath.Dir(out), err)
	}
	if err := tmpl.WriteTo(out); err != nil {
		die("write %s: %v", out, err)
	}

	nTri := threemf.CountTriangles(trianglesXML)
	fmt.Printf("%s -> %s\n", in, out)
	fmt.Printf("  template: %s\n", tmplPath)
	fmt.Printf("  bbox: (%.1f × %.1f × %.1f) mm, %d triangles\n",
		mx[0]-mn[0], mx[1]-mn[1], mx[2]-mn[2], nTri)
	fmt.Printf("  centred at (%.1f, %.1f) on %.0f×%.0f plate\n",
		tx, ty, plateW, plateD)
	if resolvedFilament != "" {
		fmt.Printf("  filament: %s (%s)\n", resolvedFilament, filamentSource)
	}
	if amsSynced {
		kind := "cached"
		if amsSource.Live {
			kind = "live"
		}
		fmt.Printf("  AMS state: %s (%s) — synced into all slots\n", kind, amsSource.Note)
	}
	if len(overrides) > 0 {
		var pairs []string
		for k, v := range overrides {
			pairs = append(pairs, k+"="+v)
		}
		fmt.Printf("  overrides: %s\n", strings.Join(pairs, ", "))
	}
}

// getAMSState attempts MQTT (if creds are available via env vars or
// CLAUDE.md), falling back to the BambuStudio.conf cache. Returns
// (nil, zero-source, nil) when no source is available — the caller decides
// whether that's fatal.
func getAMSState(studioDir string) (bambu.AMSState, bambu.AMSSource, error) {
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	creds, _ := claudemd.Resolve(claudemd.SearchPaths(cwd, home))
	if creds == nil {
		creds = &claudemd.Creds{Sources: map[string]string{}}
	}
	if v := os.Getenv("BAMBU_SERIAL"); v != "" {
		creds.Serial = v
		creds.Sources["serial"] = "env BAMBU_SERIAL"
	}
	if v := os.Getenv("BAMBU_ACCESS_CODE"); v != "" {
		creds.AccessCode = v
		creds.Sources["access_code"] = "env BAMBU_ACCESS_CODE"
	}
	if v := os.Getenv("BAMBU_HOST"); v != "" {
		creds.Host = v
		creds.Sources["host"] = "env BAMBU_HOST"
	}

	if creds.HasMQTT() {
		state, err := bambu.ReadAMSLive(creds.Host, creds.Serial, creds.AccessCode, mqttTimeout)
		if err == nil && len(state) > 0 {
			return state, bambu.AMSSource{Live: true, Note: fmt.Sprintf("MQTT %s", creds.Host)}, nil
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "openscad-pack-3mf: MQTT live read failed (%v); falling back to BambuStudio.conf\n", err)
		}
	}

	state, err := bambu.ReadAMSCache(studioDir)
	if err != nil {
		return nil, bambu.AMSSource{}, err
	}
	if len(state) == 0 {
		return nil, bambu.AMSSource{}, nil
	}
	return state, bambu.AMSSource{Live: false, Note: "BambuStudio.conf"}, nil
}

func scadToThreeMF(scad, out string) error {
	bin := findOpenSCAD()
	if bin == "" {
		return fmt.Errorf("openscad binary not found; set $OPENSCAD or install OpenSCAD")
	}
	cmd := exec.Command(bin, "-o", out, scad)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("openscad failed: %v\n%s", err, output)
	}
	return nil
}

func findOpenSCAD() string {
	if env := os.Getenv("OPENSCAD"); env != "" {
		return env
	}
	if p, err := exec.LookPath("openscad"); err == nil {
		return p
	}
	for _, c := range []string{
		"/Applications/OpenSCAD.app/Contents/MacOS/OpenSCAD",
		"/usr/bin/openscad",
		"/usr/local/bin/openscad",
		"/snap/bin/openscad",
	} {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return ""
}
