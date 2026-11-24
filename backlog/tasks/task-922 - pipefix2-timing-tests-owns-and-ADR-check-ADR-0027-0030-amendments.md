---
id: TASK-922
title: 'pipefix2: timing tests, owns and ADR check, ADR 0027/0030 amendments'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T16:51:00Z'
updated_date: '2026-11-24T17:14:00Z'
labels:
  - ops
  - chore
dependencies: []
priority: medium
ordinal: 425000
card: "pipefix2"
owns:
  - "internal/bot/"
  - "internal/loot/"
  - "internal/nav/"
  - "internal/watch/"
  - "internal/d2dctl/"
  - "internal/d2denv/"
  - "internal/d2dtown/"
  - "cmd/d2dview/"
  - "tools/ci/"
  - "docs/process/"
  - "docs/adr/"
  - "docs/decisions.md"
  - "docs/backlog-howto.md"
  - "AGENTS.md"
exclusive: []
contracts: []
result: "audit fixes 3, 5, 6, 7: timing tests fixed or guarded (docs/process/timing-tests.md), tools/ci/ownscheck.py with tests wired in check.sh and the merge preflight, ADR 0027 and 0030 amended; unit tests and python tests run, whole-repo gate not run on this host, not verified live"
allow-adr: true
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Pipeline audit fixes 3, 5, 6, 7 (user approved 2026-11-24). Done when the checks and docs are in.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Wall-clock test ceilings fixed or race-guarded, policy in docs/process/timing-tests.md
- [x] #2 ownscheck.py wired in check.sh and merge preflight, tests pass
- [x] #3 ADR 0027 and 0030 amended, decisions.md updated
<!-- AC:END -->
