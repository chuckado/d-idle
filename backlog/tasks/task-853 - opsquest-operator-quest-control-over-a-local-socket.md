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
  - "cmd/d2bot/main.go"
  - "cmd/d2bot/questwire.go"
  - "internal/ops/**"
  - "internal/questboard/audit.go"
  - "docs/operator.md"
  - "tools/questloop.sh"
  - "tools/questloop_test.sh"
exclusive: []
contracts: []
result: "Operator quest control built: unix socket 0600 (D2BOT_OPERATOR=1, never in a measured run), allowlisted quest verbs, audited with remote operator, rate limited, ops CLI with wait, tools/questloop.sh with a fake-docker test. Offline only; not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the sandbox needs quest retries and tests without the viewer passcode (user decision 2026-11-23). Done when: the criteria hold; the coordinator enables it on the sandbox after merge.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 off by default; socket 0600; quest verbs only; audited; rate limited
- [x] #2 d2bot ops quest CLI incl. wait
- [x] #3 tools/questloop.sh with fake-docker test
- [x] #4 docs/operator.md
<!-- AC:END -->
