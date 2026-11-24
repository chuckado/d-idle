# PR body (local, NOT posted): a monster tries a player's newest pet first

> status: ready for the coordinator's decision 2026-11-24 (agent fidbatch, after the second independent verification: [pr-fidbatch-pet-order-verifier2.md](pr-fidbatch-pet-order-verifier2.md), READY AFTER EDITS; this file applies its edits and the upstream AGENTS.md rules of 2026-11-18..20). Nothing is pushed to GitHub. Fill the numbers from the "Result block" at the end before posting; labels `trace` and `port` need the user's go (ask first). Background and the earlier drafts: [pr-fidbatch-pet-order-draft.md](pr-fidbatch-pet-order-draft.md).

Title: `fight: a monster tries a player's newest pet first, the merc by its last hire or revive`

Branch: private remote `didle/fidbatch-pet-search-order-up-v5` (one commit, on upstream main; see the result block for the hash).

---

## What changes

A hostile monster picks its target by walking the player's list and taking a candidate only when it is strictly nearer than the best so far, so on a tie the first one tried wins. game.exe builds that list so the pet that entered it last comes first: a summon enters where it is made, the merc when it is hired or joins and again when it is revived. d2d tried the merc first, then the summons oldest first. Each pet now carries when it entered the list and the search takes them in game.exe's order.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `0x5b1956`..`0x5b197c` (FUN_005b1900) | a new list node goes right after the player's list head: head, newest, .., oldest |
| `0x5dd960` (FUN_005dd7f0) | `cmp edi,[best]; jae skip`: a candidate is taken only when strictly nearer; the first tried wins a tie |
| `0x579b43` (FUN_00579aa0, only caller FUN_00579c00) | the revive handler inserts the merc again with FUN_005b1900: it goes to the front |
| `0x5732cb` (FUN_00573270, from FUN_005770e0, FUN_005774f0, FUN_00579180) | the merc maker inserts the merc into the list when it is made (its 5th argument nonzero skips it) |
| `0x5a6586` (FUN_005a6520 -> FUN_005b1a90, `0x5b1a90`..`0x5b1b6c`) | a dying unit is unlinked: the others keep their order |
| `0x5d0501` (FUN_005d0350, Conversion) | a converted monster is inserted with FUN_005b1900 too (not changed here, see Limits) |

## Claims (a batch only)

One commit, one claim.

| Commit | Claim or change | Evidence | Re-check |
|---|---|---|---|
| `fight: a monster tries a player's newest pet first, the merc by its last hire or revive` | the pet that entered the player's list last is tried first; a summon's stamp is set where it is made, the merc's in `World::spawn_merc` (hire, join, revive); `rules::search_order` orders the foes | `0x5b1956`, `0x5dd960`, `0x579b43` | the oracle and `test_pet_search_order` below |

## How to re-check

The oracle `tools/emu/pets_merc.py` builds the player's list with game.exe's own FUN_005b1880 and FUN_005b1900 (and, in a share of the cases, inserts the merc after the summons as a revive or a later hire does), then runs FUN_005dd7f0 natively. `pets_merc_harness.cpp` runs d2d's real `rules::search_order` and `rules::search_pick` on the same cases. From `tools/emu`, game.exe as `$D2_MPQ_DIR/bin/game.exe`, the library built (`build/components/rules/librules.a`):

```sh
for rs in 0 0.3 1.0; do
  for s in 21 22 23 24; do REVIVE_SHARE=$rs uv run python pets_merc.py 25000 $s c$s.txt > g$s.txt; done
  cat c2?.txt > c.txt; cat g2?.txt > g.txt
  ./pm_h before < c.txt | diff g.txt - | grep -c '^>'    # the old order's misses
  ./pm_h after  < c.txt | diff g.txt -                   # nothing
done
```

(`pm_h`: built as the header of `pets_merc_harness.cpp` says.) The in-game test drives the real `Fight::world`, `Fight::summon_one` and `World::spawn_merc` (no stamp is set by hand): `build/tests/test_pet_search_order`. It compiles on the parent and fails there for the right reason (a tie goes to the older summon or the merc).

```sh
python3 tools/pr_guard.py
cmake --build build && ctest --test-dir build --output-on-failure
```

Numbers of the runs of 2026-11-24 (the commands above print them): see the Result block below, copied in when posting.

## Checklist

