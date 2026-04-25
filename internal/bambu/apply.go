package bambu

import (
	"fmt"
	"sort"
	"strings"

	"github.com/swh/openscad-skill/internal/threemf"
)

// ApplyFilament splices a Bambu Studio filament profile into slot 0 of a
// project's parallel arrays, sets every object's extruder to slot 0, and
// clears the slot-0 filament-diff list so Bambu Studio doesn't re-sync the
// values from a different system preset on open.
func ApplyFilament(t *threemf.ThreeMF, idx *FilamentIndex, name string) error {
	if !idx.HasName(name) {
		return fmt.Errorf("filament profile not found: %q", name)
	}
	profile, err := idx.Effective(name)
	if err != nil {
		return err
	}

	cfg, err := t.ProjectSettings()
	if err != nil {
		return err
	}

	// Splat slot 0 values from the resolved profile.
	for k, v := range profile {
		if _, skip := FilamentMetaSkip[k]; skip {
			continue
		}
		existing, ok := cfg[k]
		if !ok {
			continue
		}
		existingList, ok := existing.([]any)
		if !ok {
			continue
		}
		newVal := unwrapFirst(v)
		if newVal == "nil" {
			continue
		}
		if len(existingList) == 0 {
			cfg[k] = []any{newVal}
		} else {
			existingList[0] = newVal
		}
	}

	// Force the chosen profile's name into filament_settings_id[0].
	if list, ok := cfg["filament_settings_id"].([]any); ok && len(list) > 0 {
		list[0] = name
	}

	// Clear slot 0's filament diff list — values now match a system preset.
	cfg["different_settings_to_system"] = clearFilamentSlotDiff(cfg["different_settings_to_system"], 0)

	if err := t.WriteProjectSettings(cfg); err != nil {
		return err
	}

	t.SetEveryObjectExtruder(1)
	return nil
}

// ApplyOverrides writes user --set KEY=VALUE pairs into project_settings.
// Keys whose existing value is a list are written into slot 0 only and
// added to different_settings_to_system[2] (slot-0 filament diffs).
// Scalar keys replace the whole value and are added to diff[0]
// (print/process diffs).
func ApplyOverrides(t *threemf.ThreeMF, overrides map[string]string) error {
	if len(overrides) == 0 {
		return nil
	}
	cfg, err := t.ProjectSettings()
	if err != nil {
		return err
	}

	var printKeys, filamentKeys []string
	for k, v := range overrides {
		existing, ok := cfg[k]
		if ok {
			if list, isList := existing.([]any); isList {
				if len(list) == 0 {
					cfg[k] = []any{v}
				} else {
					list[0] = v
				}
				filamentKeys = append(filamentKeys, k)
				continue
			}
		}
		cfg[k] = v
		printKeys = append(printKeys, k)
	}

	diffs := normaliseDiffs(cfg["different_settings_to_system"])
	if len(printKeys) > 0 {
		diffs[0] = mergeSemicolon(diffs[0], printKeys)
	}
	if len(filamentKeys) > 0 {
		diffs[2] = mergeSemicolon(diffs[2], filamentKeys)
	}
	cfg["different_settings_to_system"] = toAnySlice(diffs)

	return t.WriteProjectSettings(cfg)
}

// ApplyTemplateDefaults is the build-template equivalent of ApplyOverrides:
// writes the configured override keys into project_settings AND adds them
// to different_settings_to_system[0] (canonical defaults are a print-side
// concern, not filament-side).
func ApplyTemplateDefaults(t *threemf.ThreeMF, defaults map[string]string) error {
	if len(defaults) == 0 {
		return nil
	}
	cfg, err := t.ProjectSettings()
	if err != nil {
		return err
	}
	for k, v := range defaults {
		cfg[k] = v
	}
	keys := make([]string, 0, len(defaults))
	for k := range defaults {
		keys = append(keys, k)
	}
	diffs := normaliseDiffs(cfg["different_settings_to_system"])
	diffs[0] = mergeSemicolon(diffs[0], keys)
	cfg["different_settings_to_system"] = toAnySlice(diffs)
	return t.WriteProjectSettings(cfg)
}

func unwrapFirst(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		if len(t) == 0 {
			return ""
		}
		s, _ := t[0].(string)
		return s
	}
	return ""
}

func mergeSemicolon(existing string, add []string) string {
	have := map[string]struct{}{}
	for _, s := range strings.Split(existing, ";") {
		if s != "" {
			have[s] = struct{}{}
		}
	}
	for _, s := range add {
		if s != "" {
			have[s] = struct{}{}
		}
	}
	keys := make([]string, 0, len(have))
	for k := range have {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ";")
}

func normaliseDiffs(v any) []string {
	switch t := v.(type) {
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			s, _ := x.(string)
			out = append(out, s)
		}
		for len(out) < 3 {
			out = append(out, "")
		}
		return out
	}
	return []string{"", "", ""}
}

func toAnySlice(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

func clearFilamentSlotDiff(v any, slot int) []any {
	d := normaliseDiffs(v)
	idx := slot + 2 // index 0 = print, 1 = printer, 2.. = filament slots
	for len(d) <= idx {
		d = append(d, "")
	}
	d[idx] = ""
	return toAnySlice(d)
}
