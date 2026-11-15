---
id: TASK-485
title: 'uis7: The cockpit becomes the default; polish and accessibility'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T08:45:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 270000
card: "uis7"
owns:
  - "internal/watch/web/layout.json"
  - "internal/watch/web/layouts/"
  - "internal/watch/web/components.css"
  - "internal/watch/web/panels/{statusline,needs-count,needs-you,orders,zone-map,area-map,progress,journey-strip,health,log}/panel.css"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/layout.js"
  - "internal/watch/layout_test.go"
  - "internal/watch/testdata/layout_default.js"
  - "internal/watch/testdata/layout_cells.golden"
  - "tools/uishots/layout-rows.json"
  - "tools/uishots/genlayouts.js"
  - "tools/uishots.sh"
  - "internal/watch/web/shell.css"
  - "internal/watch/web/tokens.css"
  - "internal/watch/glance.go"
  - "internal/watch/glance_questions.go"
  - "internal/watch/glance_questions_test.go"
  - "tools/uishots/fit.js"
  - "internal/watch/web/update.js"
  - "internal/watch/web/panels/needs-you/panel.json"
  - "internal/watch/testdata/glancepanels/"
  - "internal/watch/web/panels/needs-you/panel.js"
  - "tools/uishots/fixtures_update.go"
  - "tools/uishots/fixtures_glance.go"
  - "tools/uishots/serve.go"
exclusive: []
result: "Cockpit is the default (web/layout.json = layouts/cockpit.json, Standard kept in layouts/standard.json). Choice rule: a stored layout or a menu choice is never changed; a Simple pinned by an older shell (no d2layoutpick marker) follows the default; a new hero gets the cockpit, no pin. Cockpit entry in the Layout menu (Classic is the way back); layout note floats in a cockpit; first-run opens the Getting started drawer once; startup self-check falls back to Standard when a slot panel is missing; OBS strip keeps Standard. Open hero questions are now Needs you rows (the old Thoughts card is hidden in a cockpit). Old message bar floats. Contrast: text-3 text became text-2, light ok/warn chips 4.58:1. Fit check on the default page passes at the 8 sizes plus 1024x768, 900x700, 1279x720 (see commit log for the last run). Renders looked at: 1440x900, 1920x1080, 1366x768, 1280x620, 390x844, dark and light, empty and full. Fable pre-flip review (agent/uis7review, APPROVE WITH CHANGES): the approval is conditional on R1 to R6, all applied here: update banner reserves its height and has Later and a readable button; layout note uses the new key d2layoutnote2; hero questions show every answer, two lines and the fade line; a panel that fails to load offers Use Classic; Needs you buttons and the Orders verb row are 44 px on the phone. RELEASE LINE: the default watch layout changed to the cockpit; a saved layout or a menu choice is kept; the way back is Menu (...) then Layout: Classic, or the address /watch?layout=classic. Not verified live (fixture viewer only): live stream tile, update banner, unlock reveal sheet (fixed overlays, not rendered), screen reader, real phone, Electron window."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fit check passes for the default page (no layout param) at the eight sizes, dark and light, with zero, two and five needs rows and with one and eight orders
- [x] #2 A stored layout is untouched (test); an auto-pinned Simple follows the default; the Layout menu has Cockpit and returns to every older layout
- [x] #3 Renders at 1440x900, 1920x1080, 1366x768, 1280x620 and the phone looked at, dark and light, empty and full data
- [x] #4 Tab order follows reading order; every control reachable by keyboard; contrast 4.5:1; reduced motion
<!-- AC:END -->
