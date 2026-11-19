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
result: "Crush D6 built as local private-fork branch fid/crush-d6-r1 (oracle e177207f, claim a35db81c on upstream da98b7f8, pushed to the private remote only): a Crushing Blow life is not in the get-hit size; parent 72,108 to 72,163 of 100,800 cases differ on three seeds, claim 0, control 0; 7 of 9 mutants caught, one equivalent, two disclosed escapes. P08: curse_put on upstream main 0 of 100k puts differ, no candidate. Magic resistance T3: not worth building for Normal Act 1-3. Scan: P07 monster block gate first (blocked on the spawn variant). Offline, no live game; whole-tree ctest not run; verifier pending."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the next upstream candidates after decomp41 (crush D6, P08, magic resistance, the scan). Consumer: the independent verifier of fid/crush-d6-r1 (coordinator schedules). Done when: trace doc, oracle numbers and the local branch exist; nothing posted.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 trace doc: D6 rule with addresses and tags, d2d and bot compared
- [x] #2 oracle differential of the get-hit size after a Crushing Blow against game.exe (V run, 100k cases, 3 seeds, control)
- [x] #3 P08 curse_put re-run on current upstream main decided
- [x] #4 T3 magic resistance decided by player-visible impact
- [x] #5 scan of the next most player-visible deviations
<!-- AC:END -->
