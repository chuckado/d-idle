---
id: TASK-820
title: 'uinav: page links, stale texts, drop the old cards Inventory (review E1 to E3)'
status: Proposed
result: "Verified offline (Go and node tests, fixture-viewer pictures on the worker host), not live. Links added: Inventory to Points and Vault (band and phone list), Vault to Inventory and Skills, Skills to Points; Vault and Characters Back is /watch?live=1. Removed invPage and ?view=classic; Skills read-only text fixed; Vault name in the menu; duplicate layout-note branch; Inventory loot-rule and stash placeholders and the Respec stub removed. Not done: layout files mockup/live/command (need F2); Loot rules button on Inventory (rewire, follow-up)."
assignee:
  - credit-runner
created_date: '2026-11-22T20:12:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 374000
card: "uinav"
owns:
  - "internal/watch/linkmatrix_test.go"
  - "internal/watch/invui.go"
  - "internal/watch/skillsui.go"
  - "internal/watch/pointsui.go"
  - "internal/watch/invapi.go"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "internal/watch/web/vault.css"
  - "internal/watch/web/characters.js"
  - "internal/watch/testdata/vault_nav.mjs"
  - "docs/ui/nav-links"
  - "docs/ui-design-system.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Size: S. Review E1 to E3 (docs/design/fable-review-2026-11-22, commit fe101397a on agent/fableui). Done when the matrix test, internal/watch and uishots tests are green.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Link matrix test fails on old pages, passes now
- [x] #2 Inventory links Vault and Points; Vault links Inventory and Skills; Skills links Points; Back = /watch?live=1
- [x] #3 E2 stale texts fixed; E3 invPage and ?view=classic removed (layout files left: need F2)
- [x] #4 Before/after pictures at 390x844 and 1366x768 in docs/ui/nav-links
<!-- AC:END -->
