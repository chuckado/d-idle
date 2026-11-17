---
id: TASK-593
title: 'vaultv5: Ethereal Vault Keep queue (V5)'
status: In Progress
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
  - "docs/design/loot-vault-2026-11-16.md"
  - "docs/knobs.md"
  - "internal/watch/vaultapi_test.go"
  - "internal/watch/vaultui_test.go"
  - "internal/bot/vaultkeep.go"
  - "internal/bot/contracts.go"
  - "internal/bot/goal.go"
  - "internal/bot/runner.go"
  - "internal/bot/playeraction_test.go"
  - "internal/env/tuning_loot.go"
  - "cmd/d2bot/vaultpage.go"
  - "cmd/d2bot/vaultpage_test.go"
  - "cmd/d2bot/behaviors.go"
  - "cmd/d2bot/d2dwire.go"
  - "cmd/d2bot/live.go"
  - "tools/uishots.sh"
  - "tools/uishots/fixtures_vault.go"
  - "tools/uishots/fit.js"
exclusive: []
result: "Built and tested OFFLINE (VaultFake, Sorceress and Paladin heroes, fake clock); NOT verified live. POST /api/vault/keep queues (idempotent, specific refusals), internal/vaultkeep executes at the next safe moment in town (backup first, read-back, one save, kept logged after it, crash recovery by looking at the hero), knobs loot.vaultKeep and loot.vaultKeepPerVisit, Vault page Keep, states, Kept chip, progress line, fit pages vault-keep*. Live drill is a checklist in docs/vault.md."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Slice V5 of docs/design/loot-vault-2026-11-16.md. Done when the Keep queue works offline against VaultFake; the live drill is a checklist, not run.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Keep queue API and executor, idempotent, crash-safe, tested on two classes
- [x] #2 Vault page Keep button, Keep selected, states, Kept chip, progress line, fit fixtures
- [x] #3 Safety: backup before first create, kill switch, measured runs off, dead hero refused
- [x] #4 docs/vault.md Keep section and live drill checklist; design doc V5 marked done
<!-- AC:END -->
