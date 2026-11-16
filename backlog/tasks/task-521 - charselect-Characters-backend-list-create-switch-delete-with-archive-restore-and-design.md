---
id: TASK-521
title: 'charselect: Characters backend (list, create, switch, delete with archive, restore) and design'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T02:04:00Z'
labels:
  - watch
  - feature
  - tester
dependencies: []
priority: medium
ordinal: 282000
card: "charselect"
owns:
  - "internal/charsel"
  - "docs/charselect.md"
  - "internal/watch/charactersapi.go"
  - "internal/watch/charactersapi_test.go"
  - "internal/watch/testdata/characters"
  - "cmd/d2bot/charwire.go"
  - "cmd/d2bot/charwire_test.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Phase A of the Character Select screen (user request 2026-11-15). No UI here (Phase B). Done when the backend is merged and docs/charselect.md carries the design and live checklist.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 docs/charselect.md: feasibility, design, live checklist
- [ ] #2 internal/charsel: header reader, name rules, store with archive and restore, state parking, manager
- [ ] #3 watch.characters/1 routes with goldens and gates
- [ ] #4 d2bot wiring: Deps.Characters hook, supervisor switch (exit 6)
- [ ] #5 tests on Paladin and Sorceress; gate batch green
<!-- AC:END -->
