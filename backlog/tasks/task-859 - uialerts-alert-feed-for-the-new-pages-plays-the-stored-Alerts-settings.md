---
id: TASK-859
title: 'uialerts: alert feed for the new pages plays the stored Alerts settings'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T05:43:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 389000
card: "uialerts"
owns:
  - "internal/watch/web/alerts-feed.js"
  - "internal/watch/testdata/alertsfeed_test.cjs"
  - "internal/watch/settingsjs_test.go"
  - "internal/watch/web/nav.js"
  - "internal/watch/http_view.go"
  - "docs/ui-design-system.md"
exclusive: []
contracts: []
result: "alerts-feed.js polls /api/alerts only while sound is on or notifications are granted; toast + cue + hidden-tab notification; node tests pass; not live-checked in a browser; skipped where the old page (#alertspane) runs."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 feed off with both settings off; toast, cue and notification when on (node tests)
<!-- AC:END -->
