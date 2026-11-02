# PR draft: the player-side missile damage roll: element top, physical in 256ths, -100 floor (PLAYERROLL)

> status: DRAFT held for the independent verifier and the user's go (nothing opened, pushed upstream or posted). Fork branch `pr/player-roll` on the private fork remote, head `28c0540a`, four commits on upstream main `6935cf64` (read-only fetch 2026-11-02): `0e634c55` (oracle and test mode), `bb168658` (element roll top), `b543539d` (physical roll in 256ths), `28c0540a` (-100 floor). Inputs: the roll-top pair 9ba8d369 / 23e2baec (SKMISS: `docs/research/nec-skill-trace.md` D1, `sor-rest-trace.md`), the physical candidate `local/skabrecheck-phys` (`skabrecheck.md` claim 2) and its verifier record `docs/research/pr-physroll-verifier.md` (branch agent/verifyphys; READY WITH CHANGES). The burst roll (`sor-rest-trace.md` D1, with the Meteor deviation) is NOT part of this PR. Tags: V = run on the worker host this session, S = read, U = not verified.

Title: `A missile's element roll never reaches its max, and its physical roll keeps its 256ths (FUN_005a89a0)`

## What changes

game.exe rolls a skill missile's damage in FUN_005a89a0 (physical at record +8, elements at +0x10..+0x28): each part is 0 unless min and max are both over 0, else `min + rand(max - min)` on the missile's seed (FUN_005a8910 / FUN_0045c3e0, max exclusive), in 256ths; FUN_0057bf80 resists that fine value and FUN_0057c6c0 takes it off life in 256ths. d2d's `rules::missile_blow` differed in three ways, one commit each (all in `components/rules/combat.hpp`):

1. `bb168658` element roll: `Rng::range(lo, hi)` is top-inclusive; now `lo + rng(hi - lo)` (ends swapped, 0 unless the lower end is over 0). Covers fire, lightning, cold, poison and magic. Side effect on the RNG stream: a `lo == hi` element no longer draws (`rng(0)` returns 0 without a draw, as FUN_0045c3e0 does for n <= 0, S).
2. `b543539d` physical roll: the roll was shifted to whole points before the resistance (`>> 8` then `resisted`); now the fine roll goes through `resisted_fine`, the whole points go to `damage` and the part under a point to `Blow::frac`, as the element branch already does. Same single RNG draw.
3. `28c0540a` floor: for a monster defender FUN_0057be00 floors the physical resistance at -100; d2d took -150 as x2.5. One token, `std::max(target.res[0], -100)`. A separate commit because it is latent (see Limits): drop it if the maintainer prefers.

`0e634c55` is the oracle (`tools/emu/missile_hit.py`) and `test_missile_hit --missile-hit`, unchanged from the SKMISS pair.

## Reach (callers of `missile_blow`, all with a monster target; S, caller audit in `pr-physroll-verifier.md` section 4 plus a re-read of the call sites)

- Element roll (commit 1): every skill-sourced element hit: player missiles (`Fight::strike`, `missiles.cpp`), area releases, Shiver Armor / Thunder Storm / Hurricane state hits (`states.cpp`), pet and merc skill hits and auras (`pets.cpp`), trap missiles with a Skill (`objects.cpp`).
- Physical roll and floor (commits 2, 3): only a skill with Skills.txt MinDam / MaxDam over 0 reaches the physical branch (rows with no Skill, `missile.fixed`, auras and shrines carry no physical). Amazon: Magic Arrow. Assassin: Psychic Hammer and Mind Blast (through `spot()`), Blade Sentinel, Blade Fury. Barbarian: War Cry. Druid: Twister, Tornado, Shock Wave, Volcano, Armageddon. **No Sorceress, Necromancer or Paladin path.** Not reached: Holy Shield (Smite uses `player_blow`), the Druid summons (own attack stats), Molten Boulder and Blade Shield (rows have no Skill: no physical in d2d at all, U, separate gap).

## Side effect to name

`hurt()` now also receives the physical fraction. A physical hit under one point (damage 0, frac > 0) no longer returns early: it makes the monster aware and rolls its get-hit, and the get-hit size includes the fraction. Both follow game.exe's 256ths (the get-hit comes from the fine total, I, FUN_0057cb00 not re-traced here), but they change monster modes, not only life. Volcano and Armageddon also fire, so the fraction sum can reach 510 per call and `hurt` carries it (S).

## Proof (V, worker host, d2d builds -j2, game.exe's code under the fork emulator as oracle)

