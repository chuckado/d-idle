# Watch UI overlays: one at a time, one layer scale

Card: overlayfix (bug report 2026-11-15: "drawing over the menu screen").

## The bug
The "..." menu was an absolutely positioned box at `z-index: var(--d-z-sheet)` inside the status band. A zone or More sheet
(`.d-tilesheet`, same z-index) lives inside a panel tile, and every `.d-panel` is a CSS container (`container: panel / inline-size`,
which contains layout and makes a stacking context). Which of the two painted on top was decided by DOM order and those
contexts, so the sheet covered the menu and only its last entry showed. The sheet's reason line was one line cut by `overflow: hidden`.

## Policy
1. Exactly one transient overlay is open at a time: the "..." menu, a tile sheet (zone menu, Orders More), a drawer (Log, Getting
   started), the Reset confirm. Opening one closes the others. `web/overlay.js` does this: every overlay registers its element, and
   when its `hidden` attribute goes false the manager closes the rest through their own close function. Overlays keep their own
   open and close code.
2. Escape closes the one opened last and returns the focus to the element that was focused (or pressed) outside every overlay
   when it opened. An overlay that handles Escape itself (the drawer, the Reset sheet, the menu) calls `preventDefault`, so the
   manager does not close a second one.
3. Layers (tokens.css `--d-z-*`): band 10 < tile sheets and drawers 20..22 < toasts 30 < banners and the Reset confirm 40. The
   "..." menu is a `popover="manual"` element: it sits in the browser's top layer, above all of those, whatever stacking context a
   tile makes. It is placed under its button by `--menu-top` / `--menu-end` (set in `setupMenu`), scrolls inside
   (`max-block-size`) when the window is short, and falls back to the z-index when `showPopover` is missing.
4. Sheet text wraps (`overflow-wrap: anywhere`, `min-inline-size: 0` on flex children); the sheet scrolls inside its tile; the
   charter line with "Pay and go" is sticky at the bottom of the sheet; a sheet never starts above the header.

## Checks (tools/uishots.sh --fit; fit.js `overlayState`, fit.go `judge`)
- every fit page measures `ov`: more than one overlay open, a menu item whose centre is not hit-testable (`elementFromPoint`), text in a
  sheet wider than its box, a sheet above the header or below the window, "Pay and go" outside the sheet.
- pages `overlay-*`: menu over a deep-linked zone sheet, a zone tapped after the menu, light theme, loop, phone sizes.
- must-fail pages `overlaydefect-{under,clip,header,two}` (fixtures_overlay.go) fail with the named verdict or the run fails.
- node: `testdata/overlay_dom.mjs` (every ordered pair of overlay kinds, Escape, focus return), run by `TestOverlayManagerJS`.

Not verified live: a real phone browser (Safari has no focus on button press: the manager also records the pressed button).

## The "..." menu is robust by design (card menufit2, 2026-11-17)
The menu failed the 1280x620 fit three times because every new entry (Pause, Loot rules, Vault, Quests, Support, Stats...) added a row to a flat list. Now (shell.js `groupMenu`): a **primary row** (Pause, Reset) that never scrolls, then ONE scrolling body (`.d-menu-body`, `overflow-y: auto`, inside the popover's `max-block-size`) with collapsible groups **Hero** (Characters, Vault, Quests, Loot rules, Getting started), **Tools** (Edit layout, Help, Theme, Log, Support bundle, Return screen) and **Layout** (the presets, folded by default). A new entry goes into a group; when the groups are taller than the window the body scrolls and a sticky "more" strip says so. Rule for authors: never append to `#menu-more` directly; use a group. `pauseui.js` puts Pause into `.d-menu-primary`.
Checks (fit.js `overlayState`, `menuScroll`): the primary row is on show without scrolling; a body taller than its window must scroll and show the hint; every entry is scrolled to and hit-tested (`elementFromPoint`); the Orders More list must scroll when taller than its sheet. Pages `menu-tall*` (a 700 px group: must pass), `menu-groups*` (Layout group opened), and the must-fail `menudefect-{noscroll,primary,hint}` (fixtures_menu.go) at 1280x620, 1366x768, 900x700, 412x780, 360x640. Pictures: `tools/uishots.sh --fit --out DIR --only 'menu*'`.
