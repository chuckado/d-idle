---
id: TASK-450
title: 'verifyp03b: second verifier of P03 block length after fixes'
status: In Progress
assignee:
  - live-agent
created_date: '2026-11-14T22:22:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 260000
card: "verifyp03b"
owns:
  - "docs/upstream/verifier2-p03-2026-11-14.md"
exclusive: []
contracts: []
result: "partial, not verified: run stopped while the head build was running on the NAS builder; no verdict doc yet"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the P03 fixes (agent/prfix2, fork head 3264d61d) need an independent second check before posting. Consumer: the P03 upstream PR draft (docs/upstream/drafts/hitrec-modes.md). Done when the verdict doc is committed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Rebuild oracle, head and fix-reverted mutant on the NAS builder; rerun the tests and the before/after differential
- [ ] #2 Game-side oracle and closed form checked against game.exe and the HITREC trace
- [ ] #3 clang-tidy, docs, body, identity, scope and regression risk checked
- [ ] #4 Verdict doc docs/upstream/verifier2-p03-2026-11-14.md written
<!-- AC:END -->
