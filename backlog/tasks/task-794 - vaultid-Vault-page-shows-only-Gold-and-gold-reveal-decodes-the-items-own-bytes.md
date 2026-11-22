---
id: TASK-794
title: 'vaultid: Vault page shows only Gold, and gold reveal decodes the item''s own bytes'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T01:10:00Z'
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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Free, Ether and scroll options hidden on the Vault page behind loot.vaultRevealMethods (default gold only); the Reveal bar shows only Gold
- [ ] #2 Fork verb item decode (read-only) on didle/devp18-20261121 with tests/itemdecode_d2d.py: 0 differences on 4 items, state digest unchanged; regression set run
- [ ] #3 Gold reveal works instantly for a Sealed-less record: decode, pay, unseal in one request; record-time decode; must-fail tests on Paladin and Sorceress
- [ ] #4 devp18 image built and loaded on the sandbox host, not switched; docs/private/devp18.md
- [ ] #5 Scenarios vault.reveal-gold and vault.reveal-gold-short flipped to active and run on a lab slot
<!-- AC:END -->
