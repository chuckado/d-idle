---
id: TASK-481
title: 'uifix1: cockpit review fixes R1 to R13 and the missing fit fixtures'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T07:31:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 269000
card: "uifix1"
owns:
  - "internal/watch/web/panels/statusline/"
  - "internal/watch/web/panels/needs-you/"
  - "internal/watch/web/panels/health/"
  - "internal/watch/web/panels/journey-strip/"
  - "internal/watch/web/panels/orders/"
  - "internal/watch/web/panels/zone-map/"
  - "internal/watch/web/panels/area-map/panel.json"
  - "internal/watch/web/shell.css"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/watch.html"
  - "internal/watch/whyline.json"
  - "internal/watch/whychange.json"
  - "internal/watch/whyline_test.go"
  - "internal/watch/testdata/"
  - "cmd/d2bot/glancewire.go"
  - "cmd/d2bot/glancewire_pause.go"
  - "cmd/d2bot/glancewire_pause_test.go"
  - "tools/uishots/"
  - "tools/uishots.sh"
  - "internal/bot/whyline_test.go"
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
- [x] #1 R1 to R13 of docs/ui/redesign-2026-11-14/reviews/cockpit-review.md fixed
- [x] #2 Fixtures for paused, Loop row, dead mercenary, low supplies and the live picture tile; tools/uishots.sh --fit passes at the eight sizes
- [x] #3 Renders at 1280x620, 1366x768, 1440x900, 1914x930 and the phone looked at
<!-- AC:END -->
