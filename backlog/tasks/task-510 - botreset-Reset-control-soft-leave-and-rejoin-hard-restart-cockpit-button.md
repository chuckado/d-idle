---
id: TASK-510
title: 'botreset: Reset control (soft leave-and-rejoin, hard restart, cockpit button)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T16:59:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 278000
card: "botreset"
owns:
  - "internal/bot/playerreset.go"
  - "internal/bot/playerreset_test.go"
  - "internal/watch/resetapi.go"
  - "internal/watch/resetapi_test.go"
  - "internal/watch/resetui_test.go"
  - "internal/watch/web/resetui.js"
  - "cmd/d2bot/resetwire.go"
  - "cmd/d2bot/resetwire_test.go"
  - "cmd/d2bot/glancewire_reset.go"
  - "docs/reset.md"
  - "internal/watch/testdata/reset"
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
- [ ] #1 GET/POST /api/reset with goldens, refusals, double press
- [ ] #2 soft and hard levels, state kept across a reset
- [ ] #3 cockpit confirm sheet, strip, menu, More item, Needs you chip
- [ ] #4 fit fixtures at 8 sizes and the real phone sizes
- [ ] #5 docs/reset.md with the live checklist
<!-- AC:END -->
