# openscad-bosl2 — a Claude Skill

A [Claude Code](https://claude.ai/code) skill that helps Claude write OpenSCAD `.scad` files using the [BOSL2](https://github.com/BelfrySCAD/BOSL2) library, in a style suited to 3D-printable utility parts (jigs, holders, brackets, fixtures, household items).

It encodes a house style — BOSL2 idioms, anchor-driven layout, `epsilon` / `layer_height` baselines, Customizer parameters — plus filament and slicer recommendations, print-orientation reasoning, and reference docs for the BOSL2 API and the OpenSCAD Customizer.

## Install

Open Claude Code in any directory and paste this prompt, replacing the project directory and printer model with your own:

> Install the `openscad-bosl2` skill from `https://github.com/swh/openscad-skill` for me.
>
> 1. Clone it to **`~/Projects/openscad-skill`** and symlink the **`openscad-bosl2/` subdirectory** to `~/.claude/skills/openscad-bosl2` (only that subdir is the skill — the repo root holds Go source and build infrastructure).
> 2. Run `make` in the clone to build the Go helper binaries into `openscad-bosl2/tools/`.
> 3. Make sure OpenSCAD is installed (point me at the download page if not), and that BOSL2 is cloned into my OpenSCAD libraries directory.
> 4. I have a **Bambu X1C**. Add a short `## Printer` section to `~/.claude/CLAUDE.md` (creating the file if needed) that names the printer and points at `~/.claude/skills/openscad-bosl2/references/printers.md` for its specs.
> 5. Confirm the install with a one-line summary.

That's the whole install. Restart your Claude Code session afterwards so the skill is picked up.

The Go binaries are needed for `openscad-pack-3mf` (3MF splicing with filament/AMS-aware settings) and `openscad-build-template`. Building requires Go ≥ 1.21 — install from [go.dev](https://go.dev/dl/) if needed.

For other printers, swap the model name in step 3 — `references/printers.md` has spec blocks for Bambu A1 Mini / A1 / P1P / P1S / X1C / X1E / H2D, Prusa MK4S / Core One / XL, and Creality K1 / K1 Max / K2 Plus. If your printer isn't listed, tell Claude what it is and ask it to add a row for it.

### Manual install

If you'd rather do it yourself:

```sh
git clone https://github.com/swh/openscad-skill.git
cd openscad-skill && make            # builds the Go helpers into openscad-bosl2/tools/
mkdir -p ~/.claude/skills
ln -s "$PWD/openscad-bosl2" ~/.claude/skills/openscad-bosl2
```

The repo layout:

- `openscad-bosl2/` — the skill itself: `SKILL.md`, `references/`, and a `tools/` directory that holds the bash helpers, seed files, and (after `make`) the Go binaries. This subdir is what gets symlinked or installed under `~/.claude/skills/`.
- `cmd/`, `internal/`, `go.mod`, `Makefile`, this `README.md` — build infrastructure at the repo root. Not part of the skill, not installed.

The Go binaries (`openscad-bosl2/tools/openscad-pack-3mf`, `openscad-bosl2/tools/openscad-build-template`) are gitignored — `make` writes them into the skill's tools/ directly.

Then add a `## Printer` section to `~/.claude/CLAUDE.md` naming your model and pointing at the skill's spec table. Three lines is enough:

```markdown
## Printer

**Bambu X1C** — see `~/.claude/skills/openscad-bosl2/references/printers.md` for specs.
```

You also need [OpenSCAD](https://openscad.org/downloads.html), [BOSL2](https://github.com/BelfrySCAD/BOSL2) (clone into your OpenSCAD libraries directory — `~/Documents/OpenSCAD/libraries/` on macOS / Windows, `~/.local/share/OpenSCAD/libraries/` on Linux), and [Go ≥ 1.21](https://go.dev/dl/) at install time (only needed to build the helpers).

### AMS state: cached vs live (optional)

Most users don't need to do anything here — `tools/openscad-pack-3mf` reads AMS state from Bambu Studio's local cache (`BambuStudio.conf`), which Bambu Studio updates over MQTT whenever it's running and connected to your printer. As long as Bambu Studio has been open recently with the printer online, the cache reflects what's currently loaded.

The cache only goes stale if you swap spools while Bambu Studio is **closed**. If that happens to you regularly, you can opt into live MQTT reads — but it requires enabling LAN-only mode on the printer.

#### Live MQTT (opt-in, LAN-mode only)

To read AMS state straight from the printer, set credentials and enable LAN-only mode. Credentials come from env vars and/or a CLAUDE.md block; env vars override CLAUDE.md.

**The access code is a secret** — put it in an env var, not CLAUDE.md (which is loaded into Claude's prompt context):

```sh
# ~/.zshrc or ~/.bashrc
export BAMBU_ACCESS_CODE=12345678          # from the printer screen
export BAMBU_SERIAL=00M09C460402058        # optional — can also live in CLAUDE.md
export BAMBU_HOST=192.168.1.42             # optional — can also live in CLAUDE.md
```

Less sensitive fields (serial, host) can live in `~/.claude/CLAUDE.md` if you'd rather Claude knows them:

```markdown
## Bambu printer (MQTT)

- Serial: 00M09C460402058
- Host: 192.168.1.42
```

The full set of recognised env vars: `BAMBU_ACCESS_CODE`, `BAMBU_SERIAL`, `BAMBU_HOST`.

**LAN-only mode caveat.** On firmware released after ~May 2024, the printer's local MQTT broker is reachable *only* when LAN-only mode is enabled (Settings → WLAN → LAN Only Mode — the IP and access code are both on that page). LAN-only mode disables Bambu Cloud features: no Bambu Handy app monitoring, no MakerWorld direct print, no cloud slicing. Most users prefer to leave cloud mode on and rely on the cache, which is what the tool does by default. See the [Bambu Lab Wiki](https://wiki.bambulab.com/en/knowledge-sharing/enable-lan-mode) for the trade-offs.

If MQTT is configured but unreachable (cloud mode without LAN-only, printer asleep, wrong IP, etc.), the tool prints a one-line warning to stderr and falls through to the cache — so it's safe to leave the env vars set; you'll just transparently get cache when MQTT can't reach.

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

**The skill (`openscad-bosl2/`)**:

- [`openscad-bosl2/SKILL.md`](openscad-bosl2/SKILL.md) — entry point loaded by Claude. House style, baselines, filament guidance, slicer-parameter recommendations, common gotchas, a starter skeleton.
- [`openscad-bosl2/references/`](openscad-bosl2/references/) — deeper sub-docs:
  - `bosl2-shapes.md` — 2D / 3D primitives and the anchor / spin / orient system.
  - `bosl2-transforms.md` — movement helpers, distributors, mutators.
  - `bosl2-rounding.md` — built-in rounding/chamfer, `rounded_prism`, edge profiles and masks.
  - `customizer.md` — full OpenSCAD Customizer annotation reference.
  - `printers.md` — common-printer spec table.
  - `bosl2-tutorials.md` — index of upstream BOSL2 tutorials.
- [`openscad-bosl2/tools/`](openscad-bosl2/tools/) — Bash: `openscad-check` (syntax / semantic check) and `openscad-render` (PNG render). Go: `openscad-pack-3mf` (geometry → configured Bambu Studio project 3MF, with AMS-aware filament selection) and `openscad-build-template` (canonical template builder). The Go binaries are gitignored and built by `make`.

**Build infrastructure (repo root)**:

- [`cmd/`](cmd/) and [`internal/`](internal/) — Go source for the helpers.
- [`Makefile`](Makefile) — `make build` to build, `make install` to copy the skill into `~/.claude/skills/<name>/`, `make clean` / `make test`.
- `go.mod` / `go.sum` — Go module definition.

## Updating

```sh
cd path/to/repo/openscad-skill && git pull && make
```

Or ask Claude: *"Update the openscad-bosl2 skill."* Run `make` after pulling if anything under `cmd/` or `internal/` changed; the binaries are not committed.
