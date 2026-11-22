---
id: TASK-822
title: 'buildcatalog: build catalogue data, loader and contract test'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T20:33:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 375000
card: "buildcatalog"
owns:
  - "internal/plandata/build-catalog.json"
  - "internal/plandata/catalog.go"
  - "internal/plandata/catalog_test.go"
  - "internal/plandata/plandata_test.go"
  - "internal/buildfile/plans.go"
  - "docs/builds-catalog.md"
exclusive: []
contracts: []
result: "build-catalog.json: 15 builds (30 plan tables, all covered, uncovered list empty), seven estimate ratings each from the SME doc (Summoner Defense N4/H3, Blizzard Idle 3), SME ballpark thresholds as data; catalog.go loader validates and builds the small card (Defense, two highest, lowest, fixed tie order) and accessible star text; TestBuildCatalog plus card tests on Paladin, Sorceress, Necromancer. Build-specific: all text and stars are per build data. Not verified live; all stars are unverified estimates. build-catalog.json is skipped by buildfile.PlanNames (shared build- prefix)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 build-catalog.json covers every plan table; loader validates
- [x] #2 TestBuildCatalog and small-card tests pass (Paladin, Sorceress, Necromancer)
- [x] #3 docs/builds-catalog.md written
<!-- AC:END -->
