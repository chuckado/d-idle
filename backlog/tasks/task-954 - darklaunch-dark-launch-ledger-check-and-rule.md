---
id: TASK-954
title: 'darklaunch: dark-launch ledger, check and rule'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-26T03:33:00Z'
updated_date: '2026-11-26T05:37:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 454000
card: "darklaunch"
owns:
  - "data/darklaunch.json"
  - "tools/darklaunch"
  - "docs/process/dark-launch.md"
  - "tools/ci/check.sh"
exclusive: []
result: "Ledger of 33 dark features, check.py wired into tools/ci/check.sh, report.py, rule doc. Offline only; criteria_met is U for all rows (not verified live)."
contracts: []
lane: next
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 data/darklaunch.json seeded
- [x] #2 tools/darklaunch/check.py wired into tools/ci/check.sh with tests
- [x] #3 docs/process/dark-launch.md
<!-- AC:END -->
