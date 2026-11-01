## Why

The first det run (docs/research/det-hashes-2026-11-01.md) left coverage gaps: no cold missile visibly hit a monster, no merc or pet was present, the player position was not hashed, and the freeze-rules candidate was not run. This change closes them with new scripted scenarios and re-runs the whole set on five builds.

## What Changes

- New scenario scripts under tools/emu/ (headless d2d with the lab patch set, one hidden slot, no game.exe).
- docs/research/det-hashes-2026-11-01b.md: commands, builds, per-candidate verdicts, remaining coverage gaps.
- Queue rows and a BACKLOG line only where a verdict allows.

## Impact

Docs and tools only; nothing pushed upstream or posted. No spec delta.
