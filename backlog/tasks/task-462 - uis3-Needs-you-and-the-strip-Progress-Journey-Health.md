---
id: TASK-462
title: 'uis3: Needs you and the strip: Progress, Journey, Health'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T04:08:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 265000
card: "uis3"
owns:
  - "internal/watch/web/panels/needs-you/"
  - "internal/watch/web/panels/progress/"
  - "internal/watch/web/panels/journey-strip/"
  - "internal/watch/web/panels/health/"
  - "internal/watch/testdata/glancepanels/"
  - "tools/uishots/"
  - "tools/uishots.sh"
  - "internal/watch/glancepanels_test.go"
exclusive: []
contracts: []
result: "Built needs-you (rail), progress, journey-strip, health (strip), all reading /api/glance, each in a fixed box; popovers use the top layer (popover attribute), no timers or fetches. DOM goldens (testdata/glancepanels, node) for every state named in the card, button posts checked against post+body, notification switch off by default and pings only for a NEW red row. tools/uishots: glance fixture variants (needs=N, glance=measuring|blocked|unknown|gearbroken|ch9|nojourney), --pics mode, strip fit pages; full tools/uishots.sh --fit on the builder after the last merge: 352 ok, 0 FAIL (every page, all 8 sizes); labrun tower --gate: TOWERGATE PASS. Renders viewed at all 8 sizes: critique 1) at 1280x620 the health words are hidden and labels are 9 px, readable but small; 2) with five needs and a short window the Orders list inside its box shows a cut row (it scrolls, by design); 3) the rail shows dead space below Orders when the map is not wide. Not verified: live viewer, real notifications (Notification permission), the tally and marks in the Progress popover (glance has no field; ask S2 for one), gear chips/resistances/belt in the Health popover (same). Survival margin: words only (readiness word safe/risky/too hard)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 DOM goldens: needs-you empty, one red, red + amber, five rows (overflow line); progress measuring and full; journey-strip for chapter 1, 3, 9 and blocked; health all ok, gear broken, two unknown
- [x] #2 A button posts exactly the post and body the reply names (node test)
- [x] #3 At the 8 fit sizes the strip is one row of three cells with no clipped text; tools/uishots.sh --fit passes with zero, two and five needs rows
- [x] #4 No number without a unit or label; survival margin shows as words only
<!-- AC:END -->
