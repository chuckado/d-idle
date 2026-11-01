# PR draft: a curse takes a fifth of its lowering off an immune monster (SKRESISTFIX)

> status: READY (independent verifier READY, three text edits applied), held for the user's go. Local candidate only (`fix/curse-immune`, and its rebased copy `fix/curse-immune-rebased` on the worker host); nothing pushed or posted. All six earlier PRs (#152 #153 #155 #156 #157 #158) are merged upstream; the candidate was rechecked clean against upstream main 52362d52 (UPSTREAMSYNC; merge-tree only, not rebuilt on that head). Evidence and commands: [pr-curseimmune-proof.md](pr-curseimmune-proof.md); verifier: [pr-curseimmune-verifier.md](pr-curseimmune-verifier.md); trace: `docs/research/sk-resist-trace.md` (card skresist, D1).

Title: `Curses and auras: a lowering counts a fifth on a monster whose base resistance is 100 or more`

---

## Limits

- Only the lowering of a monster's resistance by its curse, cry or the player's enemy aura (`Fight::target_of` -> `apply_target_stats`: Lower Resist, Conviction, Amplify Damage, Decrepify, the other aurastats on stats 36 / 37 / 39 / 41 / 43 / 45). Everything else in the resistance path is unchanged and was already equal in the differential: base columns, pierce only under 100, the -100 floor of fire..poison in `missile_blow`, the 100 line.
- The curse VALUE calcs (dm56, ln34) are not part of the claim: the test mode feeds one aurastat value per case. d2d's `eval_calc` is shared with earlier proven calcs (I, not re-run here).
- Hirelings and players as defenders are not modelled by d2d's `target_of` (monsters only); game.exe gives them the full value (FUN_005c3540 excludes them), so nothing to change there.
- Not claimed (separate code path, `rules::missile_blow` / `player_blow`): D2, the missing -100 floor on magic and physical. With 1.14d data it is unreachable on the missile path (no curse lowers magic resistance, smallest ResMa 10, Amplify Damage -100 on the smallest ResDm, which is 0 (blank cells), gives -100, equal to the floor). The differential still shows it (magic total under -100, 1,958 cases) because the grid goes past the data. Also not claimed: melee elements without pierce or floor (D3), Sanctuary's undead rule (D4).
- A curse whose FIRST aurastat rounds to 0 on an immune applies no curse at all in game.exe (0x5c3639); unreachable with 1.14d skills (smallest first-stat value on an immune is Lower Resist's level 1 -31 -> -6; Decrepify's first stat is velocitypercent), not modelled.
- Pets as defenders: d2d gives a pet a monster's curse at the full value (`pets.cpp:815`) while game.exe would judge it by the rule (a pet is not a hireling); not modelled, unreachable in 1.14d data (the only pet bases of 100 or more are Fire Golem fire and Iron Golem poison, and the monster casters of these curses lower damage resistance only).
- Not run: det state hashes, an MSVC build, a live game.

## What changes

game.exe's curse and aura callbacks (srvdofunc 30, 65, 66) store each aurastat through FUN_005c3540. For a monster that is not a hireling, when the value is negative and the monster's BASE value of that resistance stat (before any curse) is 100 or more, the stored value is value / 5, truncated toward 0. The monster's total is base plus the stored values, and a hit is immune when the total is 100 or more. So Lower Resist level 1 (-31) leaves a lightning-110 Hell monster at 104: still immune.

d2d added the full value (`target.res[i] += value`), so any curse or Conviction broke an immunity game.exe keeps: the same Lower Resist 1 left that monster at 79, taking 21 % of a lightning hit instead of 0.

`target_of` now remembers the monster's resistances before any curse and `apply_target_stats` counts a negative change as a fifth against that base when it is 100 or more. Positive changes and bases under 100 take the full value as before.

## Proof

`tests/game/test_sk_resist.cpp` (`--resist CASES`) puts the case's one aurastat on the monster's curse and reads the real `Fight::target_of`, then the real `rules::missile_blow` for the damage; `tools/emu/sk_resist.py game` runs FUN_005c3540, FUN_0057be00 and FUN_0057c6c0 natively under the emulator on the same cases (6 resistance types x base values around the 100 line x curse values -150..+75 x pierce x rolls: 47,628 grid cases plus 100,000 random per seed).

| build | adj differs (of 147,628) | dealt differs | seed 7 adj / dealt |
|---|---|---|---|
| upstream main + the test | 88,101 | 14,355 | 88,224 / 14,289 |
| this branch | 0 | 1,958 (D2, magic under -100, not claimed) | 0 / 1,977 (same) |

Every difference before is "base 100 or more and a negative curse". After, `adj` matches on every case, and `dealt` / `effres` differ only where the total is under -100 on magic or physical (D2, outside the claim). A line-for-line Python port of the fixed rule equals the real build on all cases of both seeds.

pr_guard 0 FAIL / 0 WARN on the range and on each commit; presubmit --quick 0 FAIL; test_game, test_rules, test_combat, test_protocol, test_monsters pass; gcc -Wshadow -Wunused -Wsign-compare and clang-tidy (names and includes) clean on the touched lines; ruff clean.

## How to review

1. Read the claim: `components/game/fight/missiles.cpp` (`target_of` takes the base, `apply_target_stats` has one lambda), 11 lines; the `fight.hpp` signature.
2. Disassemble FUN_005c3540 (0x5c3540..0x5c35a3) with `tools/emu/disasm.py`: the base stat comes from FUN_006253b0, the `cmp 100` and the divide by 5 sit in the server range.
3. Run the differential (commands in the proof): `python sk_resist.py gen 100000 1 > c.tsv`, `game c.tsv`, `test_sk_resist --resist c.tsv`, `cmp`.
4. Siblings: the change is inside `target_of`, so every path that reads a monster's resistances (missile, melee, Static Field, pets) gets it; no caller of `apply_target_stats` exists elsewhere.

## Files

`components/game/fight.hpp`, `components/game/fight/missiles.cpp`, `docs/research/re/skills.md` (one line); tooling `tests/game/test_sk_resist.cpp`, `tools/emu/sk_resist.py`.
