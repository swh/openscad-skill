---
name: openscad-bosl2
description: Write or edit OpenSCAD (.scad) files using the BOSL2 library. Use whenever the user is designing a 3D-printable part — jigs, holders, brackets, enclosures, fixtures, household items — and the file is OpenSCAD. Covers the user's house style (BOSL2 idioms, anchors, epsilon/layer_height baselines, Customizer parameters) and links to detailed BOSL2 reference sub-docs.
---

# OpenSCAD with BOSL2

This skill guides Claude in writing OpenSCAD code that matches the user's house style: BOSL2-first, anchor-driven, parametric, Customizer-friendly, and tuned for FDM 3D printing.

The user's specific printer model is named in the project / parent `CLAUDE.md`. The CLAUDE.md typically only names the model — look up its specs (build volume, max nozzle/bed temps, enclosure, multi-material, hardened nozzle) in [`references/printers.md`](references/printers.md). If the CLAUDE.md doesn't name a printer, or names one that isn't in the table, **ask** before designing around capacity or material assumptions.

## When to use

Whenever you are creating or editing a `.scad` file. Most of the user's parts are utility prints — workshop jigs, tool holders, brackets, household fixtures — though some are display / cosmetic models.

## Required preamble

Every `.scad` file should start like this:

```scad
include <BOSL2/std.scad>
// include <BOSL2/rounding.scad>     // add when using offset_sweep, os_circle, etc.

/* [Geometry] */
// ... user-facing customizer parameters ...

/* [Hidden] */
$fn = 256;              // 100 for drafts, 256 for finished parts
epsilon = 0.01;         // offset cutters in difference() to avoid z-fighting
layer_height = 0.2;     // pick from context; do NOT expose to the Customizer
```

The three baselines are non-negotiable:

| Variable        | Purpose                                                                 |
|-----------------|-------------------------------------------------------------------------|
| `epsilon`       | Extend cutters in `difference()` so coplanar faces don't z-fight. Typical: `0.01` (or `0.001` for tight tolerances). |
| `layer_height`  | Snap z-axis features (floor thickness, label depths, embossed text) to integer multiples for clean prints. Common emboss depth: `2 * layer_height` to `4 * layer_height`. |
| `$fn`           | Polygon resolution — set globally, override per-call only when needed.   |

### Picking a `layer_height` value

`layer_height` is a constant in `[Hidden]`, **not** a Customizer parameter. Infer a value from the user's request:

- **Fine detail / cosmetic part** (visible features, embossed text, organic curves): `0.12` or `0.16`.
- **Default / general utility part** (most jigs, holders, brackets): `0.2`.
- **Fast / structural part** (large, low-detail, strength matters more than finish): `0.28` or `0.32`.

If the user's intent is unclear — e.g. they could plausibly want either a fast draft or a polished finish — **ask** before guessing. A one-line question is cheaper than printing the wrong thing.

## House style

### Prefer BOSL2 primitives over OpenSCAD built-ins

| Use this              | Not this              | Why                                                |
|-----------------------|-----------------------|----------------------------------------------------|
| `cuboid([x,y,z])`     | `cube([x,y,z])`       | Accepts `chamfer=`, `rounding=`, `edges=`, `anchor=` |
| `cyl(d=, h=)`         | `cylinder(d=, h=)`    | Accepts `chamfer1=`, `chamfer2=`, `rounding1/2=`, `anchor=` |
| `right(5)`, `up(10)`  | `translate([5,0,0])`  | One-line clarity for single-axis moves             |
| `rot(from=v1, to=v2)` | manual Euler angles   | Aligns one direction with another without trig    |

Use raw `cube`/`cylinder` only when you genuinely need them.

### Anchor everything explicitly

Pick a deliberate reference point on every primitive — `anchor=BOTTOM+LEFT+FRONT`, `anchor=TOP`, etc. Don't rely on default centering. This makes positioning robust to dimension changes.

```scad
cuboid([60, 40, 10], anchor=BOTTOM+LEFT+FRONT);     // corner at origin
cyl(d=20, h=30, anchor=BOTTOM);                      // sits on the build plate
```

### Chamfer or round every exposed vertex

