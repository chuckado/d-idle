## Why

The live drill D1 and the soak logs show the ground-potion pickup disagreeing with game.exe (docs/research/belt-pickup-trace.md, proven by the emulator): a clicked ground potion goes to the belt first (first column whose bottom potion is the same kind, else the first empty column), else the pack. The pickit assumed the pack (939 "potion already in the pack" skips with belt room free), treated any free belt box as room, and reported "picked 0" for a potion that did land on the belt.

## What Changes

- internal/town/belt: pure `Place` and `WillBelt` (the game's box choice), shared by the pickit and the belt fill.
- internal/loot/loot.go: a potion the plan carries is wanted when `WillBelt`; the pack buffer limits only pack-bound potions.
- internal/loot/pickup.go: belt room means `WillBelt`, and arrival is also proven by the same-class pack-plus-belt count going up (a unit id may not survive the pickup).
- internal/town/beltfill.go: `pickBeltBox` is `belt.Place` (the plan column no longer steers).
- internal/town/sim/sim.go: Add sets the item cell like the real reader.
- Delta: town-basics ADDED `Ground potion goes where the game puts it`.

code merged e4bb88d; owns released so later cards can edit; live proof pending.
