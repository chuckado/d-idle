# UI references for the watch screen and the character screen

> status: research input dated 2026-11-14, not a design. Gathered from the web in one session: reference screenshots were downloaded to a scratch area (never into the repo) and looked at one by one; documentation pages were fetched as text (the fetch tool returns a summary, not the raw page). Where a site blocked fetching it was skipped. Target is classic 1.14d; D2R-only facts are flagged D2R.

Tags: **V** = I fetched the page or looked at the image (named in the line), **S** = secondary description or search snippet only, **U** = unverified or from memory. Pixel figures are measured by eye on 1920x1080 screenshots and are rough (plus or minus 5 percent). Screenshot source for most images: Interface In Game (interfaceingame.com, per-screen pages under /screenshots/<game>-<screen>/), Steam store screenshots, Wikimedia Commons.

## Coverage and honesty table

| Reference | Fetched | Image looked at |
|---|---|---|
| Diablo II Resurrected (technical alpha) inventory + stash | interfaceingame.com D2R pages | yes (2) |
| Diablo III hero details, loot HUD | interfaceingame.com D3 pages | yes (2) |
| Diablo IV beta weapon tooltip | interfaceingame.com D4 pages | yes |
| Path of Exile 1 inventory | interfaceingame.com | yes |
| Path of Exile 2 in-game HUD | Steam store screenshot | yes (HUD only, no inventory) |
| Last Epoch | search results + one patch-notes page | images checked, no UI visible |
| Grim Dawn inventory | interfaceingame.com | yes |
| Warframe equipment | interfaceingame.com | yes |
| Destiny 2 character | interfaceingame.com | yes |
| Moonlighter inventory (extra ARPG-lite) | interfaceingame.com | yes |
| Genshin Impact mobile (hub screens) | interfaceingame.com | yes (2) |
| Melvor Idle, Cookie Clicker, Antimatter Dimensions, Idle Champions, Kittens Game | Steam store screenshots | yes (5) |
| Grafana | Wikimedia Commons screenshot + official best-practices page | yes |
| Bloomberg terminal | Wikimedia Commons photo | yes (angled photo) |
| Home Assistant dashboards | official docs (sections, panel) + release blog | no |
| Linear | official redesign blog + display-options docs | no |
| Datadog | official executive-dashboard blog; one best-practices page 404 | no |
| Stream overlays | search snippets only (the Elgato page 404) | no |
| Apex Legends quick swap overlay (HUD) | interfaceingame.com | yes |
| Factorio, RimWorld, Dwarf Fortress, Frostpunk | Commons / Steam / interfaceingame | yes (4) |
| Oxygen Not Included | Steam | yes, but the shot shows no HUD (not analysable) |

Not verified at all (no usable source): Torchlight Infinite UI (the one image I looked at was a character render), Titan Quest 2 UI (gameplay shot only), Vercel dashboard (docs URL 404), Game UI Database (blocked, 403), Fandom/wiki pages (blocked), Diablo wiki (TLS error).

---

## A. ARPG inventory and character screens

### A1. Diablo II Resurrected inventory and stash (D2R; layout is the 1.14d layout re-skinned)
URLs: interfaceingame.com/screenshots/diablo-ii-resurrected-technical-alpha-inventory/ and .../-stash/. Image: yes (2) [V].
- Inventory is a right-docked panel about 570x750 px (30 percent of width, 70 percent of height) over a live map; the game stays visible on the left 60 percent [V image]. 
- Doll uses a fixed 3-column slot map: weapon slot left, head+body centre, weapon slot right, then gloves / ring / belt / ring / boots on one row; slot sizes are multiples of the 49 px cell [V image].
- Pack grid is 10 columns by 4 rows directly under the doll, gold in a bar under the pack, no stats on this screen [V image].
- Weapon-swap is two small tabs (I, II) above each hand slot, not a separate screen [V image]. The 1.14d original has the same I/II concept via the W key [U].
- Stash screen mirrors the layout: stash panel left, inventory right, character stays centre-bottom; stash 10x10 with Personal/Shared tabs is D2R-only; classic 1.14d is 6 wide x 8 tall (our brief) [V image for D2R; classic size from the brief].
- Item tooltip is not in these shots [V image]. No mercenary panel visible; in classic 1.14d the mercenary has its own equipment window opened from the merc portrait [U].
- **Take:** the fixed slot map (players have muscle memory for it) and the I/II swap tabs on the hand slots; inventory beside the stash with both visible at once.
- **Reject:** heavy ornate frame (about 15 percent of each panel is border art), and three separate windows to compare doll, stash and merc.

