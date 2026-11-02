# Re-verifier: local/missilediff-record-up and local/missilepoison-skillrows (REVERIFYMISSILE, 2026-11-01)

> status: current re-verifier record for the two private-fork candidates; verdicts in section 9, updated for heads 1824870b / 71f98145 in "Third check" (T5). Supersedes the verdict of `docs/research/pr-missilerecord-verifier.md` (first head, NOT READY) for the fixed heads; that doc stays the record of the first round.

Candidates (private fork remote only; nothing pushed by me, nothing posted): `local/missilediff-record-up` = upstream main 6935cf64 + 554067e0 / 4224a2be / ec09c73f (unrebased: 9d7e7d81 / 312b0bbb / c301eaef on the roll-top pair 23e2baec); `local/missilepoison-skillrows` = 70fa479d on 6935cf64. Author's evidence (`docs/research/missilediff.md`, `docs/research/pr-missilerecord-draft.md`, branch agent/missilefix at e9d5b3e2) read only after sections 3-7 were run. Offline: no game, no game instance, no lab binary, no det scenarios this round.

Tags: **V run** game.exe code run natively in MY oracle under upstream's emulator, or a real d2d build run with MY probe; **V read** disassembled by me (`tools/emu/disasm.py`) or source read at the named commit; **D** 1.14d tables; **S** stated, not re-traced; **U** not traced; **O** opinion.

## 1. Verdicts in one paragraph

**Record branch (`local/missilediff-record-up`, ec09c73f): READY WITH CHANGES.** The record claim holds on c301eaef and ec09c73f with the poison read per frame (0 differences in physical, elements and poison over two new seeds; V run), and the consumer now applies exactly the record: through the real `monster_blow`, 0 of 18,092 (seed C, at four defender settings) and 0 of 22,743 (all acts) poison hits differ from game.exe's record, and game.exe's own resist stage (FUN_0057c1e0, run natively) equals `monster_blow`'s per-frame and length arithmetic in all 5,926 in-domain cases (V run). The change needed is the owner-poison add: game.exe does **not** add the owner's poison on any of the 32 poison rows (all have SrcDamage 0, and FUN_0064b860 calls FUN_0064a930 only when SrcDamage != 0), and where it does add (SrcDamage != 0) it **sums** the length and scales the rate by SrcDamage / 128 (V read + V run). The head adds the owner's poison unscaled and takes the max length: wrong for Baal Subject 2 (unraveler5: mummy4 + poisonhit) and any poisonhit shooter of a poison row. The PR text's "No Act 1 / 2 MissA row has an owner poison" sidesteps that, and its "independent counts" paragraph was written before this re-verification (section 8). **Skill branch (`local/missilepoison-skillrows`, 70fa479d): READY WITH CHANGES.** It is sound and minimal: `missile_blow` goes from x 18 to x 77 too strong on upstream main (Andariel alone; V run) to 0 differences in 12,586 hits. But two sibling writers with the same unit mistake remain (`objects.cpp` traps, `deaths.cpp` death clouds), and the PR must name them.

## 2. My harness, extended (worker host, `~/d2d-lab/verifymissile/harness/`)

