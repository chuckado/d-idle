---
id: TASK-949
title: 'epic-nightmare: Nightmare difficulty complete'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:05:00Z'
labels:
  - bot
  - feature
  - T4
dependencies: []
priority: low
ordinal: 450000
card: "epic-nightmare"
owns:
  - "docs/epics/epic-nightmare.md"
exclusive: []
contracts: []
lane: later
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Difficulty epic: Nightmare never ran live (docs/nightmare-campaign.md status banner: OFFLINE ONLY, U). Templates nightmare_all exist; quest board rows list `nm` difficulties with meta levels (data/questboard/act*.json). Unmet; waits for epic-normal. Done when: the criteria below are ticked.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 epic-normal is Done.
- [ ] #2 The quest board data and drills cover the `nm` meta rows for all five acts (data/questboard/act*.json, a Nightmare drill matrix).
- [ ] #3 Nightmare readiness floors defined per boss and leg (docs/difficulty.md, internal/task/nightmare.go) and measured, not guessed.
- [ ] #4 Resist penalty and immunity handling (curses, merc, gear) proven for at least two classes.
- [ ] #5 A fresh hero completes Nightmare Acts I to V live, writes after start 0.
<!-- AC:END -->
