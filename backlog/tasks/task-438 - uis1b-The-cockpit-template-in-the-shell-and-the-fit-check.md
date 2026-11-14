---
id: TASK-438
title: 'uis1b: The cockpit template in the shell, and the fit check'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-14T20:52:00Z'
updated_date: '2026-11-14T20:52:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 259000
card: "uis1b"
owns:
  - "internal/watch/web/shell.js"
  - "internal/watch/web/shell.css"
  - "internal/watch/web/layout.js"
  - "internal/watch/web/layoutedit.js"
  - "internal/watch/web/shell.json"
  - "internal/watch/web/layouts/cockpit.json"
  - "internal/watch/layout.go"
  - "internal/watch/layout_test.go"
  - "internal/watch/panels.go"
  - "internal/watch/http_view.go"
  - "tools/uilint/"
  - "tools/uishots/"
  - "tools/uishots.sh"
  - "internal/watch/testdata/layout_cells.js"
  - "internal/watch/testdata/layout_cells.golden"
  - "internal/watch/testdata/layout_cockpit.js"
  - "internal/watch/testdata/layout_edit.js"
exclusive: []
contracts: []
result: "Built, not verified live (no game, no deploy). Cockpit template cockpit/1 in the shell: layout keys template/slots/sizes (additive, ParseLayout + CheckSlots), pure plan in layout.js (cockpitPlan, cockpitSwap, cockpitSize), slot DOM and geometry in shell.js/shell.css (band 44 px, stage 4:3 of the height left, strip, rail, two rail columns from 1700 px, phone tabs under 700 px, no page scroll, slots without a mounted panel are left out and the neighbours take the room, no old-tile fallback), manifest key slot: true (and shell.json layout.slot for video and hero-meters, whose panel directories are not ours), ?layout=cockpit (web/layouts/cockpit.json with old-style breakpoints; a stored cockpit wins). Editor on a cockpit (layoutedit.js startCockpit): drag a tile's name onto another to swap, a slot-panel list per tile, strip-height and rail-width dividers with limits (keys too), Save/Reset/Done. uilint knows the keys, refuses an unknown slot name or a non-slot panel. tools/uishots --fit (fit.go, fit.js, fixtures.go): loads each page at 1920x1080, 1440x900, 1366x768, 1914x930, 1904x1000, 1536x730, 1280x620 and 390x844 in headless Chromium, fails on page scroll, [data-slot] or [data-slotpanel] scrollHeight above clientHeight (data-scroll excepted) and on dead space (largest uncovered square above 10 px); the overflowing fixture (and the classic page, which scrolls) must FAIL. Verified on the herdr box with headless Chromium (the Tower job queue was wedged, see report): cockpit, demo with 0/2/5 Needs you rows, demo with a 20-row Orders list pass at all 8 sizes, overflow fixtures fail at all 8. Not verified: the Tower run, whole-repo vet and tests, the zone-menu cases (no zone panel exists yet), the real cockpit panels (other slices). Critique of own renders: (1) the picture is exactly 4:3 and the rail takes the rest, no dead space at any size, but with only video and hero-meters mounted the picture floats in a black frame at wide sizes; (2) strip cells and rail tiles align to the same 8 px gap and edges; the fixture rows are plain text, the real panels must add the accent use J asks for; (3) at 1280x620 a Needs you with five rows leaves the zone grid about 22 px per row, the real zone map must keep its own minimum."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the first attempt's free grid cannot fit a screen (content-height rows). The shell gains a second, additive kind of layout: template cockpit/1 with named slots (band, stage, strip, rail, railWide, phoneTabs) that fills the viewport, next to the unchanged free grid; tools/uishots gets the fit check that proves a page fits. Done when: the criteria below are ticked and the result says what was verified.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every existing layout (stored, Classic, Standard, Simple, Stream, Debug, the three old presets) renders the same cells as before (the S1 compat test still passes, plus a new one for template-less files)
- [x] #2 ?layout=cockpit with the panels that exist today (video, hero-meters) fills the viewport and does not scroll at the sizes of the fit rule; empty slots are left out without holes (fit check output in the result)
- [x] #3 Editor on the cockpit: swap two slots, move both dividers, save, reload: the same arrangement (node test of the plan; pictures looked at)
- [x] #4 uilint accepts the new keys and rejects an unknown slot name or a non-slot panel in a slot
- [x] #5 The fit check fails on a deliberately overflowing fixture page and passes on the cockpit
<!-- AC:END -->
