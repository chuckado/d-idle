## What changes

A big hit puts the player in a get-hit from a swing or a cast only if the used skill's Skills.txt `interrupt` column is set, and never while the unit is in state 0x36 (uninterruptable). `FUN_005809d0`, the server's player mode switch (called by the hit at `0x57d383`), calls `FUN_0057edd0` (the mode table) and then, as its fifth argument is 0, `FUN_0057eec0(unit, new mode, 1, 0)`. The second tests state 0x36 first (`0x57eec7`: a Leap, Leap Attack or Whirlwind under way refuses every new mode); then it reads the unit's used skill (`FUN_00620250`, id by `FUN_00643ce0`): a row whose flags byte `+7` has the `interrupt` bit clear (the mask at `0x6ce284`, `0x57ef46`) refuses mode 4 and 9, unless the unit is idle; with no used skill it allows. Checked over all 357 rows, 22 have it clear: Kick, Jab, Strafe, Fend, Smite, Zeal, Charge, Concentrate, Frenzy, Whirlwind, the Identify and Town Portal scrolls and books, Werewolf, Werebear and six monster rows.

d2d tested `!attack_mode(pmode)`: no melee swing was ever interrupted (Attack and Sacrifice included) and every cast was (Werewolf and Werebear included, and a Leap in flight). Now `Fight::hit_interrupts` refuses while `Fight::uninterruptable()` (state 0x36), and otherwise reads the new `Skill::interrupt` for the swing and cast modes (A1, A2, SC, TH, KK, S1, S3, S4); the barrel kick, which d2d plays as KK with no skill id, is read as Kick, the skill game.exe kicks a barrel with (`FUN_005868a0` -> `FUN_00580a70`, which stores it as the used skill, `0x580b43..0x580b4c`). `Fight::cancel_swing` drops the swing's repeats, sequence and movement when the get-hit starts. Two commits: the oracle (`tools/emu/interrupt_swing.py` and `test_hit_recovery --interrupt`, which builds and prints what the parent does) and the claim.

Scope: the used-skill gate and state 0x36 only. #548 (a hit does not restart a get-hit) is merged and kept as it is in the get-hit condition; the mode table and the knock line stay as they are. Built on upstream main `77990456`.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `0x57d383`, `FUN_005809d0` `0x580a1c` | the get-hit call; with its last argument 0 it asks `FUN_0057eec0` after `FUN_0057edd0` |
| `FUN_0057eec0` `0x57eec7..0x57eed6` | state 0x36 first: refuses every new mode |
| `FUN_0057eec0` `0x57ef07..0x57ef4c` | used skill present, its row's `+7` against the mask: clear refuses (the jump table at `0x57efd2` is for other new modes) |
| `FUN_005868a0` `0x5868ce..0x58690e`, `FUN_00580a70` `0x580b43..0x580b4c` | the barrel kick starts with the unit's skill 1, Kick, and stores it as the used skill; Kick's `interrupt` is blank |
| `FUN_0057edd0` | A1, A2, SC, TH allowed, KK with an end event queued; class gates S1 (Amazon), S3 (Druid) |

## Counted differentials (V run: game.exe's own functions under the emulator, d2d built and run, same day)

`interrupt_swing.py game` hooks the used skill (`FUN_00620250` a fake pointer, `FUN_00643ce0` the row id) and the unit state (`FUN_00639df0` answers 1 for one state at a time: none, 0x36) over the real Skills.txt table and runs both functions for 357 rows x 8 modes x 2 states, with a mode-end event queued as a swing has one. `test_hit_recovery --interrupt` gives d2d's side: one big hit on a Paladin whose `swing_skill` is the row, in each mode, with repeats, a sequence and a movement pending, the movement flying (a Leap or Whirlwind under way) for state 0x36; and the barrel kick (`swing_skill` 0 in KK, what `World::operate` leaves), compared with game.exe's (Kick, KK) cell.

| Build | cells differ (of 5,711, plus the barrel kick) | state none (of 2,855) | state 0x36 (of 2,856) | barrel kick | own-mode cells differ (of 357) | interrupted with swing state left |
|---|---|---|---|---|---|---|
| oracle commit `113ab0d7` (parent) | 2,811 | 1,740 | 1,071 | agrees (refused by the attack-mode rule) | 210 | 1,005 |
| a claim without state 0x36 and the barrel rule, same oracle | 2,681 | 0 | 2,680 | differs (interrupted) | 0 | 0 |
| claim `5fa67510` | 0 | 0 | 0 | agrees | 0 | 0 |

