---
id: TASK-563
title: 'survrescreen: Paladin survival A/B rescreen after the clock fix'
status: Done
assignee:
  - live-agent
created_date: '2026-11-16T19:40:00Z'
updated_date: '2026-11-16T21:10:00Z'
labels:
  - bot
  - d2d-measure
dependencies: []
priority: medium
ordinal: 294000
card: "survrescreen"
owns:
  - "docs/probes/survival-ab-rescreen-2026-11-16.md"
  - "docs/probes/test-characters.md"
exclusive: []
contracts: []
result: "Paladin rescreen PASS (deaths 0 vs 0, chickens 1 vs 2 within slack 1, XP ratio 1.18, interval spans zero); ledger now learns chickens (hard 1.5) but k still prints 1.00 and merc deaths and fights are not in the ledger (suspects). Run on lockstep slots, writes after start 0 in 6 of 6; verified live on the lab d2d, not game.exe."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the first screen failed because of the clock bug (fixed in survfix). Done when: the Paladin 3-seed screen is rerun and recorded.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Paladin 3-seed screen rerun on the current main binary, abreport with slack 1
- [x] #2 k behaviour and zone choice differences checked in the margin logs
- [x] #3 record written, verdict and 8-seed cost stated
<!-- AC:END -->
