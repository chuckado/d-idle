# decomp27: P07 a monster's melee block gate, P03c a knocking hit inside a get-hit (game.exe 1.14d vs d2d)

> status: finished 2026-11-15 (card decomp27, Sonnet). Offline: emulator on the NAS builder (the image of the merged oracles), no game, no live d2d, nothing posted. Fork branches (private fork remote, base upstream `77990456`, nothing force-pushed): `fid/melee-block-r1` (oracle only, `3d3d6b25`) and `fid/knock-in-gethit-r1` (oracle `64047f23`, claim `4e911f21`). Tags: V run (game.exe's own code under the emulator, compared with d2d's REAL rules), V read, D (tables), S, I, U. Follows [decomp26-trace.md](decomp26-trace.md) (P06, the missile half of the same gate) and the P03b trace (`docs/upstream/verifier-p03b-2026-11-14.md`).

## Verdict

1. **P07, a monster's melee swing: d2d already matches. No PR.** The melee hit entry `FUN_0057ec10` passes can-block = 1 and missile = 0 as constants (`0x57ece1`..`0x57ece8`), so there is no gate to port: every monster swing and every monster skill's hit that rolls goes through it. 180 of 180 cells (19 defender modes minus the town walk, 10 block chances) agree between game.exe's `FUN_0057ec10` and d2d's `monster_blow`. Not a PR; the oracle is on `fid/melee-block-r1` for a verifier.
2. **P03c is proven and built.** game.exe refuses a knockback (new mode 0x13) from GH, BL, DT and DD; d2d knocked a player standing in a get-hit and moved its hold. Parent 1 of 13 cells differ (GH), claim 0. Branch `fid/knock-in-gethit-r1`, two commits, ready for the Opus verifier.

