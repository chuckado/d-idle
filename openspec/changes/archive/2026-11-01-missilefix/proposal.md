## Why

The independent verifier of `local/missilediff-record` (docs/research/pr-missilerecord-verifier.md, verdict NOT READY) found that the record claim holds but the row's poison is stored in `MonStats::elements` as a whole-point total, which `monster_blow` reads in the El convention (roll x 10 per frame for 2 x dur frames): up to 62x too strong. This card fixes the candidate on the private fork, re-runs both differentials, tries the rebase onto upstream main, and updates the doc and PR draft.

## What Changes

- Private fork (worker host, remote private only): the row poison moves to `MonStats::poison_min` / `poison_max` / `poison_len`; a `monster_blow` test; presubmit short-name, ruff and clang-tidy findings cleared; D5 stays out; rebase result on upstream main.
- docs/research/missilediff.md: corrected result and a short "verifier findings and fixes" section.
- docs/research/pr-missilerecord-draft.md: PR text draft held for the user's go.

## Impact

Docs only in this repo. No Go, no live game, no spec, ADR or glossary change. Overlap (declared): card missilediff (archived) owned docs/research/missilediff.md; this card edits it again.