As a default, every external vertex (edges and corners on the printed surface) should have a small **chamfer or roundover, typically 0.5–1.0 mm**. Sharp printed corners chip easily; a tiny break makes the part dramatically more robust and pleasant to handle. Use the built-in `chamfer=` / `rounding=` on `cuboid()` / `cyl()` rather than adding it after the fact.

```scad
cuboid([60, 40, 10], anchor=BOTTOM, chamfer=0.5);          // simple break
cyl(d=20, h=30, anchor=BOTTOM, chamfer1=1, chamfer2=0.5);  // bigger at the base
```

Skip this only when there's a reason — a mating face that needs to stay flat, an edge that disappears against another part, or a cosmetic line that should remain crisp.

### Bounding-box vectors

Name dimensions as `[x, y, z]` lists and access via `.x`/`.y`/`.z`:

```scad
BASE = [60, 40, 10];
cuboid(BASE, anchor=BOTTOM);
right(BASE.x/2) up(BASE.z) cyl(d=8, h=20);
```

This avoids passing three loose numbers around and makes the code resilient to resizes.

### Customizer parameters

Top of file. Use OpenSCAD Customizer annotations:

```scad
// Wall thickness in mm
wall = 4;             // [3:0.5:8]
// Number of legs
n_legs = 4;           // [3, 4, 5, 6]
// Print left-handed
left_handed = false;
```

Group related params with tab markers (`/* [Geometry] */`, `/* [Print] */`, `/* [Hidden] */`). Variables before the first `{` and outside `[Hidden]` show up in the Customizer.

See [`references/customizer.md`](references/customizer.md) for the full annotation grammar.

### Difference() with epsilon

When subtracting, extend the cutter past the surface by `epsilon` so the faces don't coincide:

```scad
difference() {
    cuboid(BASE, anchor=BOTTOM);
    down(epsilon)                                    // extend below the floor
        cyl(d=8, h=BASE.z + 2*epsilon, anchor=BOTTOM);
}
```

### Wrap multi-part output in `union()` or `module()`

A comment helps the slicer (and the next reader) understand intent:

```scad
// One object for the slicer
union() {
    base();
    handle();
}
```

If there are multiple discrete objects in one file, each should be in their own `module()` or `union()` at the top level.

### Echo derived dimensions

When a value is computed from inputs, echo it so the user can sanity-check:

```scad
slot_w = (assembly_w - 7 * gap) / 6;
echo(str("Slot width: ", slot_w, "mm"));
```

### Print orientation

Model the part in its final / functional orientation, then add a single top-level `rotate()` to flip it for printing. Leave a comment so the orientation can be toggled:

```scad
// Orient for printing — comment out for editing
rotate([90, 0, 0]) down(BASE.z) right(BASE.x/2)
    part();
```

**Pick the print orientation from the part's mechanical job, not just what lays flat.** FDM prints are anisotropic: layer-adhesion bonds are weaker than the filament itself, so prints are **weak in tension along the build (Z) axis** and prone to splitting between layers under load. When choosing an orientation:

- **Identify the load paths.** Where will tension, bending, or impact actually hit the part? A hook, hinge, bracket arm, lever, or screw boss has a clear "this must not snap" axis.
- **Align primary tensile loads with the build plate (X/Y), not Z.** A hook that pulls upward in use should be printed lying on its side, so the pull direction runs along the layers, not across them.
- **For multi-axis strength**, consider printing diagonally — rotated ~30–45° around X or Y so no single load axis is purely Z. This trades a small support penalty for much better all-direction toughness, and is often worth it for jigs and fixtures that get knocked around.
- **Overhangs and bridging** also matter, but mechanical strength usually wins when the two conflict — supports are fixable, a snapped part isn't.

If the load path isn't obvious from the brief — e.g. a holder that might or might not get yanked on — **ask** before committing to an orientation.

### Filament choice

Suggest a filament once you understand what the part is *for*. **Strongly prefer "easy" filaments** — they print reliably on most consumer FDM printers without an enclosure, filament dryer, or hardened nozzle. Only reach for harder materials when there's a specific technical reason, and check the project / parent `CLAUDE.md` to confirm the printer can handle them (max nozzle/bed temp, enclosure, hardened nozzle).

