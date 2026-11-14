---
id: TASK-430
title: 'survivaldesign: survival review and unified power-vs-danger design'
status: Done
assignee:
  - user
created_date: '2026-11-14T15:57:00Z'
updated_date: '2026-11-14T17:23:00Z'
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
result: "Review (54 mechanisms inventoried with file:line, thresholds, origin and log counts; nine notions of ready; diagnosis), design (danger profile, power estimate, margin, risk budget, ledger, five archetypes with an override policy, slices S-A to S-F, validation plan, 10 open questions) and README written under docs/survival. Docs only, proposal: nothing implemented, nothing verified live; log counts are V over three recorded logs. Consumer: the S-A to S-E cards (not created; the user decides)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

User request 2026-11-14: review all survival and bot-behaviour changes and design one flexible system that judges character power against zone danger. Docs only, no code. Done when the three docs are merged. Consumer: the slice cards S-A to S-E of docs/survival/design-2026-11-14.md section 10 (created after the user decides).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Review doc docs/survival/review-2026-11-14.md: inventory of every safety and zone-choice mechanism with file:line, thresholds, origin, log counts, conflicts, thrash history, diagnosis
- [x] #2 Design doc docs/survival/design-2026-11-14.md: power estimate, zone danger profile, margin, risk budget, feedback controller, archetype profiles with override policy, rollout slices S-A to S-E, validation plan, open questions
- [x] #3 README docs/survival/README.md: one-page summary (diagnosis, model, slices, open questions)
- [x] #4 Checks pass: backlog, contracts, doclinks, leakcheck
<!-- AC:END -->
