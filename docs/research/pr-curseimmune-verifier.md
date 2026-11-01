# Independent verifier: fix/curse-immune (a curse or enemy aura lowering an immune monster counts a fifth) (VERIFYCURSEIMMUNE, card verifycurseimmune, 2026-11-01)

> status: verifier record for the local upstream candidate `fix/curse-immune` (claim 6a04cb3e, tooling 557d14e3, base upstream main 72132b99) on the worker host. Author's proof and draft: `docs/research/pr-curseimmune-proof.md` / `pr-curseimmune-draft.md`; trace `docs/research/sk-resist-trace.md` (D1). Offline; no game, no d2d slot; nothing pushed or posted; the author's branch and worktrees untouched. Verdict in section 7.

Tags: **V read** disassembled by me with upstream's `tools/emu/disasm.py`; **V run** game.exe code run under upstream's emulator in my own oracle, or the real d2d build run with my own harness; **D** the patched 1.14d tables read from the MPQ stack by my oracle; **S** d2d source; **I** inferred; **U** not traced.

Order of work: the claim diff and the disassembly first (section 1), then my own generator, oracle and harness (section 2), then the author's proof and draft. I did not open the author's `tools/emu/sk_resist.py` or `tests/game/test_sk_resist.cpp`.

## 1. The rule in game.exe (my reading)

