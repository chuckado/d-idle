---
id: TASK-718
title: 'dlv-data: one capability table, generated skills.json/skillKinds and tiers'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:41:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 341000
card: "dlv-data"
owns:
  - "data/capabilities/"
  - "internal/delivery/"
  - "tools/capgen/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Schema didle.cap/2 with the delivery object (appendix B); capabilities.json is the ONLY hand-edited skill table (Q1); skills.json and skillKinds generated then removed in dlv-cleanup; Amazon/Assassin rows generated experimental; tier generator per Q2 (unsupported = no valid delivery or no driver; experimental = valid + driver; supported = plus a green Proving Ground row on the backend in use). Gate: golden parity over the existing simulator tests and scenario corpus (Q11), byte-identical decisions. Any hand-typed supported row that drops to experimental is listed in the result and the engine decision stays unchanged unless the row is unused.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every row loads; generated views equal the checked-in files except the listed corrections
- [ ] #2 CI check: skills.json and skillKinds equal the generated views
- [ ] #3 Tier generator output listed; each supported->experimental drop named in result
- [ ] #4 Golden parity test green (agency agreement fixture unchanged)
- [ ] #5 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #6 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #7 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
