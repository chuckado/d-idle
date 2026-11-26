---
id: TASK-1008
title: 'stepexec2: talk and use adapters, Akara bark reading on the executor'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T17:56:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 482000
card: "stepexec2"
lane: next
owns:
  - "internal/questspec/exec.go"
  - "internal/questspec/exec_test.go"
  - "internal/bot/stepexec.go"
  - "internal/bot/stepexec_test.go"
  - "internal/bot/task.go"
  - "data/questspec/act1.json"
  - "tools/provingground/scenarios/cain.json"
  - "data/darklaunch.json"
  - "docs/design/stepexec-p9.md"
exclusive: []
contracts: []
result: "The step executor has TALK and USE adapters; Akara reading the bark scroll (search_for_cain/read_scroll) runs on it behind quests.stepExec with scenario cain.read.stepexec; unit tests on two classes with a must-fail case. Not run on the lab or live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

P9 remainder (Fable review 2026-11-25).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 TALK and USE adapters in the step executor
- [x] #2 second chunk (Akara reads the bark) behind quests.stepExec with a scenario row and must-fail tests on two classes
<!-- AC:END -->
