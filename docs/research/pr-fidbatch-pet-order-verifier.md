# Verifier report: "a monster tries a player's newest summon first, its merc last" (fidbatch)

> status: verifier report 2026-11-23 (independent Opus verifier, docs only, nothing pushed or posted). Subject: [pr-fidbatch-pet-order-draft.md](pr-fidbatch-pet-order-draft.md); fork branches `didle/fidbatch-pet-search-order-up` @ `21df4aa6` (one commit on upstream main `ff6d47a66`) and `didle/fidbatch-20261124` @ `678ef868`. Tags: V = run or read by this verifier (how), S = read in a doc, U = unverified.

## Verdict

PENDING (filled below once the builder jobs finish).

## 1. The claim re-derived from game.exe (V: disassembly read with capstone on the builder's game.exe, `dis.py` / `xref.py` of the lab)

| Link | What the code does | Tag |
|---|---|---|
| FUN_005b1880 | makes the player's list head (a pool node at `game+0x10f8+4*i`, first free of 8, the unit's `+0xd0` = i) | V read |
| FUN_005b1900 (0x5b1956..0x5b197c) | new node: `new.prev(+0xc) = head`, `new.next(+8) = head.next`, `head.next.prev = new`, `head.next = new`, unit `+0xd0` = list. Insert right after the head: the list runs head, newest, ..., oldest | V read |
| FUN_005dd7f0 walk (0x5dd8e0..0x5dd99e) | from each head along `+8`; the player is skipped with its whole list when another act, no room, a town room or `dist >= 0x37`; a dead player scores 0x7fffffff; then each node: `cmp edi,[best]; jae skip` (0x5dd960): taken only when strictly nearer (unsigned), so the first tried wins a tie | V read |
| Callers of FUN_005b1900 | 15 sites (xref): FUN_00573270 (0x5732cb, monster pet maker incl. the merc, guarded by its 5th arg), FUN_00579aa0 (0x579b43), Raise Skeleton FUN_005c4b00, Golem FUN_005c5100, Revive FUN_005c56c0, Druid summons FUN_005c7390, Conversion FUN_005d0350 (0x5d0501), Decoy FUN_005dc000, Valkyrie FUN_005dc1e0, and FUN_005c07a0 / 005c5250 / 005c6a80 / 005c7520 / 005d6e70 / 005d7430 (not named here) | V xref; names S (repo docs) |
| Merc re-insert on revive | the revive handler FUN_00579c00 (S: merc-revive docs) calls FUN_00579aa0 (0x579cf0), which calls FUN_005b1900(merc, player's list) at 0x579b43: a revived merc goes to the FRONT (newest) | V read |
| Merc leaves the list on death | FUN_005a6520 -> FUN_005b1a90 (0x5a6586) | V xref; S (trace) |
| Merc hire | FUN_00573270 is called from FUN_005770e0, FUN_005774f0 and FUN_00579180 (the free Rogue, S); which of the first two is hire and which load/join is not named | V xref, U names |
| Converted monster | Conversion (FUN_005d0350) inserts the converted monster with FUN_005b1900 at 0x5d0501: it is the newest pet of the caster | V xref |

So the claim "newest summon first" is exact (V). "Its merc last" is true only while every living summon was made after the merc's last insertion (hire, join, revive). After a revive (or a hire) with summons alive, game.exe tries the merc FIRST among those older summons.

## 2. The oracle and the in-game test, rerun (V: builder long lane, own trees `/data/src/vfp-*`, own build dirs `/data/build/vfp-*`, no shared binary replaced)

Builds: `ff6d47a66` plus only `tests/game/test_pet_search_order.cpp` (the "parent"), and `21df4aa6`; GCC Release, `D2D_WERROR=OFF` (as the author), game data of the lab image.

| Check | Result |
|---|---|
| `test_pet_search_order` on the parent | 3 FAIL lines (took 2 / 2 / 1, game.exe's order 3 / 4 / 2), rc 1 |
| `test_pet_search_order` on `21df4aa6` | `OK: ...`, rc 0 |
| `pets_merc.py` seeds 21..24 x 25,000 (S and M lines, 200,000) | md5 prefixes cases `b4a2be90`, game `168c5cf7`: equal to the draft's |
| `pm_h before` / `pm_h after` against the game side | 1,254 / 0 differ |
| breakdown of the 1,254 (re-tabulated here, no longer I) | 780: game.exe a summon, d2d the merc, same distance; 474: game.exe a newer summon, d2d an older one, same distance; none with the player or a distance change |

Adversarial variant (`pets_merc_revive.py`, verifier's copy of the oracle with one change: in a share of the cases with a merc and at least one summon, game.exe's list is built with the summons first and the merc last, as FUN_00579c00 -> FUN_00579aa0 -> FUN_005b1900 does on a revive or a hire after the summons; d2d's case line is unchanged because d2d keeps no insertion order). Seeds 31..34 x 25,000:

| share of revived cases | revived cases | old d2d order differs | new order (`21df4aa6`) differs |
|---|---|---|---|
| 0.0 (control: insert calls deferred only) | 0 | 1,281 | 0 |
| 1.0 | 50,944 | 515 | **766** |

So in the revive (or hire-after-summon) domain the commit makes d2d WORSE than upstream main: 766 against 515 picks off game.exe. In that state game.exe tries the merc first, then the summons newest first; the old d2d (merc first, summons oldest first) was right on every merc-vs-summon tie there and the commit is wrong on all of them. V.

The harness itself: its `after` labels slot 1 `merc` and the rest `summon` by position; it never runs ai.cpp's real classification (`foe.monster ? other : !foe.pet ? player : foe.id < 0 ? merc : summon`). That mapping is covered only by the in-game test (V read: `merc_foe_id()` is `-3 - 2 * actor->id`, always negative; summons carry their monster id). Its header cites `ai.cpp:546..563 at upstream 72132b99` and `merc.cpp:302..330` (stale: the branch's lines are ai.cpp:937..953) and says `-std=c++20` where the README says c++23 (V read).

## 3. What the draft did not run

| Check | Result (V unless said) |
|---|---|
| `-Wshadow -Wconversion -Werror` (the GCC proxy warnings CMake adds by default) on the three changed TUs, compile command from `compile_commands.json`, `-fsyntax-only` | ai.cpp, monsters.cpp, test_pet_search_order.cpp: rc 0, no warning, on base and patch |
| clang-tidy as CI's PR mode (`tools/tidy-pr-select.py` over the four changed C++ files: 84 TUs, `clang-tidy-19 -p build --quiet`, the repo's `.clang-tidy`, warnings are errors) | **FAIL in the new test**: `test_pet_search_order.cpp:48` no header providing `d2d::d2s::kLevel`, `:61` none providing `std::pair`, `:68` none providing `std::move` (misc-include-cleaner). The other two errors (`apps/d2d/main.cpp:56` screenshot.hpp, `apps/d2d/video.hpp:22` libavcodec) are this builder's missing ffmpeg build (CI builds `ffmpeg_build` first), not in the diff. ai.cpp, monsters.cpp/.hpp: clean |
| `pr_guard.py upstream/main HEAD` | 0 FAIL, 0 WARN |
| `presubmit.py --quick --base upstream/main` | 2 FAIL (both `tools/emu/test_structs.py`), 4 WARN (siblings: co-op partner; doc-claims README:294; build skipped; no compile_commands) |
| ctest (whole tree) | see section 5 |
| det state hashes | see section 5 |

## 4. The presubmit FAIL is pre-existing (V)

`python3 -m unittest test_structs.py` on a clean export of upstream main `ff6d47a66`: FAILED (failures=1), `509 not less than or equal to 371 : unit 0x90 ptr events ... components/game/ai.hpp:509`. `git diff --stat ff6d47a66 21df4aa6 -- docs components/game/ai.hpp` is empty: the commit touches neither `structs.tsv` nor `ai.hpp`. Not this PR's to fix; one line in the body is enough.
