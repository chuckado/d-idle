## Why

The early fidelity baseline failed (d2d about 48 % below Game.exe on XP/h) and DECOMPMON found that most of the extra time per kill is walking (docs/research/monster-attack-trace.md); the post-fix soak is still 37.6 % below the d2d baseline (docs/probes/postfix-soak1.md). The risk map ranks walk speed 10th (docs/research/d2d-fidelity-riskmap.md). This card (decomp-queue item 12) traces walk and run speed and the movement step in game.exe 1.14d, compares d2d and our bot, and attributes the walking time from the soak logs. No observable behaviour change: research, docs and emulator scripts only; no Go code, no game instance.

## What Changes

- `docs/research/walk-speed-trace.md`: Q1-Q5 with addresses and tags, the deviation table, the walking-time attribution, bot findings B1-B4.
- `tools/emu/walk_speed.py` + `walk_speed_harness.cpp`: exhaustive differential (game.exe's own code vs d2d before / after); `tools/emu/walk_soak_speed.py`, `walk_soak_backoff.py`: soak log readers.
- `docs/research/pr-walk-proof.md`, `pr-walk-draft.md`, `pr-walk-evidence-pack.md`: local PR candidate (3 claims), a row in `docs/research/pr-bundle-queue.md`.
- `docs/decomp-queue.md` item 12 row and result line; the `docs/research/README.md` and `docs/INDEX.md` lines.

owns: docs/research/walk-speed-trace.md, docs/research/pr-walk-*.md, tools/emu/walk_*, openspec/changes/decompwalk/**
exclusive: none (offline; no game instance, no d2d slot, no Go changes)
gates: none (research)