| Filament    | Strength / toughness         | Heat resistance | UV / outdoor | Ease     | Best for                                                       |
|-------------|------------------------------|-----------------|--------------|----------|----------------------------------------------------------------|
| **PLA**     | Stiff but brittle, low impact| ~60 °C (sags in a hot car) | Poor         | Easiest  | Indoor / display / models, low-load jigs, prototypes, cosmetic parts |
| **PETG**    | Tough, decent impact, slight flex | ~75 °C   | OK           | Easy     | **Default for utility parts** — workshop jigs, holders, brackets, fixtures, anything mildly mechanical or outdoor-adjacent |
| **PETG-CF** | Stiffer than PETG, dimensionally stable, more brittle | ~75 °C | OK | Easy     | Tight-tolerance jigs, parts where deflection matters more than impact, structural inserts |

Mention harder filaments **only when justified**:

- **ABS / ASA** — needs an enclosure; better heat resistance (~95 °C) and UV (ASA) than PETG. Suggest only for hot-environment or sun-exposed outdoor parts.
- **Nylon (PA / PA-CF)** — abrasion-resistant and tough; needs drying. Suggest for living hinges, gears, snap-fits seeing real cycles.
- **Polycarbonate (PC)** — extreme strength and heat; hard to print. Suggest only for high-load structural parts.
- **TPU** — flexible. Suggest only when the part *must* compress, bend, or grip.

When suggesting, name **one primary recommendation and one fallback**, and give the reason in one sentence:

> *"PETG is the default here — it'll handle being knocked around in the workshop and resists the occasional drop better than PLA. PLA would also work if it'll only see indoor light duty."*

If the use case is genuinely ambiguous (display *or* working tool? indoor *or* outdoor?), ask before recommending.

### Export format

When telling the user how to take a `.scad` file into the slicer, **recommend 3MF over STL** wherever the slicer supports it (Bambu Studio, OrcaSlicer, PrusaSlicer, recent Cura, SuperSlicer — i.e. essentially every modern slicer). 3MF preserves precision, includes units, supports multi-material/colour, and produces smaller files than ASCII STL.

OpenSCAD exports 3MF directly:

```sh
openscad -o part.3mf part.scad
```

Suggest STL only when the target tool genuinely doesn't accept 3MF.

