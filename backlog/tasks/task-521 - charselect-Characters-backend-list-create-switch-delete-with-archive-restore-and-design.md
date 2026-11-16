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
result: "Phase A built offline, not verified live. d2d has no list/create/delete verb; list, delete and restore are save-folder operations, create and switch are a relaunch with --start-class or --char under the supervisor (exit code 6), so no fork verb is needed. State is per instance, not per character: a switch parks and unparks the per-character files (internal/charsel/perchar.json). Routes GET and POST /api/characters (watch.characters/1) with goldens; archive then delete with typed name; active delete and measured-run refused. Live checklist and the Phase B UI plan in docs/charselect.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Phase A of the Character Select screen (user request 2026-11-15). No UI here (Phase B). Done when the backend is merged and docs/charselect.md carries the design and live checklist.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/charselect.md: feasibility, design, live checklist
- [x] #2 internal/charsel: header reader, name rules, store with archive and restore, state parking, manager
- [x] #3 watch.characters/1 routes with goldens and gates
- [x] #4 d2bot wiring: Deps.Characters hook, supervisor switch (exit 6)
- [x] #5 tests on Paladin and Sorceress; gate batch green
<!-- AC:END -->
