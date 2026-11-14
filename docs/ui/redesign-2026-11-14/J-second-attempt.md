# J. Second attempt: the cockpit and the character canvas

> status: design proposal, second attempt (2026-11-14, branch `agent/uidesign3`), not built; waits for the user's verdict. It replaces the Standard and Simple presets of D-layouts.md and the canvas of I-character-screen.md section 3. It keeps: the Orders model (C), the requirement rules and instant actions (I.5 to I.10), the loot-rules needs, the compatibility promise, every answer the user gave. Tags: V = seen in a render named or read in code, U = unverified.

## 0. Why the first attempt failed (the user's verdict, 2026-11-14, after seeing wave 1 live)

"Both are better, but this overall is disappointing." Watch: about 2000 px tall at 1920 wide, dead space, panels all over the place. Inventory: a masonry of unrelated cards, the hero and the mercenary in different regions, the tooltip is a card.

What I got wrong, plainly:

1. **I rearranged cards on a scrolling grid.** A 12-column grid with content-height rows cannot produce a screen that fits: every tile has its own height, so bottoms are ragged and the page grows. The first design never had a height budget.
2. **I reused too many existing tiles** and let fallbacks to old tiles stand in, so the live page was the old tiles in new places.
3. **I never rendered at the target viewports** and judged a mockup that scrolled.
4. **The character screen kept three zones with different inner heights** and put the mercenary where there was room, not next to the hero.

## 1. Acceptance criteria (from the user, binding for this design and for the build)

1. The default watch screen fits without scrolling at 1440x900, 1920x1080 and 1366x768, and degrades by design at 390x844. Visible at once: the picture, the status sentence, what it is doing and why, needs you, health, progress, the map to send the hero, the orders.
2. Zero dead space: tiles tile the screen exactly.
3. The character screen is one continuous canvas, the hero and the mercenary in one frame, storage as one connected area, a real hover tooltip, actions as a contextual menu and toolbar, gear sets as a compact bar, repair and gold as inline chips; it fits 1440x900 and 1920x1080.

Each is checked on renders at those exact sizes (section 6), not by eye on a scrolling page.

## 2. The watch cockpit

A fixed frame that fills the viewport. Nothing in it has a content-driven height; every region has a budget and its content adapts (ellipsis, internal scroll for a long list, fewer rows).

```
1440 x 900
+----------------------------------------------------------------------------------------------------------+ 44
| d.idle  [FARMING] Cold Plains . you ordered it, 12 min left      life == 86  mana == 105  (2 need you)  Character Skills ... |
+-----------------------------------------------------------------------+----------------------------------+
|                                                                       | NEEDS YOU                        |
|                                                                       |  Weapon broken ...  [Repair now] |
|                                                                       |  Go for Blood Raven? [Go][Not yet]|
|                       game picture, 4:3                               +----------------------------------+
|                       height = viewport - band - strip                | ORDERS                    1 open |
|                       (960 x 720 at 1440 x 900)                       |  [Go to][Town][Repair][Pause][More]|
|                                                                       |  > Farm Cold Plains 30 min  ACTIVE x|
|                                                                       |  ~ then the plan: level 10 ...   |
|                                                                       +----------------------------------+
|                                                                       | WHERE TO        [Act I | Area]   |
|                                                                       |  Monastery line                  |
|                                                                       |  Outer Cloister                  |
+-----------------------+---------------------------+-------------------+  ... 3 columns x 10 rows of zone |
| PROGRESS  level 7     | JOURNEY  chapter 3 of 9   | HEALTH            |  buttons, band stripe, pins,     |
| to 10 ====-- 25 min   | o-o-(3)-o-o-o-o-o-o       |  o    o    o  o  o|  route line, here and target     |
| xp/h 12.4k  gold/h 2.1k| Blood Raven: too early [Go][Hold]| Life Gear Ready Merc Supplies|                    |
+-----------------------+---------------------------+-------------------+----------------------------------+
```

| Region | Size rule | Holds | When empty or long |
|---|---|---|---|
| Band | 44 px | brand, the status sentence (state chip, place, reason; click for Now / Would change it / Next / uptime), life and mana bars with one number each, the needs-you count, Character, Skills, the "..." menu | the reason is cut with an ellipsis; the full text is in the popover |
| Picture | height = viewport height - band - strip - gaps; width = height x 4/3 (capped so the rail keeps at least 400 px) | the game, edge to edge in its tile, no title bar, no margin box; the zone name and Full screen as two small labels on the picture's edge | no picture: the same box says why ("reconnecting") |
| Strip | 112 px under the picture, three cells | Progress (level meter with ETA, xp/h with a sparkline, gold/h); Journey (the nine chapters as a dotted path, the next chapter with its band, Go now and Hold); Health (five lights with a word each) | a rate still building shows a dash |
| Rail | the rest of the width, full height | Needs you (0 to 3 one-line rows, each with its button); Orders (verb row, one-line rows, the plan row); Where to (the act map; a switch to the local area map) | Needs you takes no height when empty and the map grows; more than three alerts show "+N"; a long order list scrolls inside its tile |

