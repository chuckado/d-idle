# Review of slice S1b: the cockpit template in the shell, and the fit check

> status: review of branch `agent/uis1b` at `d9389fc56` (card TASK-438) against `origin/main` at `f050a6fab`, 2026-11-14. Reviewer: the design reviewer named in [../briefs/README.md](../briefs/README.md) ("What Fable reviews"). Nothing here was checked on a live game or a deployed viewer. Tags: V = verified (how is said), S = read in code or a doc only, U = unverified.

## Verdict: APPROVE WITH CHANGES

The compatibility promise holds for every older layout (section 1), the shell change is small and well separated, and the pure plan is tested. Eight things must change before the merge: R0 because the "..." menu cannot be used on a cockpit page, R1 and R2 because the fit check passes pages it must fail, R3 to R5 because the template's keys and attribute names freeze under the compatibility promise the day this merges and three of them are wrong or collide, R6 because a later-version layout file is refused instead of ignored, R7 because it changes the OBS strip the day the default flips. None needs a redesign; all are inside the card's owned paths.

## Required changes (before merge)

| # | Where (branch head) | What is wrong | Change |
|---|---|---|---|
| R0 | `internal/watch/web/shell.css:122` | The band clips (`overflow: hidden`) and the "..." menu's sheet is positioned inside it (`.d-menu-pop`, `shell.css:182`), so on a cockpit page the open menu is invisible and cannot be clicked. V on the builder at all 8 sizes: the sheet's box is 224 x 328 px under the band, the element at its centre is a tile of the page, the Edit layout button is not hittable. Help, Edit layout, Theme and the layout presets (the way back to Classic) are unreachable except by `?edit=1` or by editing the address. | The band must not clip what opens from it: clip `.d-band-main` only (it already does, `:123`), or give `.d-menu-pop` `position: fixed` under a cockpit. Add a fit page with the menu open (see section 5, the click step) that asserts the sheet is hittable. |
| R1 | `tools/uishots/fit.js:7-13`, `:18-24` | Four crafted pages pass at all 8 sizes that must fail (V on the builder, section 3): content cut one level below the panel root; `data-scroll` put on an element that does not scroll; a slot whose only panel is hidden (a blank rail); a frame taller than the viewport that the shell clips. | (a) Test every descendant of a slot panel for `scrollHeight - clientHeight > 1`, not only the root; except elements whose computed `overflow-y` is `auto` or `scroll` and that carry `data-scroll`, and single-line text with `text-overflow: ellipsis` (width only). (b) `data-scroll` counts only on an element that really scrolls (`overflow-y: auto|scroll`), never on the panel root. (c) Dead space is measured from the visible `[data-slotpanel]` boxes plus the band, not from the slot containers; the stage's own letterbox is the one named exception. (d) Fail when a slot or panel box ends outside the viewport (`right > vw + 1`, `bottom > vh + 1`). Keep the four pages as must-fail fixtures. |
| R2 | `tools/uishots/fit.go:105`, `:171-177` | A dump named `fit-overflow-*` counts as "failed as it must" for ANY reason, also `{"error": ...}` and a viewport mismatch. V: a dump holding `{"error":"boom"}` under that name printed "ok ... failed as it must (the measurement failed: boom)" and the command exited 0. So a harness that breaks reports the self-test as green. | Each must-fail page names the reason it must fail with (`overflow-fixture`: "overflows its box"; `overflow-classic`: "the page scrolls"); an error, "no measurement", "no [data-slot]" or a viewport mismatch is a FAIL of the run for every page. Add the cases to `TestFitJudgeAndDir`. |
| R3 | `internal/watch/web/shell.js:481`, `:539`; `shell.css:137-148`; `tools/uishots/fit.js:9`, `:18` | `data-slot` is already the equipment-slot attribute of the character page (V: `web/inventory.js:68`, `inventory-merc.js:43`, `components.css:162-193`, values `head`, `belt`, ...). The fit check is meant to run on the character page next (brief, Do 6): there `[data-slot]` matches ten gear buttons, which then count as layout regions that "cover" the frame. After the merge the attribute cannot be renamed (compat promise: no data attribute renamed). | Rename the layout attribute now, while nothing reads it: `data-cslot` (and keep `data-slotpanel`). Update shell.css, fit.js, layoutedit.js `slotEl`, the tests, and the wording of the fit rule in the briefs (report line for the coordinator; the brief is not this card's file). |
| R4 | `internal/watch/web/layouts/cockpit.json:25-34`, `layout.go:76-79` (`LayoutRailWide`), `layout.js:319`, `shell.css:147` | The wide rail is two equal columns, left = Needs you, Orders, area map; right = act map. The accepted renders (V: `mockups/v2/watch.html:181`, `research/iterations/final/src/watch.html:178`) have Needs you and Orders across the whole rail and under them the act map (1.7fr) beside the area map (1fr). J section 2's sentence and the JSON of J.3 describe the earlier arrangement; the renders are what the user accepted, and they apply at 1920x1080, 1914x930 and 1904x1000 (three of the eight sizes, one of them the user's own window). `railWide {left, right}` cannot express it, and the key freezes at merge. | `railWide: { "top": [...], "left": [...], "right": [...] }` with `top` optional and spanning both columns, columns `1.7fr 1fr`; `cockpit.json`: `top: [needs-you, orders]`, `left: [zone-map]`, `right: [area-map]`. `CheckSlots`, `pools`, `cockpitSwap`, `cockpitPlan` and the node test learn `top`. If the coordinator prefers J's sentence over the renders, that is a user decision to record; do not merge with the question open. |
| R5 | `internal/watch/web/shell.css:116`, `:131`; `layout.js:300-307` (`cockpitSizes`) | The strip is 112 px at every height. J section 2 ("On short windows the strip is 92 px") and the accepted mockup (V: `mockups/v2/watch.html:209`, `@media (max-height:760px) and (min-width:701px){:root{--strip:92px}`) make it 92 px at 1536x730 and 1280x620. At 1280x620 this is 20 px of picture height. The strip cells are also equal thirds where the renders have `1fr 1.25fr 1.1fr` (`mockups/v2/watch.html:68`). | Under 760 px of height (and not phone) the strip is `min(sizes.strip, 92)`; in CSS (`min()` in the media query) or in `cockpitSizes(layout, height)`; add the two short sizes to the node test. Cell weights: nit N3 unless cheap. |
| R6 | `internal/watch/layout.go:42-44`, `:255-259` | `ParseLayout` type-decodes `slots` and `sizes` whatever the template is, so a file of a later version is refused, not ignored (V: a throwaway test on the branch: `template: "cockpit/2"` with `"stage": {"id": ...}` gives "want a panel id or a list of ids"; `slots` as a list and `sizes.strip: "12%"` fail too). `postLayout` then answers 400 and `layoutStore.load` drops a stored file to the default without a word. J.3 promises the opposite ("a shell that does not know `template` ignores it and uses `breakpoints`"), and layout.js does it right (`isCockpit` is false, the grid is planned). | Keep `Slots` and `Sizes` as `json.RawMessage` in `LayoutManifest`; decode them only when `Template == "cockpit/1"`, and report a decode error from `CheckSlots` (the strict check uilint runs), never from `ParseLayout`. Test: the three files above parse, store and are served back byte for byte. |
| R7 | `internal/watch/web/layout.js:153`; `testdata/layout_cockpit.js:83` | `named()` now starts from `cockpitIds(layout)`. A cockpit on the page returns early in `selectPanels`, and a file without a template has no cockpit ids, so the only effect of this line is in overlay mode under a cockpit layout: every opt-in panel a slot names is mounted in the OBS strip (the node test asserts `orders` is). The day S7 makes the cockpit the default, the overlay of every streamer gains `statusline`, `needs-count` and the rest. | `named()` goes back to `new Set()`; the overlay under a cockpit mounts what it mounts under Classic. Replace the assertion at `layout_cockpit.js:83` with: `selectPanels(list, cockpit, 'overlay')` equals `selectPanels(list, classic, 'overlay')`. |

