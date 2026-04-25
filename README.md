# openscad-bosl2 — a Claude Skill

A [Claude Code](https://claude.ai/code) skill that helps Claude write OpenSCAD `.scad` files using the [BOSL2](https://github.com/BelfrySCAD/BOSL2) library, in a style suited to 3D-printable utility parts (jigs, holders, brackets, fixtures, household items).

It encodes a house style — BOSL2 idioms, anchor-driven layout, `epsilon` / `layer_height` baselines, Customizer parameters — plus filament and slicer recommendations, print-orientation reasoning, and reference docs for the BOSL2 API and the OpenSCAD Customizer.

## Install

Open Claude Code in any directory and paste this prompt, replacing the printer model with your own:

> Install the `openscad-bosl2` skill from `https://github.com/swh/openscad-skill` for me.
>
> 1. Clone it to `~/Projects/openscad-skill` and symlink it to `~/.claude/skills/openscad-bosl2`.
> 2. Make sure OpenSCAD is installed (point me at the download page if not), and that BOSL2 is cloned into my OpenSCAD libraries directory.
> 3. I have a **Bambu X1C**. Add a short `## Printer` section to `~/.claude/CLAUDE.md` (creating the file if needed) that names the printer and points at `~/.claude/skills/openscad-bosl2/references/printers.md` for its specs.
> 4. Confirm the install with a one-line summary.

That's the whole install. Restart your Claude Code session afterwards so the skill is picked up.

For other printers, swap the model name in step 3 — `references/printers.md` has spec blocks for Bambu A1 Mini / A1 / P1P / P1S / X1C / X1E / H2D, Prusa MK4S / Core One / XL, and Creality K1 / K1 Max / K2 Plus. If your printer isn't listed, tell Claude what it is and ask it to add a row for it.

### Manual install

If you'd rather do it yourself:

```sh
git clone https://github.com/swh/openscad-skill.git
mkdir -p ~/.claude/skills
ln -s openscad-skill ~/.claude/skills/openscad-bosl2
```

Then add a `## Printer` section to `~/.claude/CLAUDE.md` naming your model and pointing at the skill's spec table. Three lines is enough:

```markdown
## Printer

**Bambu X1C** — see `~/.claude/skills/openscad-bosl2/references/printers.md` for specs.
```

You also need [OpenSCAD](https://openscad.org/downloads.html) and [BOSL2](https://github.com/BelfrySCAD/BOSL2) (clone into your OpenSCAD libraries directory — `~/Documents/OpenSCAD/libraries/` on macOS / Windows, `~/.local/share/OpenSCAD/libraries/` on Linux).

## Multiple printers

If you switch between printers, scope the printer block to whichever directory tree you use that printer in. Claude Code loads `CLAUDE.md` files from the working directory and its parents, so a child file overrides a more general one:

```
~/Documents/3D prints/
├── CLAUDE.md                   # default printer (e.g. X1C)
├── for-the-mk4s/
│   ├── CLAUDE.md               # Prusa MK4S
│   └── ...
└── for-the-a1/
    ├── CLAUDE.md               # Bambu A1
    └── ...
```

Ask Claude to set this up for you the same way as the install: tell it which printers go where.

## What's in the skill

- [`SKILL.md`](SKILL.md) — entry point loaded by Claude. House style, baselines, filament guidance, slicer-parameter recommendations, common gotchas, a starter skeleton.
- [`references/`](references/) — deeper sub-docs:
  - `bosl2-shapes.md` — 2D / 3D primitives and the anchor / spin / orient system.
  - `bosl2-transforms.md` — movement helpers, distributors, mutators.
  - `bosl2-rounding.md` — built-in rounding/chamfer, `rounded_prism`, edge profiles and masks.
  - `customizer.md` — full OpenSCAD Customizer annotation reference.
  - `printers.md` — common-printer spec table.
  - `bosl2-tutorials.md` — index of upstream BOSL2 tutorials.
- [`tools/`](tools/) — `openscad-check` (syntax / semantic check) and `openscad-render` (render to PNG). Both shell out to the `openscad` binary; honour `$OPENSCAD` to override its location.

## Updating

```sh
cd path/to/repo/openscad-skill && git pull
```

Or ask Claude: *"Update the openscad-bosl2 skill."* Changes are live in any new Claude Code session.
