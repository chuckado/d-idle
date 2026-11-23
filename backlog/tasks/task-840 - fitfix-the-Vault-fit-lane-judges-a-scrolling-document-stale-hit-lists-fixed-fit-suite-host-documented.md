---
id: TASK-840
title: 'fitfix: the Vault fit lane judges a scrolling document, stale hit lists fixed, fit suite host documented'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T02:50:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 381000
card: "fitfix"
owns:
  - "tools/uishots"
  - "tools/uishots.sh"
  - "docs/ui-regression-plan.md"
exclusive: []
contracts: []
result: "Not an environment fault: the Vault became a scrolling document (4e7ab2d2f) and Select hit lists went stale (12e632b72); fit judge and rows updated, vault* fit 70 ok -> 539 ok on the worker host (12 flaky cases remain, see docs/ui-regression-plan.md section 7). Verified on the worker host, not live game."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 First bad commit found by bisect (4e7ab2d2f vaultdesktop for scroll, 12e632b72 vaultclean for hit lists), not an environment fault
- [x] #2 Vault fit rules follow the scrolling document; unit test keeps sideways overflow failing
- [x] #3 Doc says the fit suite runs on the worker host (the worker host) or the builder long lane, never the small head box
<!-- AC:END -->
