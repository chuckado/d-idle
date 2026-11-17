## What changes

Open Wounds (item property 135) takes the life game.exe's handler gives it. d2d had the right per-level numbers but two other rules wrong:

1. **A champion, a unique or a superunique takes half.** The item event of stat 135 is function 15 of the event table at `0x7325b0`, `FUN_005bfe60` (Crushing Blow's `FUN_005bffc0` is function 16). It draws `rand % 100` under the stat (`0x5bfe9e..0x5bfec9`), takes `FUN_005bfda0(level, 9, 18, 27, 36, 45)` plus 40 (`0x5bff14..0x5bff20`: 9 a level up to 15, then 18 to 30, 27 to 45, 36 to 60, 45 above), and halves it toward zero when the struck monster's data word `+0x16` has bit 4 or 8 (`FUN_005a0180` with mask 12, `0x5bff3c..0x5bff56`: champion, unique, superunique, not a minion). d2d never halved.
2. **256ths a frame, 200 frames.** The handler hands `FUN_0056e970` stat `0x4a` (life regeneration, 256ths a frame, negated), state `0x3e` (`openwounds`) and 200 frames (`0x5bff8a..0x5bffa2`). d2d cut the value to whole points a second (`value x 25 / 256`, truncated) for 8 seconds: up to a quarter of the wound lost at level 1 (3 a second against 3.9).

`rules::open_wounds_fine(level, special)` is the value (it replaces `open_wounds_per_sec`); `kOpenWoundsFrames` is 200; `player_blow` puts it in `Blow::bleed_fine` when the roll lands; `Target::special` (champion, unique, superunique, set in `Monster::target`) is the halving; `Fight::land` sets the monster's `bleed_rate` to `bleed_fine / (256 x 40)` a millisecond for 200 frames (`monster_dots` is unchanged: its per-frame life change is the same arithmetic `FUN_005a6920` does with stat `0x4a`).

Two commits on upstream `ed007569`: the oracle `5e7a69b1` (`tools/emu/open_wounds.py`, `open_wounds_harness.cpp`, a README row), then the claim `dabc3ff4` (code, the `player_blow` pins in `test_combat` (49, halved 24), a `docs/research/re/combat.md` note, and `tests/game/test_open_wounds.cpp`: a blow through `Fight::land` and `monster_dots` on a plain monster, a champion, a unique, a superunique and a minion, game.exe's life lost at levels 2, 30 and 61; ctest runs both). The oracle's `compare` exits 1 when no monster case is compared.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| ItemStatCost row 135 | `itemevent1 domeleedamage` func 15, `itemevent2 domissiledamage` func 15 (Crushing Blow row 136 is func 16) |
| `0x7325ec` (entry 15; `0x7325f0` entry 16 holds `0x5bffc0`) | the handler `FUN_005bfe60` |
| `FUN_005bfe60` `0x5bfe9e..0x5bfec9` | the draw on the attacker's seed, `% 100`, under the stat |
| `FUN_005bfe60` `0x5bfed8..0x5bff20` | level (stat 12, at least 1), `FUN_005bfda0` over 9, 18, 27, 36, 45, plus 40 |
| `FUN_005bfe60` `0x5bff25..0x5bff56` | a player struck divides by 4 (by 2 more for a missile, event 6); a monster with data word `& 12` halves |
| `FUN_005bfe60` `0x5bff8a..0x5bffa2` | the record: 200 frames, stat `0x4a`, the negated value, state `0x3e`, to `FUN_0056e970` |
| `FUN_005a6920` `0x5a6926..0x5a69f9` | each frame: stat 6 (life, 256ths) plus stat `0x4a`, clamped to max (read, not run) |

## Counted differentials (V run, tree = upstream `ed007569` plus the branch)

Oracle: `open_wounds.py game` calls `FUN_005bfe60` natively with the chance fixed at 100 (the draw is the same LCG step the crush handler makes), an attacker level, a struck player or monster whose data word is set, and event 5 or 6, and reads the record handed to `FUN_0056e970`. The life the wound takes over its whole length (256ths) is compared: `value x frames` on game.exe, `open_wounds_per_sec x 8 x 256` on the parent, `open_wounds_fine x kOpenWoundsFrames` on the claim (`open_wounds_harness`). Rows that strike a player are printed by game.exe but not compared (d2d has no monster that wounds a player). The emulator compare's d2d side is the rule (`open_wounds_fine`); `test_open_wounds` checks the path (`Monster::target`, `player_blow`, `Fight::land`, `monster_dots`).

| cases | differ before (parent) | differ after |
|---|---|---|
| seed 7 (`gen 200 7`): 1,738 compared, 244 player rows not compared | **1,729** (plain 849 of 858, special 880 of 880) | **0** |
| seed 1 (`gen 20000 1`): 16,526 compared, 5,256 player rows not compared | **16,463** | **0** |
| seed 5001 (fresh): 3,069 compared | **3,054** | **0** |

The level grid is 1 to 99 x data words {0, 2, 4, 8, 0xa, 0xc, 0x10, 0x12} x events 5 and 6, then random. Bit 2 and bit 0x10 do not halve; bits 4 and 8 do.

Mutants of the claim (local, never pushed), each aborted by a test:

| Mutant | Aborted by |
|---|---|
| no halving | `test_combat` |
| halving rounded up | `test_combat` (`player_blow` pin, 24) |
| breakpoint 16 instead of 15 | `test_combat` |
| `player_blow` ignores `special` | `test_combat` |
| a minion counted as special | `test_open_wounds` (minion case) |
| rate taken per 25 ms instead of per frame | `test_open_wounds` |
| wound 200 x 25 ms | `test_open_wounds` |
| whole points a second (the parent's truncation in `Fight::land`) | `test_open_wounds` |

`test_combat` pins game.exe's values: levels 1, 15, 16, 30, 61, 99 = 40, 166, 184, 436, 1,426, 3,136; halved at 1, 30, 61, 99 = 20, 218, 713, 1,568; level 0 as 1. The pins use no random number. `test_open_wounds` fails on the parent plus that test alone (V run).

Build and tests (V run, GCC, Release, `-j2`, the builder): `test_combat` and `test_open_wounds` ok; full ctest 358 of 358 pass on the branch merged with upstream main `f1533770`.

## In play

Life taken from a plain monster over the 8 seconds, game.exe against the parent: level 1 31.3 against 24, level 5 59.4 against 56, level 12 108.6 against 104, level 30 340.6 against 336. A champion or unique takes half in game.exe and all of it on the parent. Halved, level 30 over 8 s: 340 points on a plain monster, 170 on a unique (game.exe; d2d after the claim the same, V run by `test_open_wounds`).

## Limits

- A player struck (a monster with the property, divided by 4): game.exe's rule is in the oracle's output, not modelled (no monster of d2d carries it).
- Which data-word bits the spawners set: bit 4 on a champion, bit 8 (`FUN_005a0320`) on every unique, champion and superunique placed, bit 0x10 on minions (read in the spawners; d2d's `Boss` kinds are mapped as its stun code already does: special is not none and not minion). Not run on a spawned monster.
- The chance draw's place among the events of one blow (crush, Open Wounds) follows game.exe's property order; d2d draws the crush first, as before.
- A second Open Wounds on a wounded monster refreshes it in d2d, as before; whether `FUN_0056e970` replaces or adds was not read.
- An act boss (MonStats boss, not a placed unique) is `Boss::none` in d2d and takes the whole wound; whether game.exe's act-boss spawn sets bit 4 or 8 of the data word was not read.
- In d2d's tick order `monster_dots` runs before the blows, so a wound applies 199 of its 200 updates in play (`test_open_wounds` starts at the hit tick and counts 200); where in a frame game.exe first applies the new regeneration was not run, and the end of poison and bleed (`<` against `<=`, and a monster left at one point with a fraction) is not changed here.
- The consumer `FUN_005a6920` is read, not run; d2d's `monster_dots` is unchanged.

## Answers to the presubmit WARNs

- Unit kinds: a merc (`merc_strike`: `player_blow` over its gear, its level) and a Shadow (the owner's fighter, the Shadow's level) wound through the same `player_blow` and `Fight::land`, so they take the new rule with their own level, as game.exe's handler reads the attacker's stat 12; a player struck is not modelled (no monster carries the property in d2d); a co-op partner's hero is a player (the same path).
- `build`, `gcc`: built and ctest run (Release, GCC): 358 of 358 pass on the branch merged with upstream main `f1533770`.

## Claims

| Commit | Claim or change | Evidence | Re-check |
|---|---|---|---|
| `5e7a69b1` | the oracle: game.exe's wound value against d2d's rule | seed 7: 1,729 of 1,738 differ on the parent; seed 1: 16,463 of 16,526; seed 5001: 3,054 of 3,069 | `open_wounds.py compare ...` exits 1 on the parent |
| `dabc3ff4` | special halves, 256ths a frame for 200 frames, `test_open_wounds`, `test_combat` pins | 0 differ on all three seeds; `test_open_wounds` aborts on the parent; mutants above all abort | `compare` exits 0; `ctest --test-dir build -R 'test_combat|test_open_wounds' --output-on-failure` |

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_combat test_game
cd tools/emu && uv sync
set -o pipefail
uv run python open_wounds.py gen 20000 1 > /tmp/c.tsv
uv run python open_wounds.py game /tmp/c.tsv > /tmp/g.tsv
g++ -O2 -std=c++23 -I../../components/rules -I../../components -I../../components/d2s -I../../components/txt \
    -o /tmp/open_wounds_harness open_wounds_harness.cpp && /tmp/open_wounds_harness < /tmp/c.tsv > /tmp/d.tsv
uv run python open_wounds.py compare /tmp/c.tsv /tmp/g.tsv /tmp/d.tsv   # parent: 16463 differ, exit 1; claim: 0, exit 0
../../build/tests/test_combat
cd ../..
ctest --test-dir build -R 'test_combat|test_open_wounds' --output-on-failure   # D2_MPQ_DIR set; test_open_wounds skips without the data
```

## Checklist

- [x] Tests: `tests/test_combat.cpp`, `tests/game/test_open_wounds.cpp`
- [x] The finding in `docs/research/re/`: `combat.md`
- [x] `python3 tools/pr_guard.py` passes (0 FAIL, 0 WARN)

---

Agent: Claude Code
