## Why

User request (2026-11-01): audit whether the original-game baselines (`tools/d2dsoak/baselines.json`, `baselines-early.json`) are trustworthy enough to gate on while d2d fidelity is still being finalised.

## What Changes

- docs/research/baseline-audit.md: per-file numbers, spec minimums PASS/FAIL, statistical power, confounds, a verdict per gate criterion, and the extra baseline runs that would fix the weakest spots.
- No code, no spec change, no new run on the game host (read-only).

## Impact

Docs only. Shared-file lines in merge-lines.txt.
