---
id: TASK-743
title: 'pgledger: results ledger: JSONL record, ingest command, results written on the board host (no pull loss)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:48:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 349000
card: "pgledger"
owns:
  - "cmd/d2scenario/ledger.go"
  - "internal/scenario/ledger.go"
  - "docs/process/pg-ledger-ops.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design section 5, 7 card 9 (Q6: ledger on the board host next to the status files, one generated status doc in the repo). Ops-repo part is a commit on a branch of the ops repo. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Record per design 5.1: run id, verdict, D1, D2, host role, commit, setup tag EARNED/SETUP/CHEATED/IMPOSSIBLE, history
- [ ] #2 runner appends at the end of exec-one through one ingest command; append-only
- [ ] #3 ingest and board-host side are an ops-repo change, described by role
- [ ] #4 must-fail test: a run without provenance is refused
<!-- AC:END -->
