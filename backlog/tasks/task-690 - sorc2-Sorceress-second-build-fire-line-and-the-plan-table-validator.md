---
id: TASK-690
title: 'sorc2: Sorceress second build (fire line) and the plan table validator'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-20T18:20:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 336000
card: "sorc2"
owns:
  - "internal/buildfile/tablecheck.go"
  - "internal/buildfile/tablecheck_test.go"
  - "internal/plandata/build-sorceress_fire_normal30.json"
  - "internal/leveler/sorcfire_test.go"
  - "internal/plandata/build-primary-blizzard.json"
  - "internal/plandata/build-fallback-lightning.json"
  - "internal/plandata/build-fallback-lightning-v2.json"
  - "docs/plan/data/build-primary-blizzard.json"
  - "docs/plan/data/build-fallback-lightning.json"
  - "docs/plan/data/build-fallback-lightning-v2.json"
  - "docs/plan/data/build-sorceress_fire_normal30.json"
  - "internal/combat/classdata/profiles.json"
  - "internal/combat/profileid_test.go"
  - "internal/agency/buildrules_test.go"
  - "internal/buildfile/plans_test.go"
  - "tools/provingground/scenarios/points.json"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: Fable roadmap Q8: the Sorceress second build, and the shipped plan tables had illegal rows (primary-blizzard 43 points into Static Field, cap 20; the lightning tables took Lightning Mastery at clvl 26, needs 30, and Chain Lightning at 27 without Lightning).
Done when: buildfile.CheckTable (cap, required level, prerequisite per row) passes on every shipped Sorceress and Paladin table and rejects the three defects (two classes); build sorceress_fire_normal30 (tier experimental, combat line fire) is selectable by name and the default is unchanged.
result: Done offline, not verified live. Chosen build: fire line S1 (Fire Bolt then Fire Ball to 16, Fire Mastery, Teleport 18), guide-track S. Remaining combat support: Fire Ball has no live verdict (experimental), Teleport hops are not planned by the walker, Telekinesis is a prerequisite point only. The Proving Ground scenario points.sorc-fire-lands-at-30 is pending: the format has no build-select trigger and no per-skill assert.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 CheckTable rejects over-cap, early-level and missing-prerequisite rows; shipped Sorceress and Paladin tables pass
- [x] #2 primary-blizzard and fallback-lightning/-v2 fixed as data
- [x] #3 sorceress_fire_normal30 selectable by name, default unchanged
- [x] #4 Proving Ground scenario written (pending: format gap)
<!-- AC:END -->