**Pre-configured 3MF projects.** If the user has set up a slicer template (see [`openscad-pack-3mf`](#openscad-pack-3mf) under Tools), prefer it — it produces a `.3mf` that opens in Bambu Studio / OrcaSlicer with their printer, filament, and process already set up, plus the template's baked default settings. **For any setting where your part-specific recommendation diverges from those defaults, pass `--set KEY=VALUE` so the override travels with the file.** Don't fall back to plain `openscad -o part.3mf` just because you'd change one or two settings — list the divergences as `--set` flags and the user gets a one-click slice.

### Slicer parameters

After delivering the `.scad` file, recommend slicer settings unless the user has already specified them. Always include layer height; include the others when they're relevant to the part. Keep the recommendation to a few lines — not a full slicer profile.

**Always recommend:**

- **Layer height** — match the value in the `.scad` file (so embossed text, label depths, and floor thicknesses snap cleanly). Mention it explicitly so the user knows to set the slicer accordingly.
- **Supports** — say whether they're likely needed in the recommended print orientation. Most parts can be oriented to avoid them; flag any overhangs >45° from vertical or unbridgeable spans (>~10 mm). If supports are needed, suggest **tree supports** painted onto the offending faces (most modern slicers — Bambu Studio, OrcaSlicer, PrusaSlicer — support this).

**Recommend when relevant:**

- **Infill type:**
  - **Grid** — **default**. Fast crosshatch-style pattern with decent strength in XY; weak in Z but that's true of all infills. (Bambu Studio also has a similar "Cross Hatch" pattern.)
  - **Cubic / Adaptive Cubic** — slightly tougher than Grid with similar speed; Adaptive Cubic concentrates infill near walls for less material at the same effective strength. Good upgrade when the part takes real load.
  - **Gyroid** — near-isotropic strength, but **slow**. Reserve for parts that genuinely need omnidirectional toughness (hooks, pulled-on brackets) — not as a blanket default.
  - **Lightning** — fastest, but only supports the top layers; only use for cosmetic / hollow parts.
  - **Concentric** — for flexible (TPU) parts that need to compress evenly.
- **Infill density:**
  - 10–15 % — cosmetic / display / hollow models.
  - 15–25 % — light utility (most jigs, holders).
  - 30–50 % — load-bearing, mechanical, brackets that get yanked on.
  - 60 %+ — high-stress structural parts. Consider thicker walls instead — wall count usually contributes more to strength than infill density past ~40 %.
- **Wall (perimeter) count:**
  - 2 — cosmetic / non-structural.
  - **3 — default**.
  - 4–5 — strength-critical parts (hooks, brackets, anything that bears tension or impact). Often a better strength investment than raising infill.
- **Top / bottom layer count:**
  - 4 — default.
  - 5–6 — pressure-tight, watertight, or cosmetic flat tops where pillowing would show.
- **Brim** — recommend for small-footprint parts, tall thin parts, or anything with marginal bed contact.
- **Ironing** — recommend for visible flat top surfaces where appearance matters (badges, labelled tops, display pieces).

Skip parameters that don't apply (e.g. don't recommend ironing on a part with no flat top). A typical recommendation might read:

> *"Suggested slicer settings: 0.2 mm layer height, PETG, 4 walls, 25 % cubic infill, 6 top / 5 bottom layers, no supports needed in this orientation, auto brim 7 mm."*

These match the defaults baked into the canonical `openscad-pack-3mf` template — biased a notch toward strength and cosmetic flat-top quality vs. Bambu's stock 0.2 mm preset. Change them in the recommendation when the part calls for it (e.g. raise infill density for a load-bearing bracket, change pattern to gyroid for genuinely omnidirectional load, drop walls/shells back to 3/4 for fast cosmetic prints).

## Tools

This skill ships helper scripts and binaries. **Always invoke them by their full path inside the skill** — the working directory is rarely the skill repo, so `tools/openscad-check` will fail in most projects. Use:

```sh
~/.claude/skills/openscad-bosl2/tools/openscad-check FILE.scad
~/.claude/skills/openscad-bosl2/tools/openscad-render FILE.scad
```

(That path works because the skill is installed user-level via symlink. If a particular project installs the skill at `<project>/.claude/skills/openscad-bosl2/` instead, adjust accordingly.)

`openscad-check` and `openscad-render` are bash scripts. `openscad-pack-3mf` and `openscad-build-template` are Go binaries built from `cmd/` — run `make` in the skill's clone after `git pull` if anything under `cmd/` or `internal/` changed.

### `openscad-check`

Syntax / semantic check without doing any geometry work. Run this **after every meaningful edit** to a `.scad` file — it catches parse errors, unknown modules, undefined variables, and (by default) treats warnings as failures.

```sh
~/.claude/skills/openscad-bosl2/tools/openscad-check path/to/part.scad
~/.claude/skills/openscad-bosl2/tools/openscad-check --lenient path/to/part.scad   # don't fail on warnings
~/.claude/skills/openscad-bosl2/tools/openscad-check a.scad b.scad c.scad          # batch
```

Exit codes: `0` clean, `1` one or more files failed, `2` usage / OpenSCAD missing.

Don't claim a file "compiles" or "is correct" without running this.

### `openscad-render`

Render a `.scad` file to a PNG for visual review. Default is a full CGAL render with the camera framed to fit (`--viewall --autocenter`).

```sh
~/.claude/skills/openscad-bosl2/tools/openscad-render part.scad                       # -> part.png
~/.claude/skills/openscad-bosl2/tools/openscad-render --preview part.scad             # fast ThrownTogether
~/.claude/skills/openscad-bosl2/tools/openscad-render --size 1600x1200 part.scad      # custom resolution
~/.claude/skills/openscad-bosl2/tools/openscad-render -o /tmp/p.png part.scad         # explicit output path
~/.claude/skills/openscad-bosl2/tools/openscad-render --camera 0,0,0,55,0,25,200 part.scad  # specific view
~/.claude/skills/openscad-bosl2/tools/openscad-render --view axes,edges part.scad     # show axes + edges
```

Use `--preview` for quick visual sanity-checks on large/heavy files; the default render is slower but matches what slicers will see.

**Where the PNG goes.** Without `-o`, the tool writes `part.png` *next to* the source `.scad` — leave it there. **Don't pass `-o /tmp/...`** unless the user asks. Renders living next to the source make them easy to find later, easy to diff against future renders, and easy to surface back to the user via [`openscad-show`](#openscad-show).

### `openscad-show`

Open one or more rendered PNGs in the platform's default image viewer (`open` on macOS, `xdg-open` on Linux, `start` on Windows). Use this whenever the user asks to *see* / *look at* / *show me* a render.

```sh
~/.claude/skills/openscad-bosl2/tools/openscad-show part.png
~/.claude/skills/openscad-bosl2/tools/openscad-show a.png b.png c.png    # multiple
~/.claude/skills/openscad-bosl2/tools/openscad-show /path/to/dir          # most recent .png in that dir
~/.claude/skills/openscad-bosl2/tools/openscad-show -n 3 /path/to/dir     # top 3 most recent
```

For a directory argument, the tool picks the most recently modified `.png` — useful when the user says "show me the latest render of the bracket" and Claude knows the project dir but not the exact filename. The tool prints each path it opens to stdout, so Claude can see what was actually launched.

Default flow when the user asks to see a render:
1. If a render hasn't happened yet for the current part, run `openscad-render` first (writes the PNG next to the .scad).
2. Pass that PNG (or the project directory) to `openscad-show`.

Both scripts honour `$OPENSCAD` to override the binary path; otherwise they search `PATH` and standard install locations including `/Applications/OpenSCAD.app/Contents/MacOS/OpenSCAD`.

### `openscad-pack-3mf`

Splice OpenSCAD geometry into a slicer **template** project 3MF — the output opens in Bambu Studio / OrcaSlicer with the user's printer, filament, and process settings already applied (Bambu's project format is the only way slicer settings travel with a 3MF; this tool reuses a pre-saved project as a template).

