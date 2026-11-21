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

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Capability rows for all 30 Druid skills (V skills.txt requirements, delivery incl. ground, summons, shapeshift, tier with reason, unsupported flagged)
- [ ] #2 druid_fire_guide table from a build file passes buildfile.CheckTable; opt-in by selector, default unchanged (Druid plus three other classes)
- [ ] #3 druid/fire combat profile (data only), unknown profile or class is an error; Druid class index test
- [ ] #4 planner shadow plan for clvl 1 to 30 and agreement fixture at 6/12/18/24/30
- [ ] #5 Proving Ground scenario points.druid-fire-lands-at-30 real and run once on the builder
- [ ] #6 docs/process/druid-support.md with the shared section ground-target delivery for all classes
<!-- AC:END -->
