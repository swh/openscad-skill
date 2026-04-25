package bambu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FilamentMetaSkip lists JSON keys that are profile metadata (not slot
// values) and shouldn't be splatted into project_settings.config when
// applying a filament.
var FilamentMetaSkip = map[string]struct{}{
	"type": {}, "name": {}, "inherits": {}, "from": {}, "setting_id": {},
	"instantiation": {}, "filament_id": {}, "include": {},
	"compatible_printers": {}, "compatible_printers_condition": {},
	"compatible_prints": {}, "compatible_prints_condition": {},
	"version": {}, "description": {}, "is_custom_defined": {},
}

// FilamentIndex holds every filament JSON in the Bambu Studio system dir,
// indexed by `name`, plus a precomputed map of effective filament_id ->
// instantiable profile names (so we can resolve AMS codes to picks).
type FilamentIndex struct {
	byName               map[string]map[string]any
	instantiableByID     map[string][]string
	effectiveCache       map[string]map[string]any
}

// LoadFilamentIndex reads every *.json in filamentDir, walks inheritance
// chains, and prepares the lookups we need.
func LoadFilamentIndex(filamentDir string) (*FilamentIndex, error) {
	idx := &FilamentIndex{
		byName:           map[string]map[string]any{},
		instantiableByID: map[string][]string{},
		effectiveCache:   map[string]map[string]any{},
	}
	entries, err := os.ReadDir(filamentDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		full := filepath.Join(filamentDir, e.Name())
		data, err := os.ReadFile(full)
		if err != nil {
			continue
		}
		var j map[string]any
		if err := json.Unmarshal(data, &j); err != nil {
			continue
		}
		if t, _ := j["type"].(string); t != "filament" {
			continue
		}
		name, _ := j["name"].(string)
		if name == "" {
			continue
		}
		idx.byName[name] = j
	}

	// Build instantiableByID after all profiles are loaded so inheritance
	// can resolve.
	for name, j := range idx.byName {
		if v, _ := j["instantiation"].(string); v != "true" {
			continue
		}
		eff, err := idx.Effective(name)
		if err != nil {
			continue
		}
		fid := firstString(eff["filament_id"])
		if fid == "" {
			continue
		}
		idx.instantiableByID[fid] = append(idx.instantiableByID[fid], name)
	}
	return idx, nil
}

// Effective walks the inheritance chain (root-first, so leaves override
// roots) and returns the merged dict for the named profile.
func (i *FilamentIndex) Effective(name string) (map[string]any, error) {
	if cached, ok := i.effectiveCache[name]; ok {
		return cached, nil
	}
	out, err := i.effective(name, map[string]bool{})
	if err == nil {
		i.effectiveCache[name] = out
	}
	return out, err
}

func (i *FilamentIndex) effective(name string, seen map[string]bool) (map[string]any, error) {
	if name == "" {
		return map[string]any{}, nil
	}
	if seen[name] {
		return nil, fmt.Errorf("inheritance cycle at %q", name)
	}
	seen[name] = true
	j, ok := i.byName[name]
	if !ok {
		return nil, fmt.Errorf("filament profile not found: %q", name)
	}
	parent, _ := j["inherits"].(string)
	merged, err := i.effective(parent, seen)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	for k, v := range merged {
		out[k] = v
	}
	for k, v := range j {
		out[k] = v
	}
	return out, nil
}

// HasName reports whether a profile with the given name exists in the index.
func (i *FilamentIndex) HasName(name string) bool {
	_, ok := i.byName[name]
	return ok
}

// NameForID returns an instantiable profile name whose effective filament_id
// matches id. When multiple profiles match (e.g. nozzle variants), prefer
// those whose `compatible_printers` includes preferPrinter; among the rest,
// pick the shortest name (so canonical 0.4 wins over "0.6 nozzle" etc.).
func (i *FilamentIndex) NameForID(id, preferPrinter string) (string, error) {
	candidates := i.instantiableByID[id]
	if len(candidates) == 0 {
		return "", fmt.Errorf("no instantiable filament with filament_id=%q", id)
	}
	var compatible []string
	for _, n := range candidates {
		eff, err := i.Effective(n)
		if err != nil {
			continue
		}
		cp, _ := eff["compatible_printers"].([]any)
		for _, v := range cp {
			if s, _ := v.(string); s == preferPrinter {
				compatible = append(compatible, n)
				break
			}
		}
	}
	pool := compatible
	if len(pool) == 0 {
		pool = candidates
	}
	sort.Slice(pool, func(a, b int) bool {
		if len(pool[a]) != len(pool[b]) {
			return len(pool[a]) < len(pool[b])
		}
		return pool[a] < pool[b]
	})
	return pool[0], nil
}

// FilamentTypeForName returns the filament_type of a profile, walking
// inheritance to find it (e.g. "PETG", "PLA", "PETG-CF").
func (i *FilamentIndex) FilamentTypeForName(name string) string {
	eff, err := i.Effective(name)
	if err != nil {
		return ""
	}
	return firstString(eff["filament_type"])
}

func firstString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		if len(t) > 0 {
			s, _ := t[0].(string)
			return s
		}
	}
	return ""
}
