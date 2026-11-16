---
id: TASK-542
title: 'support: send a support bundle to the updates site'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T04:02:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 286000
card: "support"
owns:
  - "internal/support"
  - "internal/watch/supportapi.go"
  - "internal/watch/supportapi_test.go"
  - "internal/watch/web/supportui.js"
  - "internal/watch/web/support.css"
  - "cmd/d2supportrecv"
  - "cmd/d2bot/supportwire.go"
  - "docs/support-bundle.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Bundle builder, redaction, fail-closed scan, tests
- [ ] #2 Receiver (cmd/d2supportrecv) with auth, quota, purge, traversal tests; README, nginx snippet, deploy steps
- [ ] #3 Watch API + UI sheet + node tests + fit fixtures
- [ ] #4 Docs: support-bundle.md, friend guide section, knobs
- [ ] #5 Live check: real upload on the updates site (owner, after deploy)
<!-- AC:END -->
