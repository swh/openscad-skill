# BOSL2 transforms, distributors, mutators

Summary of the [Transforms](https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Transforms.md), [Distributors](https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Distributors.md), and [Mutators](https://github.com/BelfrySCAD/BOSL2/blob/master/tutorials/Mutators.md) tutorials.

## Movement (translate)

`translate([x,y,z])` and its longer cousins are rarely the right call in BOSL2 code — single-axis helpers are clearer:

| Helper          | Effect       |
|-----------------|--------------|
| `right(d)`      | +X           |
| `left(d)`       | −X           |
| `back(d)`       | +Y           |
| `fwd(d)`        | −Y           |
| `up(d)`         | +Z           |
| `down(d)`       | −Z           |
| `move([x,y,z])` | same as `translate` |

Use `right(5)` not `translate([5,0,0])`. Chain them for compound moves: `up(10) right(5) cuboid(...)`.

## Rotation

`rot(...)` is the BOSL2 wrapper around `rotate()`:

```
rot([x, y, z])
rot(a=degrees, v=axis_vector)
rot(from=vector, to=vector)        // rotate one direction onto another
```

Single-axis helpers (all take optional `cp=[x,y,z]` centerpoint):

```
xrot(degrees, cp=)
yrot(degrees, cp=)
zrot(degrees, cp=)
```

`rot(from=, to=)` is the cleanest way to align a part with an arbitrary direction without computing Euler angles.

## Scaling

```
scale(factor)              // uniform or [x,y,z]
xscale(factor)
yscale(factor)
zscale(factor)
```

## Mirroring

```
mirror([nx, ny, nz])       // raw OpenSCAD, by plane normal
xflip(x=offset)            // mirror across YZ plane (optionally offset)
yflip(y=offset)            // mirror across XZ plane
zflip(z=offset)            // mirror across XY plane
```

## Skew

```
skew(sxy=, sxz=, syx=, syz=, szx=, szy=)
```

Each parameter is a multiplier: `sxy=0.5` shifts X by 0.5 × Y, etc.

---

## Distributors

Distributors take their child block and replicate it. The `n=` parameter controls how many copies; `spacing=` controls separation.

### Linear

```
xcopies(spacing=, n=2, sp=[0,0,0])
ycopies(spacing=, n=2, sp=[0,0,0])
zcopies(spacing=, n=2, sp=[0,0,0])
line_copies(spacing=, n=, p1=, p2=)
```

`sp=` is the start point; without it, copies are centred on the origin.

### Grid

```
grid_copies(spacing=, n=, size=, stagger=false, inside=)
```

`stagger=true` produces a hex-packed grid. `inside=polygon` clips to a 2D region.

### Rotational / mirror

```
zrot_copies(n=)            // n copies around Z
xrot_copies(n=) / yrot_copies(n=) / rot_copies(...)
arc_copies(n=, r=, sa=, ea=)
xflip_copy() / yflip_copy() / zflip_copy()
mirror_copy(v=)
move_copies([positions])
```

The mirror/flip variants emit the original **and** the mirrored child (i.e. one in, two out).

`arc_copies()` distributes along an arc — useful for radial spokes / bolt circles.

---

## Mutators

Mutators wrap children and modify the resulting shape.

### Half-space cuts (3D)

Each takes an optional axis offset and a "size" parameter that sets how big the implicit cutter cube is (default `s=1000`):

```
top_half(z=0, s=1000)
bottom_half(z=0, s=1000)
left_half(x=0, s=1000)
right_half(x=0, s=1000)
front_half(y=0, s=1000)
back_half(y=0, s=1000)
half_of(plane_normal_vector, cp=[0,0,0], s=1000)
```

The 2D versions take `planar=true`:

```
left_half(planar=true, x=0)
right_half(planar=true, x=0)
front_half(planar=true, y=0)
back_half(planar=true, y=0)
```

Use these to cut a model in half for printing, or to expose a cross-section in preview.

### Hull operations

```
chain_hull() { a; b; c; ... }
```

Hulls each consecutive pair (`a∪b`, `b∪c`, ...) and unions the results. Cheaper and often more useful than a full N-way `hull()`.

### Offset / shells (2D)

```
round2d(r=, or=, ir=)
shell2d(thickness, or=, ir=)
```

`round2d()` rounds inside and/or outside corners of a 2D shape; `shell2d()` produces an outline of a given thickness (positive grows outward, negative inward).

### Inverse Minkowski

```
minkowski_difference(planar=false) { base; tool; }
```

Subtracts the tool's surface profile from the base — useful for "shrink by N mm" operations that `offset()` can't do in 3D.

### Color helpers

```
hsl(h, s, l)
hsv(h, s, v)
```

For preview-only colouring of children.
