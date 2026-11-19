---
id: TASK-673
title: 'agencyp0: Agency P0 stat and skill points by verb, sheet, autopilot switch, legality'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-19T18:45:00Z'
labels:
  - bot
  - feature
  - T0
dependencies: []
priority: medium
ordinal: 329000
card: "agencyp0"
owns:
  - "internal/points"
  - "internal/watch/pointsapi.go"
  - "internal/watch/pointsapi_test.go"
  - "internal/watch/pointsui.go"
  - "internal/watch/web/points.js"
  - "internal/watch/web/points.css"
  - "internal/watch/web/skills.js"
  - "internal/watch/web/watch.html"
  - "internal/watch/skillsapi.go"
  - "internal/watch/funwire.go"
  - "cmd/d2bot/pointswire.go"
  - "cmd/d2bot/pointswire_test.go"
  - "cmd/d2bot/gearpolicy.go"
  - "cmd/d2bot/behaviors.go"
  - "cmd/d2bot/glancewire.go"
  - "cmd/d2bot/ladder.go"
  - "cmd/d2bot/d2dbackend.go"
  - "cmd/d2bot/d2dwire.go"
  - "cmd/d2bot/live.go"
  - "internal/bot/pointsgate.go"
  - "internal/bot/pointsgate_test.go"
  - "internal/bot/leveling.go"
  - "internal/bot/runner.go"
  - "internal/combat/playable.go"
  - "internal/leveler/leveler.go"
  - "internal/leveler/build.go"
  - "internal/leveler/verbs_test.go"
  - "internal/d2dctl/points_verbs.go"
  - "internal/d2dctl/points_verbs_test.go"
  - "internal/d2dsrc/points_sheet_test.go"
  - "internal/skilltree/skilltree.go"
  - "internal/power/profiles.json"
  - "internal/power/skilldamage_test.go"
  - "internal/plandata"
  - "docs/plan/data"
  - "internal/env/tuning_resource.go"
  - "internal/resource/tuning.go"
  - "docs/knobs.md"
  - "internal/scenario/state.go"
  - "internal/scenario/scenario_test.go"
  - "cmd/d2scenario/run.go"
  - "cmd/d2scenario/main_test.go"
  - "tools/provingground/scenarios/points.json"
  - "tools/provingground/scenarios/behaviours.json"
  - "docs/points.md"
  - "docs/process/points-verbs-devp14.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Roadmap R3 of the fable review (docs/design/fable-review-2026-11-19/04 section 7, P0; ADR 0020). Spend stat and skill points by the d2d verbs, a points sheet and API, the per-hero autopilot switch (default off: bank and ask), legality fixes, Proving Ground scenarios. Doc: docs/points.md. Done when the checks pass and the branch is merged.

result: P0 built and tested offline, NOT verified live (no game was run; live checklist in docs/points.md section 7). Verbs `cmd stat` and `cmd skillpt` already exist in d2d (no devp14 addition needed, wish list in docs/process/points-verbs-devp14.md). d2d does not enforce the per-level skill cap (V from the fork source), so it is the knob resource.skillCapPerLevel (default 0). Scenarios: two runnable (bank by default, autopilot spends by verb), two pending for want of a trigger.api. Build-specific data: class base and per-point gains come from charstats (classes from tables, tests run Sorceress and Paladin); capability labels from combat profiles and power skillKinds.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 spend stat and skill points by d2d verb behind a capability check, read back
- [x] #2 points sheet API and Points tile; POSTs audited
- [x] #3 per-hero autopilot switch default off, points waiting state
- [x] #4 legality fixes: skill cap check, ladder runeword tags, missing skill kinds, resist order
- [x] #5 Proving Ground scenarios for points
<!-- AC:END -->
