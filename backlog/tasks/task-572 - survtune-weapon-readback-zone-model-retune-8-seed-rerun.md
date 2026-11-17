---
id: TASK-572
title: 'survtune: weapon readback, zone-model retune, 8-seed rerun'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T23:37:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: high
ordinal: 298000
card: "survtune"
owns:
  - "internal/survival"
  - "tools/ab/prep"
  - "tools/survival"
  - "docs/survival/ab-harness.md"
  - "docs/probes/survival-ab-8seed-rerun-2026-11-16.md"
  - "docs/probes/test-characters.md"
exclusive: []
contracts: []
result: "branch 1 (code): prep reads the SETUP weapon back from dump stats (hax 3-6, bot reads dmg 3-6, zone shadow src=skill dps 19.9, verified on the lab, one 30 minute smoke), d2d must come from fork branch didle/rebase-20261115 (the Nov 6 lab build has no weapons key); pay-for-risk route in the dial (params.json typicalEnter/badFloor/riskValueMul, default zoneModel stays 0, tests on two archetypes); reruns and recommendation come on the docs branch"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the 8-seed Paladin A/B (docs/probes/survival-ab-8seed-2026-11-16.md) failed the XP part (ratio 0.823, interval wholly below zero) while winning safety, and the harness Paladin read src=default. Done when: the weapon is read back, the controller trades pay for risk within the dial, and the 8-seed rerun with a plain recommendation is in docs/probes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 harness Paladin weapon read back (dps and src not default), smoke proven on the lab
- [x] #2 controller retune data-driven, tests on two classes, safety kept
- [ ] #3 8-seed Paladin rerun, abreport verdict recorded, writes after start 0
- [ ] #4 rerun doc with a plain recommendation
<!-- AC:END -->
