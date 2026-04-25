# BOSL2 tutorial index

Links to every official tutorial in the [BOSL2 repo](https://github.com/BelfrySCAD/BOSL2/tree/master/tutorials). Items marked **(summarised locally)** have a sub-doc in this folder; everything else is "fetch on demand."

## Summarised locally

- **Shapes2d / Shapes3d** → [`bosl2-shapes.md`](bosl2-shapes.md) — `square`, `rect`, `circle`, `ellipse`, polygons, stars, `teardrop2d`, `cuboid`, `cyl`, `sphere`, `spheroid`, anchor/spin/orient, basic attachment.
- **Transforms / Distributors / Mutators** → [`bosl2-transforms.md`](bosl2-transforms.md) — `right`/`up`/`back`, `rot`, `xcopies`/`grid_copies`/`arc_copies`, `top_half`/`half_of`, `chain_hull`, `round2d`/`shell2d`, `minkowski_difference`.
- **Rounding the Cube** → [`bosl2-rounding.md`](bosl2-rounding.md) — built-in rounding/chamfer, `rounded_prism()`, `edge_profile()` + `mask2d_*`, 3D `edge_mask()` / `corner_mask()`, `diff()` integration.

## Fetch on demand

These are less commonly needed for the kind of household / workshop parts in the user's corpus. Fetch the raw markdown if a job calls for them — don't preload.

- **Attachment-Overview** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Overview.md
- **Attachment-Basic-Positioning** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Basic-Positioning.md
- **Attachment-Position** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Position.md
- **Attachment-Attach** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Attach.md
- **Attachment-Align** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Align.md
- **Attachment-Relative-Positioning** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Relative-Positioning.md
- **Attachment-Tags** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Tags.md (tag-based diff/intersect — relevant if using `diff()` heavily)
- **Attachment-Color** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Color.md
- **Attachment-Parts** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Parts.md
- **Attachment-Edge-Profiling** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Edge-Profiling.md
- **Attachment-Making** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Attachment-Making.md (writing your own attachable modules)
- **Paths** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Paths.md (paths and regions — needed for `offset_sweep`, `path_sweep`, `linear_sweep` workflows)
- **Beziers_for_Beginners** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Beziers_for_Beginners.md
- **VNF** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/VNF.md (vertex/face polyhedra — advanced)
- **FractalTree** — https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/FractalTree.md (worked example, not reference)

## When to look further

The tutorials cover *concepts*. For complete API signatures (every parameter, every overload), the source-of-truth is the BOSL2 wiki cheatsheet:

- **Cheatsheet** — https://github.com/BelfrySCAD/BOSL2/wiki/CheatSheet (every public function, one-line signature)
- **Wiki home** — https://github.com/BelfrySCAD/BOSL2/wiki (per-module pages: `shapes3d.scad`, `rounding.scad`, `screws.scad`, `gears.scad`, etc.)

If a sub-doc here doesn't answer the question, hit the cheatsheet first.
