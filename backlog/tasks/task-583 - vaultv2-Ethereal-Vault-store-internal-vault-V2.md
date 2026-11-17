---
id: TASK-583
title: 'vaultv2: Ethereal Vault store (internal/vault, V2)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-17T03:23:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 301000
card: "vaultv2"
owns:
  - "internal/vault"
  - "data/vault"
  - "docs/vault.md"
  - "docs/design/loot-vault-2026-11-16.md"
exclusive: []
result: "Offline only, not run live: internal/vault store (log, replay, dedupe, state machine, keep crash recovery, capacity/trim/shrink, filters, archive, export/import) with tests and benchmarks; unused until V3."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 internal/vault: JSONL event log, replay index, content-hash dedupe, state machine, keep crash recovery
- [x] #2 capacity tiers and trim/shrink from data/vault, pins, filters, archive-before-delete, export/import
- [x] #3 2-build, crash-between-every-step and random replay tests; benchmarks at 10k and 50k
- [x] #4 docs/vault.md formats; design doc marks V2 done
<!-- AC:END -->
