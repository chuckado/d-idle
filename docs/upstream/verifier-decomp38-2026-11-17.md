# Verifier: decomp38, a Mana Burn monster's hit (2026-11-17)

> status: current, 2026-11-17 (card verifydecomp38, TASK-608, Opus, singly; independent verifier before posting). Nothing posted. Subject: private fork branch `fid/mana-burn-amount-r1` (head `d7a97516`: oracle `a6093f9d`, claim `d7a97516`, on upstream `ed007569`), body `docs/upstream/bodies/mana-burn-amount-body.md`, trace `docs/research/dc38-lane-d.md` (TASK-606). Checklist from `verifier-decomp35-2026-11-16.md`, `verifier-decomp36-2026-11-16.md`, `verifier-p13-2026-11-16.md`, `verifier-p18-2026-11-15.md`. The fixed `fid/mana-burn-amount-r2` is on the private fork remote only.

Tags: **V run** (run by this verifier today, NAS builder), **V read** (disassembly or source read by this verifier), **I** (inferred), **S** (an earlier doc, named), **U** (unverified).

## Verdict: r1 FAIL, r2 PASS WITH CHANGES

Post **`fid/mana-burn-amount-r2`**: oracle `9136b9a0`, claim `0243f24f`, on upstream main `ed007569`; body after M1 to M8 below.

The rule is game.exe's and the claim implements it (re-derived below). The 256 times is not a unit artefact: the stat is stored shifted by 8, rolled shifted by 8 again, and taken off stat 8, which the same getter returns in 256ths. r1 cannot be posted:

