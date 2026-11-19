---
id: TASK-546
title: 'idwaitfit: instant-actions indicator fits the character screen at 1280x620'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T15:06:00Z'
updated_date: '2026-11-16T15:19:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 288000
card: "idwaitfit"
owns:
  - "internal/watch/web/inventory-actions.js"
  - "internal/watch/web/inventory-actions.css"
exclusive: []
contracts: []
result: "Indicator is one nowrap line with the closed rule as tooltip; fit pages character-idwait and character-iddoing green at all eight sizes (uishots --fit on the worker host). PNG not viewed."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Follow-up of idfast: the indicator line wrapped and overflowed cv-side at 1280x620.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 character-id* fit pages green at all sizes
<!-- AC:END -->