**When to use it:** by default, every time. The template carries one fixed set of slicer defaults (baked in by `openscad-build-template`); for any setting where your part-specific recommendation diverges, **pass `--set KEY=VALUE`** to override on top of the template. The override is written into the project's settings *and* added to the slicer's "differs from preset" list, so it survives a project open instead of silently re-syncing.

For **filament selection**, use `--filament-type` / `--ams-slot` / `--filament` instead of `--set` (filament is a multi-key concern — temps, flow ratios, fan speeds — that only stays consistent when set together from a real Bambu profile).

```sh
# Recommendation aligns with template defaults — splice and go:
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf part.scad

# Recommendation diverges (load-bearing bracket: more infill, omnidirectional pattern, supports on):
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf \
    --set sparse_infill_density=50% \
    --set sparse_infill_pattern=gyroid \
    --set enable_support=1 \
    part.scad

# Use whatever PETG is currently in the AMS (preferred — survives spool changes):
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf --filament-type PETG part.scad

# Pick a specific AMS slot:
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf --ams-slot 2 part.scad

# Force a specific Bambu profile regardless of AMS state:
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf \
    --filament "Bambu PETG HF @BBL X1C" part.scad
```

**Filament selection rules:**

- Prefer `--filament-type TYPE` (e.g. `PLA`, `PETG`, `PETG-CF`, `ABS`) — the tool reads Bambu Studio's cached AMS state and picks the first slot loaded with that type. The user doesn't have to re-edit the recommendation when they change spools.
- Use `--ams-slot N` (0-indexed) when you specifically need slot N's filament.
- Use `--filament NAME` only when AMS isn't loaded with the right type or you need a specific profile (e.g. `Bambu PLA Matte @BBL X1C` over a Generic PLA).
- At most one of these flags per call. Apply *before* `--set`, so `--set nozzle_temperature=260` cleanly fine-tunes the filament's value.
- Caveat: AMS state is updated via MQTT when Bambu Studio is connected to the printer. If spools are swapped while Bambu Studio is closed, the cache goes stale.

Common Bambu Studio override keys (use the slicer-native names verbatim):

| Key                           | Example values                                      |
|-------------------------------|------------------------------------------------------|
| `layer_height`                | `0.12`, `0.16`, `0.2`, `0.28`                        |
| `wall_loops`                  | `2`, `3`, `4`, `5`                                   |
| `top_shell_layers`            | `4`, `5`, `6`                                        |
| `bottom_shell_layers`         | `4`, `5`                                             |
| `sparse_infill_density`       | `15%`, `25%`, `40%`, `60%`                           |
| `sparse_infill_pattern`       | `crosshatch`, `cubic`, `gyroid`, `grid`, `lightning`, `concentric` |
| `enable_support`              | `0` (off), `1` (on)                                  |
| `support_type`                | `tree(auto)`, `tree(organic)`, `normal(auto)`        |
| `brim_type`                   | `auto_brim`, `outer_only`, `no_brim`                 |
| `brim_width`                  | `5`, `8`                                             |
| `ironing_type`                | `topmost`, `top`, `no ironing`                       |

