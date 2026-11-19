# Verifier: Crushing Blow D5, the divisor of a plain unique (2026-11-19)

> status: current, 2026-11-19 (card verifycrushd5, Opus, singly; independent verifier before posting). Nothing posted. Subject: private fork branch `fid/crush-d5-r1` (oracle `9eac185e`, claim `bfef14cf`, on upstream `9b64c11a`), body `docs/upstream/bodies/crush-d5-body.md`, trace `docs/research/decomp41-merc-crush-trace.md` section 2. The fixed `fid/crush-d5-r2` is on the private fork remote only.

Tags: **V run** (run by this verifier today, the builder container), **V read** (disassembly or source read by this verifier), **I** (inferred), **S** (an earlier doc, named), **U** (unverified).

## Verdict: r1 FAIL (stale base, pin gaps, scan wording), r2 PASS WITH CHANGES

The rule is game.exe's and r1's one-line fix implements it: 0 of 29,360 differ on the claim, 5,632 on the parent, every one a plain unique (V run, the grid is exhaustive: 734 MonStats rows x 5 kinds x 1 to 8 players, so there is no seed to vary). The class mapping holds on both sides and no other row of d2d's divisor table is wrong (below). r1 should not be posted as it is:

1. **Stale base (high).** r1 is on `9b64c11a`; upstream main is `d5de50ca`, 83 commits on. It rebases with no conflict, but the posted commits must be the rebased ones, built and tested there.
2. **The pins and the oracle miss mutants (high).** Four plausible mutants pass r1's `test_crush_divisor` (V run, section Mutants): a hireling given the player share (v10), a superunique hireling given 8 (v11), and two in `Monster::target`, which hands the divisor to `player_blow` (v15: a constant 4; v16: the parent's 8 back for a unique). v15 and v16 also pass r1's ORACLE (0 differ): its d2d side calls `Monster::crush_divisor` directly, not the divisor the player's blow takes, so the measured quantity is one step short of the player-visible one (the decomp37 / decomp38 lesson). No test in the tree reads `Monster::target(...).crush_divisor` (V read, `git grep`). r2 measures through `Monster::target` in the oracle commit (same measuring code in both commits) and adds two game.exe pins (a hireling at 4 players, a hireling made superunique): 17 of 17 non-equivalent mutants abort.
3. **The writer scan overclaims (medium, claims).** "the only two writes of bit 2 in the executable" and "only the two superunique branches write bit 2": a widened scan finds a third writer of the word, `FUN_004ac810` (a setter with a mask, three calls, one with a mask read from memory), and shows that `0x5a4599` is not a creator's superunique branch but `FUN_005a4440` re-making a monster from a saved record, gated on a record flag the saver sets from bit 2 alone (section "Who writes bit 2"). The conclusion stands (a plain unique never gets bit 2 on the game side); the wording must change. r2's `drops.md`, the oracle docstring and the claim message say it.
4. **Stale "boss or unique" comments (low).** `tests/test_combat.cpp:423` and `tools/emu/crush.py:18` (both from #609) still say a unique divides by 8; r2's claim corrects both.

**Post `fid/crush-d5-r2`**: oracle `3d51e67e`, claim `5824a42f`, on upstream main `d5de50ca` (pushed to the private fork remote today, V run `git ls-remote`); author and committer `shatteredeagle <26069301+shatteredeagle@users.noreply.github.com>`, oracle first, `Agent: Claude Code` the only parsed trailer after a blank line in both (V run, `git interpret-trailers --parse`). Against r1 rebased: `ai.hpp` byte for byte; the oracle's d2d side through `Monster::target`; pins 13 and 14; `drops.md` writer paragraph; two comments. The body then needs C1 to C8 below.

## Game side, re-derived (V read by this verifier)

- **`FUN_005bffc0` `0x5c0031..0x5c00a8`.** A type-0 defender (player): 10, straight to the missile test (`0x5c003c`). Type 1: `FUN_0063ee90` (a hireling) 10, no player share (`0x5c0050`); else `ebx = 4`, then 8 when `FUN_0063e9f0(0, unit)` is not 0 (MonStats boss row) or, only when it is 0, `FUN_005a0180(unit, 2)` (`lea edx, [eax+2]` with eax 0 at `0x5c0063`); then `+ MulDiv(ebx, FUN_005738f0(max(stat 0x64, 1)), 100)`; doubled for event 6 (`0x5c00a8`). Other types 4.
- **`FUN_005a0180` (`0x5a0180..0x5a01a1`).** Type-1 unit, data `[unit+0x14]`, `(word [data+0x16] & dx) != 0`.
- **The mask-2 bit is the superunique bit, from three readers.** `FUN_005a03a0` reads the SuperUniques index only under it (S: `verifier-crush3`); the S->C monster packet writer `0x53e4bf..0x53e58d` packs `FUN_005a0180` with 4, 8, 2, 0x10, 0x40 and, under 2 only, a 16-bit `FUN_005a01e0(unit, 0x10)` (the SuperUniques index); the saver `0x542305..0x54231c` sets record flag 0x800 and copies `FUN_005a01e0` into the record's `+0x3c` under 2 only (V read).

### Who writes bit 2 (the brief's (b))

The author's scan matched `or|mov|and|xor|add|sub` of a **word or byte** at `[reg + 0x16]`. This verifier widened it (`scan.py`, capstone over every executable section; only `.text` is executable): every instruction whose first operand is memory and whose byte range covers displacement `0x16` (byte, word or **dword at `+0x13..+0x16`**), any base or index, `mov/or/xor/add/adc/bts/xchg/inc/shl/rol/set*/movs/stos/cmpxchg` and the rest, plus every `lea reg, [reg+0x16]`: 1,286 hits. Writes of an immediate without the bit dropped: 1,146. Then the base register's last definition traced in a linear sweep (`scan2.py`): 24 writes whose base came from `mov reg, [x + 0x14]` (a unit's data pointer). V read of each:

| Site | What it is | Bit 2? |
|---|---|---|
| `0x5a4a99` | `FUN_005a49b0`, a monster made from a SuperUniques row (spawned-bit table `+0x1d30` beside it) | yes |
| `0x5a4599` | `FUN_005a4440`, `or ..., 2` only when its argument `[ebp+0x20]` is set; its one caller `0x5426c3` passes `record[+0x18] & 0x800` (and `& 2` as the champion argument): a monster made again from a saved record | yes, only for a record the saver flagged from bit 2 (`0x542305..0x54230e`) |
| `0x5a01c9` | `FUN_005a01b0`, setter with a mask | no caller, no pointer to it in any section (V run) |
| `0x5a0375` | `FUN_005a0320`, `dx = 8` fixed at `0x5a034a` (callers `0x5a0b61`, `0x5a487b`, `0x5a48f4`, `0x5a496e`) | no |
| `0x5a4967` | `FUN_005a4940` (unique), `or ..., cx` with `cx` = the unit type, 1 by the test at `0x5a494c` | no |
| `0x5a4905`, `0x5a44a3`, `0x5a083e`, `0x5a0b76` | champion `5`, champion `4`, `4`, `1` | no |
| `0x4ac829` | **`FUN_004ac810`, a second setter with a mask, missed by the author's list**: calls `0x45e465` (4), `0x45e473` (8), `0x4664bd` (a word read from `[edi+0x10]`) | possible for the unit it is given; its callers sit in the client's address range (I) and the game-side handler never reads a client unit (I) |
| `0x62818f`, `0x64a1d9`, `0x64acae` | type-4 (item) / type-3 (missile) data, guarded by `cmp [unit], 4` / `3` | not a monster |
| `0x66ac18` | `or dword [x+0x14], 8`: bit 3 of byte `+0x14` | no |
| the other 10 | list-link writes (`[x+0x14]` / `[x+0x18]` pairs: `0x4bc762`, `0x50eb87`, `0x575d64`, `0x5ff24b`, `0x5ff4d6`, `0x5ff4e1`) and a pointer that is a stack argument, not a unit (`[ebp+0x14]` / `[esp+0x14]`: `0x62af32`, `0x663941`, `0x66394c`, `0x6b089f`) | not the monster data (V read) |

The rest of the 1,146 have bases not traced to a `+0x14` load (stack frames, `alloc` results, table rows): not checked one by one (U), as the author's. The SuperUniques.txt and Levels tables only feed `FUN_005a49b0` (which row), never the word itself: no table value is copied into `+0x16` (no `movs`/`rep` or dword write at `+0x14..+0x16` traced to a data pointer). Result: on the game side a plain unique never carries bit 2 (V read), with the client setter as the one caveat the body must not deny.

### Which d2d kind is which (the brief's (a))

- d2d sets `Boss::superunique` only for a SuperUniques row: the preset path (`objgroups.cpp:718`, `FUN_005a49b0`'s) and `World::seal_boss` (`act4.cpp:62`: Diablo's seals, the Ancients through `act5.cpp:490`, Baal's waves through `act5.cpp:525`) (V read). game.exe makes all of these through `FUN_005a49b0` (I, the d2d comments name it; the seal path `FUN_0054e600` not re-read).
- `Boss::unique` comes from the random unique pack (`objgroups.cpp:704`, `833`), the Gidbinn guardian (`act3.cpp:234`, game.exe `FUN_005a43e0`, which writes no flag), and the act-boss rows made without a pack (`ai.cpp:219`, `act4.cpp:335`: `FUN_005b1cf0`, "flag 8"): none of them gets bit 2 in game.exe (the writer list above). With the claim they take 4 unless their row is a MonStats boss row (Andariel, Duriel, Mephisto, Diablo, Baal keep 8 through `boss_column`). So Blood Raven, the Maggot Queen rows, the compelling orb and Nihlathak's non-superunique row change from 8 to 4 too (I from the table; each is a row of the oracle's 734, which agrees).
- Champion, minion, plain: 4 on both sides before and after; the oracle's other kinds agree in all 23,728 cases (V run below). No other row of d2d's table is wrong: no further fix belongs in the PR.

## Rebase (the brief's (c))

r1 sits on `9b64c11a`; upstream main is `d5de50ca` (83 commits on; V run `git fetch`). Since then nothing in upstream touched `crush_divisor`, `Monster::target`'s crush argument, `player_blow`'s crush (#609, #615) or the leech (#587); the two commits rebase with no conflict (V run). The crush base #609 / #615 is unchanged; this PR changes only the divisor the target carries into it.

## Differential (V run)

Builder container, the r2 tree. `crush_div.py gen 8` (29,360 cases), `game` once (game.exe natively, every table loaded), the d2d side from each tree's `test_crush_divisor --crush`, `compare`:

| Tree | plain | champion | unique | superunique | minion | total |
|---|---|---|---|---|---|---|
| parent (r2 oracle commit `3d51e67e`, built alone) | 0 | 0 | **5,632** | 0 | 0 | 5,632 of 29,360, exit 1 |
| r2 claim `5824a42f` | 0 | 0 | 0 | 0 | 0 | 0 of 29,360, exit 0 |

The parent's first difference: row 0 unique at 8 players, game.exe 3,728,270, d2d 1,864,135 (a quarter plus 350 percent against an eighth plus 350 percent). This reproduces the author's counts exactly; the grid is exhaustive (every row, kind and player count), so a fresh seed does not apply. `compare` on hand-made files (V run, the r2 script): equal files exit 0; an empty d2d file, a short one, a malformed line, a missing file, an empty case file, no argument, an unknown verb and one changed value each exit 1.

## Mutants (V run, builder, local, never pushed)

Each one text change in `components/game/ai.hpp` of the r2 claim; the test's translation unit recompiled with the build's own command and linked with the unmutated objects (`mutb.py`, the author's method); the pins run alone (as ctest does) and the oracle compare on all 29,360 cases. The r1 column is the same mutant against r1's `test_crush_divisor.cpp` (its 12 pins and its d2d side).

| Mutant | r2 pins | r2 oracle (of 29,360) | r1 pins | r1 oracle | `test_game_tables` |
|---|---|---|---|---|---|
| v01 a unique counts again (the parent) | abort | 5,632 | abort | 5,632 | abort |
| v02 unique and superunique swapped | abort | 11,264 | abort | 11,264 | abort |
| v03 a champion gets 8 | abort | 5,632 | abort | 5,632 | abort |
| v04 a minion gets 8 | abort | 5,632 | abort | 5,632 | passes |
| v05 the superunique no longer gets 8 | abort | 5,632 | abort | 5,632 | abort |
| v06 the MonStats boss column dropped | abort | 800 | abort | 800 | abort |
| v07 the base 4 becomes 2 | abort | 22,528 | abort | 22,528 | abort |
| v08 the 8 becomes 16 | abort | 6,632 | abort | 6,632 | abort |
| v09 the 8 becomes 7 | abort | 6,632 | abort | 6,632 | abort |
| v10 a hireling takes the player share | abort | 175 | **passes** | 175 | abort |
| v11 a superunique hireling gets 8 | abort | 40 | **passes** | 40 | passes |
| v12 the share always on a base of 4 | abort | 5,803 | abort | 5,803 | abort |
| v13 the share read one player higher | abort | 25,515 | abort | 25,515 | abort |
| v14 a superunique without the share | abort | 5,103 | abort | 5,103 | abort |
| v15 `Monster::target` hands `player_blow` a constant 4 | abort | 26,544 | **passes** | **0** | passes |
| v16 `Monster::target` gives a unique 8 again | abort | 4,968 | **passes** | **0** | passes |
| v17 the share rounded up | passes | 0 | passes | 0 | passes |
| v18 the hireling branch dropped | abort | 180 | abort | 180 | passes |

17 of 17 non-equivalent mutants abort r2's pins. v17 is equivalent: the shares are multiples of 50 and the divisor 4 or 8, so `divisor x pct` always divides by 100. r1: four pass its pins, and v15 / v16 also pass its oracle (its d2d side never went through `Monster::target`). `test_game_tables` alone catches 12 (it tests `monster_foe_props`, the MonProp path).

## Build, ctest, tidy (V run)

Builder container, Release, GCC, tests and tools on, `-j2` under the builder's lane lock, ccache; every job registered with `tools/agentrun.sh`; `D2_MPQ_DIR` and `D2_PATCH_INSTALLER` set.

| Tree | Build | Tests |
|---|---|---|
| r2 claim `5824a42f` | full, rc 0 (3,078 s) | `ctest -j2`: 407 of 408 pass, 13 skip for data (as upstream); the failure is `smoke_d2d` |
| upstream main `d5de50ca` (the app rebuilt in the same tree) | `d2d`, rc 0 | `smoke_d2d`: @@SMOKEP@@ |

`smoke_d2d` on the claim aborts at `tests/smoke_d2d.py:444` (`assert st["class"] == "3" and st["name"] == "Alpha" and st["hardcore"] == "1"`, after switching to the character Beta): a character-select / save check that no crush code reaches.

clang-tidy-19 on the r2 claim (`test_crush_divisor.cpp`, `test_game_tables.cpp`, `test_combat.cpp`; the repo's `HeaderFilterRegex` takes `components/`, so `ai.hpp` too): @@TIDY@@

## Hygiene (V run)

- `pr_guard.py upstream/main HEAD` on r2: 0 FAIL, 0 WARN. `presubmit.py --quick --base upstream/main`: 0 FAIL, 4 WARN (doc-claims on `drops.md:690` "every MonStats row", true: the oracle runs every row; doc-claims on `tools/emu/README.md:297`, the same; build and gcc skipped by `--quick`), answered by C6.
- Identity and trailers as above. Diff r2 against upstream: 8 files (`ai.hpp`, `test_crush_divisor.cpp`, `test_game_tables.cpp`, `test_combat.cpp`, `crush_div.py`, `crush.py`, `tools/emu/README.md`, `drops.md`, `monsters.md`); no private code, host names or paths (read).
- No `todo.md` / `unverified.md` row names this (V read, `git grep crush`), so the template's "row moved" part has nothing to move.

## Required changes to the body (`crush-d5-body.md`, becomes the r2 body)

- **C1 (high).** Commits: "Two commits on upstream `d5de50ca`: the oracle (`3d51e67e`), then the claim (`5824a42f`)".
- **C2 (high, claims).** Evidence rows `0x5a4599, 0x5a4a99`: replace with two rows: "`0x5a4a99` (`FUN_005a49b0`): `or word ptr [eax + 0x16], 2` for a monster made from a SuperUniques row, beside the spawned-bit table at `+0x1d30`" and "`0x5a4599` (`FUN_005a4440`): the same `or`, only when its argument is set; its one caller (`0x5426c3`) passes a saved record's flag 0x800, which the saver (`0x542305..0x54230e`) sets from bit 2 alone, so a monster made again keeps the bit it had". Add a row "`FUN_004ac810` (`0x4ac829`): a second setter with a mask; calls `0x45e465` (4), `0x45e473` (8), `0x4664bd` (a word from a record); its callers sit in the client's address range". Keep the `FUN_005a01b0` row (no caller, no pointer to it). What changes: "only the two superunique branches of the creators write it" becomes "on the game side only a monster made from a SuperUniques row, or made again from a saved record that had it, gets it, so a unique ... takes the quarter".
- **C3 (high, template).** Add the template's Checklist: a test in `tests/` (`test_crush_divisor`, `test_game_tables`); the finding in `docs/research/re/` (`drops.md`, `monsters.md`); `pr_guard` passes; "`ctest` with the game data: 407 passed, 1 failed, 13 skipped; fails on main too: `smoke_d2d`" (if the parent row above says so); labels `oracle`, `test`, `trace`.
- **C4 (medium).** Counted differentials: the d2d side is "the divisor `Monster::target` hands `player_blow`", not "`Monster::crush_divisor`". Mutants: "18 mutants (this file's table); 17 abort `test_crush_divisor`; the one that does not (the share rounded up) gives the same number for every input". Drop "Both test binaries print ok; the full ctest has not run".
- **C5 (medium).** What changes: "`test_crush_divisor` pins game.exe's crushes for fourteen rows, kinds and player counts, through `Monster::target` (a hireling at four players, a hireling made superunique)"; and "two comments that still said a unique divides by 8 (`test_combat.cpp`, `crush.py`)".
- **C6 (medium).** Limits: replace the first bullet with "The writers were found by a scan of every write that covers byte `+0x16` of a unit's data word (`.text` is the one executable section), the base traced to the unit's data pointer; writes through a pointer built another way were not traced one by one. `FUN_004ac810` writes the word for the units it is given; its callers are in the client's range and the game side's handler never reads them (not traced further)." Keep the bit-8 and D6 bullets.
- **C7 (low).** In play: add "Blood Raven, the Maggot Queen rows, the compelling orb and other act-quest monsters d2d makes as a unique without a pack take a quarter too unless their row is a MonStats boss row (as `FUN_005b1cf0`'s flag 8 gives them in game.exe)".
- **C8 (low).** How to re-check: add `ctest --test-dir build -R 'test_crush_divisor|test_game_tables|test_combat' --output-on-failure`, and the parent count "parent: 5,632 differ (all unique), exit 1".

## Not checked

- No Windows, macOS or Clang build; no live game, no lab drill; nothing on the worker host.
- The writes not traced to a data pointer (U); `FUN_004ac810`'s callers (I: client); the seal path `FUN_0054e600` calling `FUN_005a49b0` (I, d2d's comment); a full ctest of the parent tree (only `smoke_d2d` was re-run there).
