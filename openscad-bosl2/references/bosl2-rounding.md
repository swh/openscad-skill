# BOSL2 rounding, chamfering, edge profiles

Summary of the [Rounding the Cube](https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Rounding_the_Cube.md) tutorial. Despite the name, it covers all of BOSL2's edge/corner-treatment tooling — most of it works on prismoids and other shapes too.

There are three escalating tiers; pick the simplest one that does the job.

## Tier 1: built-in rounding/chamfer on `cuboid()` / `prismoid()`

Most cases. Uniform radius across selected edges.

```
cuboid(size, rounding=r, chamfer=c, edges=SPEC, except_edges=SPEC, anchor=, spin=, orient=)

prismoid(size1, size2, h, rounding=r, rounding1=r, rounding2=r, chamfer=c, ...)
```

Key points:
- `rounding=` and `chamfer=` cannot both be non-zero on the same edge. Use one or the other per edge (you can mix across edges by calling separately or via masks).
- `cuboid()` accepts **negative** rounding/chamfer on the top or bottom face — turns the cutout into an outward fillet (good for the base of a feature meeting a floor).
- `prismoid()` rounds vertical edges only; `rounding1=` / `rounding2=` for top/bottom rounding individually. Per-edge: `rounding1=[r0, r1, r2, r3]` — counter-clockwise from `BACK+RIGHT`.

### Edge selectors (for `edges=` / `except_edges=`)

| Form                      | Selects                                              |
|---------------------------|------------------------------------------------------|
| `"X"` / `"Y"` / `"Z"`     | All 4 edges parallel to that axis                    |
| `TOP`                     | All 4 edges around that face                         |
| `TOP+RIGHT`               | The single edge between those two faces              |
| `TOP+RIGHT+FRONT`         | The 3 edges meeting that corner                      |
| `[TOP, "Z", BOTTOM+RIGHT]`| List of any of the above                             |
| 3×4 binary array          | Explicit on/off per edge — escape hatch              |

`except_edges=` subtracts from `edges=` — start with `edges=EDGES_ALL` and chip away.

## Tier 2: `rounded_prism()` — continuous-curvature rounding

When constant-radius arcs look unnatural (visible "tangent kinks"), use `rounded_prism()`. It rounds with a 4th-order Bezier instead of an arc, producing G2 continuity.

```
rounded_prism(polygon, height=,
              joint_top=, joint_bot=, joint_sides=,
              k=0.5, splinesteps=)
```

- `joint_*` is the *distance* from the start of the rounding to the would-be sharp edge — not a radius. Larger joint = larger sweep.
- Pass a 2-vector (`joint_top=[a, b]`) for asymmetric rounding (different sweep lengths into adjacent faces).
- `k` ∈ [0, 1] controls how abruptly the curve transitions; default 0.5. Higher = tighter against the corner.
- Per-side variation: pass a list to `joint_sides=`.

Use this for visible / tactile features (handles, grips, organic-feeling parts). Skip it for hidden internal geometry.

## Tier 3: profile and mask attachments

When a single shape needs *different* treatments on different edges (e.g. round the top edges, chamfer the bottom, leave one face alone, teardrop a printed undercut), compose with attachments.

### 2D profile attachments

`edge_profile()` attaches a 2D profile to selected edges of an attachable parent. The profile is tagged `"remove"` by default, so it's subtracted via BOSL2's `diff()` system.

```
edge_profile(faces, except=[], excess=0)
edge_profile_asym(face, flip=, corner_type="round")    // for asymmetric fillets
corner_profile(corners, r=)
face_profile(faces, r=)
```

Pair with one of the **mask2d_*** profiles:

| Module                                                        | Shape                       |
|---------------------------------------------------------------|-----------------------------|
| `mask2d_roundover(r=, h=, inset=, mask_angle=, quarter_round=)`| Concave quarter-round       |
| `mask2d_teardrop(h=, angle=, mask_angle=)`                    | Printable undercut          |
| `mask2d_cove(...)`                                             | Concave cove                |
| `mask2d_chamfer(...)`                                          | 45° (or angled) chamfer     |
| `mask2d_rabbet(...)`                                           | Rectangular rabbet          |
| `mask2d_dovetail(...)`                                         | Dovetail joint              |
| `mask2d_ogee(...)`                                             | S-curve ogee                |

```scad
diff()
cuboid([40, 40, 20])
    edge_profile([TOP+LEFT, TOP+RIGHT])
        mask2d_roundover(r=4);
```

### 3D edge / corner masks

When a 2D profile isn't enough — e.g. teardrop edges that aren't axis-aligned — use 3D masks attached via `edge_mask()` / `corner_mask()`. The parent's `$parent_size` is available so the mask knows how long to be.

```
edge_mask(edge_spec)
corner_mask(corner_spec)

rounding_edge_mask(r=, r1=, r2=, l=, h=)
rounding_corner_mask(r=)
teardrop_edge_mask(r=, l=, angle=)
teardrop_corner_mask(r=, angle=)
chamfer_edge_mask(...)
chamfer_corner_mask(...)
```

## Choosing a tier

| Situation                                         | Tool                                |
|--------------------------------------------------|-------------------------------------|
| Uniform rounded edges on a box                    | `cuboid(rounding=)`                 |
| All edges except one face                         | `cuboid(rounding=, except_edges=)`  |
| Different radius top vs bottom                    | `prismoid` or `edge_profile` masks  |
| Visible/tactile fillet that should *flow*         | `rounded_prism()`                   |
| Printable hole / overhang on an arbitrary edge    | `teardrop_edge_mask()`              |
| Mixed treatments on the same shape                | `edge_profile()` + `mask2d_*`       |
| Custom profile (ogee, rabbet, dovetail)           | `edge_profile()` + `mask2d_*`       |

## Integration with `diff()`

Profile and mask attachments default to tag `"remove"`. They become real subtractions only when the parent (or an ancestor) is wrapped in `diff()`:

```scad
diff()
cuboid([40, 40, 20])
    edge_profile(TOP) mask2d_roundover(r=3);
```

Forgetting `diff()` is a common cause of "the rounding does nothing" — the mask is there but invisible.
