---
id: TASK-793
title: 'vaultdesktop: Vault page scrolls as one document on desktop'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T00:46:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 365000
card: "vaultdesktop"
owns:
  - "internal/watch/web/vault.css"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "tools/desktopscroll"
  - "docs/ui/desktop"
  - "docs/vault.md"
  - "internal/watch/vaultdesktop_test.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

User report 2026-11-21: Vault page on PC cannot scroll. Done when the page scrolls at every desktop width.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Must-fail desktop scroll test (1366x768, 1920x1080, 1280x720, 900 wide) fails on old CSS and passes after
- [ ] #2 Vault page is one scrolling document with collapsible Reveal, Marked and Missed blocks
- [ ] #3 Other frame pages reviewed; screenshots in docs/ui/desktop
<!-- AC:END -->
