## What changes

A monster's missile can be blocked only if its physical record is not zero. The missile's hit entry `FUN_005ad730` calls `FUN_0057dfb0(defender, 1, can_block)` at `0x5ad806` with `can_block = [record + 8] != 0` (`0x5ad7f9`). With 0, `FUN_0057dfb0` skips the block roll (and its draw on the defender's seed, `0x57dfb4`) and runs `FUN_0057dd60`, which holds the dodge, avoid and evade rolls and the claws' weapon block. The record's physical field is rolled by `FUN_005a89a0` from the missile's stats 0x15 / 0x16 (`FUN_005a8910`: min and max both over 0, else 0). d2d rolled the shield's block for every monster missile (`monster_blow`, `combat.hpp`).

New `MonStats::unblockable`, set by `monster_shot_stats` for the MissA / MissS1 / MissC shots (`!(phys.low > 0 && phys.high > 0)`, the launch record in 256ths) and by `missile_blow` for a monster skill's missile (the row's own physical: Fallen Shaman fire, Radament's and the Countess's), and read by `monster_blow` for a missile. Two commits: the oracle (`tools/emu/missile_block.py` and `test_monster_missile --missile-block` / `--missile-block-skill`, which build and print what the parent does) and the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_005ad730` `0x5ad7f9..0x5ad806` | `cmp [record + 8], 0; setne` is the third argument of `FUN_0057dfb0` |
| `FUN_0057dfb0` `0x57dfb4` | argument 0 goes straight to `FUN_0057dd60`; otherwise the chance (`FUN_00622720`), a third while running, the roll on the defender's seed |
| `FUN_005a89a0` / `FUN_005a8910` | the record's physical: min and max both over 0 else 0 (field `+8`) |

## Counted differentials (V run: game.exe's own functions under the emulator, d2d built and run, same day)

`missile_block.py game` takes the physical min and max game.exe's own launch (`FUN_0059fa30`, the existing `mm_diff.py`) put on each case's missile, rolls the record with `FUN_005a89a0` and runs `FUN_005ad730` on it with `FUN_0057dfb0` hooked to read can-block. `test_monster_missile --missile-block` gives d2d's side: the case's shot through the real `monster_shot_stats` and `monster_blow` against a defender with Block 75 and nothing else, 400 trials; `compare` requires 0 blocks where game.exe's can-block is 0 and more than 0 where it is 1.

| Build | mm_diff cases differ (of 29,421; 579 Blood Raven excluded) | skill rows differ (of 45; 12 with no d2d row, listed) |
|---|---|---|
| oracle commit `13783a6d` (parent) | 4,550 | 42 |
| claim `cd314f2c` | 0 | 0 |

`compare` exits 1 on any difference, on zero compared cases and on a missing, repeated or malformed line on either side (`missile_block.py selftest`, 12 checks, no emulator). `test_monster_missile`, `test_game`, `test_combat`, `test_rules`, `test_monster_ai`, `test_monsters`, `test_game_rules`, `test_monster_thinks`, `test_missile_blow_poison`, `test_energy_shield`: rc 0.

## How to re-check

```sh
set -o pipefail
<build test_monster_missile as usual>
cd tools/emu && uv sync
export D2_MPQ_DIR=<the game data folder, with bin/game.exe>
uv run python mm_diff.py game TABLES_DIR 30000 20261115 OUT         # TABLES_DIR: monstats, missiles, skills, levels, monlvl by mpq-cat
uv run python missile_block.py game OUT/game.tsv > block.tsv
../../build/tests/test_monster_missile --missile-block OUT/cases.txt | grep '^K ' > d2d.tsv
uv run python missile_block.py compare block.tsv d2d.tsv         # claim: 29421 cases, 0 differ; parent: 4550 differ
uv run python missile_block.py selftest
cd ../.. && python3 tools/pr_guard.py upstream/main && python3 tools/presubmit.py --quick
```

## Honest limits

- Only the block roll is skipped; dodge, avoid, evade and the claws' weapon block are unchanged, as in game.exe. d2d's shared rng loses one draw per such missile: a distribution claim (blocked fraction 0 against about 75 percent), not a seed claim.
- Blood Raven's shots are excluded from the compare: her record carries her weapon's random damage (the existing `mm_diff.py` does not compare her physical part either).
- 12 skill-launch cases (`shafire2`, `shafire4`, `unholybolt2`, `unholybolt3`) have no d2d row: d2d loads only the Missiles.txt rows its monsters and skills name, and `andariel_missile` makes no missile for a missing row. They are listed (NOROW) and not judged; whether they matter in play is not run here.
- A skill missile whose owner adds damage (SrcDamage over 0) is not compared; only `andarielspray` has a record (can-block 1, matched).
- The record's later users (`FUN_0057c6c0`, `FUN_0057cee0`, ...) are stubbed in the oracle: only the block argument is claimed.

---

Agent: Claude Code
