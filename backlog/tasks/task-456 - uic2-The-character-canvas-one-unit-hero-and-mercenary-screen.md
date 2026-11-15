---
id: TASK-456
title: 'uic2: The character canvas (one-unit hero and mercenary screen)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T22:53:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 260000
card: "uic2"
owns:
  - "internal/watch/web/inventory.js"
  - "internal/watch/web/inventory.css"
  - "internal/watch/web/inventory-merc.js"
  - "internal/watch/web/inventory-merc.css"
  - "internal/watch/web/inventory-sets.js"
  - "internal/watch/web/inventory-sets.css"
  - "internal/watch/web/inventory-canvas.js"
  - "internal/watch/web/inventory-canvas.css"
  - "internal/watch/invui.go"
  - "internal/watch/invapi.go"
  - "docs/inventory-ui.md"
  - "internal/watch/charscreen_test.go"
  - "internal/watch/testdata/charscreen/"
  - "tools/uishots.sh"
  - "tools/uishots/pages.go"
  - "tools/uishots/fixtures.go"
exclusive: []
result: "Built on agent/uis1b (S1b is queued as #711, not on main yet): /watch/inventory is the character canvas (charPage in invui.go, inventory-canvas.js and .css new, inventory-merc.js has a canvas mode, inventory.js only fires d2:repair); ?view=classic keeps the cards page byte for byte. V (headless Chromium on the builder, tools/uishots.sh --fit pages character and character-light, fixture viewer): no scroll, no overflow, no dead space at all eight sizes 1440x900, 1920x1080, 1366x768, 1914x930, 1904x1000, 1536x730, 1280x620, 390x844, both themes. V: go test ./internal/watch (node DOM goldens testdata/charscreen: hero only; requirement marks over C1 four views and one cell per reason code; selected-item buttons; filters; merc none, dead, desert, ironwolf, unknown, absent, dead with revive; page ids, routes, no-rule grep, css rules), old inventory, merc, sets, actions, drag and tooltip goldens unchanged and green; TOWERGATE PASS (gofmt, check.sh, short tests, race on internal/watch and internal/invdesk, leakcheck). Looked at renders at 1914x930, 1536x730, 1280x620 (and 1440 light, 390): critique below. NOT verified live: real data in a real browser, Waterfox (round(), container queries, :has()), game art layer on the canvas doll (CSS only), drag over the canvas doll, the sets popover, Revive (no desk binds merc.revive yet), the 21 validity cases are judged server-side by C1 (the page test shows every reason code). Gaps: /api/hero has no Str, Dex, Vit, Energy, defense or damage, so those sheet tiles are not drawn (heroTiles is the one place to add them); the loot rules summary is a placeholder strip; the click still opens the old action popover (C3 replaces it). Build-specific: nothing (class, kind and slots come from the view). Changed beyond owns: tools/uishots/fit.js names the element in an overflow verdict (verdicts unchanged), fixtures and pages for the canvas. Critique: (1) 1914x930 u=63: no dead zone left in the who column, the selected box has a gap (about 80 px) when the item has few properties, merc tiles are large and sparse; (2) 1536x730 u=47: cleanest size, side column and all labels readable; (3) 1280x620 u=37 (the weak point of J): the four-cell column now holds belt, selected item with both verdicts, two buttons and More, the errands and the lock line, nothing overflows; labels in cells are 8.5 px, the ring and amulet slot words are 6.5 px and hard to read."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the user rejected the old Inventory page as a masonry of unrelated cards with dead space. Done when: /watch/inventory is one continuous canvas on one cell unit (hero and mercenary dolls side by side, sets bar, stat sheets, pack over stash with the belt, the selected item and the town errands beside the stash, gold, repair and sell as chips) that passes the fit check at the eight sizes of the brief; the cards page stays at ?view=classic. Brief: docs/ui/redesign-2026-11-14/briefs/C2-character-canvas.md. Tooltip and right-click menu are C3.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fit check passes for the character page at all 8 sizes (1440x900, 1920x1080, 1366x768, 1914x930, 1904x1000, 1536x730, 1280x620, 390x844)
- [x] #2 DOM goldens: hero only; each mercenary state (none, dead, desert, ironwolf, unknown, absent); the requirement marks over C1's four views and one cell per reason code
- [x] #3 No rule in the page (a test greps the scripts for requirement arithmetic and class lists)
- [x] #4 Every action of the old page still works (old inventory, merc, sets and actions tests pass unchanged; the cards page stays at ?view=classic)
- [x] #5 Renders at 1914x930, 1536x730 and 1280x620 looked at, critique written in the result
<!-- AC:END -->
