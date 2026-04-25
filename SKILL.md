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

> *"Suggested slicer settings: 0.2 mm layer height, PETG, 3 walls, 20 % grid infill, 4 top/bottom layers, no supports needed in this orientation, brim recommended (small footprint)."*

## Tools

This skill ships helper scripts. **Always invoke them by their full path inside the skill** — the working directory is rarely the skill repo, so `tools/openscad-check` will fail in most projects. Use:

```sh
~/.claude/skills/openscad-bosl2/tools/openscad-check FILE.scad
~/.claude/skills/openscad-bosl2/tools/openscad-render FILE.scad
```

(That path works because the skill is installed user-level via symlink. If a particular project installs the skill at `<project>/.claude/skills/openscad-bosl2/` instead, adjust accordingly.)

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

Both scripts honour `$OPENSCAD` to override the binary path; otherwise they search `PATH` and standard install locations including `/Applications/OpenSCAD.app/Contents/MacOS/OpenSCAD`.

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
