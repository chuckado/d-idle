---
id: TASK-778
title: 'vaultreveal: pay gold to reveal vault items (the bot buys the scrolls)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T15:01:00Z'
labels:
  - watch
  - feature
  - T2
dependencies: []
priority: medium
ordinal: 359000
card: "vaultreveal"
owns:
  - "internal/vaultkeep"
  - "internal/bulkid"
  - "internal/watch/vaultapi_gold.go"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault-logic.js"
  - "internal/watch/web/vault.css"
  - "cmd/d2bot/vaultkeep.go"
  - "docs/vault.md"
  - "tools/provingground/scenarios/vaultreveal.json"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. Phone viewer: Bulk identify using gold. Done when the Vault page pays a reveal in gold that the bot spends on real Scrolls of Identify.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Mechanism decided and written in docs/vault.md
- [ ] #2 One Reveal bar with a payment selector replaces the notice paragraph
- [ ] #3 vault_reveal gold order: POST, idempotent, one active, cancel, status
- [ ] #4 Bot buys scrolls at the town visit within the purse guards and reveals the same visit
- [ ] #5 knob loot.vaultRevealGold default 0
- [ ] #6 Scenarios vault.reveal-gold and vault.reveal-gold-short (live: unticked if not run)
- [ ] #7 Tests on two classes, race, check.sh, leakcheck
<!-- AC:END -->
