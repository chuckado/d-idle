---
id: TASK-854
title: 'uisettings: Settings sheet (Rules, Alerts) in More, review E10'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T05:07:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 387000
card: "uisettings"
owns:
  - "internal/watch/web/settings.js"
  - "internal/watch/web/nav.js"
  - "internal/watch/web/nav.json"
  - "internal/watch/web/nav.css"
  - "internal/watch/web/sheet.js"
  - "internal/watch/settingsjs_test.go"
  - "tools/uishots/"
  - "tools/uishots.sh"
  - "docs/ui/settings/"
  - "docs/ui-design-system.md"
exclusive: []
contracts: []
result: "Settings sheet (Rules: revoke one/all; Alerts: sound, desktop notifications) in More on every page; node DOM tests + Go tests pass; fit pages settings-rules/-alerts pass on Orion at 390x844 and 1366x768 (fixture data). Not live-verified against a real answer route. Gap: new pages have no alert feed that plays the stored settings."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Settings sheet with Rules and Alerts opens from More on every page
- [x] #2 node DOM tests and Go nav test pass
- [x] #3 uishots fit pages settings-rules and settings-alerts pass at 390x844 and 1366x768
- [x] #4 dropped features listed
<!-- AC:END -->
