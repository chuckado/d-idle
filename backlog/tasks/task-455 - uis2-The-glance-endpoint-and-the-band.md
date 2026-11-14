---
id: TASK-455
title: 'uis2: The glance endpoint and the band'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-14T22:39:00Z'
updated_date: '2026-11-14T22:39:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 260000
card: "uis2"
owns:
  - "internal/watch/glance.go"
  - "internal/watch/glance_test.go"
  - "internal/watch/testdata/glance/"
  - "internal/watch/web/panels/statusline/"
  - "internal/watch/web/panels/needs-count/"
  - "internal/watch/web/panels/hero-meters/"
  - "internal/watch/testdata/glance_dom.js"
  - "internal/watch/testdata/glance_dom_*.txt"
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
- [ ] #1 Handler goldens testdata/glance/view-{daily,needs,blocked,paused,dead,unknown}-{sorceress,paladin}.json
- [ ] #2 For every code in whychange.json the sentence has a non-empty reason (table test)
- [ ] #3 DOM goldens of statusline (ok, warn, danger, info, loading, error; long reason cut) and needs-count (0, 1, 3)
- [ ] #4 The band is 44 px at the three desktop sizes and two lines on the phone; fit check passes
- [ ] #5 Classic and the other free-grid layouts are unchanged
<!-- AC:END -->
