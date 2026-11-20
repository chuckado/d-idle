---
id: TASK-695
title: >-
  pgfix4: Proving Ground chicken.recovery, loot ring, cancelled town-portal
  retry, cain.chain widen
status: Done
assignee:
  - credit-runner
created_date: '2026-11-20T18:45:00Z'
updated_date: '2026-11-20T21:48:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 336000
card: "pgfix4"
owns:
  - "docs/probes/provingground-pgfix4-2026-11-20.md"
  - "internal/bot/portalretry.go"
  - "internal/bot/portalretry_test.go"
  - "internal/bot/loop.go"
  - "internal/bot/runner.go"
  - "internal/town/trip.go"
  - "internal/town/town.go"
  - "cmd/d2bot/d2dwire.go"
  - "internal/env/tuning_bot.go"
  - "internal/loot/"
  - "docs/knobs.md"
  - "cmd/d2drill/portal_copy.go"
  - "internal/scenario/"
  - "tools/provingground/scenarios/behaviour-scenarios.json"
exclusive: []
contracts: []
result: "chicken.recovery GREEN Paladin 11 and Sorceress 3 (scenario fix); loot.rare-ring-classified GREEN both (setup order, wealth assertion, pickit rule for unknown-quality jewelry); cancelled portal cast retried once (unit-tested, no scenario possible); cain.chain GREEN 4 more runs. Builder only, CHEATED setups."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: pgrun2/cainfix3/cainchain left chicken.recovery and loot.rare-ring-classified red and the cancelled portal cast unretried. Done when: scenarios green with evidence in the probe doc.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 chicken.recovery green Paladin 11 and Sorceress 3 on the builder
- [x] #2 loot.rare-ring-classified green Paladin 11 and Sorceress 3
- [x] #3 cancelled town portal cast retried once (must-fail test, two classes)
- [x] #4 cain.chain two more runs each Paladin 11 and Sorceress 3, verdicts and digests recorded
<!-- AC:END -->
