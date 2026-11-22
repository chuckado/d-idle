# UI unification review: navigation, cruft and readable build pages (2026-11-22)

> status: DESIGN REVIEW, nothing built (card fableui). Docs only. The user decides the open questions in section F; no ADR or decision line is written here. Page table: [ui-unification-appendix-pages.md](ui-unification-appendix-pages.md).

Tags: **V** = verified by reading the named file at commit `28eb917f2` or by looking at the named picture in `docs/ui/`; **U** = unverified (inferred, or needs a browser); **EXP** = Diablo II play experience, not looked up and not measured. Paths are relative to `internal/watch/` unless they start with another top directory. No page was opened in a browser and no new screenshot was taken for this review (section G).

User request, 2026-11-22: "I'm really happy with the direction we're going in but I feel like we've left a lot of cruft and old UI elements. And there's some places that could use a fix-up in unification. Things like you can't get to the Vault from the inventory. Also the build pages could use a readability pass: too much information. Each build should have a summary and a quick pros and cons, and maybe a 'recommended for' section with a list of tags like farming, bossing, leveling with a star rating one to five for each. Defensive should also be in there."

## Summary

1. **The direction is right; keep it.** The cockpit (band, picture, four phone tabs), the character canvas, the Vault page and the token and component layers are sound and consistent inside themselves. Nothing below asks for a new look.
2. **The problem is that every page built its own frame.** There are five different page headers, eight sheet implementations, four toast implementations and three button families (appendix A3, section C). The `...` menu with Characters, Vault, Help and Loot rules exists only on `/watch` (V `web/shell.js:1035`: `setupMenu` runs only when the page has the pre-shell card area).
3. **The user's example is real and it is the smallest case.** The Inventory links to Live screen, Skills and Policy; on a phone to Inventory, Skills, Policy, Return. It has no link to the Vault, Points, Characters or Help (V `invui.go:88-93`). The Vault links only "Back" (V `web/vault.js:39-40`). Section B lists every dead end.
4. **The largest cruft is a whole second page.** The pre-cockpit page (14 cards, its own tabs, `web/app.js`, `web/unlockui.js`, `web/alerts.js`, 11 panels, 8 layout files) still ships and loads inside `/watch`, and five features exist only there (C2). Recommendation: move those five, then delete, which needs a decision on the compatibility promise of 2026-11-09 (F2).
5. **The build "pages" are a list of file names.** A Paladin sees six rows such as "paladin hammerdin zeal endgame90" that are one build at different stages; each row shows tier, warnings and a 30-level table, and no sentence says what the build is (D1). Proposal: one card per build with summary, pros, cons and seven star ratings from a data file, the tables behind "Details" (D2 to D6).
6. **Recommend deleting more than adding**: section E adds one shared navigation component and one build catalogue file, and removes about 20 files and four duplicate patterns.

## A. Page inventory

In the appendix: A1 pages (15 rows with entry points, links back, phone and desktop state, last change, health), A2 panels (15 used by the cockpit, 11 only by old presets), A3 overlay surfaces, A4 what a phone user cannot do.

## B. Navigation map and gaps

### B1. Who links to whom today (V, from every `href` and `location.` in `web/` and the page templates)

Rows link to columns. `band` = a button in the page header; `menu` = inside `...`; `-` = no link.

| from \ to | Live | Inventory | Skills | Points | Vault | Characters | Quests | Map | Return | Policy | Help |
|---|---|---|---|---|---|---|---|---|---|---|---|
| Live `/watch` | . | band | band | band | band, menu, 3 more | menu, Orders More | tab, drawer | tab | band (wide only), menu | **-** | menu |
| Inventory | band | . | band | **-** | **-** | **-** | **-** | **-** | phone list only | band | **-** |
| Skills | band | band | . | **-** | **-** | **-** | **-** | **-** | **-** | band | **-** |
| Points | band | band | band, text link | . | **-** | **-** | **-** | **-** | **-** | **-** | **-** |
| Vault | Back | **-** | **-** | **-** | . | **-** | **-** | **-** | **-** | **-** | **-** |
| Characters | Back | **-** | **-** | **-** | **-** | . | **-** | **-** | **-** | **-** | **-** |
| Return | button | **-** | **-** | **-** | **-** | **-** | **-** | **-** | . | **-** | **-** |
| Policy | button | **-** | **-** | **-** | **-** | **-** | **-** | **-** | button | . | **-** |

Of 80 cells, 56 are empty. Only the live screen reaches most pages; every sub-page is a spoke that must go back through it. Pacts and the Hall of Fame are not in the table because nothing links to them (appendix rows 11, 12).

### B2. Dead ends and traps (V unless marked)

