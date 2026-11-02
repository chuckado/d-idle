## Why

The trace docs/research/waypoint-trace.md claims game.exe refuses a waypoint use only within 10 s of the player's last hostile declaration (one writer of pcdata offset 0x160) and that d2d's every-arrival rule differs in 59,309 of 132,964 cases. A PR could rest on it, so an independent look was needed.

## What Changes

- docs/research/verify-waypoint.md: verdicts per claim, counts, seeds, reproduction commands, limits, deviations, bot finding.
- tools/emu/wpv_scan.py, wpv_ctx.py, wpv_dis.py, wpv_cmp.py: the scan of game.exe for every access of the field, context dumps, and the rule comparison against the game oracle.
- Docs and emulator scripts only; no Go code, no live game, no d2d slot.

## Impact

None on shipped code. Result docs/research/verify-waypoint.md.
