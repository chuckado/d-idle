---
id: TASK-775
title: 'vaultequip: one-tap Equip from the Ethereal Vault'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T14:18:00Z'
labels:
  - watch
  - feature
  - T0
dependencies: []
priority: medium
ordinal: 359000
card: "vaultequip"
result: "Built and tested OFFLINE (fake hero body over VaultFake, Sorceress and Paladin, race x20); NOT verified live. POST /api/vault/equip (+equipcancel, GET status) queues Keep to the backpack and an order; Keeper.runEquip wears it in the same town visit through Desk.EquipNow (verified, slot pinned), vaults the replaced item first (backpack under keep mark when the vault cannot take it), ring slots left/right/auto. Knob loot.vaultEquip default 0. The four Proving Ground scenarios are declared pending (runner cannot seed a vault record or a worn item yet): live RED/green is a checklist in docs/vault.md."
owns:
  - "internal/vaultkeep/equip.go"
  - "internal/vaultkeep/equip_test.go"
  - "internal/invdesk/equipnow.go"
  - "cmd/d2bot/vaultequip.go"
  - "internal/watch/vaultequip_test.go"
  - "tools/provingground/scenarios/vaultequip.json"
  - "internal/watch/web/vault-logic.js"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "internal/watch/web/vault.css"
  - "internal/watch/vaultapi.go"
  - "internal/vaultkeep/keep.go"
  - "internal/vaultkeep/keep_test.go"
  - "cmd/d2bot/vaultkeep.go"
  - "cmd/d2bot/vaultpage.go"
  - "internal/env/tuning_loot.go"
  - "docs/knobs.md"
  - "docs/vault.md"
  - "internal/watch/testdata/vault_logic.mjs"
  - "tools/provingground/scenarios/behaviours.json"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. One-tap Equip on a Vault item instead of Take out then Equip. Done when the offline tests and gates pass; live proof is a checklist.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 vault_equip order API (equip, equipcancel, GET status) with idempotent semantics, one order per slot, knob loot.vaultEquip default 0
- [x] #2 bot withdraws and wears in one town visit, verified, displaced item goes to the vault, slot pinned, rings left/right/auto
- [x] #3 item sheet with worn-vs-candidate, can-equip reason, Reveal first, Take out
- [x] #4 must-fail unit tests on two classes (swap keeps the old item, vault full keeps it in the backpack, verify, restart)
- [x] #5 Proving Ground scenarios vault.equip-one-tap, -swap, -ring-left, -ring-right declared (live RED/green run NOT done: checklist in docs/vault.md)
<!-- AC:END -->
