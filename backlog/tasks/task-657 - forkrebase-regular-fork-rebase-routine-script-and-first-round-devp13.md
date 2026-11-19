---
id: TASK-657
title: 'forkrebase: regular fork rebase routine, script and first round (devp13)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-19T14:06:00Z'
labels:
  - d2d
  - chore
dependencies: []
priority: medium
ordinal: 246000
card: "forkrebase"
owns:
  - "docs/process/fork-rebase.md"
  - "docs/process/sandbox-devp13.md"
  - "tools/fork-rebase.sh"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 docs/process/fork-rebase.md: routine, conflict policy, tests, regression checklist
- [ ] #2 tools/fork-rebase.sh measure/rebase/test/build, stops on conflict
- [ ] #3 first round: rebase onto upstream main (didle/rebase-20261119), tests, image devp13 loaded not switched
- [ ] #4 docs/process/sandbox-devp13.md: payloads, diff vs devp12, rollback
<!-- AC:END -->
