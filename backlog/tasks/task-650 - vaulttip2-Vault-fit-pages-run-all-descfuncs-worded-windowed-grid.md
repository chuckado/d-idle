---
id: TASK-650
title: 'vaulttip2: Vault fit pages run, all descfuncs worded, windowed grid'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T18:36:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 324000
card: "vaulttip2"
owns:
  - "internal/data/statdesc.go"
  - "internal/data/statdesc_test.go"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault-logic.js"
  - "internal/watch/web/vault.css"
  - "cmd/d2bot/vaulttip.go"
  - "internal/watch/vaultapi.go"
  - "tools/uishots/fixtures_vault.go"
  - "tools/uishots/fit.go"
  - "tools/uishots/fit.js"
  - "docs/vault.md"
  - "internal/watch/vaultui_test.go"
  - "internal/watch/vaulttip_test.go"
  - "internal/watch/testdata/vault_tip.mjs"
  - "internal/watch/testdata/vault_logic.mjs"
  - "tools/uishots/fit_char_test.go"
  - "tools/uishots.sh"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: follow-up to vaulttip: the tooltip fit pages never ran, only descfunc 1-5 and 12 were worded, the grid kept every row in the DOM. Done when: fit pages run and viewed, StatLines covers the common descfuncs, the grid is windowed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fit pages vault-tip* run and were viewed at phone, tablet, laptop
- [ ] #2 StatLines words the common descfuncs, one shared function, tests per descfunc
- [ ] #3 Vault grid is windowed; 2400 items stay light
- [ ] #4 Cards count and Vault tile checked
<!-- AC:END -->
