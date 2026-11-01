## Why

The user wants more work on upstream PR #155 (Static Field) before CodeRabbit's first review, so one review covers it. Decomp item 6 had found a cold / chill deviation (C1, C2) with no patch.

## What Changes

- Re-derive the chill length and applicability rule from game.exe's disassembly; confirm C1, correct C2 (not reachable with 1.14d data).
- Local upstream candidate `fix/static-field-chill` on #155's head on the worker host: one claim commit (C1) and one tooling commit (`tools/emu/chill_length.py`, `test_game --cold-length`); nothing pushed or posted.
- In this repo: `docs/research/pr-static-chill-proof.md` and `-draft.md`, `tools/emu/chill_length.py` and `tools/emu/chill_length_mode.inc`, one row each in `docs/decomp-queue.md`, `docs/research/pr-bundle-queue.md` and `docs/BACKLOG.md`.
