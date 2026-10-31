## Why

Live drills D3.s and D3.s2 (docs/probes/drills/2026-10-31-d3.txt) showed the bot cannot bank gold on d2d: the d2d amount prompt starts empty and the town controller hides TypeChar, so neither a full nor a partial deposit works, while the control verb `d2dctl MoveGold` does. Drill D3.r0 showed the repair chore opens the smith at gold 0 (101 ticks, a failed repair) before the latch stops it.

## What Changes

- Shared town package: optional controller capability `goldMover` (`MoveGold(n int, deposit bool) error`); `depositGold` uses it when the controller has it and keeps the typed Game.exe path otherwise. d2d implements it in internal/d2dtown (both town controllers) through `d2dctl.MoveGold`.
- Shared: `needRepair` also requires gold >= `repairEstimate`.
- D3.tp (tome ends with 5 charges): not a bug; policy (docs/fixbank.md). No change.
- Delta: town-basics ADDED requirements.

code merged 2e3de60; owns released so later cards can edit; live proof pending.
