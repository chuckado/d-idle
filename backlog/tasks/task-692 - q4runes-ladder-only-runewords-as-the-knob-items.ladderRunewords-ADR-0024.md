---
id: TASK-692
title: 'q4runes: ladder-only runewords as the knob items.ladderRunewords (ADR 0024)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-20T18:27:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 336000
card: "q4runes"
owns:
  - "internal/env/tuning_items.go"
  - "internal/plandata/ladder_runewords.go"
  - "internal/plandata/ladder_runewords_test.go"
  - "docs/knobs.md"
  - "internal/env/tuning.go"
  - "internal/data/db.go"
  - "internal/data/extra.go"
  - "internal/data/data_test.go"
  - "internal/town/imbue_test.go"
  - "internal/watch/returnscreen.go"
  - "internal/watch/returnscreen_test.go"
  - "internal/watch/funwire.go"
  - "cmd/d2bot/d2dbackend.go"
  - "cmd/d2bot/funwire.go"
  - "cmd/d2bot/ladder.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

ADR 0024. Gates and fidelity runs stay vanilla. Done when the above holds.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Knob items.ladderRunewords (default 0) in the knob tables; off = no behaviour change
- [ ] #2 Ladder rows count as available in the planner only with the knob; class table for Paladin and Sorceress; tests on two classes
- [ ] #3 Return screen shows the rule 'ladder runewords: on/off'
- [ ] #4 Fork switch --ladder-runewords pushed to the private fork (differential recorded)
<!-- AC:END -->
