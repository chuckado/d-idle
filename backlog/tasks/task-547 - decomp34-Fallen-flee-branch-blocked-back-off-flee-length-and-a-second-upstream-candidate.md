---
id: TASK-547
title: 'decomp34: Fallen flee branch (blocked back-off, flee length) and a second upstream candidate'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T16:07:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 289000
card: "decomp34"
owns:
  - "docs/research/decomp34-fallen-flee.md"
  - "tools/emu/dc34_ring.py"
  - "tools/emu/dc34_flee.py"
  - "tools/emu/dc34_cmp.py"
  - "tools/emu/dc34_telemana.py"
  - "tools/emu/dc34_tmcmp.py"
  - "docs/upstream/bodies/"
  - "backlog/merge-lines/decomp34.txt"
exclusive:
  - "fork-branch:fid/tele-refund-r1"
contracts: []
result: "Fallen flee scan: blocked back-off = no deviation (1,122 of 1,122 real-think cases); the scan reads the last 4 deaths per near room of any class, d2d every DT monster: 24 of 4,000 differ (ring overflow), no PR. P13 refused Teleport mana: fid/tele-refund-r1 (head 0a439860, claim 4386559e), 927 and 970 of 2,000 differ on the parent, 0 on the claim, ready for the verifier. Offline, V run on the builder emulator and build; no live game; nothing posted."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Done when: trace doc, oracle, fork branch if deviation proven.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fallen flee branch traced and differential run
- [x] #2 second candidate traced
- [x] #3 result written
<!-- AC:END -->
