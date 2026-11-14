---
id: TASK-454
title: 'uis6: The area map, the drawers, and the phone arrangement'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T22:37:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 260000
card: "uis6"
owns:
  - "internal/watch/web/panels/area-map/"
  - "internal/watch/web/panels/log/"
  - "internal/watch/web/panels/map/"
  - "internal/watch/testdata/areamap/"
  - "internal/watch/testdata/logpanel/"
  - "internal/watch/logtab.go"
  - "internal/watch/logtab_test.go"
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
- [ ] #1 DOM goldens of area-map (with trail, empty, no data) and of the drawer per tab
- [ ] #2 The map panel's goldens are unchanged and Classic at 1440 is pixel-identical (uishots diff)
- [ ] #3 At 1920x1080 the rail shows both maps; at 1440x900 the switch shows one at a time; the fit check passes in both
- [ ] #4 At 390x844 each tab fits the screen under the picture with the fixture data (render looked at)
<!-- AC:END -->
