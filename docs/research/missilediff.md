# Monster missiles Acts 1 and 2: the real-build differential for D1, D2 and D3 (card missilediff, 2026-11-01)

> status: current, corrected 2026-11-01 (card missilefix): the verifier found the row poison stored in the wrong units; fixed on the fork (section "Verifier findings and fixes"), re-verification by the independent verifier owed before any PR. Follows docs/research/monster-missile-trace.md (D1-D6 there were proven against a Python port of d2d's expression; this card replaces the port by the REAL d2d build). Offline only: no game, no d2d slot. d2d side: fork worktree `~/d2d-lab/fork/wt-missilediff` on the worker host (a scratch build dir inside it, `-j1` / `-j2`), base upstream `23e2baec`; local branch `local/missilediff-record` (commits `9d7e7d81` test mode + refactor, `312b0bbb` the fix), pushed to the private fork remote only, never upstream, nothing posted. game.exe md5 2b3600c1... under upstream's unicorn harness (`tools/emu`).

Tags: **V run** = game.exe's own code run under the emulator AND the real d2d functions built and run, outputs compared line by line (both sides are runs here); **V read**, **D** (table), **S**, **I**, **U** as in AGENTS.md.

## Verdict

| D | Verdict | Counts (V run, real d2d build before the fix; 120,000 cases, seed 20261101) |
|---|---|---|
| D1 NM / Hell level brackets (MinLevDam / MaxLevDam) dropped | **CONFIRMED** | spike1-4 and regurgitatorcorpse and pantherpotorange: 0 differ at missile level 1 (spike: 188 cases, 0 differ), differ at every level 2+ (spike 9,201 of 9,201 cases above level 1); at the game's own levels 1 / 4 / 8: 2,560 cases, 842 differ |
| D2 HitShift ignored (MinDamage as whole points) | **CONFIRMED** | goospit1-3 at level 1 (no bracket yet): 307 of 307 cases differ (d2d 1-3 points, game 2-6 256ths); same row family: 14,181 of 14,181 differ |
| D3 the row's own element (poison, fire) missing on the MissA path | **CONFIRMED** | element compared in 55,769 cases (owner without an element of the shot's mode), 28,300 differ: goospit1-3, mummy1-4, pantherpotgreen (poison), pantherpotorange (fire); d2d carries none of them |

