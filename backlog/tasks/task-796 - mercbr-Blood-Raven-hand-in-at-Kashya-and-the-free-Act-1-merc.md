---
id: TASK-796
title: 'mercbr: Blood Raven hand-in at Kashya and the free Act 1 merc'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T01:59:00Z'
labels:
  - bot
  - probe
dependencies: []
priority: medium
ordinal: 366000
card: "mercbr"
owns:
  - "tools/provingground/scenarios/merc.json"
  - "internal/scenario/state.go"
  - "internal/scenario/merc_test.go"
  - "cmd/d2scenario/run.go"
  - "docs/probes/mercbr-2026-11-21.md"
exclusive: []
contracts: []
result: "Verified live on the lab d2d (builder long lane): merc.kashya-handin GREEN Paladin/11 and Sorceress/3; hand-in sets reward bit 0 and the free Rogue appears. No bot defect found; details docs/probes/mercbr-2026-11-21.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the user presses Start on the Blood Raven hand-in; prove the hand-in and the free Rogue on the lab. Done when: scenario merc.kashya-handin is green on two classes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 merc.kashya-handin runs green live on Paladin and Sorceress
- [x] #2 merc_hired assertion has must-fail tests
<!-- AC:END -->
