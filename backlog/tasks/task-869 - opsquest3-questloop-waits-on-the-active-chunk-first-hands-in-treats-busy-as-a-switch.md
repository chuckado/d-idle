---
id: TASK-869
title: >-
  opsquest3: questloop waits on the active chunk first, hands in, treats busy as
  a switch
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T07:36:00Z'
updated_date: '2026-11-23T08:00:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 392000
card: "opsquest3"
owns:
  - "docs/operator.md"
  - "tools/questloop.sh"
  - "tools/questloop_test.sh"
exclusive: []
contracts: []
result: "questloop waits on the active chunk first, uses handin for awaiting_handin, busy switches instead of counting a try; fake-docker test; not run live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the live loop went to blood_raven while another chunk ran and burned tries on busy refusals. Done when: the criteria hold.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 questloop waits on the accepted/in_progress chunk before any other quest; busy refusal switches, not a try
- [x] #2 awaiting_handin uses handin, only when nothing else is active; decisions logged
- [x] #3 tests and docs/operator.md updated
<!-- AC:END -->
