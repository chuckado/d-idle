---
id: TASK-532
title: 'intake: observed damage intake (lifeLost per minute) in minute line, glance and ledger'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T02:57:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 284000
card: "intake"
owns:
  - "internal/bot/hpfields.go"
  - "internal/survival/ledger.go"
  - "internal/watch/glance.go"
  - "tools/levelsim/shadow.go"
  - "docs/survival/ledger.md"
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
- [ ] #1 minute line carries lifeLost, hitsTaken, hpPolls; max-life changes never count
- [ ] #2 /api/glance intake {lifeLostPerMin, window}; ledger per-zone intake; levelsim -shadow reads lifeLost
- [ ] #3 docs/survival/ledger.md section Observed intake; gate green
<!-- AC:END -->
