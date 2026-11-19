---
id: TASK-668
title: 'provingground: scenario runner, Cain step scenarios, catalogue'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-19T16:26:00Z'
updated_date: '2026-11-19T17:50:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 327000
card: "provingground"
owns:
  - "tools/provingground/"
  - "internal/scenario/"
  - "cmd/d2scenario/"
  - "docs/process/proving-ground.md"
exclusive:
  - "d2d-lab-binary"
contracts: []
result: "Proving Ground built (internal/scenario, cmd/d2scenario, tools/provingground, docs/process/proving-ground.md): scenario format, setup builder, runner on the builder long lane, dry-run gate plan, determinism checker, behaviour catalogue and coverage report. Verified live on the builder: the Cain tree step reproduces the sandbox stall (cain.scroll.field RED, still exploring for class 30, Dark Wood / Cold Plains cycle); cain.scroll from the camp is RED because the chunk stays queued behind fights. The same-seed live determinism pair, the sandbox d2d variant, cain.read, cain.report and the stash-gold scenario were not run live. No bot behaviour or gate change."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the Search for Cain tree stall lasted 17 hours in the sandbox because nothing could play a quest step in isolation and read the result from game state (design review 2026-11-19, R1 and R2). Done when: scenarios (steps and behaviours) run the real bot on a lockstep lab slot on the builder long lane and are judged from game state; a catalogue maps every bot behaviour to scenarios and lists the uncovered; docs/process/proving-ground.md says how to write, run and gate. Test tooling only: no bot behaviour change, no merge gate change.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Scenario format, runner, determinism check, dry-run gate plan
- [x] #2 Cain tree step scenario red or its difference explained; Cain chain skeletons
- [x] #3 Behaviour catalogue and coverage report
- [x] #4 docs/process/proving-ground.md
<!-- AC:END -->
