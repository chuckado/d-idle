---
id: TASK-555
title: 'lootwire3: loot rules wave 2b (sell stage, preview, undo, import/export, per-hero store)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T17:53:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 292000
card: "lootwire3"
owns:
  - "internal/lootrules"
  - "internal/loot/rulesengine.go"
  - "internal/loot/rulesengine_test.go"
  - "internal/loot/lootshadow.go"
  - "internal/loot/lootshadow_test.go"
  - "internal/loot/rulesadapt.go"
  - "internal/town/shop.go"
  - "internal/watch/lootrulesapi.go"
  - "internal/watch/lootrulesapi_test.go"
  - "internal/watch/lootrulesui_test.go"
  - "internal/watch/web/lootrulesui.js"
  - "internal/watch/web/shell.css"
  - "internal/charsel/perchar.json"
  - "cmd/d2bot/lootruleswire.go"
  - "cmd/d2bot/lootruleswire_test.go"
  - "tools/uishots"
  - "docs/lootrules.md"
  - "docs/knobs.md"
  - "internal/env/tuning_loot.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the user accepted all five sell recommendations (2026-11-16); wave 2a left the sell stage and several controls open. Done when: stage 3 exists behind the knob (default off) and the controls work, with tests and gates.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 loot.rulesEngine 3 = ground+tidy+sell with the five sell decisions as rules and engine behaviour, fallback guard, shadow test that fails on any new disagreement
- [ ] #2 preview (dry run), undo, import/export routes, per-hero store
- [ ] #3 Loot Rules sheet controls and plain-language sell kinds, fit pictures looked at
- [ ] #4 docs/lootrules.md Wave 2b, tests and gates
<!-- AC:END -->
