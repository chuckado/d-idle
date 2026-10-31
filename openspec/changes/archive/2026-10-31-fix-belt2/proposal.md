## Why

Post-fix soak 1 (Sorceress, 190 fills): 47 fills moved 0 potions, 42 of them `no belt column for Minor Mana Potion (kind mp)`, after which the belt was marked full. The cause is not the plan (a Sorceress has two mana columns) and not Place: ground healing potions (wanted whenever the game would belt them, with no count limit) opened the empty mana columns, so an 8-box belt held 5 to 6 healing potions against a plan of 4, and mana potions had no column left (docs/fixbelt2.md).

## What Changes

- internal/town/belt: `PlaceFor` / `WillBeltFor`: a potion may not open an empty column when that leaves a planned kind without a column.
- internal/town/beltfill.go: fill candidates are placed on a copy of the belt with PlaceFor (no candidate for a kind without room, kinds without a column first); `no belt column` is no failure and never marks the belt full.
- internal/loot/loot.go: a ground potion that would take a reserved column is not wanted.
- Delta: town-basics ADDED `A kind without room is not a full belt`.
