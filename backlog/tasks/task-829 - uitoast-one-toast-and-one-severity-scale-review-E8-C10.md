---
id: TASK-829
title: 'uitoast: one toast and one severity scale (review E8, C10)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T00:40:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 379000
result: "Verified offline, not live: TestToastDOM, TestSeverityScaleMatchesToastJS, uilint sev-scale rule, Vault node tests on the shared toast, go test ./internal/watch; fit suite on the worker host at 390x844 and 1366x768. toast.js replaces shell, app.js, alerts.js and Vault toasts; times in one table; Needs you rows use data-need. Chips (data-level) and badges (data-tone) are not on the scale yet."
card: "uitoast"
owns:
  - "docs/OWNERS"
  - "docs/plan/STATUS.md"
  - "docs/ui-design-system.md"
  - "docs/ui/uitoast"
  - "internal/watch/http_view.go"
  - "internal/watch/invui.go"
  - "internal/watch/ordersmap_test.go"
  - "internal/watch/pointsui.go"
  - "internal/watch/skillsui.go"
  - "internal/watch/testdata/glancepanels/needs-you_five-rows.txt"
  - "internal/watch/testdata/glancepanels/needs-you_one-red.txt"
  - "internal/watch/testdata/glancepanels/needs-you_red-amber.txt"
  - "internal/watch/testdata/toast_test.cjs"
  - "internal/watch/testdata/vault_clean.mjs"
  - "internal/watch/testdata/vault_ether.mjs"
  - "internal/watch/testdata/vault_identall.mjs"
  - "internal/watch/testdata/vault_reveal.mjs"
  - "internal/watch/testdata/vault_sell.mjs"
  - "internal/watch/testdata/vault_sellall.mjs"
  - "internal/watch/toast_test.go"
  - "internal/watch/toastcontrast_test.go"
  - "internal/watch/vaultclean_test.go"
  - "internal/watch/vaultether_test.go"
  - "internal/watch/vaultidentall_test.go"
  - "internal/watch/vaultreveal_ui_test.go"
  - "internal/watch/vaultsell_ui_test.go"
  - "internal/watch/vaultsellall_ui_test.go"
  - "internal/watch/web/alerts.js"
  - "internal/watch/web/app.css"
  - "internal/watch/web/app.js"
  - "internal/watch/web/components.css"
  - "internal/watch/web/inventory.js"
  - "internal/watch/web/layoutedit.js"
  - "internal/watch/web/panels/needs-you/panel.css"
  - "internal/watch/web/panels/needs-you/panel.js"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/toast.js"
  - "internal/watch/web/vault.css"
  - "internal/watch/web/vault.html"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/watch.html"
  - "tools/uilint/lint.go"
  - "tools/uilint/uilint_test.go"
  - "tools/uishots/fit.js"
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
- [x] #1 web/toast.js is the one toast: aria-live, role=alert for bad, one table of times, stack of four, with a node test
- [x] #2 the four toast implementations (shell.js, app.js, alerts.js, vault.js) call it
- [x] #3 data-sev takes info, good, warn, bad only: uilint rule sev-scale, scale equal to toast.js; fit suite passes at 390x844 and 1366x768
<!-- AC:END -->
