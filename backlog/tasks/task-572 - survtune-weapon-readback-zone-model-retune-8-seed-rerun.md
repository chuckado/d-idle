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
result: "weapon read back (hax 3-6, src=skill dps 19.9, V on the lab) and d2d must be fork branch didle/rebase-20261115; Paladin 8 seeds with the shipped route (badFloor 0.4, freshNeedsMerc false): abreport PASS (XP ratio 1.002, deaths 0/0, chickens 4/1, writes after start 0); first route 0.895 INCONCLUSIVE; 4 fresh holdout seeds FAIL on chickens by one (3 vs 1+1, one death, XP 0.967); Sorceress 3 seeds INCONCLUSIVE 0.836; recommendation: do not flip bot.zoneModel yet, retune badFloor 0.5 to 0.65 on new seeds (docs/probes/survival-ab-8seed-rerun-2026-11-16.md); verified live on the lab, not game.exe"
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
- [x] #3 8-seed Paladin rerun, abreport verdict recorded, writes after start 0
- [x] #4 rerun doc with a plain recommendation
<!-- AC:END -->
