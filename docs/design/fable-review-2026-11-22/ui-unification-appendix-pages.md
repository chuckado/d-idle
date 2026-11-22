# UI unification review, appendix: page and panel inventory

> status: current (review of 2026-11-22, card fableui). Belongs to [ui-unification.md](ui-unification.md). Docs only.

Tags: **V** = verified by reading the named file at the commit this branch starts from (`28eb917f2`), or by looking at the named picture; **U** = not verified (inferred, or behaviour that needs a browser). Paths are relative to `internal/watch/` unless they start with another top directory. "Last change" is `git log -1` of the page's main file, not a judgement that it was a redesign. Nothing here was run in a browser.

## A1. Pages (routes that serve HTML)

| # | Route | Purpose | Entry points (all found) | Links back / onward | Phone | Desktop | Last change | Health |
|---|---|---|---|---|---|---|---|---|
| 1 | `/watch` (cockpit, the default layout) | live screen: picture, status band, Orders, Map, Quests, Progress | root redirect; every sub-page's "Live screen" or "Back" | band links Inventory, Skills, Points, Vault, Return screen (hidden when narrow or when the bot has no return screen); `...` menu: Characters, Ethereal Vault, Quests drawer, Getting started, Loot rules, Edit layout, Help, Theme, Item labels, Log, Return screen, Send support bundle, six layout presets; break-glass Reset (V `web/watch.html:19-28`, `web/shell.js:823-889`) | designed: band, picture, four tabs Orders/Map/Quests/Progress (V `web/layout.json` `phoneTabs`; picture `docs/ui/map-phone/after-fresh-390.png`) | designed: band, stage, strip, rail | 2026-11-21 | **clean core, cruft around it** (the hidden pre-shell page rides along, row 2) |
| 2 | `/watch` under a grid preset (`Classic`, `Simple`, `Standard`, `Stream`, `Debug`; also `?layout=mockup|live|command`) | the pre-cockpit page: cards Thoughts, Now, Where to, Queue, Rules, Spend, Chronicle, Diary, Quests, Still locked, Cain's notes, Atlas, Alerts, Journal, with its own tab strip | `...` menu, group Layout (closed by default); the layout note "Classic is in the menu"; the "Use Classic" button of the failed-panel note (V `web/shell.js:403`, `:934`, `:950`) | same band and menu as row 1; way back to the cockpit is `...` > Layout > "Layout: Cockpit" (V `web/shell.js:788`, `:871-878`) | flow layout with the old tab strip (V `web/watch.html:37-49`) | 12-column grid, editable | presets 2026-11-10 to 2026-11-19; `web/app.js` 2026-11-11 | **duplicate** of row 1 for Orders, Map, Quests, Journal, Alerts; the only home of five features (main doc C2) |
| 3 | `/watch/inventory` (character canvas, `charPage`) | paper doll, mercenary, pack, stash, belt, item actions, gear sets | band link "Inventory" on `/watch`; Skills page; Points page | band: Live screen, Skills, Policy, Theme. Phone: only "Live screen" stays in the band, the rest is a `...` list of Inventory, Skills, Policy, Return (V `invui.go:88-93`, `web/inventory-canvas.css:265`). **No Vault, no Points, no Characters, no Help.** | designed fixed canvas with a two-part tab bar (V `invui.go:105`) | designed fixed canvas | 2026-11-15 | **clean page, broken nav**; two placeholder texts (main doc C7) |
| 4 | `/watch/inventory?view=classic` (`invPage`) | the older cards page of the same data | **none**: no link anywhere (V: grep of `href`, `location.` over `web/` and `*.go`; route `invapi.go:181-189`) | band: Live screen, Skills, Policy | flows | cards | 2026-11-15 | **dead route** (kept "until the canvas has been seen live") |
| 5 | `/watch/skills` | skill trees of the class, kind filter; spends a skill point | band link on `/watch`; Inventory; Points ("open the skill trees", V `web/points.js:41`) | band: Live screen, Inventory, Policy, Theme (V `skillsui.go:18-21`). **No Points, no Vault.** | one tree per row, capped width (V `docs/ui/friend-feedback-2026-11-15.md` row a) | three trees | 2026-11-21 | **cruft**: says "Read only ... nothing here changes them" while the page spends points (main doc C6) |
| 6 | `/watch/points` | attribute points, gear requirements, build picker ("Automatic Allocate Skills"), autopilot, recent actions | band link on `/watch` only | band: Live screen, Skills, Inventory, Theme (V `pointsui.go:18-21`). **No Vault.** | single column, 48rem cap (V `web/points.css:3`) | the same single column | 2026-11-21 | **cruft**: the build picker is a file list (main doc D1); `window.confirm` |
| 7 | `/watch/vault` | every recorded drop: filter, reveal, sell, equip, capacity | band link "Vault" on `/watch`; `...` menu "Ethereal Vault"; Orders More sheet; log drawer Items tab; hero tile Ether chip (V `web/shell.js:847`, `web/panels/orders/panel.js:103`, `web/panels/log/panel.js:38`, `web/panels/hero-vitals/panel.js:74`) | **one link: "Back" to `/watch`** (V `web/vault.js:39-40`, `:1202`). No Inventory, Skills, Points. Does not load `shell.js`, so no band, no Theme button (V `web/vault.html`) | designed page scroll (V picture `docs/ui/mobile/mobile-vault-dark-after.png`) | designed, docked item card | 2026-11-22 | **clean page, island**; header still sells capacity for Ether (picture above) |
| 8 | `/watch/characters` (also `/watch?view=characters`) | switch, create, delete, restore a hero | `...` menu "Characters"; Orders More sheet (V `docs/charselect-ui.md`) | "Back" to `/watch` (V `web/characters.js:80`) | bottom sheets | cards | 2026-11-16 | clean, island; own shell like the Vault |
| 9 | `/watch/return` | "While you were away": moments, deaths, rules | automatic redirect from `/watch` when a visit is due; band link "Return screen" (hidden when narrow); `...` menu; onboarding banner (V `returnscreen.go:912-916`, `web/components.css:268`) | "Go to the live screen" (V `returnscreen.go:1045`) | flows (U) | flows (U) | 2026-11-20 | **old shell**: `app.css` classes `.btn`, `.card`, server-rendered template |
| 10 | `/watch/policy` | policy sheet: how the hero handles chores | band link "Policy" on Inventory and Skills only; **not reachable from `/watch`** in any layout (V: no `watch/policy` in `web/`) | "While you were away", "Live screen" (V `policysheet.go:819`) | flows (U) | flows (U) | 2026-11-09 | **old shell, half-orphan** |
| 11 | `/watch/pacts` | pact offers | **none** (V: no `watch/pacts` link in `web/` or the page templates) | none (V `pactui.go:17`) | bare (U) | bare | 2026-11-08 | **orphan, unstyled** |
| 12 | `/watch/hall` | Hall of Fame | **none as a link**: the return screen prints the address as text ("Open /watch/hall to read it", V `hall.go:514`) | "Back to the return screen" (V `hall.go:562`) | bare | bare | 2026-11-08 | **orphan, unstyled** |
| 13 | sign-in page | token gate | automatic | to `/watch` after sign-in (U) | wizard style (`login.css`) | same | 2026-11-08 | third visual style (`wz-`), acceptable |
| 14 | setup page (`cmd/d2dsetup/page/`) | tester setup, six screens | separate program | none to the viewer (U) | U | U | 2026-11-06 | separate product surface, out of scope for unification (limits G) |
| 15 | dashboard and tile viewers (`cmd/d2dash`, `cmd/d2watch`, `cmd/d2dview`, `internal/status`) | developer and operator views | separate programs | none | not designed (U) | yes | 2026-10-30 (`internal/status/hub.go`) | developer tools: leave alone |

