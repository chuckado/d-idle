---
id: TASK-593
title: 'vaultv5: Ethereal Vault Keep queue (V5)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T06:19:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 304000
card: "vaultv5"
owns:
  - "internal/vaultkeep"
  - "internal/vault"
  - "internal/watch/vaultapi.go"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "internal/watch/web/vault.css"
  - "cmd/d2bot/vaultkeep.go"
  - "docs/vault.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Slice V5 of docs/design/loot-vault-2026-11-16.md. Done when the Keep queue works offline against VaultFake; the live drill is a checklist, not run.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Keep queue API and executor, idempotent, crash-safe, tested on two classes
- [ ] #2 Vault page Keep button, Keep selected, states, Kept chip, progress line, fit fixtures
- [ ] #3 Safety: backup before first create, kill switch, measured runs off, dead hero refused
- [ ] #4 docs/vault.md Keep section and live drill checklist; design doc V5 marked done
<!-- AC:END -->
