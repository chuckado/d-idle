# PR candidate D1: the Cairn Stones' Tristram portal offset (fid/cairn-portal-r1), 2026-11-18

> status: current, 2026-11-18 (card cairnportal, TASK-648, Sonnet). Branch `fid/cairn-portal-r1` on upstream main `52d4cdd1`, local and on the private fork remote only; NOTHING posted (an independent Opus verifier runs first). Source claim: `docs/research/decomp40-cairn-portal.md` D1. Tags: V run (game.exe's own code run in the emulator, on the builder, today), V read (disassembly), I (inferred), U (unverified).

## 1. What the PR claims, and what it does not

Claim: game.exe puts the Tristram portal of the Cairn Stones at StoneLambda's unit position plus (6, -3) in SUBTILES (a tile is 5x5 subtiles), and the timer's portal (a player who rejoins) at StoneAlpha plus (4, 4) subtiles. d2d added the same numbers to `Npc::x` / `y`, which are TILES, so the portal stood 6 tiles right and 3 tiles up of StoneLambda (outside the ring of pillars; our live drill L1 saw the row `2 4 25.1 57.7`) and the timer's one 4 tiles from StoneAlpha. The fix divides by 5 at two lines.

Not in this PR: the red or blue colour (already red upstream since 11-07), the translated scroll's pictogram (D2 / D7, a feature), the Gibbet fallback portal `shared.cpp` `open_portal_at(object.x + 6, object.y + 6)` (see section 5), the rng order (D3), the sigil (D4).

## 2. Commits (private fork, on upstream `52d4cdd1`)

| Commit | What |
|---|---|
| `463e440b` | oracle: `tools/emu/cairn_portal.py`, `tests/game/test_cairn_portal.cpp --cairn-portal N FIRST`, a README row |
| `f8e4ac32` | claim: `shared.cpp` (`lambda->x + 6.f / 5`, `lambda->y - 3.f / 5`), `setup.cpp` (`alpha->x + 4.f / 5`, `alpha->y + 4.f / 5`), the test pins game.exe's values on eight maps |
| `f81ca8da` | docs: `docs/research/re/quests-act1.md` (units) |

Author and committer the public noreply identity; `Agent: Claude Code` the only trailer, after a blank line (V run, `git interpret-trailers --parse`). Pushed to the private fork remote (`fid/cairn-portal-r1`, head `f81ca8da`); never posted.

## 3. Oracle (V run)

`cairn_portal.py game` runs game.exe's own code natively on the stones d2d lays out (`test_cairn_portal --cairn-portal` prints StoneAlpha's and StoneLambda's subtile positions and d2d's two portals, `floor(tile * 5)`, one row per map seed):

1. The fifth stone: `FUN_00593710` (OperateFn 9) on a quest record whose fifth right click this is. `FUN_00620870` reads StoneLambda's position (native), `FUN_0056ede0` (the call that makes missile 288 at `0x593a2f`) is hooked and records its stack arguments (x, y). Then `FUN_005a9930` (that missile's server function) runs native on a missile unit placed at (x, y) and its call of `FUN_0056d130` (the object maker, object class `0x3c`, the red portal) is recorded: the portal's position.
2. The timer: `FUN_00592d50` native with StoneAlpha a type 2 unit; `FUN_0056d130`'s arguments recorded.

Hooked (not run): the quest record getter, the player lookups and quest-bit tests, the item and mode setters, the room lookups and the object maker's body. Not run: `FUN_0059fa30` (the missile's creation from `FUN_0056ede0`'s record: x and y go to its `+0x10` / `+0x14`, V read at `0x56ee70..0x56ee79`); so "the missile stands at (x, y)" is V read, not V run.

Output on the parent (oracle commit, 40 map seeds, 39 rows; seed 20 gave no row on the d2d side, not diagnosed): every one of the 39 rows differs on both portals: the fifth stone's d2d offset is (30, -15) subtiles, game.exe's (6, -3); the timer's d2d (20, 20), game.exe's (4, 4). Example, seed 5: StoneLambda (255, 263), game.exe (261, 260), d2d (285, 248). `compare` exits 1.

## 4. Claim: numbers

| Tree | rows | stones portal differs | timer portal differs | exit |
|---|---|---|---|---|
| parent (oracle commit `463e440b`, seeds 1 to 40, 39 rows, built from that exact commit) | 39 | **39** | **39** | 1 |
| claim (`f8e4ac32`, same seeds) | 39 | **0** | **0** | 0 |

