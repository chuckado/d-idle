---
id: TASK-808
title: >-
  vaultsell: instant vault Sell, the engine's price credited to the stash, Sell
  count honest
status: Done
assignee:
  - credit-runner
created_date: '2026-11-22T16:02:00Z'
updated_date: '2026-11-22T16:51:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 370000
card: "vaultsell"
owns:
  - "cmd/d2bot/vaultkeep.go"
  - "cmd/d2scenario/run.go"
  - "cmd/d2scenario/vaultfacts.go"
  - "docs/knobs.md"
  - "docs/vault.md"
  - "internal/d2dctl/credit.go"
  - "internal/d2dctl/credit_test.go"
  - "internal/d2dctl/paygold.go"
  - "internal/d2dctl/vault_fake.go"
  - "internal/d2dctl/vault_verbs.go"
  - "internal/env/tuning_loot.go"
  - "internal/scenario/state.go"
  - "internal/scenario/vaultassert.go"
  - "internal/scenario/vaultseed_test.go"
  - "internal/vault/store.go"
  - "internal/vaultkeep/keep.go"
  - "internal/vaultkeep/sale.go"
  - "internal/vaultkeep/sale_test.go"
  - "internal/vaultkeep/sell.go"
  - "internal/watch/testdata/vault_logic.mjs"
  - "internal/watch/testdata/vault_sell.mjs"
  - "internal/watch/vaultapi.go"
  - "internal/watch/vaultsale_test.go"
  - "internal/watch/vaultsell_ui_test.go"
  - "internal/watch/web/vault-logic.js"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "tools/provingground/scenarios/vaultsell.json"
exclusive:
  - "d2d-lab-binary"
  - "fork-branch:didle/vaultsell-20261122"
contracts: []
result: "Sell is instant. Count bug: sellPlan skipped queued, pinned, sold and unloaded items without saying why (260 selected printed Sell (11)); the button now reads Sell n of N with the reasons on screen (node test, two fixtures, fails on main). Fork verbs debug creditgold (new, atomic, stash limit 2,500,000, idempotent, --vault-pay) and item price (read-only engine sell price) on didle/vaultsell-20261122 tip 63b1cecd, image d2dbot:devp20 (2cff4a367ccd) loaded not switched, builder tests green except itemdecode (also red on devp19). Bot: Keeper.SellNow credits the engine price and removes the records, falls back to the old queue. Scenario vault.sell-instant GREEN on a lab slot (Paladin 11, Sorceress 3; CHEATED setup, UNMEASURED run, mechanics only). Page not seen in a browser (fit pages not run)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Sell count honest: button states n of N with reasons (node test, two fixtures)
- [x] #2 fork creditgold and item price verbs, tests green on the builder
- [x] #3 Sell removes records and credits the engine price at once (Go tests, Paladin and Sorceress)
- [x] #4 image d2dbot:devp20 loaded, sandbox untouched
- [x] #5 scenario vault.sell-instant green on a lab slot
<!-- AC:END -->
