---
id: TASK-992
title: 'backlogplan: To Do status, planning lanes, plan.py view, lane migration'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-26T04:02:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 456000
card: "backlogplan"
lane: now
owns:
  - "tools/backlog"
  - "backlog/tasks"
  - "backlog/config.yml"
  - "docs/backlog-howto.md"
  - "docs/plan/QUEUE.md"
exclusive: []
contracts: []
result: "To Do status, lane/lane_note/due/epic keys (check.py), tools/backlog/plan.py plan view, migration R1-R6 of 241 open cards; offline only, tests pass, tools/ci/check.sh not seen to finish in 110 s"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: built-but-dark and stale cards were forgotten, no schedule view. Done when: the plan view runs on today's board.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 To Do status and lane/due/epic keys checked by check.py with tests
- [x] #2 plan.py plan view and --write page with tests
- [x] #3 howto section Planning lanes and migration rule R1-R6
<!-- AC:END -->
