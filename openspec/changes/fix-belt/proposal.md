## Why

docs/primitives-review.md found the belt fill failing for a reason that is not the move: a belt put that fails with `hand: item not found` right after a re-entry set a sticky "belt full" mark that survived every new game (F1), no belt step ran after the shop so d2d purchases waited in the pack (F3), combat read the belt column as the d2d box and dropped boxes 4..15 (F7), a potion could be put above another kind's column (F8), and the kit gate ignored healing potions waiting in the pack (F6). F9 (why the id vanishes) needs log lines.

## What Changes

- internal/town/beltfill.go: item-not-found never marks the belt full and retries after a fresh read (candidates are re-read with the pack open); the mark also holds the pack's potion set and resets on `ForgetBelt` (new game); `pickBeltBox` respects column kind; per-put diagnostic line and mark set/clear lines.
- internal/town/chores.go: `beltafter` step after the shop.
- internal/bot/goal.go: `forgetGame` calls `ForgetBelt`. internal/bot/kitgate.go: pack potions count as the kit; `potionDiag`. internal/bot/observe.go: minute line gets `packHP`, `packMP`, `beltBoxes`.
- internal/combat/safety.go: `BeltOf` maps box = GY*4+GX, column = box%4 (shared by both backends: Game.exe gives (column,row), d2d gives the box in GX with GY 0).
- internal/loot/loot.go: rate-limited `loot: potion skip` diagnostic line.
- Delta: town-basics ADDED requirements. Shared code; the failures and the box mapping are d2d-driven. The portal diagnostic line belongs to fix-portal (cmd/d2bot/d2dwire.go).
