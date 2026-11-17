---
id: TASK-597
title: 'vaultv7b: wire the ETHER wallet, Vault page upgrade UI, earn call sites'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T06:53:00Z'
labels:
  - watch
  - feature
  - T0
dependencies: []
priority: medium
ordinal: 304000
card: "vaultv7b"
owns:
  - "internal/ether"
  - "internal/watch/etherapi.go"
  - "internal/watch/etherearn.go"
  - "cmd/d2bot/etherwire.go"
  - "docs/ether.md"
  - "data/vault/ether.json"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "internal/watch/web/vault.css"
  - "tools/uishots/fixtures_vault.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Wallet wired in cmd/d2bot behind loot.vault, per account, survives hero switches
- [ ] #2 Vault page Upgrade/Downgrade/chip, fit and pictures
- [ ] #3 Earn call sites idempotent, tests
<!-- AC:END -->
