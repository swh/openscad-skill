# openscad-bosl2 — a Claude Skill

A [Claude Code](https://claude.ai/code) skill that helps Claude write OpenSCAD `.scad` files using the [BOSL2](https://github.com/BelfrySCAD/BOSL2) library, in a style suited to 3D-printable utility parts (jigs, holders, brackets, fixtures, household items).

It encodes a house style — BOSL2 idioms, anchor-driven layout, `epsilon` / `layer_height` baselines, Customizer parameters — plus filament and slicer recommendations, print-orientation reasoning, and reference docs for the BOSL2 API and the OpenSCAD Customizer.

## Prerequisites

- **Claude Code** (the skill is loaded by Claude Code from `~/.claude/skills/`).
- **OpenSCAD** — install from [openscad.org](https://openscad.org/downloads.html). The `tools/openscad-check` and `tools/openscad-render` helpers shell out to the `openscad` binary.
- **BOSL2** — clone into your OpenSCAD libraries directory:
  ```sh
  # macOS:
  git clone https://github.com/BelfrySCAD/BOSL2.git \
      "$HOME/Documents/OpenSCAD/libraries/BOSL2"
  # Linux:
  git clone https://github.com/BelfrySCAD/BOSL2.git \
      "$HOME/.local/share/OpenSCAD/libraries/BOSL2"
  # Windows (PowerShell):
  git clone https://github.com/BelfrySCAD/BOSL2.git `
      "$HOME\Documents\OpenSCAD\libraries\BOSL2"
  ```

## Install

Clone this repo somewhere persistent, then symlink it into your user-level skills directory:

```sh
git clone https://github.com/swh/openscad-skill.git ~/Projects/openscad-skill
mkdir -p ~/.claude/skills
ln -s ~/Projects/openscad-skill ~/.claude/skills/openscad-bosl2
```

A symlink (rather than a copy) means edits to the cloned repo apply to the installed skill instantly.

To verify, start Claude Code in any directory and ask it to "create a small openscad part" — it should pick up the skill automatically.

### Project-level install

If you'd rather scope the skill to a single project, symlink into that project's `.claude/skills/` instead:

```sh
mkdir -p path/to/project/.claude/skills
ln -s ~/Projects/openscad-skill path/to/project/.claude/skills/openscad-bosl2
```

The skill's `tools/` paths in `SKILL.md` assume the user-level location. For a project install, point Claude at the actual path (e.g. `path/to/project/.claude/skills/openscad-bosl2/tools/openscad-check`).

## Tell the skill which printer(s) you have

The skill is printer-agnostic by default. Capacity and material recommendations depend on a `## Printer` block in a `CLAUDE.md` that's in scope for your work.

**For one printer used everywhere:** put the block in `~/.claude/CLAUDE.md` (loads for every project):

```markdown
## Printer

- **Model:** Bambu X1C
- **Build volume:** 256 × 256 × 256 mm (single-colour); ≤ ~200 × 200 mm usable when multi-material with frequent transitions (purge tower)
- **Enclosed:** Yes, passive chamber ~50 °C
- **Max nozzle / bed:** 300 °C / 100 °C
- **Hardened nozzle:** Stock
- **Multi-material:** AMS (4 spools per unit, up to 4 units = 16 colours)
- **Default material:** PETG
- **Slicer:** Bambu Studio / OrcaSlicer
```

**For one printer used only in a particular tree** (e.g. you keep all your `.scad` work under `~/Documents/3D prints/`): put the block in `~/Documents/3D prints/CLAUDE.md`. Claude Code auto-loads `CLAUDE.md` files from the working directory and its parents.

**For multiple printers:** create one project per printer (or one tree per printer), each with its own `CLAUDE.md` pinning that printer. When Claude is invoked from inside a given project, it picks up that project's printer spec.

```
~/Documents/3D prints/
├── CLAUDE.md                   # default printer (e.g. X1C)
├── for-the-mk4s/
│   ├── CLAUDE.md               # overrides: Prusa MK4S
│   └── ...
└── for-the-a1/
    ├── CLAUDE.md               # overrides: Bambu A1
    └── ...
```

A child `CLAUDE.md` is loaded *in addition to* parents — not as a replacement — so keep the printer block self-contained in the most-specific file.

See [`references/printers.md`](references/printers.md) for specs of common printers (Bambu A1/P1S/X1C/X1E/H2D, Prusa MK4S/Core One/XL, Creality K1/K1 Max/K2 Plus) — copy and adapt the relevant block.

## What's in the skill

- [`SKILL.md`](SKILL.md) — entry point loaded by Claude. House style, baselines, filament guidance, slicer-parameter recommendations, common gotchas, a starter skeleton.
- [`references/`](references/) — deeper sub-docs:
  - `bosl2-shapes.md` — 2D / 3D primitives and the anchor / spin / orient system.
  - `bosl2-transforms.md` — movement helpers, distributors, mutators.
  - `bosl2-rounding.md` — built-in rounding/chamfer, `rounded_prism`, edge profiles and masks.
  - `customizer.md` — full OpenSCAD Customizer annotation reference.
  - `printers.md` — common-printer spec table.
  - `bosl2-tutorials.md` — index of upstream BOSL2 tutorials.
- [`tools/`](tools/) — `openscad-check` (syntax / semantic check) and `openscad-render` (render to PNG). Both honour `$OPENSCAD` to override the binary path.

## Updating

If you cloned and symlinked, just `git pull` in the clone:

```sh
cd ~/Projects/openscad-skill && git pull
```

Changes are live in any new Claude Code session.
