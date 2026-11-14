---
id: TASK-437
title: >-
  uiredesign4: cockpit fixes after the first look (zone menu, minimap, window
  sizes)
status: Done
assignee:
  - credit-runner
created_date: '2026-11-14T20:46:00Z'
updated_date: '2026-11-14T20:54:00Z'
labels:
  - watch
  - research
dependencies: []
priority: high
ordinal: 258000
card: "uiredesign4"
owns:
  - "docs/ui/redesign-2026-11-14/"
exclusive: []
contracts: []
branch: "agent/uidesign4"
result: "docs and mockups only, not verified live (2026-11-14): the zone menu now opens beside the map on a two-column rail and above it otherwise and never covers a zone button; the area map is scaled to fit; short windows get one-line zone rows and a 92 px strip so all ten rows show. Rendered and looked at: watch at 1914x930, 1440x900 and 1280x620 with the menu open, and 1280x620 closed; character at 1914x930, 1536x730 and 1280x620 (fits; the column beside the stash is cramped on short windows). The fit rule in the briefs lists the four browser-window sizes; the three Artifacts were updated in place. Renders made on the head box with a headless Chromium shell; nothing on the worker host"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

The user's first look at the second attempt (2026-11-14, window about 1914x930): the open zone menu hid the bottom zone rows and the area map was cropped. Docs and mockups only. Consumer: the slice cards S1b, S5, S6 whose briefs carry the fix.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The zone menu never covers a zone button; the area map is letterboxed; all zones visible at every supported height
- [x] #2 Browser-window sizes 1914x930, 1904x1000, 1536x730, 1280x620 added to the renders and to the fit rule of the briefs
- [x] #3 Watch and character re-rendered and looked at at the new sizes; the three Artifacts updated in place
<!-- AC:END -->
