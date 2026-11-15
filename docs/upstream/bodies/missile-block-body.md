## What changes

A monster's missile can be blocked only if its physical record is not zero. The missile's hit entry `FUN_005ad730` calls `FUN_0057dfb0(defender, 1, can_block)` at `0x5ad806` with `can_block = [record + 8] != 0` (`0x5ad7f9`). With 0, `FUN_0057dfb0` skips the block roll (and its draw on the defender's seed, `0x57dfb4`) and runs `FUN_0057dd60`, which holds the dodge, avoid and evade rolls and the claws' weapon block. The record's physical field is rolled by `FUN_005a89a0` from the missile's stats 0x15 / 0x16 (`FUN_005a8910`: min and max both over 0, else 0). d2d rolled the shield's block for every monster missile (`monster_blow`, `combat.hpp`).

New `MonStats::unblockable`, set by `monster_shot_stats` for the MissA / MissS1 / MissC shots (`!(phys.low > 0 && phys.high > 0)`, the launch record in 256ths) and by `missile_blow` for a monster skill's missile (the row's own physical: the Fallen Shaman's fire and Radament's bolt), by `andariel_missile` for a row that names a Skill (the carried skill's physical, after it replaces the row's: a Death Mauler's maul has 0 in the row and 10 to 20 in DeathMaul, so it stays blockable), and read by `monster_blow` for a missile. Two commits: the oracle (`tools/emu/missile_block.py` and `test_monster_missile --missile-block` / `--missile-block-skill`, which build and print what the parent does) and the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_005ad730` `0x5ad7f9..0x5ad806` | `cmp [record + 8], 0; setne` is the third argument of `FUN_0057dfb0` |
| `FUN_0057dfb0` `0x57dfb4` | argument 0 goes straight to `FUN_0057dd60`; otherwise the chance (`FUN_00622720`), a third while running, the roll on the defender's seed |
| `FUN_005a89a0` / `FUN_005a8910` | the record's physical: min and max both over 0 else 0 (field `+8`) |
| `FUN_0064b860` (`0x64b9fd`, `0x64ba10`), run through `FUN_0059fa30` on a Death Mauler | a row that names a Skill carries the skill's physical in place of the row's: the maul's record is 2,560 to 5,120 (10 to 20) on normal, can-block 1 (V run) |

## Counted differentials (V run: game.exe's own functions under the emulator, d2d built and run, same day)

`missile_block.py game` takes the physical min and max game.exe's own launch (`FUN_0059fa30`, the existing `mm_diff.py`) put on each case's missile, rolls the record with `FUN_005a89a0` and runs `FUN_005ad730` on it with `FUN_0057dfb0` hooked to read can-block. `test_monster_missile --missile-block` gives d2d's side: the case's shot through the real `monster_shot_stats` and `monster_blow` against a defender with Block 75 and nothing else, 400 trials; `--missile-block-skill` does the same for a monster skill's missile through the real `andariel_missile` (at the level d2d launches it, with the monster skill it carries) and the launch records of `mm_launch.py`, which now also launches the Death Mauler's maul. `compare` requires 0 blocks where game.exe's can-block is 0 and more than 0 where it is 1.

| Build | mm_diff cases differ (of 29,421; 579 Blood Raven excluded) | skill-launch cases differ (of 60; 12 with no d2d row, listed; 3 Blood Raven excluded) |
|---|---|---|
| oracle commit (parent) | 4,550 | 42 |
| the claim ordered as an earlier head did (`unblockable` set from the row before the carried skill's physical is swapped in) | 0 | 15 (the Death Mauler's maul, all 15 launches) |
| claim | 0 | 0 |

`compare` exits 1 on any difference, on zero compared cases and on a missing, repeated or malformed line on either side (`missile_block.py selftest`, 12 checks, no emulator). The carried path (a row with a Skill: the Death Mauler's maul) is covered by both the oracle's skill cases (game.exe: can-block 1 on the launch record 2,560 to 5,120, normal difficulty) and a unit case in `tests/game/test_carried_missile.cpp` (row physical 0 with a carried physical: blockable; carried physical 0 with a row physical: unblockable; no Skill and no physical: unblockable; 400 shots through the real `monster_blow`); the case fails when `andariel_missile` sets `unblockable` before the swap (checked: abort).

## How to re-check

```sh
set -o pipefail
<build test_monster_missile as usual>
cd tools/emu && uv sync
export D2_MPQ_DIR=<the game data folder, with bin/game.exe>   # emu.py is found next to missile_block.py (EMU=<dir> only when run from elsewhere)
uv run python mm_diff.py game TABLES_DIR 30000 20261115 OUT         # TABLES_DIR: monstats, missiles, skills, levels, monlvl by mpq-cat
uv run python missile_block.py game OUT/game.tsv > block.tsv
../../build/tests/test_monster_missile --missile-block OUT/cases.txt | grep '^K ' > d2d.tsv
uv run python missile_block.py compare block.tsv d2d.tsv         # claim: 29421 cases, 0 differ; parent: 4550 differ
uv run python missile_block.py selftest
uv run python mm_launch.py TABLES_DIR > launch.out                   # the skill launches (with the Death Mauler's maul)
uv run python missile_block.py skill-cases launch.out > skill_cases.txt && uv run python missile_block.py game-skill launch.out > skill_block.tsv
../../build/tests/test_monster_missile --missile-block-skill skill_cases.txt | grep -E '^[KN] ' > d2d_skill.tsv
uv run python missile_block.py compare skill_block.tsv d2d_skill.tsv   # claim: 60 cases, 0 differ; parent: 42 differ
cd ../.. && python3 tools/pr_guard.py upstream/main && python3 tools/presubmit.py --quick
```

## Honest limits

- Only the block roll is skipped; dodge, avoid, evade and the claws' weapon block are unchanged, as in game.exe. d2d's shared rng loses one draw per such missile: a distribution claim (blocked fraction 0 against about 75 percent), not a seed claim.
- Blood Raven's shots are excluded from the compare: her record carries her weapon's random damage (the existing `mm_diff.py` does not compare her physical part either).
- The grid is `mm_diff.py`'s Act 1 / 2 shooter rows (36 rows, 51 monsters) and the skill launches of `mm_launch.py` (the Death Mauler's maul added). Act 3 to 5 shooters go through the same `monster_shot_stats` (read in the source) but were not run.
- The area path: game.exe also reaches `FUN_005ad730` from `0x57e203` (a loop over units that copies a caller's record); the same `record + 8` rule applies there. d2d's monster area missiles were not checked.
- The Countess's and the vampires' firewalls are not changed by this PR: d2d already burns them without a block roll (a direct burn in `ai.hpp`, no `monster_blow`), so the claim names only the Fallen Shaman's fire, Radament's bolt and the Death Mauler's maul among the skill missiles.
- 12 skill-launch cases (`shafire2`, `shafire4`, `unholybolt2`, `unholybolt3`) have no d2d row: d2d loads only the Missiles.txt rows its monsters and skills name, and `andariel_missile` makes no missile for a missing row. They are listed (NOROW) and not judged; a separate loading fix.
- A skill missile whose owner adds damage (SrcDamage over 0) is compared only for `andarielspray`.
- The record's later users (`FUN_0057c6c0`, `FUN_0057cee0`, ...) are stubbed in the oracle: only the block argument is claimed.

## Checklist

- [x] test in `tests/` (`test_monster_missile` unit and `--missile-block` modes, `test_carried_missile` for the carried path)
- [x] finding in `docs/research/re/` (`combat.md`)
- [x] `tools/pr_guard.py upstream/main` passes (0 FAIL, 0 WARN)

Presubmit WARN `siblings` (hit paths): melee. The monster's own melee block gate is a separate PR (P07); the missile and skill paths are covered here through `monster_shot_stats`, `missile_blow` and `andariel_missile`.

---

Agent: Claude Code
