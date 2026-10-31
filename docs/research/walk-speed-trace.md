# Walk speed and movement step: game.exe 1.14d vs d2d vs our bot (agent DECOMPWALK, 2026-10-31)

> status: research, current (card decompwalk, decomp-queue item 12). Offline only: no game instance, no d2d slot; nothing pushed, posted or filed. Follows docs/research/movement-and-recovery-trace.md (DECOMPREC), which this doc extends; d2d read at upstream main d0ea14b7 (worktree `trace/walk` on the worker host). Tags: V = read in game.exe 1.14d (the read-only copy, md5 2b3600c1...) or run in its own code under unicorn (`tools/emu/walk_speed.py`); V-d2d = read in d2d's source; D = patched tables via mpq-cat (docs/process/mpq-data-reading.md); S = upstream docs or ours; I = inference; U = unverified.

## Answer in four lines

1. **Player speed: equal for our bot's case.** Every class walks at CharStats WalkVelocity 6 and runs at RunVelocity 9 (D, loader 0x613472 / 0x613498 name them +0x40 / +0x41): 0.375 / 0.5625 subtiles a tick (1.875 / 2.81 cells/s). d2d gets the same numbers with no faster run/walk and no velocity skill, which is our early Sorceress.
2. **Monster speed: d2d deviates.** game.exe moves a monster in ANY moving mode, RN included, at MonStats **Velocity** x walk %; d2d moves RN at MonStats **Run**. A running Zombie (Velocity 1, Run 3) is 3x too fast in d2d. Two smaller terms (chill, floor) and one player term (velocitypercent diminished) come with it. Proven exhaustively: 712,362 cases, 425,665 differ before, 0 after (table below). Local PR candidate built.
3. **None of that explains the bot's walking time.** The deviations make d2d monsters faster (harder), not the bot slower. The bot spends 68 % of its time per kill walking (8.27 of 12.19 virtual s, post-fix soak), and the time is lost on the bot side (finding B1 below: after the first fresh game the bot most likely walks for the rest of the session, I), plus routes and stalls.
4. Bot fixes first: call `ResetRun` on every new game (B1), then log stamina and the run belief so a soak can show the run share.

## Q1 player speed (rules)

| Address | Rule | Tag |
|---|---|---|
| `00623f50` 0x624132..0x6241c9 | moving mode: walk % = `00621740`(row 4) + stat 0x43 (velocitypercent), at least 25 (0x62414e: `cmp esi, 0x19`); path velocity = `00621360` x % / 100 (signed, truncated), stored by `00648690` at path +0x7c / +0x84; the animation rate uses the same % | V (read; oracle) |
| `00621740` row 4 | stat 96 (faster run/walk) only: 150 v / (150 + v) for ANY v != 0, negative too (`test ecx, ecx; je` skips only 0); C division | V (read; oracle `frw` in move_recovery.py) |
| `00621360` | velocity base: player CharStats +0x40 (WalkVelocity) << 8 whatever the mode; monster MonStats +0x32 (Velocity) << 8 whatever the mode | V (read); field names V (loader descriptors 0x65161a Velocity +0x32, 0x651640 Run +0x34) |
| `00620e80` | a running player gets stat 0x43 + (RunVelocity x 100 / WalkVelocity - 100) = +50 | V (read) |
| `0057f240` | stamina a running tick: CharStats RunDrain (+0x42) x 2 x (body armor speed (+0xd8) / 10 + 1), less stat 154 % (signed magic for / -100), at least 1; none in town (`0061ab00`); at 0 the run ends | V (read) |
| `0057af80` | chill: stat 0x43 += MonStats coldeffect (byte +0x168 + difficulty) for a monster, -50 for a player; skipped when coldeffect is 0; then `00623f50` again so the slow applies mid-move | V (read) |

## Q2 the movement step

Velocity is set once when a moving mode starts (and again when a stat change calls `00623f50`, e.g. chill). Each tick `0064fe40` advances by `velocity x unit vector (length 4096) >> 8` in 16.16 subtiles, i.e. velocity / 4096 subtiles, the same on a diagonal (V, DECOMPREC oracle `step`). No acceleration, no start-up delay. d2d: `follow_path` with `cells_per_sec(vel) x elapsed` per 40 ms tick (tick.cpp:356..361, gamedata.hpp:810), equal to < 1/4096 subtile; a re-issued move re-plans in the same tick (tick.cpp:345), so a re-click costs no ticks in d2d (V-d2d). The click-to-move pathing (type 7, A* near / straight far) is S (DECOMPREC, upstream `moves.py` diff).

## Q3 monster walk / run

game.exe: velocity = Velocity << 8 x max(100 + stat 0x43 deltas, 25) / 100 in WL and RN. Run is read only where an AI turns it into a pace, `Run x 100 / Velocity - 100` clamped (0x5eccec, 0x5ed07a, 0x5ed461, 0x5ed7e3, 0x5f0d15 Baboon, 0x5f6e2b ZakarumZealot, 0x5e48e0 pets), and by the charges (0x5cf808, 0x5d9fdd: path velocity = Run << 8) (V). d2d (ai.cpp:2724..2727, 643): RN moves by `cells_per_sec(Run)`, chill multiplies the whole % by (100 + coldeffect) / 100, floor 10 (V-d2d). 57 of 73 MonStats rows with an RN mode have Run != Velocity (D); early ones: zombie1 1 / 3, corruptrogue2 6 / 8.

