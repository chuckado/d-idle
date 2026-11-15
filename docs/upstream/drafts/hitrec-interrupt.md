# Draft (local, NOT posted): "A hit interrupts a swing or cast whose skill allows it" (P04)

> status: 2026-11-15 (card prfix7, TASK-509). Fork branch `fid/hitrec-interrupt-r3` (private fork remote only), two commits on upstream main `77990456`: oracle `113ab0d7`, claim `5fa67510`. Applies the independent verifier's required changes ([../verifier-p04-2026-11-15.md](../verifier-p04-2026-11-15.md) on `agent/verifyp04`): a Leap Attack in flight (state 0x36) and the barrel kick (Kick) refuse the get-hit, the oracle gets a unit-state dimension, wording and checklist fixed. **A second verifier pass is needed before posting**; not posted. #548 is merged. Body: [bodies/hitrec-interrupt-body.md](../bodies/hitrec-interrupt-body.md). Record: [../prfix7-p04-r3-2026-11-15.md](../prfix7-p04-r3-2026-11-15.md). Trace: [../../research/decomp26-trace.md](../../research/decomp26-trace.md).

PR title (when unblocked): `A hit interrupts a swing or cast unless its skill refuses it`

## What changes

`FUN_005809d0` (the server's player mode switch, called by the hit at `0x57d383`) asks two questions before a get-hit: `FUN_0057edd0` (modes, done by #548 and the earlier PRs) and, because its fifth argument is 0, `FUN_0057eec0`, which reads the unit's used skill (`FUN_00620250`). A used skill whose Skills.txt `interrupt` column is clear refuses the get-hit; with it set (Attack, Sacrifice, Vengeance, Conversion, most casts) the hit enters mode 4 from A1, A2, SC, TH, S1, S3, S4 (KK with an end event queued) and the swing's queued strike goes with the old mode.

d2d (`tick.cpp`): `!attack_mode(pmode)`, so no melee swing was ever interrupted (Attack and Sacrifice included) and every cast was (Werewolf and Werebear included; d2d reads scrolls and tomes at once, with no cast mode), also a Leap in flight. One new row field (`Skill::interrupt`), `Fight::hit_interrupts` (state 0x36 refuses first; a KK swing with no skill id, the barrel kick, is read as Kick) and `Fight::cancel_swing`.

## Claims

| Claim | Evidence | Re-check |
|---|---|---|
| game.exe's answer for each (skill row, mode, unit state) | V run (builder container, emulator): `interrupt_swing.py game` runs `FUN_0057edd0` and `FUN_0057eec0` for 357 Skills.txt rows x 8 modes x state none and 0x36, Paladin; state none byte-identical to the verifier's `game.tsv` (2,856 cells, 22 rows refuse); state 0x36 refuses all 2,856; also counted, not compared: state 0xf and state 0x2a with the roll at 100 refuse all, with it at 0 equal state none | `interrupt_swing.py game > game.tsv` |
| d2d before | V run (oracle `113ab0d7`): 2,811 of 5,711 cells differ (1,740 state none, 1,071 state 0x36), 210 of 357 own-mode, 1,005 keep swing state; the barrel kick agrees | `interrupt_swing.py compare game.tsv d2d.tsv` (exit 1) |
| the r2 claim, under the same oracle | V run: 2,681 differ: 2,680 state 0x36 cells (Leap Attack, Leap, Whirlwind in flight interrupted where game.exe refuses) and the barrel kick (interrupted, game.exe refuses) | same |
| d2d after | V run (claim `5fa67510`): 0 (5,711 cells and the barrel kick); the barrel-only mutant leaves 1 and aborts the unit assert | same |

## Limits (stated in the body)

- The used-skill pointer lives from the skill's start to the end of its swing or cast (idle, walk, run clear it): V run, drill R1 emulator part (`docs/research/warp-walk-game-trace.md` Part B, 2,856 swing ends and 714 walk starts, 0 failures).
- The gate's other two state branches (Concentrate 0xf, the Paladin's Concentration aura 0x2a, which refuses 20 percent of big hits) are not modelled; the oracle counts them and holds them off (V run, one state at a time).
- d2d never casts a scroll or tome (read at once, no cast mode): the gate's answer for them has no d2d counterpart.
- Not changed: the struck cue for a refused hit, the queued weapon swing cue of a cut swing, the block animation inside a swing, the knockback inside a swing or a get-hit and on a cast with `interrupt` blank (P03c), where an interrupted click-attack resumes.
- The barrel kick test sets the fields `World::operate` leaves, it does not run a level's barrel object; the Leap Attack test sets a flying movement.
