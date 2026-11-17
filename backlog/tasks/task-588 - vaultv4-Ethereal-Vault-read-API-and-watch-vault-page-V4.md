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
- [ ] #1 GET /api/vault, GET /api/vault/item/<id>, POST pin/unpin/discard with 10k-item tests
- [ ] #2 /watch/vault page with filters, sort, multi-select, capacity meter, placeholder Upgrade, disabled Keep
- [ ] #3 entry points: menu, Orders More, Drops feed
- [ ] #4 uishots fit pages and must-fail defect pages, pictures viewed
- [ ] #5 docs/vault.md UI section, V4 marked done
<!-- AC:END -->
