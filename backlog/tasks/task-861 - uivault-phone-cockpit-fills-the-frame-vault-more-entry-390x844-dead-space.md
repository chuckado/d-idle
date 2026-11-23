---
id: TASK-861
title: 'uivault: phone cockpit fills the frame (vault-more-entry-390x844 dead space)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T05:54:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 389000
card: "uivault"
owns:
  - "internal/watch/web/utilities.css"
exclusive: []
contracts: []
result: "Phone cockpit: the shown phoneTabs panel and the grid rows now fill the frame above the bar. fit vault-more-entry-390x844 passes; full fit at 390x844 on the worker host: 222 red before, 143 after, none newly red (79 fixed). The 143 left are pre-existing, not touched. Not looked at on a real phone."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fit page vault-more-entry passes at 390x844 and the other fit pages at 390x844 stay green
<!-- AC:END -->
