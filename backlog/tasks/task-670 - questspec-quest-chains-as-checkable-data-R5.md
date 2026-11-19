---
id: TASK-670
title: 'questspec: quest chains as checkable data (R5)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-19T17:17:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 328000
card: "questspec"
owns:
  - "internal/questspec"
  - "data/questspec"
  - "docs/questspec.md"
exclusive: []
result: "Checker (11 rules) plus 8 Act 1 chains, 2 capability sets (sandbox devp13 shipped, lab) and an allow list, run by go test in internal/questspec; must-fail fixtures reproduce RC1-RC3, green on a fixed backend. Today: 3 allow-listed findings, all find_tree (RC1 residual). Not verified live; no bot change; data/questboard untouched. docs/questspec.md"
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

R5 of the Fable review (docs/design/fable-review-2026-11-19/06-roadmap.md). Checker plus data annotation, no bot behaviour change. Done when the checker runs in go test and docs/questspec.md exists.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 internal/questspec loads data/questspec and checks every chain against the capability sets in CI (go test)
- [x] #2 must-fail fixtures for the old Cain bugs, green on the fixed set
- [x] #3 Act 1 chains converted without changing data/questboard or the executor
<!-- AC:END -->
