---
id: TASK-792
title: 'parkrelease: gave_up park retried once after a new bot release'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T21:32:00Z'
updated_date: '2026-11-21T22:04:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 364000
card: "parkrelease"
owns:
  - "internal/questboard"
  - "internal/scenario"
  - "cmd/d2scenario"
  - "internal/bot/questboard.go"
  - "tools/provingground/scenarios/board.json"
  - "docs/probes/park-retry-after-release-2026-11-21.md"
exclusive: []
contracts: []
result: "gave_up parks of another release retried once (RetriesAfterRelease); unit tests Paladin+Sorceress; scenario added as pending, NOT run live (no lab slot)"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Unit tests on Paladin and Sorceress
- [x] #2 Scenario board.park-retry-after-release added (lab run pending)
<!-- AC:END -->