1. Inventory to Vault, and Vault to Inventory: none. This hurts most after **Equip from the Vault** (docs/vault.md "Equip"): the item goes to the hero and there is no "see it on the hero" link.
2. Inventory, Skills, Vault to Points: none. Points is reachable only from the live band.
3. Skills to Points: none, although Points links to Skills. The two pages spend the same resource family.
4. Quests and Map are tabs of the live screen and cannot be linked from other pages by a button, although the address `/watch?tab=Quests` already works (V `web/shell.js:616-618`, `ck.want`).
5. Help exists only in the `...` menu of `/watch`. A user confused on the Inventory or the Vault has no Help.
6. Policy is linked from Inventory and Skills but from nowhere on the live screen.
7. **"Back" on the Vault and Characters pages goes to `/watch`, not `/watch?live=1`** (V `web/vault.js:1202`, `web/characters.js:80`). `/watch` without `live` redirects to the Return screen when a visit is due (V `returnscreen.go:912-916`); the Vault page does not poll `/api/thoughts`, which is what marks a look in progress (V `returnscreen.go:926-928`). So after a long Vault session "Back" can land on "While you were away" (U: the timing was not run). Every other page links `/watch?live=1`.
8. The way back from Classic to the new layout is `...` > Layout (a group that starts closed, V `web/shell.js:880`) > "Layout: Cockpit". The user already asked "how do I get back to the new UI". The word "Cockpit" is never shown elsewhere as the name of the default.
9. Three names for one place: "Live screen" (sub-pages), "Back" (Vault, Characters), "Go to the live screen" (Return). Two names for another: "Vault" (band) and "Ethereal Vault" (menu, page title).
10. Page titles: sub-pages show the brand "d.idle" in the `h1` and the page name only in a card (`skillsui.go:16`, `:25`); the Vault and Characters pages show the page name as `h1` and no brand. There is no visible "you are here" except on the Inventory phone list (`aria-current`, V `invui.go:93`).

### B3. Recommendation: one navigation component, the same on every page

One list of destinations in one file (`web/nav.json`), drawn by one module (`web/nav.js`) that every page loads, including the Vault and Characters pages that do not load `shell.js`. The current entry is marked with `aria-current="page"`.

**Five destinations, in this order:** `Live` (`/watch?live=1`), `Gear` (the Inventory), `Skills` (Skills, Points and the build as three sub-tabs of one area, F4), `Vault`, `More`. `More` opens one sheet, the same on every page: Characters, Return screen, Policy, Loot rules, Help, Theme, and on the live screen only: Edit layout (hidden on a phone), Item labels, Log, Send support bundle. Reset and Pause stay where they are (the break-glass Reset in the band of the live screen).

Why five: the phone band already holds Inventory, Skills, Points, Vault, Reset and `...` in one row at 390 px with 5 px padding (V picture `docs/ui/map-phone/after-fresh-390.png`); there is no room to add a sixth, and Points belongs with Skills.

**Phone (under 700 px): a bottom tab bar.** The in-page tabs of the live screen (Orders, Map, Quests, Progress) stay at the top under the picture: the bottom bar answers "which page", the top tabs answer "which part of this page". The Inventory keeps its own two-part tab bar the same way.

```
+--------------------------------------+        +--------------------------------------+
| d.idle   WALKING Rogue Encampment    |        | d.idle   Gear              Sorc  L24 |
| life ---- 36  mana ---- 105  2 need  |        | [search pack and stash]  Gold 12,400 |
+--------------------------------------+        +--------------------------------------+
|            game picture              |        | [ Hero and merc ] [ Pack and stash ] |
| [Full screen][Low rate][Bright][Stats|        |                                      |
+--------------------------------------+        |        paper doll / cells            |
| [Orders] [ Map ] [Quests] [Progress] |        |                                      |
|                                      |        |  selected item: ... [Sell] [Stash]   |
|        the chosen tab's panels       |        |  -> "Find more like this in Vault"   |
|                                      |        |                                      |
+--------------------------------------+        +--------------------------------------+
| *Live*   Gear   Skills  Vault  More  |        |  Live  *Gear*  Skills  Vault   More  |
+--------------------------------------+        +--------------------------------------+
```

**Desktop (700 px and wider): the same five in the top band**, left of the status sentence's row end, where the links are today. Sub-pages get the same band; the page name sits beside the brand.

```
+----------------------------------------------------------------------------------------------+
| d.idle | *Live*  Gear  Skills  Vault  More v |  WALKING Rogue Encampment ...  life mana | Reset |
+----------------------------------------------------------------------------------------------+
```

**Header rule for every page:** brand, navigation, then the page's own status or tools. No "Back" button on a top-level page (the navigation is the way back); "Back" stays only inside a flow that has steps (create a hero). Every link to the live screen is `/watch?live=1`.

**Contextual links on top of the global ones** (cheap, each one line): Vault item sheet after Equip: "Open Gear"; Gear selected item: "Open the Vault" in the toolbar next to Sell; Skills banked-points line: "Attribute points and build"; Quests row with an area: "Show on map" (`?tab=Map`); Needs-you rows that are about points or gear: a link to that page.

What NOT to do: no hamburger drawer on a phone (five items fit; a drawer hides the Vault again), no breadcrumbs (two levels only), no second global bar on desktop.

## C. Cruft list

Size: S = under a day, M = one to two days, L = more. Risk is the risk of breaking something a user has.

