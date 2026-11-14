# Research and iteration log for the second attempt

> status: working record (2026-11-14, branch `agent/uidesign3`). The design it led to is [../J-second-attempt.md](../J-second-attempt.md). Tags: V = I looked at the render or page named, S = from the research file's own tags, U = unverified.

## 1. What is here

| File | What |
|---|---|
| [references.md](references.md) | The reference study: about 28 interfaces (ARPG inventories, idle dashboards, monitoring tools, management sims, a mobile hub), gathered from the web by a research agent that downloaded and looked at about 25 usable screenshots; each with its layout grammar, what to take and what to reject; synthesis rules, candidate arrangements and anti-patterns. No reference image is in this repository |
| [render.sh](render.sh) | Renders the mockups at 1440x900, 1920x1080, 1366x768 and 390x844 with headless Chromium on the builder |
| [iterations/](iterations/) | One folder per iteration: the page as it was (`src/`), its renders, and the critique in section 4 below |

## 2. The pages the user rejected, measured (V: two screenshots of the live sandbox the coordinator saved; not committed, they contain the game's art)

Watch, 1920 px wide, **2113 px tall** (1.96 screens at 1080):

| Block | Box (x, y, w, h) | Problem |
|---|---|---|
| Band | 0, 0, 1920, 58 | fine |
| Voice line | 12, 78, 1896, 40 | a full-width row for one sentence |
| Getting started | 12, 138, 1896, 106 | a full-width band holding "1 of 5 done" and one button |
| Game | 12, 258, 940, 764 | a title bar and padding around the picture |
| Map | 966, 258, 464, 826 | thirteen buttons, then a map that uses about a tenth of its black box, then five lines of legend |
| Gear condition, Hero now, Why, Thoughts | 1444, 258 to 1084, 464 wide | four cards for four sentences |
| Tally | 12, 1038, 940, 166 | two numbers in a half-screen-wide card |
| Journey | 12, 1218, 940, 286 | a third of it empty |
| Where to | 966, 1098, 464, 706 | the act map, one screen below the fold |
| Hero | 490, 1518, 462, 346 | alone, with an empty 464 by 346 hole to its left |
| (empty) | 1444, 1098, 464, 766 | a dead column |
| Doing now | 12, 1878, 1896, 222 | a full-width band repeating Hero now |

About 0.6 million of the page's 4.06 million pixels are empty holes; another large share is padding inside under-filled cards. Nothing below y = 1080 is visible without scrolling: the act map, the journey, the hero, the order.

Inventory, 1920 px wide, **1965 px tall**:

| Block | Box | Problem |
|---|---|---|
| Equipped | 12, 122, 1088, 978 | one card holding the doll (316 px wide), a 730 px item panel, the mercenary fold, gear sets |
| Backpack | 1114, 122, 776, 278 | the pack uses 480 px of 776 |
| (empty) | 1114, 400, 776, 248 | dead |
| Repair | 1114, 648, 776, 218 | a card for one button |
| (empty) | 1114, 866, 776, 248 | dead |
| Belt | 12, 1114, 1088, 106 | four cells in a 1088 px card |
| (empty) | 12, 1220, 1088, 348 | dead |
| Stash | 1114, 1114, 776, 440 | the grid uses 288 px of 776 |
| Item actions | 12, 1568, 1088, 356 | a form at the bottom of the page |

## 3. What the research changed