Reused from round 1 (my own; written from the disassembly): `vm_game.py` (game.exe's FUN_0064b860 record, native), `vm_gen.py`, `test_vm_probe.cpp`. Added this round:

- `vm_game2.py`: `vm_game.py` plus the owner's poison stats 0x39 / 0x3a / 0x3b fed to the stat getter, and a SrcDamage override written into the row (+0x12d) for synthetic cases.
- `test_vm_probe2.cpp` (installed as `tests/game/test_vm_probe.cpp` in each tree): prints `MonStats::poison_*` (`P min max len`) and takes an owner poison triple. It adds a `--skill` mode (row_damage -> `missile_blow`, `andariel_missile`'s seam), a `--dot` mode (the real `monster_blow` on a defender with a given `res[3]` / `plr`), and `VM_RES` / `VM_PLR` for the shot hits. It reads the physical pair with a structured binding, so one file builds on both the `lo/hi` and the `low/high` trees. `VM_NO_SHOT` covers trees without the seam.
- `vm_dot.py` (new oracle): game.exe's resist stage for a poison hit on a PLAYER. It runs FUN_0057c1e0 natively on a damage record with +0x28 (poison per frame, 256ths) and +0x2c (frames) set, together with FUN_0057bf80 over the table at 0x732980, FUN_0057be00, FUN_0057bf10, MulDiv and FUN_00611d30 on the booted difficulty rows. Hooked: the stat getter, the hireling predicate, FUN_0057c060 (PvP / hireling %, 100), events, FUN_0057c140 (cold / freeze lengths) and logging.
- `vm_cmp_s.py` (element and poison compare), `vm_hit.py` (what the hit applies vs the record), `vm_dotgen.py`.
- Cases (new seeds, none of round 1's): C = Act 1 / 2 rows, `vm_gen.py rows.txt 60000 90210` (120,200 cases, md5 d375ead79ba9); D = all 78 monster Miss rows, seed 5150 (123,912, 1d4b0716f618); OP = 17 rows x levels 1 / 4 / 8 / 20 x 4 owner poisons x SrcDamage {row's own, 128, 64, 255} x A1 / A2 (2,176, 726246508bae); S = the 32 poison rows + the 15 skill-sourced rows, levels 1..40 / 50 / 60 / 99 + 20,000 random (seed 4711; 21,892, a248925c805c); DOT = per x len x res x plr grid + 20,000 random (seed 31337; 24,890, 8fc29bbc00bf).

Process note (O): the author wrote `vm_cmp_p.py` into my harness directory and ran my harness on the fixed head. The draft then cites those runs as "Independent counts (the verifier's own harness ...)", but they were the author's runs. Every count in this doc comes from my own runs, and the poison counts use my new `vm_cmp_s.py` / `vm_hit.py`. The author's `vm_cmp_p.py` agrees with them on seed C.

## 3. The record claim, poison per frame (V run)

| set, build | cases compared | physical differ | other elements compared / differ | poison compared / differ |
|---|---|---|---|---|
| C, c301eaef | 108,140 (2,408 on rows d2d does not load, 9,652 skill-path rows set aside) | 0 | 16,903 / 0 | 23,884 / 0 |
| D all acts, c301eaef | 112,760 (4,858 not loaded, 6,294 skill path) | 0 | 7,916 / 0 | 23,926 / 0 |
| C, ec09c73f | 108,140 | 0 | 16,903 / 0 | 23,884 / 0 |
| D, ec09c73f | 112,760 | 0 | 7,916 / 0 | 23,926 / 0 |

The ec09c73f probe output (C, D and OP sets, hits included) is byte-identical to c301eaef's on every numeric line; the hit compare gives 0 of 18,092 (C) and 0 of 22,743 (D) on ec09c73f too.

Rebase: `git range-diff 23e2baec..c301eaef 6935cf64..ec09c73f` shows the three patches identical (`=`). The tree differences are upstream's own (Monster / Foe / Missile fields), none in the shot functions (V run). `git merge-tree ec09c73f 70fa479d`: clean (V run). The author's unit test (`test_monster_missile` with no arguments) passes on c301eaef and on ec09c73f (rc 0).

## 4. The consumer: what a hit applies (V run, V read)

**Path in game.exe** (V read), step by step:

1. Record +0x34 / +0x38 / +0x3c -> missile stats 0x39 / 0x3a / 0x3b, set as they are (FUN_0064ac60).
2. At the hit, the damage record gets the poison (code at 0x57bb60..0x57bbee; S that this builder is the one a missile hit takes):
   - +0x28 = a roll in the stats' range (FUN_0057a880 via FUN_0045c3e0) x a byte / 128;
   - +0x2c += stat 0x3b, or stat 0x65 when > 0, divided by stat 0x146 when > 1.
3. FUN_0057c1e0 resists +0x28 by poison resist (stat 0x2d, max 0x2e, pierce 0x150; table entry 8) and +0x2c by poison length reduction (stat 0x6e; entry 7).
4. FUN_0057ac50 sets state 2 for +0x2c frames with hp regen (stat 0x4a) = -(+0x28), and keeps an existing poison whose rate is higher.

So the poison is +0x28 256ths a frame for +0x2c frames, which is d2d's model.

**Resist stage, game.exe vs `monster_blow`** (DOT set; my `vm_dot.py` vs the probe's `--dot`):

| domain | cases | differ |
|---|---|---|
| Normal, poison res -100..75, PLR 0..75, no max-res bonus (what d2d's Fighter holds: effective values) | 5,926 | **0** (per frame and length, so the total too) |
| outside it (res > 75 or < -100, PLR > 75, Nightmare / Hell, max-res stat) | 18,964 | differ by design: there game.exe applies the floor -100, the cap of 75 + max-res, the PLR cap of 75 and the difficulty penalty itself, which d2d does in the Fighter builder |

Penalty as game.exe applies it (V run): Nightmare -40 and Hell -100, on the poison resistance AND on the poison length reduction. A record of 88 x 400 on a player with 0 / 0 becomes 123 x 560 in Nightmare and 176 x 800 in Hell (x 4 in all); with PLR 50 it is 176 x 600 in Hell. d2d's player Fighter folds the penalty into both values (`fight/strike.cpp:103`: plr = stat 110 + penalty, clamped to -100..75; the res side: S, not re-read), so the in-domain identity is the right test.

**Shot hits** (real `monster_blow`, record from game.exe, `vm_hit.py`). To-hit is capped at 95 %, so about 5 % of the lines are misses, and those are skipped:

| set, defender res / plr | poison hits | differ (per frame outside game's range, or ticks != game's length) |
|---|---|---|
| C (MissA rows), 0 / 0, -100 / 0, 40 / 25, 75 / 50 | 18,092 each | 0, 0, 0, 0 |
| D (all acts), 0 / 0 | 22,743 | 0 |

Hell examples (head, res -100; game.exe's length also doubles through the PLR penalty):

- goospit1 level 8: record 50 x 200 -> 100 a frame.
- mummy4 level 8: 88 x 400 -> 176 a frame.
- pantherpotgreen level 8: 102-108 x 400 -> 210 a frame.

A full Hell pot on a player with no PLR is 204-216 x 800 = 637-675 points. The author's det drill counts 500 frames, so its 359 comes from a truncated window, not from "x 2 for the resistance".

**Roll range (low, pre-existing, not this PR):** FUN_0057a880 rolls `min + rand(max - min)` (rand(n) = 0..n-1, FUN_0045c3e0), so game.exe never reaches the poison max. `monster_blow` rolls `rng.range(min, max)` inclusive (rules.hpp:172). My runs show it: pantherpotgreen level 1 rolled 24 of 18..24. It is the same top-of-range issue as the roll-top pair (U that this builder is the missile-hit path).

## 5. The owner-poison add (`max` length): traced (V read, V run)

FUN_0064b860 (V read; round-1 disassembly re-read) in order:

1. The row element is stored with `mov` (the record slots are set).
2. **Only if SrcDamage (+0x12d) != 0**, FUN_0064a930 adds the owner's stats:
   - `+0x34 += stat 0x39` and `+0x38 += stat 0x3a`;
   - `+0x3c += stat 0x65` if > 0, else `+0x3c += stat 0x3b` and `+0x40 += stat 0x146`.
3. FUN_0064aa60 then scales +0x34 / +0x38 (and the rest) by SrcDamage / 128, truncated toward 0, but **not +0x3c**.

With SrcDamage 0 and +0x12e (SrcMissDmg) set, the source missile's stats are added instead (not a MissA case).

The OP set through `vm_game2.py` confirms it (V run). goospit1 level 1 with an owner poison of 40 / 40 / 300:

| row's SrcDamage | game.exe record | head |
|---|---|---|
| 0 (real) | 8 / 8 / 200 (owner dropped) | 48 / 48 / 300 |
| 128 (patched) | 48 / 48 / 500 (length summed) | |
| 64 (patched) | 24 / 24 / 500 | |
| 255 (patched) | 95 / 95 / 500 | |

Over the 272 OP cases with the rows' real SrcDamage, the head differs in **180 of the 204 with an owner poison**, all on SrcDamage-0 poison rows. The 24 that match are bighead1 / spike1 (SrcDamage 128, no row poison) (V run).

Data (D): all 32 poison rows of Missiles.txt have SrcDamage 0. Only `skmagepois` has 128, and it has no row element. Which shooters of a poison row carry `MonStats::poison_*`? `poison_hit` (ai.cpp:2907) fills it for umod 23 (poisonhit):

- **Baal Subject 2** (superunique, unraveler5, Mod1 23, MissA mummy4), through `make_boss` (V read + D).
- The Maggot Queen class (act_boss 0x11c: poisonhit, ai.cpp:146-151), when spawned as a non-boss. Whether Coldworm the Burrower (a superunique of maggotqueen1 with goospit1) gets it too: U.

So the risk is real but narrow. Those shooters' spit carries the owner's poison hit plus the row's, with the owner's length, where game.exe carries the row's alone. The owner's poison riding on a SrcDamage-0 shot is pre-existing (base copies `monster.stats` into the shot, poison included); the head turns it into owner + row.

## 6. Readers of `elements` and `poison_*` across the tree (V read, at c301eaef)

- `MonStats::poison_*`: the only reader is `monster_blow` (combat.hpp:609). Writers: `poison_hit` (ai.cpp:2912-2914), the new `monster_shot_stats` and the skill branch's `missile_blow`. `Fighter::poison_len` (states.cpp, pets.cpp) belongs to another struct. Nothing assumed the old total.
- `elements` readers: `monster_blow` (combat.hpp:595, the El convention: roll x 10 a frame for 2 x dur frames) and `swing_max_damage` (monsters.hpp:287, "poison max x 10 >> 8"). Both read the El convention, which the record branch no longer writes for poison. So no consumer assumed the old total; it existed only in 312b0bbb and in `missile_blow`.
- Writers of a type-3 `elements` entry that is NOT in the El convention (siblings, pre-existing on upstream, in neither branch):
  - `world/objects.cpp:706-707` (trap missiles, `kTrapMissile`): `pts()` stores `P x ELen >> 8` exactly as the first head did, so it is x 20 x ELen / 256 too strong (trappoisonjavcloud, ELen 500: x 39) (V read; the arithmetic of round 1, section 5).
  - `fight/deaths.cpp:135` (`boss_missile`: the mummies' `corpsepoisoncloud` on death): `elo / 10` with `dur = elen`. The rate is right to within 9 / 256, but the length doubles (2 x ELen frames): x 2 (V read).
  - `ai.cpp:434` (a monster melee skill's added element, "A1"): `elo >> 8` for any etype, so a poison skill would be x 10 / 256 x 2. U whether any monster attack skill has EType pois.
  - `ai.cpp:2404` (kDiabCold): cold only, fine.

## 7. The Andariel split (`local/missilepoison-skillrows`, 70fa479d)

- **Diff** (V read): `missile_blow` puts a type-3 element into `poison_min / poison_max / poison_len` as they are (per frame, frames). Other elements stay as `>> 8` in `elements[0]`, byte-identical to before (`elo x 1 >> 8`). One commit, two files (ai.hpp, `tests/game/test_missile_blow_poison.cpp`). Callers covered: `andariel_missile` (ai.cpp:380: every skill missile of Andariel, the shamans, Unravelers, the Summoner, the vampires ...) and `missile_from` (ai.hpp:708, sub-missiles). Minimal: yes.
- **The x 62 on upstream main is real** (V run: my probe on 6935cf64, `--skill`, defender 0 / 0, against game.exe's record). Every poison row through `missile_blow` is stored as an `elements` entry (13,768 of 18,312 cases), and 13,272 of 13,272 poison comparisons differ. What the hit applies, against the record:

  | row, level | game.exe record | upstream main's hit | ratio to the record max |
  |---|---|---|---|
  | andypoisonbolt, 1 | 32-64 x 800 (100-200 points) | 1,130-1,520 x 1,600 (7,063-9,500 points) | x 35-48 |
  | andypoisonbolt, 8 | | | x 62-66 |
  | andarielspray | | | x 18-44 |
  | undeadmissile1 | | | x 26-27 |
  | necromage1 (ELevLen 250) | | | x 37 at level 1, x 410 at level 20 (whether a monster fires it through this path: U) |

  The fork's c301eaef tree gives the same, because it does not contain the split.
- **After the split** (70fa479d, same S set, V run): poison compared 13,272, **differ 0**; no poison left as an `elements` entry; other elements 3,018 compared, 0 differ. Through the real `monster_blow`: 12,586 poison hits, **0 differ** from the record (939 misses at the 95 % cap skipped). The author's `test_missile_blow_poison` passes (rc 0).
- **Complete?** For `missile_blow`, yes. For "monster poison written in the wrong units", no: the two siblings of section 6 (traps, death clouds) remain. Name them in the PR body, or fold `objects.cpp` in: it has the same one-line shape.
- Independent of the record branch: no shared hunk, and the merge is clean. The record PR does not need it, but both edit `ai.hpp` near each other, so whichever merges second should re-run `test_monster_missile` and `test_missile_blow_poison`.

## 8. Suites, presubmit, the PR text

**Suites (V run):** RESULT_SUITES

**pr_guard / presubmit (V run, ruff via uvx):**

| branch | pr_guard vs 6935cf64 | presubmit --quick vs 6935cf64 |
|---|---|---|
| record-up ec09c73f | 0 FAIL, 0 WARN | 0 FAIL, 4 WARN: siblings (unit kinds), siblings (hit paths), docs-sync, build skipped; ruff 2 files clean, short-names ok |
| skillrows 70fa479d | 0 FAIL, 0 WARN | 0 FAIL, 2 WARN: docs-sync, build skipped |

The author's answers to the 4 WARNs:

- **build**: answered by the ctest runs.
- **hit paths**: partly answered. The skill-sourced sibling is named; traps and death clouds are not.
- **unit kinds**: not answered in the body. The answer: player, merc, pet and co-op partner are all hit through the same `monster_blow` with their own Fighter, so the units hold for all of them (the merc's and pet's `plr` / res penalty: S).
- **docs-sync**: not answered. Add a line in `docs/research/re/`, or a body line saying no doc covers this.

**PR text claims vs counted runs** (draft `pr-missilerecord-draft.md` at e9d5b3e2):

| claim | supported? |
|---|---|
| record 0 differ on c301eaef / ec09c73f (the author's 120,000) | yes in kind: my C / D sets give 0 (section 3) |
| "Independent counts (the verifier's own harness ...) re-run on the fixed head ... 0 differences ... on both heads" | **overclaim**: written before any re-verification, and those were the author's runs of my harness. Replace with this doc's counts and a reference |
| consumer numbers for goospit1 / mummy4 / pantherpotgreen equal the record | yes (section 4), at res 0 / plr 0 and at three other defender settings |
| "pantherpotgreen 102-108 for 400 (159-169)" | yes at level 8 in Normal arithmetic; in Hell the player's total is x 4, not x 2 (section 4) |
| "drill ... player poison resistance -100: ... now 359" | number not re-run by me (S). Its explanation "x 2 for the resistance" (missilediff.md) is incomplete: the length doubles too, and the 500-frame window cuts it |
| "pr_guard 0 / 0; presubmit 0 FAIL" | yes (V run) |
| test_game, test_combat, test_game_rules, test_rules, test_monster_ai, test_monsters pass | RESULT_SUITES_SHORT |
| Limits: "No Act 1 / 2 MissA row has an owner poison" / "the length is the larger (FUN_0064a930 +0x3c sum not traced)" | **wrong, and now traced**: game.exe adds no owner poison on SrcDamage-0 rows (every poison row); where it does add, it sums the length. Baal Subject 2 is affected (section 5) |
| test_mpq / smoke_d2d / quests_act1 / quests_act4 fail for environment / pre-existing reasons | RESULT_FAIL_SHORT |
| "MissS1 / MissC / MissSQ never fire", imp double fire, SrcDamage -1 | yes (round 1, V run / V read) |

## 9. Verdicts and the exact changes

**`local/missilediff-record-up` (ec09c73f): READY WITH CHANGES.**

1. `monster_shot_stats`: on a SrcDamage-0 row (every poison row), do not add the owner's `poison_*`; clear it before adding the row's. Where SrcDamage != 0, scale the owner's per-frame poison by SrcDamage / 128 and SUM the lengths (FUN_0064a930 / FUN_0064aa60). Replace the "U: ... not traced" comment at ai.hpp:696 with the traced rule. Extend the unit test's `champion` case: goospit1 with an owner poison of 40 / 40 / 300 must give 8 / 8 / 200. If this is kept out of the PR instead, Limits must say that the owner's poison rides on SrcDamage-0 shots (pre-existing) and that Baal Subject 2's spit is owner + row instead of the row alone.
2. PR text: drop the "Independent counts" paragraph or replace it with this doc's numbers; fix the owner-poison Limits line; name the trap (`objects.cpp:706`) and death-cloud (`deaths.cpp:135`) siblings; answer the unit-kinds and docs-sync WARNs; say a Hell pot is x 4 in total, not x 2.
3. Not required: a known-gap line for the top of the poison roll (section 4).

**`local/missilepoison-skillrows` (70fa479d): READY WITH CHANGES.**

1. PR body: state that the x 62 is real on upstream main (section 7 numbers); name the two remaining siblings (traps x 20 x ELen / 256, death clouds x 2), or fold the trap one in (same shape, `objects.cpp:706-707`).
2. A merge-order note with the record PR: both edit ai.hpp next to each other, so re-run both unit tests on whichever merges second.

## Reproduce (worker host)

```sh
cd ~/d2d-lab/verifymissile/harness
python3 vm_gen.py rows.txt 60000 90210 > casesC.tsv; python3 vm_gen.py allmiss.txt 30000 5150 > casesD_all.tsv; python3 vm_dotgen.py 31337 20000 > dot_cases.tsv
cd ../head/tools/emu && export D2_MPQ_DIR=<MPQ dir with bin/game.exe> D2_PATCH_INSTALLER=<LODPatch_114d.exe> D2_BUILD=<build with mpq-cat> PYTHONPATH=$PWD
.venv/bin/python ../../../harness/vm_game2.py game ../../../harness/casesC.tsv > ../../../harness/game_C.txt     # also casesD_all, cases_op, casesS
.venv/bin/python ../../../harness/vm_dot.py ../../../harness/dot_cases.tsv > ../../../harness/dot_game.txt
cd ../../../harness && VM_BLOW=1 ../head/build/tests/test_vm_probe casesC.tsv > rvhead_casesC.txt && ../head/build/tests/test_vm_probe --dot dot_cases.tsv > rvhead_dot.txt
python3 vm_cmp_s.py casesC.tsv game_C.txt rvhead_casesC.txt; python3 vm_hit.py casesC.tsv game_C.txt rvhead_casesC.txt 0 0
VM_BLOW=1 ../base/build/tests/test_vm_probe --skill casesS.tsv > rvup_skill_r0.txt && python3 vm_cmp_s.py casesS.tsv game_S.txt rvup_skill_r0.txt
```

## Left on the worker host

RESULT_LEFT

## Third check (2026-11-02, narrow: heads 1824870b and 71f98145)

Scope as briefed: only the owner-poison commit, the trap commit, the merge of the two, and the two PR drafts (`agent/missilefix2` at 33b1c4e2: `pr-missilerecord-draft.md`, `pr-missilepoison-draft.md`). The rest of sections 3 to 7 is not re-derived. My oracle files (`vm_game.py`, `vm_game2.py`, `vm_dot.py`) and all case files are unchanged. New on the worker host: `test_vm_probe3.cpp` (my own copy of the probe change: column 16 of a case line, the SrcDamage override that `vm_game2.py` writes into the row's +0x12d, now also goes into the probe's copy of `MissileInfo::src_damage`; probe2 ignored it), `cases_trap.tsv`, `rv3_*.sh`. Builds -j1 / -j2, niced, capacity WARN (allowed).

### T1. The owner-poison rule on 1824870b (V run)

| set (probe3 on 1824870b vs my game.exe record) | compared | differ |
|---|---|---|
| OP, poison (2,176 cases: 544 at the rows' real SrcDamage, 1,632 with SrcDamage patched to 128 / 64 / 255) | 2,112 | **0** |
| OP, physical | 2,176 (544 real, 1,632 patched) | 0 |
| OP, poison, the previous head ec09c73f (same oracle) | 2,112 | 1,776 |
| C: physical / other elements / poison | 108,140 / 16,903 / 23,884 | 0 / 0 / 0 |
| D (all acts): physical / other elements / poison | 112,760 / 7,916 / 23,926 | 0 / 0 / 0 |
| C hits at res / plr 0/0, -100/0, 40/25, 75/50; D hits at 0/0 | 18,092 each; 22,743 | 0 |
| DOT (resist stage), in domain | 5,926 | 0 |

For C, D, the four C hit runs and DOT, every numeric output line of 1824870b is byte-identical to ec09c73f's; only the log timestamps differ. So the new rule changes nothing in the record claim sets (they carry no owner poison).

**Non-poison rows.** In game.exe the gate is SrcDamage, not the row's element type. FUN_0064b860 calls FUN_0064a930 whenever +0x12d != 0, for any row (V read, section 5). So applying the rule to every row is game.exe's behaviour. The non-poison rows in OP (bighead1 and spike1: SrcDamage 128, no row poison) match the oracle at their real SrcDamage and also patched to 64 and 255, with and without an owner poison (0 differ; V run). Compared with ec09c73f, at the rows' real SrcDamage only the 360 poison-row cases that carry an owner poison change. The 120 poison-row cases without an owner poison, and all 64 bighead1 / spike1 cases, give the same output. In real play there is one change on non-poison rows: an owner's poisonhit is now dropped on a SrcDamage-0 non-poison row as well, where before it rode along. That matches game.exe. Which shooters it touches beyond Baal Subject 2 and the Maggot Queen class: S.

**SrcDamage -1 (low, new).**
- The new gate tests `info.src_damage != 0`, so a -1 row scales the owner's and the row's poison by -1 / 128. With probe3 and a -1 override:
  - goospit1 with an owner poison of 40 / 40 / 300 gives poison 0;
  - bighead1 with an owner poison of 200 gives `poison_min = poison_max = -1`.
- game.exe stores -1 as the byte 0xff and reads it as 255. My 255 override gives 95 / 95 / 500 for the goospit1 case (V run: `vm_game2.py` writes `src & 0xff`).
- `monster_shot_phys` treats `src_damage <= 0` as 0, so the two halves of the same function disagree.
- Not reachable through a monster Miss row: only poisonjavcloud, plaguejavcloud and furylightning have -1, all player skill missiles (D).
- Fix: gate on `info.src_damage > 0`, as `monster_shot_phys` does and as the draft's "treated as 0 here" says. Otherwise the text must say the poison half scales by -1 / 128.

U, as the author also says: whether a cast (MissC) shot takes this path.

### T2. The diffs against my previous heads (V read, V run)

- **`local/missilediff-record-up`**: parent of 1824870b = ec09c73f. One commit, two files: `components/game/ai.hpp` (`monster_shot_stats`, 20 lines) and `tests/game/test_monster_missile.cpp`. Nothing else changed. The order in the code: clear the owner's poison when SrcDamage is 0, add the row's poison (lengths summed), then scale the pair by SrcDamage / 128 with int64 truncation toward 0. That is section 9 change 1, and the oracle agrees: (8 + 40) x 64 / 128 = 24, not 8 + 20. `test_monster_missile` passes on 1824870b (rc 0).
- **`local/missilepoison-skillrows`**: parent of 71f98145 = 70fa479d. One commit, three files:
  - `ai.hpp`: `put_missile_element` is the old `missile_blow` body moved out unchanged.
  - `world/objects.cpp`: the `pts()` lambda is replaced by the helper. Non-poison trap elements are unchanged (`value >> 8` before and after). Poison now goes to `poison_*` per frame.
  - the test.
  `missile_blow` is unchanged in effect: my S-set probe output on 71f98145 is byte-identical to 70fa479d's on every numeric line. Poison 13,272 compared, 0 differ; other elements 3,018 compared, 0 differ; hits through `monster_blow` 12,586, 0 differ. (Section 7's "939 misses" is the row-filtered `vm_hit.py` count; unfiltered there are 686.) `test_missile_blow_poison` passes on 71f98145 (rc 0).
- **The trap path (`objects.cpp:706`).**
  - Correction to my own section 6: the trap that d2d fires is `kTrapMissile[3]` = **primepoisoncloud** (rules/shrines.hpp:284), not trappoisonjavcloud. No trap entry names trappoisonjavcloud. Both drafts, the commit message of 71f98145 and its test comment copied my "trappoisonjavcloud, ELen 500, x 39".
  - primepoisoncloud has no Skill, so `spring_trap` takes `row_damage(...)` at `kTrapLevel` 1 / 4 / 8 and then the helper. Upstream's old store was x 20 x ELen / 256 too strong: x 23 / 28 / 34 at those levels (ELen 300 / 360 / 440; arithmetic on the old `pts()` line, V read).
  - My check: `vm_game2.py` (game.exe's FUN_0064b860 row path, zero owner) on primepoisoncloud, trappoisonjavcloud, trapfirebolt and trapnova at levels 1..40 / 50 / 60 / 99 (172 cases), against probe3 `--skill` on 71f98145. `--skill` is the same `row_damage` call followed by the same `put_missile_element`.
  - Result: poison 86 compared, 0 differ; other elements 86 compared, 0 differ; hits through `monster_blow` 80, 0 differ (6 misses). primepoisoncloud is 10 x 300 at level 1, 25 x 360 at level 4 and 45 x 440 at level 8, in game.exe and on the head.
  - So the trap record arithmetic equals game.exe's FUN_0064b860 row path: **V run**.
  - **U**: that game.exe's object-trap launcher reaches the row path at that level with nothing added. My oracle uses a zero-stat monster owner, not a trap object.
  - **U**: whether a cloud missile (pSrvDoFunc 3) hits more than once.
  - `spring_trap` itself is not executed by any test: the unit case calls the helper only. Its call is V read.
- Low (readability): the helper's comment sits under `missile_blow`'s old comment block. That block ("...up to 62x as much") now reads as the helper's description.

### T3. Merge, suites, guards (V run)

**Merge.** 1824870b with 71f98145 merges cleanly (one auto-merge in `ai.hpp`; throwaway commit 2fd6d693 in my `mid` tree, not pushed). Built with -j2 at the test targets: `test_game`, `test_combat`, `test_game_rules`, `test_rules`, `test_monster_ai`, `test_monsters`, `test_monster_missile`, `test_missile_blow_poison` and my probe (rc 0). Each passes under ctest on the merge (8 of 8). The apps and crashpad were not rebuilt; the merged code is in the `game` library, which these targets build. On the merge, my probe output is byte-identical to the single heads' on every numeric line: OP and C against 1824870b, S and trap against 71f98145. So neither branch changes the other's results.

| branch | pr_guard vs 6935cf64 | presubmit --quick vs 6935cf64 |
|---|---|---|
| record-up 1824870b | 0 FAIL, 0 WARN | 0 FAIL, 4 WARN (unit kinds, hit paths, docs-sync, build skipped) |
| skillrows 71f98145 | 0 FAIL, 0 WARN | 0 FAIL, 3 WARN (hit paths, docs-sync, build skipped) |

Housekeeping: my round-2 lane `rv_run4.sh` (the upstream ctest baseline) was still running under my agent name, in the `base` tree I needed. I stopped it with `agentreap.sh reverifymissile`. Its partial log is kept as `logs/rv-up-ctest.partial.log`. Section 8's `RESULT_*` placeholders were never filled by round 2. They are left as they are; T3 holds this round's suite numbers. Left on the worker host from this round, nothing running (`agentreap.sh --list` is empty for me):

- `verifymissile/{head,base,mid}` detached at 1824870b / 71f98145 / 2fd6d693, each with my untracked probe;
- `harness/test_vm_probe3.cpp`, `cases_trap.tsv`, `game_trap.txt`, `rv3*` outputs and `rv3_*.sh`.

### T4. PR draft claims vs counted runs

**`pr-missilerecord-draft.md`**

| claim | verdict |
|---|---|
| owner-poison rule and its numbers (8 / 8 / 200; 48 / 24 / 95 x 500) | yes (T1, my oracle and my probe copy) |
| "owner-poison set ... 0 differences in poison (2,112 compared)" on 1824870b | yes (T1) |
| "the head before it differed in 1,416 of them" | **does not reproduce**: 1,776 of 2,112 with my oracle on ec09c73f. The old head ignores `src_damage` for poison, so a patched probe cannot change that count. Use 1,776 |
| re-verifier counts on ec09c73f; the earlier commits' output "unchanged on every numeric line" | yes (T1: byte-identical on C, D, hits, DOT) |
| Hell pot x 4 in total; the drill number is S | yes (section 4) |
| "a SrcDamage -1 (game.exe: 255, treated as 0 here)" | **half true**: physical yes, poison no (T1). Fix the code (`> 0`) or the text |
| "trap missiles (... trappoisonjavcloud x 39)" under Not in this PR | **wrong row**: primepoisoncloud, x 23 / 28 / 34 at trap levels 1 / 4 / 8 (T2) |
| ctest 7 of 7 on 1824870b | `test_monster_missile` rc 0 on 1824870b (mine); the six suites: see T3 (I ran them on the merge, not on 1824870b alone) |
| "Third check (narrow, owed ...)" section | superseded by this section: replace it with a pointer here |
| rows that moved on base; the first verifier's 11,804 / 34,067 | S (not re-run this round) |

**`pr-missilepoison-draft.md`**

| claim | verdict |
|---|---|
| x 62 table on upstream main; 13,272 / 3,018 / 12,586 with 0 differ after the split | yes (section 7; re-run on 71f98145, T2) |
| "trappoisonjavcloud (ELen 500) was x 39" (Siblings, Proof, the commit message, the test comment) | **wrong row**: say primepoisoncloud, x 23 / 28 / 34 at levels 1 / 4 / 8. The test values can stay (they test the helper), but its comment names the wrong row |
| "U that the trap record has no further difference (no live or game.exe differential was run on traps)" and Limits "a trap in game.exe" not run | now partly run: the row-path record arithmetic is V run, 0 of 86 (T2); the launcher remains U. Update the line |
| `test_missile_blow_poison` "covers" the trap | **overclaim (low)**: it covers the helper, not `spring_trap` |
| pr_guard 0 / 0, presubmit 0 FAIL 3 WARN | yes (T3) |
| merge clean, both unit tests pass on the merge | yes (T3: clean, 8 of 8 suites pass, outputs identical to each head) |
| death clouds x 2, melee-skill element U, named as siblings | consistent with section 6 |

### T5. Verdicts

**`local/missilediff-record-up` (1824870b): READY WITH CHANGES.** The code matches game.exe on every set (T1). Changes before the user's go:
1. Draft: replace 1,416 with 1,776 (T1).
2. Draft: the trap sibling is primepoisoncloud, x 23 / 28 / 34 at trap levels 1 / 4 / 8, not trappoisonjavcloud x 39.
3. Draft: replace the "Third check (narrow, owed)" section with a pointer to this section.
4. SrcDamage -1: either gate the poison on `info.src_damage > 0` (a one-token code change; then `test_monster_missile` and my OP set must be re-run, and they should still give 0) or say in the draft that the poison half scales by -1 / 128. Recommended: the code change, to match `monster_shot_phys`. Not blocking: no monster Miss row has -1.

**`local/missilepoison-skillrows` (71f98145): READY WITH CHANGES** (text only; the code is right, T2).
1. Commit message, test comment and draft: name primepoisoncloud (x 23 / 28 / 34 at trap levels 1 / 4 / 8) instead of trappoisonjavcloud x 39.
2. Draft: the trap record arithmetic is now V run against game.exe's row path (0 of 86 poison, 0 of 80 hits). The trap launcher and the cloud's hit count stay U, and the unit test covers the helper, not `spring_trap`.
3. Optional: move the helper's comment so `missile_blow`'s old comment stays with `missile_blow`.

Merge order: no constraint (T3).
