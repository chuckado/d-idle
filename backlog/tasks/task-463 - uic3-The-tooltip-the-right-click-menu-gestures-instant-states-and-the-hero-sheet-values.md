---
id: TASK-463
title: 'uic3: The tooltip, the right-click menu, gestures, instant states and the hero sheet values'
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
card: "uic3"
owns:
  - "internal/watch/web/inventory-tip.js"
  - "internal/watch/web/inventory-tip.css"
  - "internal/watch/web/inventory-actions.js"
  - "internal/watch/web/inventory-actions.css"
  - "internal/watch/web/inventory-drag.js"
  - "internal/watch/web/inventory-canvas.js"
  - "internal/watch/web/inventory-canvas.css"
  - "internal/watch/testdata/uic3/"
  - "internal/watch/herosheet_test.go"
  - "internal/watch/hero.go"
  - "internal/bot/heroattrs.go"
  - "internal/bot/heroattrs_test.go"
  - "cmd/d2bot/readywire.go"
  - "tools/uishots/"
  - "tools/uishots.sh"
  - "internal/watch/uic3_test.go"
  - "internal/watch/web/inventory-sets.css"
  - "docs/inventory-ui.md"
  - "docs/ui-design-system.md"
exclusive: []
contracts: []
result: "Built on origin/main after uic2 (the canvas), not verified live. Tooltip (inventory-tip.js, canvas only): 270 px beside the cell, flips right to left, a bottom or top sheet on touch and under 40rem, opaque, never scrolls; blocks: the game's lines, where it is (worn, pack, stash, belt, queued move, TOWN), requirement rows for you and for the mercenary from req and usable (a refusal is never a met row), comparison old to new with signed deltas for every slot it fits (both rings, both hands), hint line. Right-click menu with a key on every entry and greyed entries with the reason (Enter, Ctrl+click, Alt+click, Del with one inline confirmation, M greyed with the verdict and the bot's own reason, R repair); double-click or Enter runs the first moving entry; a click only selects. Feedback from the reply: data-sending (dashed), done (green, fades), refused (red, reason in data-why, line above the stash, the verb's code in the item actions list), queued (view queue, TOWN badge); nothing moves before the view says so. Drag: footprint of the held item over pack and stash. Hero values: /api/hero stats gains attrs, defense, damage, absent when unread (Runner.Sheet, readywire.go); sheet tiles bound, 16 tiles. V (node and Go tests on the worker host: go test ./internal/watch ./internal/bot, testdata/uic3/c3.js: 21 validity cases of I.10 x 2 classes against tooltip rows and menu entries, goldens for tooltip, menu, feedback per action kind, placement of 5000+ cells at the eight sizes never off-screen and never over its cell; sheet.js; TestHeroStatsSheetFields; TestSheetOf). V (headless Chromium on the builder): tools/uishots.sh --fit --only 'character*' passes at all eight sizes (1440x900, 1920x1080, 1366x768, 1914x930, 1904x1000, 1536x730, 1280x620, 390x844) for the page, the light page, five tooltip-open and three menu-open fixtures (hover and right-click in the harness: the popup opens, lies inside the window, the tooltip does not cover its cell). Looked at renders at 1914x930, 1280x620 and 390x844 (tooltip on a pack item, a ring pair, a worn item, a stash item; menu on a pack item and a stash item): critique: (1) the tooltip over the stash and pack covers neighbouring cells by design, fine, but at 1280x620 it also hides most of the selected box beside it; (2) the first hint line wraps on a long refusal, now shortened to refused, see above; (3) on the phone the sheet is clean, key hints are hidden there. NOT verified live: a real browser and real data, Waterfox, touch long-press, the instant lane (invdesk.instant is off by default: every reply is queued), Give to mercenary (the bot cannot dress him yet), Repair from the menu, comparison for an item every slot refuses for the same requirement (the view has no item.slots field), defense is stat 31 (U live), damage is stats 21 and 22 not the weapon base (S). Build-specific: nothing in the page; class and kind come from the view; the merc kind names are a display table."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the character canvas (uic2) still opened the old click popover, had no real tooltip with requirements and comparison, and its hero stat tiles were blank. Done when: a floating tooltip with signed deltas, a right-click menu with keys and gestures, instant-action states and the missing hero values work and pass the fit check at the eight sizes with the tooltip and the menu open. Brief: docs/ui/redesign-2026-11-14/briefs/C3-tooltip-menu-gestures.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 DOM goldens: tooltip (no requirements, each unmet kind, class item, mercenary block, comparison with deltas, both rings), menu per place, each feedback state
- [x] #2 Tooltip never leaves the viewport nor overlaps its cell at the eight sizes (node test of placement)
- [x] #3 A refused requirement or action never renders as done (node test)
- [x] #4 Hero Str Dex Vit Energy defense damage in /api/hero stats, absent when unread, tiles bound
- [x] #5 Fit check passes at all 8 sizes incl. tooltip open and menu open fixtures; renders looked at
<!-- AC:END -->
