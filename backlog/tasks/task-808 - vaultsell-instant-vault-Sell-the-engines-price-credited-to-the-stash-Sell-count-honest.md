---
id: TASK-808
title: 'vaultsell: instant vault Sell, the engine''s price credited to the stash, Sell count honest'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T16:02:00Z'
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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Sell count honest: button states n of N with reasons (node test, two fixtures)
- [ ] #2 fork creditgold and item price verbs, tests green on the builder
- [ ] #3 Sell removes records and credits the engine price at once (Go tests, Paladin and Sorceress)
- [ ] #4 image d2dbot:devp20 loaded, sandbox untouched
- [ ] #5 scenario vault.sell-instant green on a lab slot
<!-- AC:END -->
