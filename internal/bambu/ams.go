package bambu

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AMSState maps printer serial → ordered list of filament_id codes per
// AMS slot. Empty slots are represented as empty strings; this matches the
// shape produced by both the cache and the live MQTT reader so callers
// don't need to care which path was taken.
type AMSState map[string][]string

// AMSSource records where a state came from for the success summary.
type AMSSource struct {
	Live bool   // true for MQTT, false for cache
	Note string // freeform info ("BambuStudio.conf", "MQTT 192.168.1.42", etc.)
}

// ReadAMSCache parses BambuStudio.conf's `ams_filament_ids` field. Returns
// an empty (not nil) state if the file or key is missing — callers should
// check len() to decide whether to fall through.
func ReadAMSCache(studioDir string) (AMSState, error) {
	conf := filepath.Join(studioDir, "BambuStudio.conf")
	data, err := os.ReadFile(conf)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return AMSState{}, nil
		}
		return nil, err
	}
	var top map[string]any
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("parse BambuStudio.conf: %w", err)
	}
	raw, _ := top["ams_filament_ids"].(map[string]any)
	out := AMSState{}
	for serial, v := range raw {
		csv, _ := v.(string)
		parts := strings.Split(csv, ",")
		// Drop a single trailing empty (Bambu writes "id1,id2,id3,id4,").
		if len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) == "" {
			parts = parts[:len(parts)-1]
		}
		slots := make([]string, len(parts))
		for i, p := range parts {
			slots[i] = strings.TrimSpace(p)
		}
		out[serial] = slots
	}
	return out, nil
}

// PickFilamentNameBySlot resolves AMS slot N's filament_id to an instantiable
// profile name compatible with the project's printer.
func PickFilamentNameBySlot(state AMSState, projectPrinter string, idx *FilamentIndex, slot int) (string, error) {
	slots, err := singleSlots(state)
	if err != nil {
		return "", err
	}
	if slot < 0 || slot >= len(slots) {
		return "", fmt.Errorf("AMS slot %d out of range (have %d slots: %v)", slot, len(slots), slots)
	}
	id := slots[slot]
	if id == "" {
		return "", fmt.Errorf("AMS slot %d is empty", slot)
	}
	return idx.NameForID(id, projectPrinter)
}

// PickFilamentNameByType walks AMS slots in order, returning the first
// instantiable profile whose filament_type matches `wantType`
// (case-insensitive).
func PickFilamentNameByType(state AMSState, projectPrinter string, idx *FilamentIndex, wantType string) (string, error) {
	slots, err := singleSlots(state)
	if err != nil {
		return "", err
	}
	target := strings.ToUpper(strings.TrimSpace(wantType))
	for _, id := range slots {
		if id == "" {
			continue
		}
		name, err := idx.NameForID(id, projectPrinter)
		if err != nil {
			continue
		}
		if strings.ToUpper(idx.FilamentTypeForName(name)) == target {
			return name, nil
		}
	}
	return "", fmt.Errorf("no AMS slot loaded with filament_type=%q (slots: %v)", wantType, slots)
}

// singleSlots picks the right printer's slot list. If exactly one printer is
// in the state we use it; if multiple, we error rather than guess.
func singleSlots(state AMSState) ([]string, error) {
	if len(state) == 0 {
		return nil, errors.New("AMS state is empty")
	}
	if len(state) > 1 {
		serials := make([]string, 0, len(state))
		for k := range state {
			serials = append(serials, k)
		}
		return nil, fmt.Errorf("multiple printers in AMS state (%v); pass --filament NAME explicitly", serials)
	}
	for _, slots := range state {
		return slots, nil
	}
	return nil, errors.New("unreachable")
}
