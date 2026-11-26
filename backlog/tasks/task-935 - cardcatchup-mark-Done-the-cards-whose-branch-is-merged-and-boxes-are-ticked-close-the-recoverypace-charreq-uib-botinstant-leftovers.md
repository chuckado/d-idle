---
id: TASK-935
title: 'cardcatchup: mark Done the cards whose branch is merged and boxes are ticked'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-26T01:48:00Z'
labels:
  - docs
  - chore
dependencies: []
priority: medium
ordinal: 436000
card: "cardcatchup"
owns:
  - "backlog/tasks"
  - "docs/BACKLOG.md"
exclusive: []
contracts: []
result: "catchup_done.py --apply: 224 cards set Done (222 by the backlog CLI, TASK-393 and TASK-597 by hand because the CLI could not find them); check.py and contracts check clean. Not a live matter."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: 224 cards had merged branches and all boxes ticked but stayed open (pipeline audit finding 2). Done when: they are Done.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 catchup_done.py --apply run, check.py clean
<!-- AC:END -->
