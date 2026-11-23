---
id: TASK-884
title: 'uia11y: accessibility and text-scale pass (contrast, names, landmarks, reduced motion, 200% text)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T10:24:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 397000
card: "uia11y"
owns:
  - "internal/watch/a11ycontrast_test.go"
  - "internal/watch/a11ypages_test.go"
  - "internal/watch/web/tokens.css"
  - "internal/watch/web/components.css"
  - "internal/watch/web/shell.css"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/panels/hero-status/panel.css"
  - "internal/watch/returnscreen.go"
  - "internal/watch/policysheet.go"
  - "tools/uishots/"
  - "tools/uishots.sh"
  - "docs/ui-design-system.md"
exclusive: []
contracts: []
result: "Tokens: control border token (3:1) and light good/warn darkened to pass AA; Return and Policy got an h1; the inventory pulse honours reduced motion; ts200 fit pages (320x640 at 200% text) pass for Vault, Quests, Build, Settings after fixes (statusline word wraps, tab bar wraps, slot chip row marked). Tests: TestTokenContrastAA, TestPagesLandmarksAndNames, TestMotionHonoursReducedMotion, TestFitTextScale. Not covered: vertical clipping at 200%, a keyboard walk, screen readers. Full fit not re-run."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 token contrast test (AA) in both themes; tokens fixed
- [x] #2 landmark and name audit of every page; one h1 on Return and Policy
- [x] #3 reduced-motion rule and test; ts200 fit pages pass at 320x640
<!-- AC:END -->
