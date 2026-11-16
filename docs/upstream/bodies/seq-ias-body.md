## What changes

A player's attack sequence (Jab, Impale, Double Swing, Double Throw, Frenzy, and the Assassin's Fists of Fire, Claws of Thunder, Blades of Ice and Dragon Claw: Skills.txt `anim` SQ) runs 30 points slower than a plain swing. FUN_00623f50's attack branch takes 30 off the percent `effective IAS + stat 68` for a unit of type 0 in mode 0x12 (`0x624290..0x62429b`: `cmp [unit], 0; cmp [unit+0x10], 0x12; sub esi, 0x1e`), before the clamp to 15..175. Which skills reach that branch is the record's: FUN_006216e0 sends `seqtrans` SC (Lightning, Chain Lightning) to the cast branch, FUN_00621580 sends a UseAttackRate skill to the attack branch (record `+6` against the mask at `0x6ce274`), and a record with neither (Charge, Leap Attack, Whirlwind, Inferno, Arctic Blast) takes the rate from another stat and is not touched here.

`rules::sequence_sias(use_attack_rate, seqtrans)` is the term (-30 or 0); `Fight::start_sequence` adds it to the skill speed it already gives `attack_ticks`. A Jab with bare hands at 0 IAS lasts 19 ticks, not 13; Frenzy 25, not 17.

Two commits: the oracle (`tools/emu/seq_rate.py`, `tests/game/test_seq_rate.cpp --seq-rate`, README row; it prints what `start_sequence` does on the parent), then the claim (the rule, the call site, 36 frozen cells and the rule asserts in the test, a paragraph in `docs/research/re/combat.md`).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00623f50` `0x6241d7..0x62430a` | the attack branch: `ebx = FUN_00621740(unit, 0)` (stat 93), `esi = ebx + stat 68`, at `0x624290` a player in mode 0x12 loses 30, clamp 15..175, rate = record `+0xc` x esi / 100, stored at `+0x4c` and `+0x3c` |
| `FUN_006216e0` `0x6216f5..0x62172c` | mode 0x12 with an active skill whose record byte `+0x12` (seqtrans) is 0xa (SC) takes the cast branch |
| `FUN_00621580` `0x6215f4..0x621622` | a mode whose table row has `+8` set (S1..S4, 0x12) takes the attack branch when the active skill's record `+6` has a bit of `0x6ce274` (value 8, V run) |
| `0x62442f`, `0x62443f` | mode 0x12 runs the sequence init FUN_00621260 first, then FUN_00623f50: the rate it stores is the one the sequence steps by (`+0x3c`) |
| `docs/research/re/skills.md` "Cast rate" (this repo) | the sequence ends when its countdown is <= 0: `ceil(256 steps / speed)` |

## Counted differentials (V run, tree = upstream `d1a10a38` plus the branch)

Oracle: `seq_rate.py compare` runs FUN_00623f50 natively on the real Skills.txt records game.exe's loader read (`skill_tips.boot`), mode 0x12, for each cell `test_seq_rate --seq-rate` prints: a class x its own sequence skills x main-hand weapons (none, sword, dagger, great sword, spear, staff, bow, crossbow, claws) x IAS {0 .. 200} x WSM {-30 .. 30} x skill speed {0, -50, 20, 40}, 17,600 cells; the game ticks are `ceil(256 steps / rate)` with the rate game.exe prints for stat 93 = IAS and stat 68 = 100 - WSM + skill speed. The branch each record takes is told from the rate game.exe prints (`seq_rate.py table`): 7,600 cells are skills it times in the attack branch and are compared; Lightning, Chain Lightning (cast) and Charge, Leap Attack, Whirlwind, Inferno, Arctic Blast (another rate) are listed by name and left out (their d2d lengths are the same on the parent and the claim: 0 lines differ).

| 7,600 cells | differ before (parent) | differ after |
|---|---|---|
| Jab, Impale (Amazon); Double Swing, Double Throw, Frenzy (Barbarian); Fists of Fire, Claws of Thunder, Blades of Ice, Dragon Claw (Assassin) | **6,941** | **0** |

`compare` exits 1 on a difference, on a malformed line and on a file with no compared cell (all three run). Mutants (V run): `kSequenceRatePenalty = 0` aborts `test_seq_rate` at the rule assert; dropping the call-site term aborts it at the first frozen cell.

Build and tests (V run, GCC 14, Release, on the builder): `presubmit.py --quick --base upstream/main` 0 FAIL (3 WARN, answered below), `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN, clang-tidy-19 on the 265 translation units that include a changed file: 0 findings in the changed files (3 `misc-include-cleaner` errors in `debug_hero.cpp` and `test_akara_decipher.cpp`, files this branch does not touch). Full ctest: parent 5 of 326 fail (`test_outdoor` aborted, `smoke_d2d`, `touch_controls_d2d`, `touch_corpse_d2d`, `pad_use_d2d`), the claim the same 5 of 327 (the new test passes). Linux GCC only; no Windows, macOS or Clang build run (U).

## Limits

- Only the rate term is claimed. The step count (`seq.size()`), the AnimData speed d2d counts a sequence at (the class's A1 record) and the length `ceil(256 steps / rate)` are unchanged and not re-proved here (U). For the AnimData record, FUN_00620f00 gives the `GH` record for mode 0x12 with no weapon (speed 256 for every class, V run); with a weapon in hand it was not run. If it is the weapon's `GH` record, an Assassin with claws counts at 256 where d2d uses 208.
- stat 68 = 100 - WSM + skill speed is the convention of the merged rate checks (`player_attack_rate.py`); how a weapon writes -WSM into stat 68 is I.
- That a player's mode is 0x12 while a Skills.txt `anim` SQ skill plays follows `swing_mode`'s existing reading of `anim` (FUN_004cc5b0); not re-run here.
- No kill-rate or damage-per-second effect is claimed.

## Answers to the presubmit WARNs

- Hit paths (melee, skill, missile): the rule is the player's sequence timing only; Whirlwind and Charge (also SQ) take another rate branch in game.exe and keep their length here; a missile has no sequence.
- Build and gcc skipped under `--quick`: run on the builder as above.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_seq_rate
cd tools/emu && uv sync
set -o pipefail
../../build/tests/test_seq_rate --seq-rate | grep '^D ' > /tmp/seq.tsv
uv run python seq_rate.py compare /tmp/seq.tsv     # last line: 7600 cells  same (parent: 6941 differ)
uv run python seq_rate.py table
../../build/tests/test_seq_rate
```

(The test needs `D2_MPQ_DIR` and the 1.14d data set; the compare needs the game image the other emulator checks use, `tools/emu/README.md`.)

---

Agent: Claude Code
