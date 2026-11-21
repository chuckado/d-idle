---
id: TASK-787
title: 'autoalloc: Automatic Allocate Skills slice, respec knobs, currency design'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T19:18:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 361000
card: "autoalloc"
owns:
  - "docs/design/autoalloc-and-currencies-2026-11-21.md"
  - "docs/autoalloc.md"
  - "internal/points/"
  - "internal/leveler/autoalloc.go"
  - "internal/leveler/autoalloc_test.go"
  - "cmd/d2bot/autoallocwire.go"
  - "cmd/d2bot/autoallocwire_test.go"
  - "internal/env/tuning_respec.go"
  - "internal/env/tuning_agency.go"
  - "internal/env/tuning.go"
  - "docs/knobs.md"
  - "internal/watch/pointsapi.go"
  - "internal/watch/pointsapi_test.go"
  - "internal/watch/pointsui.go"
  - "internal/watch/web/points.js"
  - "internal/watch/web/points.css"
  - "tools/provingground/scenarios/autoalloc.json"
  - "tools/provingground/scenarios/behaviours.json"
  - "internal/leveler/leveler.go"
  - "internal/watch/funwire.go"
  - "cmd/d2bot/pointswire.go"
exclusive: []
contracts: []
result: "Done offline, NOT verified live. Design doc (feature spec, respec model, currency inventory and gold/rune consolidation, 14 cards, 10 open questions). Built: build lock in <hero>.points.json, GET/POST /api/hero/build, allocator behind agency.autoAlloc (default 0) closing the leveler gate while locked, Points page picker and locked state, respec.* price knobs with a live preview on the Points page (placeholders, deliberately undecided), 3 Proving Ground scenarios written (points.autoalloc-follows-build, -sorc-fire, points.autoalloc-refuses-locked-switch) but NOT run: no lab slot was used. Unit tests on Paladin and Sorceress, -race -count=20 green on points, watch, leveler. Respec pay is disabled until the fork card (cmd respec). The dashboard's generic knob screen shows the respec knobs but has no live price line (card C11)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the user wants an Automatic Allocate Skills button with a locked build and a gold respec, and gold and runes as the meta currencies. Done when: the design doc exists and the first slice (lock, allocator, picker, respec knobs) is merged with the knob off.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Design doc with feature spec, respec, currency inventory, cards and open questions
- [x] #2 Build lock stored in the hero plan file, GET and POST /api/hero/build
- [x] #3 Auto-allocator follows the locked build behind agency.autoAlloc (default 0)
- [x] #4 Points page picker, locked state, respec price preview, respec knobs respec.*
- [x] #5 Proving Ground scenarios written (run on a lab slot: not run yet)
- [x] #6 Unit tests on Paladin and Sorceress, race count 20 on touched packages
<!-- AC:END -->
