# Verifier report: "a monster tries a player's newest summon first, its merc last" (fidbatch)

> status: verifier report 2026-11-23 (independent Opus verifier, docs only, nothing pushed or posted). Subject: [pr-fidbatch-pet-order-draft.md](pr-fidbatch-pet-order-draft.md); fork branches `didle/fidbatch-pet-search-order-up` @ `21df4aa6` (one commit on upstream main `ff6d47a66`) and `didle/fidbatch-20261124` @ `678ef868`. Tags: V = run or read by this verifier (how), S = read in a doc, U = unverified.

## Verdict

**NOT READY** as committed (`21df4aa6`). Two blockers, both fixable in the same branch:

1. BLOCKER (V): the merc is put LAST unconditionally, but game.exe re-inserts it at the front of the player's list on a revive (FUN_00579c00 -> FUN_00579aa0 -> FUN_005b1900 at 0x579b43) and on a hire. With summons alive from before the revive, the commit moves d2d AWAY from game.exe: in the verifier's revive cases it differs on 766 picks where upstream main differs on 515 (section 2). The PR standard asks that the change push towards 1:1; this one trades one deviation for another on a path a Necromancer or Druid player hits every time the merc dies and is revived in town with summons following.
2. BLOCKER (V): CI's "Names and includes" clang-tidy job fails on the new test (`misc-include-cleaner`: `d2d::d2s::kLevel`, `std::pair`, `std::move` without their headers; section 3). Fix: include the d2s stats header that declares `kLevel` and `<utility>`.

READY AFTER EDITS once (1) is fixed one of two ways (section 6) and (2) is fixed, the oracle (both case sets) rerun to 0, and the high findings below answered.

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

## 5. ctest and det state hashes

- Whole tree at `21df4aa6` (GCC Release, `D2D_WERROR=OFF`, `-j4`, builder long lane, 1,406 s), then `ctest --test-dir build -j4 --output-on-failure`: **506 tests, 0 failed**, 13 skipped (they want data this builder lacks: test_armor_hit, test_boss_level, test_cold_length, test_freeze_resume, test_freeze_rules, test_hammer_hit, test_light_pierce, test_missile_hit, test_next_hit, test_sk_resist, test_static_field, test_walk_velocity, test_umod_stats). The build has 28 warnings, none in the changed files. V. Not run: CI's RelWithDebInfo and `D2D_WERROR=ON` whole-tree build (the changed TUs pass `-Werror` above).
- Det state hashes (`tools/emu/det_scenarios_lab.py` / `det_scenarios_run.py` of d2lod-bot, all 14 scenarios, seed 3, lockstep, headless d2d on the builder, not the worker host): lab d2d of the integration branch without the claim (`29e28671a`, d2d sha256 prefix `8866230b4ac3`) and with it (`678ef868`, `9bc51ecdbd36`). Every scenario: `mon` and `state` hashes identical; `merc`, `pet`, `follow` rerun on the base give the same hashes (deterministic). V. Two changes to the driver were needed for this fork: d2d now wants `--lab-verbs` (and `--no-save`) for the debug verbs (the first attempt without them ran every scenario empty and is discarded); `debug monstate` does not exist in this lab set (its line is an `err` in every snapshot, same on both), and `debug skill left 44 / 42` answers "not usable" in the hit scenarios (same on both). What it shows: nothing outside the claim moved. What it does not show: none of these scenarios has two pets at the same distance from a monster's think, so the claimed change itself is not exercised by them.

## 6. Findings, ranked, with the fix

