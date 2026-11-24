# Draft v2 (local, NOT posted): "fight: a monster tries a player's pets newest first, the merc by when it was hired or revived (FUN_005dd7f0)"

> status: draft v2 2026-11-24 (v1 was verified NOT READY: [pr-fidbatch-pet-order-verifier.md](pr-fidbatch-pet-order-verifier.md); v2 answers its two blockers with an insertion counter). Agent fidbatch (bundle brief: turn the open fidelity cards into proven upstream candidates). One claim. Nothing is pushed to GitHub or posted; the private fork holds the branches. Source trace: [pets-merc-trace.md](pets-merc-trace.md) (D1, 2026-11-01); this doc re-proves it on upstream main `ff6d47a66` and builds the fix.

## What the open cards turned out to be (checked 2026-11-23 against upstream main and `gh`)

| Card / item | Verdict | Evidence |
|---|---|---|
| TASK-101 fid-load-life-mana | already upstream, MERGED | psi29a/d2decomped #242 (2026-11-08) |
| TASK-125 fid-revive-refill | already upstream, MERGED | #245 (2026-11-08) |
| TASK-203 fid-load-gold-clamp | already upstream, MERGED | #539 "a saved purse or stash over its cap is 0 on entering a game" (2026-11-15) |
| TASK-102 fid-load-maxmana | NO DEVIATION in d2d (bot reader defect) | task result of 2026-11-08, `fid-loadmaxmana-trace.md`; nothing to propose |
| TASK-179 ptrclick | UNVERIFIED, no PR candidate | needs a live probe (docs/probes/ptrclick.md); a click verb artifact of the lab fork, not an upstream d2d behaviour |
| monster missile gates (can-block, MissA ToHit) | already upstream | #553 (can-block); ToHit gate: `stats.no_hit_roll = !info.to_hit` on every monster missile path (ai.cpp:3480, 3510, deaths.cpp:357, objects.cpp:1035) |
| monster mods D1 D2 G1 | already upstream | #157 (boss stats), ghostly speed skipped (`components/rules/uniques.cpp:198`), Cold Enchanted length (`uniques.hpp` elem_len) |
| pet tie order D1 | OPEN on upstream main; this draft | below |

Open PR #708 (merc RogueMissile) is not touched.

## Title and body (PR template)

Title: `fight: a monster tries a player's pets newest first, the merc by when it was hired or revived (FUN_005dd7f0)`

> **What changes.** A hostile monster's target search (FUN_005dd7f0) walks the player's list in order and takes a candidate only when it is strictly nearer than the best so far (0x5dd960), so on a tie the first one tried wins. game.exe builds that list with FUN_005b1900, which inserts each pet right after the player's list head (0x5b1966..0x5b1979): the pet that entered last is tried first. A summon enters where it is made; the merc enters when it is hired or joins and again when it is revived (FUN_00579aa0 at 0x579b43). d2d tried the player, the merc, then the summons oldest first. Each pet now carries when it entered the list (`Pet::listed`, `Hero::merc_listed`, set where summons are made and in `World::spawn_merc`, the one place a merc is hired, joined or revived) and `rules::search_order` gives game.exe's order.
> **Evidence.** `tools/emu/pets_merc.py` builds the list with game.exe's own FUN_005b1880 / FUN_005b1900 and runs FUN_005dd7f0 natively (100,000 cases per run: one player, the merc in 85 %, 0 to 3 summons, distances biased to the ties and the 0x37 / aidist / 6 / 0x13 / 0x31 edges); in a share of the cases (`REVIVE_SHARE`) the merc's insert comes after the summons, as a revive or a later hire does. `pets_merc_harness.cpp` runs d2d's real `rules::search_order` and `rules::search_pick` on the same cases: old order differs in 1,212 / 1,006 / 527 cases (revive share 0 / 0.3 / 1.0), new order in 0 of each. `tests/game/test_pet_search_order.cpp` drives the real `Fight::world`: it fails on the parent and passes with the change.

Claims table (one commit):

| Claim | Evidence | Check |
|---|---|---|
| FUN_005b1900 inserts a pet right after the player's list head, so the list runs newest .. oldest | verifier read 0x5b1956..0x5b197c; oracle builds the list by calling FUN_005b1880 then FUN_005b1900 (V run) | `pets_merc.py`, README row |
| FUN_005dd7f0 takes a candidate only when strictly nearer: the first tried wins a tie | same oracle (V run); verifier read 0x5dd960 | `./pm_h before` against `./pm_h after` |
| a revived or later-hired merc is inserted again, at the head | verifier read: FUN_00579c00 -> FUN_00579aa0 -> FUN_005b1900 at 0x579b43 (V read); oracle with the merc's call last (the removal FUN_005b1a90 is not run, I) | revive share 0.3 and 1.0 |
| d2d now orders its pets the same way | `search_order` + `test_pet_search_order` (merc first, summon first, revive, nearer-wins cases) | `test_pet_search_order` (needs the game data) |

