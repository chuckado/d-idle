## What changes

A player's Crushing Blow takes the life the target has BEFORE the hit, in 256ths. FUN_005bffc0 is a strike-event handler: FUN_0057c6c0 raises the attacker's melee event 5 (a missile's event 6 doubles the divisor, `0x5c00a2`) and the defender's 1 before it takes the record's damage at `LAB_0057c852`; the handler reads the defender's stat 6 as it is then and subtracts `life / divisor` (256ths, whole division) less `MulDiv(that, min(res, 100), 100)` when the damage resist is above 0 (`0x5c00aa..0x5c00e2`). d2d took `(hit points - physical - elements) / divisor` in whole points, the life left after the hit; the crush was smaller whenever the hit was above zero.

`player_blow` (`components/rules/combat.hpp`) now computes `crushing_fine = life_fine / (divisor x missile factor)` less the resist share with `mul_div`, and adds it to the blow's fine total (`damage` whole points, `frac` 256ths). It uses `Target::life_fine`, added by the leech change this branch is stacked on (`fid/leech-r1`). `test_combat` pinned `(400 - 50) / 4 = 87` less 50 percent (44); it now pins `400 / 4 = 100` less 50 percent (50) and the missile row 12.5 hp (12 and 128 / 256ths).

Two commits: the oracle (`tools/emu/crush.py`, `crush_harness.cpp`, a README row), then the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0057c6c0` `0x57c7b7..0x57c7f3` | after the record is built (FUN_0057c1e0), the strike events: attacker 5 then defender 1 (melee, second argument 0), or 6 and 2 (missile), through FUN_005c0c30 |
| `FUN_0057c6c0` `0x57c7f8..0x57c826`, `0x57c852..0x57c879` | only then the physical is cut to the life and FUN_0057c420 leeches (`0x57c826`), and the record's total `+0x4c` comes off stat 6 at `LAB_0057c852` |
| `FUN_005bffc0` `0x5bfff6..0x5c0026` | the chance: a draw on the attacker's seed `% 100` under the item stat (136) |
| `FUN_005bffc0` `0x5c0031..0x5c00a8` | the divisor: 4; a boss or unique 8 (FUN_0063e9f0, FUN_005a0180); a hireling 10 (FUN_0063ee90); plus the player-count share (FUN_005738f0); doubled for event 6 |
| `FUN_005bffc0` `0x5c00aa..0x5c010b` | `life / divisor` less the resist share, stat 6 set (0 when not above 0), the kill bit `rec + 4 \| 2` |

## Counted differentials (V run, tree = `fid/leech-r1` plus the branch)

Oracle: `crush.py game` runs FUN_005bffc0 on game.exe natively (a defender with stat 6 = the case's life, stat 0x24 = its resist, the divisor selected by the hireling / boss hooks of the 10 and 8 branches) for each case and prints the life it takes. `crush.py order` runs FUN_0057c6c0 natively with the dispatcher FUN_005c0c30 replaced by a recorder and prints the defender's life at events 5 and 1 and at the end. `crush.py compare` sets the first against `crush_harness`, which runs d2d's real `player_blow` with Crushing Blow 100 percent and prints the blow's damage less its physical, both in 256ths. Grid: 6 lives (1 to 2,000 points) x 4 hits x 5 resists {0, 25, 50, 100, -20} x 3 divisors {4, 8, 10}, then 2,000 random.

| cases | differ before (parent) | differ after |
|---|---|---|
| 2,360 | **2,223** (res 0: 960 of 960, res under 0: 142 of 142, res over 0: 1,121 of 1,258) | **0** |

Order (`crush.py order`, FUN_0057c6c0 natively, 2,360 cases): the life at event 5 and at event 1 is the life before the damage in 2,360 of 2,360, the end life is `life - damage` (0 under one point) in 2,360 of 2,360.

Mutants of the claim (each is caught by the compare): no resist share 1,258 differ, whole points of life 1,905, the life after the hit (the old rule) 2,241.

Build and tests (V run, GCC 14, Release, on the builder): `presubmit.py --quick --base upstream/main` 0 FAIL, `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN, `test_combat` and `test_game` as the leech branch. Full ctest: the result is in the branch's task card.

## In play

A 100 hp monster hit for 30 with a 100 percent Crushing Blow: game.exe takes 25 + 30, the parent took 17 + 30. A boss (divisor 8) at 2,000 hp hit for 60: 250 against 242 on the parent. The gap is `hit / divisor` per proc.

## Limits

- The monsters' MonProp Crushing Blow on the player, the merc and the pets (`monprop_hit`, `tick.cpp:634`, `:762`) is not changed here: it uses the same FUN_005bffc0 and the same FUN_0057c6c0 order, so the game also takes the life before the hit there, but that is a different subsystem and the d2d side has no oracle yet (a follow-up).
- Which events the item property registers FUN_005bffc0 on is read from the handler's own tests (event 6 doubles the divisor) and the event order of FUN_0057c6c0; the item-property table that assigns events was not read.
- Open Wounds, the other half of the same hit, was not traced.

## Answers to the presubmit WARNs

- Unit kinds: the diff touches the player's blow. Pets and the merc call `player_blow` too (`pets.cpp`, `deaths.cpp`) and take the same new crush; game.exe runs one handler for every attacker, so the rule is the same there. A monster attacker's crush is the limit above.
- Hit paths: the melee path is what the oracle covers; the missile path reaches the same function with `swing.missile`, which still doubles the divisor (`test_combat` pins the 12.5 hp row).
- Build: `--quick` skips it; the builds are above.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_combat test_game
cd tools/emu && uv sync
set -o pipefail
uv run python crush.py gen 2000 7 > /tmp/c.tsv
uv run python crush.py game /tmp/c.tsv > /tmp/g.tsv
uv run python crush.py order /tmp/c.tsv > /tmp/o.tsv   # life at events 5 and 1, then the end life
g++ -O2 -std=c++23 -I../../components/rules -I../../components -I../../components/d2s -I../../components/txt \
    -o /tmp/crush_harness crush_harness.cpp && /tmp/crush_harness < /tmp/c.tsv > /tmp/d.tsv
uv run python crush.py compare /tmp/c.tsv /tmp/g.tsv /tmp/d.tsv   # parent: 2223 differ; claim: 0, exit 0
../../build/tests/test_combat
```

---

Agent: Claude Code
