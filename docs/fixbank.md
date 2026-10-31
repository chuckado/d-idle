# FIXBANK: gold deposit and repair gold (card fix-bank)

> status: current. Findings of the D3 drills (docs/probes/drills/2026-10-31-d3.txt); live proof is a drill rerun.

## Fixed (shared code, offline-tested; the d2d part is d2d-only)
- P5 (D3.s, D3.s2): on d2d the gold amount prompt starts EMPTY and `hoverFirst` hides `TypeChar`, so `depositGold` failed for full and partial amounts. New optional controller capability `town.goldMover` (`MoveGold(n int, deposit bool) error`). `internal/d2dtown` implements it on both town controllers (`HoverFirstBelt`, `Input`) through `d2dctl.MoveGold` (V live in the drill control). `depositGold` calls it when present, then confirms by memory as before. Game.exe's controller does not have it: its typed path is unchanged (shared interface, d2d implements).
- Caps: the stash cap clamp (`bankAmount`, StashGoldCap 2,500,000 UNVERIFIED, learned from a refusal) runs before the move. The purse cap (10,000 at clvl 1, D4.cap) matters only for a withdrawal; the town package has no withdraw chore, so nothing can overfill the purse. `MoveGold(n, false)` exists for a future one, which must clamp to purse room.
- D3.r0: `needRepair` also needs gold >= `repairEstimate` (a quarter of each worn item's base cost times its missing durability, at least 1 gold when damaged). UNVERIFIED as a price; it is a lower bound, so a character that can pay is never turned away, and gold 0 opens no smith. Shared.

## D3.tp: not a bug, no change
The tome ended with 5 charges because `buyScrolls` lifts the target to 5 when a tome is held (`j.target < 5 -> 5`), and `TPScrollTarget` (2) is a knob; the poor-character rule (TripGoldFloor) can cap it lower. The loop buys one scroll per iteration until `have == target`. 20 is the tome's maximum, not the bot's target. Policy, not a defect.

## Drills to rerun after merge
D3.s (deposit all 1500: purse 0, stash 1500), D3.s2 (deposit 1000 of 1500: purse 500, stash 1000), D3.r0 (gold 0: no repair step runs, no ticks spent at the smith), D3.tp (unchanged expectation: tome plus 5 charges).
