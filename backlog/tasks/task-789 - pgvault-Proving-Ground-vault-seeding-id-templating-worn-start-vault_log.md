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
result: "Runner: pre.pack, pre.worn, pre.vault_records (real items taken into the vault store), pre.ether, {{vault.N.id}} and {{pack.item[name=..].id}} placeholders in trigger.api, api_reply field assertions, vault_log, vault_count, vault_revealed, vault_records_present, vault_missed, vault_api, a d2d capability check for needs. Verified live on the builder (devp16, Paladin 11 and Sorceress 3): seeding and templating work; vault.full-no-trim and vault.full-notice GREEN on Sorceress 3 (full vault trims nothing, API shows full, autoTrim false), manage.equip-immediate GREEN on Paladin 11; the four vault.equip scenarios, manage.recall-now and manage.identify-immediate RED with findings F1 (the d2d backend never gives the runner its VaultKeep, so Keep, Equip and Reveal orders never run), F2 (equip after Keep does not find the created item) and F3 (identify verb ok but no change); vault.reveal-gold and -short stay pending (no debug paygold in devp16)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: L. Why: four shipped vault features have pending scenarios because the runner cannot seed vault records, template record ids or start with a worn item. Done when those scenarios run.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 pre.vault_records, pre.worn, pre.pack, pre.ether in the runner with must-fail tests (Paladin, Sorceress)
- [x] #2 {{...}} id templating in trigger.api and api_reply field assertions
- [x] #3 vault_log assertion kind
- [x] #4 vaultequip, vaultreveal, manage, vaultfull scenarios active and run (reveal stay pending: no debug paygold in devp16); live verdicts in docs/process/proving-ground.md section 10
- [x] #5 docs/process/proving-ground.md format section
<!-- AC:END -->