## Questions

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q1 | Who calls the block roll `FUN_0057dfb0` and with what can-block? | Exactly three call sites (scan of every `call` in `.text`): `0x57ece8` (the melee hit entry `FUN_0057ec10`, `push 1 / push 0 / push defender`: can-block 1, missile 0), `0x5ad806` (the missile hit `FUN_005ad730`, can-block = record +8 != 0, P06) and `0x56ba0d` (`FUN_0056b9c0`, can-block 0, missile 1: a unit-enumeration callback, below). | V read |
| Q2 | Who reaches `FUN_0057ec10` (melee)? | 33 call sites. By the skill tables: the basic Attack (do-function 1, `FUN_0056f070`, site `0x56f151`), the skill hit `FUN_0056e680` (`0x56e6b3`; `ResultFlags` bit 0 set skips it, `0x56e68e`), Jab (do 7), Charge (do 67), Whirlwind (do 76), double swing / throw, Power Strike (start 6, `FUN_005da940`) and the other melee do-functions. All pass the same constants: the melee entry has no way to ask for can-block 0. | V read |
| Q3 | Which monster skills skip it? | Only `ResultFlags` bit 0 (D, skills.txt): DiabRun (5) and Siege Beast Stomp (1). d2d's `sure` (`result_flags & 1`, `ai.cpp:1630`) is the same test. Charge's hit and the cold touch / Corpse Explosion records are sure by their own code (`ai.cpp:2134`, `2174`, `2447`), as the existing comments read. | V read + D |
| Q4 | Is there any can-block 0 path for a melee-like hit? | `FUN_0056b9c0` (called through the callbacks `FUN_0056ba90` and `FUN_005d3cb0`, passed by `FUN_005d3e80` and `FUN_005d7b60`, do-functions 38 "release area hit" and 53): can-block 0, missile 1, so such a skill's area hit is never blocked. No monster skill in skills.txt uses do 38 or 53 (D: 0 of 147 monster skills), so d2d has no monster path to compare; the player's own use is not this card's. | V read + D |
| Q5 | What does the block roll do for a player defender? | `FUN_0057dfb0` (`0x57dfb0`..`0x57e08b`): chance `FUN_00622720`; when `FUN_00622d00` (the moving test) holds and the mode is not WL (2), the chance is divided by 3 (`imul 0x55555556`, truncating); roll = the defender's 64-bit LCG (`lo x 0x6ac690c5 + hi`, `rand % 100`) `< chance`; returns 1 on a block. The moving test holds in RN (3) and TW (6) and in no other mode. | V read + V run (mode scan below) |
| Q6 | game.exe against d2d over modes and chances (P07)? | `mel_block.py game` runs `FUN_0057ec10` natively (hooked: the chance, dodge, to-hit, reach, bookkeeping) with a monster attacker and a player defender in modes 0..18 and chances 0, 1, 5, 9, 10, 20, 25, 33, 50, 75, 400 swings per cell (seeds from a fixed xorshift): 190 cells. The blocked count equals the seed-exact LCG reading in every cell (`R` lines: 0 mismatches; before the TW fix of the reader 9 TW cells differed: the third applies in TW, Q5). The only argument pair `FUN_0057dfb0` sees is can-block 1, missile 0 (`A 1 0`). `test_melee_block` counts the REAL `rules::monster_blow` (to-hit sure; WL walking, RN running, every other mode standing): `mel_block.py compare` = **180 cells, 0 differ** (TW excluded: d2d has no town fight, game 93 of 400 at chance 75 there). Mutants: d2d's mode GH at 75 forced to 0 blocks, and RN at 75 forced to 300: 1 differ each, exit 1. | V run |
| Q7 | Does game.exe's knockback ask `FUN_0057edd0` for mode 0x13? | `FUN_0057cee0` (takes damage) at `0x57d332`: when the hit's flag bit 8 is set it calls `FUN_00580a70` with new mode `0x13` (`push 0x13` at `0x57d350`, call `0x57d356`) and returns, so a knocking hit never also tries the get-hit. `FUN_00580a70` calls `FUN_0057edd0` first (`0x580adc`) and ends without a mode change when it returns 0. | V read |
| Q8 | What does `FUN_0057edd0` answer for 0x13? | New modes above 0x11 take the current-mode jump table (`0x57ee08`..`0x57ee17`), not the new-mode table. V run of the function (class 3, hold queued = 1, the P03b setup extended to new mode 19): **refused from DT, GH, BL and DD; allowed from every other mode, including the swing modes and KB (0x13 from 0x13).** With queued = 0 (nothing pending) the swing and cast modes A1, A2, SC, TH are refused too (d2d models a running hold, queued = 1). Classes 0..6: no class-dependent 0x13 cell beyond the known modes 13 and 15 cells of classes 0 and 5, which hold for 4 and 9 as well. | V run |
| Q9 | d2d before and after (P03c)? | `test_hit_recovery --knock-modes` (the real `Fight::hurt_side`, a 1-point knocking hit, hold running): parent `mode_switch.py compare-knock` **13 cells, 1 differ (GH: d2d starts the slide and moves the hold, game.exe refuses), 7 excluded** (dead modes DT, DD; swing modes A1, KK, S1, S3, S4: d2d never lets a knock interrupt a swing, game.exe allows it: P04, D2 below). Claim: **13 cells, 0 differ**; the P03b cells (`compare`, big hit) still 13 / 0; `selftest` 12 of 12 (the parent's selftest plus the knock checks, one missing-mode, one mutant). `test_hit_recovery` (with the new assert: a knock in GH moves neither hold nor position), `test_game`, `test_combat`, `test_rules`, `test_swing_start` rc 0. | V run |

Commands (builder, `D2_MPQ_DIR` the game data, `GAMEEXE` game.exe): P07 `mel_block.py game 400 > game.tsv`, `test_melee_block 400 | grep '^D ' > d2d.tsv`, `mel_block.py compare game.tsv d2d.tsv`. P03c `mode_switch.py game > game.tsv`, `test_hit_recovery --knock-modes | grep '^K ' > knock.tsv`, `mode_switch.py compare-knock game.tsv knock.tsv` (parent `4e911f21^` build: exit 1; claim: exit 0).

## Deviation list (d2d vs game.exe)

| id | Rule | game.exe | d2d | Tag | Differential |
|---|---|---|---|---|---|
| D1 | A knocking hit in a get-hit | refused (`FUN_0057edd0`, new mode 0x13 from GH) | before: starts the slide, moves the hold; claim: refused | V run both sides | yes, built (`compare-knock`), claim on `fid/knock-in-gethit-r1` |
| D2 | A knocking hit in a swing (A1, KK, S1, S3, S4; A2, SC, TH when idle) | allowed (queued 1) | never interrupts a swing (`!attack_mode`, `tick.cpp:674`) | V run game side, d2d V run (the `excluded` lines) | yes, same harness; NOT in the claim: P04 changes what a hit does inside a swing, so it follows P04 |
| D3 | A monster's melee swing and block | can-block 1 | same | V run 180 cells | built, 0 differ: no deviation |
| D4 | Town walk (TW) third | the block chance is a third in TW as in RN | d2d has no town fight | V run game side | not needed (no hostile melee in town) |
| D5 | An area hit by a skill (do 38, 53) | can-block 0 | monsters use neither | V read + D | none: no monster skill uses them; a player's release skill is another card |
| D6 | BearSmite (do 2, ResultFlags 8, kick function `FUN_0056f1f0` has no call to `FUN_0057ec10`) | does not roll the melee entry through that function | not traced | U | read `FUN_0056f1f0`'s damage call and the BearSmite owners; a lab drill (below) |

## Bot findings (described, not changed)

B1 No bot code models a monster's melee block or a knockback: grep of `internal` for "knock" finds only data and walking (`internal/data/extra.go`, `internal/nav/walk.go`, `internal/nav/mouthgate.go`); for block chance only `internal/boss/act3.go:34` (a mode-name comment). Nothing to fix. A future Paladin kill-time model: the player's block is `FUN_00622720` (playerdef-trace Q9); a running player blocks a third in RN and TW only (Q5), walking blocks at the full rate (`WL` is not a third: V run). The earlier bot note B2 of monster-attack-trace (run policy blocks a third) stays true for RN.

## PR candidates

- **P03c: yes (local, not posted).** `fid/knock-in-gethit-r1` on the private fork remote: oracle `64047f23` (`mode_switch.py` new mode 19 and `compare-knock`, `test_hit_recovery --knock-modes`, README row), claim `4e911f21` (`tick.cpp:674`: `&& (actor->pmode != kModeGH || actor->knock)`, with an assert in `test_hit_recovery`). Parent 1 differ (exit 1), claim 0. `presubmit --quick` 0 FAIL / 2 WARN, `pr_guard.py upstream/main` 0 FAIL / 0 WARN. Not done: clang-tidy, the whole ctest (the verifier runs them). Brief: [verify-knock-gethit.md](../upstream/briefs/verify-knock-gethit.md); draft [hitrec-knock-in-hit.md](../upstream/drafts/hitrec-knock-in-hit.md), body [hitrec-knock-in-hit-body.md](../upstream/bodies/hitrec-knock-in-hit-body.md). Depends on #548 (merged upstream; the branch is on `77990456`, which has it).
- **P07 melee: no PR candidate.** d2d matches (Q6). The oracle commit `3d3d6b25` on `fid/melee-block-r1` is a regression test for the merged behaviour; whether it is worth upstreaming is the maintainer's call (tools only, no claim). Brief: [verify-melee-block.md](../upstream/briefs/verify-melee-block.md) asks the verifier to try to break the negative result.

## Drill recipe (lab binary, `docs/forkdbg.md`, `docs/forkcheats.md`; writing only, nothing run)

K1 (D3 melee block, U for the whole pipeline in d2d): give the lab Paladin a block of 75 (a shield and Holy Shield or the cheat for stat 20), `debug spawn fallen1 20` next to her, `debug mon set <id> dmg 1 1`, stand still and `step` through 400 swings; count mode 9 (BL) plays: about 75 percent of landed blows. Walk (`debug`-driven path) and run: expect about 75 percent while walking, 25 percent running.
K2 (D1 knock in GH): a Paladin with Smite, a monster in reach, `debug hit` a get-hit first, then a Smite hit within the hold: expect no slide and the hold end unchanged (the claim), the slide only from NU / WL / RN.
K3 (D6): a bear (`debug spawn` of a BearSmite owner, if the lab has one) hits a shielded player; whether blocks occur is the U.

## Matrix-cell updates (proposed text)

- `skill-matrix` MA10 / MA13 (monster melee): "UPDATE 2026-11-15 (decomp27): a monster's melee block gate does not exist: `FUN_0057ec10` passes can-block 1 and missile 0 (V read, `0x57ece1`); `mel_block.py` 180 of 180 cells equal d2d's `monster_blow` over modes and chances (V run); only RN and TW get a third (V run); a skill with `ResultFlags` bit 0 (DiabRun, Siege Beast Stomp) skips the whole roll, d2d `sure` matches (V read)".
- Hit recovery matrix cell for "knock in GH": "DEV, fixed on `fid/knock-in-gethit-r1`: new mode 0x13 refused from DT, GH, BL, DD, allowed from every other mode (V run); d2d refused BL already, GH was the one cell (parent 1 of 13 differ, claim 0); knock in a swing is D2, after P04".

## Missing glossary terms (for the report)

"can-block" (the third argument of `FUN_0057dfb0`), "queued event" (the type-1 mode-end event `FUN_005415a0` returns), "new mode 0x13" (the knockback request of `FUN_00580a70`).

## Shared-file lines for the merger

See `backlog/merge-lines/decomp27.txt`.
