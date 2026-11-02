# PR draft: monster movement pace: the base velocity percent (75) and the paces game.exe's thinks set (MONSTERPACE)

> status: DRAFT held for independent verification and the user's go (nothing pushed, opened or posted upstream). Candidate on the private fork remote only: `local/monsterpace-clean` (head `eeebb4b6`, 14 commits on upstream main `6935cf64`, read-only fetch 2026-11-02); `local/monsterpace` (head `e6a09533`, 15 commits) has the same tree plus one fix-forward commit (`58974838`: the Blood Hawk commit used the local name `far`, a windows.h macro name, and did not compile; the clean branch has it squashed into the Hawk commit). Evidence: [monsterai2-trace.md](monsterai2-trace.md) (branch agent/monsterai2 until merged), [monster-ai-trace.md](monster-ai-trace.md), [walk-speed-trace.md](walk-speed-trace.md). Format follows [pr-missilerecord-draft.md](pr-missilerecord-draft.md).

Title: `A monster starts at 75 percent velocity and its thinks carry the pace game.exe gives them (FUN_00573cb0, FUN_005de190)`

## What changes

Eight claims, one commit each (the pace claims per AI) (plus tests, tools, docs), each with its own differential against game.exe 1.14d under upstream's unicorn harness (`tools/emu`):

1. **A monster's base velocitypercent is 75, not 100.** game.exe's monster stat initialiser FUN_00573cb0 ends its stat block with `SetStat(unit, 0x43, 0x4b)` (0x573f16); the walk percent is that plus chill, pace, aura and Spider Lay, at least 25 (FUN_00623f50). d2d summed from 100, so a plain walk was 4/3 too fast, a pace of 100 gave 200 percent where game.exe gives 175. `ai.cpp` sums from `kMonsterVelocityPct` (75).
2. **Zombie** runs at pace 100 (`FUN_005de190(100, 0)` before its run, 0x5eff19).
3. **Brute** walks in at pace 100 - clamp(life %, 40, 100) (0x5efcbf).
4. **Blood Hawk**: the charge sets pace aip5 and steps = its distance, the aip2 wander -50, the back-off (also in melee) aip4.
5. **Corrupt Rogue** runs at pace aip4 (both runs); **Corrupt Lancer** runs at pace 100 (both runs).
6. **Fallen** flees a dying kin at pace 50 (0x5f03a8, kept when the back-off is blocked).
7. **Blood Raven**: pace 100 on the walk home, the walk about the target, the 5 percent move about and the back-off, pace 50 on the circle; and her home rule: the flag, pace 100 and the walk home when the **target** is 50 or more from home or she is **more than** 50 (game: `jbe`, d2d tested `>= 50` on her own distance only), both distances FUN_005dc480's (each axis less half the unit's size + 1, `point_apart`; `ai.cpp` feeds them for her only).
8. **Fallen Shaman** (unique): corpse reach by the room-search callback FUN_005f1380: a dead class 19 / 58 monster within aip4 squared, not flagged champion or unique (d2d took any within 80 by `ai_distance`, champions included).

## Proof (counts V: the worker host, seed 20261102, game.exe's own code run, d2d's own code compiled and run)

