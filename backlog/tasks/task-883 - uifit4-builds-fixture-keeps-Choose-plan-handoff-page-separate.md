---
id: TASK-883
title: 'uifit4: builds fixture keeps Choose; plan-handoff page separate'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T09:57:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 396000
card: "uifit4"
owns:
  - "tools/uishots/fixtures_builds.go"
  - "tools/uishots.sh"
  - "docs/ui/fit-triage-2026-11-23.md"
exclusive: []
contracts: []
result: "Full fit on the worker host: 14 red of 2751 (10 are R5 needs-you at 360x640, 4 were builds-choose pages broken by my locked fixture, fixed here and re-measured: all builds-* pass at both sizes)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 builds-choose and builds-choose-adv pass again at 390x844 and 1366x768; builds-plan shows the locked plan
<!-- AC:END -->
