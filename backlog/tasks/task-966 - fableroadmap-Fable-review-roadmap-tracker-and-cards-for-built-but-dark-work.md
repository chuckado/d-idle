---
id: TASK-966
title: 'fableroadmap: Fable review roadmap tracker and cards for built-but-dark work'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:50:00Z'
labels:
  - docs
  - chore
dependencies: []
priority: high
ordinal: 455000
card: "fableroadmap"
owns:
  - "docs/roadmap-fable.md"
  - "data/darklaunch.json"
  - "backlog/tasks/task-96[6-9] - *"
  - "backlog/tasks/task-9[7-9][0-9] - *"
exclusive: []
result: "Docs and cards only, no code run. docs/roadmap-fable.md: one row per recommendation of all Fable reviews with state from git log and the knob table (V), Built but dark flip order, Operating rule; data/darklaunch.json: four rows adjusted (arbiter outcome rule, town.plan, boardSM, checkpoint). 17 Proposed cards created (TASK-970 to TASK-991 without the pieces other agents carded; ids can shift at merge). Not verified live."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the Fable reviews recommended structural work that was built and left at default 0; a living tracker and cards make it get acted on (user popup 2026-11-26). Done when: tracker merged, check.py clean, cards on the board.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/roadmap-fable.md: one row per Fable recommendation with state, evidence, blocker, next action, enable criteria
- [x] #2 Built-but-dark section with flip order tied to data/darklaunch.json
- [x] #3 Cards created for recommendations that had none
- [x] #4 Operating rule paragraph (cause class check before a new patch card)
<!-- AC:END -->
