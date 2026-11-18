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
result: "Fit pages vault-tip* run on the worker host (Chrome) and viewed (390x844, 1024x768, 1440x900); StatLines (internal/data) words descfunc 1-10, 12-16, 19-21, 24, 27, 28 and the grouped lines, tests per descfunc; grid windowed (2400-item fixture, fit judge and must-fail pages); Cards count is not on the Vault. NOT verified live; string keys of grouped lines and descfunc 13, 14 are U"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: follow-up to vaulttip: the tooltip fit pages never ran, only descfunc 1-5 and 12 were worded, the grid kept every row in the DOM. Done when: fit pages run and viewed, StatLines covers the common descfuncs, the grid is windowed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fit pages vault-tip* run and were viewed at phone, tablet, laptop
- [x] #2 StatLines words the common descfuncs, one shared function, tests per descfunc
- [x] #3 Vault grid is windowed; 2400 items stay light
- [x] #4 Cards count and Vault tile checked
<!-- AC:END -->