## Q4 differential (tools/emu/walk_speed.py + walk_speed_harness.cpp)

Exhaustive grid: every MonStats row x WL (and RN where MonStats2 has it) x deltas -100..200 x chill; every class x walk/run x FRW -60..200 x velocitypercent -50..100 step 5 x chill. game.exe's own code from 0x624113 to 0x6241ce; d2d line for line before / after the candidate; a mismatch is >= 1 unit of 1/4096 subtile a tick.

| Term | before | after | what differs |
|---|---|---|---|
| M1 monster RN moves by Run (rows with Run != Velocity) | 34,190 | 0 | Zombie RN 0.31 vs 0.94 cells/s; realistic slice (deltas -50..100, no chill) 8,607 of 11,023 RN cases |
| M2 chill multiplies instead of adding coldeffect | 124,301 | 0 | chilled champion (+20, coldeffect -50): 70 % vs 60 % |
| M3 monster floor 10 instead of 25 | 58,588 | 0 | only below 25 % (Spider Lay, deep slows) |
| P1 player velocitypercent diminished with FRW | 207,466 | 0 | Vigor +13 % gives +11 % in d2d; none for our early Sorceress |
| P2 negative FRW not diminished | 1,120 | 0 | -20 gives -23 in game.exe; no early source |
| total (712,362 cases) | 425,665 | 0 | player 208,586 / 226,548, monster 217,079 / 485,814 |

## Q5 our bot, and where the walking time goes

Soak `postfix1` (bot f8c1b7f, run policy on, 1,253 virtual minutes, read-only greps on the worker host): per kill 12.19 virtual s = walk 8.27 (68 %), cast 0.85, wait 3.09 (6,176 `killtime` lines). Rule speeds: walk 9.375, run 14.06 subtiles/s. Realised: net displacement between 30 s `status` samples 3.18 subtiles/s while clearing with no hostile near, 1.93 with hostiles (a lower bound: turns and back-and-forth hide inside a sample); out-of-mana back-off steps advance their target 7.8 subtiles per virtual s (3,970 step pairs, 10.4 subtiles a step every 0.81 s), 83 % of the WALK rule and 55 % of the run rule. Stall lines: 1,797 `nav stuck`, 635 `stalled by monsters`, 1,375 `combat move failed` (no path), 1,113 re-plans, 619 cave-mouth step-aways; 6,040 back-off steps (62 virtual min, 5 %). So the realised speed is well below the rule: the bot loses the time (walking instead of running, stalls, routes), not d2d's physics.

| id | finding (bot side) | file:line | tag |
|---|---|---|---|
| B1 | `ResetRun` has no caller: the run belief survives a fresh game or death, d2d starts every game walking, and with belief "running" and a full bar the policy never presses `r` again, so the bot walks until stamina could drain, which a walk never does. 19 fresh games in postfix1. Call it where a game starts (d2denv / lifecycle) | internal/d2dinput/run.go:66, controller.go:42 | V (grep), effect I |
| B2 | the run state is unobservable: no stamina or belief in any log line (0 `stamina` lines), so no soak can show the run share; add both to the `minute` line | internal/d2dinput/run.go:69..87 | V |
| B3 | `MoveTo` re-issues `cmd move` after 400 ms without 1 subtile of progress and kicks once; costs no d2d ticks, but each stall waits 400 ms of virtual time | internal/d2dinput/controller.go:185..238 | V-read, cost I |
| B4 | no travel-time model in nav (no speed constant found); a model would need 9.375 / 14.06 subtiles/s | internal/nav (grep) | V (grep) |

> correction 2026-10-31: the B1 premise is WRONG: `ResetRun` is wired (`d2denv.Build` sets `Proc.OnInGame = ctl.ResetRun`, a method value the grep missed); the walking share is still unexplained and the live proof recipe of docs/fixrun.md decides.

What the gap must come from: B1 (up to 1.5x on every leg after the first restart, I), then route length and stalls. A live check is needed for B1's real share: one d2d game with a restart, logging stamina.

## PR candidate (local only)

Branch `fix/walk-velocity` on upstream d0ea14b7 in the worker host's `trace/walk` worktree, 3 commits (one claim each, authored by the user, body `Agent: Claude Code`): 96953bfe M1, 7e8feb87 M2 + M3, 97d55732 P1 + P2. Upstream test_game / test_rules / test_combat pass, pr_guard 0 FAIL / 0 WARN against upstream/main (docs/research/pr-walk-evidence-pack.md, proof pr-walk-proof.md, body pr-walk-draft.md). Not pushed or posted; the independent verifier is not run yet (the next step), nor a det state-hash run.
