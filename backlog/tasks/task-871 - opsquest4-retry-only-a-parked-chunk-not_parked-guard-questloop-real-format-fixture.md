---
id: TASK-871
title: 'opsquest4: retry only a parked chunk (not_parked guard), questloop real-format fixture'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T07:57:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 393000
card: "opsquest4"
owns:
  - "internal/ops/**"
  - "cmd/d2bot/ops*.go"
  - "docs/operator.md"
  - "tools/questloop.sh"
  - "tools/questloop_test.sh"
exclusive: []
contracts: []
result: "Bot-side not_parked guard on operator retry (deploy needed), questloop retries only a blocked chunk with a park code after a second read, real status format in the fake-docker test. Not run live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the live loop called retry on an in_progress chunk. Done when: the criteria hold.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ops retry refuses not_parked unless a matching chunk is blocked (must-fail test)
- [x] #2 questloop retries only on blocked with a park code; test fixture uses the real status format
- [x] #3 docs/operator.md updated
<!-- AC:END -->
