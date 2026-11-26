---
id: TASK-998
title: 'flipwave2: flip arbiterControl and stepExec on one sandbox, 24 h watch'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T17:40:00Z'
labels:
  - ops
  - drill
dependencies: []
priority: high
ordinal: 482000
card: "flipwave2"
lane: next
epic: fableroadmap
owns:
  - "docs/probes/flipwave2.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Both knobs are built dark with ledger rows and no card owns their flip (arbmig and stepexec1 own the code). Done when: both flips have a recorded 24 h outcome from tools/flip/flip.py.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 arbiterControl families survive and player: arbiter-replay verdict and incident rows green, then tools/flip/flip.py on ONE sandbox with bot.arbiterShadow 1, 24 h watch result recorded (CLOSE or REVERT line)
- [ ] #2 quests.stepExec: scenario cain.scroll.stepexec green 2 of 3, then flip on one sandbox, 24 h watch result recorded
- [ ] #3 ledger rows set by the tool (trial, enabled or dark with revert_reason); no box ticked without the tool's line
<!-- AC:END -->
