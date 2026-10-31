# Map contiguity across levels (stream and viewer map pane)

> status: implemented offline in `cmd/d2dview` (card mapitem); live check pending (last section). Tags: V verified (data or decompiled code read), I inferred, U unverified.

## The game's rule (game.exe 1.14d automap, from the decompile notes in the d2decomped tree `docs/research/re/automap.md`, read on the worker host)
- (a) V: the automap keeps one cell store per **layer**, the Levels.txt `Layer` column of the level the player is in (the saved `.ma` file is keyed the same way). Cells are placed by act-absolute tile position (`x = (tx - ty) * 80 / 10`), so every explored room of a layer is drawn together, whichever level of that layer the player stands in. The player's room and its near rooms are revealed as the player moves (every 80 client px, only tiles that were on screen); a room created for another layer switches the store and back.
- (b) V: the Rogue Encampment, Blood Moor and the other Act 1 wilderness share layer 0, so town and field are one map (the same for Lut Gholein, Kurast, Harrogath with their fields).
- (c) V: caves, crypts, towers, Tristram, Cow level, sewers, tombs and the other dungeons have a layer of their own (Levels.txt gives each its own number), so their map stands alone and the field is not drawn in it.
- (d) V: the scroll is recomputed from the player's client position every frame (`rules::automap_scroll`), the layer's cells stay: the view re-centres on the player, the layout does not change. A town portal is the same layer in the field and the town, so both parts of the map stay and the view jumps with the player (I: no special case in the code).
- (e) V: explored cells are saved per character (`Name.map`, `.ma0` to `.ma3`), per map seed (the last four are kept), per layer. U: whether d2d reveals the same set of rooms as game.exe on a room callback (d2d reveals what its renderer drew, see the automap notes).
- SME not asked: the decompile and Levels.txt settle the rule. Caveat: d2d's `Layer` values were read from a stock Levels.txt copy on the worker host (I: same as the 1.14d MPQ).

## Which levels are stitched (shared Levels.txt layer, V; level numbers as in `cmd/d2dview/areas.go`)
- Act 1, layer 0: 1 Rogue Encampment, 2 Blood Moor, 3 Cold Plains, 4 Stony Field, 5 Dark Wood, 6 Black Marsh, 7 Tamoe Highland, 17 Burial Grounds, 26 Monastery Gate, 27 Outer Cloister, 28 Barracks, 32 Inner Cloister, 33 Cathedral.
- Alone (own layer): 8 Den of Evil, 9 to 16 caves and holes, 18 Crypt, 19 Mausoleum, 20 Forgotten Tower, 21 to 25 Tower cellars, 29 to 31 Jail, 34 to 37 Catacombs, 38 Tristram, 39 Moo Moo Farm.
- Act 2: 40 to 45 (Lut Gholein, Rocky Waste, Dry Hills, Far Oasis, Lost City, Valley of Snakes) are one layer; 46 Canyon of the Magi and everything below it alone.
- Act 3: 75 to 83 (Kurast Docks to Travincal) one layer. Act 4: 103 to 106 one layer, 107 to 108 another. Act 5: 109 to 112 one layer; caves alone.
- Levels.txt also gives 125 and 135 a shared layer 92 (unused ids in the stock table): not stitched. The table is `stitchGroups` in `cmd/d2dview/stitch.go`.

## What d2d reports
- `debug level` gives the player cell (level-relative, tiles) and the level's act-tile origin; d2dworld keeps one grid per visited level (level-relative subtiles) and `LevelMap.Origin` (5 x origin, V live in docs/fixwp.md W5). Level origins are act-absolute, so two levels that touch have origins that differ by exactly the width or height of one of them (the DRLG places them edge to edge, docs/d2d-exit.md). Adjacent entries are not needed for the stitch.
- Never-visited neighbours have no grid, so only visited levels are drawn (the game also draws only what was explored).

## The design (cmd/d2dview)
- One frame: `Snap.X/Y`, monsters, exits and `MapRect` are in act tiles and act subtiles (level origin added). A crossing moves nothing on screen, so the camera is continuous without any smoothing.
- `/map.png` is the composite of every known level of the player's stitch group (`stitchGroups`), one pixel per subtile, with its box in the `X-Map-Rect` header. A cell known to two grids (each reaches one tile past its edge) is drawn from the lower level number, a fixed order, so crossing does not redraw. The picture is rebuilt only when a member grid grows or a level joins; `mapVer` counts rebuilds; `Pub.stitch` names the group.
- The pages fetch the new image aside and swap it in with its box when loaded (no blank frame); the trail survives a stitched crossing and restarts when the group changes or on a jump over 30 tiles (town portal).
- Explored memory is the d2dworld per-level grid: it survives transitions, is dropped only on a new game (`Reset`). Public safety unchanged: only tiles, counts and the group number reach `/api/pub`.
- The dashboard pane `web/live/panels/world.js` (bot server, one level) is not changed; item 13 covered the stream and d2dview pages.

## Live check (later, not done)
Hidden slot recipe: `docs/probes/drills/README.md`. Start a slot d2d with the viewer (`d2dview`), open `/stream` and `/` in a headless browser, walk (or `debug goto`) Blood Moor to Cold Plains and back, screenshot the map pane every 2 s across the crossing: expect one picture with both fields, the player dot moving without a jump, no blank frame. Then enter the Den of Evil: only the cave map; leave: the field map is back. Then a town portal from the field: both parts kept, the trail restarts. Also read `X-Map-Rect` and a border cell of both levels (U: level origins of live Blood Moor and Cold Plains touch as the DRLG says).
