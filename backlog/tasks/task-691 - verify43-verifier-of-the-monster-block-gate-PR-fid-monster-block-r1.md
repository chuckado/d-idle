---
id: TASK-691
title: 'verify43: verifier of the monster block gate PR (fid/monster-block-r1)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T18:24:00Z'
labels:
  - d2d
  - readonly
  - d2d-combat
  - fidelity
dependencies: []
priority: high
ordinal: 336000
card: "verify43"
owns:
  - "docs/research/decomp43-monster-block-verify.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Consumer: the coordinator's upstream decision on fid/monster-block-r1 (decomp43, TASK-683). Why: the independent Opus verifier is required before any upstream post. Done when: the verify doc has a verdict and the evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Claim re-derived from the oracle and the decomp trace; compare exit codes checked on bad input
- [ ] #2 Remaining mutants run on the builder long lane with results recorded
- [ ] #3 Commit hygiene, CI/CodeRabbit fit and side-effect risks reviewed; open upstream PRs #634 and #637 checked read-only
- [ ] #4 Verdict GO / GO WITH FIXES / NO-GO written in docs/research/decomp43-monster-block-verify.md
<!-- AC:END -->
