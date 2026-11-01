## Why

The walk-velocity candidate was proven and verified (docs/research/pr-walk-proof.md) and the user gave the go on 2026-11-01 to open it upstream. This card records the replay, the review-lesson changes made before opening, and the PR.

## What Changes

- Branch fix/walk-velocity-v4 on the worker host (from v3 head 8db11206, upstream main unchanged at 72132b99): the reviewer mode moved out of test_game.cpp (which would have passed upstream's 1,500-line limit) into its own tests/game/test_walk_velocity.cpp, the oracle's seed masked to 32 bits, include-cleaner fixes, a patched-data gate; pushed to the fork as fix/walk-velocity and opened as a PR.
- Docs only in this repo: posted lines in the evidence pack, the bundle-queue row, a BACKLOG line.
- No product code changes in this repo.

## Impact

Owns the evidence pack, the bundle-queue walk row and its BACKLOG line. Nothing live.
