---
id: TASK-918
title: 'toolstrade41: kit mana minimum vs belt plan, stand click retryable'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T04:36:00Z'
updated_date: '2026-11-24T05:10:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 422000
card: "toolstrade41"
owns:
  - "cmd/d2bot/objects.go"
  - "cmd/d2bot/objects_reach_test.go"
  - "internal/resource/beltspec.go"
  - "internal/bot/kitgate.go"
  - "internal/bot/kitmana_test.go"
  - "internal/bot/fakes_test.go"
  - "tools/provingground/scenarios/stepgrounds.json"
exclusive: []
contracts: []
result: "Cause of room_stand RED: not the plan but the kit gate: kitMinMP 3 mana potions in the belt against a plan that holds 1 (Necromancer clvl 31+) or 2, and the scenario belt had none, so the hero went to town at once; kit minimum now capped by ManaCapacity. Stand click failures are KindNotFound (retried 3 times) instead of danger or timeout (a timeout skipped the optional step on the first try). Verified in the lab (devp26-fixes, job 20261123-233159-b34b): room_stand.necro, stand_click_afar.necro, stand_click_pickup.necro GREEN (V)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Kit mana minimum capped by belt plan (Necromancer clvl 40, Sorceress): must-fail test
- [x] #2 Stand click failures retryable (not-found); necro room_stand, stand_click_afar and stand_click_pickup GREEN in the lab
<!-- AC:END -->
