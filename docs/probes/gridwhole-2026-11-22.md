# gridwhole: is `debug grid 5` whole at arrival? (card gridwhole, 2026-11-22)

> status: current. Card `gridwhole` (first slice of `docs/design/fable-review-2026-11-22/map-knowledge.md`, user answers of 2026-11-22: Q1 drop the whole-map verb pending this probe). Tags: V = measured in a named run, S = source or doc reading, U = unverified. Setup is CHEATED (`debug goto`, `debug warp`, `debug kill` are SETUP verbs of a lab d2d on a throwaway hero): mechanics evidence about the engine's own collision, not about play.

## Verdict

**V: the collision grid of the current level is whole on arrival on every level probed, mazes included.** In 106 runs (64 levels, seeds 5, 7, 11 for 22 levels and seed 5 for 42 more: all of act 1 including Catacombs 1 to 4, Jail 1 to 3, Tower Cellar 1 to 5, Crypt, Mausoleum, caves and passages, plus acts 2 to 5) the reply of `debug grid 5` has exactly 5 x width columns and 5 x height rows of `debug level`, holds only `.` and `#` (no unknown cell), and the grid read after the hero was warped over the whole level (every room laid: `debug laid`) has the same size. So `FullyKnown() == true` is TRUE as a fact for d2d, and the "known grid edge" reading of the Countess walk is wrong (see the conflict below). Maze levels do not differ from the Blood Moor in this respect, so the whole-map verb of the review's question stays dropped (Q1).

**Not literally "0 differences".** The arrival grid and the all-rooms-up grid differ in a few cells: raw 0 to 1586 cells per run (rooms bring objects, doors and monsters with them), after taking away every cell within 15 subtiles of a monster or object of either snapshot: 0 in 75 of 106 runs, at most 103 cells (0.15 percent of a level; Worldstone Keep 1) in the others, always compact patches (for example a 2 x 6 block in the Crypt, seed 5, level 18 subtile x 77..78, y 134..139, opened) and never a region: a missing room would be at least a DS1 room of 8 x 8 tiles (1600 subtiles). Repeating the runs with every monster killed before both reads (`GW_KILL=1`, 23 runs on 8 levels, the job was still running when this was written) left the same kind of residue (0 to 93 cells), so it is not monsters. The cause of the residue is U: compact footprints of objects the `debug objects` list does not carry at the read (doors, scenery) or tile re-picking when rooms come up (`relevel`, docs/d2d-backend.md section 3). Tower Cellar 1 to 5 (the maze of the Countess) have 0 residue in all 15 runs.

Rooms: the hero warp laid every room in 98 of 106 runs; in 8 it did not (`rooms_up_final` below `rooms_total`: Monastery Gate 23 of 24, Outer Cloister 32 of 35, levels 76, 77, 78 (act 3 jungle, up to 192 rooms), 107 (River of Flame, 139 of 297), 110 (164 of 180), 114 (61 of 64)), so for those the "after" grid is not the grid of a fully laid level; they stay whole in size and symbols.

The engine's own collision query: `debug blocked x y` (cells) on 200 random cells of each grid disagrees with the grid in at most 6 cells at arrival and 7 after (3.5 percent) in the first 64 runs, which counted every mismatch; later runs only count those away from every monster and object, at most 3. U why (a live unit blocks the query but may not the grid is the likely reading). The grid verb and the query read the same collision.

## The conflict of the review (section 2), settled

`internal/nav/mouthgoal_test.go` says the Countess' stand (21,70) was "100 subtiles west and outside the known map" and that (121,70) was "the known grid's edge". V (Tower Cellar 5, level 25, all three seeds): the level is 30 x 30 cells = **150 x 150 subtiles**, whole at arrival (150 rows of 150 columns). (21,70) and (121,70) are both inside it, 129 subtiles from the east edge; neither is an edge. The nav needs no map beyond the grid to plan the whole route; (121,70) was the first corner of a normal path (S: the walk's own plan), refused by the guessed mouth bound. The job log named in the review (20261121-005917) is no longer on the builder (the job list was pruned), so the comment could not be checked against it; the probe is the evidence. Card knownedge corrects the test comment.

## Table (per level: seeds probed, A = at arrival, B = after the warp over all rooms)

