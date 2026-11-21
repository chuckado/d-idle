---
id: TASK-697
title: 'necro1: Necromancer summoner (capability rows, build, combat profile, planner shadow)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T18:49:00Z'
labels:
  - bot
  - feature
  - T0
dependencies: []
priority: medium
ordinal: 337000
card: "necro1"
owns:
  - "data/capabilities/capabilities.json"
  - "internal/buildfile/necrosummoner_test.go"
  - "internal/buildfile/testdata/necromancer-summoner-guide.json"
  - "internal/buildfile/tablecheck_test.go"
  - "internal/plandata/build-necromancer_summoner_guide.json"
  - "docs/plan/data/build-necromancer_summoner_guide.json"
  - "internal/leveler/necrosummoner_test.go"
  - "internal/combat/classdata/profiles.json"
  - "internal/combat/profileid_test.go"
  - "internal/agency/necrosummoner_test.go"
  - "internal/agency/planner_test.go"
  - "internal/agency/buildrules_test.go"
  - "internal/agency/testdata/agreement.json"
  - "tools/provingground/scenarios/points.json"
  - "docs/process/necromancer-support.md"
  - "docs/agency.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M

Why: Fable roadmap Q8 (user decision 2026-11-19): class order Sorceress 2nd build, then Necromancer summoner; the table had 14 of 30 Necromancer rows and no requirements.
Done when: see the criteria; no combat code and no live play change.
result: Done offline, not verified live. 30 Necromancer capability rows: 5 supported, 10 experimental, 15 unsupported (the engine already has summon upkeep, Amplify/Weaken/Lower Resist and Corpse Explosion code, so it is further than assumed). Build necromancer_summoner_guide (experimental, opt-in by selector, combat line summoner, expanded from a didle.build/1 file, passes CheckTable). Planner shadow sane for clvl 1 to 30; attack agreement with the profile 0 of 5 by design (pets have no damage in the model), summon slot agrees. Proving Ground scenario points.necro-summoner-lands-at-30 is pending (trigger.build and skill_points are on agent/pgformat). Later card needs: docs/process/necromancer-support.md section 4.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Capability rows for all 30 Necromancer skills with tier reasons, requirements (V skills.txt) and synergies
- [x] #2 necromancer_summoner_guide table from a build file passes buildfile.CheckTable; opt-in by selector, default unchanged
- [x] #3 summoner combat profile (data only), unknown profile or class is an error
- [x] #4 planner shadow plan for clvl 1 to 30 and agreement fixture, two classes in the must-fail tests
- [x] #5 Proving Ground scenario written (pending: pgformat not merged)
- [x] #6 docs/process/necromancer-support.md lists what a later card builds
<!-- AC:END -->
