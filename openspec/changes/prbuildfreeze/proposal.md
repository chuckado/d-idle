# prbuildfreeze: upstream PR prep for the freeze rules (local/freeze-rules-v3)

## Why

The user authorised up to five more upstream PRs (2026-11-02), each only after an independent re-verification. Candidate 1 of docs/research/pr-candidates-20261102.md is the freeze rules branch `fix/freeze-rules-v2`; it conflicts with open upstream PR 159 on one tool file. This card picks the order, rebuilds, re-runs the suites and differentials and fixes the draft.

## What Changes

- Private fork: branch `local/freeze-rules-v3` on upstream main, independent of PR 159 (merges cleanly in either order).
- Adds docs/research/pr-freeze-rules-draft.md (public PR body draft, replaces the stale pr-freeze-draft.md as the one to use).
- No PR opened, nothing pushed to GitHub.

## Overlap (declared)

docs/research/pr-freeze-draft.md and pr-freeze-proof.md are owned by earlier cards (archived); this card only adds a new draft and leaves them as the evidence trail.
