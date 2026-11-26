---
id: TASK-984
title: 'agencyp1b: Agency P1b: planner drives combat choice (agency.planner 2) and the h'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:57:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 472000
card: "agencyp1b"
owns:
  - "docs/design/agency-p1b.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Recommendation of the 2026-11-19 design review (06-roadmap R7; ADR 0020 item 2): after the capability table and planner in shadow (agencyp1a, merged), the planner drives (mode 2) and a viability line is shown; guide tracks as build files. No card exists past P1a. Knob agency.planner is default 0 (mode 1 shadow; ledger row says: shadow lines reviewed over 3 days with no wrong pick; owner go for mode 2). Depends on dlv-data (one capability table) for the tier source. Risk: medium. Enable criteria: shadow lines reviewed 3 days, no wrong pick; owner go.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Review of 3 days of planner shadow lines recorded
- [ ] #2 Mode 2 behind its knob with the viability line
<!-- AC:END -->