For filament-side keys (temperatures, flow ratios, fan speeds, etc.) prefer `--filament*` over piecewise `--set`. Use `--set nozzle_temperature=260` only as a fine-tune *on top* of a filament selection. Values are written as strings — Bambu's `project_settings.config` is all-strings. `--set` / `--filament` etc. only work with `--slicer bambu-studio` for now.

Fall back to plain `openscad -o part.3mf part.scad` only if the user explicitly doesn't want a configured project, or if they don't have a template set up yet.

```sh
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf part.scad         # -> part-packed.3mf
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf -o /tmp/p.3mf part.scad
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf part.3mf          # accepts pre-rendered 3MF too
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf --slicer prusa-slicer part.scad   # different slicer template
~/.claude/skills/openscad-bosl2/tools/openscad-pack-3mf -t /path/to/custom.3mf part.scad
```

By default it looks up `~/.config/openscad-bosl2/templates/<slicer>.3mf` (default `<slicer>` = `bambu-studio`). The template stores the settings; the tool replaces the geometry and centres the new part on the build plate.

**To set up a template the first time:**

1. Render the seed: `openscad -o /tmp/seed.3mf ~/.claude/skills/openscad-bosl2/tools/template-seed.scad`
2. Open `/tmp/seed.3mf` in your slicer; set printer / filament / process / supports / etc. as you'd like them.
3. **File → Save Project As…** to anywhere, e.g. `/tmp/my-template.3mf`.
4. Run [`openscad-build-template`](#openscad-build-template) on the saved project. It applies the skill's default overrides (and strips placement transforms / thumbnails) and writes the canonical template to `~/.config/openscad-bosl2/templates/bambu-studio.3mf`.

After that, every `openscad-pack-3mf` run produces a 3MF ready to slice with those settings.

**AMS state — cache by default, live MQTT opt-in:** the tool **always syncs the printer's current AMS slot contents** (from `BambuStudio.conf`'s cache, or live MQTT if creds are set) into every slot of the output project, so the user's saved template doesn't show stale filaments. The cache stays current as long as Bambu Studio has recently been open with the printer online. **Don't push the user toward live MQTT unless they ask** for it; the cache is the right answer for almost everyone.

`--ams-slot N` and `--filament-type T` both build on top of the default sync — they additionally set the active object's filament (slot 0) to a specific AMS slot's profile. `--filament NAME` skips the sync entirely on the assumption that the user wants a specific profile not necessarily in the AMS.

Live MQTT reads are useful only in a narrow case: the user swaps AMS spools while Bambu Studio is closed, *and* they're willing to enable LAN-only mode (Bambu firmware after ~May 2024 only allows local MQTT in LAN-only mode, which disables Bambu Cloud / Handy / MakerWorld direct-print).

If credentials are configured (env vars and/or a CLAUDE.md block), the tool tries MQTT first and falls back to the cache transparently when it can't connect — no error, just a one-line stderr note. Recognised env vars (env overrides CLAUDE.md):

```sh
export BAMBU_ACCESS_CODE=12345678          # ALWAYS recommend env, never CLAUDE.md
export BAMBU_SERIAL=00M09C460402058        # optional; may also be in CLAUDE.md
export BAMBU_HOST=192.168.1.42             # optional; may also be in CLAUDE.md
```

```markdown
## Bambu printer (MQTT)

- Serial: 00M09C460402058
- Host: 192.168.1.42
```

**The access code is a secret — never tell the user to put it in CLAUDE.md** (CLAUDE.md is loaded into prompt context). Always recommend `BAMBU_ACCESS_CODE`. Serial and host are fine in CLAUDE.md if the user wants Claude to know which printer is in use.

The success summary reports `AMS state: live (MQTT, <host>)` when MQTT succeeds, otherwise `cached (BambuStudio.conf)`.

**Limitations (v1):** single object only; single plate; the canonical template's transform is overwritten on every splice (parts always land centred on plate, on z=0); MQTT requires `Host` (auto-discovery via mDNS may come later).