Limits (for the PR body):
- One player. With more than one player entry in the foes `search_order` returns d2d's order unchanged (co-op was not modelled: d2d lists every pet after the last player, a separate quirk).
- A converted monster (Conversion, FUN_005d0350 at 0x5d0501) is the caster's newest pet in game.exe; d2d keeps it after the pets, as before. A later claim could give it a serial.
- The revive oracle puts the merc's insert after the summons; the removal on death (FUN_005a6520 -> FUN_005b1a90) is read, not run.
- The harness tags slot 0 player and the rest pet and takes the serials from the case line; ai.cpp's own tagging (`foe.monster`, `foe.pet`) is covered only by the in-game test for the cases it builds.
- Dead pets are not generated (game.exe takes a dying monster out of its list).
- Only the monster's search changes; the merc's own search (FUN_005ddc30: 0 of 100,000 differ) and the pets' thinks do not.
- `tools/emu/test_structs.py` fails on upstream main too (a `structs.tsv` row cites `ai.hpp:509`, the file has 371 lines); not touched here.
- Hooked in the oracle (not the claim): the room, outdoors, town, sight, in-melee, alignment, size, area, FUN_005dd510's pathfinder, the pool allocator.

## Reproduce

Fork branch `didle/fidbatch-pet-search-order-up-v2` on the private remote, commit `02368c52`, one commit on upstream main `ff6d47a66`; game.exe as `$D2_MPQ_DIR/bin/game.exe`; the library built from the commit (`build/components/rules/librules.a`):

```sh
cd tools/emu
for rs in 0 0.3 1.0; do
  for s in 21 22 23 24; do REVIVE_SHARE=$rs uv run python pets_merc.py 25000 $s c$s.txt > g$s.txt; done
  cat c2?.txt > c.txt; cat g2?.txt > g.txt
  ./pm_h before < c.txt | diff g.txt - | grep -c '^>'    # 1212 / 1006 / 527
  ./pm_h after  < c.txt | diff g.txt -                   # nothing
done
# md5 of g.txt / c.txt: share 0 d24291ff / 250eb5c1, 0.3 b9566f9d / 36a3277d, 1.0 782e4567 / 95f579f4
build/tests/test_pet_search_order                       # parent (test file only): FAIL lines; this commit: OK
```
(`pm_h`: build as the header of `pets_merc_harness.cpp` says.)

## Checks run (V run unless said; builder long lane, GCC Release, game data of the lab image)

- Oracle: above. The verifier's v1 numbers (1,254 before, 766 worse-than-main in the revive state) were reproduced by it; v2 is 0 in all three shares.
- `test_pet_search_order` (revive, merc-vs-summon, summon-vs-summon, nearer-wins cases), `test_combat`, `test_monster_ai`, `test_rules`, `test_monsters`, `test_game_rules`, `test_game`: rc 0.
- `-Wshadow -Wconversion -Werror -fsyntax-only` on ai.cpp, monsters.cpp, the test: rc 0. clang-tidy-19 (the repo's `.clang-tidy`, the compile database of the build) on the test, monsters.cpp, ai.cpp, tick.cpp, pets.cpp, summons.cpp, setup.cpp: no warning or error (the v1 include-cleaner failure is fixed: `d2s_items.hpp` and `<utility>`).
- `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN.
- ctest and det hashes: see the result block at the end of this file.

## Independent verification

v1: verifier verdict NOT READY (two blockers, both answered above). v2 has NOT been verified independently; the author does not certify it. Requested: a second verifier, with the revive oracle of the first verifier (its own `pets_merc_revive.py`), the co-op case (two players), and a check that no other path puts a merc or summon into `actor->pets` / `actor->merc` without a serial (grep found one push site, summons.cpp, and one merc site, `spawn_merc`; a pet created by another route would get serial 0 and sort last).

## Result block (v2, 2026-11-24)

- ctest on the whole tree of `02368c52` (GCC Release, `D2D_WERROR=OFF`, `-j4`, build 1,457 s): 506 tests, 0 failed, 13 skipped (they want data the lab image lacks; the same 13 as the verifier's v1 run). V run.
- Det state hashes: NOT re-run for v2. The verifier's v1 run (14 scenarios, `mon` and `state` identical with and without the claim) did not exercise the claim (no scene has two pets or a merc and a summon at one distance from a monster's think), and its driver edits for this fork were not committed. v2 adds only list serials, which change no state a scenario hashes unless a pick differs. U for v2; to redo: base `didle/rebase-20261124` (`0fcd512a`) against `didle/fidbatch-20261124b` (`42cfeaad`).
- Branches (private remote, pushed): upstream-ready `didle/fidbatch-pet-search-order-up-v2` @ `02368c52` (one commit on upstream main `ff6d47a66`; v1 `didle/fidbatch-pet-search-order-up` @ `21df4aa6` stays, superseded); integration `didle/fidbatch-20261124b` @ `42cfeaad` (two commits on `didle/rebase-20261124`; the earlier `didle/fidbatch-20261124` is superseded). The integration branch was not rebuilt after the cherry-pick (clean pick, no conflict); the upstream-ready tree was.
