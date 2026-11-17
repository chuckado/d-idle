---
id: TASK-591
title: 'vaultv7: ETHER meta wallet and vault capacity upgrades'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T05:14:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 303000
card: "vaultv7"
owns:
  - "internal/ether"
  - "data/vault/ether.json"
  - "internal/watch/etherapi.go"
  - "internal/watch/etherapi_test.go"
  - "docs/ether.md"
  - "docs/vault.md"
  - "docs/tally-strip.md"
  - "docs/design/loot-vault-2026-11-16.md"
  - "data/vault/vault.go"
exclusive: []
contracts: []
result: "ETHER wallet built offline (internal/ether, own ether.jsonl ledger, earn table as data, upgrade and refund-last-step, GET /api/ether and POST upgrade/downgrade); tests incl. property and torn tail; not wired into the bot, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: V7 of the Ethereal Vault design. Done when: the criteria below are ticked.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ETHER ledger wallet: replayed balance, idempotent keys, torn-tail repair
- [x] #2 Earn sources as data (data/vault/ether.json), daily-capped kill trickle
- [x] #3 Upgrade, refund-last-step (fits rule), vault.Account and vault.Wallet implemented
- [x] #4 GET /api/ether, POST /api/vault/upgrade and /downgrade
- [x] #5 docs/ether.md, vault.md wallet section, design doc V7 marked done
<!-- AC:END -->
