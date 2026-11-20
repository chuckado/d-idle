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
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Capability rows for the Paladin audited against evidence; changed only where V says wrong
- [ ] #2 paladin_hammerdin_guide table from a build file passes buildfile.CheckTable; opt-in by selector, default unchanged
- [ ] #3 hammerdin combat profile (data only), unknown profile or class is an error
- [ ] #4 planner shadow plan for clvl 1 to 30 and agreement fixture, Paladin plus one other class
- [ ] #5 Proving Ground scenario points.paladin-hammerdin-lands-at-30 written and run once on the builder
- [ ] #6 docs/process/hammerdin-support.md lists what a later card builds
<!-- AC:END -->
