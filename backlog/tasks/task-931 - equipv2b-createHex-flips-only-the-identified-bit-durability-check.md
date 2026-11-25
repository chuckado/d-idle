---
id: TASK-931
title: 'equipv2b: createHex flips only the identified bit (durability check)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-25T14:35:00Z'
labels:
  - watch
  - probe
dependencies: []
priority: medium
ordinal: 433000
card: "equipv2b"
owns:
  - "internal/vaultkeep/equipdirect_test.go"
  - "docs/vault.md"
exclusive: []
contracts: []
result: "createHex is clean: V live devp30 probe, durability and props kept, exactly one bit differs. The 0 percent and 105 percent readings come from elsewhere (U); needs the sandbox unit read."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Size: S. Coordinator suspicion: worn Hand Axe 0 percent, Skull Cap 105 percent. Done when createHex is cleared or fixed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Probe on devp30: durability kept through create and equipdirect
- [x] #2 Unit test: exactly one bit differs
<!-- AC:END -->