Width classes, all by rule, none by wrapping:

| Viewport | Picture | Rail | Difference |
|---|---|---|---|
| 1366 x 768 | 784 x 588 | 558 | as drawn |
| 1440 x 900 | 960 x 720 | 456 | as drawn |
| 1920 x 1080 | 1200 x 900 | 696 | the rail is two columns: Needs you, Orders and the area map on the left, the act map full height on the right, so both maps show |
| 390 x 844 | 390 x 292 | - | band (two lines), picture, then one of three tabs fills the rest: Orders (with Needs you on top), Map, Progress (progress, journey, health). No page scroll |

What left the default screen (each is one click away, none is deleted): the onboarding card (a "Getting started 1 of 5" chip in the band's menu until done, opening as a sheet), Hero sheet (on the character screen), Why as a tile (the sentence and its popover), Tally and Marks (the progress popover), Thoughts and the logs (a Log drawer from the "..." menu and from the sentence popover), the voice line (it becomes the popover's last line), Gear condition (a Needs you row and the Gear light), the map's thirteen controls (zoom and Layers on the area map only), Rules, Quests, Cain's notes, Atlas, Still locked (the Debug and Classic layouts).

### The act map in a rail

The map is three columns by ten rows of zones (V `internal/watch/mapdata.go`: X is one of three values, Y one of ten). Drawn as a grid of one-line buttons it needs about 440 by 380 px to be readable at 11 px text, which is what the rail has. Each button: a stripe for the level band (green in range, amber above, red too hard, grey below, dashed not orderable), the name, and at the right the level range, "here", or the queue number. The route to the current target is a gold line behind the buttons. A click opens the zone menu as a sheet over the lower part of the map tile (never over the picture).

### Fixes after the user's first look (2026-11-14, window about 1914x930)

"Massive improvement", and: was it meant to be cut off? Two things were cut, neither by the frame:

| Seen | Cause | Fix |
|---|---|---|
| The bottom zone rows were hidden | the zone menu opened as a sheet over the lower part of the map | the menu never covers a zone button: beside the map (in the area tile) on a two-column rail, above it (over the Orders tile) otherwise; the picked zone is outlined |
| The area map showed only a strip of a room | the drawing was scaled to cover its tile | it is scaled to fit, whole |

Supported sizes are now two lists, and the fit check uses both: full screen 1920x1080, 1440x900, 1366x768; browser windows 1914x930, 1904x1000, 1536x730, 1280x620; phone 390x844. On short windows the strip is 92 px, zone rows are one line (all ten rows always show; none under 20 px), and the strip's cells drop their second words. Renders: [research/iterations/fix/](research/iterations/fix/).

## 3. How this fits the shell (compatibility)

The free grid (`watch.layout/1`, 12 columns, 48 px rows, fit rows, the editor) stays exactly as it is for every existing layout and for Classic, Debug and any layout the user made. A screen that must fit the viewport cannot be expressed in content-height rows, so the cockpit is a second, additive kind of layout:

```json
{ "schema": "watch.layout/1",
  "template": "cockpit/1",
  "slots": {
    "band":  ["statusline", "hero-meters", "needs-count"],
    "stage": "video",
    "strip": ["progress", "journey-strip", "health"],
    "rail":  ["needs-you", "orders", "zone-map"],
    "railWide": { "left": ["needs-you", "orders", "area-map"], "right": ["zone-map"] },
    "phoneTabs": { "Orders": ["needs-you", "orders"], "Map": ["zone-map"], "Progress": ["progress", "journey-strip", "health"] }
  },
  "sizes": { "strip": 112, "railMin": 400 },
  "breakpoints": { } }
```

- A shell that does not know `template` ignores it and uses `breakpoints` (the file also carries the old-style placements as a fallback, so it never renders blank).
- A stored layout without `template` is untouched.
- The editor on a cockpit layout: tiles are still movable, into slots. Drag a tile onto another slot to swap them; drag the two dividers (strip height, rail width) within limits; a slot's tile can be replaced from the list of panels that declare `fits: "slot"`. Free placement stays in the free-grid layouts.
- Panels that may sit in a slot declare `"slot": true` in their manifest and must render inside a fixed box (no content-driven height). The eight cockpit panels are new panels (`statusline`, `needs-count`, `needs-you`, `orders`, `zone-map`, `area-map`, `progress`, `journey-strip`, `health`); none is an old tile restyled. The old tiles stay for the free-grid layouts.
- Fallbacks (D.4) are dropped for the cockpit: a slot whose panel does not exist yet is not shown as an old tile. The cockpit ships as the default only when its panels exist (section 7).

## 4. The character canvas

One frame, one cell unit `u` derived from the viewport height (12 cells of storage plus chrome), shared by the dolls, the pack, the stash and the belt. Two zones that touch.