### A2. Diablo III hero details and loot HUD
URLs: interfaceingame.com/screenshots/diablo-iii-hero-details/ , .../diablo-iii-loot/. Image: yes (2) [V].
- Hero details is a modal on a modal: a doll block about 280 px wide at left, and a right block with Kanai's cube powers (3 big tiles) and skills (6+4 icons) [V image]. Almost no numbers: it is a "build summary", not a stat sheet.
- Doll slot colour coding (green frames with a star glyph for set or legendary) shows item class at a glance without text [V image].
- Loot HUD: action bar plus orbs occupy the bottom 13 percent; party frames top-left in a 100 px rail; minimap and objectives top-right; chat bottom-left; the game view is the hero at roughly 80 percent [V image].
- **Take:** colour-coded slot frame as a status signal; one glyph (star) as an extra flag.
- **Reject:** stacked modals; large decorative headers (the "Profile 3950" banner eats 12 percent of height).

### A3. Diablo IV beta weapon tooltip
URL: interfaceingame.com/screenshots/diablo-iv-beta-weapon/. Image: yes [V]. (This capture is a localised build.)
- Tooltip is a tall card (380x790 px, 41 percent of height) anchored to the left edge of the character panel and overlapping the game view, not the panel [V image].
- Header has item name and a power number; the key row is "81 damage per second (+34)" with the delta in green next to the base value, i.e. comparison is inline [V image].
- Affixes are bullet rows with a diamond marker; the legendary effect is a coloured paragraph and the card scrolls ("scroll down" prompt) when long [V image].
- Context actions are a labelled list directly under the tooltip with key chips: Equip, Shift+click link, Shift compare, mark as junk [V image].
- Character panel is right-docked at about 40 percent width: tabs (Character / Expertise / Skills) on top, doll with a large 3D model, a tab row (Consumables / Quest / Aspects) between doll and pack, pack 7-wide, currency strip under the pack [V image].
- **Take:** delta shown beside the number it changes; action list under the tooltip with visible hotkeys; tab row between doll and pack.
- **Reject:** the tall scrolling tooltip (a scrolling tooltip cannot be read by a mouse hover without stealing focus); a 3D model as the doll centrepiece (about 25 percent of the panel is art).

### A4. Path of Exile 1 inventory
URL: interfaceingame.com/screenshots/path-of-exile-inventory/. Image: yes [V].
- Right-docked panel, 640 px wide (33 percent) and about 65 percent of height; top strip has tabs (Inventory / Cosmetics); doll takes the top 45 percent, flask belt in a row below it, pack 12 columns x 5 rows takes the lower 25 percent [V image].
- Item tooltip is centred-top of the screen, 410 px wide, with a header ribbon (two lines: name, base type), then base stats (white), then a single requirement row, then affixes in blue centred, one per line [V image].
- Requirements are one row: "Requires Level 53 (gem), 131 Int", unmet requirements are red in the live game [U for the colour; the row itself V].
- Stack counts are digits in the cell corner; the grid has no empty-space guides, so empty cells are the same dark as the border [V image].
- No built-in mercenary; companions are not on this screen [V image].
- **Take:** requirement as one compact row; affixes one per line with colour = type; stack numbers inside cells.
- **Reject:** tooltip positioned far from the cursor and the panel (it is detached at the top centre in this capture); flask row that is not part of the pack grid.

### A5. Path of Exile 2 in-game HUD
Source: Steam store screenshots. Image: yes (HUD only) [V]. The wider PoE2 inventory redesign is discussed in search results only: stash tab affinities (items go to a tab on ctrl-click), new fragment tab, complaints that the stash looks dated [S: web search, several news pages].
- HUD is minimal: two orbs bottom corners about 220 px, a skill bar of 5 skills on the right and a flask pair on the left, a boss health bar 770x40 at the top centre with debuff icons beneath [V image]. Total HUD about 7 percent of the screen.
- **Take:** the boss bar pattern (one wide bar at the top centre, state icons under it) for a single "what is happening now" line.
- **Reject:** nothing to reject from a screen that is mostly view; not analysable as an inventory.

### A6. Last Epoch
URLs checked: search results and a patch-notes summary page (maxroll.gg news). Image: viewed 2 store screenshots, both gameplay with no UI [V that no UI is visible].
- Inventory is split into separate tabs for crafting materials, dungeon keys and woven echoes; keys got their own inventory space; forum feedback asks for a character preview and stat tooltips on the inventory screen [S: maxroll patch notes page V text, forum snippets S].
- Idols sit in a separate section beside the gear screen and do not use pack space [S: guide snippets].
- Unmet level requirement is marked with an X on the item in the pack or stash [S: developer notes snippet].
- **Take:** an unmet-requirement mark visible on the grid cell itself, not only in the tooltip.
- **Reject:** separate tabs for each material type when the pack is only 10x4.

