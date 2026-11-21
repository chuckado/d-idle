---
id: TASK-706
title: 'barb1: Barbarian Frenzy (capability rows, build, combat profile, planner shadow, Proving Ground scenario)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T01:01:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 339000
card: "barb1"
owns:
  - "data/capabilities/capabilities.json"
  - "internal/buildfile/frenzy_test.go"
  - "internal/buildfile/tablecheck_test.go"
  - "internal/buildfile/testdata/barbarian-frenzy-guide.json"
  - "internal/plandata/build-barbarian_frenzy_guide.json"
  - "docs/plan/data/build-barbarian_frenzy_guide.json"
  - "internal/leveler/frenzy_test.go"
  - "internal/combat/classdata/profiles.json"
  - "internal/combat/profileid_test.go"
  - "internal/agency/frenzy_test.go"
  - "internal/agency/buildrules_test.go"
  - "internal/agency/planner_test.go"
  - "internal/agency/testdata/agreement.json"
  - "tools/provingground/scenarios/points.json"
  - "docs/process/barbarian-support.md"
  - "docs/agency.md"
  - "internal/scenario/scenario.go"
  - "internal/scenario/scenario_test.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
result: Done offline. First Proving Ground run (builder long lane, job 20261120-200655-c7ae, CHEATED level-30 setups, mechanics only) found two things: the scenario harness class index had Barbarian and Amazon swapped (the Barbarian scenario ran an Amazon; fixed in internal/scenario, V by that run and docs/d2d-fidelity.md) and the necro Skeletal Mage bound 4 was too high (got 3; set to 3). The corrected rerun, builder long lane job 20261120-203155-86f6, is GREEN for both: points.barbarian-frenzy-lands-at-30 (Barbarian seed 11: Double Swing 13, Frenzy 4, Axe Mastery 3, 29 points spent, start line combatProfile=barbarian/frenzy levelerBuild=barbarian_frenzy_guide, writes after start 0) and points.necro-summoner-lands-at-30 (Necromancer seed 11: Raise Skeleton 10, Skeleton Mastery 10, Skeletal Mage 3, 30 spent, writes after start 0). Setups CHEATED (level 30), mechanics evidence only. Verified live: the points verb and the selected table; NOT verified: any Frenzy fight, dual wield, war cry upkeep. Capability rows for all 30 Barbarian skills (V skills.txt requirements; 10 supported passives, 10 experimental, 10 unsupported; Iron Skin corrected from buff to passive). Build barbarian_frenzy_guide (the review's B1 Frenzy, not Whirlwind: review 03 says do not start there) experimental, opt-in by selector, combat line frenzy, from a didle.build/1 file, passes CheckTable; profile barbarian/frenzy; planner shadow sane for clvl 1 to 30, plan and profile agree at 6/12/18/24/30. necro-summoner scenario flipped from pending to real. Later card needs: docs/process/barbarian-support.md section 4.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Capability rows for all 30 Barbarian skills (V skills.txt requirements, tier with reason, unsupported flagged)
- [x] #2 barbarian_frenzy_guide table from a build file passes buildfile.CheckTable; opt-in by selector, default unchanged
- [x] #3 barbarian/frenzy combat profile (data only), unknown profile or class is an error
- [x] #4 planner shadow plan for clvl 1 to 30 and agreement fixture at 6/12/18/24/30, Barbarian plus one other class
- [x] #5 Proving Ground scenarios points.necro-summoner-lands-at-30 (real) and points.barbarian-frenzy-lands-at-30 written and run once on the builder
- [x] #6 docs/process/barbarian-support.md lists what a later combat card builds
<!-- AC:END -->
