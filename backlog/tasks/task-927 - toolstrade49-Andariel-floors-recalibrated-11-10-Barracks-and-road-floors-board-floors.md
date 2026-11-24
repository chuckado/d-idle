---
id: TASK-927
title: >-
  toolstrade49: Andariel floors recalibrated (11 / 10, Barracks and road floors,
  board floors)
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T18:36:00Z'
updated_date: '2026-11-24T19:06:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 429000
card: "toolstrade49"
owns:
  - "internal/env/tuning_boss.go"
  - "internal/env/tuning_route.go"
  - "internal/plandata/boss_levels.go"
  - "internal/plandata/boss_levels_test.go"
  - "internal/plandata/runs.json"
  - "docs/plan/data/runs.json"
  - "data/questboard/act1.json"
  - "internal/task/campaign.go"
  - "internal/task/campsim_floors_test.go"
  - "internal/task/campsim_test.go"
  - "internal/task/floors_test.go"
  - "internal/task/task_test.go"
  - "internal/task/andariel_test.go"
  - "internal/bot/beltfill_test.go"
  - "internal/bot/routeguard_test.go"
  - "internal/campaign/campaign_test.go"
  - "internal/levelplan/levelplan_test.go"
  - "docs/knobs.md"
  - "docs/andarielgate.md"
  - "docs/levelplan3.md"
  - "docs/probes/andariel-gate-recalibration-2026-11-24.md"
exclusive: []
contracts: []
result: "Applied with the user go (2026-11-24): env.DefaultAndarielMinLevel 11 and ...Merc 10, class rows floor 11 / 10 and recommended 14 / 13, runs.json 11, route.monasteryMinLevel/Merc 10, route.barracksMinLevel/Caster 11, route.barracksMinLife/Caster 0, board legacyFloor 11 for Andariel and Tools of the Trade; boss.andarielHold stays 1 in code. The optional Forgotten Tower is left behind once the hero is in Act 2 (campaign.go). Tests updated (floor table rows at clvl 8, 9, 10, 11 for no merc, dead and living merc). Offline only, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Andariel floors 11 / 10, recommended 14 / 13, Barracks floors 11 with life floors 0, road floor 10, board floors 11, all as defaults with tests on two classes
- [x] #2 Sandbox tuning block for the same numbers in the report
<!-- AC:END -->
