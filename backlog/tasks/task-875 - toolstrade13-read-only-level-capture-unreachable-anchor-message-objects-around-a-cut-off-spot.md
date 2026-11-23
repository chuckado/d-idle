---
id: TASK-875
title: 'toolstrade13: read-only level capture, unreachable anchor message, objects around a cut-off spot'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T08:38:00Z'
labels:
  - bot
  - probe
dependencies: []
priority: medium
ordinal: 394000
card: "toolstrade13"
owns:
  - "tools/probes/capture.py"
  - "tools/probes/capture_test.py"
  - "internal/boss/run.go"
  - "internal/boss/boss_test.go"
  - "internal/boss/fake_test.go"
  - "internal/nav/gatedoor.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Read-only capture script, clearer boss-not-found message and an object listing in the cut-off log; the cause of the unreachable Smith anchor on the real engine is NOT found yet (needs the capture or the next log)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Live sandbox 2026-11-23: the Smith anchor (73,267) is 80 subtiles west of the closest reachable point (152,267) on the real engine and no door opens it. Done when merged; the capture is run on the live container by the coordinator.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 tools/probes/capture.py captures one level through read verbs only (allow-list tested)
- [x] #2 A boss not found after a failed walk to its anchor says the spawn point could not be reached; must-fail test, two bosses
- [x] #3 The cut-off log lists the objects nearest to the character and to the spot
<!-- AC:END -->
