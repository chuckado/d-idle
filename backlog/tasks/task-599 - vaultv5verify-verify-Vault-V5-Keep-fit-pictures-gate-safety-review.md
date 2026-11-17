---
id: TASK-599
title: 'vaultv5verify: verify Vault V5 Keep (fit, pictures, gate, safety review)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-17T08:02:00Z'
updated_date: '2026-11-17T08:48:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 305000
card: "vaultv5verify"
owns:
  - "internal/vaultkeep"
  - "tools/uishots/fixtures_vault.go"
exclusive: []
contracts: []
result: "Vault V5 verified offline on the worker host (not live). Fit vault* all ok. Viewed 14 Keep pictures at the five sizes, dark and light. Fixed in the fixtures: the Keep and ether presses answered origin not allowed, the page query rewrite dropped the fixture state (Kept chip showed nothing), the Keep states were off the first screen. Fixed in vaultkeep: the kill switch and safe moment are re-judged before every create. Gate PASS after merging main (an earlier run hit a data race in the quest board desk test, not Vault code). Not a gap: identical items cannot both be in a store."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 vault fit pages and Keep pictures viewed and defects fixed
- [x] #2 gate on vaultkeep, vault, loot, d2dctl, watch, d2bot passes
- [x] #3 vaultkeep safety review done, gaps tested
<!-- AC:END -->