- [x] A test in `tests/` (`test_pet_search_order`)
- [x] The finding in `docs/research/re/monster-ai.md` (the target search section)
- [x] `python3 tools/pr_guard.py` passes
- [ ] `ctest` with the game data: N passed, N failed, N skipped (the Result block; "not run" for what needs data the author's lab image lacks, named there); fails on main too: `tools/emu/test_structs.py` (presubmit tool test, a `structs.tsv` row cites a line past the end of `ai.hpp`)
- [ ] Labels: `trace`, `port` (the coordinator asks first)

### Limits and notes for the reviewer

- One player only: with several players in the foes the order is unchanged (co-op is not modelled; d2d lists every pet after the last player, a separate quirk). presubmit's "net roles" warning: the list order is the server's pick; this changes the host's search, a client only sees the result.
- A converted monster (Conversion) is the caster's newest pet in game.exe; here it stays after the pets as before. `// ponytail:` note above `search_order`.
- Merc re-inserts on other paths: FUN_005394a0 (the per-client join walk, which recreates a joining client's own saved merc; S: `docs/research/re/hireling.md`, `net-join.md`) and FUN_0056aa50 (the loader's mercenary reader inside FUN_0056b180; S: `fid-loadhp-trace.md`) both call the merc maker on a join or a load, where d2d stamps in `spawn_merc` too. Whether FUN_005394a0 also runs on an act change (town-start.md lists "game join / act entry") is U; d2d does not re-stamp on a level change either.
- The oracle's revive and the verifier's timeline runs (hire, summon, merc death with FUN_005b1a90 run natively, revive, summon death, in any order) agree after the change; the harness tags slot 0 player and the rest pet, and ai.cpp's own tagging is covered by the in-game test.
- Dead pets are not generated (game.exe unlinks a dying unit); a dead merc stays as a dead foe entry and never wins.
- clang-tidy: the PR touches `fight.hpp` and `ai/units.hpp`, so CI's PR mode reaches four `std::erase_if` include-cleaner findings that are already on main and outside this diff (`apps/d2d/devctl/debug_hero.cpp`, `tests/game/test_stand_attack.cpp`, `tests/game/test_two_hands_blow.cpp`); none are in a changed file.
- Lab fork only (not in this PR): the lab verbs `merc level` / `merc heal` / `merc hire` call `spawn_merc` and so restamp the merc.
- The det scenes have no tie, so their state hashes do not exercise the claim; they were identical with and without it (verifier's run of the earlier version).

---

Agent: Claude Code

## Result block (runs of 2026-11-24, builder long lane, GCC Release, `D2D_WERROR=OFF`; game data of the lab image)

- Commit `87d9d7428` (private remote branch `didle/fidbatch-pet-search-order-up-v5`, author the user's numeric noreply, one commit on upstream main `dadda9b98`). The ctest run below used the same tree before two edits that touch no code path (the harness include line and the amend's author; the test's `level.hpp` include came first), and the oracle and tidy runs used the final tree.
- `ctest --test-dir build -j4`: 506 tests, 505 passed, 1 failed, 13 skipped. Skipped for data the lab image lacks: `test_armor_hit`, `test_boss_level`, `test_cold_length`, `test_freeze_resume`, `test_freeze_rules`, `test_hammer_hit`, `test_light_pierce`, `test_missile_hit`, `test_next_hit`, `test_sk_resist`, `test_static_field`, `test_walk_velocity`, `test_umod_stats` ("not run" for those). The failure, `test_two_hands_blow` (`tests/game/test_two_hands_blow.cpp:58`, `phys_hi == 100`), fails the same way on upstream main `dadda9b98` (built and run on its own).
- Oracle (`REVIVE_SHARE` 0 / 0.3 / 1.0, seeds 21..24, 25,000 cases each): old order misses 1,212 / 1,006 / 527 S cases, the new order 0 / 0 / 0. md5 of `g.txt` / `c.txt`: d24291ff / 250eb5c1, b9566f9d / 36a3277d, 782e4567 / 95f579f4.
- `test_pet_search_order` on the commit: OK. On the parent (upstream main plus only the test file, it compiles): 3 of its 9 checks fail (foe 2 for 3, foe 2 for 4, foe 1 for 2), the ties between summons and merc against summon.
- Second verifier (earlier version `02368c52`): own timeline oracle with hire, summon, merc death (FUN_005b1a90 natively), revive, summon death in any order: 664 misses before, 0 after; det hashes of 14 scenes identical with and without (no scene holds a tie).
- clang-tidy-19 (`.clang-tidy`, the build's compile database) on the 7 changed `.cpp` files: no warning. `-Wshadow -Wconversion` is in the build flags: no warning in the changed files. `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN. `presubmit.py --quick`: 2 FAIL (`tools/emu/test_structs.py`, fails on main too), 2 WARN answered above (siblings).
- Integration branch for the lab: `didle/fidbatch-20261124c` @ `8195a62f` (the same claim, v3 text, on `didle/rebase-20261124`); not rebuilt after the pick.
- Not done: independent verification of this exact commit (the second verifier's edits were applied after its run); det hashes for v5.