## 1. The compatibility promise

| Claim | Result |
|---|---|
| The cells golden was taken with the layout.js of before the change | V: `layout_cells.js` of the branch run against `web/layout.js` of `origin/main` (`f050a6fab`, sha256 `113b0868...`) with `--write` gives a file identical to the branch's `testdata/layout_cells.golden` (12 lines, all equal). Done twice: on the builder (node v20.19.2, as the brief asked) and on the head box; both identical. |
| The golden can fail | V: one mutated condition in `named()` (layout.js of main) changed the hashes. |
| What the golden covers | `plan()` at 390, 768, 1280, 1440, 1920 in the page and in the editor, the overlay plan at 1920, and `selectPanels` in both modes, for `layout.json`, the five presets, the three editor presets and the three v1 compat fixtures. It does not cover the DOM the shell builds; the existing `Compat` tests do (V: `go test -p 1 -run 'Compat|Layout|Cockpit|Templateless|Manifest' ./internal/watch/` passes on the branch). |
| Stored layouts keep loading | S: `layoutStore.load` and `postLayout` are unchanged; a file without the new keys parses as before (V: `TestTemplatelessLayoutsRenderTheSameCells` also parses each file and finds no template). The exception is R6 (files with later-version keys). |
| Schema additive | S: three new optional keys (`template`, `slots`, `sizes`), one new manifest key (`slot`), one new `shell.json` list (`layout.slot`). `watch.layout/1` unchanged. |
| `?layout=cockpit` is not the default | V: `web/layout.json` is untouched in the diff and `TestShippedCockpitLayout` asserts it has no template; `PRESETS` is unchanged, so the menu does not offer it. |
| `X-Layout-Source` | S: unchanged in `layout.go`; the new `?layout=cockpit` branch of `loadLayout` (`shell.js:370-375`) reads it to let a stored cockpit win over the shipped file. |
| Layouts without a template untouched in the shell | S: every new branch in `shell.js` is behind `isCockpit(sel.layout) && mode !== 'overlay'`; the new CSS is scoped to `[data-template="cockpit"]` or to new class names. One side effect on other layouts: none found. |
| Overlay | Unchanged for older layouts (golden). Changed for cockpit layouts: R7. |