| # | What | Evidence (V) | Replace by | Delete or unify | Size / risk |
|---|---|---|---|---|---|
| C1 | Five page headers, each hand-written in a Go string or built in JS | `invui.go:15-25` and `:83-95`, `skillsui.go:13-23`, `pointsui.go:13-23`, `web/vault.js:37-57`, `web/characters.js:80`, `returnscreen.go:1044-1045`, `policysheet.go:818-819` | the one navigation of B3 | UNIFY | M / low |
| C2 | The pre-cockpit page inside `/watch`: 14 cards and a tab strip in `web/watch.html:34-112`, `web/app.js` (509 lines), `web/unlockui.js`, `web/alerts.js`, `web/gear.js`, `web/update.js`, the compatibility CSS `web/app.css:53-67`, the old `#sheet`, `#scrim`, `#toast` | appendix row 2; `web/shell.json` `compat` lists the 14 ids | first MOVE the five things that live only there: revoke a learned rule and "Revoke all" (`web/app.js:92-152`, `:469-470`), Spend, Cain's notes, the Atlas with "Still locked" (`web/unlockui.js:96`, `:123`), alert sound and desktop notifications (`web/alerts.js:14-54`). Rules and alert settings go into `More` > Settings; the three collections become tabs of the log drawer or are dropped (F3) | then DELETE | L / medium: needs F2 |
| C3 | Six layout presets in the menu plus three unlisted files | `web/shell.js:403`, `:788`; `web/layouts/` has `classic`, `simple`, `standard`, `stream`, `debug`, `cockpit`, `mockup`, `live`, `command`; `mockup`, `live` and `command` are reachable only by `?layout=` (`web/shell.js:436`) | Cockpit, plus `Stream` if the OBS strip is used (the overlay mode falls back to `standard`, `web/shell.js:449`) | DELETE `mockup`, `live`, `command`, `simple`, `debug`; decide `classic`, `standard` with C2 | S / low for the three unlisted; rest with F2 |
| C4 | 11 panel directories used only by those presets | appendix A2 | nothing; the cockpit panels cover them (`hero-vitals` by the band and Progress, `map` by `area-map`, `alerts-count` by `needs-count`) | DELETE with C2 (or leave stubs, F2) | M / with F2 |
| C5 | The old cards Inventory | appendix row 4: route with no link | the canvas | DELETE `invPage` and the `?view=classic` branch (`invapi.go:183-186`) | S / low |
| C6 | Stale texts | Skills: "Read only: points are spent by the player in the game, nothing here changes them" (`skillsui.go:28`) while the page spends points (`web/skills.js:102-115`). Vault header "Upgrade +200 for 20 Ether" (picture `docs/ui/mobile/mobile-vault-dark-after.png`) while the reveal-by-Ether, scroll and charge texts are hidden by knob (`docs/vault.md` "Gold is the one way to reveal"; about 19 such strings in `web/vault.json:188`, `:329`, `:347-362`, `:436-448`). Menu says "Ethereal Vault", band says "Vault". Layout note: both branches of a condition print the same sentence (`web/shell.js:934`) | fix the sentence; one name "Vault" in navigation with "Ethereal Vault" as the page title; Ether texts follow the currency decision (F7) | FIX now (Skills, names); DELETE the hidden reveal texts only after F7 | S / low |
| C7 | Placeholders shipped to users | Inventory: "Your rules and the starter preset show here once the loot rules tab is built" (`invui.go:117`) although the loot rules editor exists (`web/lootrulesui.js`, opened only from the `...` menu of `/watch`); "Stash from anywhere: locked (skips a game rule)" (`invui.go:136`); Points: a disabled button "Respec: coming soon" (`pointsui.go:48`) | the Inventory block becomes a "Loot rules" button that opens the existing editor; the stash line goes; the respec button is hidden until it works, the price line stays | DELETE / rewire | S / low |
| C8 | Browser dialogs | `window.confirm` in `web/points.js:142`, `:144`, `web/skills.js:110`, `web/app.js:398`, `:470`; `alert` in `web/update.js:41`. Locking a build, which the design says costs a respec to undo, is confirmed by an unstyled browser box | the one confirm sheet (C12) | UNIFY | S / low |
| C9 | Three button families | old `.btn` with `.primary .danger .ghost .small` (`web/app.css:31-37`; used by `web/watch.html`, `web/app.js`, `web/gear.js`, `web/alerts.js`, `returnscreen.go`, `policysheet.go`, `gate.go`), `d-btn` with data attributes (everything new), `cv-btn` (`invui.go:102-103`, `web/inventory-merc.js`) | `d-btn` only | UNIFY: `cv-btn` becomes `d-btn` with a canvas size; `.btn` dies with C2 and the two template pages | M / low |
| C10 | Four toast implementations and three severity vocabularies | `toast()` in `web/shell.js:94-108` (`data-sev` info, good, warn, alarm), in `web/app.js:208`, in `web/alerts.js:30`, and `vt-toast` in `web/vault.js:139-141` (`data-level` good, bad); chips use `data-level` ok, warn, bad and badges `data-tone` (`web/components.css:50-52`) | the shell toast, exported so the Vault and Characters pages can use it; one scale `info, good, warn, bad` for toast, note, chip and badge | UNIFY | M / low |
| C11 | Eight sheet implementations | appendix A3 | two surfaces only: **sheet** (bottom sheet on a phone, centred card on desktop; modal; one at a time through `web/overlay.js`, which already exists) and **drawer** (side panel for reading: log, help, quests) | UNIFY over time: new work uses the shared sheet; `vt-sheet`, `cs-sheet`, `lr-sheet`, `cv-pop`, the reset and support dialogs move one card each | L / medium |
| C12 | Four ways to confirm | browser box (C8); a second press on the same bar with Yes and Back (`web/vault.js:113-115`, `:394-397`); a modal sheet (Reset, Vault sell, Identify all); typing the hero's name (Characters delete); and act-then-Undo toast (`web/panels/orders/panel.js:59-63`) | the rule in the vocabulary below: three levels, chosen by what is lost | UNIFY | M / low |
| C13 | Disabled reasons as hover text | appendix A4 item 2 | a reason line under the control (`d-why`), always visible while the control is disabled | UNIFY | M / low |
| C14 | Empty, loading and error states are per page | one shared `d-empty` used 4 times in `web/`; own texts everywhere else: "This panel failed to load." (`web/shell.js:1054`), "could not read the build state" (`web/points.js:153`), "No answer from the bot" (Characters), "The Ethereal Vault page could not load its texts." (`web/vault.js:1200`); Points and Skills show an empty page while loading (U) | three shared states with one sentence pattern each (vocabulary below) | UNIFY | M / low |
| C15 | Old colour names next to the tokens | `web/app.css:5` aliases `--surface`, `--line`, `--bad` ... to the `--d-*` tokens and uses them 59 times; two literal colours `#3a3`, `#d90` (`web/app.css:188-189`) | nothing: goes away with C2 | DELETE with C2 | - |
| C16 | Orphan pages | Pacts and Hall of Fame (appendix rows 11, 12) | either a `More` entry and the shared frame, or remove the routes | decide (F6) | S / low |
| C17 | The Help button fires its event twice with a 900 ms timer | `web/shell.js:1025` | harmless; note only | keep | - |

