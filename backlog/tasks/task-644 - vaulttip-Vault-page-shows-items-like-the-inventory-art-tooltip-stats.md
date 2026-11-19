---
id: TASK-644
title: 'vaulttip: Vault page shows items like the inventory (art, tooltip stats)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T16:02:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 322000
card: "vaulttip"
owns:
  - "internal/watch/vaultapi.go"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault-logic.js"
  - "internal/watch/web/vault.css"
  - "cmd/d2bot/vaulttip.go"
  - "cmd/d2bot/vaultpage.go"
  - "internal/watch/vaulttip_test.go"
  - "internal/watch/testdata/vault_tip.mjs"
  - "tools/uishots/fixtures_vault.go"
  - "docs/vault.md"
  - "tools/uishots.sh"
  - "internal/vault/types.go"
  - "internal/loot/vault.go"
  - "internal/vaultkeep/reveal.go"
  - "internal/d2dctl/vault_verbs.go"
  - "internal/d2dctl/vault_verbs_test.go"
  - "cmd/d2bot/ladder.go"
  - "cmd/d2bot/d2dbackend.go"
exclusive: []
contracts: []
result: "offline only (tests, no live game): rows carry base facts, art key, vendor price and stat lines (identified only)"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the Vault page showed names only; the user wants it to look like the inventory with the same tooltip. Done when: art cells and the inventory tooltip (stat lines for identified items only) show on the Vault page.

result: offline only, not verified live. Rows carry base facts, art key (picture variant, tint, unique/set row stored with the record), vendor price and stat lines (identified only, descfunc 1-5 and 12 worded). Fit pages and live check not run.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Vault rows carry the inventory tooltip facts (base, art, props); identified only for props
- [x] #2 Unidentified items never expose affix text in the API payload (must-fail test)
- [x] #3 Node tooltip test: weapon and jewel, vault text equals inventory text
<!-- AC:END -->
