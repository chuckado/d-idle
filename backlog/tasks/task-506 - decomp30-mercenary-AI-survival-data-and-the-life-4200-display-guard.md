---
id: TASK-506
title: 'decomp30: mercenary AI, survival data and the life 4200% display guard'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T15:33:00Z'
labels:
  - bot
  - research
dependencies: []
priority: medium
ordinal: 277000
card: "decomp30"
owns:
  - "docs/research/decomp30-mercai.md"
  - "internal/d2/merc.go"
  - "internal/d2/merc_test.go"
  - "internal/watch/web/inventory-merc.js"
  - "docs/upstream/pipeline-2026-11-14.md"
exclusive: []
contracts: []
result: "trace doc plus display clamp (B1) with test; no emulator run (worker down); not verified live; root cause of life 4200% unread (U)"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the merc dies constantly; learn what game.exe's merc AI does so the survival model and merc policy are data-driven. Consumer: docs/survival/design-2026-11-14.md section 5.5. Done when: the trace doc exists and the life-percent display is clamped with a test. No emulator run (worker down).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/research/decomp30-mercai.md written with V/U tags
- [x] #2 life percent display guard with test
<!-- AC:END -->
