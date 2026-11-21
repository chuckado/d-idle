---
id: TASK-749
title: 'mglanes: merge class, agent quotas, whole-repo refusal, wait budget, load guard'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-21T04:50:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 353000
card: "mglanes"
owns:
  - "docs/process/pg-ops-mglanes.md"
exclusive: []
contracts: []
result: "Builder lane wrapper (ops repo) enforces per-agent quotas (main 1 running + 1 waiting, long 2), refuses whole-repo go vet/test from agent worktrees, gives merge-class jobs priority (plus a 2 min grace), cancels agent jobs queued over 20 min, delays a merge gate under load (marks contended after 10 min) and runs two long slots behind the memory guard (second slot needs 4500 MB). Warn-only for 24 h after the install, then enforced; offline drill test_lockrun.sh 14 checks; one live main-lane job verified queueing behind a real merge. Not run: enforcement phase."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design 4.5, 7 card 13. Private ops repo work. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Quotas: 1 running + 1 waiting main-lane job per agent, 2 on the long lane, no whole-repo from a branch (Q5)
- [x] #2 long lane two slots behind the memory guard (Q9)
- [x] #3 load guard and wait budget
<!-- AC:END -->
