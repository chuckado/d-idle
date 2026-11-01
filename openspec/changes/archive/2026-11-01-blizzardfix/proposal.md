## Why

The independent verifier rated the local d2d PR candidate `fix/blizzard-shard-offsets` READY WITH CHANGES (docs/research/pr-blizzard-verifier.md on branch agent/verifyblizzard). This card applies its three required changes and brings the PR draft up to date. Nothing is pushed or opened upstream; that needs the user's go.

## What Changes

- Candidate rebuilt in a new scratch worktree (branch `fix/blizzard-shard-offsets-v2`, the author's branch and worktree untouched): `cx` / `cy` renamed to `cell_x` / `cell_y` in the test and the Python tool, ruff findings fixed, the claim commit names Eruption.
- docs/research/pr-blizzard-draft.md: Eruption sentence, corrected D3 statement, the verifier's counts and verdict, final hashes, status READY, held for the user's go.

## Impact

Docs and the local scratch checkout only. No repo code, spec, ADR or glossary change.
