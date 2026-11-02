## Why

docs/coverage/rule-coverage.md lists monster.stats (734 rows) at 0 percent: no base stat of a plain monster was ever compared with game.exe, and the XP/h baseline audit flagged per-kill experience. Only boss spawn stats (DECOMPMODS) and the XP award chain (DECOMPXP) were traced.

## What Changes

- docs/research/monster-stats-trace.md: how game.exe 1.14d builds a plain monster's life, defence, to-hit, damage, experience and resistances (FUN_00573cb0, FUN_006538a0), compared with d2d's real `rules::monster_stats` (242,220 row cases) and with the real unit stats of a populated level (Act 1 and 2, three difficulties); deviations D1.., bot findings B1.., drill recipe, matrix lines.
- tools/emu/ms_*: the differential (game side in the emulator, d2d side a small test linked to a scratch upstream build).

## Impact

Docs and tools only. No spec, policy, ADR or glossary change.