1. **FUN_005c3540** (ECX unit, EDX stat, one stack value; `ret 4`) (V read 0x5c3540..0x5c359b): a byte table at 0x5c35b0 for stat - 0x24 in 0..9 picks the arm. On the rule's arm: value > 0 passes; a player (unit type 0) passes; a hireling class (FUN_0063ee90) passes; else the unit's OWN base stat (FUN_006253b0(unit, stat, 0) -> FUN_00624ed0 on the unit's +0x5c list) >= 100 gives value / 5 (the 0x66666667 multiply, truncated toward 0). V run: of stats 0..0x16f with value -50 and base 150 it changes exactly 36, 37, 39, 41, 43, 45; over 6 stats x base 99 / 100 / 101 x value -260..40 it equals `trunc(v / 5) if v < 0 and base >= 100 else v` in every case (0 mismatches).
2. **The base includes the boss mods.** FUN_005a1370 writes Magic Resistant's and the enchanteds' adds with FUN_00627260(unit, stat, total + add, 0) (V read 0x5a13a3..0x5a15e4), the setter spawn uses for the MonStats columns (sk-resist-trace Q1); FUN_00627260 writes the unit's +0x5c list through FUN_006270b0, the list FUN_006253b0 reads (V read). So `Monster::target()`'s `type.diff[d].res + boss_res` is the right base for d2d (S, ai.hpp:366).
3. **The curse callback FUN_005c35c0** (V read 0x5c35c0..0x5c3794): when the record's first aurastat (+0x1c) is >= 0, its value (+0x34) goes through FUN_005c3540 first and **a result of 0 applies no curse at all** (0x5c3639 -> return 0, no state). The state is then built (FUN_0056e970) with that first adjusted value; aurastats 2..6 (+0x20.., values +0x38..) each go through FUN_005c3540 and a non-zero result is added to the state's list (FUN_00627150). Curse replacement (FUN_005c34b0) and the immunity checks before (FUN_0056e2f0 on states 0x17 / 0x1b / 0x38) are outside the claim.
4. **The aura callback FUN_005cedc0** (V read 0x5cedc0..0x5cf00d): for each of six aurastats with a non-zero value and a stat id > 0: stat 0x6e takes its own path; else FUN_005c3540, a 0 result is skipped; the ItemStatCost `direct` stats (flag at 0x6ce278) are written to the base, the rest go to the aura state's list (FUN_00627150). Stats 36..45 have `direct` empty (D), so a resistance change of an aura is a state-list entry like a curse's, also judged against the base. No "first stat 0" rule here.
5. **The dm formula FUN_00645b20(lvl, min, max)** (V read 0x645b20..0x645b5e; the only `imul 0x6e; add 6; idiv` in the binary): min + (110 lvl / (lvl + 6)) (max - min) / 100, capped at max. V run for Lower Resist (25, 70): L1 31, L2 37, L5 47, L10 55, L20 62, L60 70, L99 70. That the calc VM's `dm` op calls this function is I (the shape); `ln34` = par3 + (lvl - 1) par4 is I (calc language, S upstream skills.md).

Skills with a resistance aurastat (D, Skills.txt patched): Lower Resist 91 (curse, srvdofunc 30: fire / light / cold / poison `-dm56`, par5 25, par6 70), Amplify Damage 66 (curse: damageresist `-par5` = -100), Decrepify 87 (curse: velocitypercent, damagepercent, damageresist, attackrate all `par5` = -50; damageresist is the THIRD stat), Conviction 123 (enemy aura, srvdofunc 66: skill_armor_percent `-dm56`, then fire / cold / light `-min(ln34,150)`, par3 30, par4 5). The rest (Resist auras, Salvation: friendly; Valkyrie, Dopplezon, golems, wolves, Fade, Berserk: own stats) are not enemy effects. **No 1.14d skill lowers magic resistance.**

## 2. Own differential (V run, the worker host)

- **Generator** `tools/emu/verify_curse_immune.py gen SEED N`: a 32-bit LCG (`x * 22695477 + 1`, seed masked to 32 bits; the author used xorshift32), seeds 0x5EED0001 and 0x0BADC0DE, 328,080 cases each. Every case names a REAL skill and level, not a stat value: the value calcs run on both sides.
  - class 0, exhaustive: Lower Resist, Amplify Damage, Decrepify (monster curse) and Conviction (player aura) alone, levels 1..60, x 21 bases (-150, -100, -50, -1, 0, 1, 20, 50, 75, 95, 99, 100, 101, 104, 105, 110, 115, 125, 150, 200, 1000) on all six types: 5,040.
  - class 1, exhaustive stacked: each curse at 1 / 5 / 10 / 20 / 30 / 60 x Conviction at 1 / 5 / 10 / 20 / 24 / 25 / 26 / 60 (the 150 cap at 25) x the 21 bases: 3,024.
  - class 2, spawn values: every distinct six-tuple of the MonStats ResDm..ResPo columns of Normal, Nightmare and Hell (D), with no add, Magic Resistant (+40 fire / light / cold), +75 on one element, or MR + Lightning Enchanted; a random curse and / or Conviction at 1..60: 100,016.
  - class 3, random one effect (base -150..1000 with weight on 90..130, levels 1..99): 100,000. Class 4, random curse + Conviction: 100,000.
  - class 5, hireling and player defenders (game side only, d2d has no such `target_of` path): 20,000.
- **game side** `verify_curse_immune.py game`: FUN_005c3540 run natively per aurastat (memoized on stat, value, base, defender kind: 119,900 distinct native calls per seed), the curse's first-stat rule of FUN_005c35c0 and the aura's skip of 0 as read (section 1), Lower Resist's and Conviction's dm56 through FUN_00645b20 natively; totals = base + the adjusted changes. Self-checks printed each run: the stat set, the trunc rule (0 mismatches), the `direct` column, the dm values.
- **d2d side** `tools/emu/test_verify_curse_immune.cpp` (as `tests/game/test_verify_curse_immune.cpp`): a Paladin `World`, zombie1 adjacent, `boss_res` pins the six bases (BASEFAIL check: 0), `Monster::curse = {skill, level}`, `actor->aura` = Conviction with `skill_level` = the case's level (AURAOUT check: 0), then the REAL `Fight::target_of(0).res`. `--vlist` prints d2d's `eval_calc` values of every Skills row with a resistance aurastat or srvdofunc 30, levels 1..99.

| build | class 0 (5,040) | class 1 (3,024) | class 2 (100,016) | class 3 (100,000) | class 4 (100,000) | immunity verdict (>= 100) differs, all classes |
|---|---|---|---|---|---|---|
| base 72132b99, seed 0x5EED0001 | 2,400 | 1,440 | 50,950 | 71,088 | 93,863 | 82,257 |
| base, seed 0x0BADC0DE | 2,400 | 1,440 | 50,872 | 76,793 | 94,017 | 85,357 |
| claim 6a04cb3e alone, both seeds | **0** | **0** | **0** | **0** | **0** | **0** |
| head 557d14e3, both seeds | **0** | **0** | **0** | **0** | **0** | **0** |

(rows where any of the six totals differs). On base every differing type is "base >= 100 and lowered" (398,637 / 404,474 type-cells), nothing else; e.g. Lower Resist 1 on base 100: game.exe 94, d2d 69. Further results, same runs:

- Class 5 (V run, game side): hirelings and players keep the full value in all 20,000 cases. d2d: the hireling's curse goes through `merc.cpp:70-74` at the full value (S), equal; the claim does not touch it.
- Value calcs: d2d's `eval_calc` equals the game side for Lower Resist, Amplify Damage, Decrepify and Conviction at levels 1..99 (396 rows, 0 differ). Limit: dm via game.exe's own FUN_00645b20 (V run), `ln34` and the `par` / `min` parts by the calc text (I); game.exe's calc VM itself was not run.
- D1b reach: no srvdofunc-30 row's first aurastat comes out 0 after the rule at any level 1..99 (0 rows; d2d's values, equal to game.exe's for the four skills). Lower Resist's smallest is L1 -31 -> -6, Amplify -100 -> -20, Decrepify's first stat is velocitypercent (not a resistance).
- md5 (first 12): c1 30a4b8af72d6, c2 ee13ae9bdb80, g1 036267a0ed41, g2 0354e67dfaa0, d-base 1 / 2 a44ec4b786be / 08beaa1ae9f2, d-claim = d-head 1 / 2 d18f49749aaa / db8b9b56c09e.

Not covered by my run: the hit itself (`effres`, `dealt`; the claim does not touch `missile_blow`, and the trace's run matched it outside D1 / D2, S), monster cries (no 1.14d cry has a resistance aurastat, D), the merc's own enemy aura (d2d applies only `actor->aura`; pre-existing, not the claim), the curse's duration and replacement.

## 3. Isolation

Three Release builds of my harness (D2D_WERROR=OFF): base, claim alone, head. claim = head on both seeds (md5 above): the tooling commit changes nothing in d2d. `--vlist` equal on base and head (the calcs are untouched).

## 4. Side claims

- **D2 not touched** (V read of the diff): 6a04cb3e changes only `Fight::target_of`, `apply_target_stats` (fight.hpp signature, missiles.cpp) and one skills.md line; `rules::missile_blow` / `resisted` are untouched.
- **D2 unreachable with 1.14d data** (D + I): no skill lowers magic resistance (section 1), so a monster's magic total is its column plus boss adds, never below 0. Physical: the lowest ResDm is 0 (blank cells; the author's "smallest ResDm 10" skips blanks) and a monster carries one curse: Amplify Damage -100 gives exactly -100, which the floor leaves at -100; Decrepify -50. So no case under -100. The conclusion holds; the draft's "-90" should read "-100, equal to the floor".
- **No other caller of `apply_target_stats`** (V read, `git grep`): one declaration and three uses in missiles.cpp (definition and the two calls in `target_of`) on the candidate, on upstream main 7c3a50af, #155..#158 and all local candidates and their rebased copies.
- **Siblings outside the claim** (I + D): pets as defenders get monster curses in d2d at the full value (`pets.cpp:815`, S) while game.exe would judge them by the rule (a pet is type 1, not a hireling); unreachable in 1.14d: the only pet bases >= 100 are Fire Golem fire and Iron Golem poison, and the only monster casters of these curses (succubus witches: Amplify Damage; doom knights, Baal on his throne: Decrepify) lower damage resistance only.

## 5. Merges (textual, `git merge-tree --write-tree`, both orders)

Candidate head 557d14e3 against #155 (8c56e96a; shares fight.hpp), #156 (3f418788), #157 (cf85634a), #158 (6b3964fa; shares missiles.cpp and re/skills.md), fix/armor-hit (d5bd19c1; shares re/skills.md), fix/freeze-resume (efacc7f4), fix/frozen-armor-and-resume-2c (f9e38b15; shares re/skills.md) and the new upstream main 7c3a50af: **clean in both orders, all eight**. The rebased copy `fix/curse-immune-rebased` (4132071c, 16d6663a on 7c3a50af, card upstreamsync) is the same patch (`git range-diff`: only the cherry-pick line added) and merges clean with #155, #157, #158 and the rebased armor / resume / bundle copies; against #156 it reports #156's own README conflict with 7c3a50af (the same as 7c3a50af vs #156), not one of this claim's files. I did not build the rebased copy (no file of the claim changed upstream).

## 6. Review notes

1. LOW (draft text): "smallest first-stat value on an immune is Lower Resist's -25 -> -5": it is -31 -> -6 at level 1 (FUN_00645b20 run), and Decrepify's first stat is velocitypercent; the conclusion (unreachable) holds.
2. LOW (draft text): D2 "Amplify Damage -100 on the smallest ResDm 10 gives -90": the smallest is 0 (blank), giving -100, which the floor does not change; still unreachable.
3. LOW (draft text, Limits): add the pet sibling (section 4) as not modelled and unreachable in 1.14d data.
4. INFO: `lowered` uses `value < 0`, game.exe `value <= 0`: the same result for 0. C++ `/` truncation equals the game's (V run, 0 mismatches over -260..40).

## 7. Verdict

**READY** (text edits only, no code change). The claim holds: 0 of 616,160 compared cases (656,160 minus the 40,000 game-only class 5) differ on claim alone and on head (both seeds, every class, the real curse and aura skills and levels, the calc path included), against 82,257 / 85,357 immunity verdicts and 219,741 / 225,522 rows on base. Side claims hold (D2 untouched and unreachable; no other caller); merges clean in both orders with #155..#158, the three local candidates and upstream main 7c3a50af. Draft edits: review notes 1 to 3; proof table: add this run.

## Commands (worker host)

```sh
# trees: base 72132b99, claim 6a04cb3e, head 557d14e3; each gets tests/game/test_verify_curse_immune.cpp
cd <upstream>/tools/emu       # D2_MPQ_DIR / D2_PATCH_INSTALLER as for the other oracles
python verify_curse_immune.py gen 0x5EED0001 100000 > c1.tsv
python verify_curse_immune.py game c1.tsv > g1.tsv            # writes c1.tsv.gvals; self-checks on stderr
build/tests/test_verify_curse_immune --vcurse c1.tsv | grep -E '^[0-9]+ ' > d1.tsv
build/tests/test_verify_curse_immune --vlist | grep '^L ' > vlist.txt
python verify_curse_immune.py compare c1.tsv g1.tsv d1.tsv vlist.txt
```

## Shared-file lines for the merger

- docs/research/pr-bundle-queue.md (owned by another card), append: `- CURSEIMMUNE verifier (VERIFYCURSEIMMUNE 2026-11-01, docs/research/pr-curseimmune-verifier.md): fix/curse-immune **READY after draft text edits** (no code change): own oracle + harness with the real curse / aura skills and levels, 616,160 compared cases on 2 seeds, base 219,741 + 225,522 rows differ -> claim alone = head 0; hireling / player defenders full value (game.exe, 20,000 x 2); calcs equal levels 1..99; D2 untouched and unreachable; merges clean both orders with #155..#158, armor-hit, freeze-resume, bundle-2c and main 7c3a50af`
- docs/BACKLOG.md: `- SKRESISTFIX: verifier READY (docs/research/pr-curseimmune-verifier.md): draft text edits (first-stat -31 -> -6, D2 -100 not -90, pet sibling); next the det state hashes / private-fork carry card and the user's go`
- docs/INDEX.md, research: `- docs/research/pr-curseimmune-verifier.md: independent verifier of fix/curse-immune (own oracle and harness, real skill calcs, merges)`
- CHANGELOG.md: none. GLOSSARY.md: none missing.
