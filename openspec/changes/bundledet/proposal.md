## Why

The user decided (2026-11-01) on one PR with two commits for the Frozen Armor and freeze-resume claims. The combined branch had four commits, and the determinism state hashes for the resume claim and for the combination were still owed (the armor claim alone had run 14 of 14 identical).

## What Changes

- Worker host (not in this repo): new local branch `fix/frozen-armor-and-resume-2c` in the PR checkout `bundlear`, the same tree as the four-commit branch squashed to two commits (armor claim plus tooling, resume claim plus tooling).
- Det state hashes: the 14 DETHASH2 scenarios on the lab base, the lab base plus the resume claim, and the lab base plus both claims.
- docs/research/pr-bundle-armor-resume-det.md: the det hash log and the per-commit build and test results.

## Impact

Docs only in this repo. Nothing is pushed, posted or opened. The PR body file belongs to the bundlearmor card; the edit to it is an overlap, declared in tasks.md. No spec, policy, ADR or glossary change.
