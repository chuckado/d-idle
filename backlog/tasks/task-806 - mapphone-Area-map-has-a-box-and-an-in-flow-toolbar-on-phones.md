---
id: TASK-806
title: 'mapphone: Area map has a box and an in-flow toolbar on phones'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T15:39:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 369000
card: "mapphone"
owns:
  - "internal/watch/web/panels/area-map"
  - "internal/watch/web/utilities.css"
  - "tools/uishots/mapphone.go"
  - "tools/uishots/mapphone.js"
  - "internal/watch/testdata/areamap"
  - "docs/ui/map-phone"
  - "tools/uishots.sh"
  - "tools/uishots/fit.go"
  - "tools/uishots/fixtures.go"
  - "tools/uishots/serve.go"
  - "docs/mapfog.md"
exclusive: []
contracts: []
result: "Root cause: the phone page became a scrolling page (mobilequests) and panels take their content height, but the Area map was an absolute canvas in a 100 percent box, so it had height 0 and its controls floated over the tab row. Fix: the toolbar and Layers list sit in the flow above a square picture box (min 260 px) and the Act|Area switch is the first row of the tile. Verified in headless Chrome on the worker host at 390x844 and 1366x768 in four senses states (the check failed before, passes after); not verified on the real viewer."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

User bug 2026-11-22 (phone, Map tab empty, controls over the Quests and Progress tabs). Why: the box collapsed to 0 px on the scrolling phone page. Done when: tools/uishots.sh --map passes (docs/mapfog.md, section "Map on a phone").
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Browser check tools/uishots.sh --map fails before the fix and passes after at 390x844 and 1366x768 (legacy, fresh, walked, waiting)
- [x] #2 Area map canvas > 200 px on a phone, toolbar and Act|Area switch do not intersect the tab row, pixels drawn when the hero is known
- [x] #3 Waiting state says why
- [x] #4 docs/mapfog.md and screenshots in docs/ui/map-phone
<!-- AC:END -->
