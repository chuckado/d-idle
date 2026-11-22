---
id: TASK-793
title: 'vaultdesktop: Vault page scrolls as one document on desktop'
status: In Progress
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
  - "internal/watch/web/vault-logic.js"
  - "tools/uishots/fixtures_vault.go"
  - "docs/vault.md"
  - "internal/watch/vaultdesktop_test.go"
exclusive: []
contracts: []
result: "The Vault page is one scrolling document at every width (no 100dvh frame, no inner grid scroller; the windowed grid follows the page scroll), desktop cells sized for about 9-10 columns, Reveal bar and Marked for sale are one-line folds (state in localStorage; open by default only with an active order), the Missed drops block is gone from the page (user decision 2026-11-21; API counter kept). Verified in headless Chrome at 1366x768, 1920x1080, 1280x720 and 900x700 with tools/desktopscroll/run.sh on the builder (old frame caught: grid area 170-306 px, no page scroll); NOT verified on a real PC. Other frame pages: second merge."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

User report 2026-11-21: Vault page on PC cannot scroll. Done when the page scrolls at every desktop width.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Must-fail desktop scroll test (1366x768, 1920x1080, 1280x720, 900 wide) fails on old CSS and passes after
- [x] #2 Vault page is one scrolling document with collapsible Reveal, Marked and Missed blocks
- [x] #3 Other frame pages reviewed; screenshots in docs/ui/desktop
<!-- AC:END -->
