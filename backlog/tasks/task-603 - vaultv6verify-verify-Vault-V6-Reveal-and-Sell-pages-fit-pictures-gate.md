---
id: TASK-603
title: 'vaultv6verify: verify Vault V6 Reveal and Sell pages (fit, pictures, gate)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T09:28:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 306000
card: "vaultv6verify"
owns:
  - "internal/watch/web/vault.css"
  - "internal/watch/web/vault.json"
  - "tools/uishots/fixtures_vault.go"
exclusive: []
contracts: []
result: "Vault V6 verified offline on the worker host (not live). Fit vault* all ok except the cockpit More menu at 1280x620 (Layout: Debug covered; chars-menu-entry fails the same, not Vault). Pictures viewed for reveal, sell and keep states at the five sizes. Fixed: at 360x640 the selection bar (six rows after Reveal and Sell) scrolled the page, now short labels Keep (n), Reveal (n), Sell (n); the open filter panel plus the Keep row scrolled the page, now capped to the room left. Gate PASS, leakcheck 0 hits."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 vault reveal and sell fit pages pass and pictures viewed
- [x] #2 gate on vault, vaultkeep, town, invdesk, loot, watch passes
<!-- AC:END -->
