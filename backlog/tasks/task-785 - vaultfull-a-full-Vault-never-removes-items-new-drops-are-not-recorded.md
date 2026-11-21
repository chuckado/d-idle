---
id: TASK-785
title: 'vaultfull: a full Vault never removes items, new drops are not recorded'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T19:08:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 361000
card: "vaultfull"
owns:
  - "internal/vault/"
  - "internal/vaultkeep/"
  - "internal/loot/vault.go"
  - "internal/loot/vault_test.go"
  - "internal/loot/pickup.go"
  - "internal/watch/vaultapi.go"
  - "internal/watch/vaultapi_test.go"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "internal/watch/web/vault-logic.js"
  - "cmd/d2bot/vaultkeep.go"
  - "cmd/d2bot/vaultpage.go"
  - "internal/env/tuning_loot.go"
  - "docs/vault.md"
  - "docs/knobs.md"
  - "tools/provingground/scenarios/vaultfull.json"
exclusive: []
contracts: []
result: "Offline, not verified live. Full vault never trims or sells by itself (loot.vaultAutoTrim 0 default; 1 restores the old trim and auto-sell of marks). New drops are counted as missed drops (persisted, best 10, Clear button); sell-queued flags without a player order (the 196 of the sandbox vault) show as Marked for sale with a Sell these N now button and are never sold automatically. Tests on Paladin and Sorceress; scenarios vault.full-no-trim and vault.full-notice pending (runner cannot seed vault records)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Full vault: no trim, no sell, new drops not recorded (knob loot.vaultAutoTrim 0)
- [x] #2 Missed drops counter and list persisted; marked-for-sale list with a confirm button
- [x] #3 loot.vaultAutoTrim 1 reproduces the old trim and auto-sell (two classes)
- [x] #4 docs/vault.md Decision 2026-11-21 and docs/knobs.md
<!-- AC:END -->