**Keep as it is (looked at, found sound):** the token file and cascade layers; the panel contract and the scheduler; the cockpit plan and its phone tabs; the character canvas layout; the Vault page layout and its read-back rule ("every press is read back", `web/vault.js:7`); texts in JSON files for the Vault and Characters; the Orders "More" sheet with a filter; Undo toasts for orders; the typed-name delete.

### C-vocabulary: the one component set (proposal, 11 lines)

1. **Button** `d-btn`: variants `primary` (one per view), default, `ghost`, `danger`; sizes `sm` (36 px, dense desktop rows) and `md` (44 px, every phone control). No other button class.
2. **Chip** `d-chip`: a small fact or a filter. With `aria-pressed` it is a filter toggle. Never an action that changes the hero.
3. **Toggle**: a `d-btn` with `aria-pressed` and the state in its text ("Autopilot: on"). No checkbox styled as a switch.
4. **Tabs** `d-tabbar`: parts of one page (`role=tablist`, arrow keys). The global navigation is not tabs; it is links.
5. **Sheet**: the one modal surface. Title, body, a button row with the primary on the right, Escape and scrim close it, focus returns to the opener. Bottom sheet under 700 px.
6. **Drawer**: the one non-modal side surface, for reading (log, help, quests). Never holds a confirm.
7. **Toast**: what a press did, 8 s, optional one action (Undo). Levels `info, good, warn, bad`. Never the only place an error is shown: the control's reason line also says it.
8. **Note** `d-note`: a standing message inside a page (refusal, "the bot is not answering"), same four levels.
9. **Empty, loading, error**: `d-empty` with one sentence in the pattern "Nothing here yet: <what fills it>", a skeleton for loading (the shell already has `d-skel`), and for an error "Could not load <thing>. Try again" with a button.
10. **Confirm, three levels by what is lost**: (a) undoable: act, then a toast with Undo; (b) costs gold or cannot be undone cheaply (sell, pay, lock a build, reset): a sheet that names the thing, the price and what is lost, primary button carries the price; (c) destroys a hero or its saves: the sheet plus the typed name. No browser dialogs, no "press again" bars.
11. **Reason line** `d-why`: visible text under a disabled control. `title` is never the only carrier of a reason.

## D. Build pages

### D1. What a player sees today (V)

There is no build page. The picker is a list inside the "Automatic Allocate Skills" card of `/watch/points`, drawn by `renderPicker` (`web/points.js:112-139`) from `GET /api/hero/build`. For each build file of the hero's class it shows:

- a heading that is the **file's build name with underscores turned into spaces** (`cmd/d2bot/autoallocwire.go:71`: `Label: strings.ReplaceAll(name, "_", " ")`), followed by the tier in brackets, for example "paladin hammerdin zeal endgame90 (supported)" or "fallback-lightning-v3 (experimental)";
- a red line of skills the hero "cannot use yet and will hold the point for you";
- a count line "N more skills are not proven in the fighter yet";
- a fold "First 30 levels" with one text line per level ("Level 7: Might x1, +5 strength");
- a button "Choose and lock", confirmed by a browser dialog.

It does not show: what the build is, how it plays, what it is good or bad at, what gear it wants, or which rows belong together. The `BuildInfo.Note` field exists (`internal/points/autoalloc.go:27`) but the wire does not fill it and the page does not draw it.

The data behind it (V, `internal/plandata/build-*.json`, 28 files): `build`, `class`, `rows` (per level: stat points and skills), and unevenly `tier`, `combat`, `selector`, `hotkeys`, `main_attack`, `respec`, `mana_heavy`, `ladder_only`, and a `note` or `tags` string of 450 to 1,400 characters that is **provenance for developers** ("V end state Maxroll ... U adaptation: no respec, so clvl 2-17 spend only ..."), not player text. Combat lines are in `internal/combat/classdata/profiles.json` (per class and build: attack bands, buffs, key layout). No file holds a summary, pros, cons or ratings.

**The core readability problem is not the amount of text per row; it is that rows are tables, not builds.** A Paladin gets six rows (`paladin-normal30`, `paladin-normal30-v2`, `paladin-zealot-lite`, `paladin_hammerdin_endgame90`, `paladin_hammerdin_guide`, `paladin_hammerdin_zeal_endgame90`) and all six put their points into Blessed Hammer (V: summed from the rows). They are one build, the Hammerdin, with two ways to level and three endgame tables. The Sorceress gets lightning three times with version suffixes (`fallback-lightning`, `-v2`, `-v3`).

### D2. Proposal: one card per build, tables behind it

A **build** is what the player chooses (Hammerdin). A **table** is a file the allocator follows (normal30, endgame90, guide). A new catalogue groups tables into builds and carries the player text. Each build card has, top to bottom:

1. **Identity line**: name, class, one line ("Spins magic hammers around himself; almost nothing resists them").
2. **Summary**, two or three lines of plain words: how it fights, when it gets strong, what it needs from you.
3. **Pros and cons**, three each, plain words, no numbers that are not measured.
4. **Recommended for**: seven ratings, one to five stars each, always in the same order (D3).
5. **Status chip**: Supported / Experimental (the existing tier), and "the hero cannot use: X" when the capability table says so.
6. **Details** (closed by default): the level table, attribute plan, main attack by level, hotkeys, gear rules (`ladder_only`, strength and dexterity targets), table variants, and the provenance note for the curious.
7. **Choose and lock**, confirmed by a sheet (C12 level b) that shows the respec price.

### D3. The seven ratings

| # | Rating | Means | Can it be measured here? |
|---|---|---|---|
| 1 | Leveling | how smoothly it gets from level 1 through Normal | YES: time to clvl 30 and deaths on the way, from a fresh-start soak |
| 2 | Farming (items) | repeating runs for drops late in the game | PARTLY: runs per hour of a farm order; needs endgame tables to be playable, which most are not yet |
| 3 | Farming (XP) | experience per hour in dense areas | YES: XP per hour from a soak in a named area |
| 4 | Bossing | act bosses and single hard targets | YES: boss kill time and deaths in a boss scenario |
| 5 | Defense | how hard it is to kill | YES: survival margin (lowest life per fight), potions used, deaths per hour |
| 6 | Clear speed | how fast it empties an area | YES: clear time of a fixed area in a Proving Ground scenario |
| 7 | Hands-off | how little it needs from you: no held points, few "needs you" rows, no stalls on mana, corpses or a dead mercenary | PARTLY: needs-you rows, held points and stall minutes per hour from the soak log; "how fiddly it feels" stays opinion |

Seven is the cap. Not added: "gear dependence" (a con line says it better), "fun" (opinion with no data), "party value" (the product is single player).

**Honesty rule for stars.** Every rating carries a source: `exp` (opinion of a veteran, reviewed by the d2-sme agent), or `measured` (a number from a named scenario or soak, turned into stars by a threshold table that is also data). The page shows the difference without a legend hunt: opinion stars are outline-only with the word "estimate"; measured stars are filled and the detail line names the number ("measured: 0.4 deaths per hour, soak of 2026-12-01"). A build never shows a measured star from a CHEATED or impossible character (test-character provenance rule). **Today every star would be `exp`**: no build has the scenarios, so the first release ships estimates and says so.

### D4. Where it lives and how it is checked