`compare` exits 1 on any difference, on a missing, repeated or malformed line on either side, or on a cell that keeps swing state; `interrupt_swing.py selftest` runs 16 checks of that without the emulator. The Attack row in KK (swing_skill 0, KK) is d2d's barrel kick state, since Attack never plays KK, so that one cell is judged as the barrel kick against game.exe's Kick and not as Attack (the comparison prints it). With the claim's tests on the code without the two rules, `test_hit_recovery` aborts at its assertion for a Leap Attack in flight; with only the barrel rule removed it aborts at the barrel kick assertion and the comparison shows that one cell. `test_hit_recovery` (Attack and Sacrifice interrupted, Zeal and Smite not, a Leap Attack in flight and the barrel kick not), `test_game`, `test_combat`, `test_rules`, `test_swing_start`, `test_swing_players`, `test_classic_swing`, `test_game_rules`, `test_monster_ai`, `test_monsters`: rc 0. The own-mode column is the mode d2d's swing table gives the row (a Leap plays SC through `cast_missile`, not A1), not a count of cases that can happen in play.

## How to re-check

```sh
set -o pipefail
<build test_hit_recovery and mpq-cat as usual>
cd tools/emu && uv sync
export D2_MPQ_DIR=<the game data folder, with bin/game.exe>
export D2_PATCH_INSTALLER=<LODPatch_114d.exe>     # `game` loads the 1.14d tables through it (emu.py's default is under the home directory)
export D2_BUILD=<the build directory that holds tools/mpq-cat/mpq-cat>   # default "build" next to tools/
../../build/tests/test_hit_recovery --interrupt | grep '^[IB] ' > d2d.tsv
uv run python interrupt_swing.py game > game.tsv
uv run python interrupt_swing.py compare game.tsv d2d.tsv      # claim: 5,711 cells and the barrel kick, 0 differ; parent: 2,811 differ
uv run python interrupt_swing.py selftest
cd ../.. && python3 tools/pr_guard.py upstream/main && python3 tools/presubmit.py --quick
```

`EMU` (the folder with emu.py) is optional; it defaults to the script's own folder. Any other argument prints the usage.

## Honest limits

- The used-skill pointer's lifetime is what makes "only a swing or a cast holds it": it is cleared by idle, walk, run and the end of a swing (V run, the P04 drill R1 emulator part, 2,856 swing ends and 714 walk starts, 0 failures), and d2d reads `swing_skill` only in those modes.
- The gate has two more state branches that the oracle holds off and d2d does not model: with `interrupt` set, a unit in state 0xf (Concentrate, the Barbarian's) or state 0x2a (the Paladin's Concentration aura) still refuses the hit when `rand(100)` on the unit's seed is below its stat `0xa4` (`skill_concentration`, 20 for Concentration; `0x57ef51..0x57ef8f`). V run, one state at a time: 0xf refuses; 0x2a refuses with the stat at 100 and allows it at 0. So in game.exe a Paladin in a Concentration aura is not interrupted on about 20 percent of big hits (the value from Skills.txt, D); d2d interrupts him on every one (before this PR never). The comparison counts those cells (2,680 each for state 0xf and for 0x2a at stat 100, where game.exe differs from the no-state answer) and does not compare them. Modelling it takes one more random draw; left to a follow-up.
- d2d never casts a scroll or a tome: `World::read_portal` and the Identify use act at once, with no SC mode and no `swing_skill`, so rows 217 to 220 never reach `hit_interrupts`, before or after. The gate's answer for them (refused) has no d2d counterpart.
- A refused big hit plays no struck cue in game.exe (V read: only the too-small path sets `0x8000`); d2d's soft-hit branch plays it for a refused hit, as before for every swing. The weapon swing cue queued by `set_pmode` is not dropped when a swing is cut, so an Attack cut early still whooshes a few ticks later (game.exe's client plays it from the mode's frame, which the get-hit replaces: I). Not changed.
- Not changed: a block animation inside a swing (d2d's `pmode < 0` rule), a knocking hit inside a swing or a get-hit (`tick.cpp`'s knock line: game.exe also refuses a knockback, new mode 0x13, for a used skill with `interrupt` clear; d2d still knocks a Werewolf cast), and whether an interrupted click-attack resumes (d2d keeps `attack_mon`; game.exe not traced).
- The oracle runs the mode switch for a Paladin; class 0 in S1 and class 5 in S3 refuse every row (class gates of `FUN_0057edd0`, not part of this claim; `C` lines).
- The barrel kick case is d2d's state as `World::operate` leaves it (the test sets the same fields: `swing_skill` 0, KK), not a run through a loaded level's barrel object. The Leap Attack case sets a flying movement, as the Leap's start does, not a played Leap.
- Siblings: the melee and skill paths are touched and the missile path is not: a monster missile never sets the player's used skill (`swing_skill` is written only by the player's own swing, cast and summon paths).
- Data gating: `test_hit_recovery` needs the `interrupt` column and the rows it names (Attack, Sacrifice, Zeal, Smite, Kick, Leap Attack) in every data set it runs on; the 1.14d tables have them, and the test fails loudly if a row is missing.

## Checklist

- [x] A test in `tests/` (`tests/game/test_hit_recovery.cpp`: the `--interrupt` grid and the asserts) and an emulator oracle (`tools/emu/interrupt_swing.py`).
- [x] The finding is in `docs/research/re/combat.md` (the P04 bullet, with the addresses).
- [x] `python3 tools/pr_guard.py upstream/main` passes.

Agent: Claude Code
