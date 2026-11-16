---
id: TASK-525
title: 'lootwire-L4: shadow stage logging disagreements between today and the Starter loot rules'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T02:11:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-524
priority: medium
ordinal: 285000
card: "lootwireL4"
owns:
  - "internal/loot/lootshadow.go"
  - "internal/loot/lootshadow_test.go"
  - "internal/watch/lootshadowapi.go"
  - "internal/watch/lootshadowapi_test.go"
  - "docs/lootrules.md"
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
- [ ] #1 shadow log line, counters, GET /api/lootrules/shadow, bot.lootShadow knob
- [ ] #2 zero Starter pack disagreements on fixtures, ground equivalent test, latency budget test
- [ ] #3 Shadow findings section in docs/lootrules.md
<!-- AC:END -->
