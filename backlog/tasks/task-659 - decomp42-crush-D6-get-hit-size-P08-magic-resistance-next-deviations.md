---
id: TASK-659
title: 'decomp42: crush D6 get-hit size, P08, magic resistance, next deviations'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-19T14:32:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 326000
card: "decomp42"
owns:
  - "docs/research/decomp42-crush-d6-trace.md"
  - "docs/upstream/bodies/crush-d6-body.md"
  - "tools/emu/d42_crush_hit.py"
exclusive:
  - "fork-branch:fid/crush-d6-r1"
contracts: []
result: "PENDING"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the next upstream candidates after decomp41 (crush D6, P08, magic resistance, the scan). Consumer: the independent verifier of fid/crush-d6-r1 (coordinator schedules). Done when: trace doc, oracle numbers and the local branch exist; nothing posted.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 trace doc: D6 rule with addresses and tags, d2d and bot compared
- [ ] #2 oracle differential of the get-hit size after a Crushing Blow against game.exe (V run, 100k cases, 3 seeds, control)
- [ ] #3 P08 curse_put re-run on current upstream main decided
- [ ] #4 T3 magic resistance decided by player-visible impact
- [ ] #5 scan of the next most player-visible deviations
<!-- AC:END -->
