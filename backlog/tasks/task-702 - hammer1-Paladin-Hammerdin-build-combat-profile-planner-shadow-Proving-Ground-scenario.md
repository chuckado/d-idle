---
id: TASK-702
title: 'hammer1: Paladin Hammerdin (build, combat profile, planner shadow, Proving Ground scenario)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T23:50:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 338000
card: "hammer1"
owns:
  - "internal/buildfile/hammerdin_test.go"
  - "internal/buildfile/testdata/paladin-hammerdin-guide.json"
  - "internal/plandata/build-paladin_hammerdin_guide.json"
  - "docs/plan/data/build-paladin_hammerdin_guide.json"
  - "internal/leveler/hammerdin_test.go"
  - "internal/combat/classdata/profiles.json"
  - "internal/combat/profileid_test.go"
  - "internal/agency/hammerdin_test.go"
  - "internal/agency/planner_test.go"
  - "internal/agency/testdata/agreement.json"
  - "tools/provingground/scenarios/points.json"
  - "docs/process/hammerdin-support.md"
  - "docs/agency.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
result: Done offline; one Proving Ground run GREEN on the builder long lane (job 20261120-191542-06cc, Paladin seed 11, CHEATED level-30 setup, mechanics only, writes after start 0; a first run 20261120-191335-4c75 was RED on my own too-high Concentration bound, corrected to the table's 4). 30 Paladin capability rows audited, none changed (no V evidence of a wrong tier). Build paladin_hammerdin_guide (experimental, opt-in by selector, combat line hammerdin, from a didle.build/1 file, passes CheckTable); profile paladin/hammerdin (start line V live: combatProfile=paladin/hammerdin); planner shadow sane for clvl 1 to 30, plan and profile agree on Blessed Hammer from clvl 18 (3 of 5 checkpoints; Smite is 0 dps in the estimator). Not verified live: hammer aiming, Concentration while casting, mana. Later card needs: docs/process/hammerdin-support.md section 4.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Capability rows for the Paladin audited against evidence; changed only where V says wrong
- [x] #2 paladin_hammerdin_guide table from a build file passes buildfile.CheckTable; opt-in by selector, default unchanged
- [x] #3 hammerdin combat profile (data only), unknown profile or class is an error
- [x] #4 planner shadow plan for clvl 1 to 30 and agreement fixture, Paladin plus one other class
- [x] #5 Proving Ground scenario points.paladin-hammerdin-lands-at-30 written and run once on the builder
- [x] #6 docs/process/hammerdin-support.md lists what a later card builds
<!-- AC:END -->
