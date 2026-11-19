---
id: TASK-662
title: 'etherreveal: Ether replaces the scroll for vault reveals'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-19T14:49:00Z'
updated_date: '2026-11-19T16:41:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 327000
card: "etherreveal"
owns:
  - "internal/vaultkeep"
  - "internal/vault"
  - "internal/bulkid"
  - "internal/watch/vaultapi.go"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "docs/design/bulk-identify-2026-11-18.md"
  - "internal/loot/vault.go"
  - "cmd/d2bot/vaultkeep.go"
  - "internal/env/tuning_identify.go"
  - "docs/knobs.md"
  - "internal/watch/vaultmass_test.go"
exclusive: []
contracts: []
result: "Ether tap replaces the scroll: unlock from held affixes, 1 Ether default, charged once; Cain free; previews for both. Offline only, not verified live; Keep of a revealed item stays unidentified in game (fork has no flag)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: user decision 2026-11-19, Ether paid on top of the tool. Done when: docs/design/bulk-identify-2026-11-18.md section 12 holds.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Ether pays instead of the scroll, once, after the unlock
- [x] #2 Cain free, scroll and Ether previews before the tap
- [x] #3 Unidentified items never expose affixes before payment
<!-- AC:END -->
