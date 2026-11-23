# Search for Cain, CAIN-4 and CAIN-5a: cairn stones and the way to Tristram (2026-11-18)

> status: current, offline work (agent cain45). Tags: V verified (live lab drill L1 or a test), S source or doc read, U unverified. Inputs: `docs/probes/search-for-cain-l1-2026-11-18.md` (rule, V on the d2d lab build, character CHEATED: mechanics only), `docs/probes/search-for-cain-2026-11-17.md` (slice plan).

## What changed

- `cmd/d2bot/cain.go`: `solveCairn` rewritten (the old backtracking assumed a wrong stone resets the lit ones, WRONG per L1). Cycle through the stones that are not lit, read the lit state after every click, stop at the portal. Bounds: 15 clicks with the lit state (5+4+3+2+1, tight for the reverse order), 25 blind (23 in the worst of the 120 orders). Error with the lit set and "Key to the Cairn Stones" when a whole cycle lit nothing. Stones are resolved by class (17..21) before every click (indexes and ids shift as rooms load, V). V tests: all 120 orders, both modes, no click on a lit stone, missing key, bound, portal lag, ctx cancel.
- `cairnStones` enters the portal and checks level 38 (`tristramArea`); any other level is an error naming it. Wrong area stays a prereq.
- `internal/d2dsrc/items.go`: Unit.Mode of the five stones (classes 17..21) = state column of `debug objects` (NU 0, ON 1) joined to the preset by index; read only when a stone is in view, every snapshot. A `debug portals` row with which >= 2 becomes class 60 (the generic red portal, `quest.ObjRedPortal`), which 0/1 stay class 59 (town portal). This also stops a quest portal counting as a town portal cast in the d2d portal verb.
- `internal/task/act1.go`: the stones step has a Ready fault (prereq) "no Key to the Cairn Stones in the pack" unless the key (`bkd`, found by code since CAIN-2) is held or the portal is open. The step is then blocked with the reason instead of clicking.
- `data/questboard/act1.json`: chunk `free_cain` split in `stones` (travel_stony_field, touch_stones, enter_tristram; SUPPORTED) and `free_cain` (clear_tristram, free_cain, return_camp; still "not supported yet": Tristram clear, the cage and Cain never ran live, drill L2). The board shows "part N of 5" for the quest (the count follows the chunk list).

## Can the sandbox d2d build (devp10) see the stones and the portal?

- `debug portals`: YES (V: the bot reads it in every snapshot, `readBatch`, in the sandbox since before devp10; the row `2 4 25.1 57.7` is the same verb).
- `debug presets` (stone classes, positions): YES (V: the sandbox log of 2026-11-17 reads presets, `fieldObjects`).
- `debug objects` (the NU/ON state): U. The verb is listed in the stock d2d `debug` family (docs/d2d-spike.md, S) and `d2dctl.Objects` exists, but no sandbox log or doc shows a reply from it, and L0 and L1 ran on the lab build. The private fork checkout is not on this machine, so the source gate was not read. devp10 carries `--lab-verbs` only as an opt-in that must stay off in the sandbox.
- Fallback built in (works with only the shipped verbs): when `debug objects` fails or a stone has no state, `Unit.Mode` stays -1, the solver goes blind (cycle all five, portal watch, at most 25 clicks) and the portal is seen through `debug portals` (class 60, which 2). A lit stone ignoring a click is harmless (V), so the blind cycle is correct, only slower.
- Private-fork addition if `debug objects` turns out to be lab-only (a devp11 change, fork branch pushed the same day, no GPL patch in this repo): serve the `debug objects` read verb (same rows: `index x y op<N> class locked NU|ON`) without `--lab-verbs`, behind `--item-verbs` or a new read-only flag; nothing else is needed for the stones. For the quest portal no change is needed (`debug portals` row which 2).

## Live checks (not run; no live game in this slice)

1. Sandbox: send `debug objects` in Stony Field; expect rows `op9 17..21 NU`. If it errs, the blind path runs (log line `object.mode` unknown).
2. With `bkd` in the pack, run the board chunk `stones`: expect <= 15 clicks (state read) or <= 25 (blind), the portal, level 38.
3. Game.exe: the stone Mode after a click (U); the red portal class 60 near StoneLambda + (6,-3) (U).
