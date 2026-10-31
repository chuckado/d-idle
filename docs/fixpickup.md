# FIXPICKUP: ground potion pickup fixes (card fix-pickup)

> status: current. Findings B1, B2, B4, B5 of the P7 review; live proof is drill D1/P7.

Rule (V by emulator, docs/research/belt-pickup-trace.md): panel closed, a clicked ground potion takes the first column whose bottom potion is the same kind (lowest free box), else the first empty column, else the pack. Panel open: cursor.

## Fixed (shared code, offline-tested)
- `belt.Place(slots, items, kind)` / `belt.WillBelt`: the rule as pure functions (table test TestPlace: matching room, empty column, mixed full, rejuv, 4/8/12/16 boxes).
- B1: `needsOf.Want` wants a potion the plan carries when WillBelt, whatever the pack holds; the pack buffer only limits pack-bound potions. Policy kept: a kind the plan has no column for (Barbarian mana in Normal) still follows the buffer rule, because it would spoil a column.
- B2: `candidates` uses WillBelt instead of Free() > 0.
- B4: `pickBeltBox` was the game rule plus one extra pass preferring the plan column for the kind; it is now `belt.Place` alone, since a Game.exe shift-click lands where the game puts it.
- B5: PickUp and pickOne already close open panels (`hand.PanelOpen` then `CloseAll`) before the ground click, re-checking the state; backend-neutral, no change.
- Arrival: `arrived` also accepts a same-class count increase in pack plus belt, so a potion whose unit id changes on pickup is not reported "picked 0". Cause was code-read only, UNVERIFIED live.
- sim.Add now sets the item cell (Item.GX/GY) like the real reader; the belt tests had all boxes at 0.

Unchanged: rejuvenation keeps its carry cap (wantRejuv) when the belt would not take it. Shared by both backends.
