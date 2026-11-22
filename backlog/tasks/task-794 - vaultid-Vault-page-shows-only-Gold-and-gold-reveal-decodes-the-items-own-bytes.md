---
id: TASK-794
title: >-
  vaultid: Vault page shows only Gold, and gold reveal decodes the item's own
  bytes
status: Done
assignee:
  - credit-runner
created_date: '2026-11-22T01:10:00Z'
updated_date: '2026-11-22T02:56:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 365000
card: "vaultid"
owns:
  - "internal/vaultkeep"
  - "internal/bulkid"
  - "internal/vault"
  - "internal/loot/vault.go"
  - "internal/loot/vault_test.go"
  - "internal/d2dctl/vault_verbs.go"
  - "internal/d2dctl/vault_verbs_test.go"
  - "internal/d2dctl/vault_fake.go"
  - "internal/watch"
  - "cmd/d2bot/vaultkeep.go"
  - "cmd/d2scenario"
  - "internal/scenario"
  - "internal/env/tuning_loot.go"
  - "docs/knobs.md"
  - "docs/vault.md"
  - "tools/provingground/scenarios/vaultreveal.json"
exclusive:
  - "fork-branch:didle/devp18-20261121"
contracts: []
result: 'Vault page: only Gold (loot.vaultRevealMethods, default 4), one Reveal with gold button on the toolbar and sheet, Equip needs no reveal. Fork didle/devp18-20261121 verb item decode: pure decode works (the unidentified bytes carry the affixes), itemdecode_d2d 4 items 0 differences, regression set green (smoke_d2d known fail). Image d2dbot:devp18 60df7519efb9 loaded, not switched. Scenarios vault.reveal-gold(+short) active, run live on a lab slot: the order decodes 6 bare records and pays 480 from the stash (verified live, UNMEASURED run, CHEATED setup, mechanics only); final verdict in the merge report.'
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Free, Ether and scroll options hidden on the Vault page behind loot.vaultRevealMethods (default gold only); the Reveal bar shows only Gold
- [x] #2 Fork verb item decode (read-only) on didle/devp18-20261121 with tests/itemdecode_d2d.py: 0 differences on 4 items, state digest unchanged; regression set run
- [x] #3 Gold reveal works instantly for a Sealed-less record: decode, pay, unseal in one request; record-time decode; must-fail tests on Paladin and Sorceress
- [x] #4 devp18 image built and loaded on the sandbox host, not switched; docs/private/devp18.md
- [x] #5 Scenarios vault.reveal-gold and vault.reveal-gold-short flipped to active and run on a lab slot
<!-- AC:END -->
