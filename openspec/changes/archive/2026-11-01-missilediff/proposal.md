## Why

The monster-missile trace proved D1 (level brackets), D2 (HitShift) and D3 (the row's own element) only against a Python port of d2d's expression. Rule-coverage ledger rows for the Act 1 / 2 monster missiles stayed READ-ONLY. This card runs the real d2d build.

## What Changes

- docs/research/missilediff.md: the real-build differential (game.exe launch record against the real d2d `monster_shot_phys` / `monster_shot_stats`, 120,000 cases), D1-D3 CONFIRMED or REFUTED with counts, a local PR candidate branch, drill notes.
- tools/emu/mm_diff.py and mm_diff_run.sh: the game.exe side and the comparison (extends mm_launch.py).
- docs/coverage/ledger.tsv: rows the run proves.
- The d2d side: a local branch `local/missilediff-record` in the private fork checkout (never pushed upstream, never posted).

## Impact

Docs and emulator scripts only in this repo. No Go, no live game, no d2d slot, no spec, ADR or glossary change. Missing glossary term for the coordinator: "launch record".