```
1440 x 900 (u = 58)
+ band: d.idle  Simmy . Sorceress 7 . Cold Plains, in the field              (1 needs you)  Watch Skills ... +
+ bar: [search /] (Usable by me)(By mercenary)(Unidentified)(Marked sell)   Gold 1,240  Stash 5,300 [Repair 180 g][Sell marked (1)] +
+--------------------------------------------------------------+--------------------------------------------+
| HERO  Sorceress 7            | MERCENARY Desert Guard 14     | PACK 10 x 4                                |
| +------+ [helm] [am] +------+| ALIVE  life ====  Str 60 Dex 40| +--+--+--+--+--+--+--+--+--+--+            |
| |weapon| [armor]     |shield|| [weapon][helm ]               | |  |  |  |  |  |  |  |  |  |  |            |
| |      |             |      || [      ][armor]               | +--+--+--+--+--+--+--+--+--+--+            |
| +------+             +------+|                               +--------------------------+-----------------+
| [gloves][ring][belt][ring][boots]                            | STASH 6 x 8              | BELT [][][][]   |
| SETS (Fire damage)(Town)(Magic find)(+)                      | +--+--+--+--+--+--+      | SELECTED        |
| +---------+---------+---------+---------+ +------+------+    | |  |  |  |  |  |  |      |  Light Plate    |
| | Level 7 | Str 38  | Dex 25  | Vit 30  | | Level| Life |    | |  |  |  |  |  |  |      |  [Give to merc] |
| | Life    | Mana    | Defense | Damage  | | Str  | Dex  |    | ...                      |  [Stash][Sell]  |
| | Fire    | Cold    | Light   | Poison  | | Def  | Aura |    |                          | LOOT RULES      |
| | xp      | next lvl| Gear    | Points  | |             |    |                          |  1 Runes.. 4    |
| +---------+---------+---------+---------+ +-------------+    |                          |  2 ...          |
+--------------------------------------------------------------+--------------------------+-----------------+
```

| Part | Rule |
|---|---|
| Who wears (left, takes the spare width) | The hero's doll (8 by 6 cells) and beside it the mercenary's (only the slots his type has) with his state chip, life bar and a six-value sheet under his doll; the gear set bar under both; the hero's stat sheet as one block of sixteen values; under that, side by side, the loot rules summary and the latest item actions. This is where the watch screen's Hero tile went |
| Storage (right, exactly 10 cells wide) | The pack on top; under it the stash (6 by 8) and beside the stash a 4-cell column: the belt, the selected item (its requirements, comparison and its actions as buttons), the town errands (deposit and withdraw gold, buy scrolls, identify all), the locked "Stash from anywhere" line |
| Bar | Search, four filter chips, then gold, stash gold, Repair (with the price, red when something is broken) and Sell marked as chips. No card for repair or gold |
| Tooltip | A real floating tooltip anchored beside the hovered cell (flips at the edge, never scrolls): name in its quality colour, the game's lines, where the item is, requirement rows for the hero and for the mercenary with met and unmet marks and the gap, the comparison against the worn item as old to new values with signed deltas, one line of hints. On a phone the same content is a bottom sheet |
| Actions | Right-click: the context menu with every action and its key; a refused action is listed greyed with its reason. Double-click: the first action. The selected item's box shows the same list as buttons. No form at the bottom of the page |
| Mercenary states | none: the column says so and how one is hired; dead: the state chip is red and Revive is the one primary button; unknown: header only |
| Widths | 1366 x 768: u = 50, same arrangement. 1920 x 1080: u = 72. 390: two tabs (Hero and mercenary; Pack and stash); the bar scrolls sideways; the storage tab scrolls inside itself |

The sketch above shows the first arrangement of this section; the final one is in the renders ([research/iterations/final/](research/iterations/final/)).

## 5. Visual system changes (on top of E)

- Density: body 12.5 px, labels 10 px small caps, rows 28 px (38 on touch), tile padding 8 to 10 px, gaps 8 px. The first attempt's 14 px body and 12 px gaps cost about a fifth of the screen.
- Tiles are flat surfaces with no border and no title bar of their own; a label of one line is the whole chrome.
- One accent use per region: the primary verb, the current target, the current chapter.
- The picture has no frame except a hairline; nothing overlaps it but two small labels on its edge.
- State colours unchanged (E.2).

## 6. How it was checked

Renders at 1440x900, 1920x1080, 1366x768 and 390x844 with headless Chromium on the builder (`research/render.sh`), six renders in three rounds of critique; the pictures and the critique are in [research/](research/README.md). The live pages the user saw are described there as measured block diagrams, not as pictures: the live screenshots contain the game's own art, which this repository and its published pages never carry.

## 7. What this does to the plan

See G-build-plan.md and briefs/README.md for the revised cards. In short: a new shell slice for the cockpit template; the cockpit panels are built as slot panels with fixed boxes; Standard-as-default by fallbacks is withdrawn (the cockpit becomes the default when its panels exist; until then the default stays Classic); the character screen brief is rewritten around the one-unit canvas.
