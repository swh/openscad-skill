include <BOSL2/std.scad>

// template-seed.scad
// Tiny placeholder geometry used to bootstrap a slicer template `.3mf`.
//
// Workflow:
//   1. openscad -o /tmp/seed.3mf template-seed.scad
//   2. Open /tmp/seed.3mf in your slicer.
//   3. Pick your printer / filament / process / AMS / supports / etc.
//   4. File -> Save Project As... -> /tmp/my-template.3mf
//   5. openscad-build-template /tmp/my-template.3mf
//
// Step 5 applies the skill's default overrides and writes the canonical
// template to ~/.config/openscad-bosl2/templates/bambu-studio.3mf;
// openscad-pack-3mf then splices new OpenSCAD geometry into it without
// touching the settings.

/* [Hidden] */
$fn = 64;
epsilon = 0.01;
layer_height = 0.2;

BASE = [40, 25, 4];

// One object for the slicer
union() {
    difference() {
        cuboid(BASE, anchor=BOTTOM, chamfer=0.5);

        // Embossed label so a slicer preview makes it obvious this is the template
        up(BASE.z - 2 * layer_height) linear_extrude(2 * layer_height + epsilon)
            text("TEMPLATE", size=5, halign="center", valign="center");
    }
}
