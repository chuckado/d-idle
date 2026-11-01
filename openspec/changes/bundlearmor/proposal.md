## Why

The user decided (2026-11-01) to bundle the Frozen Armor candidate (`fix/armor-hit`) and the freeze-resume candidate (`fix/freeze-resume`) as one upstream PR. Both are READY locally on the worker host, but they were proven on separate branches. A combined branch and one PR body were missing.

## What Changes

- Worker host (not in this repo): local branch `fix/frozen-armor-and-resume` in the PR checkout `bundlear`, built from upstream main 72132b99 by cherry-picking the two claims and their tooling; the comment citation fix; every differential and check re-run on the combined head.
- docs/research/pr-bundle-armor-resume-body.md: the combined PR body (draft status READY, held for the user's go), with per-commit claims, evidence, residuals, review order and check results.

## Impact

Docs only in this repo. Nothing is pushed, posted or opened; no spec, policy, ADR or glossary change. The pr-bundle-queue line is in the closing report for the merger (that file belongs to another card).
