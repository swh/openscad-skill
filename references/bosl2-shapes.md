# BOSL2 shape primitives

Summary of the [Shapes2d](https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Shapes2d.md) and [Shapes3d](https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Shapes3d.md) tutorials. These tutorials cover **basic primitives only** — rounding/filleting beyond corner-rounding, attachable parts, paths, regions, sweeps, and screws live in separate tutorials.

The single biggest BOSL2 idea: every primitive accepts `anchor=`, `spin=`, and (for 3D) `orient=`, applied in that order. Once you internalise that, most of the API follows.

## Anchor / spin / orient

Direction constants (vectors). The same constants work in 2D and 3D:

| Constant            | Vector       |
|---------------------|--------------|
| `LEFT`              | `[-1, 0, 0]` |
| `RIGHT`             | `[ 1, 0, 0]` |
| `FRONT` / `FWD`     | `[ 0,-1, 0]` |
| `BACK`              | `[ 0, 1, 0]` |
| `BOTTOM` / `BOT` / `DOWN` | `[ 0, 0,-1]` |
| `TOP` / `UP`        | `[ 0, 0, 1]` |
| `CENTER` / `CTR`    | `[ 0, 0, 0]` |

**Combine with `+`** to get edges and corners: `TOP+RIGHT`, `BOTTOM+FRONT+LEFT`, etc.

- `anchor=`  — which point on the shape lands at the origin. The vector also points *outward* from the shape at that face/edge/corner.
- `spin=`    — clockwise rotation in degrees, applied **after** anchoring. In 2D this is rotation in the XY plane; in 3D it's around Z.
- `orient=`  — (3D only) vector that the shape's local "up" should align with. Applied after spin.

Per-corner lists (for `rounding=` / `chamfer=` on rectangles) are indexed **[quadrant I, II, III, IV]** — i.e. counter-clockwise starting from back-right (`+X+Y`).

For 3D `cuboid()`, `edges=` and `except_edges=` accept:

- A face vector (`TOP`) — all 4 edges of that face.
- A corner vector (`TOP+RIGHT+FRONT`) — the 3 edges meeting that corner.
- A single edge vector (`TOP+FRONT`).
- An axis string: `"X"`, `"Y"`, or `"Z"` — all 4 edges parallel to that axis.
- A **list** combining any of the above: `edges=[TOP, "Z", BOTTOM+RIGHT]`.

Combine `edges=` and `except_edges=` to round most of the shape and skip a few.

## 2D primitives

### `square(size, anchor=[0,0], spin=0, center=true)`

BOSL2's override of the built-in. Adds `anchor`/`spin` (the `center=` flag stays for compatibility but `anchor=CENTER` is preferred).

### `rect([w, h], anchor=[0,0], spin=0, rounding=0, chamfer=0)`

Rectangle with optional **per-corner** rounding or chamfering. `rounding=` and `chamfer=` accept either a scalar (all four corners) or a 4-list `[I, II, III, IV]` with `0` to leave a corner sharp. They can be mixed (round one corner, chamfer another).

### `circle(r=, d=, anchor=, spin=, $fn=)`

Override of the built-in. Anchor points are anywhere on the circumference — pass any direction vector and you get the corresponding perimeter point.

### `ellipse(r=[rx,ry], d=[dx,dy], anchor=, spin=, realign=false, circum=false, $fn=)`

Independent X/Y radii. `realign=true` rotates by half a side angle (useful so a flat is on top instead of a vertex). `circum=true` circumscribes the ideal ellipse instead of inscribing it.

### `right_triangle([w, h], anchor=, spin=)`

Right triangle in quadrant I. Use `xflip()`, `yflip()`, or `spin=` to reorient.

### `trapezoid(w1=, w2=, h=, shift=0, anchor=, spin=)`

Generalised quadrilateral:
- `w1=0` → triangle with apex at back.
- `w2=0` → triangle with apex at front.
- `shift=` slides the back edge along X to make a parallelogram.

Anchor directions follow the side angles, which can be surprising — visualise with `show_anchors()` if unsure.

### `pentagon()`, `hexagon()`, `octagon()`, `regular_ngon(n=, ...)`

Signature (same for all four):
```
pentagon(d=, r=, side=, ir=, id=, rounding=0, realign=false, anchor=, spin=)
```
- `r`/`d` — circumradius / circumdiameter (vertex distance).
- `ir`/`id` — inradius / indiameter (flat-to-flat).
- `side=` — specify by edge length.
- `rounding=` — curve the corners.
- `realign=true` — rotate by half a side so a flat sits on top.
- Named anchors: `"side0"`, `"tip0"`, etc.

Prefer these over `circle($fn=n)` — the anchor points sit on the polygon's vertices/edges, not on the inscribed circle.

### `star(n=, ir=, or=, id=, d=, step=1, realign=false, align_tip=, align_pit=, anchor=, spin=)`

- `n=` number of points; `ir`/`or` (or `id`/`d`) inner and outer radii.
- `step=` connects every Nth vertex — produces different star topologies.
- `align_tip=vector` / `align_pit=vector` orient the first tip or pit toward a direction.
- Named anchors: `"tip0"`, `"pit0"`, `"midpoint0"`.

