---
id: TASK-845
title: 'buildchoose: Choose sheet in plain words, default plan, Advanced variants'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T03:43:00Z'
updated_date: '2026-11-23T04:13:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 384000
card: "buildchoose"
owns:
  - "internal/plandata/build-catalog.json"
  - "internal/plandata/catalog.go"
  - "internal/plandata/catalog_test.go"
  - "internal/points/buildcards.go"
  - "internal/points/buildcards_test.go"
  - "cmd/d2bot/autoallocwire_test.go"
  - "internal/watch/web/sheet.js"
  - "internal/watch/web/buildcards.js"
  - "internal/watch/web/points.js"
  - "internal/watch/web/components.css"
  - "internal/watch/testdata"
  - "internal/watch/buildcards_test.go"
  - "tools/uishots.sh"
  - "docs/ui/build-choose"
  - "docs/builds-catalog.md"
  - "docs/ui-design-system.md"
exclusive: []
contracts: []
result: "Catalogue variants (label, why, defaultVariant, fromLevel), plain Choose sheet with collapsed Advanced, tests and fit pages at 390x844 and 1366x768 verified (fit run on the worker host, not on a real phone). No automatic table switch at level 30 exists in the bot; not added."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: user feedback 2026-11-23: the Choose sheet showed raw plan table ids as radios. Done when: one paragraph, collapsed Advanced, catalogue variants with plain text.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Catalogue builds carry defaultVariant and plain labels; contract tests pass
- [x] #2 Choose sheet shows one paragraph, no radio by default, Advanced lists plans
- [x] #3 DOM golden, node audit and fit pages builds-choose pass at 390x844 and 1366x768
<!-- AC:END -->
