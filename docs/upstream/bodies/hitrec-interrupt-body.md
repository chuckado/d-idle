## What changes

A big hit puts the player in a get-hit from a swing or a cast only if the used skill's Skills.txt `interrupt` column is set. `FUN_005809d0`, the server's player mode switch (called by the hit at `0x57d383`), calls `FUN_0057edd0` (the mode table) and then, as its fifth argument is 0, `FUN_0057eec0(unit, new mode, 1, 0)`. The second reads the unit's used skill (`FUN_00620250`, id by `FUN_00643ce0`): a row whose flags byte `+7` has the `interrupt` bit clear (the mask at `0x6ce284`, `0x57ef46`) refuses mode 4 and 9, unless the unit is idle; with no used skill it allows. Checked over all 357 rows, 22 have it clear: Kick, Jab, Strafe, Fend, Smite, Zeal, Charge, Concentrate, Frenzy, Whirlwind, the Identify and Town Portal scrolls and books, Werewolf, Werebear and six monster rows.

d2d tested `!attack_mode(pmode)`: no melee swing was ever interrupted (Attack and Sacrifice included) and every cast was. Now `Fight::hit_interrupts` reads the new `Skill::interrupt` for the swing and cast modes (A1, A2, SC, TH, KK, S1, S3, S4) and `Fight::cancel_swing` drops the swing's repeats, sequence and movement when the get-hit starts. Two commits: the oracle (`tools/emu/interrupt_swing.py` and `test_hit_recovery --interrupt`, which builds and prints what the parent does) and the claim.

Scope: the used-skill gate only. #548 (a hit does not restart a get-hit) and the mode table stay as they are; this PR is built on upstream main and rebased on #548 once it merges (a one-line conflict at the get-hit condition).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `0x57d383`, `FUN_005809d0` `0x580a1c` | the get-hit call; with its last argument 0 it asks `FUN_0057eec0` after `FUN_0057edd0` |
| `FUN_0057eec0` `0x57ef07..0x57ef4c` | used skill present, its row's `+7` against the mask: clear refuses (the jump table at `0x57efd2` is for other new modes) |
| `FUN_0057edd0` | A1, A2, SC, TH allowed, KK with an end event queued; class gates S1 (Amazon), S3 (Druid) |

## Counted differentials (V run: game.exe's own functions under the emulator, d2d built and run, same day)

`interrupt_swing.py game` hooks the used skill (`FUN_00620250` a fake pointer, `FUN_00643ce0` the row id, `FUN_00639df0` no state) over the real Skills.txt table and runs both functions for 357 rows x 8 modes, with a mode-end event queued as a swing has one. `test_hit_recovery --interrupt` gives d2d's side: one big hit on a Paladin whose `swing_skill` is the row, in each mode, with repeats, a sequence and a movement pending.

| Build | cells differ (of 2,856) | own-mode cells differ (of 357) | interrupted with swing state left |
|---|---|---|---|
| oracle commit `308e96f8` (parent) | 1,741 | 210 | 1,005 |
| claim `a78f66a3` | 0 | 0 | 0 |

`compare` exits 1 on any difference, on a missing, repeated or malformed line on either side, or on a cell that keeps swing state; `interrupt_swing.py selftest` runs 9 checks of that without the emulator. `test_hit_recovery` (Attack and Sacrifice interrupted, Zeal and Smite not), `test_game`, `test_combat`, `test_rules`, `test_swing_start`, `test_swing_players`, `test_classic_swing`, `test_game_rules`, `test_monster_ai`, `test_monsters`: rc 0.

## How to re-check

```sh
set -o pipefail
<build test_hit_recovery as usual>
cd tools/emu && uv sync
export D2_MPQ_DIR=<the game data folder, with bin/game.exe>
../../build/tests/test_hit_recovery --interrupt | grep '^I ' > d2d.tsv
uv run python interrupt_swing.py game > game.tsv
uv run python interrupt_swing.py compare game.tsv d2d.tsv      # claim: 2856 cells, 0 differ; parent: 1741 differ
uv run python interrupt_swing.py selftest
cd ../.. && python3 tools/pr_guard.py upstream/main && python3 tools/presubmit.py --quick
```

## Honest limits

- The used-skill pointer's lifetime is what makes "only a swing or a cast holds it": it is cleared by idle, walk, run and the end of a swing (V run, the P04 drill R1 emulator part, 2,856 swing ends and 714 walk starts, 0 failures), and d2d reads `swing_skill` only in those modes.
- A refused big hit plays no struck cue in game.exe (V read: only the too-small path sets `0x8000`); d2d's soft-hit branch plays it for a refused hit, as before for every swing. Not changed.
- Not changed: a block animation inside a swing (d2d's `pmode < 0` rule), a knocking hit inside a swing or a get-hit (`tick.cpp:660`), and whether an interrupted click-attack resumes (d2d keeps `attack_mon`; game.exe not traced).
- The oracle runs the mode switch for a Paladin; class 0 in S1 and class 5 in S3 refuse every row (class gates of `FUN_0057edd0`, not part of this claim; `C` lines).
- Whether d2d's tome and scroll use reaches the SC cast with the portal skill as `swing_skill` is not run end to end.

---

Agent: Claude Code
