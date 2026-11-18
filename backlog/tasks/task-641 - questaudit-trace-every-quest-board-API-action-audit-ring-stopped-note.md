---
id: TASK-641
title: 'questaudit: trace every quest board API action, audit ring, stopped note'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-18T15:17:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 321000
card: "questaudit"
owns:
  - "internal/questboard/audit.go"
  - "internal/questboard/audit_test.go"
  - "internal/watch/questsaudit.go"
  - "internal/watch/questsaudit_test.go"
  - "docs/questboard.md"
  - "internal/questboard/desk.go"
  - "internal/watch/questsapi.go"
  - "internal/watch/web/panels/quest-board/panel.js"
  - "internal/watch/web/panels/quest-strip/panel.js"
  - "internal/watch/testdata/questui/quest_ui.mjs"
exclusive: []
contracts: []
result: "built and tested offline (go test -race -count=20 V); not deployed, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the accepted Search for Cain vanished on 2026-11-18 with no trace. Done when: every quest API action is logged and audited with a source class, and a stop shows on the board for 24 h.

result: built and tested offline, not deployed, not verified live (Go tests -race -count=20 V).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 INFO log line per quest API action without credentials
- [x] #2 audit ring quests/<hero>.audit.jsonl capped at 500, not in the support bundle
- [x] #3 Stopped by you note on the board for 24 h; two-tap confirm kept and recorded
- [x] #4 tests: log, ring, redaction, cancel record with source class, UI note
<!-- AC:END -->
