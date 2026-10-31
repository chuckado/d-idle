## Why

The early fidelity baseline failed (d2d -48 % XP/h, -52 % kills/h against Game.exe, docs/d2d-baselines-early.md). The risk map ranks monster ranged attacks, specials, melee and walk speed as the next fidelity risks (docs/research/d2d-fidelity-riskmap.md). This card (decomp-queue item 7) traces monster melee and missile attack resolution in game.exe 1.14d against d2d on the current upstream main, runs game.exe's own functions as the oracle, and measures from the soak logs how much of the gap monsters could explain at all. no observable behaviour change: research, docs and emulator scripts only; no Go code, no game instance.

## What Changes

- `docs/research/monster-attack-trace.md`: second pass (melee re-run on upstream 0dc4c62a, missiles added, Q4 time split, Q5 bot findings); the first pass archived to `docs/archive/monster-attack-trace-2026-10-30.md` with a status banner.
- `tools/emu/mon_missile.py`: oracle for the missile block gate (FUN_0057dfb0) and the missile damage roll (FUN_005a89a0), plus an exhaustive table check of the ToHit / physical gates; `tools/emu/monster_hit.py`: the d2d get-hit port follows current `rules::get_hit`.
- `docs/decomp-queue.md` item 7 row and result line, the `docs/research/README.md` and `docs/INDEX.md` lines, a row in `docs/research/pr-bundle-queue.md`.
- Result: melee clean; two narrow missile gate deviations proven (none in the early game); PR candidate not built.

owns: docs/research/monster-attack-trace.md, docs/archive/monster-attack-trace-2026-10-30.md, tools/emu/mon_*.py, tools/emu/monster_hit.py, openspec/changes/decompmon/**
exclusive: none (offline; no game instance, no d2d slot, no Go changes)
gates: none (research)