One behaviour to know, not a defect: Save in the cockpit editor makes the cockpit this profile's stored layout, so plain `/watch` then opens the cockpit (today: the picture and the meters only). The way back is the "..." menu (Classic) or Reset. Today that menu does not work on a cockpit page (R0), so a user who presses Save there is stuck on it.

## 2. The slot rules

Against J section 2 (numbers computed from `shell.css:116-135`, S; the same numbers as J's table):

| Size | Picture | Rail | J |
|---|---|---|---|
| 1366x768 | 784 x 588 | 558 | same |
| 1440x900 | 960 x 720 | 456 | same |
| 1920x1080 | 1200 x 900 | 696 | same size; arrangement differs (R4) |
| 1536x730, 1280x620 | strip 112 | | J: strip 92 (R5) |
| 390x844 | 374 x 280 inside an 8 px frame | tabs | J: 390 x 292 edge to edge (nit N4) |

- Band 44 px, phone 84 px (two lines): as designed. The band clips (`overflow: hidden`, `shell.css:122`), which is right for the sentence and wrong for anything that opens from the band (section 5).
- "A slot with no mounted panel is left out, no old-tile fallback": right, and it is what J.3 decided. Two holes in it for a partly built cockpit and for the flip in S7:
  1. A panel that is mounted and hides itself still holds its slot. V: a rail whose one panel calls `ctx.hide(true)` leaves a blank column of the full rail width at every size, and the fit check passes it. The plan counts `ck.roots.has(id)` (`shell.js:486`), not what is shown. Between S3 and S4 the rail of `?layout=cockpit` is exactly this: Needs you alone, hidden when there is nothing to say. Fix with R1 (the check) and in `applyCockpit`: a slot whose panels are all hidden is not built (re-plan on `ctx.hide`).
  2. A rail column has no floor for its last panel: every panel before it may take 50 % (`shell.css:152`), so Needs you plus a long Orders list leave the act map what is left. The card's own critique saw 22 px zone rows at 1280x620; J requires all ten rows and none under 20 px (the map needs about 380 px). The template needs a minimum for the growing panel (a manifest key such as `slotMin`, or the cap of the others computed from it). Needed by S4/S5, not by this merge; write it into those briefs.
- `shell.json` `layout.slot` listing `video` and `hero-meters`: clean enough for one slice, and it is linted (`layout-slot`: the id must be a panel). It is a second source of truth for one fact, so it must end: S2 (which owns the band panels) moves `hero-meters` to `"slot": true` in its manifest, S6 or S7 does `video`, and S7 deletes the list and the `LayoutConf.Slot` field's use. Add a uilint finding when an id is both in the list and has `slot: true`.
- `slot: true` is one flag for three different boxes (a 44 px band, a 4:3 stage, a tile). The editor's list offers every slot panel for every place, so the meters can be put on the stage and the picture in a 112 px strip cell. It is reversible, so not a blocker; before S7 the editor should offer only panels that fit the kind of place (an additive manifest key, for example `slotKinds: ["band"]`, default tile).
- uilint: accepts the keys, refuses an unknown slot name, a compat card or alias in a slot, a non-slot panel, a panel twice in one arrangement, two stage panels, sizes out of range, an unknown template (V: the test cases in `tools/uilint/uilint_test.go` read; not rerun here beyond the builder run of the card). Missing: `slot: true` together with `fit: true` (a contradiction) is not a finding.

### The editor

| Question | Answer |
|---|---|
| Can a panel be lost with no way back? | No (S). A swap is a permutation in every arrangement; a replaced panel stays in the list of every tile (`slotPanels` is every slot manifest), Done drops unsaved changes, Reset drops the stored copy. One trap: swapping with a panel that stands only in another arrangement (Orders with the area map at 1440) removes Orders from the page arrangement; it comes back only through a tile's list. Say so in the bar's hint, or refuse a swap that removes a panel from the arrangement on screen. |
| Number of panels per slot | Fixed: nothing adds a fourth rail tile or removes one. Fine for S1b; S7 should decide. |
| Keyboard | Swap: the list in each tag (a real `select`, labelled) is the keyboard path; the drag is pointer only. Dividers: focusable, arrows move 8 px, limits enforced. |
| Accessibility gaps | (a) `role="separator"` on a focusable element needs `aria-valuenow`, `aria-valuemin`, `aria-valuemax` and `aria-orientation` (`layoutedit.js:347`). (b) The tags are appended to band panels too, but only tiles in `.d-main` get `position: relative` (`shell.css:168`), so a band panel's tag is positioned against the band and clipped by it. (c) Focus is not moved to the editor bar on entry and Escape does not leave. (d) Editing is refused under 700 px with a toast: fine. |
| Reset | `postJSON {reset: true}`, then the stored copy is gone and the page reloads after 700 ms; with `?layout=cockpit` in the address the shipped cockpit loads again, without it the default. Correct. |
| Phone tabs | `role="tablist"` and `role="tab"` without a `tabpanel`, no arrow keys, and each click rebuilds the bar (`shell.js:520`, `:498`), so the focused tab button is destroyed and keyboard focus falls to the body. Must be fixed by S6 (which owns the phone arrangement); cheap to fix here: keep the bar, swap only the body. |

## 3. The fit check

What it measures (S, `fit.js`): page scroll (`scrollingElement.scrollHeight/Width` against the iframe's inner size), `scrollHeight - clientHeight > 1` (and width) of each `[data-slot]` and `[data-slotpanel]` element, and the largest square no `[data-slot]` rectangle covers (over 10 px fails). The iframe is exactly w by h, so `innerWidth`, `dvh` and container queries are those of a window of that size; a page that overflows gets a scrollbar inside the iframe and is caught by the scroll test.

Tried to break it on the builder (a throwaway copy of the branch with extra fixture pages, never committed; 96 page loads; nothing ran on the worker host that runs the games):

| Page | Must | Got at all 8 sizes |
|---|---|---|
| the branch's own 7 pages (cockpit, demo with 0, 2, 5 needs, 20 orders, the two must-fail pages) | as the card says | V: 56 of 56 as the card says; the overflow fixture fails with 668 to 1272 px, Classic with a scroll height of 4627 to 11167 px |
| a panel whose inner box (`block-size: 100%; overflow: hidden`) holds 60 rows | fail (content is cut, no scroll, no ellipsis) | **passes** |
| 60 rows under the panel root marked `data-scroll`, which does not scroll | fail | **passes** |
| a rail whose only panel hides itself | fail (a blank column 400 to 700 px wide) | **passes** ("4 slots, largest gap 8 px") |
| `.d-main` forced to 130dvh (the shell clips it) | fail (the strip and the rail's end are below the window) | **passes** |
| the demo with the "..." menu opened before measuring | (probe of R0) | fails: "slot band overflows its box by 323 px" (the clipped sheet), so the check does see an open sheet inside a clipping slot |
| a must-fail dump that holds a measurement error | fail the run | **"ok, failed as it must"**, exit 0 (R2) |

So the check does fail on the overflow it was built for (a slot panel root taller than its box), on page scroll and on a missing slot; it does not see clipping below the root, abuse of `data-scroll`, blank slots or a frame larger than the window. R1 closes these.

Other false-pass risks:

- Only the first phone tab is measured at 390x844; Map and Progress are never loaded (the tab is page state, not addressable). Add `?tab=<name>` to the shell and two pages to the list. Needed before S6 at the latest.
- The viewport test (`fit.go:168`) compares the iframe's size with the size in the file name; both come from the same two numbers, so it cannot fail. Harmless, but it is not evidence.
- Fonts: the builder's fonts are not the user's. Fixed boxes make the outer numbers font-independent; what a font changes is what is cut inside a box, which the check cannot see until R1 lands. Keep the pictures for the eye.
- A panel that fails to load or is still `loading` after the 5 s wait is measured as it is (an empty box that covers its slot). The judge should fail a page with a `[data-slotpanel]` whose `data-state` is `loading` or that holds `.d-empty` "failed to load".
- Nothing asserts a least size (a tile squeezed to 10 px "fits"). See 2, the floor of the growing panel.

Runtime: about 44 s on the builder for 96 page loads (V: the script's own timestamps), so about half a minute for the branch's 56; the wait in the builder's queue was far longer than the run. The fit check is not part of `tools/ci/check.sh` or of the merge gate (V: no `uishots` in `tools/ci/`), so "a slice that fails the fit check is not done" rests on each agent running it. Ask the merge step to run `tools/uishots.sh --fit` for branches that touch `internal/watch/web/` (a report line; not this card's file). The node tests skip when node is missing (`t.Skip`): the builder has node (V: v20.19.2), so they run there; a gate host without node would pass the cells golden without running it. Make the skip a failure when `CI` or the gate's variable is set.

## 4. Code quality, files outside the brief, security

- Files outside the brief's list, all now in the card's `owns:`: `internal/watch/panels.go` (one additive manifest field, `omitempty`; low risk), `internal/watch/http_view.go` (one exact path added to the static allowlist; low risk), `testdata/layout_edit.js` (the grid editor's harness now cuts the source at the marker comment `// ---- Edit layout on a cockpit`; a rename of that comment would make the old test read the cockpit editor too, but `TestCockpitEditorContract` fails first; acceptable), three new testdata files. No compat fixture is touched (V: diff stat).
- Route allowlist: the server serves `/watch/layouts/cockpit.json` only. The shell's test `/^(mockup|live|command|cockpit[a-z0-9-]*)$/` (`shell.js:376`) lets `?layout=cockpit-anything` ask for a file; the name cannot leave the directory (letters, digits, hyphen), the server answers 404 and the shell falls through to the stored or default layout. Safe. It is a test hook in production code (the fixture viewer's `cockpit-demo`, `cockpit-overflow`): acceptable, say so in the comment.
- `fitHarness` (`fit.go:44-53`) prints `u` into a script with `%q`, which does not escape `<`; `u=/watch?</script>...` breaks out. Only in the fixture viewer on the loopback address, so a nit: use `json.Marshal`.
- `framedPage` relaxes `frame-ancestors` only in the fixture viewer and only for `fitframe=1`; the product's CSP is untouched (V: no change to `pageCSP`).
- `layout.go` `CheckSlots`: clear. `sortedKeys` generic helper is fine. `LayoutSlots.UnmarshalJSON` decodes twice (once typed, once raw); fine at this size, and it goes away with R6.
- `shell.js` `applyCockpit` rebuilds the slot elements whenever the arrangement signature changes and moves the existing roots; panels keep their DOM and listeners. Good. `ck.plan.left` (ids named and not mounted) is computed and shown nowhere: show it in the editor bar ("not built yet: ...") or drop it.
- `cockpit.json` old-style `breakpoints` hold only the picture, so an older shell (a rollback with a stored cockpit) shows the picture and the meters and nothing else. Not blank, as J asked, but poor: copy Classic's placements instead (nit N1).

## 5. What S2 to S7 need that S1b does not give yet

| Need | For | Missing |
|---|---|---|
| A panel knows where it stands | S2 (`hero-meters` serves the band of both pages), S3 | `ctx` has no `slot` field; only CSS can tell (`data-slotpanel`). Add `ctx.slot` (`band`, `stage`, `strip`, `rail`, `tab`, or empty). |
| Panels adapt to their box | S3 (strip cells "drop their second words" on short windows), S4, S5 (zone rows) | Only the stage is a size container (`shell.css:143`). Make every slot panel root `container-type: size` with one name in the shell, or each panel invents its own. |
| How the band shares its width | S2 | `.d-band-main` is `nowrap` and clips; nothing says who shrinks. Rule needed in shell.css: the first band panel `flex: 1 1 0; min-inline-size: 0`, the others `flex: none`. S2 does not own shell.css. |
| Things that open over other tiles | S2 (the sentence's popover), S5 (the zone menu: beside the map on a wide rail, over the Orders tile otherwise), S6 (the Log drawer) | Every slot clips (`shell.css:142`), the band clips (`:122`). The shell needs one layer outside the slots (`ctx.sheet(...)` or a `[data-layer]` element above `.d-main`) and the fit check must ignore it. R0 is the first case of this. |
| The fit check on real panels | S2 onward | The pages list is fixed in `tools/uishots.sh` and the states (`needs=0/2/5`, `orders=20`) are read from the address by the stand-in panels' own code. Real panels read endpoints: the fixture viewer needs a state parameter that changes what its endpoints answer, and each slice must replace its stand-in in `cockpit-demo` with the real panel (rule for the briefs). |
| The fit check with something open | S5 (zone menu open and closed, brief Do 6) | The harness only reads. It needs one step before measuring (`&click=<selector>`). |
| Phone tabs | S6 | Addressable tab (`?tab=`), remembered tab, tabpanel roles, focus kept (section 2), each tab in the fit check. |
| The wide arrangement | S4, S5, S6 | R4. |
| The minimum of the growing tile | S5 | Section 2. |
| The slot list in shell.json ends | S2, S7 | Section 2. |

## Nits

- N1 `cockpit.json` breakpoints: Classic's placements instead of the picture alone.
- N2 `fit.go`: `%q` into a script; the viewport test that cannot fail.
- N3 Strip cells `1fr 1.25fr 1.1fr` as in the renders (a `sizes` key is not needed; CSS on `:nth-child` is enough).
- N4 Phone: the picture is inset 8 px; J draws it edge to edge (390 x 292).
- N5 `shell.css`: `data-has` has no rule for `p` alone or for `tp` (a phone layout with no stage); the tabs then land in an implicit row. Add `"p"` or make uilint require a stage panel that is mounted.
- N6 The editor bar's hint does not say that a picked panel that is not built appears only after Save and a reload (the toast does, once).
- N7 uilint: `slot: true` with `fit: true` should be a finding; an id in `layout.slot` whose manifest already says `slot: true` should be one.
- N8 The card's result names the builder host; prose should say "the builder".

## To verify live (nobody has)

1. `/watch` with no parameter on a profile with a stored Classic or user-made layout, before and after the deploy: same page (the golden says the plan is the same; the eye has not seen it).
2. `/watch?layout=cockpit` on the live viewer at the user's window (about 1914x930): the picture and the meters, no scroll, the "..." menu opens fully and offers Edit layout and Classic.
3. In the cockpit editor: swap two tiles, drag both dividers, Save, reload; then Reset to default, and Classic from the menu: the old page is back unchanged.
4. `/watch?mode=overlay` with a stored cockpit: the OBS strip is the same as with Classic (after R7).
5. The same page in the user's own browser fonts and with the game picture running (the fixture viewer has a still picture).

## How this was checked

- Read: `git diff origin/main...agent/uis1b` (24 files), J sections 2 and 3, the S1b brief, the briefs README, `mockups/v2/watch.html`, `research/iterations/final/src/watch.html`.
- Golden: rebuilt from `origin/main`'s `layout.js` and compared (section 1).
- Go: `go test -p 1 -count=1 -run 'Compat|Layout|Cockpit|Templateless|Manifest' ./internal/watch/` on the branch: ok. A throwaway test for R6 (not committed).
- Fit check: a throwaway copy of the branch with probe fixtures (not committed anywhere), run on the builder. Output summarised in section 3. The golden rebuild ran there too.
