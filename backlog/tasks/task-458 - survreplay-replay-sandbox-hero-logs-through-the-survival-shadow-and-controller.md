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
result: "Ran on the builder, not live (offline replay of 7.8 h of a real sandbox log, clvl 11 to 14). S-B pass rule FAILS (balanced: bad stays under the enter line 6 of 14, quiet stays 8 of 8). S-C controller: no second Black Marsh entry for 2 h 31 min after the death sequence, 0 empty candidate sets in 445 minutes, second revive refused; but k_hero 6.2 (replay lacks weapon and defence) makes it far too timid. Nothing in params.json or profiles.json changed; proposals in the report."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the S-B margin and S-C controller had never run on a real log. Done when: the sandbox hero's retained bot log is a sanitized fixture, `levelsim -controller` replays it, and docs/survival/replay-2026-11-14.md reports the pass rules and the constants that look wrong.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Sanitized replay fixture from the sandbox log, under 1 MB
- [x] #2 levelsim -controller replays the ledger and shadow controller on it (builder)
- [x] #3 Calibration report docs/survival/replay-2026-11-14.md with the section 9 pass rule result
<!-- AC:END -->
