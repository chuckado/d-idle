---
id: TASK-458
title: 'survreplay: replay sandbox hero logs through the survival shadow and controller'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T23:08:00Z'
labels:
  - bot
  - model-check
dependencies: []
priority: medium
ordinal: 262000
card: "survreplay"
owns:
  - "docs/survival/replay-2026-11-14.md"
  - "internal/survival/testdata"
  - "tools/levelsim/controller.go"
  - "tools/levelsim/shadow_test.go"
  - "tools/levelsim/cmd/levelsim/main.go"
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
- [ ] #1 Sanitized replay fixture from the sandbox log, under 1 MB
- [ ] #2 levelsim -controller replays the ledger and shadow controller on it (Tower builder)
- [ ] #3 Calibration report docs/survival/replay-2026-11-14.md with the section 9 pass rule result
<!-- AC:END -->