### `openscad-build-template`

Convert a slicer-saved project `.3mf` into a canonical `openscad-pack-3mf` template — applies the skill's default overrides for layer height / wall count / shell layers / infill / brim, ensures Bambu Studio honours them on open (by adding the keys to `different_settings_to_system`), strips placement transforms, and drops stale thumbnails.

```sh
~/.claude/skills/openscad-bosl2/tools/openscad-build-template /tmp/my-template.3mf
~/.claude/skills/openscad-bosl2/tools/openscad-build-template --slicer bambu-studio /tmp/my.3mf
~/.claude/skills/openscad-bosl2/tools/openscad-build-template -o ~/foo.3mf /tmp/my.3mf
```

Run this any time the SKILL.md slicer-parameter defaults change, or whenever you save a fresh slicer project you want to use as the basis. Currently only `bambu-studio` is implemented; PrusaSlicer / OrcaSlicer / Cura would each need their own override-key set added to the tool.

## Reference sub-docs

Read the relevant sub-doc before reaching for a primitive you don't know:

- [`references/bosl2-shapes.md`](references/bosl2-shapes.md) — 2D and 3D primitives (`square`, `rect`, `circle`, `ellipse`, polygons, stars, `teardrop2d`, `cuboid`, `cyl`, `sphere`, `spheroid`), the anchor/spin/orient system, basic attachment.
- [`references/bosl2-transforms.md`](references/bosl2-transforms.md) — movement helpers, rotation (`rot`/`xrot`/`from=,to=`), distributors (`xcopies`, `grid_copies`, `arc_copies`, `mirror_copy`), mutators (`top_half`, `chain_hull`, `round2d`, `shell2d`, `minkowski_difference`).
- [`references/bosl2-rounding.md`](references/bosl2-rounding.md) — built-in rounding/chamfer, `rounded_prism()`, `edge_profile()` + `mask2d_*`, 3D edge/corner masks. **Three-tier model**: pick the simplest tool that does the job.
- [`references/customizer.md`](references/customizer.md) — full Customizer annotation reference: dropdowns, sliders, spinboxes, text, vectors, tabs, presets, `-p` / `-P` CLI.
- [`references/printers.md`](references/printers.md) — common FDM printer specs (Bambu A1/P1S/X1C/X1E/H2D, Prusa MK4S/Core One/XL, Creality K1/K1 Max/K2 Plus): build volume, max temps, enclosure, multi-material support. Use when the user mentions a printer or the project's `CLAUDE.md` doesn't capture limits.
- [`references/bosl2-tutorials.md`](references/bosl2-tutorials.md) — index of all BOSL2 tutorials (most linked, not preloaded — fetch on demand). Also points at the BOSL2 cheatsheet for full API lookup.

## Common gotchas

- **`cube()` defaults to `center=false`, `cuboid()` defaults to centred.** Pick one style per file.
- **Forgetting `epsilon` in `difference()`** causes z-fighting in the preview and (sometimes) holes in the render.
- **Variables after the first `{` are invisible to the Customizer.** Keep all customizable params at the top.
- **`edge_profile()` and `mask2d_*` need `diff()`** wrapping the parent to actually subtract.
- **Quadrant order for per-corner rounding** is **counter-clockwise from back-right** (I, II, III, IV). Test with distinct values if unsure.
- **Mixing `rounding=` and `chamfer=` on the same edge** is not allowed. Pick one per edge.

## Skeleton

```scad
include <BOSL2/std.scad>

/* [Geometry] */
// Outer width
width = 60;             // [20:5:200]
// Outer depth
depth = 40;             // [20:5:200]
// Wall thickness
wall = 3;               // [1:0.5:6]

/* [Hidden] */
$fn = 256;
epsilon = 0.01;
layer_height = 0.2;     // chosen for general utility part; not a Customizer param

BASE = [width, depth, 4 * layer_height];

echo(str("Base: ", BASE.x, " × ", BASE.y, " × ", BASE.z, " mm"));

// One object for the slicer
union() {
    difference() {
        cuboid(BASE, anchor=BOTTOM, rounding=2, edges="Z");
        // ... cutouts, with epsilon offsets ...
    }
    // ... additive features ...
}
```