`compare` also exits 1 on an empty d2d file, an empty game file, a malformed row and a missing file; `game` and no argument exit 1 (V run). The eight distinct stone places (StoneLambda always 8 subtiles below StoneAlpha) and game.exe's values are the pins of `test_cairn_portal` (no emulator needed): OK on the claim; each source mutant below fails it.

Mutants (V run, one text change each in the claim's `shared.cpp` / `setup.cpp`, `test_cairn_portal` rebuilt each time): 12 of 12 fail the test (exit 1, a `FAIL:` line naming the game.exe value): tile units for the stones' portal (the parent), wrong sign of y, wrong sign of x, axes swapped, off by one in x, off by one in y, the other stone (StoneAlpha) as the anchor, tile units for the timer, wrong sign of the timer's x, timer off by one in y, timer axes (y negated), the timer anchored on StoneLambda.

Build and tests (V run, the builder, Release, GCC, tests and tools on, `-j2` under the lane lock, every job registered): full build rc 0 (1,262 targets, 1,597 s); `ctest -j2` on the claim tree: 394 tests, 393 passed, 13 skipped (data or emulator), 1 failed: `smoke_d2d` (`tests/smoke_d2d.py:444`, the character-select assert `st["class"] == "3"`), which fails the same way on upstream `52d4cdd1` rebuilt alone (V run): not this PR's. `test_cairn_missile`, `test_cairn_dip`, `test_cairn_client_event` pass. clang-tidy-19 on `test_cairn_portal.cpp`: 0 findings (one `misc-include-cleaner` finding fixed). `tools/pr_guard.py upstream/main HEAD`: 0 FAIL 0 WARN. `tools/presubmit.py --quick --base upstream/main`: 0 FAIL, 2 WARN (build skipped under --quick, answered above; no compile_commands in the local checkout, tidy run on the builder).

## 5. Limits (honest)

- Units: that a unit's integer position is a subtile (5 per tile) is the project's standing convention (`rules`, the Duriel portal `orifice->x - 13.f/5`, the `warp_spot` oracle); this PR does not measure the scale itself. The 100-unit distance gate `FUN_006417f0 <= 0x64` at the top of `FUN_0056ede0` (V read) fits a subtile unit.
- The step from `FUN_0056ede0` to the missile's position is read, not run (section 3).
- `nearest_free`: d2d moves the portal to the nearest free subtile when the spot is blocked. On all 39 maps the spot after the fix equals game.exe's own value (so no nudging happened there); whether `FUN_0056d130` itself nudges a blocked spot in game.exe is not measured.
- The Gibbet fallback at `shared.cpp` (`object.x + 6, object.y + 6`) is NOT changed: the game.exe calls of `FUN_0056d130` found by a scan of the call sites are `0x54b8f3 0x582aa8 0x58b01d 0x58bd33 0x592de8 0x5933eb 0x594228 0x5964d3 0x59b0e7 0x59cba8 0x5a99bd 0x5be32a`; `0x5933eb` adds (3, 3) to a unit's position and makes object `0x3b` (a blue portal), `0x594228` / `0x59b0e7` pick a free spot with `FUN_00545340`. None is the "(6, 6)" the fork comment cites; the Gibbet path is U. Left for a card of its own.
- Map layouts: Stony Field's stone piece appears at eight places over seeds 1 to 40, always with StoneLambda 8 subtiles below StoneAlpha; the pins cover all eight.
- No game, no lab drill; no Windows or macOS build.
- The portal object's colour and the 300-frame flash missile are untouched.

## 6. Drill recipe (not run)

Lab binary with the cain verbs (docs/forkdbg.md): fresh Sorceress, `debug give bkd`, warp to level 4, click the five stones by the cycle of L1, then `debug portals` and `debug presets`: expected after the fix `portal 2` at StoneLambda + (1.2, -0.6) tiles, within 0.5 tile of the class 61 preset (the Stone Sigil). Rejoin: leave and re-enter the game after the stones opened; the timer's portal at StoneAlpha + (0.8, 0.8).

## 7. Bot consequence (not changed)

`docs/research/decomp40-cairn-portal.md` section 5: the bot takes the portal by the portal list or object class 60, never from a computed offset, so the d2d fix changes nothing in the bot; after it the portal stands inside the ring (about 1.2 tiles from StoneLambda), next to stone footprints: keep the generous interaction radius. B3 of that trace stays true.

## 8. Matrix row (proposed text)

`A1Q4 Tristram portal | red object 60 at StoneLambda + (6,-3) subtiles (timer: StoneAlpha + (4,4)) | d2d: red, offset in tiles until fid/cairn-portal-r1 (D1, V run)`.

## Shared-file lines for the merger
Generated from the task card at merge.