Totals before the fix: **44,456 of 120,000 cases differ** (physical 32,690; element 28,300; both 16,534; physical only 16,156; element only 11,766), 14 of the 36 rows. After the fix: **0 of 120,000** (physical 0; element 55,769 compared, 0 differ; at the game's own levels 2,560 cases, 0 differ). 21 rows show 0 differences in the BEFORE run too (raven1 is not compared) (they carry no bracket, shift 8 and no row element): bighead1-4, cr_arrow1-4, blowgun, pantherjav1/2/4, skbowarrow1-5, skmage1-4.

PR candidate: **yes, local** (branch above; the independent verifier comes next). Limits below (the roll stays whole points, D5; only the record is exact).

## Questions answered

| Q | Answer | Tag |
|---|---|---|
| Q1 | What does d2d's record function do for a MissA / MissS1 / MissC shot? Extracted without a behaviour change from `ai.cpp:2014-2020` into `monster_shot_phys` / `monster_shot_stats` (components/game/ai.hpp), `Ai::tick` calls them; the test mode calls the same functions. Before the fix: `(src x SrcDamage / 128) + MinDamage` in whole points, no bracket, no shift, no row element | V read + V run (the extraction is the before binary) |
| Q2 | What does game.exe do? `FUN_0059fa30` -> `0059f900` -> `0064b860` / `0064ac60`: physical min = `(owner stat x 256 x SrcDamage / 128, truncated) + ((MinDamage + MinLevDam bracket at the missile level) << HitShift)`; the row's own element by `0064b100 / 0064b1d0 / 0064b2a0` at the level. The formula implemented by the fix equals game.exe's output in all 120,000 cases | V run (real code, stats read with `00625480`) |
| Q3 | Is the earlier D1-D3 evidence (a port) right? Yes: the real build reproduces every conclusion, with the sharper numbers above | V run |
| Q4 | Does the missile level matter beyond 1 / 4 / 8? The grid covers levels 1..40 (boundaries 8 / 9 / 16 / 17 / 22 / 23 / 28 / 29 inside) and random levels to 60; the game's own levels are 1 / 4 / 8 | V run |

## The differential

- **Game side** `tools/emu/mm_diff.py game TABLES N SEED OUT` (extends `mm_launch.py`: same boot, same spawn, same mode overlay `005a4f50`): per difficulty (a fresh game: difficulty byte 0 / 1 / 2) every Act 1 / 2 shooter row whose source is MissA1 / MissA2 / MissS1 / MissC (36 distinct rows; the skill-sourced rows go through `andariel_missile` in d2d and are not on this path), missile level 1..40 x 11 owner pairs per source (boundaries 0, 1, 127, 128, 255, 256, 1000-1500 and three draws) plus random cases to the budget. The owner's stats 0x15 / 0x16 are REWRITTEN in its stat list (an entry inserted or grown when the monster has none; points, as the getter returns them: V run, the getter read back equals the written value in every case), the missile made by `0059fa30` at the case's level, its stats read with `00625480`: physical 0x15 / 0x16 (256ths), elements 0x30..0x3b. xorshift32 masked to 32 bits, seed 20261101, shuffled so every row is touched.
- **d2d side** `tests/game/test_monster_missile.cpp` (`--monster-missile CASES`): builds in the scratch dir, reads the same game data (`load_game_data` over the MPQ copy and the 1.14d patch), calls the real `monster_shot_phys` (256ths) and `monster_shot_stats` (the elements the missile's `src` carries) per case line `ID ROW FIRST CAST S1 LVL A1MIN A1MAX A2MIN A2MAX S1MIN S1MAX`.
- **Compare** `mm_diff.py cmp OUT D2D_OUTPUT`: physical in 256ths exactly; elements per type: fire / lightning / magic min and max `>> 8`, cold the same plus its length (stat 0x38), poison per frame in 256ths and its length (stats 0x39 / 0x3a / 0x3b against `MonStats::poison_min / poison_max / poison_len`, corrected in card missilefix; the first version compared a whole-point total, the convention d2d's own `missile_blow` used, which the hit does not apply as such, see below).
- Skipped by design: the element comparison where the shooter's own El row fires in the shot's mode (its element rides on the game's missile too; physical still compared) and Blood Raven's rows (raven1: her record carries her weapon, D5 of the trace; nothing compared, 2,379 cases). Together 64,231 cases have no element comparison. `D2_MPQ_DIR=$HOME/mmtab/emu` (the 1.14d patch layer through `D2_PATCH_INSTALLER`), tables in `~/mmtab`.
- Outputs (worker host, `/tmp/md_full`): cases md5 `10ac65af...`, game.tsv `4775fa30...`, d2d-before `a67cc453...`, d2d-after `9003bf02...`.

| Build | Cases | Physical differ | Element compared / differ | Any differ | At level 1 / 4 / 8 |
|---|---|---|---|---|---|
| base 23e2baec + test mode and extraction (`9d7e7d81`) BEFORE | 120,000 | 32,690 | 55,769 / 28,300 | 44,456 | 2,560 cases, 842 differ |
| + the fix (`312b0bbb`) AFTER | 120,000 | 0 | 55,769 / 0 | 0 | 2,560 cases, 0 differ |

Rows with a difference before (row: cases differing of cases): goospit1 4,768 / goospit2 4,667 / goospit3 4,746 (all of them), mummy1-4 2,322 / 2,400 / 2,293 / 2,352 (element only), pantherpotgreen 2,355 (element only), pantherpotorange 2,397, regurgitatorcorpse 6,955 of 7,103, spike1-4 2,272 / 2,306 / 2,272 / 2,351 (levels above 1). Row data (D): goospit1 Min/Max 1 / 3, MinLevDam1 3, HitShift 1, pois 4 x 200 frames; spike1 1 / 2, MinLevDam1 3, SrcDamage 128; pantherpotorange 3 / 5, MinLevDam1 5, fire 2-5; regurgitatorcorpse 25 / 55, MinLevDam1 15; mummy1-4 pois 9 x 100..400 frames; pantherpotgreen pois 9-12 x 400.

## The fix (local branch, two commits on 23e2baec)

1. `9d7e7d81`: `monster_shot_phys` and `monster_shot_stats` in `components/game/ai.hpp` (the lines of `ai.cpp:2014-2020` moved, `Ai::tick` calls `x.src = monster_shot_stats(...)`), `tests/game/test_monster_missile.cpp`. No behaviour change (BEFORE binary).
2. `312b0bbb`: `MissileInfo::min_lev / max_lev` (loader reads MinLevDam1..5 / MaxLevDam1..5, `gamedata_load.cpp`), `monster_shot_phys` in 256ths with `rules::level_bonus` and `<< HitShift`, a negative or missing SrcDamage no longer subtracts (a `-1` cell read as -1 before: no Act 1 / 2 row has one, D), `monster_shot_stats` puts the row's own element (`rules::row_damage`) into a free `elements` slot (poison: see the correction below, it goes in `poison_*`); plus `tools/emu/mm_diff.py` / `mm_launch.py`.

Limits (all disclosed in a PR body):
- `Ai::tick` stores `a2_min = phys >> 8`: a physical part under one point (the Sand Maggot spit, 2-6 256ths) is 0 whole points on d2d's roll (`monster_blow` rolls whole points inclusive; game.exe rolls 256ths top exclusive, D5 of sk-missile-trace). The RECORD is exact; the roll is not changed here.
- The row element takes a free slot of `MonStats::elements` (3 slots). A monster whose three El rows all fire in the shot's mode would lose the row element: no Act 1 / 2 shooter has more than two (D, `monstats.txt`; not run as a case: I).
- The impact path (`ai.hpp:793` `monster_blow(..., missile.src)`) is unchanged and was not re-run here: the claim is the record the missile carries, not the damage applied (block, resistance, roll: monster-attack-trace). `test_game` / `test_combat` of the branch were not run in this card (only the three targets built); the verifier must.
- Weapon rows (Blood Raven) and the skill-sourced rows (shamans, Unraveler bolts, Andariel, Summoner) are not on this path.
- Synergy calc (`EDmgSymPerCalc` in a monster row): no Act 1 / 2 row sets one (D); the fix does not evaluate it (game's `0064b100` does).

## Verifier findings and fixes (card missilefix, 2026-11-01)

The independent verifier (`docs/research/pr-missilerecord-verifier.md` on branch agent/verifymissile, verdict NOT READY) reproduced the record claim (0 of ~250,000 cases differ on the head; base 11,804 of 34,067 at levels 1 / 4 / 8) but found one blocker the author's "I" tag had left open: the head stored the row's poison as a whole-point total (`P x ELen >> 8`) in `MonStats::elements`, and `monster_blow` reads a type-3 element in the MonStats El convention (roll x 10 a frame for 2 x dur frames), so a hit dealt x 7.8 to x 31 of game.exe's record on the Act 2 rows (V run by the verifier: one Hell Night Slinger pot took 5,922 life in 20 s against about 165 points; Andariel's poison bolt x 62 through `missile_blow`, which upstream main has too).

| Verifier item | Fix (private fork, V = run on the worker host) |
|---|---|
| Poison units (blocker) | `c301eaef` on `local/missilediff-record`: the row poison goes into `MonStats::poison_min / poison_max / poison_len` (per frame in 256ths, frames: the record's +0x34 / +0x38 / +0x3c; `monster_blow` already applies these on every hit), added to an owner's own (U: the length is `max`, FUN_0064a930's +0x3c sum not traced; no Act 1 / 2 shooter has its own poison). Fire / cold / lightning / magic stay in `elements`. V: `test_monster_missile` (no arguments) runs the poison through the real `monster_blow`: goospit1 Normal 8 a frame for 200 frames, Hell 50 x 200; mummy4 Hell 88 x 400; pantherpotgreen Hell 102-108 x 400; it FAILS on the old head (assert on goospit1) and passes now |
| `missile_blow` (Andariel) | SPLIT, not in this candidate: `local/missilepoison-skillrows` (`70fa479d` on upstream main 6935cf64, one commit + `test_missile_blow_poison`). Reason: another consumer (`andariel_missile`: every poison skill missile of the shamans, Unravelers, Andariel, Summoner), a larger blast radius than the record claim, and it is wrong on upstream today, so it stands alone as a bug fix; the candidate's PR text names it as the sibling (presubmit's "siblings" WARN) |
| presubmit FAILs | `MonsterShotPhys { low, high }`; ruff 0.8.6 (presubmit's rules) clean on `mm_diff.py` / `mm_launch.py` (unused imports, loop variables, the B023 in `mm_launch.py` bound by default arguments); clang-tidy: the unused `combat.hpp` include is used now by the unit test, `s1` renamed `s1_shot`; only `outdoor.hpp` findings remain, which are not ours. `pr_guard.py 23e2baec HEAD` 0 FAIL / 0 WARN; `presubmit.py --quick` 0 FAIL, 4 WARN (siblings x2, docs-sync, skipped build: answered in the draft) |
| Base | see the PR draft: the roll-top pair 9ba8d369 / 23e2baec is neither upstream nor an open PR; the candidate rebases cleanly (`local/missilediff-record-up`) |
| D5 | stays out (Limits) |

Re-run on the fixed head `c301eaef` (V run):
- Author's differential (`mm_diff.py cmp` against the same 120,000 cases of `/tmp/md_full`, poison now per frame): physical 0 differ, element compared 55,769, 0 differ, at levels 1 / 4 / 8 2,560 cases 0 differ.
- Verifier's harness (his cases and game.txt, three sets: seed A, seed B, all acts 123,912 cases; poison compared per frame): physical 0 differ in every class (MissA rows and the skill-sourced rows), elements 0 differ, poison compared in about 22,000 (seed A: 8,095 + 11,236 MissA, 1,995 + 2,792 skill) cases per set with 0 differ. Through the real `monster_blow` (`VM_BLOW=1`): goospit1 8 / 50 a frame for 200 frames at levels 1 / 8, mummy4 18 / 88 for 400, pantherpotgreen 24..108 for 400.
- Det scenarios (lab binaries: the lab patch set + the roll-top pair, then + the candidate; one hidden slot, no game.exe): the 14 old scenarios (passive, fight, cold, static, cold2, hit_plain / champion / unique / boss / immune / minion, merc, pet, follow) hash the same on mon and state, base against head. The verifier's `spit_*` scenarios (player life lost in 500 frames, player poison resistance -100 in Hell, so the poison counts double): Sand Maggot Normal 22 -> 42 (old head 83), Hell 351 -> 452 (old 1,168); Night Slinger Hell 0 -> 359 (old 5,922; a pot's record 159-169 points, x 2 for the resistance); the other rows move only by the brackets and the sub-point physical.
- Suites: ctest of the full build (61 tests, `D2_MPQ_DIR` and `D2_PATCH_INSTALLER` set): `test_game`, `test_combat`, `test_game_rules`, `test_rules`, `test_monster_ai`, `test_monsters` pass; failures `test_mpq` (the integrity assertion, environment), `smoke_d2d` (no voice, no d2speech in this MPQ folder), `quests_act1` / `quests_act4` (a quake timeout; the same assertion on the lab binary WITHOUT the candidate, pre-existing); skipped `test_missile_hit`, `test_iso9660`. The verifier's base had 10 failures only because its tests-only build had no d2d binary.
- The rebased branch `local/missilediff-record-up` (`554067e0` + `4224a2be` + `ec09c73f` on upstream main 6935cf64, no conflict): the same author and verifier differentials (0 differ; its probe output is byte-identical to the unrebased head's), `test_game`, `test_combat`, `test_game_rules`, `test_rules`, `test_monster_ai`, `test_monsters` pass, pr_guard 0 / 0.

Not changed: the 14 deterministic scenarios do not fire an affected row (so identical hashes are expected); the poison length of an owner with its own poison hit stays a `max` (U).

## Deviations d2d vs game.exe

| D | Status | Rule game.exe | Rule d2d (before) | Tag | Differential |
|---|---|---|---|---|---|
| D1 | CONFIRMED, fixed on the branch | brackets `(level - 1) x MinLevDam1` (to 8, then the later steps) on min and max, in the physical | `missile_info.min` raw | V run both sides (44,456 differ before incl. D2, D3; spike alone 9,201 of 9,201 above level 1) | yes: this one |
| D2 | CONFIRMED, fixed | `(MinDamage + bracket) << HitShift`, 256ths | whole points | V run both sides (goospit level 1: 307 of 307) | yes |
| D3 | CONFIRMED, fixed | the row's element (poison per frame x length, fire) rides on the missile | MissA path carries only the monster's El rows | V run both sides (28,300 of 55,769 compared) | yes |
| D4 | unchanged, not touched here: 25 of 49 rows unblockable | | | V run (record, trace) + S | `mon_missile.py block` (existing) |
| D5 | unchanged: Blood Raven weapon | | | V run | none |
| D6 | unchanged: flag bit 9 | | | V read | none |

## Bot findings

None new. B1-B6 of monster-missile-trace stand; the record table the bot lacks (B3: `internal/combat/fight.go:492-507` dodges by distance, not damage; `internal/d2/missiles.go:45-57` names four kinds) can now be built from game.tsv with confidence that d2d, after the fix, deals the same record. Described, not changed.

## PR candidate

Local branch `local/missilediff-record` (private fork remote only; head `c301eaef` after the verifier's fix): proof above (before 44,456 / 120,000, after 0). Not for upstream until the independent verifier runs `mm_diff.py` with its own harness and `test_game`, `test_combat`, `pr_guard.py`, `presubmit.py --quick` (not run here). Draft text: this doc (Verdict, The differential, Limits). Upstream-PR-standard siblings to answer: the physical roll (D5), skill-sourced rows (`andariel_missile` already uses `row_damage`), traps (`objects.cpp:699`, same fields).

## Drill recipe (lab only; docs/forkdbg.md, docs/forkcheats.md; no game was run)

Lab binary built from the branch (`bin/d2d.lab-debug-skills` plus the patch) as `D2D_BIN`, one hidden slot, throwaway Sorceress `--no-save`, `d2drill -probe`. Nightmare / Hell game needed (U whether the lab has a difficulty verb).
- M1 D1: `debug spawn quillrat1 1`, 30 hits, life before / after (`dump stats` raw[6]): expected NM spike 19-26, Hell 44-61 after the fix; 10-17 / 23-40 before.
- M2 D2 / D3: `debug spawn sandmaggot1 1` in an Act 2 level, one spit: poison state from the row (6 points over 200 frames at level 1), `debug states`; before the fix none.
- M3 control: `cr_archer1` / `sk_archer1` arrows equal before and after in every difficulty.

## Matrix cell updates (proposed text)

| Row | Cell | New value |
|---|---|---|
| rule-coverage `missiles` 7-10, 63-66, 70-72, 238, 239, 341 | status | TESTED-DEVIATION, 44,792 cases, docs/research/missilediff.md D1-D3 |
| `missiles` 2-5, 13-16, 125-128, 196-200, 226, 227, 229, 266 | status | TESTED-0-DIFF, 72,829 cases (record, levels 1..40, clean-owner elements) |
| `missiles` 22-25, 32, 77-79, 130, 133, 135, 346, 653 | status | stays READ-ONLY (skill-sourced, weapon, no damage record) |
| monster-missile-trace D1-D3 | tag | V run both sides (this doc) |

## Dead ends

- First owner-stat rewrite wrote value << 8 (the getter shows points, not 256ths) and then wrote into a list entry that did not exist; the stat list of a monster with no damage in that mode has no 0x15 / 0x16 entry (bighead A2): the script inserts it in key order (entries 8 bytes {stat << 16, value}, count and capacity words at head + 4 / + 6) and grows a full list.
- First count cap read as a dword (786,444) scanned past the array: count is a word.
- Not run: `test_game`, `test_combat`, presubmit on the branch; the impact path; skill-sourced rows; the roll (D5).

## Reproduce (the worker host)

```sh
# d2d side (fork worktree, scratch build dir inside it; -j1 or -j2 while a soak runs)
cd ~/d2d-lab/fork/wt-missilediff/build && nice ninja -j2 test_monster_missile mpq-cat
# game side and compare
tools/emu/mm_diff_run.sh game $HOME/mmtab 120000 20261101 /tmp/md_full
D2_MPQ_DIR=$HOME/mmtab/emu D2_PATCH_INSTALLER=<LODPatch_114d.exe> test_monster_missile --monster-missile /tmp/md_full/cases.txt | grep -E '^[0-9]+ ' > d2d.txt
tools/emu/mm_diff_run.sh cmp /tmp/md_full d2d.txt        # BEFORE: build 9d7e7d81; AFTER: 312b0bbb
```

## Shared-file lines for the merger

In `openspec/changes/missilediff/merge-lines.txt`.
