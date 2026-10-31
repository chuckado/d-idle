## Why

The independent verifier (card verifyprs) found the walk velocity candidate `fix/walk-velocity` HOLDS WITH LIMITS but NOT READY: the branch sat on an old upstream base, mercs and pets still ran at MonStats Run, and the d2d side of the proof was a transcription with no reviewer-runnable mode. This card fixes those on a new branch `fix/walk-velocity-v2` (worker host, own worktree; the author's branch is untouched), re-checks the CI-class gates, and updates the PR docs. Local only: nothing is pushed or posted upstream.

## What Changes

- Branch `fix/walk-velocity-v2`: the three original commits replayed on upstream main, a merc commit, a pet commit, and a tooling commit (`tests/game/test_game.cpp --walk-velocity`, `tools/emu/walk_velocity.py`, README row).
- `tools/emu/walk_velocity.py` and `walk_velocity_mode.inc` here are copies of the branch's tooling files.
- `docs/research/pr-walk-proof.md`, `pr-walk-draft.md`, `pr-walk-evidence-pack.md` and the walk row of `pr-bundle-queue.md` updated (READY or NOT READY with reasons).

owns: docs/research/pr-walk-proof.md, docs/research/pr-walk-draft.md, docs/research/pr-walk-evidence-pack.md, docs/research/pr-bundle-queue.md, tools/emu/walk_velocity.py, tools/emu/walk_velocity_mode.inc, openspec/changes/walkfix/**
exclusive: none (offline; own worktree and build dirs of the candidate on the worker host, no game instance, no d2d slot; nothing pushed or posted)
gates: none (no spec numbers change)