### A7. Grim Dawn inventory
URL: interfaceingame.com/screenshots/grim-dawn-inventory/. Image: yes [V].
- One centred window, 495x530 px (26 percent of width, 49 percent of height), split into left (doll with 14 slots around the figure) and right (attributes, health/energy, combat stats, a 2x5 row of resistance icons with percentages) [V image].
- Pack grid sits under both halves in the left 60 percent, and a currency strip on its top edge; the right-bottom pane is a large empty textured box (dead space, a drawn-but-unused panel) [V image].
- Hover compare: two tooltips side by side, left "Currently Equipped" and right "Item in Inventory", each with armour value, affix lines and two requirement lines ("Required Physique: 33") at the bottom [V image].
- Item frames are tinted by rarity (yellow frame = rare/epic) with the item art inside; equipped slot dark frames for empty ones [V image].
- **Take:** automatic side-by-side compare without a modifier key; stat column in the same frame as the doll so stats update while you look at gear; resist icons as a compact 2x5 block.
- **Reject:** the empty right-bottom pane; tooltips at 11 px type that are unreadable at 1920 and worse at 1366.

### A8. Warframe equipment
URL: interfaceingame.com/screenshots/warframe-equipment/. Image: yes [V].
- Left rail 410 px (21 percent) lists loadout slots as big rows (name + rank), centre is a 3D model hero about 40 percent, right column is an 8-line stat table (armor, energy, health, shield, sprint speed, duration, efficiency, range, strength) in 14 px type [V image].
- Actions attach to the selected row: three adjacent text buttons (Equip, Appearance, Abilities) appear beside it, bottom bar has Select mode / Loadout options / Exit [V image].
- Currency in the top-right corner as two icon+number pairs [V image].
- **Take:** selected-row-owns-its-actions (buttons appear beside the selection, not in a global toolbar); stat table as plain two-column rows.
- **Reject:** the 3D model hero (not useful for a bot's paper doll) and row heights of 70 px (too tall for 30-item lists).

### A9. Destiny 2 character
URL: interfaceingame.com/screenshots/destiny-2-character/. Image: yes [V].
- Top bar 90 px high (nav tabs plus name, level, xp bar under it); model centre; five weapon/ghost tiles in a left column and five armour tiles in a right column, tiles 96 px on a 126 px pitch; one hero stat (Power, 60 px digits) beside the right column [V image].
- Footer bar 60 px with key hints (Toggle Subscreen, Gear Stats, Dismiss) [V image].
- About 35 percent of the frame is empty gradient on the left and right of the columns (dead space) [V image].
- **Take:** one hero number (Power) as the single summary stat of the screen; key-hint footer.
- **Reject:** symmetrical columns with empty gutters; the screen has room for about 10 numbers and shows 1.

### A10. Moonlighter inventory (small-pack reference)
URL: interfaceingame.com/screenshots/moonlighter-inventory/. Image: yes [V].
- A two-page book: left page is the pack grid (5 columns x 4 rows of 105 px cells, plus a quick row above), right page is the paper doll with a 4-line stat tray [V image].
- Tabs along the top edge (grid, star, backpack, map) [V image]; footer bar lists controller hints [V image]; HP and gold stay visible top-left while the screen is open [V image].
- **Take:** vital signs (HP, gold) stay on screen while the inventory is open; pack and doll as the two halves of one frame.
- **Reject:** 105 px cells (4 rows fill 60 percent of the height).

---

## B. Idle and incremental dashboards

### B1. Melvor Idle
Source: Steam store screenshot (Atlas of Discovery build). Image: yes [V]. Text: a wiki Settings page (403) and the beginners guide (403) were blocked; a search result says the sidebar can be collapsed to icons, hidden groups via an eye toggle, and a combat minibar can show on non-combat pages [S].
- Left sidebar 240 px (12.5 percent) with grouped rows: each row is 40 px tall with icon, skill name and "level / 99" right-aligned; groups (Combat, Passive, Non-combat) have eye toggles [V image, toggles S].
- Top header strip 65 px green with page name and save/cloud controls, then a thin stat strip (skill level, skill xp, current tool) 45 px [V image].
- Main area is a grid of cards (4 columns x 2 rows, 385x260 px), each with XP/time, art, progress bar and mastery count; locked cards are shown dimmed with a red level chip, they still take a full cell [V image].
- Progress is a big green bar across the top of the main area plus an inline bar per card [V image].
- **Take:** sidebar rows with a trailing number (level) as the density device; locked items shown with the unlock condition instead of hidden.
- **Reject:** 260 px cards that carry 3 numbers each (about 70 percent of each card is art or empty).

### B2. Cookie Clicker
Source: Steam store screenshot (early state). Image: yes [V].
- Three columns: left 580 px (30 percent) holds the hero (big cookie plus name and counters), centre 940 px (49 percent) is a news ticker at the top and a mostly empty content area, right 360 px store list [V image].
- Store list rows are 80 px tall; a Buy/Sell toggle with 1/10/100 above; hover tooltip appears to the left of the row, overlapping the centre column, 460x110 px [V image].
- In this early state the centre column is about 85 percent empty: an unfilled state is blank, not designed [V image].
- **Take:** the buy-quantity toggle (1/10/100) as a mode switch above a list; tooltip anchored beside the hovered row.
- **Reject:** a centre column that is empty until content arrives (this is exactly the dead-space complaint).

### B3. Antimatter Dimensions
Source: Steam store screenshot (late game). Image: yes [V].
- Left tab rail 108 px (7 percent), nine tabs, each row 38 px, colour-coded by prestige layer (Infinity orange, Eternity purple, Reality green) [V image].
- A currency strip at the top (three groups of two lines), then the main body is a strict table: 8 rows x 47 px, columns name, multiplier, amount, rate, buy button right-aligned; alternate rows are striped [V image].
- Bulk controls on one line: "Until 10", "Max All (M)" with the hotkey in the label; a footer shows time since last save [V image].
- The whole screen is one table plus two action blocks, no hero image, and no scrolling at 1624x912 [V image].
- **Take:** strict table rows with striping, right-aligned primary action, hotkey in the button label; colour by category on the nav rail.
- **Reject:** exponent notation in running text is unreadable; keep numbers short and use units.

### B4. Idle Champions of the Forgotten Realms
Source: Steam store screenshot. Image: yes [V].
- The scene is the hero at 100 percent (full-bleed), HUD overlays occupy about 25 percent: top-left resource trio (gold, static dps, boss resource) on a dark gradient, top-centre zone path with 5 nodes and an auto-progress toggle, top-right quest card, bottom a 12-portrait roster strip 160 px high [V image].
- The selected champion's card expands in place in the roster strip (680 px wide) with name, level, the next upgrade cost and button hints, while the neighbours shrink to portraits [V image].
- Skill/ult icons float above the roster strip, with cooldown timers drawn on the icon [V image].
- **Take:** full-bleed live view with the HUD as translucent overlays; roster strip where the focused item expands in place; cooldown timers drawn on icons.
- **Reject:** dark-gradient overlays that cut into the scene at the top (about 12 percent of the view).

### B5. Kittens Game
Source: Steam store screenshot. Image: yes [V].
- Pure text and buttons in a dark theme, three columns: resources list left (a name, value, max and rate per row, 22 px high), action buttons in two columns centre (each 40 px, with "(count)"), log and tooltip right [V image].
- A text tab bar along the top (Bonfire, Deities, Science, ...) with a red "(!)" marker on tabs that need attention [V image].
- A thin green bar on the left edge of a button shows its progress or capacity; "+" and "-" mini buttons are attached to the button edge [V image].
- The page continues below the fold (the window is cut at the bottom): the layout does not fit one screen by design [V image].
- **Take:** an attention marker on the tab itself; mini-actions attached to the button (+/-/auto) instead of a modal.
- **Reject:** the scrolling column; tabs of 14 items.

---

## C. Dense monitoring and ops UIs

### C1. Grafana
URLs: Wikimedia Commons file "Grafana Dashboard (2017).jpg" (image yes [V]); grafana.com/docs/grafana/latest/visualizations/dashboards/build-dashboards/best-practices/ (V text).
- Layout: top bar 40 px, variable/filter row 60 px, then three panel rows; first row is five stat tiles 305 px wide x 410 px tall with a sparkline behind a 60 px big number; second row has four pie panels; third is a world map filling the full width [V image].
- Panels are laid out on a 24-unit grid with fixed row heights (resize handle at each corner) [V image: handles visible; the 24-unit number is U].
- The pie panels leave the lower 35 percent of each panel empty beside their legend table, an example of dead space inside a panel [V image].
- Docs advice: one row per service in data-flow order, group related signals side by side (rates left, latency right), keep each dashboard answering one question, use drill-down to detail dashboards, avoid stacked graphs, use colour only for good/bad states with thresholds [V text].
- **Take:** stat tile with a sparkline behind the number; filter row on top; drill-down instead of cramming.
- **Reject:** pie charts with legends (waste a quarter of the tile); panels that are not sized to their content.

### C2. Bloomberg terminal
URL: Wikimedia Commons file "Bloomberg Terminal.jpg". Image: yes (an angled photo of a monitor, so values are rough) [V].
- Every panel is a mono-type table with a two-line header: a command line on top (yellow-on-black), numbered function buttons in a red bar, tabs numbered 21 to 27 under the first block [V image].
- Colour is the only emphasis: amber labels, white or cyan numbers, red bars for headings, green and red for up and down; type about 9 to 11 px equivalent [V image].
- The lower half of this window is still empty black (a tab body not yet filled), so even this reference has dead zones when data is missing [V image].
- Action model is keyboard-first: numbered buttons "96) Output", "97) Settings", with number = hotkey [V image].
- **Take:** numbered actions so every control has a hotkey; label-colour/value-colour split (amber/white) for dense key-value pairs.
- **Reject:** the monochrome dense look for a casual viewer.

### C3. Home Assistant dashboards
URLs: home-assistant.io/dashboards/sections/ , /dashboards/panel/ , blog post "dashboard chapter 1" (2024-03-04). Image: no [V text].
- Sections view puts cards in sections on a grid with a maximum column count; a section can span several columns; an optional dense setting back-fills horizontal gaps at the cost of card order [V text].
- Blog: chosen over masonry on purpose: masonry packs tighter but is unpredictable; the grid accepts more space for predictability; tile cards have standard sizes; on small screens sections rearrange while columns inside a section stay constant [V text].
- Panel view: one card spanning the full width, no badges; recommended for maps and picture elements [V text].
- **Take:** panel view for the video (one hero card), sections as stable units on narrow screens.
- **Reject:** masonry (the reason this redesign exists).

### C4. Linear
URLs: linear.app/blog/how-we-redesigned-the-linear-ui , linear.app/docs/display-options. Image: no [V text].
- Views follow one structure: navigation, a header with filters and display options, optional side panel for metadata, main content (list, board, timeline, split, fullscreen) [V text].
- Density came from trimming sidebar, tabs and headers, aligning icons and labels on one grid, and dimming navigation chrome; the post does not give row heights [V text].
- Display options choose which properties show on each row or card; they hide data but do not filter [V text].
- **Take:** per-view toggles for which fields a row shows; chrome dimmer than content.
- **Reject:** none; Linear is a list product, only the density rules apply.

### C5. Datadog
URL: datadoghq.com/blog/datadog-executive-dashboards/ (V text); one docs best-practices URL returned 404.
- Most important summary at the top; sections as groups that can collapse; template variables to filter; avoid many small competing widgets (no number given) [V text].
- Third-party guidance suggests under about 20 widgets per dashboard; this is a rule of thumb, not an official limit [S].
- **Take:** top-line summary row, then groups.
- **Reject:** collapsible groups (anything collapsed on a no-scroll screen is hidden state).

---

## D. Stream overlays and game HUDs

### D1. Stream overlay layout guidance
Source: search results from several overlay guides (the Elgato page returned 404) [S].
- Common advice: design at 1920x1080, keep the centre clear, webcam small in a corner or edge, alerts and goal bars in corners, check game UI zones before placing widgets, thin borders and translucent panels, one zone map (avoid / sparing / free) [S].
- **Take:** zones as a design tool: mark the corners and edges where overlay is allowed.
- **Reject:** unverified specific pixel safe areas; no source gave them.

### D2. Apex Legends quick-swap overlay
URL: interfaceingame.com/screenshots/apex-legends-quick-swap/. Image: yes [V].
- A full-width translucent band (about 250 px high, 23 percent of the screen) sits across the middle of the view; left half names the held item and quantity, right half is a 4x2 icon grid with stack counts [V image].
- The band is dimmed behind and the world stays visible above and below [V image].
- **Take:** a transient contextual overlay for item actions that keeps the world visible.
- **Reject:** none.

### D3. Diablo III and PoE2 in-game HUDs
See A2 and A5: view hero at 80 to 93 percent, status at the corners [V image].

---

## E. Management sims

### E1. Factorio
Source: Wikimedia Commons "Factorio Space Age Gleba Screenshot.jpg". Image: yes [V]. The official wiki GUI page returned 404.
- The map fills the screen; HUD is 4 small islands: top-right research progress (190x70 px, a bar and 8 shortcut icons), bottom-left character toolbar, bottom-centre quickbar two rows of 10 slots (about 840 px), and a pair of alert chips (a red and an amber triangle with counts) above the quickbar [V image].
- Alerts are tiny icon+count chips, not text; they are positioned at the bottom centre where the eye returns to the action [V image].
- FPS/UPS readout in 12 px at the top centre [V image].
- **Take:** alerts as icon+count chips; HUD as 4 islands, never a bar around the scene.
- **Reject:** none for density; the tiny hit targets (24 px) do not work for touch.

### E2. RimWorld
Source: Steam store screenshot. Image: yes [V].
- Map is the hero, with a top colonist portrait bar (about 80 px, 10 portraits with a status strip below), a left resource list 160 px wide (icon + number per row, 24 px per row), a bottom main-tab bar 40 px (Architect, Work, Schedule, Assign, Animals, Wildlife, Research, Quests, World), and a bottom-right time controls cluster [V image].
- Alerts are a right-side stack of text lines on red fills: "Colonist needs rescue", "Fire!", "Major break risk x2" (red = urgent, plain = informative) [V image].
- Selection info is a bottom-left panel 430x200 with a tab row (Log, Gear, Social, Bio, Needs, Health) and a pop-over detail pane 630x430 above it showing an overview/operations tab pair and a long health table of 14 rows at 20 px [V image].
- **Take:** alert stack with red fill = needs the player; the left resource list as the density model (icon + number only); tab row on the selection panel.
- **Reject:** the large overlay pane over the map for detail (it hides 14 percent of the view).

### E3. Dwarf Fortress (Steam)
Source: Steam store screenshot. Image: yes [V].
- Top strip 60 px: the fort name, population count with 7 mood faces (counts under each), a Stocks button, a row of colour-coded resources (Food, Drink, Seeds, Meat, Fish, Plant, Other) with a value under each, a date block [V image].
- A left icon rail 40 px with 13 round icons, a bottom toolbar 40 px, a top-right minimap 240x270 [V image].
- One announcement line in a bar under the top strip; the view stays the hero at about 85 percent [V image].
- **Take:** one strip of coloured label over value pairs for health lights; a single announcement line.
- **Reject:** icon-only rail with no labels (not self-explaining for first-time users).

### E4. Oxygen Not Included
Source: Steam store screenshot. Image: yes, but it is a world cutaway without HUD [V that no HUD is shown]. Not analysable; listed for completeness.

### E5. Frostpunk technology tree
URL: interfaceingame.com/screenshots/frostpunk-technology-tree/ (listed as the tree). Image: yes [V].
- Top bar with three resource counts (icon + number), a title and a close button; four category tabs under the title; a left rail of five tiers with roman numerals and a check mark when complete [V image].
- Locked rows are dimmed and boxed together at the bottom, so progress is read as "bright = done" [V image].
- **Take:** completion as a visible state (check badge on done steps) and locked steps dimmed together.
- **Reject:** a node-graph for a quest-chapter list; a list is denser.

---

## F. Mobile game hub screens

### F1. Genshin Impact (mobile layout, inventory and character)
URLs: interfaceingame.com/screenshots/genshin-impact-mobile-weapons-2/ , .../genshin-impact-mobile-attributes/. Image: yes (2) [V].
- Inventory: left icon rail 100 px (7 category icons), item grid 5 columns of 183x190 tiles in the left 60 percent, detail card on the right 615 px (32 percent) with name band, type, main stat, level, effect text and an "Equipped: <name>" strip at the card foot; top bar shows currencies and a capacity counter "Weapons 11/2000"; bottom bar sort dropdown, sort direction and trash [V image].
- Character: left nav list of 6 text entries (Attributes, Weapons, Artifacts, ...), a model hero at the centre, stats on the right as 5 label/value rows with a "Details" button and a primary Level Up button bottom-right [V image].
- A "New" badge on tiles is the unread mark; item level is a bottom chip [V image].
- **Take:** detail card in a fixed right column for the selected tile (hover on desktop); owner strip "Equipped: ..." on the card (maps to "worn by hero or merc"); capacity counter in the header.
- **Reject:** 183 px tiles (at 1366 wide the pack would not fit).

---

## G. Synthesis

### G1. Rules for a no-scroll, zero-dead-space watch screen with a video as hero

1. The hero (video) takes about 50 to 60 percent of the viewport width and its natural 16:9 height; everything else lives in rails and strips around it. Idle Champions, Factorio, RimWorld and Dwarf Fortress give the live view 75 to 100 percent and overlay the rest; we cannot overlay (the game must stay readable), so we dock instead [V images].
2. One status strip above or below the hero carries the sentence (what and why), in one line of 16 to 18 px; Dwarf Fortress and PoE2's boss bar both use a single wide line at the top [V images].
3. Alerts that need the player are chips with a count, red fill = needs you, otherwise neutral (Factorio chips, RimWorld stack) [V images]. An empty alert area shows a calm "nothing needs you" line, not a blank box.
4. Health lights are a single row of coloured label-over-value pairs (Dwarf Fortress top strip) [V image]; 7 to 9 items fit in 600 px.
5. Progress (level, xp/h, chapter) is a tile row with big number plus tiny sparkline or bar (Grafana stat tile, Melvor header strip) [V images]; 3 to 4 tiles, 90 to 120 px high.
6. Navigation and lists use fixed row heights of 24 to 40 px with the key number right-aligned (Melvor sidebar 40 px, Antimatter 47 px, RimWorld resources 24 px) [V images]. Do not use cards taller than about 100 px for list items.
7. Orders is a list in a rail, not a card grid: one row per order, status glyph + short text + one action (Antimatter table, Kittens button list) [V images].
8. The map is one panel (Home Assistant panel view logic) and doubles as the target picker: click sets the destination; place it where the minimap sits in games (a corner of the hero, or a side rail) [V text for HA, V image for DF/D3 corner minimaps].
9. Use a grid with fixed rows and columns, not masonry (HA blog: predictability beats packing) [V text]. Panels may span rows or columns but share one gutter (8 px).
10. Every region has a designed empty state at its normal size (Cookie Clicker's blank centre and Bloomberg's empty tab body are the failure) [V images]. An empty region shows a one-line hint plus the next expected event.
11. Breakpoints are layouts, not zoom: reflow whole regions (HA: section order changes, columns in a section stay) rather than shrinking type [V text]. At 1366x768 drop secondary text before shrinking the hero below about 640 px wide.
12. Hotkeys are shown in labels ("Max All (M)", "1) Output") and the screen needs no scrolling at all; every control reachable without it [V images].
13. Chrome is dimmer than content: Linear's redesign dims navigation and uses elevation levels for panels [V text]. Borders 1 px, no ornamental frames (D2R, D3 frames consume 10 to 15 percent of a panel) [V images].
14. Anything collapsible or tabbed is hidden state; on a monitoring screen use tabs only for rarely needed secondary views, never for the five required items. Datadog suggests collapsible groups for detail; here the brief forbids it for the required set [V text for Datadog, brief for the constraint].

### G2. Rules for a one-canvas character screen

1. One frame, fixed slot map: weapon left, head and body centre, weapon right, boots/gloves/rings/belt around (D2R, PoE1, Grim Dawn) [V images]. Hero and mercenary share the frame as two docs side by side, or hero doc with merc as a narrower second column.
2. Doll about 45 percent of the frame height, pack grid about 30 percent, belt as a row at the pack's top or bottom edge (D2R puts gold directly under the pack) [V images].
3. Stats live in the same frame as the doll (Grim Dawn right half, Warframe right column) so changing gear shows the number change; use two-column label/value rows 20 to 24 px high [V images].
4. Stash is a second panel at the same height as the pack, not a separate screen (D2R stash screen mirrors inventory) [V image]. 6x8 is 48 cells; at 36 px cells that is 216x288 px, fits a column.
5. Tooltip anchors beside the hovered cell, with a fixed width of 300 to 360 px, never scrolling (D4 tooltip scrolls: avoid; Cookie Clicker anchors beside the row: good) [V images].
6. Comparison is automatic (Grim Dawn: equipped beside hovered), and deltas are inline in green/red next to the changed stat (D4 +34) [V images]. Shift is an optional modifier only to swap sides.
7. Requirements are a single row at the foot of the tooltip, one chip per requirement, red when unmet, plus a mark on the cell itself when unmet (PoE1 row, Last Epoch X mark [S]) [V image PoE1].
8. Contextual actions are a labelled list under the tooltip or beside the selection with key chips (D4 list; Warframe buttons beside the selected row) [V images].
9. Currency and capacity sit in the header or foot as icon+number pairs (Genshin header, D2R gold bar) [V images].
10. Owner badge on items: "worn by hero or merc" shown on the tile and in the detail (Genshin "Equipped:" strip) [V image].
11. Gear sets are a tab or segmented control on the doll (D2R I/II tabs), not a new window [V image].
12. Fixed cell size from a tokens table (36 px at 1440x900) and the whole frame derives from it: grid of cells is the only unit; this avoids the mismatched edges seen in Grim Dawn and PoE1 (empty textured boxes, floating flask row) [V images].

### G3. Candidate arrangements at 1440x900

Pixel budgets assume a 1440x900 viewport, 8 px gutters, 44 px top bar. Panel widths are budgets, not measured.

**Watch screen**

Option W1 "Hero left, dashboard rail right" (Idle Champions + RimWorld + DF):
```
+-------------------------------------------------------------+ 44 top bar: name, health lights (DF style strip)
| status sentence (full width, 36 px)                         | 36
+----------------------------------+--------------------------+
| VIDEO 16:9   872 x 490           | needs-you chips   (alert  | 120
|                                  | stack, RimWorld)          |
|                                  +--------------------------+
+----------------------------------+ progress tiles 3x1  100  |
| ZONE MAP (click to send) 872x230 +--------------------------+
|                                  | ORDERS list  rows 32 px   | 430
+----------------------------------+--------------------------+
```
Left column 872 px (60 percent), right rail 552 px. Supported by: B4 (hero + overlay), E2 (alert stack), E3 (strip), C1 (stat tile), B3 (list rows).

Option W2 "Three bands" (Grafana grid + HA panel view):
```
top: health lights + status sentence in one 64 px band
mid: VIDEO 960x540 | MAP 456x540 (square-ish, click to send)
bot: progress tiles x4 | needs-you | orders (3 columns of 262 px, 220 px high)
```
Rows 64 / 540 / 220 / gutters = 900. Supported by: C1 (rows of tiles), C3 (panel view for hero), E3.

Option W3 "Overlay-free sidebar, video as full-bleed centre" (Antimatter rail + Idle Champions):
```
left rail 220 px: progress (level, xp/h, chapter), health lights, orders (40 px rows)
centre 880 px: VIDEO 880x495, status sentence under it (40 px), map 880x300 (with map as the picker)
right rail 320 px: needs-you chips, pending actions, log tail
```
Supported by: B1 and B3 (rail rows), D1 (centre clear), E2.

**Character screen**

Option C1 "Docs left, pack centre, stash right" (D2R stash screen + Grim Dawn):
```
hero doll 300 | merc doll 200 | pack 10x4 (36 px) 400 x 150 | stash 6x8 (36 px) 230 x 300
stats under dolls, belt row under pack, gold + sets in a foot bar
```
Supported by: A1, A7, A10.

Option C2 "Spread" (Moonlighter two pages + Warframe row actions):
```
left page 640: stash 6x8 above, pack 10x4 below      right page 640: hero | merc dolls (45 percent height), stats table (30 percent), sets and gold foot
tooltip card fixed in a 340 px column at the far right, contextual actions under it
```
Supported by: A10, A8, A3.

Option C3 "Selected-item column" (Genshin inventory + D4 tooltip):
```
left 800: dolls (hero + merc) top 400, pack 10x4 + belt below
centre 300: stash 6x8
right 340: sticky detail card (hover or select): name, requirement row, stats with deltas vs equipped, actions list, owner strip
```
Supported by: F1, A3, A7.

### G4. Anti-patterns (seen in the references)
- Masonry packing: unpredictable placement; the HA team rejected it on purpose [V text].
- Blank regions until data arrives (Cookie Clicker centre column, Bloomberg empty tab body, Grim Dawn empty right-bottom pane) [V images].
- Pie charts with legends leave a quarter of the tile empty (Grafana) [V image].
- Cards taller than 250 px for one number (Melvor skill cards) [V image].
- Stacked modals (D3 hero details) and ornate frames (D2R and D3 consume 10 to 15 percent of the panel) [V images].
- Tall scrolling tooltips (D4) and tooltips far from the cursor (PoE1 capture) [V images].
- 3D model as the centrepiece when the info is the point (Warframe, Genshin, D4) [V images].
- Symmetric columns with empty gutters (Destiny 2) [V image].
- Icon-only rails with no labels (Dwarf Fortress left rail) [V image].
- Collapsible groups and tabs for required information on a no-scroll monitor [V text for Datadog groups].
- Scrolling column layouts for a status page (Kittens Game) [V image].
- Unverified: a fixed 'ideal' widget count; one source says under about 20 per dashboard, I found no primary source [S].
