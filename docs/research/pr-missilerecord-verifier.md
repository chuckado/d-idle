# Independent verifier: local/missilediff-record (monster missile launch record) (VERIFYMISSILE, 2026-11-01)

> status: current verifier record (verdict NOT READY, section 7) for the private-fork candidate `local/missilediff-record` (9d7e7d81 refactor + test mode, 312b0bbb fix; parent 23e2baec = upstream 72132b99 + the unsubmitted roll-top pair 9ba8d369 / 23e2baec). Author's evidence: `docs/research/missilediff.md` (commit a8515f8, card missilediff), read only after my own result. Offline: no game, no game instance; nothing pushed or posted; the candidate branch and the author's worktree untouched (my trees are separate detached worktrees on the worker host). Verdict in section 7.

Tags: **V run** game.exe code run under upstream's emulator in MY oracle, or real d2d builds run with MY probe; **V read** disassembled by me (upstream `tools/emu/disasm.py`) or d2d source read at the named commit; **D** 1.14d MPQ tables; **S** stated in a doc or comment, not re-traced; **U** not traced; **O** opinion.

## 1. Verdict in one paragraph

The record claim holds: for every Act 1 / 2 MissA-path row (and all 78 MissA1 / MissA2 / MissS1 / MissC rows of every act), the head's `monster_shot_phys` / `monster_shot_stats` equal game.exe's `FUN_0064b860` record in 0 of ~250,000 cases, versus thousands of differences on base (V run, two seeds plus an all-acts set). Base vs head moves exactly the rows with brackets, a HitShift other than 8 on a non-zero MinDamage, or an own element; every other row is byte-identical (V run). **But the new row element breaks at the consumer**: the head stores the row's poison as a whole-point total over ELen in `MonStats::elements`, and `monster_blow` reads a type-3 element in the MonStats `El` convention (per frame = roll x 10 in 256ths, for 2 x dur frames). A Sand Maggot spit (goospit1, Normal) then poisons for 93.75 points instead of game.exe's 6.25, an unraveler4 / Radament spit (mummy4, Hell) for 4,281 instead of 137.5: x 20 x ELen / 256, i.e. x 7.8 to x 31 on the Act 2 rows (V run through the real `monster_blow`, V read for the units). Base deals no row poison at all, so the head turns a missing 6-170 point poison into a lethal one. **Verdict: NOT READY** until the poison goes in with the right units (section 7; a small change).

## 2. game.exe: what the record is (V read, V run)