- **File**: `internal/plandata/build-catalog.json` (new, one owner per wave because it is a shared data file). One entry per build: `id`, `class`, `name`, `identity`, `summary`, `pros[3]`, `cons[3]`, `ratings{leveling, farmItems, farmXP, bossing, defense, clear, handsOff}` each `{stars, source, evidence, why}`, `tables[]` (file build names with a `stage`: leveling, endgame, variant) and `default_table`. Recommended over a block inside each `build-*.json` because six Paladin tables would repeat the same text and drift (alternative in F5).
- **Contract test** `TestBuildCatalog` in `internal/plandata`: every `build-*.json` belongs to exactly one catalogue entry or is listed under `retired`; every entry has exactly three pros and three cons of at most 90 characters, a summary of at most 280, all seven ratings with stars 1 to 5; `source: measured` requires a non-empty `evidence` that names a scenario id; classes match; no skill name appears that is not in the class's skill table. Per the build-flexibility rule the test runs over every class in the file, not only the Sorceress.
- **API**: `GET /api/hero/build` gains `cards[]` (additive; `builds[]` stays so an older page keeps working). `label` comes from the catalogue, never from the file name.
- **Authoring and review**: an agent drafts an entry with every star `exp` and a one-line `why`; the d2-sme agent reviews the text and stars; the user corrects. A star becomes `measured` only by a card that names the scenario and the threshold row.
- **Threshold table** (data, same file, section `scales`): for each measurable rating, four cut points that map the number to 1 to 5 stars. The cut points are a user decision once real numbers exist (F9).

### D5. First-draft ratings (ALL UNVERIFIED / EXP: for the user to correct; no number here was measured)

Order: Leveling, Farming (items), Farming (XP), Bossing, Defense, Clear speed, Hands-off.

| Build | Lv | It | XP | Bo | De | Cl | HO | One-line reason (EXP) |
|---|---|---|---|---|---|---|---|---|
| Paladin: Hammerdin | 3 | 5 | 5 | 4 | 5 | 4 | 3 | Slow until Blessed Hammer at 18 and its synergies; magic damage that almost nothing resists; Holy Shield and block make it the sturdiest caster; hammers miss in corridors and drink mana. |
| Paladin: Hammerdin, Zeal start (leveling variant) | 4 | 2 | 2 | 3 | 4 | 3 | 4 | Melee with Might or Holy Fire is steady early and needs little mana; it is a bridge to hammers, weak if kept late. |
| Sorceress: Blizzard | 4 | 5 | 4 | 4 | 2 | 4 | 4 | Ground-target cold that kills from range and slows; Teleport; thin life, and cold-immune packs in Hell must be skipped. |
| Sorceress: Fire Ball | 5 | 3 | 3 | 4 | 2 | 4 | 4 | Strongest early caster and simple to play; many fire immunes late; marked experimental in the fighter today (V `build-sorceress_fire_normal30.json` `tier`). |
| Sorceress: Lightning | 3 | 4 | 4 | 4 | 2 | 5 | 2 | Highest clear speed when it works, with very uneven damage and the largest mana bill; the fighter cannot use its attacks yet (V `docs/design/autoalloc-and-currencies-2026-11-21.md` section 1). |
| Necromancer: Summoner | 3 | 3 | 3 | 2 | 5 | 2 | 4 | The army takes the hits, so the hero rarely dies; slow, and act bosses in the clvl 25 to 35 range take minutes (the SME answers of 2026-11-22 call this the main risk). |
| Necromancer: Bone and Corpse Explosion (leveling table) | 3 | 2 | 3 | 3 | 2 | 3 | 2 | Teeth and Bone Spear from range, Corpse Explosion needs a first corpse and aim; no tank. |
| Necromancer: Poison Nova (being added) | 2 | 4 | 4 | 3 | 2 | 4 | 3 | No aim, flat mana cost, strong on packs once Nova and its synergies are in, which is about clvl 36 to 40; poison immunes and a fragile hero (both from the SME answers of 2026-11-22). |

Draft pros and cons for one build, as the pattern (EXP): **Hammerdin** pros: "Almost no monster resists its damage", "Very hard to kill with a shield", "Scales to the end of the game with cheap gear". Cons: "Weak and slow before level 18", "Hammers miss in narrow corridors", "Uses a lot of mana".

### D6. Wireframes and accessibility

Phone, build list (the Skills area, sub-tab Build):

```
+--------------------------------------+
| d.idle  Skills            Pala  L24  |
| [ Tree ]  [ Points ]  [*Build*]      |
+--------------------------------------+
| Following: none. Points are banked.  |
+--------------------------------------+
| HAMMERDIN                 Supported  |
| Spins magic hammers; almost nothing  |
| resists them.                        |
| Leveling ***..  Defense  *****       |
| Bossing  ****.  Clear    ****.       |
| estimate                 [ Open > ]  |
+--------------------------------------+
| ZEALOT START           Supported     |
| ...                                  |
+--------------------------------------+
|  Live   Gear  *Skills*  Vault  More  |
+--------------------------------------+
```

The card shows four of the seven ratings (the two best, Defense always, and the worst), the detail shows all seven. Phone, build detail:

```
+--------------------------------------+
| < Builds        HAMMERDIN  Supported |
+--------------------------------------+
| Spins magic hammers around himself.  |
| Slow until level 18, then one of the |
| safest and strongest builds.         |
|                                      |
| GOOD                 NOT SO GOOD     |
| + nothing resists    - weak to 18    |
| + very hard to kill  - corridors     |
| + cheap gear         - mana          |
|                                      |
| RECOMMENDED FOR            estimate  |
| Leveling        ***..  3 of 5        |
| Farming (items) *****  5 of 5        |
| Farming (XP)    *****  5 of 5        |
| Bossing         ****.  4 of 5        |
| Defense         *****  5 of 5        |
| Clear speed     ****.  4 of 5        |
| Hands-off       ***..  3 of 5        |
|                                      |
| > Details: levels, stats, gear, keys |
| [ Choose and lock  (respec 6,500 g) ]|
+--------------------------------------+
```

Desktop: the list is a grid of cards three across with all seven ratings on the card; opening a card puts the detail in a docked column on the right (the pattern the Vault item card already uses), so comparing two builds is two clicks without leaving the list.

Accessibility: a rating is text first. Markup is a `dl` row "Defense" / "5 of 5 (estimate)", the star glyphs are `aria-hidden`, and the number is visible, not only in a label (colour-blind and low-vision users, and it survives a copy and paste). Stars use the text colour, not red and green. Pros and cons are two lists with headings, the plus and minus signs are decoration. The fold is a native `details`. Every control is 44 px on a phone.

## E. Prioritised plan

Order by what unblocks most with least risk. Days are agent-days. Tests named are new unless they exist.

| # | Card | Owner paths | Effort | Tests | Needs |
|---|---|---|---|---|---|
| E1 | Quick links, one hour each: Inventory gets Vault and Points (band and phone list); Vault gets Gear, Skills and Live; Skills gets Points; every "Back" uses `/watch?live=1` | `invui.go`, `skillsui.go`, `pointsui.go`, `web/vault.js`, `web/vault.json`, `web/characters.js` | 0.5 | `TestPageLinks` (Go: every page template contains the required hrefs); node test of the Vault header | nothing |
| E2 | Stale texts and placeholders (C6 Skills sentence, one name "Vault", C7) | `skillsui.go`, `invui.go`, `pointsui.go`, `web/shell.js` | 0.5 | golden DOM texts under `testdata/` | nothing |
| E3 | Delete the dead: `invPage` and `?view=classic`; layout files `mockup`, `live`, `command` | `invui.go`, `invapi.go`, `web/layouts/`, `http_view.go`, `tools/uishots/` | 0.5 | `compat_test.go` must stay green or F2 decides | F2 for the layout files only |
| E4 | The shared navigation: `web/nav.json`, `web/nav.js`, bottom bar under 700 px, band links above; all pages load it; `More` sheet replaces the `...` menu and the Inventory phone list | `web/nav.*`, `web/shell.js`, `web/shell.css`, the page templates, `web/vault.js`, `web/characters.js` | 3 | `TestNavEveryPage` (each page lists the five destinations with one `aria-current`); uishots fit pages `nav-<page>` at 390x844 and 1366x768; a node test of `nav.js` | F1, F4 |
| E5 | Build catalogue data and contract test, first-draft texts and stars for Paladin, Sorceress, Necromancer | `internal/plandata/build-catalog.json`, `internal/plandata/*_test.go` | 1.5 | `TestBuildCatalog` | F5; d2-sme review |
| E6 | Build cards in the API and the page (list, detail, stars with text, confirm sheet) | `internal/points/autoalloc.go`, `cmd/d2bot/autoallocwire.go`, `web/points.js` or a new `web/builds.js`, CSS | 3 | Go test of `cards[]` for two classes; DOM golden of a card; uishots fit `builds-list`, `builds-detail` at both sizes | E5 |
| E7 | One confirm sheet and no browser dialogs (C8, C12) | `web/sheet.js` (new, shared), `web/points.js`, `web/skills.js`, `web/update.js` | 1.5 | node test of the sheet (focus in, Escape, focus back); grep test: no `window.confirm` in `web/` | nothing |
| E8 | One toast and one severity scale (C10) | `web/shell.js`, `web/vault.js`, `web/components.css` | 1 | uilint rule: `data-sev` values from the one list | nothing |
| E9 | Reason lines instead of hover text (C13), Vault and Inventory first | `web/vault.js`, `web/inventory-canvas.js`, `web/panels/quest-board/` | 2 | DOM test: a disabled control has a visible sibling reason | nothing |
| E10 | Move the five pre-shell-only features into `More` > Settings and the log drawer (C2 first half) | `web/panels/log/`, `logtab.go`, a new settings sheet | 3 | Go tests of the new log tabs; DOM goldens | F3 |
| E11 | Delete the pre-shell page, the old presets and the 11 panels (C2 second half, C3, C4, C9, C15) | `web/watch.html`, `web/app.js`, `web/app.css`, `web/unlockui.js`, `web/alerts.js`, `web/layouts/`, `web/panels/` | 2 | the existing suites; `compat_test.go` changed as F2 allows | F2, E10 |
| E12 | Return and Policy pages on the shared frame and components; decide Pacts and Hall | `returnscreen.go`, `policysheet.go`, `pactui.go`, `hall.go` | 2 | DOM goldens | F6 |
| E13 | First measured stars: thresholds as data, one scenario per measurable rating for one build of two classes | `internal/plandata/build-catalog.json`, scenario files | 3 plus lab time | the scenarios themselves | F9; a lab slot |
| E14 | Remaining sheets onto the shared sheet (C11) and shared empty, loading, error states (C14) | per page | 3 | uishots fit per page | nothing |

