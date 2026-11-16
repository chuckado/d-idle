---
id: TASK-549
title: 'lootwire2: loot rules wave 2a (engine switch, store+API, Loot Rules tab)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T16:30:00Z'
updated_date: '2026-11-16T17:52:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 290000
card: "lootwire2"
owns:
  - "internal/lootrules"
  - "internal/loot/rulesengine.go"
  - "internal/loot/rulesengine_test.go"
  - "internal/watch/lootrulesapi.go"
  - "internal/watch/lootrulesapi_test.go"
  - "internal/loot/lootshadow.go"
  - "internal/loot/pickup.go"
  - "internal/loot/loot.go"
  - "internal/town/tidy.go"
  - "internal/town/tidyengine_test.go"
  - "internal/env/tuning_loot.go"
  - "internal/watch/lootrulesui_test.go"
  - "internal/watch/web/lootrulesui.js"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/shell.css"
  - "internal/watch/funwire.go"
  - "internal/watch/http_view.go"
  - "cmd/d2bot/lootruleswire.go"
  - "cmd/d2bot/lootruleswire_test.go"
  - "cmd/d2bot/ladder.go"
  - "tools/uishots"
  - "tools/uishots.sh"
  - "docs/knobs.md"
  - "docs/lootrules.md"
exclusive: []
contracts: []
result: "knob loot.rulesEngine (0 off, 1 ground, 2 ground+tidy, 3 reserved for sell, not implemented) with a per-decision fallback; rules store, GET/PUT /api/lootrules, Starter and per-class presets with four knobs; Loot Rules sheet in the watch UI; docs/lootrules.md section 12 with the five sell kinds and recommended defaults. Offline only, NOT verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: wave 1 showed the engine equal to today's pickit on the ground and in the pack; wave 2a lets the user switch those two stages over (default off, per-decision fallback), edit the rule set through a store and API, and see why items were kept or sold. The sell stage stays on today's code until the user answers the five sell kinds (docs/lootrules.md section 12.4). Done when: the knob, store, API and sheet exist with tests, the gates pass, and the five questions are in the report.

<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 knob loot.rulesEngine (0/1/2, 3 reserved) switches ground and tidy to the engine with fallback
- [x] #2 rules store + GET/PUT /api/lootrules
- [x] #3 Loot Rules tab in the watch UI
- [x] #4 docs/lootrules.md Wave 2a + five sell kinds
- [x] #5 tests and gates pass
<!-- AC:END -->
