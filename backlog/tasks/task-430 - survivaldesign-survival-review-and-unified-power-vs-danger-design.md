---
id: TASK-430
title: 'survivaldesign: survival review and unified power-vs-danger design'
status: In Progress
assignee:
  - user
created_date: '2026-11-14T15:57:00Z'
labels:
  - bot
  - research
dependencies: []
priority: high
ordinal: 256000
card: "survivaldesign"
owns:
  - "docs/survival/**"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

User request 2026-11-14: review all survival and bot-behaviour changes and design one flexible system that judges character power against zone danger. Docs only, no code. Done when the three docs are merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Review doc docs/survival/review-2026-11-14.md: inventory of every safety and zone-choice mechanism with file:line, thresholds, origin, log counts, conflicts, thrash history, diagnosis
- [ ] #2 Design doc docs/survival/design-2026-11-14.md: power estimate, zone danger profile, margin, risk budget, feedback controller, archetype profiles with override policy, rollout slices S-A to S-E, validation plan, open questions
- [ ] #3 README docs/survival/README.md: one-page summary (diagnosis, model, slices, open questions)
- [ ] #4 Checks pass: backlog, contracts, doclinks, leakcheck
<!-- AC:END -->