Quick wins first: E1, E2, E3 are an afternoon and answer the user's example. E4 is the piece that makes the rest stay fixed: after it, a new page cannot forget a link. E5 and E6 are independent of E4 and can run beside it (different files). E10 and E11 are the big deletion and wait for two decisions.

## F. Open questions (recommendation first; the user decides)

1. **Global navigation: five destinations as a bottom bar on a phone and band links on desktop?** Recommend yes, as in B3. Alternative: keep band links on a phone and only add the missing ones (cheaper, but the band is full at 390 px).
2. **May the compatibility promise of 2026-11-09 be relaxed so old presets and panels can be deleted?** (`docs/ui-design-system.md` section 0, rule 4: "removed panels leave a stub", rule 2: "old files keep loading".) Recommend: keep the promise for the layout and panel FILE FORMATS and for a hero's own saved layout, and allow the shipped presets and their panels to be removed with a stub that says "this panel moved" for one release. Without this, C2 to C4 stay for ever.
3. **Cain's notes, the Atlas, "Still locked" and Spend: move or drop?** Recommend: move rule revoking and alert settings (they are controls); move Cain's notes and the Atlas as two more tabs of the log drawer only if the user reads them; drop Spend and "Still locked" unless the Tally loop is still wanted (see F7).
4. **Merge Points into the Skills area as sub-tabs Tree, Points, Build?** Recommend yes: one navigation entry, and the build card needs room that a card inside Points does not have. The addresses `/watch/points` and `/watch/skills` keep working.
5. **Build text in one catalogue file or inside each table file?** Recommend the catalogue (D4). The brief's wording was "in the build tables"; the catalogue is in the same package and avoids six copies per build.
6. **Pacts and Hall of Fame pages: link them or remove them?** Recommend: Hall of Fame gets a `More` entry and the shared frame (it is a reward page); Pacts stays unlinked until the pact loop is decided, and its route returns to the developer set.
7. **Ether in the Vault header.** The currency design of 2026-11-21 proposes gold for capacity and "Renown" for the score; it is a proposal. Recommend: do not touch Ether texts in a UI clean-up; delete the hidden reveal-by-Ether and scroll texts only when that proposal is decided.
8. **The name of the default layout.** Recommend: stop showing "Cockpit" and "Classic" as peers. If Classic survives F2 it is offered as "Old layout", and the old page shows a standing one-line bar "You are on the old layout: Switch back".
9. **Star thresholds for measured ratings** (what deaths per hour is five stars). Recommend: decide after the first soak numbers exist; until then all stars are estimates and labelled so.
10. **Which four ratings show on a small card?** Recommend: Defense always (the user asked for it), the two highest and the lowest. Alternative: a fixed four (Leveling, Farming, Bossing, Defense).
11. **Is `Stream` (the OBS overlay strip) still used?** If not, it goes with the other presets.

## G. Dead ends and limits of this review

- **No browser, no new screenshots.** The screenshot suite runs on the builder in its long lane; this review did not start it. Visual claims rest on five existing pictures in `docs/ui/` and on reading CSS. Spacing, colour contrast and real touch behaviour were not assessed; "colours and spacing" in the brief is covered only as far as the old colour aliases (C15) and the severity scales (C10).
- **Not opened**: the dashboard, the tile viewers, the desktop app shells and the setup page beyond their titles. They are listed (appendix rows 14, 15) and deliberately left out of the unification.
- **The link table (B1) is from static text.** A link built from server data at run time (onboarding steps carry `href` values from `data/onboarding/`) is not in it; the onboarding list may link to pages the table marks unreachable (U).
- **"Used only by presets" (appendix A2)** is from the shipped files. A hero's own saved layout can name any panel; nothing here can count those.
- **Line counts and "20 strings" style numbers** are from one read of the files at the starting commit and will drift; the file and line are the reference, not the count.
- **Build ratings are opinion.** The table in D5 is EXP from general Diablo II knowledge plus two repo documents (the class ranking and the Necromancer SME answers). No scenario, soak or lab run backs any star. The Poison Nova row describes a build that is not in `internal/plandata` yet.
- **Dead end: rating from existing data.** I looked for a source to derive stars today. `docs/classes/ranking.md` scores classes for bot suitability on six other criteria, per class not per build, and is itself tagged judgement; the capability table gives supported or experimental, not quality. Neither can fill the stars.
- **Not checked**: whether `primary-blizzard` and `fallback-lightning` (files without a `class` field) appear in the picker for a Sorceress; the picker filters by class (`internal/points/autoalloc.go:129-140`) and the wire may fill the class another way (U).
- **Effort figures** in section E are estimates from file sizes, not from a trial.