| claim | harness | before | after (head `28c0540a`) |
|---|---|---|---|
| physical arithmetic, game.exe's own roll fed to d2d (column B) | author's `skab_phys.py` + `skab_phys_test.cpp`, 101,626 cases, seed 7, compiled against each tree's header | base (upstream main header) **84,271** differ | **0**, at every res; with commits 1 and 2 only (floor off): 16,526 differ, all of them res under -100 |
| same, real `missile_blow` + real `hurt()`, life per hit in 256ths | the verifier's `physroll_game.py` + `test_physprobe`, game.exe's `FUN_0057c6c0` per hit, seed 20261101 (30,000) and 977 (60,000), edge grid, death grid, 1 to 60 hits | the verifier's own base build: 20,881 of 30,000 and 44,544 of 60,000 differ (not re-run on upstream main: that build is `ee5e06d8`, the same physical block) | **0 of 30,000 and 0 of 60,000**, every class including res under -100 (the verifier's head without the floor: 639 and 1,456) |
| element roll, Sorceress (Fire Ball, Glacial Spike) | `sk_rest_hit.py`, 31,000 cases, md5 of game.exe and d2d lines | the old roll is wrong on 30,997 of 31,000 (port) | 0 differ, md5 `f6bc4cf6` both sides |
| element roll, Necromancer (Teeth, Bone Spear, Bone Spirit) | `nec_missile_hit.py`, 36,600 cases (23,300 + 13,300), life after one hit | 10,509 differ (6,794 + 3,715) | **0** (md5 equal both sides in both sets) |

What these do not prove: column A of the author's run (d2d's own roll, same seed) still differs on 83,964 of 101,626 physical cases, because d2d's physical roll is `Rng::range` (top-inclusive, and a zero min is rolled where game.exe rolls nothing): the physical roll range is deliberately not in this PR (see "Not in this PR"). The magic element Necromancer set and the Sorceress set run the roll with the same function, so they cover commit 1 only.

