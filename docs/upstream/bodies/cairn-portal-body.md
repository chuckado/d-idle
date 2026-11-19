## What changes

When the fifth Cairn Stone is touched, game.exe opens the red portal to Tristram just inside the ring of pillars, a little right of and above StoneLambda. d2d opened it 6 tiles right and 3 tiles up of StoneLambda, outside the ring. A player who rejoins gets the portal from a timer; game.exe puts it 4 subtiles off StoneAlpha, d2d put it 4 tiles off.

The offsets are game.exe's subtile numbers; d2d added them to `Npc::x` / `y`, which are tiles. The fix divides by 5 at the two lines (`World::cain_operate` in `quests/shared.cpp`, `World::stones_portal` in `setup.cpp`). The colour (object 60, red) and the 300-frame flash are unchanged.

Three commits: the oracle (`tools/emu/cairn_portal.py`, `tests/game/test_cairn_portal.cpp --cairn-portal`, a README row), the claim, and the finding in `docs/research/re/quests-act1.md`.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00593710` `0x593a16..0x593a2f` | loads StoneLambda's x and y from `FUN_00620870`, `add eax, 6` (x), `add ecx, -3` (y), pushes `0x120` (missile 288, `cairnstones`) and calls `FUN_0056ede0` |
| `FUN_00620870` `0x6208e5..0x6208f8` | a type 2 unit (an object): x, y = the path's `+0xc`, `+0x10`: integers (subtiles) |
| `FUN_0056ede0` `0x56ee3c`, `0x56ee70..0x56ee79` | the position goes into the creation record's `+0x10` / `+0x14` (absolute, no offset added) for `FUN_0059fa30`; a distance gate of 100 units from the source first |
| `FUN_005a9930` `0x5a99a5..0x5a99bd` | the missile's server function: `FUN_0056d130(room, missile x, missile y, level, 0, 0x3c, 1)`: object 60 at the missile's own position |
| `FUN_00592d50` `0x592dd5`, `0x592dd9` | the timer's portal: StoneAlpha's x + 4 and y + 4, level 0x26, object `0x3c` |

## Counted differentials (V run, the builder, tree = upstream `52d4cdd1` plus the branch)

Oracle: `cairn_portal.py game` runs game.exe's own `FUN_00593710` (the fifth stone; `FUN_00620870` reads StoneLambda, `FUN_0056ede0`'s position is recorded, then `FUN_005a9930` runs on a missile placed there and its `FUN_0056d130` call is recorded) and `FUN_00592d50` (the timer) on the stones d2d lays out; `test_cairn_portal --cairn-portal 40 1` prints d2d's two portals for the same stones, in subtiles.

| tree | rows | stones' portal differs | timer's portal differs |
|---|---|---|---|
| parent (oracle commit) | 39 | **39** | **39** |
| claim | 39 | **0** | **0** |

On the parent the stones' portal stands (30, -15) subtiles from StoneLambda against game.exe's (6, -3), the timer's (20, 20) against (4, 4). The 39 rows are eight distinct places of the stone piece; `test_cairn_portal` pins game.exe's values for those eight (no emulator needed).

Mutants of the claim (each caught by `test_cairn_portal`, 12 of 12): tile units, a wrong sign of x or y, swapped axes, off by one in x or y, StoneAlpha as the anchor, and the same family for the timer's portal (tile units, a wrong sign, off by one, the wrong stone).

Build and tests (the builder, Release, GCC, `-j2`): full build rc 0; `ctest`: 394 tests, 393 passed, 13 skipped (data or emulator), 1 failed: `smoke_d2d` (`tests/smoke_d2d.py:444`), which fails the same way on upstream `52d4cdd1` rebuilt alone. clang-tidy-19 on the new test: 0 findings.

## In play

Open the Cairn Stones in Stony Field: the portal stands inside the ring (1.2 tiles right, 0.6 tiles up of StoneLambda, next to the Stone Sigil), as in the original. A player who rejoins an opened game finds it 4 subtiles from StoneAlpha, not 4 tiles.

## Limits

- Whether a unit's integer position is a subtile is the project's standing convention (the same one as the Duriel portal and the warp oracle); this PR does not measure the scale.
- "The missile stands at the position `FUN_0056ede0` was given" (`FUN_0059fa30`'s record) is read, not run. `FUN_005a9930`'s own call of `FUN_0056d130` is run, with the missile placed at that position.
- Not changed: the Gibbet fallback portal (`object.x + 6, object.y + 6`, `quests/shared.cpp`): no game.exe call that adds (6, 6) was found, so its source is unproven here.
- Map seed 20 gave no row on the d2d side (not diagnosed); the other 39 of 40 seeds, eight distinct layouts, are compared.
- No game, no screenshot, no Windows or macOS build.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_cairn_portal test_cairn_missile
D2_MPQ_DIR=<MPQs> D2_PATCH_INSTALLER=<LODPatch_114d.exe> build/tests/test_cairn_portal            # the pins: parent aborts, claim OK
cd tools/emu && uv sync
set -o pipefail
D2_MPQ_DIR=<MPQs> D2_PATCH_INSTALLER=<LODPatch_114d.exe> ../../build/tests/test_cairn_portal --cairn-portal 40 1 > /tmp/d.tsv
uv run python cairn_portal.py game /tmp/d.tsv > /tmp/g.tsv
uv run python cairn_portal.py compare /tmp/d.tsv /tmp/g.tsv     # parent: 39 + 39 mismatches, exit 1; claim: 0, exit 0 (also exit 1 on an empty or malformed file)
```

## Checklist

- [x] A test in `tests/` (`test_cairn_portal`, pins game.exe's values on eight maps)
- [x] The finding in `docs/research/re/` (`quests-act1.md`)
- [x] `python3 tools/pr_guard.py` passes
- [x] `ctest` with the game data: 394 tests, 393 passed, 13 skipped, 1 failed; `smoke_d2d` fails on main too
- [x] Labels: `oracle`, `test`, `documentation`

---

Agent: Claude Code