| level | seeds | grid whole | rooms up at arrival | A vs B raw | unmasked A vs B | blocked query mismatches (A, B) |
|---|---|---|---|---|---|---|
| 2 Blood Moor | 3 | all | 9 of 83/9 of 84/9 of 80 | 200/264/227 | 0/51/0 | 1/0/0, 1/0/1 |
| 3 Cold Plains | 3 | all | 9 of 94/9 of 96/9 of 95 | 405/400/355 | 11/5/1 | 0/1/1, 0/0/0 |
| 4 Stony Field | 3 | all | 9 of 92/9 of 89/9 of 92 | 625/507/543 | 27/17/4 | 1/0/0, 2/0/2 |
| 5 Dark Wood | 3 | all | 9 of 91/9 of 89/9 of 92 | 389/341/469 | 2/0/0 | 2/1/0, 0/3/0 |
| 6 Black Marsh | 1 | all | 9 of 94 | 418 | 0 | 2, 2 |
| 7 Tamoe Highland | 3 | all | 9 of 98/9 of 98/9 of 97 | 612/664/589 | 95/84/24 | 1/2/2, 3/1/0 |
| 8 Den of Evil | 3 | all | 0 of 27/0 of 27/0 of 27 | 131/151/125 | 0/0/0 | 1/1/3, 2/0/3 |
| 9 Cave 1 | 3 | all | 9 of 63/9 of 54/9 of 45 | 184/150/106 | 0/0/0 | 1/1/0, 2/3/3 |
| 10 Underground Passage 1 | 3 | all | 0 of 54/9 of 63/0 of 72 | 162/211/322 | 0/0/0 | 0/0/1, 1/2/4 |
| 13 Cave 2 | 3 | all | 9 of 9/9 of 9/9 of 9 | 0/0/0 | 0/0/0 | 1/0/2, 2/1/0 |
| 14 Underground Passage 2 | 3 | all | 9 of 9/9 of 9/9 of 9 | 0/0/0 | 0/0/0 | 0/2/0, 2/3/2 |
| 17 Burial Grounds | 3 | all | 9 of 30/9 of 30/9 of 30 | 5/5/5 | 0/0/0 | 1/4/0, 1/1/2 |
| 18 Crypt | 3 | all | 9 of 14/8 of 12/7 of 13 | 150/112/379 | 10/0/0 | 3/0/0, 2/1/3 |
| 19 Mausoleum | 3 | all | 9 of 14/7 of 14/8 of 13 | 290/388/299 | 0/0/20 | 6/2/0, 2/2/4 |
| 20 Forgotten Tower | 1 | all | 1 of 1 | 0 | 0 | 0, 0 |
| 21 Tower Cellar 1 | 3 | all | 5 of 7/6 of 7/5 of 6 | 39/5/5 | 0/0/0 | 1/0/0, 1/1/3 |
| 22 Tower Cellar 2 | 3 | all | 5 of 7/6 of 8/4 of 6 | 21/10/25 | 0/0/0 | 0/3/0, 0/0/0 |
| 23 Tower Cellar 3 | 3 | all | 5 of 7/6 of 7/5 of 7 | 39/5/5 | 0/0/0 | 2/1/3, 1/1/7 |
| 24 Tower Cellar 4 | 3 | all | 0 of 6/5 of 7/5 of 6 | 425/0/31 | 0/0/0 | 4/1/1, 1/0/3 |
| 25 Tower Cellar 5 | 3 | all | 9 of 16/9 of 16/9 of 16 | 45/24/24 | 0/0/0 | 4/1/1, 5/3/2 |
| 26 Monastery Gate | 1 | all | 9 of 24 | 20 | 10 | 0, 0 |
| 27 Outer Cloister | 1 | all | 9 of 35 | 138 | 0 | 2, 2 |
| 28 Barracks | 1 | all | 15 of 44 | 179 | 17 | 2, 2 |
| 29 Jail 1 | 3 | all | 7 of 14/6 of 14/8 of 14 | 254/202/256 | 0/0/0 | 1/1/1, 3/1/0 |
| 30 Jail 2 | 1 | all | 7 of 16 | 342 | 0 | 0, 2 |
| 31 Jail 3 | 1 | all | 8 of 16 | 381 | 0 | 1, 0 |
| 32 Inner Cloister | 1 | all | 9 of 9 | 0 | 0 | 0, 0 |
| 33 Cathedral | 1 | all | 9 of 20 | 120 | 29 | 0, 1 |
| 34 Catacombs 1 | 3 | all | 7 of 13/7 of 13/6 of 12 | 500/570/228 | 21/0/0 | 1/0/0, 2/1/1 |
| 35 Catacombs 2 | 3 | all | 0 of 15/8 of 14/8 of 15 | 818/385/518 | 24/22/11 | 3/2/1, 1/2/1 |
| 36 Catacombs 3 | 1 | all | 8 of 16 | 422 | 21 | 1, 1 |
| 37 Catacombs 4 | 1 | all | 9 of 12 | 99 | 0 | 5, 1 |
| 41  | 1 | all | 9 of 92 | 449 | 0 | 1, 2 |
| 42  | 1 | all | 9 of 94 | 616 | 0 | 0, 1 |
| 43  | 1 | all | 9 of 95 | 371 | 0 | 1, 0 |
| 44  | 1 | all | 9 of 93 | 663 | 0 | 0, 0 |
| 45  | 1 | all | 9 of 16 | 0 | 0 | 0, 1 |
| 46  | 1 | all | 9 of 98 | 712 | 0 | 1, 1 |
| 47  | 1 | all | 0 of 14 | 67 | 0 | 0, 0 |
| 48  | 1 | all | 6 of 14 | 98 | 0 | 1, 2 |
| 49  | 1 | all | 7 of 14 | 76 | 0 | 4, 1 |
| 55  | 1 | all | 8 of 16 | 51 | 0 | 1, 2 |
| 56  | 1 | all | 7 of 24 | 157 | 0 | 2, 1 |
| 62  | 1 | all | 9 of 20 | 208 | 0 | 1, 1 |
| 74  | 1 | all | 5 of 61 | 279 | 0 | 0, 1 |
| 76  | 1 | all | 9 of 192 | 752 | 0 | 0, 0 |
| 77  | 1 | all | 9 of 192 | 506 | 0 | 0, 0 |
| 78  | 1 | all | 9 of 192 | 645 | 0 | 2, 1 |
| 79  | 1 | all | 9 of 80 | 1586 | 0 | 1, 1 |
| 80  | 1 | all | 9 of 80 | 1477 | 6 | 1, 0 |
| 83  | 1 | all | 9 of 64 | 409 | 0 | 0, 0 |
| 85  | 1 | all | 9 of 16 | 96 | 39 | 3, 0 |
| 100  | 1 | all | 0 of 11 | 579 | 0 | 0, 0 |
| 104  | 1 | all | 9 of 75 | 1136 | 0 | 1, 1 |
| 105  | 1 | all | 9 of 76 | 947 | 12 | 1, 1 |
| 106  | 1 | all | 9 of 74 | 751 | 0 | 0, 0 |
| 107  | 1 | all | 9 of 297 | 308 | 0 | 0, 0 |
| 110  | 1 | all | 9 of 180 | 1459 | 61 | 1, 0 |
| 111  | 1 | all | 9 of 160 | 980 | 11 | 1, 1 |
| 112  | 1 | all | 9 of 160 | 1238 | 69 | 0, 1 |
| 113  | 1 | all | 9 of 52 | 303 | 48 | 2, 1 |
| 114  | 1 | all | 9 of 64 | 413 | 62 | 1, 0 |
| 115  | 1 | all | 7 of 44 | 263 | 34 | 0, 1 |
| 128  | 1 | all | 9 of 40 | 442 | 103 | 1, 1 |