Unit tests: `tests/test_combat.cpp` asserts the element top (`elo 256 / ehi 257` never gives a fraction, and nothing without a min over 0) and, for the physical, a roll of 300 / 256ths: 1 point and 44 / 256 at res 0; 0 and 150 at res 50; 0 and 0 at res 100; and at res -150 exactly the -100 result (2 points and 88). Checks (V, `ctest` in the build of `28c0540a`, data tests with the MPQ folder and patch installer set): `test_combat`, `test_rules`, `test_game_rules`, `test_game` pass; of the 68 tests 55 pass, 7 are skipped (they need data this folder lacks) and 6 failed in the full run: `test_physprobe` (the verifier's probe, not part of the PR: it needs arguments), `test_mpq` (integrity assertion on this MPQ folder), `smoke_d2d` (no d2speech), `quests_act1` and `quests_act4` (an Andariel / Diablo quake timeout, the same on lab binaries without this change), and `pad_use_d2d` failed once while other runs were going and passed on rerun. `pr_guard.py 6935cf64`: 0 FAIL, 0 WARN. `presubmit.py --quick --base 6935cf64`: 0 FAIL, 2 WARN (answered under Limits), ruff clean.

Warnings-as-errors check (coordinator's rule after PR 164): the CI flags are `-Wall -Wextra -Wpedantic -Wno-missing-field-initializers -Werror` (GCC / clang) and `/W4 /WX` (MSVC) in `CMakeLists.txt`. The worker host's build is `D2D_WERROR=OFF` (the 2,250 warnings of the full build are in untouched files; none in `combat.hpp`, `test_combat.cpp`, `test_missile_hit.cpp`). I compiled `test_combat.cpp` and `test_missile_hit.cpp` with the target's own command line on g++ 16.2.1 and clang++ 23.1.4, adding `-Wshadow -Wrange-loop-construct -Wconversion -Wsign-conversion -Wunused`: this found one error in my test (`hit` shadowed an outer `hit`, MSVC C4456), fixed in the same commits (renamed `phys_hit`); after it 0 warnings on every line this PR adds on both compilers. Hand review for MSVC /W4: no new locals besides `phys_rng`, `phys`, `half`, `phys_hit`, no size_t / double conversions, `std::max(int, int)` with the literal `-100`. MSVC itself was not run (U).

## Deterministic scenarios (the lab patch set; base = lab `ee5e06d8`, head = lab + these four commits, branch `local/playerroll-lab`)

The fork's det harness (`tools/emu/det_scenarios_lab.py`, `det_scenarios_run.py`) on two lab binaries, seed 3, lockstep: 14 scenarios plus a new one, `hit_phys` (this card): a Druid casts Twister, Shock Wave and Tornado one after the other point blank on a fresh plain zombie, the physical missile skill the verifier said was still owed.

| scenario | head vs base |
|---|---|
| passive, fight, cold, static, cold2, hit_immune, merc, follow | identical (mon and state) |
| pet | first run differs only in the Spirit Wolf start position (44.1,21.1 vs 44.3,21.0) before any blow, the known run-to-run flake; rerun `base2` = `player2` = `de6f6f0d6623`: identical |
| hit_plain, hit_champion, hit_unique, hit_boss, hit_minion | differ in the target's hit points only (monster lines; one point here and there), the claimed element roll: the physical-only lab binary (`ee5e06d8` + `5b13a2c4`, no roll-top) hashes the same as base on all five, so commit 1 alone moves them |
| hit_phys (new) | differs from base in 5 of 27 snapshots, monster hit points only, from Tornado on (`phys skill 245`, 125 against 126 of 147); identical to the physical-only binary, so commit 2 alone moves it, repeatable (`base2` and `player2` reproduce both hashes). Twister and Shock Wave did not move a hash here |

So 9 of the 15 scenarios are byte-identical (the pet after its flake rerun), 6 differ in exactly the claimed field by the claimed commit (five by commit 1, `hit_phys` by commit 2).

## Limits

- Commit 3 (floor) is latent: MonStats physical resistance is not below 0 and Amplify Damage gives exactly -100, so res under -100 needs a stacked source. It is a separate commit for that reason; with it the verifier's harness is 0 of 60,000 at all res (without it 1,456).
- The physical roll range is not changed: game.exe's `min + rand(max - min)` and `0 unless both ends over 0` (D3 / D5 of `skabrecheck.md`) for the physical still differ (column A above). Same function, a next commit once the maintainer has taken this one; the roll-top commit's comment in `docs/research/re/missiles.md` already names the physical as the same rule.
- Not in this PR: the srcdam residual (weapon share whole-point in `player_blow`, then the skill's fine part: Magic Arrow, Blade Fury), `Fight::land` soak using whole-point `resisted(blow.phys, ..)` and dropping `frac` when soak takes part (only monsters with an armour pool), Molten Boulder / Blade Shield carrying no physical, the percent stats FUN_005a89a0 applies (0 for skill missiles, I), the Sorceress burst roll (`sor-rest-trace.md` D1).
- Presubmit WARNs, answered: *doc-claims, `docs/research/re/missiles.md` line 106 ("never")*: the line states game.exe's rule (FUN_005a8910 never rolls the max); d2d now implements it for every element branch (commit 1) and the physical branch is named above as not yet; *build skipped (--quick)*: answered by the ctest run above.
- Not run: an MSVC build, a live game, the full `presubmit` build and clang-tidy step.

## How to review

1. `components/rules/combat.hpp`, `missile_blow` (the whole change: three small hunks).
2. FUN_005a89a0, FUN_005a8910, FUN_0045c3e0, FUN_0057be00, FUN_0057bf80, FUN_0057c6c0 with `tools/emu/disasm.py`.
3. Elements: `tools/emu/missile_hit.py` + `test_missile_hit --missile-hit`. Physical: `tools/emu/skab_phys.py gen / game` and `cmp` against `skab_phys_test.cpp --phys`; the real-`hurt()` probe is in the verifier's record (`pr-physroll-verifier.md`, Reproduce).
4. Det: `det_scenarios_run.py run BINDIR OUT hit_phys,... base,head` inside a d2d slot.

## What the independent verifier must check (narrow, third party)

- The combined head `28c0540a` on upstream main: `git diff 6935cf64 28c0540a` is the 4 commits above; `git merge-tree` clean against current upstream main.
- The floor: res under -100 at head against game.exe on a res grid that includes -101 .. -300 and a monster defender; and that FUN_0057be00's floor is only for monster defenders (every caller here has a monster target).
- The det scenario `hit_phys` exercises a real physical hit (the five moved snapshots) and the five cold hit scenarios move only through commit 1.
- The RNG draw change for `lo == hi` elements (commit 1 side effect) is consistent with FUN_0045c3e0 and with how other callers draw.
- Warnings-as-errors: build with the CI flags on a toolchain that has them (MSVC /W4 /WX was not run here).

## Shared-file lines for the merger

In `openspec/changes/playerroll/merge-lines.txt`.
