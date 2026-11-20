---
id: TASK-694
title: 'pgformat: Proving Ground format additions (trigger.build, skill_points, log_line, only_class) and the two pending scenarios run'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T18:37:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 336000
card: "pgformat"
owns:
  - "internal/scenario"
  - "cmd/d2scenario"
  - "tools/provingground"
  - "docs/process/proving-ground.md"
exclusive: []
contracts: []
result: "Format additions (all optional, old scenarios unchanged): trigger.build, skill_points, log_line, only_class; must-fail tests with Paladin and Sorceress. agency.plan-matches-profile and points.sorc-fire-lands-at-30 ran GREEN once each on the builder long lane (Sorceress seed 11; CHEATED setups, mechanics only). Not run: the Paladin of the agency scenario."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: two pending scenarios could not run (no per-run build selection, no per-skill or bot-log assertion). Done when: the format has trigger.build, skill_points, log_line, only_class and both scenarios ran.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 format additions with must-fail tests for two classes
- [x] #2 agency.plan-matches-profile and points.sorc-fire-lands-at-30 flipped from pending and run once each on the builder
- [x] #3 docs/process/proving-ground.md format and coverage sections updated
<!-- AC:END -->
