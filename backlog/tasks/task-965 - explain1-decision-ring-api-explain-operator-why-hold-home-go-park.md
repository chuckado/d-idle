---
id: TASK-965
title: 'explain1: decision ring, /api/explain, operator why hold home go park'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:49:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 455000
card: "explain1"
owns:
  - "internal/bot/explain.go"
  - "internal/bot/explain_test.go"
  - "internal/watch/explainapi.go"
  - "internal/watch/explainapi_test.go"
  - "internal/ops"
  - "cmd/d2bot/opscmd.go"
  - "cmd/d2bot/opscmd_test.go"
  - "cmd/d2bot/opswire.go"
  - "docs/operator.md"
  - "cmd/d2bot/ladder.go"
  - "cmd/d2bot/questwire.go"
  - "internal/bot/loop.go"
  - "internal/bot/runner.go"
  - "internal/bot/shadow.go"
  - "internal/watch/http_view.go"
  - "internal/watch/hub.go"
exclusive: []
contracts: []
result: "decision ring, /api/explain and operator why/hold/home/go/park built and unit-tested offline; not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ring of 200 decisions with standing refusals
- [x] #2 /api/explain read-only
- [x] #3 operator verbs why hold home go park, audited, tested on 2 shapes
- [x] #4 docs/operator.md updated
<!-- AC:END -->
