---
id: TASK-589
title: 'vaultv3: Ethereal Vault V3, rules action vault, Vault preset, looter vault path'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-17T04:29:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 302000
card: "vaultv3"
owns:
  - "internal/loot/vault.go"
  - "internal/loot/vault_test.go"
  - "internal/lootrules/vaultpreset.go"
  - "internal/lootrules/vaultpreset_test.go"
  - "cmd/d2bot/vaultwire.go"
  - "docs/vault.md"
  - "docs/lootrules.md"
  - "docs/design/loot-vault-2026-11-16.md"
  - "docs/knobs.md"
  - "internal/loot/pickup.go"
  - "internal/loot/loot.go"
  - "internal/loot/lootshadow.go"
  - "internal/lootrules/types.go"
  - "internal/lootrules/parse.go"
  - "internal/lootrules/eval.go"
  - "internal/lootrules/plain.go"
  - "internal/lootrules/schema.go"
  - "internal/env/tuning_loot.go"
  - "cmd/d2bot/d2dwire.go"
exclusive: []
contracts: []
result: "Built and tested offline only (go test of loot, lootrules, vault, d2dctl, d2bot); NOT verified live: the peek verb was never run from a bot loop. loot.vault 0 by default, Starter and golden tests unchanged."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

V3 of docs/design/loot-vault-2026-11-16.md. Done when the offline tests and the gate pass.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 rules action vault and Vault preset (default off, Starter unchanged)
- [x] #2 looter walks, peeks, records, sends no pickup click; fallback on peek failure; inert without verbs
- [x] #3 knob loot.vault forced off in measured runs; shadow expects vault only with the knob on
- [x] #4 tests on two classes with VaultFake; docs/lootrules.md Vault action, docs/vault.md, design doc V3 marked
<!-- AC:END -->
