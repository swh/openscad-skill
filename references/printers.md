# Common 3D printer specs

A reference for the consumer / prosumer FDM printers most likely to be in use. Specs are drawn from manufacturer pages and the [Polymaker printer wiki](https://wiki.polymaker.com/the-basics/3d-printers/popular-printers); always cross-check against the user's actual machine before assuming a limit. **Specs change with firmware updates and new revisions** — treat this as a starting point, not gospel.

The user records their current printer in the project / parent `CLAUDE.md`. If that's missing or ambiguous, ask before designing around capacity, material, or temperature limits.

## Quick comparison

| Printer              | Build vol (mm)        | Enclosed | Chamber       | Nozzle °C | Bed °C | Multi-material | Stock hardened nozzle |
|----------------------|-----------------------|----------|---------------|-----------|--------|-----------------|------------------------|
| **Bambu A1 Mini**    | 180 × 180 × 180       | No       | —             | 300       | 80     | AMS Lite (4)    | No                     |
| **Bambu A1**         | 256 × 256 × 256       | No       | —             | 300       | 100    | AMS Lite (4)    | No                     |
| **Bambu P1P**        | 256 × 256 × 256       | No (panel kit avail.) | —  | 300       | 100    | AMS (4, ×4)     | No                     |
| **Bambu P1S**        | 256 × 256 × 256       | Yes      | Passive ~50°C | 300       | 100    | AMS (4, ×4)     | Yes                    |
| **Bambu X1C**        | 256 × 256 × 256       | Yes      | Passive ~50°C | 300       | 100    | AMS (4, ×4)     | Yes                    |
| **Bambu X1E**        | 350 × 320 × 325       | Yes      | Active 60°C   | 350       | 120    | AMS (4, ×4)     | Yes                    |
| **Bambu H2D**        | 350 × 320 × 325       | Yes      | Active 65°C   | 350       | 120    | Dual nozzle     | Yes (tungsten carbide) |
| **Prusa MK4S**       | 250 × 210 × 220       | No (enclosure kit avail.) | — | 290       | 120    | MMU3 (5)        | No                     |
| **Prusa Core One**   | 250 × 220 × 270       | Yes      | Active ~55°C  | 290       | 120    | MMU3 (5)        | No                     |
| **Prusa XL**         | 360 × 360 × 360       | Optional | Optional      | 300       | 120    | Up to 5 toolheads | No                   |
| **Creality K1**      | 220 × 220 × 250       | Yes      | Passive       | 300       | 100    | Optional CFS    | No                     |
| **Creality K1 Max**  | 300 × 300 × 300       | Yes      | Passive       | 300       | 120    | Optional CFS    | No                     |
| **Creality K2 Plus** | 350 × 350 × 350       | Yes      | Active ~60°C  | 350       | 120    | CFS integrated  | Yes                    |

"AMS (4, ×4)" = one AMS holds 4 spools, up to 4 AMS units = 16 colours. "AMS Lite (4)" = a single 4-spool unit, no chaining.

## Multi-material effective build volume

On almost every "single-nozzle multi-material" system (Bambu AMS, Prusa MMU3, Creality CFS), you don't lose printable area to the multi-material kit itself — but **the slicer adds a purge tower** beside the part to flush colour transitions. That tower can occupy a large fraction of the bed for prints with many colour changes. Plan layouts so the actual part fits comfortably alongside it; for a 256×256 bed, **assume ≤ ~200×200 of usable area when printing multi-colour with frequent transitions**.

The **Prusa XL** is the exception — independent toolheads mean no purge tower and no effective volume penalty for multi-material.

The **Bambu H2D** uses two physical nozzles, also avoiding a purge tower for two-material prints.

## Material compatibility — rules of thumb

- **PLA, PETG, TPU, PVA**: every printer in the table.
- **PLA-CF, PETG-CF**: needs a hardened nozzle. Stock on P1S / X1C / X1E / H2D / K2 Plus; aftermarket upgrade on the others.
- **ABS, ASA**: technically prints on open-frame machines (A1, P1P, MK4S, K1) but warps badly without an enclosure. Use enclosed printers (P1S / X1C / Core One / K1 Max / etc.) in practice.
- **PA / PA-CF (Nylon)**: needs both a hardened nozzle *and* an actively heated chamber for best results. Reliable on X1E / Core One / H2D / K2 Plus; possible-but-finicky elsewhere.
- **PC (Polycarbonate)**: needs ~280–300 °C nozzle, 100–120 °C bed, *and* an enclosure (ideally heated). Realistic on X1C / X1E / Core One / K1 Max / K2 Plus / H2D.
- **PEEK / PEI / Ultem**: needs 350+ °C nozzle, ~120 °C bed, and an actively heated chamber (>60 °C). Only X1E, K2 Plus, and H2D approach this; even then, dedicated industrial machines are usually required for production-quality results.

## Per-printer notes

### Bambu A1 / A1 Mini
Open-frame i3-style. Cheapest way into Bambu's ecosystem. **No enclosure** means no engineering plastics (ABS/ASA/PC/Nylon). Stock brass nozzle — must upgrade for any composite filament. AMS Lite holds 4 spools; doesn't chain.

### Bambu P1S / P1P
Enclosed CoreXY. P1P is the open-frame version — Bambu sells side panels to convert it to a P1S. Both have hardened steel nozzles by default and accept regular AMS stacks.

### Bambu X1C
The flagship CoreXY for several years. Hardened nozzle, hardened extruder gears, lidar/camera, AMS support. Passive enclosure means no real PA/PC/PEEK without modifications.

### Bambu X1E
Enterprise X1 with **active chamber heating to 60 °C**, 350 °C hotend, 120 °C bed, ethernet, and lockdown firmware options. The serious engineering-materials Bambu.

### Bambu H2D
Dual-nozzle CoreXY (released 2025) — two physical hotends avoid purge towers for two-material prints. Tungsten-carbide nozzle option. Active chamber.

### Prusa MK4S
Bedslinger evolution of the MK series. Reliable, repairable, open ecosystem. Open frame; the official enclosure is a separate purchase. Multi-material via MMU3 (purge tower).

### Prusa Core One
Prusa's enclosed CoreXY. Active chamber to ~55 °C, 290 °C nozzle, 120 °C bed. Targets engineering materials in a Prusa-style ecosystem.

### Prusa XL
Toolchanger CoreXY. **Up to 5 independent toolheads** — true multi-material with no purge tower. Optional enclosure. Largest build volume in this table at 360 mm cube.

### Creality K1 / K1 Max / K2 Plus
Affordable enclosed CoreXYs. K1 is the smallest; K1 Max scales to 300 mm; K2 Plus adds active chamber, 350 °C hotend, and an integrated CFS (multi-material) at 350 mm cube. Ecosystem is more open than Bambu's but slicer/firmware support is less polished.

## What to record in your project's `CLAUDE.md`

When you switch or add a printer, capture these in the project / parent `CLAUDE.md` so the skill's recommendations stay accurate:

```markdown
## Printer

- **Model:** [e.g. Bambu X1C]
- **Build volume:** [X × Y × Z mm], multi-material effective: [X × Y mm]
- **Enclosed:** [yes/no — passive/active chamber °C if any]
- **Max nozzle / bed:** [°C / °C]
- **Hardened nozzle:** [yes/no/aftermarket]
- **Multi-material:** [AMS / MMU3 / CFS / toolchanger / none]
- **Default material:** [e.g. PETG]
- **Slicer:** [Bambu Studio / OrcaSlicer / PrusaSlicer / etc.]
```

That's enough for the skill to reason about whether a part will fit, whether the suggested filament is printable, and whether to assume an enclosure.

## Sources

- [Polymaker — Bambu Lab printers](https://wiki.polymaker.com/the-basics/3d-printers/popular-printers/bambu-lab)
- [Polymaker — Creality printers](https://wiki.polymaker.com/the-basics/3d-printers/popular-printers/creality)
- [Bambu Lab P1S spec sheet (PDF)](https://marketplace.createeducation.com/wp-content/uploads/2023/11/bambu-lab-P1S-tech-specs.pdf)
- [Bambu Lab A1 tech specs](https://bambulab.com/en/a1/tech-specs)
- [Original Prusa MK4S product page](https://www.prusa3d.com/product/original-prusa-mk4-2/)
- [Original Prusa Core One product page](https://www.prusa3d.com/product/prusa-core-one/)
- [Original Prusa XL product page](https://www.prusa3d.com/product/original-prusa-xl-semi-assembled-single-toolhead-3d-printer/)
- [Prusa forum: Maximum nozzle temperatures for XL, Core One, MK4/S](https://forum.prusa3d.com/forum/english-forum-general-discussion-announcements-and-releases/maximum-nozzle-temperatures-for-xl-core-one-mk4-s/)
- [Creality K1 Max vs K2 Plus vs K1C](https://crealitysg.com/blogs/news/k1-max-vs-k2-plus)
