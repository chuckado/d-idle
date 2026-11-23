# E11 deletion plan (PLAN ONLY: nothing is deleted by this document)

> status: proposal for the user to approve. Source: ui-unification.md C2, C3, C4, C9, C15, F2; user answers 2026-11-22 (Q2 compatibility relaxed for shipped presets, Q3 move rule revoking and alert settings, DROP Cain's notes, Atlas, Spend, Still locked; Q8 `Old layout` with a Switch back bar; Q11 keep the Stream page). E10 (the Settings sheet, card uisettings) is the precondition.
> Claims are from grep of the tree at the E10 merge: V = read from the files, U = not checked live.

## 1. The one thing to decide first: Stream (Q11)

`web/layouts/stream.json` names 14 panels: activity, currency, gear, hero-vitals, log, map, needs-count, onboarding, orders, progress, statusline, tally, video, why (V). Keeping the Stream page as it is therefore keeps those panel directories. Of the 11 "only by old presets" panels in the review, only `hero-detail` and `zoneclear` are not used by `stream`, the cockpit or `shell.json` (V; `zoneclear` is also named by no layout). So E11 has two shapes:

- **E11a (recommended first):** delete the pre-shell cards and scripts, the presets `classic`, `simple`, `standard`, `debug`, the unlisted `mockup`, `live`, `command`, and the panels no kept layout names. Keep `stream` and its 14 panels.
- **E11b (later, optional):** rebuild Stream from cockpit panels, then delete the remaining panels.

User question: is E11a acceptable, with Stream keeping its panels for now?

## 2. E11a deletes (exact)

Pages and scripts (the old page inside `/watch`):
- `internal/watch/web/app.js` (+ route `/watch/app.js`, `http_view.go:73`), `unlockui.js` (route `UnlockUIRoute`, `unlock.go:30`), `alerts.js` (route `AlertsUIRoute`, `unlockalerts.go:35`, served at `:772`), `web/update.js` only if nothing else loads it (U: `update.go:161` injects it into the page; check), `web/gear.js` is NOT deleted blind: `repairapi.go:14,49` serves it and `gearjs_test.go` tests it (decide with the Inventory owner).
- In `internal/watch/web/watch.html` (123 lines): the 14 old cards (`data-lid` list = `shell.json` `layout.compat`), the old `#sheet`, `#scrim`, `#toast`, the tab strip (lines 34-112), script tags at 119-122. The file stays: it is the template the cockpit shell loads into (V: `http_view.go:64`).
- `web/app.css:53-67` (compat CSS) and the old colour aliases at `:5`, `:188-189` (C15). `web/app.css` itself CANNOT go yet: `returnscreen.go` and `policysheet.go` link it and use `.btn` (V). That is C9: move those two pages to `d-btn` first (separate small card), then delete the file.
- Routes/handlers only the dropped cards used (U, list from the E10 card): `/api/bestiary` (`unlockbestiary.go`), `/api/atlas` (`unlockatlas.go`), the Spend list and `/api/currency` spend part (the band still reads currency), `/api/unlocks` locked list (`unlock.go`, `unlockui.js`), `/api/chronicle` and `/api/questlog` are kept until the log drawer decision F3 (Cain's notes, Atlas, Spend, Still locked are dropped, Chronicle/Quests/Diary are not in this answer: ask).

Layouts (Q2 relaxed): delete `web/layouts/{classic,simple,standard,debug,mockup,live,command}.json`, their route cases (`http_view.go:85-86`), the menu entries (`shell.js:403`, `:773`, `:788`, `:436`) except `cockpit` and `stream`. Keep the layout and panel FILE FORMATS and a hero's own saved layout loading (Q2). `Old layout` link and `choosePreset('classic')` (`shell.js:899`) go with the old page; the Switch back bar goes with it (it has nothing to switch back from).

Panels not named by `cockpit.json`, `shell.json` or `stream.json` (V by grep): `hero-detail`, `zoneclear` (check `zoneclear.go` serves an API used elsewhere before removing the panel directory).

## 3. The stub (Q2)

A saved layout or `?layout=<removed preset>` naming a removed preset or panel loads the cockpit and shows one line in place of the panel: "This panel moved. The new layout has it in the band, Progress or the log." One release, then the stub text goes. Texts through `textContent`.

## 4. Tests that reference the old page (V, files; each changes as noted)

- `compat_test.go`, `layout_test.go`, `watchlayout_test.go`: assert the 14 compat ids, preset files and `?layout=` presets: change to "removed presets load the cockpit with the stub", keep the file-format fixtures under `testdata/compat/` (they prove old FILES still load).
- `nocards_test.go`, `herofast_test.go`, `topband_test.go`: reference `unlockui`, `alerts.js`, the old cards: drop those assertions.
- `watch_test.go`, `currency_test.go`, `ordersmap_test.go`, `panels_test.go`, `repairapi_test.go`, `pauseui_test.go`, `onboarding_test.go`, `reasonlines_test.go`, `linkmatrix_test.go`: read `watch.html` or `app.js` strings; each needs a one-line review (many assert text now only in `app.js`).
- `uitokens_test.go`, `returnscreen_test.go`: `app.css` (stays until the C9 card).
- `unlock*_test.go` for bestiary and atlas: removed with their routes. uishots pages for the old page (`classic` fixtures) and the `settings-*` pages stay.
- New: a test that every removed preset id loads the cockpit and shows the stub.

## 5. Order and size

1. Stub and removed-preset redirect (small).  2. Delete cards, scripts, routes, tests (large, one PR).  3. C9: Return and Policy to `d-btn`, then delete `app.css` (separate PR).  4. E11b only if the user wants Stream rebuilt.

Risks: a hero's saved layout that names a removed panel (stub covers it); `update.js` injection (`update.go:161`) must keep the updater banner working on the cockpit.
