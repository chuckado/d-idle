---
id: TASK-493
title: 'herostats: hero defence, block, damage and attack rating in the d2d dump'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T11:53:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 274000
card: "herostats"
owns:
  - "docs/survival/herostats-2026-11-15.md"
  - "internal/d2dsrc/gearstats.go"
  - "internal/d2dsrc/gearstats_test.go"
  - "internal/power/hero.go"
  - "docs/survival/power.md"
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
- [ ] #1 Fork dump stats carries gear, weapons and hero (block, ar, damage per hand, dmgpct), pushed to the private fork
- [ ] #2 Go parser reads hero additively; tests pass
- [ ] #3 Docs: power.md readable table, research note with V/U tags
<!-- AC:END -->
