---
id: TASK-797
title: 'gridwhole: probe, is debug grid whole at arrival (maze levels included)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T02:18:00Z'
labels:
  - d2d
  - probe
dependencies: []
priority: medium
ordinal: 367000
card: "gridwhole"
owns:
  - "docs/probes/gridwhole-2026-11-22.md"
  - "tools/probes/gridwhole.py"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: review docs/design/fable-review-2026-11-22/map-knowledge.md section 5 card 1. Done when the probe doc has the verdict.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Probe script tools/probes/gridwhole.py runs on the lab d2d
- [ ] #2 Grid at arrival compared with the grid after every room is up and with the engine's collision query, maze levels included
- [ ] #3 docs/probes/gridwhole-2026-11-22.md has the per-level table and settles the mouthgoal_test comment
<!-- AC:END -->
