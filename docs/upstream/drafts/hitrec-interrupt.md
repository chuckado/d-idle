# Draft (local, NOT posted): "A hit interrupts a swing or cast whose skill allows it" (P04)

> status: 2026-11-15 (card decomp26). Fork branch `fid/hitrec-interrupt-r1` (private fork remote), two commits on upstream main `fb6d799d`: oracle `308e96f8`, claim `a78f66a3`. **Ready for the independent verifier**, not for posting: it edits the get-hit line of `fight/tick.cpp` that #548 (open) also edits, so it is posted after #548 merges (the rebase is a one-line conflict at `tick.cpp:715` and the comment above it). Body: [bodies/hitrec-interrupt-body.md](../bodies/hitrec-interrupt-body.md). Trace: [../../research/decomp26-trace.md](../../research/decomp26-trace.md).

PR title (when unblocked): `A hit interrupts a swing or cast unless its skill refuses it`

## What changes

`FUN_005809d0` (the server's player mode switch, called by the hit at `0x57d383`) asks two questions before a get-hit: `FUN_0057edd0` (modes, done by #548 and the earlier PRs) and, because its fifth argument is 0, `FUN_0057eec0`, which reads the unit's used skill (`FUN_00620250`). A used skill whose Skills.txt `interrupt` column is clear refuses the get-hit; with it set (Attack, Sacrifice, Vengeance, Conversion, most casts) the hit enters mode 4 from A1, A2, SC, TH, S1, S3, S4 (KK with an end event queued) and the swing's queued strike goes with the old mode.

d2d (`tick.cpp:715`): `!attack_mode(pmode)`, so no melee swing was ever interrupted (Attack and Sacrifice included) and every cast was (Town Portal and Identify scrolls and tomes, Werewolf and Werebear included). One new row field (`Skill::interrupt`), `Fight::hit_interrupts` and `Fight::cancel_swing`.

## Claims

| Claim | Evidence | Re-check |
|---|---|---|
| game.exe's answer for each (skill row, mode) | V run (worker host): `interrupt_swing.py game` runs `FUN_0057edd0` and `FUN_0057eec0` for 357 Skills.txt rows x 8 modes (A1 A2 SC TH KK S1 S3 S4), Paladin: 335 allowed, 22 refused in each mode (the 22 rows with `interrupt` blank); class gates only (Amazon in S1, Druid in S3 refuse every row) | `interrupt_swing.py game > game.tsv` |
| d2d before | V run (NAS builder, oracle commit `308e96f8`): `test_hit_recovery --interrupt`: 1,741 of 2,856 cells differ; in the 357 cells where the mode is the one d2d plays the skill in, 210 differ (every melee skill that game.exe interrupts, and the casts it does not); 1,005 interrupted cells keep swing state (repeats, sequence, movement) | `interrupt_swing.py compare game.tsv d2d.tsv` (exit 1) |
| d2d after | V run (claim `a78f66a3`, code identical to the built `b4d81433`): 0 of 2,856 differ, 0 of 357 own-mode; no cell keeps swing state; `test_hit_recovery` (Attack and Sacrifice go to GH and lose `kicks_left`, Zeal and Smite do not), `test_game`, `test_combat`, `test_rules`, `test_swing_start`, `test_swing_players`, `test_classic_swing`, `test_game_rules`, `test_monster_ai`, `test_monsters`: rc 0 | same |

## Limits (stated in the body)

- The used-skill pointer lives from the skill's start to the end of its swing or cast (idle, walk, run clear it): V run, drill R1 emulator part (`docs/research/warp-walk-game-trace.md` Part B, 2,856 swing ends and 714 walk starts, 0 failures). d2d's `swing_skill` is read only while the mode is a swing or cast, so a stale value is never read.
- Not changed: the struck cue. A refused big hit sets no flag `0x8000` in game.exe (V read, `0x57d383` path); d2d's soft-hit branch still plays it (as for every swing before this PR).
- Not changed: the block (BL) animation inside a swing (`foe.blocked && pmode < 0`, hitrec D4), the knockback inside a swing (`tick.cpp:660`, P03c), and where the interrupted click-attack resumes (`attack_mon` is kept, as for any other mode; game.exe U).
- A skill d2d does not build (`skill_built`) swings as a plain Attack with `swing_skill` 0 and is interrupted as Attack (the row of skill 0 has `interrupt` set).
- Whether d2d's tome / scroll use goes through the SC cast with the portal skill as `swing_skill` (so that the refusal reaches it) is U: checked only with the unit test, not with the real item click.
