// Package claudemd locates and parses Bambu printer credentials from
// CLAUDE.md files. Credentials live under a section header containing
// "bambu" (case-insensitive); supported keys are Serial, Access code, and
// Host (the parser is forgiving about formatting).
//
// Multiple CLAUDE.md files in the search path are merged with most-specific
// (closest to cwd) winning per-field, so a project file can override only
// Host while inheriting Serial/AccessCode from ~/.claude/CLAUDE.md.
package claudemd

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Creds holds the resolved printer credentials. Host may be empty (in
// which case the MQTT path is skipped — only the AMS-cache fallback runs).
type Creds struct {
	Serial     string
	AccessCode string
	Host       string
	// Sources records which file each field came from, for diagnostics.
	Sources map[string]string
}

// HasMQTT reports whether all three fields are present (the minimum needed
// to attempt an MQTT connection).
func (c *Creds) HasMQTT() bool {
	return c != nil && c.Serial != "" && c.AccessCode != "" && c.Host != ""
}

// SearchPaths returns the ordered list of CLAUDE.md files to consult,
// most-specific first: cwd, walking up to home, plus ~/.claude/CLAUDE.md.
// Files that don't exist are skipped at parse time.
func SearchPaths(cwd, home string) []string {
	var paths []string
	if cwd == "" || home == "" {
		return paths
	}
	cwd, _ = filepath.Abs(cwd)
	home, _ = filepath.Abs(home)

	// Walk cwd up to home (inclusive). Stop at filesystem root for safety.
	cur := cwd
	for {
		paths = append(paths, filepath.Join(cur, "CLAUDE.md"))
		if cur == home || cur == filepath.Dir(cur) {
			break
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}

	// User-global file, if not already covered.
	userGlobal := filepath.Join(home, ".claude", "CLAUDE.md")
	if !contains(paths, userGlobal) {
		paths = append(paths, userGlobal)
	}
	return paths
}

// Resolve walks the given paths in order and merges Bambu credential blocks
// from each. The first non-empty value seen for a given field wins (so the
// caller-provided list should be ordered most-specific first).
func Resolve(paths []string) (*Creds, error) {
	c := &Creds{Sources: map[string]string{}}
	for _, p := range paths {
		fields, err := ParseBambuSection(p)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		set := func(key string, val string) {
			if val == "" {
				return
			}
			switch key {
			case "serial":
				if c.Serial == "" {
					c.Serial = val
					c.Sources["serial"] = p
				}
			case "access_code":
				if c.AccessCode == "" {
					c.AccessCode = val
					c.Sources["access_code"] = p
				}
			case "host":
				if c.Host == "" {
					c.Host = val
					c.Sources["host"] = p
				}
			}
		}
		set("serial", fields["serial"])
		set("access_code", fields["access_code"])
		set("host", fields["host"])
	}
	if c.Serial == "" && c.AccessCode == "" && c.Host == "" {
		return nil, nil
	}
	return c, nil
}

var (
	// reHeader matches a markdown heading containing "bambu" (case-insensitive).
	reHeader = regexp.MustCompile(`^\s*#{1,6}\s+.*[Bb][Aa][Mm][Bb][Uu]`)
	// reBlankOrText matches lines we ignore inside the section.
	reAnyHeader = regexp.MustCompile(`^\s*#{1,6}\s`)
	// reKV captures `Key: value`, optionally bullet-prefixed, with optional
	// surrounding **bold** markers on the key.
	reKV = regexp.MustCompile(`^\s*(?:[-*]\s+)?\*{0,2}\s*([A-Za-z][A-Za-z _\-]*?)\s*\*{0,2}\s*:\s*(.+?)\s*$`)
)

// ParseBambuSection extracts the first "bambu" section from path, returning
// a map with normalised keys ("serial", "access_code", "host"). Unknown keys
// are ignored. An empty map is returned (without error) if the file has no
// matching section.
func ParseBambuSection(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := map[string]string{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	inSection := false
	for scanner.Scan() {
		line := scanner.Text()
		if reHeader.MatchString(line) {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}
		// Any other heading ends the section.
		if reAnyHeader.MatchString(line) {
			break
		}
		m := reKV.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := normaliseKey(m[1])
		if key == "" {
			continue
		}
		val := stripInlineMarkdown(m[2])
		if val != "" {
			out[key] = val
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func normaliseKey(raw string) string {
	k := strings.ToLower(strings.TrimSpace(raw))
	k = strings.ReplaceAll(k, "-", "_")
	k = strings.ReplaceAll(k, " ", "_")
	switch k {
	case "serial", "serial_number", "sn":
		return "serial"
	case "access_code", "accesscode", "access", "code", "lan_code", "lan_access_code":
		return "access_code"
	case "host", "ip", "address", "hostname":
		return "host"
	}
	return ""
}

// stripInlineMarkdown removes a single layer of `**bold**` or `` `code` ``
// wrapping from a value. Leaves nested or unbalanced markers alone.
func stripInlineMarkdown(s string) string {
	s = strings.TrimSpace(s)
	for _, pair := range []struct{ open, close string }{
		{"**", "**"}, {"`", "`"},
	} {
		if strings.HasPrefix(s, pair.open) && strings.HasSuffix(s, pair.close) &&
			len(s) >= len(pair.open)+len(pair.close) {
			s = strings.TrimSpace(s[len(pair.open) : len(s)-len(pair.close)])
		}
	}
	return s
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
