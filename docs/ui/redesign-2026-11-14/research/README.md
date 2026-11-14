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

(Filled in as each iteration was rendered and looked at.)
