## Why

The user asked (2026-11-01) for the Forgejo workers to help. The worker host is busy with soaks and drills, while the second machine (an idle arm64 Mac) can build d2d and run the portable parts of PR-candidate vetting: the differentials, the Python checks, clang-tidy and the textual merge check.

## What Changes

- A wrapper, its remote half, a setup and parity doc and a workflow sketch in the private ops repo (not in this repo).
- docs/review-strategy.md: one short section saying the portable differentials and checks can run on a second machine and that parity was checked (counts only).

## Impact

Docs only in this repo. No spec, policy, ADR or glossary change. Nothing run live; no game was used.

Overlap (declared): docs/review-strategy.md is owned by the open card soak-minutes-virtual; this card appends one new section at the end of the file (no edit of existing text) and does not list the file in owns, so the two do not collide.
