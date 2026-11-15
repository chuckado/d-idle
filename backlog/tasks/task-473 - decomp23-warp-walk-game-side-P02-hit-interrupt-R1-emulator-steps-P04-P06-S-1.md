---
id: TASK-473
title: 'decomp23: warp walk game side (P02), hit interrupt R1 emulator steps (P04), P06 S-1'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T05:01:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 266000
card: "decomp23"
owns:
  - "docs/research/warp-walk-game-trace.md"
  - "tools/emu/ww_path.py"
  - "tools/emu/ww_search.py"
  - "tools/emu/ww_used.py"
  - "tools/emu/ww_dis.py"
  - "tools/emu/ww_run.sh"
  - "tools/emu/ww_outcome.py"
  - "docs/upstream/drafts/warp-walk-plan.md"
  - "docs/upstream/bodies/warp-walk-plan-body.md"
  - "docs/upstream/briefs/verify-warp-walk-r2.md"
  - "docs/upstream/pipeline-2026-11-14.md"
  - "docs/upstream/pr-queue.md"
exclusive: []
result: "P02: d2d's plan as a unit target equals game.exe's FUN_00649970 path in 5,857 of 5,857 starts (V run); the stall is game.exe's own rule on d2d's grid; the fork shim differs from game.exe in 319 starts (not a fix); d2d planned a warp as a spot: outcome differs in 30 starts (one of 324 combinations), fixed on fid/warp-walk-r2 head 3b4439b3 (30 before, 0 after), ready for the verifier. P04 R1: the used-skill pointer is cleared by idle, walk, run and the swing end (V run, 0 failures), so it does not outlive a swing. P06 S-1: the block gate is still open on upstream 4e054f43 (V read). Not verified live; emulator and NAS builder only."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: P02 had no game.exe side (decomp22), P04 was blocked on the used-skill lifetime, P06 needed its upstream re-check. Done when: the trace doc holds the answers with tags. Consumer: docs/upstream/pipeline-2026-11-14.md rows P02, P04, P06 and the verifier brief docs/upstream/briefs/verify-warp-walk-r2.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Warp walk: game.exe side traced and oracle run against the fork sweep (V run), deviation list written
- [x] #2 P04 R1: emulator-answerable steps run, V/U per step
- [x] #3 P06 S-1: upstream re-check recorded
<!-- AC:END -->
