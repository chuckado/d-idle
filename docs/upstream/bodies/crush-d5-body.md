## What changes

A Crushing Blow on a monster divides its life by 4, or by 8 for a MonStats boss row or a superunique, plus the player-count share (`FUN_005bffc0` `0x5c0031..0x5c00a8`). d2d gave the 8 to a plain unique as well (`Monster::crush_divisor`: `Boss::unique || Boss::superunique`). The 8 comes from `FUN_005a0180(unit, 2)`, a test of bit 2 of the monster data word at `+0x16`; that bit is the superunique bit, and only the two superunique branches of the creators write it, so a unique (a rare, random-modifier monster) takes the quarter as a champion or a plain monster does.

`crush_divisor` (`components/game/ai.hpp`) now gives 8 only to a MonStats boss row or `Boss::superunique`. `test_game_tables` asserted a unique at 8; it asserts 4 now and `test_crush_divisor` pins game.exe's crushes for twelve rows, kinds and player counts. The RE notes (`monsters.md`, `drops.md`) say who writes the bit (the "possible disagreement, not established" paragraph of `drops.md` is settled).

Two commits: the oracle (`tools/emu/crush_div.py`, `tests/game/test_crush_divisor.cpp --crush`, a README row), then the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_005bffc0` `0x5c0031..0x5c00a8` | divisor 4; a hireling 10 (`FUN_0063ee90`); else `FUN_0063e9f0(0, unit)` (a MonStats boss row) or `FUN_005a0180(unit, 2)` (mask 2) gives 8; then plus `MulDiv(divisor, FUN_005738f0(max(stat 0x64, 1)), 100)` |
| `FUN_005a0180` `0x5a0193` | `(data word +0x16 & mask) != 0` for a type-1 unit |
| `0x5a4599`, `0x5a4a99` | `or word ptr [eax + 0x16], 2`: the only two writes of bit 2 in the executable (a scan of every `or`, `and`, `mov` of `[reg + 0x16]` in `.text`); the second sits beside the SuperUniques spawned-bit table at `+0x1d30` |
| `FUN_005a4940`, `FUN_005a0320` `0x5a4967`, `0x5a0375` | a unique: `or ..., cx` with cx 1, then bit 8: 9 |
| `FUN_005a48c0` `0x5a4905`, `0x5a0375` | a champion: `or ..., 5` then bit 8: 0xd |
| `FUN_005a01b0` | the setter with a mask: no caller in `.text` |

## Counted differentials (V run)

`crush_div.py game` boots game.exe with every table loaded and runs `FUN_005bffc0` natively (the attacker's Crushing Blow 100 percent, a melee strike) on a type-1 unit of every MonStats row (734), made plain, champion, unique, superunique or as a minion (the data word the creators leave, V read above), at 1 to 8 players, on a life of 262,144 points; `FUN_0063ee90`, `FUN_0063e9f0`, `FUN_005a0180` and `FUN_005738f0` all run for real. It prints the life the crush takes. `test_crush_divisor --crush` prints the same from d2d's `Monster::crush_divisor` of the same row, kind and player count.

RESULTS_PLACEHOLDER

## In play

A unique (rare) skeleton at 4,000 life hit by a 100 percent Crushing Blow: game.exe takes 1,000, the parent took 500. A boss row (Andariel, Duriel) or a superunique (Bishibosh) takes 500 on both sides. The size is the gap between a quarter and an eighth of the life, per proc, on a unique only (a champion, a minion and a plain monster were right).

## Limits

- The data word's value for each creator is read from the creators' code (`0x5a4567..0x5a4aa5`, three sites), not run: `crush_div.py` sets the word as the creators leave it. Whether any other path writes a plain unique's word with bit 2 was searched for in `.text` (the scan above; the setter `FUN_005a01b0` has no caller); a write through a pointer built another way would not appear.
- Whether a superunique also carries bit 8 (`strike.cpp`'s stun test) was not read.
- The get-hit size test counts the crush in d2d and does not in game.exe (D6 of the crush verification): not changed here.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cd build && ninja test_crush_divisor test_game_tables && cd ../tools/emu && uv sync
set -o pipefail
uv run python crush_div.py gen 8 > /tmp/c.tsv
uv run python crush_div.py game /tmp/c.tsv > /tmp/g.tsv
../../build/tests/test_crush_divisor --crush /tmp/c.tsv | grep -E '^[0-9]+ -?[0-9]+$' > /tmp/d.tsv
uv run python crush_div.py compare /tmp/c.tsv /tmp/g.tsv /tmp/d.tsv   # parent: differs, exit 1; claim: 0, exit 0
../../build/tests/test_crush_divisor && ../../build/tests/test_game_tables
```

---

Agent: Claude Code