### `teardrop2d(r=, d=, ang=45, cap_h=, anchor=, spin=)`

3D-printable hole profile — flat-top circle that bridges without supports. `ang=` is the overhang angle; `cap_h=` flattens the top to encourage bridging.

### `glued_circles(r=, d=, spread=, tangent=, anchor=, spin=)`

Two circles joined by a meniscus. `spread=` is centre-to-centre distance; `tangent=` is the meniscus angle (negative values flip the curve).

## 3D primitives

### `cube(size, center=false, anchor=, spin=, orient=)`

BOSL2 override of the built-in. Adds full anchor/spin/orient support; the original `center=` flag is retained.

### `cuboid(size, anchor=CENTER, spin=0, orient=UP, rounding=0, chamfer=0, edges=EDGES_ALL, except_edges=[])`

The workhorse rectangular solid. **Centres by default** (unlike `cube()`). Adds:
- `rounding=` / `chamfer=` — scalar applied to selected edges.
- `edges=` / `except_edges=` — see selectors above.
- `rounding` and `chamfer` cannot both be non-zero on the same edge.

Use `cuboid()` over `cube()` whenever any edge needs rounding, chamfering, or non-default anchoring.

### `cylinder(h=, r=, r1=, r2=, d=, d1=, d2=, center=false, anchor=, spin=, orient=)`

BOSL2 override of the built-in. Same parameters; adds anchor/spin/orient.

### `cyl(h=, l=, r=, d=, r1=, r2=, d1=, d2=, rounding=, rounding1=, rounding2=, chamfer=, chamfer1=, chamfer2=, anchor=CENTER, spin=0, orient=UP)`

Workhorse cylinder. **Centres by default**. Adds:
- `rounding=` / `chamfer=` for both ends.
- `rounding1`/`rounding2` and `chamfer1`/`chamfer2` for **per-end** control.
- A negative chamfer/rounding goes outward (good for fillets at the base of a feature).
- Mix freely: chamfer one end, round the other.

Both `h` and `l` are accepted as the length parameter.

Use `cyl()` over `cylinder()` whenever an end needs rounding, chamfering, or any anchor other than the default.

### `sphere(r=, d=, anchor=, spin=, orient=)`

Built-in override with anchor/spin/orient.

### `spheroid(r=, d=, circum=false, style="orig", anchor=, spin=, orient=)`

Improved sphere. `circum=true` circumscribes the ideal sphere (rather than inscribing — useful when the sphere is a cutter and you don't want it to leave gaps). `style=` controls vertex/triangulation:

- `"orig"` (default) — OpenSCAD's standard layout.
- `"aligned"` — poles aligned with axes.
- `"stagger"` — staggered rings, fewer triangles at the poles.
- `"icosa"` — icosahedral subdivision (most uniform).
- `"octa"` — octahedral subdivision.

## Attachment system

These are children of a primitive and operate against the **parent's** named anchors / direction vectors.

### `position(anchor)`
Translates the child to the parent's anchor point. No rotation.
```scad
cuboid([20,20,5])
    position(TOP+RIGHT) cyl(d=4, h=8);
```

### `orient(anchor)`
Rotates the child so its local "up" points along the parent's anchor direction. No translation.

### `attach(parent_anchor, child_anchor=FWD)`
The combination. With one argument, attaches the child at `parent_anchor` (child's `BOTTOM` faces the parent's anchor face by default — i.e. the child sits *on top of* the parent's face). With two arguments, the named child anchor mates with the parent's anchor.

```scad
cuboid([30,30,10])
    attach(TOP) sphere(d=10);          // sphere sits on top
cuboid([30,30,10])
    attach(RIGHT, BOTTOM) cyl(d=4, h=8); // cylinder sticks out the right
```

The child's anchor outward direction points away from the surface it's attached to.

## Idioms / gotchas

- **Default centering differs.** `cube()` and `cylinder()` default to `center=false`; `cuboid()` and `cyl()` default to centred. Mixing them in one file without checking is a common source of bugs — pick one style per file and prefer the BOSL2 versions with explicit `anchor=`.
- **Anchor first, transform second.** Setting `anchor=BOTTOM` on a primitive is almost always cleaner than wrapping it in `down(h/2)` or `translate([0,0,h/2])`.
- **Use named polygon helpers.** `hexagon(d=10)` is clearer than `circle(d=10, $fn=6)` and gives anchor points on vertices/edges instead of on the inscribed circle.
- **Per-corner / per-edge rounding lists.** Quadrant order is **counter-clockwise from back-right** (I, II, III, IV). Don't guess — test with a 4-list of distinct values and look at the render.
- **`teardrop2d()` for 3D-printable holes.** Round holes that overhang need bridging; the teardrop profile prints cleanly without supports.
- **`spheroid(style="icosa")` for cutters.** When using a sphere in `minkowski()` or as a cutter, the `icosa` style avoids the polar-pinch artefacts of the default style.
