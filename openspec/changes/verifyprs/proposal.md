## Why

docs/process/upstream-pr-standard.md section 5 requires an independent adversarial verifier for every upstream PR candidate before the coordinator may ask the user for a go. Two local candidates have none: `fix/static-field` (decompstatic) and `fix/walk-velocity` (decompwalk). This card is that verifier: re-derive game.exe's rule from the disassembly without the authors' oracles, run an own differential against the REAL d2d builds before and after, sweep d2d and game.exe for sibling code paths, and run upstream's CI-class checks.

## What Changes

- `tools/emu/verify_static_field.py` + `verify_static_field_mode.inc`, `tools/emu/verify_walk_velocity.py` + `verify_walk_velocity_mode.inc`: own generators, game.exe emulator sides and d2d test_game modes that drive the real `Fight::spot` / `monster_update` / `World::walk_percent`.
- A verifier section appended to `docs/research/pr-static-proof.md` and `docs/research/pr-walk-proof.md` (their cards are archived).
- `docs/research/pr-bundle-queue.md`: the walk row updated and a static row added (READY / NOT READY with reasons); one `docs/BACKLOG.md` line.

owns: tools/emu/verify_static_field.py, tools/emu/verify_static_field_mode.inc, tools/emu/verify_walk_velocity.py, tools/emu/verify_walk_velocity_mode.inc, docs/research/pr-static-proof.md, docs/research/pr-walk-proof.md, docs/research/pr-bundle-queue.md, docs/BACKLOG.md, openspec/changes/verifyprs/**
exclusive: none (offline; own throwaway worktrees of the candidate branches on the worker host, no game instance, no d2d slot; nothing pushed or posted)
gates: none (verification; no spec numbers change)
