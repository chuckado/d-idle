---
id: TASK-518
title: 'decomp33: upstream candidates P17 sequence rate, P18 Smite, P19 Holy Bolt heal'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T23:59:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 281000
card: "decomp33"
owns:
  - "docs/research/decomp33-p17-p19.md"
  - "tools/emu/dc33_seqrate.py"
  - "tools/emu/dc33_smite.py"
  - "tools/emu/dc33_holybolt.py"
  - "docs/upstream/bodies/seq-ias-body.md"
  - "docs/upstream/bodies/smite-range-body.md"
  - "docs/upstream/briefs/verify-seq-ias.md"
  - "docs/upstream/briefs/verify-smite-range.md"
  - "backlog/merge-lines/decomp33.txt"
  - "docs/upstream/pipeline-2026-11-14.md"
  - "docs/upstream/pr-queue.md"
  - "docs/decomp-queue.md"
exclusive:
  - "fork-branch:fid/seq-ias-r1"
  - "fork-branch:fid/smite-r1"
  - "fork-branch:fid/holybolt-heal-r1"
contracts: []
result: "P17 sequence rate: fid/seq-ias-r1 (oracle 320e3491, claim e177a4fa), 7,600 cells parent 6,941 differ claim 0, ready for the verifier. P18 Smite range: fid/smite-r1 (oracle 486dc3bd, claim 772458b6), 75,600 cells parent 63,709 differ claim 0, ready for the verifier. P19 Holy Bolt heal: d2d matches game.exe for pets (2,240 cells, 0 differ), no PR, oracle-only fid/holybolt-heal-r1 (21e0afb5). Offline, V run on the builder emulator; no live game; nothing posted."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the user went ahead with upstream candidates P17, P18, P19 (2026-11-15). Done when: each is re-derived against current upstream main with game.exe in the emulator and has either a fork branch ready for the verifier or a no-PR verdict.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 P17: oracle then claim on the fork, compare 0 after, mutant aborts, checks
- [x] #2 P18: re-derived against upstream main; branch or no-PR verdict
- [x] #3 P19: re-derived against upstream main; branch or no-PR verdict
- [x] #4 bodies, verifier briefs, pipeline and pr-queue rows, leakcheck clean
<!-- AC:END -->
