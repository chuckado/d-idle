---
id: TASK-545
title: 'resetfit: header Reset fit fixes'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T14:43:00Z'
updated_date: '2026-11-16T15:53:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 287000
card: "resetfit"
owns:
  - "internal/watch/web/resetui.js"
  - "internal/watch/web/shell.css"
  - "tools/uishots.sh"
  - "internal/watch/web/components.css"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/inventory-actions.js"
  - "internal/watch/web/inventory-actions.css"
exclusive: []
contracts: []
result: "Full uishots --fit green on the worker host. Real causes found by viewing renders: the narrow band (900x700, 360x640) squeezed the sentence and overlapped the links (data-narrow tightens links, drops Return screen); the idwait lane line wrapped to five lines and squeezed the selected card (one truncated line). Toast over the header Reset removed. Not live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 reset* and default* fit pages green
<!-- AC:END -->
