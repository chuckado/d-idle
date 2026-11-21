---
id: TASK-707
title: 'druid1: Druid fire (capability rows, build, combat profile, planner shadow, Proving Ground scenario, ground-target note)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T02:43:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 340000
card: "druid1"
owns:
  - "data/capabilities/capabilities.json"
  - "internal/buildfile/druidfire_test.go"
  - "internal/buildfile/tablecheck_test.go"
  - "internal/buildfile/testdata/druid-fire-guide.json"
  - "internal/plandata/build-druid_fire_guide.json"
  - "docs/plan/data/build-druid_fire_guide.json"
  - "internal/leveler/druidfire_test.go"
  - "internal/combat/classdata/profiles.json"
  - "internal/combat/profileid_test.go"
  - "internal/agency/druidfire_test.go"
  - "internal/agency/buildrules_test.go"
  - "internal/agency/planner_test.go"
  - "internal/agency/testdata/agreement.json"
  - "tools/provingground/scenarios/points.json"
  - "docs/process/druid-support.md"
  - "docs/agency.md"
  - "internal/scenario/scenario_test.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: Fable roadmap Q8 class 5: the Druid fire build (review 03 D1), capability rows and the shared ground-target note. Done when: the criteria below hold.
result: Done offline, per-package tests and -race -count=20 green (agency, buildfile, leveler, scenario, plandata, combat), check.sh ok, leakcheck 0 hits. Capability rows for all 30 Druid skills (V skills.txt requirements, delay and mana; 14 experimental, 16 unsupported, none supported because no Druid skill ever ran live; Armageddon, Hurricane and Cyclone Armor are self-cast and Tornado a missile, only Firestorm, Molten Boulder, Eruption and Volcano are ground-aimed). Build druid_fire_guide (review 03 D1) experimental, opt-in by selector, combat line fire, from a didle.build/1 file, passes CheckTable; default Druid and four other classes unchanged (test). Profile druid/fire data only. Planner shadow sane for clvl 1 to 30; plan and profile agree at 12/18/24/30 and disagree at clvl 6 (power model gives Firestorm 0 dps, U). Class index test now covers all seven classes against data/smoke/rules.json (Druid 5). Scenario points.druid-fire-lands-at-30 is real; its one run on the builder long lane is job 20261120-215709-2f47, queued behind other long jobs when this card was written (NOT verified live: result below if it ran). docs/process/druid-support.md section 5 lists what a later card needs for ground-target delivery for all classes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Capability rows for all 30 Druid skills (V skills.txt requirements, delivery incl. ground, summons, shapeshift, tier with reason, unsupported flagged)
- [x] #2 druid_fire_guide table from a build file passes buildfile.CheckTable; opt-in by selector, default unchanged (Druid plus three other classes)
- [x] #3 druid/fire combat profile (data only), unknown profile or class is an error; Druid class index test
- [x] #4 planner shadow plan for clvl 1 to 30 and agreement fixture at 6/12/18/24/30
- [ ] #5 Proving Ground scenario points.druid-fire-lands-at-30 real and run once on the builder
- [x] #6 docs/process/druid-support.md with the shared section ground-target delivery for all classes
<!-- AC:END -->