| # | Rank | Finding | Evidence | Recommendation |
|---|---|---|---|---|
| 1 | blocker | Merc always last; game.exe puts it first after a revive or a later hire. The commit is worse than upstream main in that state | V: disassembly (section 1), revive oracle 766 vs 515 | Preferred: model game.exe's list order. A per-hero insertion counter: each `Pet` gets `listed = ++actor->list_serial` where summons.cpp pushes it; the merc gets one wherever d2d creates it on hire or join (one candidate: world/setup.cpp:500, not checked further: U) and on `cmd::ResurrectMerc` (world/commands.cpp:249); `search_order` sorts the pets after the player by serial, newest first. Then the oracle and the revive variant must both give 0 (the verifier's `pets_merc_revive.py` with REVIVE_SHARE 0.5 is a ready test). Fallback with no state: reverse only the summons and leave the merc where d2d has it (first): measured with the same harness (slot 1 kept in place) it differs in 706 of the author's cases (old 1,254), 766 of the verifier's plain cases (old 1,281) and 0 of the revive cases (old 515): never worse than upstream main, but no longer "0 after", and the claim narrows to summons |
| 2 | blocker | clang-tidy include-cleaner FAIL in `tests/game/test_pet_search_order.cpp:48,61,68` | V: CI's selector + clang-tidy-19 on the builder | Add the missing includes; rerun tidy on the test TU |
| 3 | high | README row command does not paste: it runs `pets_merc.py` from tools/emu but `-Icomponents`, `find components` and `build/components/rules/librules.a` from the repo root | V read (tools/emu/README.md:294) | Write it as the draft's Reproduce block does (`../../components`, `../../build/...`), all from tools/emu |
| 4 | high | The harness's `after` assigns kinds by slot (1 = merc, rest summons), not by ai.cpp's real mapping (`foe.id < 0`); so the 0 proves `search_order`, not the call site | V read | Say so in the PR body (the in-game test is the call-site proof), or have the harness read the kind from the case line |
| 5 | medium | Co-op: d2d's foes are `p1, m1, p2, m2, pets...`; `search_order` keeps `p1, m1, p2` and moves `m2` behind both players' summons, which game.exe never does (each player has its own list) | V read (fight/tick.cpp:487..502, monsters.cpp search_order) | Limit already listed; add that the co-op order changes and is unproven, or make `search_order` work per side (`Side::at`/`pets`) |
| 6 | medium | monsters.hpp comment says anything not a merc or summon "keeps its place"; `other` entries after the last player are moved to the end | V read | Reword: "monsters (lists 8/9, a converted one) go after the pets, as in d2d's foes" |
| 7 | low | Harness header: stale line refs (`ai.cpp:546..563 at 72132b99`, `merc.cpp:302..330`, `ai.cpp:558/559`), `-std=c++20` vs README c++23 | V read | Refresh to the branch's lines |
| 8 | low | Converted monster: game.exe inserts it with FUN_005b1900 (Conversion, 0x5d0501), so it is the newest pet; d2d keeps it after every pet. Unchanged by the commit | V xref | Keep as a limit; with the serial fix it can take a serial too (a later claim) |
| 9 | low | Draft calls the 1,254 breakdown I | re-tabulated here: 780 merc vs summon, 474 summon vs summon, all same distance | Tag it V |

## 7. Claim wording, title and body

If finding 1 is fixed by the insertion counter (recommended):

- Title: `fight: a monster tries a player's pets newest first, the merc by when it was hired or revived (FUN_005b1900)`
- Claim: "FUN_005b1900 inserts each pet of a player right after the player's list head (0x5b1966..0x5b1979), and FUN_005dd7f0 walks the list from the head taking only a strictly nearer candidate (0x5dd960), so on a tie the newest pet wins. The merc is inserted when hired and again when revived (FUN_00579aa0 at 0x579b43). d2d tried the merc first and the summons oldest first; it now tries the pets newest first by when each entered the list."
- Body: keep the Evidence paragraph, add the revive oracle (`REVIVE_SHARE`) numbers, the tidy and `-Wshadow` lines, the pre-existing `test_structs.py` FAIL ("pre-existing on main: structs.tsv cites ai.hpp:509, the file has 371 lines; not touched here"), and the Limits: one player; a converted monster stays after the pets; hooked functions as listed.

If the fallback is taken (summons only):

- Title: `fight: a monster tries a player's summons newest first (FUN_005b1900)`
- Claim: "...so of two summons at the same distance the newer is chosen. The merc keeps its place (first): game.exe's merc position depends on when it was last hired or revived, which d2d does not track (Limits)." Numbers: before / after per case set as in finding 1; the "0 after" line goes, replaced by "every remaining difference is a merc-vs-summon tie where the merc is older than the summon".

Do not keep the current title: "its merc last" is false after any revive (V).
