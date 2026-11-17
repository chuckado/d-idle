---
id: TASK-588
title: 'vaultv4: Ethereal Vault read API and /watch/vault page (V4)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T04:23:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 302000
card: "vaultv4"
owns:
  - "internal/watch/vaultapi.go"
  - "internal/watch/vaultapi_test.go"
  - "internal/watch/vaultui_test.go"
  - "internal/watch/web/vault.html"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault-logic.js"
  - "internal/watch/web/vault.css"
  - "internal/watch/web/vault.json"
  - "internal/watch/testdata/vault"
  - "cmd/d2bot/vaultwire.go"
  - "tools/uishots/fixtures_vault.go"
  - "docs/vault.md"
  - "internal/vault/index.go"
  - "internal/watch/funwire.go"
  - "internal/watch/testdata/vault_logic.mjs"
  - "internal/watch/testdata/logpanel"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/panels/orders/panel.js"
  - "internal/watch/web/panels/log"
  - "internal/env/tuning_loot.go"
  - "docs/knobs.md"
  - "cmd/d2bot/vaultwire_test.go"
  - "cmd/d2bot/ladder.go"
  - "tools/uishots.sh"
  - "tools/uishots/fit.go"
  - "tools/uishots/fit.js"
  - "tools/uishots/fit_char_test.go"
  - "tools/uishots/fixtures.go"
  - "tools/uishots/fixtures_overlay.go"
  - "tools/uishots/serve.go"
  - "docs/design/loot-vault-2026-11-16.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the Ethereal Vault (design docs/design/loot-vault-2026-11-16.md) needs a read side and a page so the user can see, filter, pin and discard what the bot recorded, before the recorder (V3) and the Keep queue (V5) exist. Done when: the API and page work against a 10k-item store offline, the fit pages and must-fail defects pass on the worker host, and the entry points exist. Not live: V3 does not exist yet, so nothing real fills a vault.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 GET /api/vault, GET /api/vault/item/<id>, POST pin/unpin/discard with 10k-item tests
- [ ] #2 /watch/vault page with filters, sort, multi-select, capacity meter, placeholder Upgrade, disabled Keep
- [ ] #3 entry points: menu, Orders More, Drops feed
- [ ] #4 uishots fit pages and must-fail defect pages, pictures viewed
- [ ] #5 docs/vault.md UI section, V4 marked done
<!-- AC:END -->
