---
id: TASK-789
title: 'pgvault: Proving Ground vault seeding, id templating, worn start, vault_log'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T19:33:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 362000
card: "pgvault"
owns:
  - "tools/provingground/"
  - "internal/scenario/"
  - "cmd/d2scenario/"
  - "docs/process/proving-ground.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: L. Why: four shipped vault features have pending scenarios because the runner cannot seed vault records, template record ids or start with a worn item. Done when those scenarios run.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 pre.vault_records, pre.worn, pre.pack, pre.ether in the runner with must-fail tests (Paladin, Sorceress)
- [ ] #2 {{...}} id templating in trigger.api and api_reply field assertions
- [ ] #3 vault_log assertion kind
- [ ] #4 vaultequip, vaultreveal, manage scenarios active and run (live verdicts in result)
- [ ] #5 docs/process/proving-ground.md format section
<!-- AC:END -->
