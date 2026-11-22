# exitsread: the engine's link table against the bot's exit data (card exitsread, 2026-11-22)

> status: current. Card `exitsread` (map-knowledge review card 4). Tags: V = measured in a named run, S = source or doc reading, U = unverified. Setup is CHEATED (`debug goto` on a throwaway hero of a lab d2d): link facts only.

## Question

The stuck steps "Monastery Gate to Outer Cloister: nav no usable gate" and "no exit to that level: level 27 from 26" could be a wrong exit in the bot's data or something else (a border that is not readable yet). The bot's route graph (`internal/nav/route.go` `edges`) is built from the hand-kept hint tables (`act*hints.go`), Levels.txt Vis columns, exits and borders it has seen. The engine's own table is `debug warps` (tile exits with destination) and `debug nearby` (levels linked past the current level's edge; d2d builds neighbours on a worker thread, so the list grows after arrival).

## Method

`tools/probes/exitsread.py` (stdlib Python; lab d2d of the private fork branch `didle/devp18-20261121`, `--headless --lab-verbs`): per level and seed `debug goto L`, then `debug level`, `debug warps`, `debug layout`, and `debug nearby` polled until unchanged for 8 s. One JSON line per run. `tools/probes/exitscmp` (Go; reads the game data of the install and `nav.RouteHints`) prints, per level, the engine links the static data lacks ("engine only") and the static data links the engine never listed ("data only"). V: builder long and main lane jobs 20261121-232055-80c8 (probe, 37 levels seed 5, then 7 levels seed 7: levels 26 to 37, 1 to 10, 13, 14, 17 to 25, 40, 75, 103, 109) and 20261121-234910-dcbc (comparison). The neighbour list was complete about 1 s after arrival on every level (`settle_s` 0.1 to 1.2).

## Findings

1. V: **the bot's data has every exit the engine lists.** For all 40 levels probed, every tile exit (`debug warps` destination) and every border neighbour that is an exit in the real game is an edge of the static graph. The Monastery Gate, Outer Cloister, Barracks, Jail 1 to 3, Inner Cloister, Cathedral and Catacombs 1 to 4 chain is complete: 26 -> 27 and 27 -> 28 and 32 -> 33 are BORDER links in the engine (Monastery Gate lists no warp and `nearby` [.., 27]; Outer Cloister lists no warp and `nearby` [26, 28]; Inner Cloister `nearby` [33]), all others are warps. So the stuck step "no usable gate" at the Monastery Gate is not a data hole: the edge exists in data and in the engine, and the failure is the border strip being unreadable at that moment (the neighbour's collision is built late; `d2dworld` link wait, `internal/d2dworld/border.go`). Cause U (not reproduced here: a border read needs the walk, and this probe only reads the link table); the next step is a probe of the strip at 26 -> 27 after the link wait, not a data change.
2. V: **one wrong data edge: Burial Grounds (17) -> Pandemonium Run 1 (133).** Levels.txt lists 133 in a Vis column of the Burial Grounds; the engine's table for 17 has warps to 18 and 19 and borders to 1, 2, 3 only. 133 to 135 are the Uber levels reached by portal only. The route graph joined them to the Burial Grounds (and so to every walk's graph). Fixed: `portalOnly` in `internal/nav/route.go` skips Vis links to 133 to 135 (133 is V, 134 and 135 are S, same family); must-fail test `TestVisLinkToPortalOnlyLevelIsNotAnEdge` (fails without the filter, two Vis orders).
3. V (information, no change): `debug nearby` lists more levels than there are exits. Cold Plains (3) lists the Rogue Encampment (1), Stony Field (4) lists 1 and 2, Dark Wood 1 and 2, Black Marsh 1, 2 and the Monastery Gate, Tamoe Highland 1 and 2, the Burial Grounds 1 and 2, the Pandemonium Fortress the Chaos Sanctum: these are levels the act layout places beside or past the edge (the list is levels "linked", not "passable"). They are NOT edges to add: nothing walks from Cold Plains into the town. (These are the "engine only" lines of the comparison.) U: whether every real border link also has passable cells at its strip; the nav decides that from the grid and `debug blocked` across the edge, not from this list.
4. V: warps seen: Lut Gholein lists the Sewers (47) twice and the Palace (50) once; Den and Caves list one or two each. No level in the probed set has a warp to a level the data does not know.

## Not done

- Acts 2 to 5 beyond levels 40, 75, 103, 109 (arrival levels only), seeds beyond 5 and 7 for the cloister chain; the border strip read at 26 -> 27 (needs the walk or a `debug blocked` strip probe after the link wait); the second card for the Monastery Gate failure.
- The hint tables themselves were compared only through the route graph (Vis plus hints); an edge only in a hint would show as "data only" (none did for the levels probed).

## Repeat

```
python3 tools/probes/exitsread.py D2D DATA 26,27,28,29,33 5,7 8 > exits.jsonl
go run ./tools/probes/exitscmp -data DIR_WITH_MPQS exits.jsonl
```
