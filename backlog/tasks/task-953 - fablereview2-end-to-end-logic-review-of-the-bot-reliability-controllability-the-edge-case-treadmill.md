---
id: TASK-953
title: 'fablereview2: end-to-end logic review of the bot (reliability, controllability, the edge-case treadmill)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-26T03:17:00Z'
labels:
  - docs
  - research
dependencies: []
priority: high
ordinal: 437000
card: "fablereview2"
owns:
  - "docs/research/fable-logic-review-2026-11-25.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the owner asked on 2026-11-25 for an end-to-end review of the bot's logic because the work keeps chasing edge cases. Done when: the report is merged; docs only, no Go change. Consumer: arbitershadow (TASK-671) and the coordinator's next planning round.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 One report docs/research/fable-logic-review-2026-11-25.md answers the six questions with V/U tags and file references
- [ ] #2 Prior art section with a ranked shortlist of at most 8 patterns
- [ ] #3 Ranked proposals (at most 12), the next 5 PRs and the one change to do first
- [ ] #4 leakcheck, contracts check and backlog check pass
<!-- AC:END -->