Speed (claim 1): `tools/emu/walk_velocity.py` (PR #156's oracle, extended): the game side now runs FUN_00573cb0 (hooked: FUN_005739d0, the random draws FUN_0045c390, the level block FUN_00573930, FUN_0063ee90, FUN_00626d40; SetStat FUN_00627260 recorded) for 26 MonStats rows x 3 difficulties and takes stat 0x43 from it (75 on all 78, asserted), instead of the 100 the oracle assumed; d2d is the real build (`test_walk_velocity --walk-velocity`: `monster_update` one frame, `World::walk_percent`, `Fight::merc_turn`, `Fight::pet_walking`).

| build | cases | compared | mismatches (monster rows) | merc / pet / player |
|---|---|---|---|---|
| base `6935cf64` + the oracle change | 122,646 | 122,646 | 50,626 of 61,654 | 0 / 0 / 0 |
| head | 122,646 | 122,646 | 0 | 0 / 0 / 0 |

`mp_run.sh` (this repo, tools/emu; extends upstream's `mon_ais.py` through `mon_ais_act1.py` / `mon_ais_pace.py`, which the PR adds): game.exe's think runs natively on random scenes (the scene, hooks and case format of `mon_ais.py`: every helper hooked at its entry, the draws made as they make them), d2d's real `rules::mon_think` replays each case (the replay block of `tests/monsters/test_monster_ai.cpp`, cut out and counted); 100,000 cases per AI, leader command none or 1 for the Fallen. "pace" is the pace field alone; "all" every field (act, n, x, y, unit, wait, event, pace, steps, the AI control words, the seed after).

| AI | commit | base: all differ / pace differ | head: all / pace |
|---|---|---|---|
| Zombie | `47cc7433` | 23,050 / 23,050 | 0 / 0 |
| Brute | `485c2da4` | 60,005 / 39,818 | 60,005 (act only) / 0 |
| Blood Hawk | `a729676e` | 56,062 / 53,264 | 0 / 0 |
| Corrupt Rogue | `0c07ecad` | 31,717 / 26,149 | 31,717 (n only) / 0 |
| Corrupt Lancer | `94bdd3c5` | 55,674 / 55,674 | 0 / 0 |
| Fallen (73,453 without a dying kin, 26,547 with) | `b0176c59` | 26,547 / 26,547 | 0 / 0 (the 73,453 others 0 before and after) |
| Blood Raven | `bcbcf5e7`, `6b06d6ff` | 66,436 / 65,643 | 0 / 0 |

Blood Raven with the old feed (`ai_distance`, no size term) on the head thinks: 1,231 differ (843 pace): the feed in `ai.cpp` is part of the claim. Her replay uses game.exe's home distances with the unit size hooked at 2 (FUN_00620510): the replay exercises `mon_think`, not `ai.cpp`'s feed; the feed is `point_apart(.., (size >> 1) + 1)` with the sizes `ai.cpp` already uses at lines 741 and 1221 (V read of FUN_005dc480 / FUN_00620510: monster row + 0xd0, a player 2).

Corpse reach (claim 8): `tools/emu/mon_shaman_corpse.py` + `mon_shaman_corpse_harness.cpp` (the real `shaman_corpse_reach`): game.exe's FUN_005f1380 on 100,000 random candidates (class 0x13 / 0x3a and others, flags 0 / 4 / 8 / 0xc, corpse bit, mode, offsets up to 100 and at the squared edge, aip4 1..60): game accepts 8,268; d2d before (`--old`, ai.cpp's old arithmetic) accepts 18,489 and differs on 10,221; head differs on 0.

Committed regression: `tests/mon_ais_pace_cases.inc` (840 game.exe runs, 120 per AI, the corpus format of the other `mon_ais*_cases.inc`; 25 names appended to `mon_ais_names.inc`), replayed by `test_monster_ai`; unit asserts per think in `test_monster_thinks` (Blood Raven's 50 / 51 and target 49 / 50 boundaries included).

Commands a reviewer pastes (worker-host paths replaced by `$D2D`, the upstream checkout with the PR applied, and `$BASE`, a checkout of `6935cf64`):

```
# speed
uv run python tools/emu/walk_velocity.py gen 100000 20261102 > cases.tsv
uv run python tools/emu/walk_velocity.py game cases.tsv > game.tsv          # prints: monster base stat 0x43 (FUN_00573cb0): 75
build/tests/test_walk_velocity --walk-velocity cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
uv run python tools/emu/walk_velocity.py compare cases.tsv game.tsv d2d.tsv
# thinks (per AI group)
PYTHONPATH=tools/emu uv run python tools/emu/mon_ais_pace.py Zombie,Brute,BloodHawk 100000 20261102 --dump out/   # MAI: MAI2_CMD=none1 for the Fallen
# corpse
uv run python tools/emu/mon_shaman_corpse.py gen 100000 20261102 > cases.tsv
c++ -std=c++26 -Icomponents/rules -Icomponents/d2s -Icomponents/txt -include cmake/no_near_far.hpp -o harness tools/emu/mon_shaman_corpse_harness.cpp
harness < cases.tsv > d2d.tsv; uv run python tools/emu/mon_shaman_corpse.py compare cases.tsv d2d.tsv
```

(The counting replay `mp_run.sh` / `mp_patch.py` that produced the think table lives in the evidence repo, not in the PR: it cuts the replay block out of `test_monster_ai.cpp` at upstream's base and counts instead of asserting. The committed test is the same replay with asserts over the 840-case corpus.)

## What else reads the velocity stat (claim 1)

`grep speed_pct|boss_speed|100 +` over `components`, `apps`, `tests` (V): `monster_update` (the plain walk: changed, `kMonsterVelocityPct`), `ai.cpp` 1931 and 1985 (the Diablo run sequence and a whirl sequence's step: left at 100, U whether game.exe moves them by the same stat; not traced), `fight/pets.cpp` 218 and 286 and `fight/summons.cpp` 201 (a pet's walk, summed from 100 by PR #156: left, U: the initialiser runs for every type 1 unit, so a summon probably starts at 75 too, not traced), the merc (left, same), the player (a different stat block: 0 mismatches before and after). Callers of `mon_think` (`ai.cpp`) are listed in `docs/research/re/monster-ai.md`; the only behavioural change in them is the pace they now return, which `ai.cpp` already applies (`monster.move_pct`, the next mode only).

Det scenarios: V run on the worker host: lab binaries = the lab patch set (the det-hash runs' verbs, plus `debug monstate`) cherry-picked onto `6935cf64` with no conflict (base) and then the candidate's 14 commits (head); one hidden d2d slot, no game.exe, `tools/emu/det_scenarios_lab.py` / `det_scenarios_run.py` (PR-external harness), seed 3, lockstep. The 14 scenarios (passive, fight, cold, static, cold2, hit_plain / champion / unique / boss / immune / minion, merc, pet, follow): **`passive` is identical on both (mon b6db27442898, state 94d8c03bf4d3: no monster walks); the other 13 differ in both hashes, as expected**: each has an awake monster walking, and its positions differ (fight block 0: chaser #1 (35.56, 38.67) on base, (35.56, 38.90) on head; chaser #2 (35.56, 37.90) and (35.38, 37.90); the four sleeping ones equal on both; the trajectories are not a plain 75 percent scaling, the chase re-paths and the frame counter moves too, so the hash is not explained by a ratio, only by "the walkers moved differently and nothing else"). base2 reruns of passive, hit_immune and pet repeat base's hashes exactly (the build is deterministic). No scenario hash is claimed unchanged except `passive`; the changed ones are the claimed field (monster positions and what depends on them). Raw lines: `fight base 1124443e0e4e / state 9f5bc7695b85` -> `head 3894564f7d40 / a9d1360892ee`; `hit_plain c24e60a01f14 / 30407a3d31f5` -> `a50888b53036 / 7962906bb946`; the full table is in the evidence repo's run log.

Their suites (V, worker host, a Release build of `local/monsterpace` head `e6a09533`, same tree as the clean branch): `test_rules`, `test_game_rules`, `test_monsters`, `test_monster_thinks`, `test_monster_ai`, `test_combat`, `test_game` all exit 0 (run from `build/tests` with the MPQ folder set); `pr_guard.py 6935cf64 HEAD`: 0 FAIL, 0 WARN; `presubmit.py --quick` (ruff 0.8.6 on PATH): 0 FAIL, 3 WARN (answered under Limits); `clang-tidy -p build` (the CI's invocation: the TUs `tools/tidy-pr-select.py` picks for the changed files, 142 of them, niced, one at a time) 32 findings, all `readability-identifier-length` on lines this branch did not add, none on an added line.

## Limits

- Brute: game.exe's move is FUN_005ded40 (a mode-2 move with path type 0xd, flags 7) and the harness labels it `run`; d2d returns `walk` 7. Left alone (the pace is the claim): the act-only differences stay, 60,005 of 100,000. Zombie / Lancer `run` are RN moves in d2d and FUN_005ded40 moves in game.exe too: the animation mode of those runs is U (not claimed).
- Corrupt Rogue: its run is FUN_005defb0 (RN, flags 3); `run`'s `n` is not read by `ai.cpp`, so the flags stay out: n-only differences remain, 31,717 of 100,000.
- Blood Hawk's `steps` (= the charge's distance) is set only when the caller gives a pointer (the tests); `ai.cpp` gives none, and what the steps do in game.exe is U.
- The initialiser run is hooked outside the claim (list above); the MonStats row's fields it reads for resists and level do not reach stat 0x43. A pet's and a merc's base, the Diablo and whirl sequences' step (above): not changed, not traced.
- Corpse reach covers a unique shaman (the room search, FUN_005a0180 8 true / 4 false). A superunique's flag (bit 2) is not in game.exe's mask 0xc: d2d keeps rejecting it as before (`Boss::superunique`), U. A non-unique shaman (the group search, type 9) was not run.
- The replay hooks the unit size at 2 and always sets off a walk (the same assumptions as `mon_ais.py`'s corpus).
- Not in this PR: the Fallen command word (no Act 1 sender), Foul Crow Nest (flag meaning U).
- Presubmit WARNs, answered: *siblings, unit kinds*: pets, mercs and players keep their own walk sums; see the section above; *docs-sync*: `docs/research/re/monster-ai.md` is updated in the last commit; *ruff*: ruff 0.8.6 clean; *build*: skipped by `--quick`, built and tested above.

## Build flags

Compiled with the CI's flags (`-Wall -Wextra -Wpedantic -Wno-missing-field-initializers -Werror`, g++ 16.2 and clang++ 23.1) and in addition `-Wshadow -Wrange-loop-construct -Wconversion -Wsign-compare` for the changed tests, the harness and `ai.cpp`: no finding on a line this branch added (`ai.cpp:2572` `shot` shadowing and the `ds1` / `dt1` / `cube` sign conversions are on upstream main). `clang-tidy` (`readability-identifier-length`, `misc-include-cleaner`) on the two changed tests: nothing on added lines. Not run: an MSVC build (reviewed by hand for C4456 / C4244 / C4267: no shadowed local, no narrowing).

## What an independent verifier must check

1. The base value: disassemble 0x573cb0..0x573f26 and read `SetStat(unit, 0x43, 0x4b)` at 0x573f16 (the only write; nothing in the think path resets it: `FUN_0057af80` chill and aura 67 add), and re-run `walk_velocity.py game` (it must print 75) and the compare.
2. That the stat initialiser applies to every plain monster unit (called from FUN_00574250 for a type-1 unit), and the opposite risk: a monster the d2d code treats as a plain walker that game.exe moves another way (whirl, charge, leap sequences) is unchanged by design.
3. Each pace against the disassembly at the addresses in the commit messages (Zombie 0x5eff19, Brute 0x5efcbf, Hawk 0x5f01d8 / 0x5f0241 / 0x5f0285, Rogue 0x5f0b5c / 0x5f0c70, Fallen 0x5f03a8, Raven FUN_005e6320), and that the harness's `pace` is the argument of the last FUN_005de190 before the act (a blocked back-off keeps it).
4. Blood Raven's boundaries (`> 50` own, `>= 50` target) with the size term, and `ai.cpp`'s feed (it is the one part the replay does not run).
5. The corpus: `mon_shaman_corpse.py` hooks FUN_005a0180 as "flags & mask"; the claim that bit 8 = unique and bit 4 = champion in monster data +0x16 (S: d2d's own `Boss` mapping, `FUN_005a0180(unit, 0xa)` in the curse code).
