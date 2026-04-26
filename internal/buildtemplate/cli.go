// Package buildtemplate is the entry-point logic for the
// openscad-build-template binary. The binary is a tiny shim that calls
// Main(); the shim's main.go is generated into cmd/ at build time by the
// Makefile and removed afterwards, so cmd/ doesn't need to be tracked in
// git.
package buildtemplate

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/swh/openscad-skill/internal/bambu"
	"github.com/swh/openscad-skill/internal/threemf"
)

const defaultSlicer = "bambu-studio"

func defaultTemplatesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "openscad-bosl2", "templates")
}

// overridesBambuStudio matches the SKILL.md "Suggested slicer settings"
// example: biased toward strength + cosmetic top quality vs. Bambu's stock
// 0.2 mm preset.
var overridesBambuStudio = map[string]string{
	"layer_height":          "0.2",
	"wall_loops":            "4",
	"top_shell_layers":      "6",
	"bottom_shell_layers":   "5",
	"sparse_infill_density": "25%",
	"sparse_infill_pattern": "cubic",
	"enable_support":        "0",
	"support_type":          "tree(auto)",
	"brim_type":             "auto_brim",
	"brim_width":            "7",
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "openscad-build-template: "+format+"\n", args...)
	os.Exit(1)
}

func usage(code int) {
	fmt.Fprint(os.Stderr,
		`usage: openscad-build-template [options] INPUT.3mf

Build a canonical openscad-pack-3mf template from a slicer-saved project 3MF.

Options:
  --slicer NAME    Slicer name (default: bambu-studio). Picks both the override
                   key set AND the output filename.
  -o, --output P   Override default output path
                   (~/.config/openscad-bosl2/templates/<slicer>.3mf).
  -h, --help       Show this help.

Currently supported slicers: bambu-studio.
`)
	os.Exit(code)
}

// Main is the entry point invoked by cmd/openscad-build-template/main.go.
func Main() {
	fs := flag.NewFlagSet("openscad-build-template", flag.ExitOnError)
	fs.Usage = func() { usage(2) }

	slicer := fs.String("slicer", defaultSlicer, "slicer name")
	output := fs.String("output", "", "output path")
	fs.StringVar(output, "o", "", "output path (alias)")
	help := fs.Bool("help", false, "show help")
	fs.BoolVar(help, "h", false, "show help (alias)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}

	if *help {
		usage(0)
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "openscad-build-template: exactly one input .3mf is required")
		usage(2)
	}
	in := fs.Arg(0)

	if !strings.EqualFold(filepath.Ext(in), ".3mf") {
		die("input must be a .3mf file: %s", in)
	}
	st, err := os.Stat(in)
	if err != nil || st.IsDir() {
		die("input not found: %s", in)
	}

	out := *output
	if out == "" {
		out = filepath.Join(defaultTemplatesDir(), *slicer+".3mf")
	}

	if *slicer != defaultSlicer {
		die("only --slicer bambu-studio is currently supported, got %q", *slicer)
	}
	defaults := overridesBambuStudio

	t, err := threemf.Open(in)
	if err != nil {
		die("%v", err)
	}

	if err := bambu.ApplyTemplateDefaults(t, defaults); err != nil {
		die("apply defaults: %v", err)
	}
	t.ZeroPlacementTransforms()
	t.DropThumbnails()

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		die("mkdir %s: %v", filepath.Dir(out), err)
	}
	if err := t.WriteTo(out); err != nil {
		die("write %s: %v", out, err)
	}

	fmt.Printf("%s -> %s\n", in, out)
	fmt.Printf("  applied %d %s overrides; transforms stripped; thumbnails dropped\n",
		len(defaults), *slicer)
}