1. **No test pins the rule (blocker).** r1's `test_mana_burn` has oracle modes only; run alone (as ctest runs it) it prints `SKIP` and exits 77, so no mutant fails ctest. r2's claim adds a no-argument mode ctest runs: game.exe's own amounts for eight cases from the rule (no data needed), and with the data the player's mana after a burn through the real `Fight::hurt_side` and a Normal level 9 unique's hit end to end. All eight mutants below abort it.
2. **The oracle measured the field, not the player's mana (high).** r1's d2d side printed `Foe::mana_burn` (shifted by 8 in the oracle commit, raw in the claim: the claim commit changed the oracle's own measurement). A mutant that keeps fight/tick.cpp's old `<< 8` on top of the new 256ths field takes 65,536 times the points off the player and passes r1's oracle (V run, `m6r1test` below). r2's d2d side runs the burn through `Fight::hurt_side` and prints the mana the player lost; the measuring code is the same in both commits.
3. **presubmit fails (high).** r1: `ruff F401` (`import random` unused in `mana_burn.py`) and `short-names` (`lo`, `hi` at `uniques.hpp:447`): **2 FAIL** (V run, `presubmit.py --quick --base upstream/main`). r2: 0 FAIL, 3 WARN (answered by M5).
4. **Usage errors and the README row (medium).** r1 `mana_burn.py` with no argument is an `IndexError`, `game` without a file likewise, a missing file a traceback, `cmp a b junk` silently runs the range compare; the README row sits in the "When it breaks" troubleshooting table. r2: `usage:` and exit 1 on no, unknown or short arguments and a bad flag, `cannot read` on a missing file, a malformed case line refused; the verb is `compare` as in the neighbouring rows; the row moved after the MonUMod row. Hand-made files (V run): empty, short, malformed, out of step, the parent's amounts: exit 1 each; equal files exit 0.
5. **No `docs/research/re/` finding (medium; the template's checklist item).** r2's claim rewrites `monsters.md` row 25 (spawn `0x5a20bf` / `0x5a20d7`, record `0x57ba96..0x57baac`, roll `0x57a8ef` / `0x57a91d`, apply `FUN_0057aa60`, the Normal minion rows).

Identity and history are already right in r1 (V run): both commits by the user's noreply identity as author and committer, oracle first, `Agent: Claude Code` the only parsed trailer, after a blank line. r2 keeps all of that (V run, `git log --format`, `git interpret-trailers --parse`).

## Game side, re-derived (V read by this verifier, then V run)

- **Spawn, `FUN_005a1f90`** (`0x5a1f90..0x5a20f7`): the MonLvl column index is `difficulty + 3 x (6 + expansion)` (`0x5a205b..0x5a206a`: the L-DM column in an expansion game); unique flag set: MonUMod constants rows `0x1c` / `0x1f` + difficulty, else `0x10` / `0x13` (`0x5a206d..0x5a209c`); `x = DM x constant / 100` (`imul 0x51eb851f; sar 5`), then **`shl esi, 8`** (`0x5a20bf`) and **`shl edi, 8`** (`0x5a20d7`) before the stat adder `FUN_006272b0` for stats `0x3e` / `0x3f`. The real spawn (`FUN_005a2120` through `umod_stats.py`'s rig, V run, six MonStats rows x three difficulties): a Normal unique at mlvl 14 gets stat 62 / 63 = 1,536 / 2,560 (x = 6..10), at mlvl 5 512 / 1,024; upstream's `test_umod_stats` already compares stat 62 with d2d's `mana_lo << 8` (S), so d2d's `mana_lo` is game.exe's `x`.
- **Normal (bot finding B1's question).** MonUMod row 25 (manahit) is enabled with `upick` 6 in **all three** difficulties and `cpick` 0 (no champion) (V run, the loaded table at `[0x96c880]`). The unique constants are 66 / 100 in every difficulty; the minion constants (rows 16 / 19) are **0 / 0 in Normal**, 33 / 50 in Nightmare and Hell, so in Normal the unique itself burns and its minions do not (V run: Normal minion stats 0 / 0). d2d's `make_boss` reads the same rows (`ai.cpp:2860..2864`), so this holds there too. A Normal-difficulty Mana Burn unique is therefore real and is the common case of the bug.
- **Record, `FUN_0057b7d0`** (`0x57ba4c..0x57bac1`): after the life-drain roll (`0x3c` / `0x3d`, record flag 4), for an attacker of type 1 that `FUN_0063ee90` calls no hireling and record flag 8 clear, `rec +0x3c += FUN_0057a8e0(0x3e, 0x3f) x SrcDam / 128` (`0x57bab1..0x57bac1`).
- **Roll, `FUN_0057a8e0` / `FUN_0057a880`**: both stats through the getter `FUN_00625480`, each **`shl 8`** (`0x57a8ef`, `0x57a91d`); max under 8 returns unchanged; else `min + rand(max - min)` (`FUN_0045c3e0` on the unit's seed `+0x20`), percent bonuses 0 here. On the stored values that is `(x_lo << 16) + rand((x_hi - x_lo) << 16)`.
- **Apply, `FUN_0057c6c0` -> `FUN_0057aa60`** (`0x57c879..0x57c87e`; `FUN_0057aa60` has this one caller, V run, a scan of `.text` for `call` rel32): stat 8 minus the field, under `0x100` set to 0 (`0x57aa76..0x57aa7e`). The same getter's stat 6 result is compared with `0x100` at `0x57c865`: the getter returns the raw 256ths and applies no ValShift, so the unit argument holds: `x_lo << 16` in 256ths is `x_lo x 256` mana points.
- **Energy Shield** (new, V read): `FUN_005ca6c0` walks the table `0x6e30e0..0x6e3120`, eight `{field, flag}` pairs: physical, fire, lightning, cold, magic (flag 0), then **`+0x38`, `+0x3c`, `+0x40` (flag 1)**. A flag-1 field is taken from only when the attacker is absent or a plain monster (`0x5ca78d..0x5ca7ca`: type 1, not a hireling). So against a Mana Burn monster, Energy Shield takes its calc1 % of the burn and pays `MulDiv(part, calc2, 16)` mana for it (`0x5ca7da..0x5ca823`). d2d's `rules::energy_shield_blows` works on the five damage fields only (`states.cpp:179..199`, `combat.hpp:636`). Not part of this PR; a Limit (M3) and a follow-up candidate.
- **Other callers.** `FUN_0057b7d0` has seven callers (V run, scan): `0x56f178` (`FUN_0056f070`), the melee gate `0x57dc37`, and the skill functions `0x5c7ecd`, `0x5c7fdb` (Hunger), `0x5d5854`, `0x5d62f5`, `0x5d7ca1`. d2d burns only in `monster_update`'s melee swing (`ai.cpp:2549`, the one writer of `Foe::mana_burn`) and only the player's mana (`Fight::hurt_side` reads `side[0]` only; a merc's mana is never burned). Whether a monster's missile or skill hit burns mana in game.exe was not run. The PR claims the melee hit only (M6).
- **Heal (D2 of the trace).** Not re-run; `FUN_0057c420` is called once, from `0x57c826`, before the apply. Kept as the trace's Limit.

## Differentials (V run)

Game side `mana_burn.py game` (r2's script) on c1 (`gen`: 8 ranges x seeds 1..250) and c2 (reseed: 10 ranges incl. `(0, 5)`, `(1, 2)`, `(111, 111)`, `(60, 111)` x seeds 5001..5200); both equal the model on 2,000 of 2,000. d2d side `test_mana_burn --mana-burn` (through `Fight::hurt_side`).

| Tree | c1 range | c2 range | roll exact |
|---|---|---|---|
| oracle commit `9136b9a0` (parent d2d) | **1,750**, exit 1 | **1,600**, exit 1 | n/a |
| claim `0243f24f` | 0, exit 0 | 0, exit 0 | c1 0 of 2,000, exit 0 |

Upstream main `ed007569` ctest (D2_MPQ_DIR and D2_PATCH_INSTALLER set): 355 of 355 pass. The claim's `test_mana_burn` with no argument: rc 0.

## Mutants (V run)

PLACEHOLDER-MUT

## Build, ctest, tidy

PLACEHOLDER-BUILD

## RNG luck

PLACEHOLDER-RNG

## Required changes to the body (`mana-burn-amount-body.md`)

- **M1 (high).** Commits and tree: "upstream `ed007569` plus `9136b9a0` (oracle) and `0243f24f` (claim)". "Two commits" paragraph: the oracle's d2d side is `monster_update` then the real `Fight::hurt_side` on a player with 60,000 mana (the mana lost, not the field); the claim also adds the `monsters.md` finding and the no-argument asserts in `test_mana_burn` (game.exe's amounts for eight cases; with the data the player's mana after a burn and a Normal level 9 unique's hit), which ctest runs.
- **M2 (high).** Re-check block: `mana_burn.py compare` instead of `cmp` (`2000 cases: game.exe off its own rule 0, d2d outside game.exe's range 0`, `ok`; on the oracle commit 1,750 of 2,000, exit 1); `compare game.tsv roll.tsv exact` for the rule alone; the d2d side needs `D2_MPQ_DIR` and `D2_PATCH_INSTALLER` (the test skips without 1.14d's MonStats); add `ctest --test-dir build -R test_mana_burn`. Say compare exits 1 on an empty or unequal file, a malformed line or cases out of step.
- **M3 (high, claims limited to proven).** Limits: "With Energy Shield, game.exe also takes the shield's calc1 % of the burn and pays `MulDiv(part, calc2, 16)` mana for it (`FUN_005ca6c0`: the table `0x6e30e0` lists the record's `+0x3c` with flag 1, taken only for a plain monster attacker, `0x5ca78d..0x5ca7ca`); d2d's shield works on the five damage fields only. Not changed here."
- **M4 (medium).** "In play": add "MonUMod row 25 can be picked for a unique in every difficulty (upick 6 in Normal, Nightmare and Hell); its minions' constants are 0 in Normal, so in Normal only the unique burns." Keep the level 9 example (MonLvl L-DM 7 x 66 / 100 and x 100 / 100: 4..7 points).
- **M5 (medium).** Answer presubmit's WARNs: `siblings` "game.exe's `FUN_0057aa60` takes any defender's stat 8; d2d applies a burn to the player only (`Fight::hurt_side` reads `side[0]`), so a merc's mana is not burned, before or after; pets have no mana; a co-op partner is a player and takes the new amount"; `build` and `gcc` "built and ctest run (Release, GCC): the same failures as upstream main".
- **M6 (medium, claims limited to proven).** Limits: "`FUN_0057b7d0` has seven callers (the melee gate, `FUN_0056f070` and five skill functions); d2d burns only on a monster's melee swing. Whether a monster's missile or skill hit burns mana in game.exe is not run; nothing there is changed." Keep r1's "own melee caller not run" line.
- **M7 (low, template).** Claims table with the counts of this file (parent 1,750 of 2,000 through `hurt_side`, claim 0; exact 2,000 of 2,000 twice; mutants abort `test_mana_burn`); tick the template's three checklist items (test `tests/game/test_mana_burn.cpp`; finding `docs/research/re/monsters.md`; `pr_guard` 0 FAIL 0 WARN). Drop "The whole test suite was not run on this branch".
- **M8 (low).** The reseed row: seeds 5001..5200 x 10 ranges (including the power-of-two bounds `(1, 2)` and `(0, 5)`... and `(111, 111)`), PLACEHOLDER-RESEED.

## For the coordinator (not body changes)

- **B1 holds and is worse than the trace says** (V read `internal/resource/hell.go:82`: the mana-potion rule needs `in.Diff != d2.Normal`). A Normal Mana Burn unique empties any Normal pool in one landed hit, so the potion rule is useless there too (the next hit empties the refill); the defence is not being hit (S, opinion: kill it from range or skip it). Energy Shield does not protect the pool (it takes from the burn and pays at least as much mana when calc2 is 16 or more; I).
- `dc38-lane-d.md` section 7: "the oracle measures what fight/tick.cpp subtracts" was the field, not the subtract; replace the branch line by r2's.

## Hygiene (V run, r2)

- Identity: both r2 commits have the user's noreply identity as author and committer, `Agent: Claude Code` the only parsed trailer.
- `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN. `presubmit.py --quick --base upstream/main`: claim 0 FAIL, 3 WARN (`siblings`, `build`, `gcc`); oracle commit 0 FAIL, 2 WARN. r1: 2 FAIL (ruff F401, short-names).
- Diff: oracle 3 files (`mana_burn.py`, `test_mana_burn.cpp`, a README row), no `components/` change, its test has no failing assert (bisect-safe); claim 7 files (`uniques.hpp`, `ai.cpp`, `ai.hpp`, `fight/tick.cpp`, `monsters.md`, the test, two docstring lines). No private-only code, no host names or paths (read of the diff).

## Not checked

- No Windows, macOS or Clang build; no live game, no lab drill (the trace's D1 recipe not run); nothing ran on the worker host.
- The heal (D2), the missile and skill paths, Energy Shield's take on d2d's side, the under-one-point rule (D3) and SrcDam other than 0x80.

## Re-check commands (as run)

Builder container, the emulator image of `docs/emu-image.md`; lab scripts `verifyd38/{bld,chain,stage,emu,emuruns,rules,dx}.sh`, `py/{dis,xref,umod,spawn,tab}.py`, `cpp/rule_harness.cpp` (lab only).

```sh
python3 mana_burn.py gen > c1.tsv; python3 mana_burn.py game c1.tsv > g1.tsv
test_mana_burn --mana-burn c1.tsv | grep -E '^[0-9]+ ' > d1.tsv
python3 mana_burn.py compare g1.tsv d1.tsv; echo $?             # oracle commit: exit 1; claim: ok, 0
test_mana_burn --mana-burn-roll c1.tsv | grep -E '^[0-9]+ ' > r1.tsv
python3 mana_burn.py compare g1.tsv r1.tsv exact; echo $?       # claim: 0 of 2,000, 0
test_mana_burn; echo $?                                          # claim: ok, 0 (D2_MPQ_DIR, D2_PATCH_INSTALLER)
```