| Finding (source group) | Change in the design |
|---|---|
| Monitoring tools and management sims give every region a fixed place and size on a grid; masonry was rejected on purpose by the dashboard teams (Home Assistant, Grafana, Factorio, RimWorld) | The cockpit: a frame with budgets instead of a scrolling grid of cards. The first attempt's whole approach is dropped |
| The live view is the hero and gets half to three fifths of the width; the rest is rails and strips docked around it, not floating over it (Idle Champions, Dwarf Fortress, Factorio; overlays rejected because the game must stay readable) | The picture is sized by the viewport height; a rail to its right, one strip under it; nothing over the picture |
| One status line (Dwarf Fortress top strip, boss bars) | The sentence stays in the band; the Why tile and Doing now tile go |
| Lists are rows of 24 to 40 px with one right-aligned action; cards taller than about 100 px for one number are an anti-pattern (Melvor) | Orders and Needs you are one-line rows; Tally, Why, Thoughts, Gear as cards are gone |
| Health as one row of label-over-value pairs; progress as three or four stat tiles 90 to 120 px high (Grafana stat tiles) | The strip: Progress, Journey, Health at 112 px |
| The map is one panel and is the target picker | The act map sits in the rail, always visible; the area map is the optional second view |
| Tabs and collapsing only for what is not required | The only switch on the default screen is Act / Area on the map tile (the area map is not in the required set); at 1920 both show |
| ARPG inventories keep the doll, the pack and the stash in one frame on one cell size; mismatched edges and empty textured boxes are the visible failures (Grim Dawn, Path of Exile) | One unit `u` for dolls, pack, stash and belt; two zones that touch |
| The stats live in the same frame as the doll (Grim Dawn, Warframe) | The stat sheets sit under the dolls; this is where the watch screen's Hero tile went |
| Tooltip beside the hovered cell, fixed width, never scrolling; comparison automatic with inline deltas; requirements as a compact block with an unmet mark on the cell; an action list with key hints (Diablo IV, Grim Dawn, Path of Exile) | The floating tooltip with requirement rows for hero and mercenary, old-to-new values with signed deltas, an owner line, and the context menu with keys |
| Currency and capacity as number pairs in a header (Diablo II's gold bar, mobile hubs) | Gold, stash gold, Repair with its price and Sell marked as chips in the bar |
| Rejected: ornate frames that eat 10 to 15 percent of a panel (Diablo II Resurrected, Diablo III), a 3D model as the centrepiece, stacked modals, a single "better" arrow | Flat surfaces, a hairline around the picture, no verdict in the comparison |

## 4. Iterations and critique

Six renders were made; each folder under [iterations/](iterations/) has the page it was rendered from (`src/`) and the picture at 1440x900 (the final one at all four sizes plus one with the tooltip open). All were rendered with headless Chromium at the exact window sizes. **Where:** the brief asked for the builder; its single job slot was held by other agents' gates for over an hour (three queued renders never started), so the renders were made on the head box with a headless Chromium shell and two libraries unpacked into a scratch directory, no root, a few seconds each. Nothing ran on the worker host that runs the games. `render.sh` runs unchanged on the builder when it is free.

| Iteration | What changed | Critique after looking at it (V) |
|---|---|---|
| it1 | The cockpit as first written; the character canvas with three zones (hero, storage, mercenary) | Watch: fits and tiles, but the progress cell is bottom-heavy with clipped text, the health words are cut, zone names are cut by the level range beside them, the map's empty corner cells read as holes. Character: the mercenary is on the far side of the storage, away from the hero: the exact complaint. A third zone only to hold him |
| it2 | Character: hero and mercenary in one frame, two zones; tooltip deltas and owner line; watch: two-line zone buttons | Character page rendered **blank**: a script error from an edit made without a render (a name used twice). Watch as it1 otherwise. Lesson kept: no edit without a render |
| it3 | Script fixed; watch: progress and health laid out top-down, range moved to the zone's second line, legend and act switch in the map's empty corners, demo control into the band | Watch: reads well at 1440. Character: works; the mercenary's note wraps one word per line, two cells of empty space under his doll, the selected-item box is mostly empty, the item-actions list sits in the storage column |
| it4 | Character: mercenary sheet under his doll, hero sheet as one eight-column block, loot rules and item actions side by side under it, town errands (gold, scrolls, Cain) beside the stash | Tooltip good. The mercenary sheet was squeezed beside the doll (unreadable); the bottom lists left a 110 px empty band; at 1920 the area map drawing is tiny in its box and zone names are cut |
| it5 | Mercenary sheet as three columns under the doll, list rows share the remaining height, wider act map at 1920 | 1366x768 and 1440x900 good. Phone: the status sentence was pushed out of the band; the storage side column is too narrow on a phone |
| final | Phone band fixed, phone storage stacked, the selected box shows the item's requirements and comparison | See "still weak" below |

### Still weak (V on the final renders)

1. The act map's shape leaves grid cells unused (the left and right columns have no zone in rows 6 and 7 and little in rows 1 to 3). The legend and the act switch fill the top corners; the middle gaps are the map's own shape. It is not dead tile space, but it is not perfectly dense either.
2. Character at 1920x1080: beside the mercenary's doll about 190 by 300 px is unused, because the dolls keep the one cell size and the zone takes the spare width.
3. Character at 1366x768: the column beside the stash is 200 px wide; two action labels are cut and the requirement block shows only its first rows.
4. The doll slots are labelled boxes. With the player's own item art they will look different; spacing between slots has not been judged with art.
5. The area map is an optional view at 1440x900 and below (a switch on the map tile); only 1920 wide shows both maps at once.
6. Phone: the Orders tab has spare room under a short list; Progress, Journey and Health are one tab away, not on the first view.
7. Everything is a static mockup with fake data. None of it has run against the real shell, real panel data or real text lengths; the fit check of slice S1b is what will hold the build to these pictures.
8. The cockpit needs a new layout kind in the shell (slots). That is more shell work than the first attempt promised, and the tile editor becomes "swap and resize slots" on this layout instead of free placement.
