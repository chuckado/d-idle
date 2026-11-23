---
id: TASK-465
title: 'drills012: friends 0.1.2 pre-release drills with real game files'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T04:15:00Z'
labels:
  - container
  - drill
dependencies: []
priority: medium
ordinal: 265000
card: "drills012"
owns:
  - "docs/releases/0.1.2-drills.md"
exclusive: []
contracts: []
result: "All four drills PASS, verified live with the user's backup zip as game source: install shows only the nine archives in /game, 21 min health run (0 restarts, 0 panics, 0 deaths, 0 bounce loops, Den entered x3), update 0.1.1 to 0.1.2 through the real client (signed set seq 3 accepted), rollback to 0.1.1. Crash-release rollback variant not rerun (needs beta key). Finding: ./didle update prints applied after a not-safe refusal. See docs/releases/0.1.2-drills.md. Cleanup done."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the 0.1.2 pack was staged but its docker drills with real archives had not run. Done when: the four drills are recorded in docs/releases/0.1.2-drills.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Drill 1: install on the extracted zip folder, only nine archives in /game, wizard and d2d start
- [x] #2 Drill 2: 20 minute health run, no panics, Den entered, no bounce loops
- [x] #3 Drill 3: update 0.1.1 to 0.1.2 through the real client from a local copy of the staged feed
- [x] #4 Drill 4: rollback drill
- [x] #5 Cleanup: project, volumes, scratch removed, nothing left
<!-- AC:END -->
