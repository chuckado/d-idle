---
id: TASK-900
title: 'toolstrade28: readiness gate for the Barracks (class level and life floors)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T21:03:00Z'
updated_date: '2026-11-23T21:29:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 408000
card: "toolstrade28"
owns:
  - "internal/env/tuning_route.go"
  - "internal/task/gate.go"
  - "internal/task/world.go"
  - "internal/task/route_test.go"
  - "internal/bot/routeguard.go"
  - "docs/knobs.md"
  - "tools/questloop.sh"
  - "tools/questloop_test.sh"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Readiness floors for the Barracks as knobs (melee 24 and 400 life, caster 22 and 300 life; EXP numbers from the d2-sme), refusal too_weak with level N needed, questloop waits on it; unit and shell tests; not run live; strikes already reset on a level-up (expireStrikes)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. Live and lab: a clvl 20 hero cannot clear the Barracks arrival. d2-sme ballpark: Paladin clvl 24-26 and 400 life, Necromancer clvl 22-24 and 300 life. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Tools of the Trade gate needs the class level and max life (route.barracks* knobs) and says 'level N needed'; the hero levels meanwhile; must-fail test, melee and caster pairs
- [x] #2 questloop treats a strength hold as waiting, not a try or a skip; shell test
<!-- AC:END -->
