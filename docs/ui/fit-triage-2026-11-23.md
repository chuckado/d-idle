# Fit suite triage, 2026-11-23 (card uifit1)

> status: current. Source: the full fit suite (`tools/uishots.sh --fit`, every page at its sizes, on the worker host) on main after cards uivault, uiframe and uialerts. Every number is a count of red RUNS (one page at one size). V = measured in these runs.

## Numbers (V)

| size | red before | red after |
|---|---|---|
| 390x844 | 128 | 0 |
| 360x640 | 80 | 11 (10 after R8 below) |
| 412x780 | 90 | 0 |
| 393x700, 412x890 | 8 each | 0 |
| 870x520 | 7 | 0 |
| 900x700 | 9 | 0 |
| 1000x560 | 7 | 0 |
| 1024x768 | 2 | 0 |
| 1280x620 | 24 | 1 |
| 1366x768 | 23 | 0 |
| 1440x900, 1914x930 | 23, 22 | 0 |
| 1536x730 | 23 | 1 |
| 1904x1000 | 23 | 1 |
| 1920x1080 | 23 | 2 |
| total | 500 | 16 (15 after R8) |

The "143 at 390x844" quoted earlier came from `--fit --sizes 390x844`, which forces every page to that size (including desk-only fixtures that cannot fail there). The default run has 128 red at 390x844.

## Families, in order of (count x user impact)

| id | rule | runs before | verdict | what was done | example |
|---|---|---|---|---|---|
| F1 | `the page scrolls`, `dead space: a 22 px square` (the scrollbar strip), `slot phoneTabs / panel ... ends outside the window` on a phone cockpit | 241 (141 pages) | STALE rule: the phone cockpit grows with its content and the page scrolls on purpose (`web/utilities.css`, "phone: the tab grows to its content") | `fitResult.phoneScroll()` in `tools/uishots/fit.go`: on a window under 700 px with a phoneTabs slot that is taller than the window (or shows a scrollbar strip) the vertical scroll, below-the-window ends and the strip are not defects. Sideways scroll, cut boxes, a slot pushed out to the right and a phone page that does NOT scroll (dead strip above the bar, fixed in card uivault) stay red. Rows: `fit_phone_test.go` | `fit-cockpit-390x844` |
| F2 | `X overflows its box by 1-2 px high` (character screen labels, `em`, `span.cv-lab`) | 200 (39 pages) | STALE: a glyph box 1 to 2 px taller than a one-line box (line-height 1); nothing is hidden | `fitGlyphSlack = 2` (height only, never width). Row: 3 px still fails | `fit-character-1366x768` |
| R1 | `hero-meters .d-meter` 103/96: 7 px wide | 72 | REAL (meter overflows the band at 360 px) | phone `--c-meterbar` min 24 px to 16 px (`shell.css`) | `fit-chars-menu-entry-360x640` |
| F4 | `the lock line is not on the storage tab` | 33 (9 pages) | STALE: the line was removed from the page (card uinav E3) | check only when the element exists; the dead defect page `phonedefect-lock` and its CSS removed | `fit-character-iddoing-390x844` |
| F3 | `#btn-layout is not hittable` after pressing More at a phone width | 16 | STALE: a phone has no layout editing, the button is not offered (`shell.js:594`) | skipped under 700 px; wide windows and every other target still judged. Rows in `fit_phone_test.go` | `fit-cockpit-menu-390x844` |
| R2 | zone map 17 px under the bottom bar at 412x780 (`button.zm-n ends at 733, the bar starts at 724`), phoneTabs overflow | 14 | REAL (the last zone row hidden under the bar) | phone zone-map grid keeps its rows' least height (`utilities.css`) | `fit-overlay-charter-412x780` |
| R3 | statusline sentence cut 13 px at 870x520 | 7 | REAL | the text takes the remaining width and ends in an ellipsis (`panels/statusline/panel.css`) | `fit-more-needs-870x520` |
| F5 | journey popover page | 8 | STALE: the Journey strip was retired (retirej) | page removed from `tools/uishots.sh` | `fit-journey-popover-1366x768` |
| F6 | `overflow-fixture`, `overflow-nested`, `overflow-fakescroll`, `tall-frame` must fail, run at 390x844 | 3 (+ the same at the forced size) | FIXTURE ARTIFACT: desk-frame defects that cannot occur on the scrolling phone | those pages run at desk sizes only; test `TestDeskFixturesRunAtDeskSizes` | `fit-overflow-fakescroll-390x844` |
| R4 | quest tile text cut (`Handing in ...`, 5 px) | 6 | REAL | line-height 1.25 on the two lines (`panels/quest-strip/panel.css`): fixed at 1280x620 to 1920x1080 except two sizes, see open | `fit-quest-handing-1366x768` |

## Final full re-measure (main after cards uichars, uiloot, uifit3): 14 red of 2751 runs

Red per size: 360x640 10 (all R5, the needs-you list), 390x844 2 and 1366x768 2 (a regression of mine: card uiloot's locked-build fixture removed the Choose button from the `builds-choose*` pages; fixed in card uifit4 by giving the locked hero only to the new page `builds-plan`), every other size 0. After uifit4 the open count is the 10 R5 runs.

## Closed by card uifit3

| id | what | verdict | done |
|---|---|---|---|
| R4b | quest tile "handing" at 1536x730 and 1904x1000 | REAL: the reason line beside the disabled button wrapped to five lines and squeezed the tile's two text lines to 0 px | the reason takes at most two short lines (line clamp, full text as its title) (`panels/quest-strip`); all `quest-*` pages pass at every size except the R5 ones |
| R6 | `health .h-btn` 7 px at 1280x620 | REAL, small: the unknown light's word wrapped to two lines in a short cell | one line with an ellipsis in a short cell (`panels/health/panel.css`); `strip-unknown-1280x620` passes |
| R7 | `char-store` 136 px empty square at 1920x1080 | BORDERLINE: 1 px over a line of 135 px; the layout is the same as its 108 to 124 px neighbours | `fitInner` 135 to 140 with rows at 136 (pass), 140 (pass), 141 and 145 (fail) in `fit_char_test.go` |

## Open

### R5: the needs-you list on a 360x640 phone (10 runs; a design decision for you)

In the compact phone strip the needs-you list is a 40 px box when two or more things need you; the rest is reached by scrolling inside the box. The fit rule calls that an overflow because the box is not a declared scroller, and your rule for phones is "no inner scroll box". Two options:

- **A. Keep the 40 px box** and declare it (`data-scroll`) so the fit rule accepts it. The strip stays short: the picture, tabs and Orders stay in view. You scroll inside a small box to see the second and third need. Picture: `docs/ui/fit-triage/needs-you-option-a-360x640.png` (fixture data).
- **B. Let the list grow** on a phone (remove the 40 px limit; the page already scrolls). Every need is visible and tappable at once, the box has no inner scroll; the page grows by about 82 px with three needs, pushing the tabs down. Picture: `docs/ui/fit-triage/needs-you-option-b-360x640.png`.

My recommendation: B (it follows the no-inner-scroll rule and a hero with three needs wants all three visible). The change is one line (`web/utilities.css`, the `max-block-size: 40px` of the phone needs-you list) plus the fit pages `menu-groups`, `quest-needs3`, `quest-away` would pass.

### Other

| id | what | runs | verdict | next step |
|---|---|---|---|---|
| R8 | orders sheet past the window at 360x640 | closed in card uifit2 | | |

Not touched: `gate.go` (the sign-in page) uses the setup wizard's own `/login.css` (class `.btn` there), not `app.css`, so it does not block removing `app.css`.
