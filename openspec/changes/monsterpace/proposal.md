## Why

The second monster AI pass (monsterai2) and the first (monster-ai-trace) proved that game.exe starts a monster at 75 percent velocity and that its Act 1 thinks set a pace d2d leaves out. The owner authorised one bundled upstream PR for the monster movement pace, each after independent Opus verification. This card prepares the candidate on the private fork and the PR draft; it opens nothing.

## What Changes

- Private fork (not this repo): branches `local/monsterpace-clean` (14 commits) and `local/monsterpace` (15, same tree) on upstream main 6935cf64; counts per commit in the draft.
- docs/research/pr-monsterpace-draft.md: the PR draft, proofs, limits and the verifier's checklist.
- tools/emu/mp_run.sh, mp_patch.py: the counting replay (evidence tooling; the oracle files are monsterai2's).

## Overlap (declared)

tools/emu/mai_ais.py, mai_*.py and mai2_*.py are monsterai2's (branch agent/monsterai2, owns them): this branch carries identical copies (checked out of that branch unchanged) because mp_run.sh imports them. Same bytes, so the merges agree; whoever merges second drops no change.

## Impact

Docs and evidence tooling here; the code change is upstream-bound and lives on the fork. No spec, policy, ADR or glossary change. Nothing posted.
