---
id: TASK-479
title: >-
  uic4: Apply the C2C3 design review to the character screen (R1 to R14) and the
  fit fixtures
status: Done
assignee:
  - credit-runner
created_date: '2026-11-15T06:38:00Z'
updated_date: '2026-11-15T11:15:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: high
ordinal: 268000
card: "uic4"
owns:
  - "internal/watch/web/inventory-canvas.js"
  - "internal/watch/web/inventory-canvas.css"
  - "internal/watch/web/inventory-tip.css"
  - "internal/watch/web/inventory-tip.js"
  - "internal/watch/web/inventory-actions.js"
  - "internal/watch/web/inventory-actions.css"
  - "internal/watch/web/inventory-drag.js"
  - "internal/watch/web/inventory-merc.js"
  - "internal/watch/web/inventory-merc.css"
  - "internal/watch/web/inventory.js"
  - "internal/watch/web/tokens.css"
  - "internal/watch/invui.go"
  - "internal/watch/tipcontrast_test.go"
  - "internal/watch/uic4_test.go"
  - "internal/watch/testdata/uic4/"
  - "internal/watch/testdata/charscreen/"
  - "tools/uishots/"
  - "tools/uishots.sh"
exclusive: []
contracts: []
result: "Review c2c3 R1 to R14 applied on the character screen; not verified live (no game, no browser of the user, no touch device). Phone storage tab: the side column no longer shrinks to 0 px (belt, selected box, errands are back); phone hero sheet four across; merc beside his doll. No mercenary (none, absent block, unreadable): the set bar and the hero values move into the freed column (inventory-merc.js solo, 3 or 4 across), the hire hint is one line. Merc tiles keep the hero tiles height; the room holds his worn items by name and the actWhy line; dead header is chip plus primary Revive with price. Selected box is built from the menu model (d2InvActions.entries/runEntry/cancelRow) and the tooltip rows (d2InvTip.rich): same labels, Give to mercenary, queued rows with Cancel, wrapping verdict lines, comparison. Pending outline and TOWN badge are redrawn after inventory.js rebuilds the cells; the grids repaint on d2:inventory-fresh; first request waits for window load (a real race: an early view left tooltip, menu, merc and canvas undrawn; this was the flaky menu-stash 1366x768 and half-drawn pages). Band one line (meters hidden, nowrap). Light tooltip uses tip tokens (--d-tip-good/warn), TestTooltipContrastBothThemes 4.5:1 in both themes. Long slot names cut against their own box. Touch: no contextmenu menu under a drag or from a touch pointer, All actions button in the box on a coarse pointer. Delete then Enter lands on Keep it. Fit fixtures (tools/uishots fxviews.go, fx= variants nomerc, absentmerc, deadmerc, longnames, pending, plus phone storage tab and light tooltip pages) and new checks (content with no width, cut values tiles, stretched tiles, dead space inside panels, expected selectors, deep band) with Go tests that fail when each defect returns. V: tools/uishots.sh --fit full run on the builder, 456 ok at the eight sizes incl. 160 character pages; node and Go tests (TestUIC4PendingMenuTouch, TestTooltipContrastBothThemes, fit_char_test.go). Looked at renders at 1914x930 (no mercenary), 1536x730, 1280x620 and 390x844 (hero and storage tabs). Not done: grid keyboard navigation, comparison for items no slot accepts (needs item.slots), merc worn items have no tooltip, the phone's search placeholder is cut."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 R1 to R13 of docs/ui/redesign-2026-11-14/reviews/c2c3-review.md applied
- [x] #2 R14 fit fixtures: no mercenary, dead merc, phone storage tab, long names, pending state, light tooltip fail the fit check or tests when the defect returns
- [x] #3 tools/uishots.sh --fit passes at the eight sizes and the renders were looked at
- [x] #4 the whole-repo gate passes on the builder
<!-- AC:END -->
