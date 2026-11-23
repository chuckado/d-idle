---
id: TASK-831
title: 'uireason: reason lines instead of hover text (review E9, C13)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T00:57:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 379000
result: "Verified offline, not live: Vault node tests, TestQuestUIJS and TestStaticDisabledControlsHaveReasonLines run web/why.js (a disabled control has a visible reason line, no title-only reasons); fit on the worker host at 390x844 and 1366x768 (see the merge note). Vault, quest board and strip, inventory bar converted; the inventory bar grows one line on a phone while a reason shows."
card: "uireason"
owns:
  - "docs/ui-design-system.md"
  - "docs/ui/uireason"
  - "internal/watch/invui.go"
  - "internal/watch/reasonlines_test.go"
  - "internal/watch/testdata/questui/quest_ui.mjs"
  - "internal/watch/testdata/vault_clean.mjs"
  - "internal/watch/testdata/vault_ether.mjs"
  - "internal/watch/testdata/vault_identall.mjs"
  - "internal/watch/testdata/vault_reveal.mjs"
  - "internal/watch/testdata/vault_sell.mjs"
  - "internal/watch/testdata/vault_sellall.mjs"
  - "internal/watch/vaultclean_test.go"
  - "internal/watch/vaultether_test.go"
  - "internal/watch/vaultidentall_test.go"
  - "internal/watch/vaultreveal_ui_test.go"
  - "internal/watch/vaultsell_ui_test.go"
  - "internal/watch/vaultsellall_ui_test.go"
  - "internal/watch/web/components.css"
  - "internal/watch/web/inventory-actions.js"
  - "internal/watch/web/inventory-canvas.css"
  - "internal/watch/web/inventory-canvas.js"
  - "internal/watch/web/panels/quest-board/panel.js"
  - "internal/watch/web/panels/quest-strip/panel.js"
  - "internal/watch/web/vault.css"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "internal/watch/web/why.js"
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
- [x] #1 a disabled control has a visible sibling reason line named by aria-describedby; no title-only reasons (Vault, quest board and strip, inventory bar)
- [x] #2 web/why.js audit runs in the Vault and quest node tests and over the static pages (Go test)
- [x] #3 fit pages at 390x844 and 1366x768 on the worker host pass (inventory bar grows by a line on a phone)
<!-- AC:END -->