| Code | What it does | Tag |
|---|---|---|
| FUN_0059fa30 -> FUN_0059f900 | spawns the missile, then `FUN_0064b860(rec[0x7c], owner, srcmiss_unit, missile, lvl)` and `FUN_0064ac60` (record -> the missile's stats) | V read |
| FUN_0064b860, row path (row `Skill` +0x194 <= 0) | phys min / max = `FUN_0064af20` / `FUN_0064aff0`; element min / max / len = `FUN_0064b100` / `0064b1d0` / `0064b2a0`, type = row +0xe4, switched by the jump table at 0x64bf24 into record slots (1 fire +0xc / +0x10 / +0x14, 2 ltng +0x18, 3 mag +0x20, 4 cold +0x28..+0x30, 5 pois +0x34 / +0x38 / +0x3c, 6..9 life / mana / stamina drain / stun, 11 burn, 12 frze = cold's slot) | V read, V run (table read from the image) |
| FUN_0064af20 | `(MinDamage + FUN_00644b70(MinLevDam1..5, lvl) [+ MulDiv(that, DmgSymPerCalc, 100)]) << HitShift (+0x196)`; lvl <= 0 takes the missile's own level | V read |
| FUN_0064b100 / 0064b2a0 | the same shape on EMin / MinELev / EDmgSymPerCalc, << HitShift; length ELen + ELevLen1..3 brackets (8 / 16) | V read |
| owner part (row SrcDamage +0x12d, a byte) | for a monster owner with no item: stat 0x15 / 0x16 << 8, then `x SrcDamage / 128` (signed, truncated toward 0), added to the row part; **only when SrcDamage != 0**: the owner's element stats 0x30..0x3b and 0x13b..0x13d added (`FUN_0064a930`) and scaled by SrcDamage / 128 (`FUN_0064aa60`, skipped at 128) | V read |
| SrcDamage `-1` | stored as the byte 255 (V run: the table check matched `SrcDamage & 0xff` on all 684 rows), so it scales x 255 / 128, not 0 | V run |
| `DmgSymPerCalc` / `EDmgSymPerCalc` | set on none of the 50 Act 1 / 2 rows nor the 78 monster Miss rows: the calc term is never exercised here | D |

Not re-traced: that a monster's MissA missile level is MonsterSkillBonus + 1 (1 / 4 / 8). Both sides take the level as an input, so the 0-diff holds at any level the two share; my grid covers 1..40, 50, 60, 99 (U for the level itself, S from d2d's comment at ai.cpp:2003).

## 3. My harness (worker host, `~/d2d-lab/verifymissile/harness/`)

Written from the claim and the sources only; the author's `mm_diff.py` / `mm_launch.py` / `test_monster_missile.cpp` were not used or read until section 6.

- **Rows** (`vm_rows.py`): Act 1 / 2 classes = Levels.txt Act 0 / 1 mon / nmon / umon + superuniques 0..20 and Corpsefire + andariel, duriel, bloodraven, radament, summoner, closed over minion1 / minion2 / spawn: 129 classes, **50 rows**: 33 on the MissA path (raven2 is Blood Raven's MissSQ, which d2d never loads) and 17 skill-sourced (srvmissile*) (D). The brief's "49" is within one of this; the author's set differs (section 6).
- **Oracle** (`vm_game.py`): boots game.exe's tables as upstream's `skill_tips.boot()` does and calls `FUN_0064b860` directly on a fake monster owner (type 1, class 0, no item, not a hireling) and a fake missile unit (type 3, the row id). Hooked: the stat getters `FUN_00625480` (owner 0x15 / 0x16 = the case's source pair, every other stat 0) and `FUN_00625500` (0), and `FUN_00623c80` (weapon-switch helper). Native: everything in section 2. Output: the record's 31 dwords. Check mode: the row table read from game memory equals Missiles.txt on MinDamage, HitShift, SrcDamage and EMin for all 684 rows. Rows with `Skill` set (fireball, frostnova, glacialspike, vampiremeteorcenter) take the skill path and are skipped.
- **d2d probe** (`test_vm_probe.cpp`, dropped into `tests/game/` of each scratch tree, `// test-links: game`): loads the game data with `load_game_data`, calls the real `monster_shot_phys` / `monster_shot_stats` (the seam `Ai::tick` calls at ai.cpp:2014) and, with `VM_BLOW=1`, the real `rules::monster_blow` on the produced `src` against a plain defender (to-hit forced, 0 resistances) to see what the hit applies.
- **Base** is 9d7e7d81, not 23e2baec: 23e2baec has no seam to call. V read: 9d7e7d81 moves ai.cpp:2014-2020 into the two functions with the same expressions (`(x << 8) >> 8` round-trips at these magnitudes), so it is the base behaviour.
- **Cases** (`vm_gen.py ROWS N SEED`, Python `random`): grid = every row x levels 1..40, 50, 60, 99 x shot A1 / A2 / S1 / SC x owner pairs (0,0) (1,1) (2,5) (5,2) [min > max] (37,61) (255,256) (1000,1500), the other two pairs random; then N random cases (levels weighted to 1 / 4 / 8, owners to 5,000). Seed A 20261101 (by chance the author's seed number, but a different generator, so different cases), seed B 4242, all-acts set (78 rows) seed 77031. Cases md5 (first 12): A 636df64716ea, B 161c21644384, all 1a992a870520.
- **Compare** (`vm_cmp.py`): physical exactly in 256ths; elements converted to d2d's own representation (min / max >> 8; poison (per-frame x len) >> 8; cold / poison length), order-insensitive on min / max because d2d clamps max >= min; plus counts of records whose physical has a fraction (D5) and of game elements d2d cannot carry (drains, stun, burn: 0 here). `vm_moves.py`: base vs head line by line, split into rows the change touches and rows it must not.
- Builds: detached worktrees `base` (23e2baec), `mid` (9d7e7d81), `head` (312b0bbb) under `~/d2d-lab/verifymissile/`, Release, `-DD2D_WERROR=OFF` (GCC 16 trips `-Werror` in `apps/d2d/ui.cpp`, unrelated to the change), `nice ninja -j2`.

## 4. Base vs head (V run)

Act 1 / 2 MissA-path rows (raven2 not loadable by d2d: 1,046 / 980 cases set aside):

| class | base (9d7e7d81), seed A / B | head (312b0bbb), seed A / B |
|---|---|---|
| at levels 1 / 4 / 8: cases | 34,067 / 34,041 | same |
| at 1 / 4 / 8: any field differs | **11,804 / 11,701** (seed A: lvl 1 3,033; lvl 4 4,418; lvl 8 4,353) | **0 / 0** |
| at 1 / 4 / 8: physical differs | 6,416 / 6,311 | 0 / 0 |
| at 1 / 4 / 8: element compared / differs | 9,108 / 9,108; 9,021 / 9,021 | 9,108 / 0; 9,021 / 0 |
| other levels: physical differs | 11,141 / 11,248 (of 47,582 / 47,616) | 0 / 0 |
| other levels: element compared / differs | 12,626 / 12,626; 12,627 / 12,627 | 0 differ |
| physical carries a fraction (D5, not in the claim) | 7,260 / 7,169 | same |

Rows that differ on base: goospit1-3 (D2 at every level, D1 from level 2, D3 poison), spike1-4 (D1 from level 2), pantherpotorange (D1 from level 2, D3 fire), mummy1-4 and pantherpotgreen (D3 poison only). Rows equal on base too: bighead1-4, cr_arrow1-4, pantherjav1/2/4, raven1, skbowarrow1-5, skmage1-4.

All 78 monster Miss rows of every act (seed 77031, 123,912 cases; d2d does not load highpriestlightning, willowisplightningbolt, willowisplightningbolt2: 4,802 cases set aside; the Miss names with a space, `ancient throwing ...` and two trap rows, were mis-split by my row list and not run): base physical 4,126 + 15,723 differ, elements 31,799 of 31,799 compared differ; **head 0 and 0**. Extra rows reached: goospit4/5/ex, mummyex, regurgitatorcorpse, spike5, frogfire / frogpois, impmissex, arcanelightningbolt, tentaclegoo, trapfirebolt, trappoisonjavcloud.

The 17 skill-sourced rows run through the same seam also match game.exe's record on head (0 of 26,535), but d2d never sends them through it: they go through `andariel_missile` / `missile_blow` (ai.cpp:380), which the candidate does not touch.

**Must stay equal** (`vm_moves.py`): of the cases on rows with no bracket, HitShift 8 (or MinDamage 0) and no own element, 0 moved between base and head (seed A 67,138 cases; seed B and all-acts likewise 0). Rows the change touches moved in every case at levels 4 / 8 / other and in 5,747 of 7,112 at level 1 (spike / regurgitator have no level-1 bracket). Output md5 (first 12): game A 450f7e3ce547, B 88287011f1e9, all f58d8a5ac3c6; base A dba169fac496, head A ec9cb7848a5d, base B 6936e0077f5a, head B a6f7d86e7242.

Break attempts:
1. min > max owner pairs, 0 / 0, 5,000-point owners, levels 50 / 60 / 99: 0 differ on head (V run).
2. Cast (SC, owner part 0) and S1 shots on every row: 0 differ (V run); but d2d loads only MissA1 / MissA2 / skill rows (gamedata_load.cpp:328-333), so a monster whose MissS1 / MissC / MissSQ row is only that (the Council's highpriestlightning, the wisps' bolts, Blood Raven's raven2) never fires it, on base and head (V read + V run "norow"; pre-existing, not this PR).
3. Three El rows of the shot's mode would leave no free slot for the row element: no such shooter in monstats (D, all acts). Own El of the shot's mode plus a row element: only the Act 5 imps (imp6-8, impmissex fire, SrcDamage 0). game.exe drops the owner's elements when SrcDamage is 0 (section 2); d2d carries them on base and head, and head adds the row's fire on top (V read; Act 5, low).
4. SrcDamage -1: no monster Miss row has one (D: only poisonjavcloud, plaguejavcloud, furylightning); head treats it as 0, game.exe as 255 (latent, low).

## 5. The consumer: what the hit applies (the blocker)

`x.src` goes to `rules::monster_blow(..., missile.src, true, rng, true)` on impact (ai.hpp:829). There a type-3 element is applied as `per = resisted(roll * 10, res)` (256ths a frame) for `resisted(dur * 2, plr)` frames (combat.hpp:600): the MonStats `El` convention (`swing_max_damage`: "poison x 10"). The player's poison runs `per / 40` 256ths a millisecond (fight/tick.cpp:512, fight/setup.cpp:232), the merc's "256ths a frame" (fight/tick.cpp:472). game.exe's record carries per-frame 256ths (+0x34 / +0x38) for +0x3c frames.

The head writes `{3, 100, P_lo x L >> 8, P_hi x L >> 8, L, "A2"}` (a whole-point total), so the hit deals `10 x (P x L / 256)` per frame for `2L` frames: **x 20 L / 256** of game.exe's total (V run, the real `monster_blow`):

| row, level | game.exe record: per frame, frames -> total points | head's hit: per frame, frames -> total | ratio |
|---|---|---|---|
| goospit1, 1 (Sand Maggot, Normal) | 8, 200 -> 6.25 | 60, 400 -> 93.75 | x 15.0 |
| goospit1, 8 (Hell) | 50, 200 -> 39.1 | 390, 400 -> 609 | x 15.6 |
| mummy1, 1 (unraveler1) | 18, 100 -> 7.0 | 70, 200 -> 54.7 | x 7.8 |
| mummy4, 1 (Radament, unraveler4/5) | 18, 400 -> 28.1 | 280, 800 -> 875 | x 31 |
| mummy4, 8 | 88, 400 -> 137.5 | 1,370, 800 -> 4,281 | x 31 |
| pantherpotgreen, 8 (slinger6) | 102-108, 400 -> 159-169 | 1,680, 800 -> 5,250 | x 31 |
| andypoisonbolt, 1 (Andariel; NOT this path) | 32-64, 800 -> 100-200 | through `missile_blow`, same encoding: 1,410, 1,600 -> 8,812 | x 62 |

Base deals 0 row poison (the D3 bug). Head deals 8 to 31 times game.exe's on the Act 2 MissA rows. The last row shows the encoding is inherited from `missile_blow` (ai.hpp:653, Andariel's poison bolt and spray), which is wrong the same way on the fork today, and on upstream main (same lines; V read) (pre-existing; same root cause; V run).

Fire (pantherpotorange) is fine: `elements` fire min / max are whole points and the hit rolls them as points (game 13,056-13,824 256ths at level 8 = head 51-54 points; only the fraction is lost, D5).

## 6. Against the author's doc (read after section 5)

| Point | Author | Me | Agreement |
|---|---|---|---|
| Record after the fix | 0 of 120,000 | 0 of ~250,000 (two seeds, all-acts set) | agree |
| Record before | 44,456 of 120,000; 842 of 2,560 at levels 1 / 4 / 8 | 11,804 of 34,067 at levels 1 / 4 / 8 (another grid and denominators) | agree in kind; counts not comparable |
| Rows that differ | goospit1-3, mummy1-4, pantherpotgreen / orange, regurgitatorcorpse, spike1-4 | the same; regurgitatorcorpse only in my all-acts set (it is not an Act 1 / 2 shooter in my enumeration, nor is blowgun) | agree |
| Poison convention | "I: d2d's existing convention, not proven equal to what the hit applies"; impact path not run | run: the hit applies x 20 L / 256 | **disagree: their open "I" is the blocker** |
| Game owner | real spawned monster, mode overlay 005a4f50, owner-El cases skipped | fake owner, stats hooked, no owner El | neither tests the owner-El path |
| raven1, skill rows | not compared | raven1 compared (0 diff, no row damage); skill rows through the seam for completeness (0 diff), not on d2d's path | agree |
| Suites | test_game, test_combat, presubmit, pr_guard not run | run here (section 8) | |

The author's ruff findings (8, a presubmit FAIL) include two `B023` "function definition does not bind loop variable" at `mm_launch.py:133`, worth a look as a possible oracle bug (O; their 0-diff result is reproduced independently here, so it does not change the conclusion).

## 7. Verdict and the changes

**NOT READY.** Required:
1. **Poison units (blocker).** Carry the row's poison as per-frame 256ths for ELen frames, not as a whole-point total in `elements`. Least invasive: `MonStats::poison_min / poison_max / poison_len` (combat.hpp: "per tick in 256ths, ticks", already applied by `monster_blow` on every hit), added to any owner value the way `FUN_0064a930` adds into +0x34 / +0x38 / +0x3c; keep fire / cold / lightning / magic in `elements`. Add a test that runs `monster_blow` on a goospit1 Normal shot and expects 8 a frame for 200 frames (6.25 points), and a mummy4 Hell one (88 for 400).
2. **`missile_blow`'s poison** (ai.hpp:653; Andariel's poison bolt and spray, x 62 today): same root cause. Fix it here (and say so) or name a sibling PR in the body; presubmit's "siblings" WARN asks exactly this.
3. **presubmit FAILs:** `short-names` on `MonsterShotPhys { int lo, hi }` (ai.hpp:670; rename e.g. `low` / `high`); `ruff`: 8 findings in `tools/emu/mm_diff.py` / `mm_launch.py` (unused imports, unused loop variables, B023). Answer the WARNs in the body (siblings: player / merc / pet paths unchanged; docs-sync: no `docs/research/re/missiles.md` line yet).
4. **Base:** the candidate sits on 9ba8d369 / 23e2baec (the roll-top pair), which is neither upstream nor an open PR (read-only `gh pr list`). Open that first, or rebase the two commits on upstream main: they apply cleanly on current upstream main 6935cf64 (`git merge-tree --merge-base=23e2baec`, no conflict) and the whole branch merges cleanly too (V run).
5. **D5 (whole-point `>> 8`): exclude** from this PR (O). It changes `monster_blow`'s roll for every monster missile (and the hit's 256ths), another function with a bigger blast radius; for Act 1 / 2 it only matters on goospit's sub-point physical (2..48 256ths, under one point at every level to 8) and the fire fraction. Keep it in Limits, as the author does.
6. Body text: the S1 / C branches are mostly unreachable (MissS1 / MissC rows are loaded only when also A1 / A2 / skill rows); the imp double fire (SrcDamage 0 should drop the owner's El, section 4 item 3) as a known gap.

Not required but cheap: record cases for the three space-named rows I did not run; the author's lab drill M2 after change 1, reading the poison state.

## 8. Fork suites, style checks, merge, det hashes

| Check | Result | Tag |
|---|---|---|
| `ctest` on head 312b0bbb (all 52 test executables built; `D2_MPQ_DIR` / `D2_PATCH_INSTALLER` set, so data tests ran) | test_game, test_game_rules, test_combat, test_monsters, test_monster_ai, test_rules, test_skills, test_drops and every other unit test pass; skipped: test_missile_hit / test_monster_missile (oracle modes), test_iso9660; failed: test_mpq (the integrity assertion on my MPQ folder: no d2video / d2xvideo there, environment) and the 9 tests that run the d2d binary (smoke, co-op, pad, quests act 1-5: no d2d app in a tests-only build) | V run |
| `ctest` on base 23e2baec, same set (minus test_monster_missile) | the same 10 failures (test_mpq + the 9 d2d-binary tests), everything else passes: the candidate adds no test failure | V run |
| `tools/pr_guard.py 23e2baec HEAD` | 0 FAIL, 0 WARN | V run |
| `tools/presubmit.py --quick --base 23e2baec` (ruff via uvx) | **2 FAIL**: `short-names` ai.hpp:670 `lo` (and `hi`); `ruff` 8 findings in mm_diff.py / mm_launch.py. 3 WARN: siblings (unit kinds, hit paths), docs-sync. Same against upstream/main | V run |
| clang-tidy (repo `.clang-tidy`, LLVM 23) on test_monster_missile.cpp and gamedata_load.cpp | **new-file errors**: test_monster_missile.cpp:13 `combat.hpp` included but unused (misc-include-cleaner); :47 variable `s1` too short. (outdoor.hpp `x` / `y` binding names: pre-existing, newer tidy only) | V run |
| merge with upstream main 6935cf64 (read-only fetch) | whole branch: clean (`git merge-tree --write-tree`, no conflict); the two candidate commits alone on upstream (`--merge-base=23e2baec`): clean | V run |
| base commits upstream? | 9ba8d369 / 23e2baec not in upstream main, not an open PR | V run |

**Det state hashes and a d2d drill** (V run; one hidden d2d slot on the worker host, no game.exe, no game instance). Lab binaries = the lab patch set of the det-hash runs (docs/research/det-hashes-2026-11-01b.md) cherry-picked with 9ba8d369 + 23e2baec (base, d2d sha256 69308544c353) and then 9d7e7d81 + 312b0bbb (head, d6c0bc50e482), all clean; `det_scenarios_lab.py` / `det_scenarios_run.py` copied into my harness dir, seed 3, lockstep.

- The 14 existing scenarios (passive, fight, cold, static, cold2, hit_plain / champion / unique / boss / immune / minion, merc, pet, follow): **base = head on every `mon` and `state` hash** (e.g. passive d1ca6c463f45 / 84942b4a0e26, the det-hash runs' upstream values; hit_minion, Hell Blood Moor, 778c963eb8da / 367f7e302e78). base2 reruns of passive, hit_immune, spit_mummy repeat base's hashes. Expected: no affected row fires in those scenarios.
- My added scenario `spit_<shooter>[_h]` (in my copy only): one shooter spawned (`debug spawn`), a passive player 4 cells south (900 life Normal, 9000 Hell), 10 blocks of 50 frames, the player's life (`dump stats` raw[6]) per block. Player life lost over the 500 frames, base -> head:

| shooter (row) | Normal | Hell |
|---|---|---|
| sandmaggot1 (goospit1) | 22 -> 83 | 351 -> 1,168 |
| unraveler1 (mummy1) | 43 -> 115 | 1,004 -> 1,485 (both include melee) |
| slinger6 (pantherpotgreen) | no throw in the window | **0 -> 5,922** (one pot, then 637 a block) |
| slinger5 (pantherpotorange) | no throw | 5 -> 144 (one hit: D1 physical + D3 fire, as game.exe's record) |
| quillrat1 (spike1) | no shot | 36 -> 57 (D1 bracket, +21 = 7 x 3) |

The Hell Night Slinger line is the blocker of section 5 in the real game loop: one green pot takes 5,922 life over 20 s and keeps going (game.exe's record: 102-108 a frame for 400 frames, 159-169 points in all).

## Reproduce (worker host)

```sh
cd ~/d2d-lab/verifymissile/harness
python3 vm_rows.py > rows.txt && python3 vm_gen.py rows.txt 60000 20261101 > cases.tsv
cd ../head/tools/emu && D2_MPQ_DIR=<MPQ dir with bin/game.exe> D2_BUILD=<build with mpq-cat> D2_PATCH_INSTALLER=<LODPatch_114d.exe> PYTHONPATH=$PWD \
  .venv/bin/python ../../../harness/vm_game.py game ../../../harness/cases.tsv > ../../../harness/game.txt
cd ../../../harness && for b in mid head; do VM_BLOW=1 ../$b/build/tests/test_vm_probe cases.tsv > ${b}_cases.txt; done
python3 vm_cmp.py cases.tsv game.txt head_cases.txt --ids 5; python3 vm_moves.py cases.tsv mid_cases.txt head_cases.txt
```

## Left on the worker host

`~/d2d-lab/verifymissile/`: the harness (`harness/`, with outputs), detached worktrees `base` / `mid` / `head` of the fork checkout (my probe file untracked in each `tests/game/`; remove with `git -C ~/d2d-lab/fork/src worktree remove --force <dir>`), the two lab d2d binaries (`bin/`) and the det logs (`det/`). The lab worktrees and their temporary ref were removed; nothing was committed to or pushed from the fork. Nothing of mine is registered (`tools/agentreap.sh --list`).

## Shared-file lines for the merger

- BACKLOG: `- MISSILEDIFF-FIX: candidate local/missilediff-record NOT READY (docs/research/pr-missilerecord-verifier.md): the row poison reaches monster_blow x 20 x ELen / 256 too strong (carry it as per-frame 256ths, MonStats::poison_*); same bug in ai.hpp missile_blow (Andariel's poison, x 62, upstream too); presubmit short-names + ruff FAILs; tidy include / s1 in test_monster_missile.cpp; rebase off the unsubmitted roll-top pair`
- INDEX: `- docs/research/pr-missilerecord-verifier.md: independent verifier of the monster missile record candidate (record 0-diff confirmed; poison units blocker; det hashes equal)`
