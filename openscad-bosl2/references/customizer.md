# OpenSCAD Customizer reference

Concise summary of the [OpenSCAD Customizer](https://en.wikibooks.org/wiki/OpenSCAD_User_Manual/Customizer) syntax. The Customizer is a GUI panel that turns top-of-file variables into form controls. Authoring `.scad` files with proper Customizer annotations means users can tweak parameters without editing source.

## Visibility rules

A variable shows up in the Customizer **only if all** of these hold:

- It's assigned in the main file (not via `include` / `use`).
- The value is a literal: number, string, boolean, or numeric vector of length ≤ 4.
- It's not inside a `/* [Hidden] */` section.
- It appears before the first `{` in the file.

Expressions like `e = str("a", "b")` or `f = 12 + 0.5` are **not** customizable.

To stop the Customizer from picking up further variables, define a no-op module:

```scad
module __Customizer_Limit__() {}
debug_mode = false;  // won't be shown in the Customizer
```

## Description / label

A single-line comment **directly above** a variable (no blank line between) becomes that variable's label:

```scad
// Wall thickness in mm
wall = 4;
```

## Tabs / sections

Block comments group variables into tabs:

```scad
/* [Geometry] */
width = 30;
height = 20;

/* [Print settings] */
layer_height = 0.2;
```

Special tab names:

- `/* [Global] */` — parameters show on **every** tab.
- `/* [Hidden] */` — parameters are **never** shown. Use this for internal constants. Can appear multiple times in a file.
- Variables before any tab marker live in a default tab named `parameters`.

## Control types and exact syntax

The control type is determined by the variable's value type plus the trailing `// [...]` annotation.

### Drop-down (combo box)

Provide a list of allowed values in `[ ]`. Works for numbers and strings, with optional `value:label` pairs.

```scad
Numbers       = 2;     // [0, 1, 2, 3]
Strings       = "foo"; // [foo, bar, baz]
LabeledNums   = 10;    // [10:S, 20:M, 30:L]
LabeledStrs   = "S";   // [S:Small, M:Medium, L:Large]
```

### Slider (numeric)

```scad
maxOnly   = 34;  // [50]            -> 0..50
range     = 34;  // [10:100]        -> 10..100
withStep  = 2;   // [0:5:100]       -> 0..100, step 5
centered  = 0;   // [-10:0.1:10]    -> negative ranges OK
```

Format: `[max]`, `[min:max]`, or `[min:step:max]`.

### Spinbox (numeric, no range)

A bare numeric variable with no annotation:

```scad
Spinbox = 5;        // unit step
SpinboxF = 5.5;     // step inferred from decimals (here 0.5)
```

You can also explicitly hint the step:

```scad
Spinbox = 5; // .5
```

### Checkbox

A boolean with an optional description above it:

```scad
// Enable preview helpers
debug = true;
```

### Text field

Bare string:

```scad
name = "hello";
```

With a length hint (display width, not a hard limit):

```scad
name = "length"; // 8
```

Multi-line text fields are **not** supported.

### Vector (2–4 elements)

Numeric vectors render as a row of spinboxes. Annotations apply to every element.

```scad
v2 = [12, 34];
v3 = [12, 34, 45];
v4 = [12, 34, 45, 23];

vRange3 = [12, 34, 46];     // [1:2:50]
vRange4 = [12, 34, 45, 23]; // [1:50]
```

## Presets / JSON

The Customizer can save sets of values as named **presets**. Stored as JSON:

```json
{
  "parameterSets": {
    "Small":  { "wall": "3", "size": "10" },
    "Large":  { "wall": "5", "size": "30" }
  },
  "fileFormatVersion": "1"
}
```

Command line:

```sh
openscad -o out.stl -p params.json -P Small input.scad
```

- `-p file.json` — preset file.
- `-P NameOfSet` — which preset to use.
- Loading a preset only overwrites the variables it mentions; others keep their current values (so partial preset files are useful).
- Variables in `[Hidden]` sections **are saved** to the JSON but **not loaded back** — handy for reserving names for future Customizer fields without breaking old presets.

### `-D` interaction

`-D var=value` **cannot** override a parameter that's set by a loaded preset. The workaround is an indirection: expose `cpart` to the Customizer, copy it into `part` outside any tab, and override `part` with `-D`:

```scad
/* [Part] */
// Which part to render
cpart = 1; // [1:Body, 2:Lid, 3:Insert]

/* [Hidden] */
part = cpart;
```

Then `openscad -p p.json -P Default -D part=2 ...` works.

## Gotchas

- **Range/step inference is unreliable.** Always specify ranges explicitly via `// [min:max]` or `// [min:step:max]` rather than relying on the Customizer to guess.
- **Variables after the first `{` are invisible.** Keep all customizable parameters at the top of the file.
- **No expressions.** A line like `radius = diameter / 2;` won't appear — assign literals to Customizer-visible variables, then derive in `[Hidden]` or below the limit module.
- **Checkboxes aren't supported by Thingiverse's Customizer.** Plain OpenSCAD handles them fine.
- **The textbox feature** is documented as working in 2021.01 with no guarantee for future versions.

## Skeleton template

```scad
include <BOSL2/std.scad>

/* [Geometry] */
// Outer width in mm
width = 60;        // [20:5:200]
// Outer depth in mm
depth = 40;        // [20:5:200]
// Wall thickness
wall = 3;          // [1:0.5:6]

/* [Print] */
// Layer height (z-axis features will snap to multiples of this)
layer_height = 0.2; // [0.1, 0.15, 0.2, 0.25, 0.3]

/* [Hidden] */
$fn = 128;
epsilon = 0.01;

// ... geometry below ...
```
