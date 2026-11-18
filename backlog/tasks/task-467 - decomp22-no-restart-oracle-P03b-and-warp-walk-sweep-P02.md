---
id: TASK-467
title: 'decomp22: no-restart oracle (P03b) and warp walk sweep (P02)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T04:25:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 265000
card: "decomp22"
owns:
  - "docs/research/emu-runs-2026-11-14b.md"
  - "docs/research/warp-walk-trace.md"
  - "docs/research/data22"
  - "tools/emu/mode_switch.py"
  - "tools/emu/dc22_dis.py"
  - "docs/upstream/pipeline-2026-11-14.md"
exclusive: []
result: "P03b game side CONFIRMS the claim (V run: parent 1 of 12 cells differ, claim 0; fork branch fid/hitrec-norestart-r2 ready for the verifier). P02 S-1: stall persists on upstream 4e054f43 (83.3 percent of starts take the warp), shim loses 97 starts, game side untraced, no patch. Not verified live."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the P03b claim and P02 lacked a game.exe column. Done when: both answered in docs/research/emu-runs-2026-11-14b.md. Consumer: P03b and P02 rows of docs/upstream/pipeline-2026-11-14.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 P03b: game-side oracle run, compare before/after
- [x] #2 P02: S-1 sweep on upstream/main and trace notes
<!-- AC:END -->
