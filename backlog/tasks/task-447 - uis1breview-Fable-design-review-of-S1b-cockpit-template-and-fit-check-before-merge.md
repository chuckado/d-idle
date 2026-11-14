---
id: TASK-447
title: 'uis1breview: Fable design review of S1b (cockpit template and fit check) before merge'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-14T21:57:00Z'
labels:
  - watch
  - research
dependencies: []
priority: medium
ordinal: 259000
card: "uis1breview"
owns:
  - "docs/ui/redesign-2026-11-14/reviews/s1b-review.md"
exclusive: []
contracts: []
result: "Review written, not verified live (no game, no deploy). Verdict APPROVE WITH CHANGES for agent/uis1b at d9389fc56: eight required changes (R0 the band clips the open menu so Edit layout and the presets are unreachable on a cockpit page; R1 the fit check passes four crafted failing pages; R2 a must-fail fixture counts any failure, also a measurement error; R3 data-slot collides with the character page; R4 the wide rail differs from the accepted renders and the key freezes at merge; R5 no 92 px strip on short windows; R6 ParseLayout refuses later-version slots and sizes; R7 overlay mounts the cockpit opt-in panels). Verified on the builder: the cells golden rebuilt from the layout.js of origin/main is identical to the branch golden; 96 fit page loads including the probes. Not build-specific."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: briefs/README.md asks for a design review of S1b before it merges (the shell diff, older layouts render the same cells, the slot rules, the fit check really fails on overflow). Done when: the review doc gives the verdict and the exact changes.

Consumer: uis1b (TASK-438)
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/ui/redesign-2026-11-14/reviews/s1b-review.md gives a verdict with exact required changes, nits and the live checks
<!-- AC:END -->
