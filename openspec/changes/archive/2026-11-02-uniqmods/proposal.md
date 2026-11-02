## Why

Decomp follow-up of queue item 13 (monster mods): champions, uniques, minions and the Act 1 / 2 superuniques. The earlier traces (DECOMPMODS, BUILDMODS) proved what each mod init writes; what was missing is the pick side (MonUMod weights per difficulty, how many mods, exclusions and fPick rules), the pack and minion counts, the Act 1 superunique records, the multipliers by class and difficulty in one place, and the bot's handling (which modifiers it knows, kill-first rules).

## What Changes

- docs/research/uniqmods-trace.md: questions Q1..Q9 with address tags, d2d deviations D1..D4 (read-only, no PR candidate), bot findings B1..B7, drill recipe G1..G6, proposed matrix text, dead ends.
- tools/emu/um_* (own prefix): `um_extract.sh` (one table through mpq-cat on the worker host), `um_tab.py` (MonUMod weights, MonStats rows), `um_fpick.py` (rows where fPick 1 / 3 differ between game.exe and d2d), `um_summ.py` (pack and minion summary of a umod_spawn log).

## Impact

Docs and offline tooling only. No Go code, no live game, no d2d slot. Nothing pushed or posted upstream. No coverage ledger row (no new differential could be run: the runner refused remote scripts this session).
