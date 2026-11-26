---
id: TASK-995
title: 'cardclose2: close merged cards whose boxes or status were stale'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-26T05:37:00Z'
labels:
  - docs
  - chore
dependencies: []
priority: medium
ordinal: 480000
card: "cardclose2"
lane: next
owns:
  - "backlog/tasks/task-*"
exclusive: []
contracts: []
result: "Closed 7 merged cards: 954 and 966 by catchup_done --apply; 953, 955, 959, 962, 934 by hand (status Done, boxes ticked where git log on origin/main shows the work). Not closed: 957, 958, 960, 961, 963 (outcome follow-ups, no merge), 965 explain1 (no merge), toolstrade78 and 79 (queued), 9 and 138 (open boxes). Docs and cards only, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 catch-up run and manual close of merged cards 953 955 959 962 934 applied
- [x] #2 backlog and contracts checks pass
<!-- AC:END -->