Redirects: `/watch/video` and `/watch/game` go to `/watch` (V `http_view.go:70-72`).

## A2. Panels of `/watch` (26 directories under `web/panels/`)

Used by the cockpit (V `web/layout.json` slots, `web/shell.json` `drawer`): `statusline`, `hero-meters`, `needs-count` (band); `video` (stage); `progress`, `quest-strip`, `health` (strip); `needs-you`, `orders`, `zone-map`, `area-map` (rail); `quest-list` (phone tab Quests); drawers `log`, `onboarding`, `quest-board`. That is 15.

Not in the cockpit, only in grid presets (V: ids per preset read from `web/layouts/*.json`): `activity`, `alerts-count`, `currency`, `gear`, `hero-status`, `hero-vitals`, `hero-detail` (manifest says `mergedInto: hero-vitals`), `map`, `tally`, `why`, `zoneclear`. That is 11 directories whose only users are the presets of row 2.

Three displays of quests exist at once: `quest-strip` (one line), `quest-list` (phone tab), `quest-board` (drawer), plus the pre-shell `quests` card of `web/watch.html:82-85` drawn by `web/unlockui.js`. Two maps of the area exist: `area-map` (cockpit) and `map` (presets).

## A3. Sheets, drawers and dialogs (overlay surfaces)

