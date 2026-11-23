---
id: TASK-834
title: 'skelspeed: skeleton minion speed and pathing on the lab d2d'
status: Done
assignee:
  - live-agent
created_date: '2026-11-23T01:45:00Z'
labels:
  - d2d
  - probe
dependencies: []
priority: medium
ordinal: 246000
card: "skelspeed"
owns:
  - "docs/probes/skeleton-speed-2026-11-23.md"
  - "docs/probes/skeleton-speed-trace-*.csv"
  - "docs/probes/skeleton-speed-foes-*.csv"
  - "docs/probes/skeleton_speed_*.py"
exclusive: []
contracts: []
result: "Measured on the lab d2d (devp20 build, lockstep, V): skeleton speed equals the table (Velocity 13: 0.8125 subtile/tick observed 0.807 to 0.812; Run 15 pace 0.931) and the follow rule's catch-up band (peak 23.3, none above); hero walk 6 and run 9 also exact. Pathing: 0 stuck runs, snap exactly at the 10-cell teleport, lag up to 6.9 cells in a corridor. Upstream #675 is in devp20; no upstream commit since e80f59042 changes pet speed or path. No deviation found, no PR draft, no rebase. Not verified against game.exe on screen (U)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: user report 2026-11-22, skeletons too fast and odd pathing. Done when: speed and pathing measured on the lab d2d with traces and a decision. Consumer: d2d fidelity review (docs/research/README, none needed).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Skeleton speed measured on the lab d2d against the table and the follow rule
- [x] #2 Pathing checked on traces (stuck, snaps, lag)
- [x] #3 Upstream commits since e80f59042 listed; decision recorded
<!-- AC:END -->
