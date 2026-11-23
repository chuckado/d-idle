---
id: TASK-866
title: 'opsquest2: questloop follows slow progress, backs off, runs to a deadline'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T07:04:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 391000
card: "opsquest2"
owns:
  - "internal/ops/**"
  - "cmd/d2bot/ops*.go"
  - "docs/operator.md"
  - "tools/questloop.sh"
  - "tools/questloop_test.sh"
exclusive: []
contracts: []
result: "questloop follows slow progress (progress counter), budget, deadline, 20 min backoff after tries, heartbeat, idempotent restart, --all continues; fake-docker test; not run live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the first live run treated a slow in-progress chunk as failed (coordinator report 2026-11-23). Done when: the criteria hold.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ops status exposes a per-chunk progress counter; wait prints it
- [x] #2 questloop: slow progress is not a failed try, total budget, backoff after tries, deadline, heartbeat, idempotent restart, --all continues
- [x] #3 docs/operator.md updated
<!-- AC:END -->