| Surface | Where | Class | Opens from |
|---|---|---|---|
| side drawer | log, Getting started / Help, Quests board | `d-drawer` (V `web/shell.css`, 18 rules) | `...` menu, status sentence |
| tile sheet | Orders "More", zone menu | `d-tilesheet` (V `web/shell.js:114-147`) | panel buttons |
| modal confirm | Reset, Pause | built in `web/resetui.js`, `web/pauseui.js` (`role=dialog`) | menu, Orders More |
| loot rules sheet | loot rules editor | `lr-sheet` (V `web/lootrulesui.js`) | `...` menu on `/watch` only |
| support sheet | send support bundle | `role=dialog` in `web/supportui.js` | `...` menu |
| vault sheets | sell, identify all, marked, capacity | `vt-sheet` (V `web/vault.css`, 7 rules) | Vault page |
| characters sheets | switch, delete, create, restore | `cs-sheet` (V `web/characters.css`) | Characters page |
| gear sets popover | gear sets on the canvas | `cv-pop` (V `invui.go:143`) | Inventory |
| pre-shell sheet and scrim | old order sheet | `#sheet.sheet`, `#scrim` (V `web/watch.html:120-122`) | old cards (presets only) |
| browser dialog | lock a build, confirm an unsupported skill, revoke all rules, pay a charter, update error | `window.confirm` / `alert` (V `web/points.js:142,144`, `web/skills.js:110`, `web/app.js:398,470`, `web/update.js:41`) | Points, Skills, old cards |

Eight sheet implementations and one browser dialog for the same job.

## A4. Things a phone user cannot do at all, or only by luck (V unless marked)

1. **Edit layout**: refused under 700 px with a toast (V `web/shell.js:683`). Correct: keep it desktop-only, but hide the menu entry on a phone instead of offering a button that refuses.
2. **Read why a button is disabled**: reasons are set as the `title` attribute (hover text), which a touch screen never shows. 12 such writes in `web/vault.js` (for example `:388`, `:417`, `:425`, `:1055`), 23 in `web/inventory-canvas.js`, 7 in `web/panels/quest-board/panel.js`, and the "N need you" pill's list of what needs you (V `web/panels/needs-count/panel.js:27`). Some of these also print the reason in a cost line; which ones do not is U (needs a browser pass).
3. **Reach the Vault, Points, Characters or Help from the Inventory**: the phone `...` list has four entries, none of them these (V `invui.go:93`).
4. **Reach the Policy sheet from the live screen**: no link in any layout (row 10).
5. **Reach Pacts and the Hall of Fame**: no link (rows 11, 12); a phone user cannot type the address comfortably.
6. **Revoke a learned rule, read Spend, Cain's notes, the Atlas, "Still locked", or switch alert sound and notifications**: these live only in the pre-shell cards, which the cockpit does not show; the way there is `...` > Layout (a closed group) > Classic, which reloads the page into another layout (main doc C2). The log drawer covers Journal, Story and Alerts rows but not the alert settings (V `logtab.go:24-30`).
7. **Open the Return screen from the band**: the band link is hidden when narrow (V `web/components.css:268`); it is in the `...` menu, so this one is reachable.
8. **Change the theme on the Vault and Characters pages**: no Theme button there; they read the stored choice or `?theme=` (V `web/vault.js:1188-1191`).
9. **Drag an item**: `web/inventory-drag.js` handles touch pointers (V `:194`); whether the long-press drag works beside page scroll on a real phone is U.
