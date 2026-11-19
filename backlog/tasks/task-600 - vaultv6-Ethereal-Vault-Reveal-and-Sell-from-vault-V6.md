---
id: TASK-600
title: 'vaultv6: Ethereal Vault Reveal and Sell-from-vault (V6)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-17T08:05:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 305000
card: "vaultv6"
result: "Built and tested OFFLINE (VaultFake, Sorceress and Paladin heroes, fake clock, restart at every call); NOT verified live and the fit pages and pictures were NOT run (no browser on the build hosts). POST /api/vault/{reveal,sell,cancel,sellprompt}, Reveal (create into pack, identify, supersede in place, remove copy; free at Cain; no scroll refused with the buy hint; Keep+Reveal one flow), Sell-from-vault (create, sell chore, read back gold and item, pinned never, per-visit cap, mismatch trips), opt-in prompt loot.vaultSellPrompt, poor-mode pickup loot.vaultPoorPickup, page buttons with the cost line. Live drill is a checklist in docs/vault.md."
owns:
  - "internal/vaultkeep"
  - "internal/vault"
  - "internal/d2dctl/vault_fake.go"
  - "internal/watch/vaultapi.go"
  - "internal/watch/vaultapi_test.go"
  - "internal/watch/vaultui_test.go"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault-logic.js"
  - "internal/watch/web/vault.json"
  - "internal/watch/web/vault.css"
  - "internal/watch/testdata/vault_logic.mjs"
  - "cmd/d2bot/vaultkeep.go"
  - "cmd/d2bot/vaultpage.go"
  - "cmd/d2bot/vaultact_test.go"
  - "internal/loot/vault.go"
  - "internal/loot/pickup.go"
  - "internal/loot/vault_test.go"
  - "internal/env/tuning_loot.go"
  - "docs/knobs.md"
  - "docs/vault.md"
  - "docs/design/loot-vault-2026-11-16.md"
  - "tools/uishots.sh"
  - "tools/uishots/fixtures_vault.go"
  - "tools/uishots/fit.js"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Slice V6 of docs/design/loot-vault-2026-11-16.md. Done when Reveal and Sell-from-vault work offline against VaultFake on two classes; the live drill is a checklist in docs/vault.md, not run.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Reveal: POST /api/vault/reveal, executor with identify through the real path, record superseded in place, restart-safe, no-scroll and free-at-Cain, Keep+Reveal one flow, tested on two classes
- [x] #2 Sell-from-vault: POST /api/vault/sell, create-sell-read-back at the town visit, pinned never sold, per-visit cap, gold mismatch trips, restart-safe, two classes
- [x] #3 Opt-in sell prompt (loot.vaultSellPrompt, default off) as a Needs-you row and the poor-mode pickup rule
- [x] #4 Vault page Reveal and Sell buttons, cost line, states, fit fixtures and picture pages (not run: no browser on the build hosts)
- [x] #5 docs/vault.md sections and live drill checklist; design doc V6 marked done
<!-- AC:END -->
