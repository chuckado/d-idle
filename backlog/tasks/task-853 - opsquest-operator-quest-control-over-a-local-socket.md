---
id: TASK-853
title: 'opsquest: operator quest control over a local socket'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T05:05:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 387000
card: "opsquest"
owns:
  - "cmd/d2bot/ops*.go"
  - "internal/ops/**"
  - "internal/questboard/audit.go"
  - "docs/operator.md"
  - "tools/questloop.sh"
  - "tools/questloop_test.sh"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 off by default; socket 0600; quest verbs only; audited; rate limited
- [ ] #2 d2bot ops quest CLI incl. wait
- [ ] #3 tools/questloop.sh with fake-docker test
- [ ] #4 docs/operator.md
<!-- AC:END -->
