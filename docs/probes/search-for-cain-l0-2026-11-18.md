# Search for Cain, drill L0: object listings of the tree and the stones (2026-11-18)

> status: result of the read-only drill CAIN-L0 named in `docs/probes/search-for-cain-2026-11-17.md` section 6 (that diagnosis is on branch `agent/cainfix`). Tags: V verified live (fixture named), S secondary, U unverified.

## What was run, and one deviation from the brief

The questions are about the d2d listings (`debug presets`, `debug objects`), which exist only in the d2d lab build, not in the Game.exe instance. So NO Wine instance was started (i1 stays stopped; nothing of the real game was touched). The drill ran on a d2d lab slot of the worker host: lab binary `d2d.lab3-labverbs` (lockstep, no save, lab verbs), one fresh throwaway Sorceress "Drilly" (`d2drill -probe`, every `debug` verb ledgered with a reason, no memory write of a running game; setup verbs used: `debug goto`, `debug warp`, `debug kill`, `debug operate` as DIAGNOSIS). Command: `docs/probes/cainl0/run.sh LABEL 'verb;verb;...' [SLOT]` (env `SEED`, default 11), output in `docs/probes/cainl0/testdata/` (each file is the probe transcript). The character is CHEATED (warped): mechanics only, never evidence about balance; logged in `docs/probes/test-characters.md`.

## (a) The tree of Inifuss

1. The reply of `debug presets` is not only the placed-object rows: after them it prints one `room <type> <x> <y>` row per room of the level (about 60 of the 100 slots; V `s1.txt`, `w4.txt`). `d2dctl.ParsePreset` does not read these, they end in `Reply.Bad` and nobody looks. `room 161` is the Dark Wood room that holds the tree (V seeds 11, 3, 5, 7, 9: rooms (44,44), (52,12), (60,68), (52,60), (36,28); the tree row then stands at room origin + 0.5, +0.5: (44.5,44.5), (52.5,12.5), (60.5,68.5), (52.5,60.5), (36.5,28.5); files `s1`, `tr3`, `tr5`, `tr7`, `sd9`).
2. The tree's place is RANDOM per seed (five seeds, five places over the whole 80x80 level). It is not "west/north-west". A bot that explores by coverage and waits for the preset to show finds it at a random time.
3. The preset row exists only after its room is built. Class is 30, name `Tree of Inifuss`, hidden flag 0, so `fieldObjects` would list it (V). Arriving from the Underground Passage (the real entry, Dark Wood cell 14.5,18.7) only 5 rows (torches, class 269) are listed and no tree (V `w4`: tree room at (44,44), 40 cells away). Warping along the diagonal and stepping 25 ticks each: not listed at (30,33) (17.8 cells from the room origin), listed at (34,36) (12.8 cells) and closer (V `w5`). So the threshold lies between 12.8 and 17.8 cells from the room origin (about one room more than the 8 cell room size; U exact).
4. Never absent: with the player near, `debug presets` always had the row (5 of 5 seeds). The class id 30 is right (V). The sandbox park "object class 30 not found" therefore is the exploration budget, not a wrong class or a missing verb (the log shows 79 percent explored at the last try, not 100).
5. `debug objects` lists the same object (operable ones) with the same index as the preset row; the tree does not show in it unless built (`s1`: index 0 `op12 30`).

## (b) Mode and the Cairn Stones

- Preset rows carry no mode (V: columns are index, class, x, y, hidden, name). `debug objects` has columns index, x, y, kind/op, class, locked flag, state: `NU` or `ON` (V `st2`; torches class 37 `ON`, chests, shrines, stones `NU`). It is a real second source for a lit state (the index equals the preset index when both list the object; objects with no operate function, e.g. class 22 and 61, are missing from it).
- Stony Field stones (V `st1`, `st2`, `st4`): rooms 160 (20,60) and 162 (28,20). Presets at room 160: class 17 (19.1,59.1), 20 (20.7,62.1), 21 (19.1,60.7), 22 (18.1,63.1, no name), 19 (21.9,60.9), 18 (20.7,59.1), and class 61 (20.5,60.5, the Tristram-portal spot, absent from objects). The stones' state is `NU` and stayed `NU` after `cmd interact <index>` for the class 17 stone (player stood next to it, seven runs), and `debug portals` stayed empty. `debug operate` returned a shrine style reply and changed nothing (`st8`).
- So in this lab build the stones show NO signal: no state change, no anim field, no portal, empty interact reply (`st6`, `st7`, `st8`). Whether the fork implements the stone handler at all, and whether a CORRECT sequence lights them, is U (only the first stone was touched; the right order is not known to us, hover scans for the stone found nothing, `hover=-1`).
- Room 162 (28,20) holds the Moldy Tome (class 8) and torches; the many rows of classes 160 to 162 at (18..21,26..29) around it are U (decoration, not stones).

## What the fix must be (feeds the CAIN-1 slice)

- Parse `room` rows in `debug presets` (new `d2dctl.Rooms()` or `Preset` with Kind room), and in d2dworld expose the position of room type 161 for Dark Wood (also 160/162 for Stony Field). Walk to room origin + (0.5,0.5) with nav; the preset row appears on the way. No coverage exploration needed; the 100 s `objectExploreBudget` and the retry counter stop mattering. Data: room type to quest object table in the quest data, not Go literals.
- Keep a bounded fallback when no room row (older fork): explore as today, but count a try only when coverage did not grow.
- Stones: keep the lit-state read from `debug objects` (`NU`/`ON`) behind an interface but plan the no-state path (try orders, wait for the portal object, class 61) as primary until a live check shows the stones change state when touched in the right order. Needs one more drill (CAIN-L2): all five stones in a known order on a throwaway, in Stony Field room 160, watching `debug objects`, `debug portals` and the presets for class 61.
- Class 22 in the stone room is hidden from `debug objects` and unnamed: U what it is.

## Fixtures
`docs/probes/cainl0/testdata/*.txt`: `s1` (Dark Wood presets and objects at the start, seed 11), `w4` (arrival from the Underground Passage), `w5` (walk-in distance series), `sd3 sd5 sd7 sd9` and `tr3 tr5 tr7` (other seeds), `w3` (warps of area 10 and 5), `st1 st2 st4` (Stony Field listings), `st6 st7 st8` (interact and operate on a stone), `w1` (Stony Field warps and layout).