Columns: "rooms up at arrival" is `debug laid` (rooms in play of all rooms); "A vs B raw" the number of differing subtile cells between the two grid reads (per seed); "unmasked" the same without the cells within 15 subtiles of any monster or object of the two snapshots; "blocked query mismatches" disagreements of 200 sampled cells with `debug blocked`, away from units and objects (probe version of 22:00; the first 64 runs counted every mismatch).

## Method and how to repeat

- Probe `tools/probes/gridwhole.py` (stdlib Python over the control socket; lab binary `d2d` of the private fork branch `didle/devp18-20261121`, 0efbf3be, `--headless --lab-verbs`, one process per level and seed, game data on the builder). Steps per run: `debug goto L`, `debug level` (rectangle), `debug grid 5` (A), `debug laid` (rooms), sample `debug blocked`, warp the hero over a lattice of walkable cells until every room is up, `debug grid 5` (B), compare.
- Run: `GW_KILL=1 python3 tools/probes/gridwhole.py D2D DATA 25,2 5,7,11 200` (one JSON line per run). Builder long lane jobs 20261121-211411-c507 (act 1 mazes and outdoors, 3 seeds; then acts 1 to 5, 42 levels, seed 5) and 20261121-224639-494a (kill variant).
- Not done: a cell-by-cell classification of the residue (needs the object collision list), and levels beyond those in the table. The act 2 to 5 levels were probed with seed 5 only.

## Consequence for the cards

- knownedge: go ahead; `FullyKnown` can be measured (known cells equal the rectangle) and is TRUE on every level here; the optimistic "edge" walk has nothing to work on for d2d.
- The arrival differential of the review ("0 differences apart from unit stamps") holds up to a residue of compact object-sized patches (U cause).
