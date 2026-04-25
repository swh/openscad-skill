// Package bambu encapsulates Bambu Studio knowledge: data directory
// discovery, filament profile inheritance, AMS state (cached and live via
// MQTT), and the rules for splicing a filament into a project's slot 0.
package bambu

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// FindStudioDir returns the first existing Bambu Studio data directory.
func FindStudioDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "BambuStudio"), // macOS
		filepath.Join(home, ".config", "BambuStudio"),                        // Linux
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			candidates = append(candidates, filepath.Join(appData, "BambuStudio"))
		}
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("Bambu Studio data directory not found (tried %v)", candidates)
}

// FindFilamentDir returns <studioDir>/system/BBL/filament if present.
func FindFilamentDir(studioDir string) (string, error) {
	p := filepath.Join(studioDir, "system", "BBL", "filament")
	if st, err := os.Stat(p); err == nil && st.IsDir() {
		return p, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("filament dir missing: %s", p)
	} else {
		return "", err
	}
}
