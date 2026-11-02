# Backend parity audit: original game (memory backend) vs d2d

> status: current (card parityaudit). Read-only audit; no game, no instance, no soak was started. Decision behind it: d2d stays the long-term primary target, the bot resumes on the ORIGINAL game (memory backend under Wine on the worker host) as the ground-truth validation backend (user decision 2026-11-02).

Tags: V run (a command ran, output read), V read (code or doc read), D (data table), S (secondary: a doc or trace says so), I (inference), U (unverified). "M" = memory backend (Game.exe, `d2.Reader` over `/proc/pid/mem`, `d2input` helper), "D" = d2d backend (control socket, `d2dsrc` + `d2denv`). Status words per cell: PRESENT-TESTED (ran against the real game of that backend), PRESENT-UNTESTED (code and offline tests only), PARTIAL, ABSENT, UNKNOWN. "Frozen" = no live Game.exe run since 2026-10-30 (last result: the Fresha soak in docs/g1-log.md).

## 1. Build and test state

| Check | Where | Result | Tag |
|---|---|---|---|
| `go build ./cmd/d2bot ./cmd/d2soak ./cmd/d2dsoak ./cmd/d2input ./cmd/d2cheat` | this checkout (worktree of current main) | exit 0, `gofmt -l internal cmd` empty | V run |
| `go vet ./internal/mem ./internal/d2 ./internal/input ./internal/d2dsrc` and `go test -short -count=1 -p 1 ./internal/mem ./internal/d2 ./internal/input ./internal/d2dsrc ./internal/cheat` | this checkout | all `ok` | V run |
| `labrun the worker host -- sh -c 'go vet ./... && go test -short -count=1 -p 4 ./...'` run from inside the worktree | the worker host | exit 0 after 120 s; 88 packages `ok`, 0 FAIL, the rest "no test files" | V run |
| Per-package run with the memory-path set (`internal/mem d2 env input cheat nav town combat d2map d2dsrc d2dcombat d2dtown`, `cmd/d2bot d2soak d2input d2cheat d2dsoak`, `go vet` first) | the worker host | exit 0 after 19 s | V run |
| `tools/soak/soak24_test.sh`, `failsafe_test.sh`, `fresh_test.sh`, `contsoak_test.sh` (offline shell tests of the Game.exe soak driver) | this checkout | all PASS / ok | V run |

The memory backend path compiles: `d2bot` without `--backend d2d` is `runLive` (cmd/d2bot/live.go:41); `--backend` defaults to `game` (cmd/d2bot/d2dbackend.go, flag `backend`, env `D2BOT_BACKEND`), d2d is chosen at cmd/d2bot/main.go:216 (V read).

Backend-contract tests (spec backend-contract): the compile-time assertions `var _ d2.StateReader = (*Reader)(nil)` (internal/d2/statereader.go:23), `d2dsrc.Source` (internal/d2dsrc/statereader.go:5), `d2denv.Reader`, `d2/fake.Reader` hold (V run: they build and the packages pass). The d2d-specific contract tests (d2dcombat belt keys and casts, d2dtown seams, d2dsrc statereader) pass. There is NO backend-contract test that runs the same scenario against both backends: the memory reader is covered by recorded fixtures (internal/d2/testdata town.snap, town_inv.snap; internal/d2/fixture_test.go) and by `mem.Fake`, the d2d source by `internal/d2dctl/fake` (V read).

Skipped tests (V run, `go test -short -v`, only 4 in the packages above):
- `TestRealD2D` (internal/d2dsrc/integration_test.go:21): needs `D2D_SOCK`, a live d2d.
- `TestRealSaves` (internal/cheat/d2s/d2s_test.go:118): needs `D2S_REAL` and copies of real `.d2s` files; this is the only test that reads real Game.exe saves.
- `TestRealSkillNames` (internal/combat/class_test.go:478): needs `D2LOD_DATA_CACHE`.
- `TestPlanLargeLevelSpeed` (internal/nav/plan_test.go:231): skipped by `-short` only.
Environment-conditional skips that did not trigger here: `TestProcUnmappedReadIsTyped` and `TestProcInstance` (internal/mem/source_test.go:89,110), banner assets, guard lock. There is no test that attaches to a real Game.exe process (`mem.OpenGame`, `internal/mem/proc.go:151`); that path is only exercised by live runs (V read).

No failing test. Counting caveat: whole-repo results are one run on one commit; use `git log` for the commit.

## 2. Feature parity matrix

Counts at the end of the table. Evidence: `file:line` is V read unless another tag is given. "g1-log" = docs/g1-log.md (V run by the soak agents, read here as S).

| # | Capability | Memory backend (M) | d2d (D) | Evidence |
|---|---|---|---|---|
| 1 | Movement and pathing (click-to-move, level map, stuck recovery) | PRESENT-TESTED | PRESENT-TESTED | M: g1-log attempts 14 to 18 and G2 8 h walked Act 1; NAVSTUCK `reclick` live in attempt 17; level map from `d2map.Tracker` over memory. D: docs/d2d-walk.md, Gate 1 rerun 2026-11-01 (docs/soaks/gate1-20261101.md) |
| 2 | Teleport primitive | PRESENT-UNTESTED | PRESENT-TESTED | Wiring is shared: cmd/d2bot/telewire.go:15 called from behaviors.go:89 (M) and d2dwire.go:69 (D). No Game.exe run had Teleport (Soaky and Fresha never reached it, I); d2d tuned it in cards teletune2/3 (openspec/changes/archive/2026-11-01-teletune*, V-d2d) |
| 3 | Town portal use and tome count | PRESENT-TESTED | PRESENT-TESTED | M: tome quantity via the unit's stat list, empty tome has no list (internal/town/shop.go:123-136, "LIVE, i4"); g1-log attempt 6 "no town portal: walking back", attempt 18 TP trips. D: the client portal command path `d2dPortal` (cmd/d2bot/d2dwire.go:212) is d2d-only: on M `ToTown` still uses the cast/click path, so the `enterportal` change (3c25dcb) is untested on Game.exe (U). Tome price differs (250 on d2d vs 450 live, docs/d2d-gaps.md (a)) |
| 4 | Belt and potion logic (fill, drink, counts) | PRESENT-TESTED | PARTIAL | M: shift-click belt fill live (g1-log attempt 11 "belt: moved 3 potions"), potion counters. D: no click modifiers (internal/town/beltfill.go:232 falls to drag), belt keys are `cmd belt N` (internal/d2dcombat/input.go, spec backend-contract), Gate 1 rerun failed the belt window (6 min, limit 5: docs/g1-log.md last section) |
| 5 | Merc handling (read, revive, hire, guard input) | ABSENT (wiring) | PRESENT-UNTESTED | The merc policy is wired only on a d2d env: `newMercWireD2D` (cmd/d2bot/mercwire.go:76-85), nothing in live.go/behaviors.go. The memory reader exists (internal/d2/merc.go:42 `FindMerc`, V read) but is unused by the bot loop. D: behind `merc.enabled` default 0 (mercwire.go:24), V offline only; hire row pick U (merc_adapter.go:11) |
| 6 | Shopping and vendor | PRESENT-TESTED | PRESENT-TESTED | M: g1-log attempt 6 (50 purchase lines), attempt 8 (78 `shop: bought`). D: d2dtown vendor dump (spec backend-contract "Vendor stock becomes stored item units"), Gate 1 rerun |
| 7 | Inventory and stash | PRESENT-TESTED | PRESENT-TESTED | M: INVSTASH and stash banking, g1-log attempts 17 and 18. D: d2dtown + `d2dsrc/items.go` (V-d2d soaks) |
| 8 | Skill casting incl. cast gap | PARTIAL | PRESENT-TESTED | Cast-first loop live on M (g1-log attempt 14: 30 kills/min). The per-skill cast gap uses `castTicks(name, fcr, d2d)` (internal/combat/castgap.go:30) with the game.exe branch `d2d=false` (set from `Fighter.D2D`, internal/combat/fighter.go:31, hard.go:151); the model is V on game.exe code (oracle) but what a command DURING a cast does on Game.exe is U (docs/fixcastgap.md section 1). D: V-d2d (docs/d2d-casts-real.md) |
| 9 | Buff retry (armor recast) | PARTIAL | PRESENT-TESTED (clock), PRESENT-UNTESTED (state path) | Clock path is backend-neutral and DRILLS13 proved it on d2d (docs/fixbuff.md "Live evidence"). State path needs `Player.States`: M reads it from the stat list (internal/d2/player.go:170, `UnitStates`), ids 20/30/88 "not seen switching live (U)". D: states only from `debug states` via `Options.BuffStates`, needs a fork verb (internal/d2dsrc/source.go:49-52), live check open (fixbuff.md "Live proof recipe (not run)") |
| 10 | Melee model inputs (stats 19, 21, 22, 25, 119, 325) | PRESENT-UNTESTED | PARTIAL | M: `meleeBlow` reads `p.Stats` of the unit's merged stat list (internal/combat/melee.go:36-61); that the player list carries 21/22 as the wielded weapon's range is U (docs/skmeleewire.md:13, drill R1); absent stats give a nil Blow, never a wrong one. D: `gearStats` derives 19/25 and 21/22 from the lab2-stats dump (`gear`, `weapons`) (internal/d2dsrc/gearstats.go:23); no weapon speed (U: no source); starter kit read live gives `gear:[]` (docs/forkdbg.md:69); an older fork binary yields nothing |
| 11 | Resist reads | PRESENT-UNTESTED | PARTIAL | M: `Player.Resist` from stats (internal/d2/player.go:182); full stat list available. D: only the resist panel (fire, cold, light, poison; magic absent) and base attributes: `Stats.Has(id)` is false for everything else (internal/d2dsrc/player.go:208-288). Neither side has a live check of the resist numbers against the panel in the reviewed docs (I) |
| 12 | Player curse and poison state | PRESENT-UNTESTED | ABSENT | M: `States` bit array (internal/d2/player.go:170, `Poisoned()` :88). D: "no state list in dump (poison, curses, buffs); poisoned() is blind" (internal/d2dsrc/player.go:204); spec backend-contract names poison state a stub gap |
| 13 | Monster identity, kind flags, enchants, life | PRESENT-TESTED | PARTIAL | M: flags byte, 9 enchant ids, unique no, boss name read at the struct (internal/d2/units.go:199-217; name pointer V live Blood Raven, comment :203-207); LifePct for damaged monsters U (units.go:29-31). D: Flags synthesised from the boss column, enchants from `Mods`, level and HP present, `UniqueNo` never set (internal/d2dsrc/items.go:351-377) |
| 14 | Monster child states (curses, auras, minion mark) | PRESENT-TESTED | ABSENT | M: `ReadStates` on the monster's stat list (units.go:214-216), `StateMinion` V live (docs/classes/smoke-necromancer.md finding 1; necro curses use it, internal/combat/necro.go:379,434). D: "no child state list (curses, minion mark)" (internal/d2dsrc/items.go:209) |
| 15 | Monster missiles in flight (dodge input) | UNKNOWN | PRESENT-UNTESTED | M: `Reader.Units(TypeMissile)` walks the client unit table, which is UNVERIFIED ("none in flight during recording", internal/d2/units.go:89-91). D: `FillMissiles` over the fork verb `debug missiles`, speed-0 rows dropped (internal/d2dsrc/missiles.go:15-50); verb vetted on the fork (docs/fixmissile.md:16), not in a gate run |
| 16 | Monster AI state (aware, target, attack in progress, wake) for the new threat findings | ABSENT | ABSENT | Neither `d2.MonsterInfo` (units.go:24-36) nor the d2d builder carries it; the findings (docs/research/monster-ai-trace.md, monster-attack-trace.md) are decomp traces, the bot sees only mode, position, class, flags, life. Whether the memory side could read the AI struct is U (not read by any code) |
| 17 | Area transitions and waypoints | PRESENT-TESTED | PRESENT-TESTED | M: WPUSE live (g1-log attempt 17), no arrival cooldown on Game.exe (internal/env/tuning_nav.go:34, docs/research/waypoint-trace.md S). D: d2dworld exits and link wait (docs/d2d-bot.md), Gate 1 rerun |
| 18 | Quest and campaign reads | PRESENT-TESTED | PRESENT-TESTED | M: `.d2s` save reader (`saveQuestReader`, cmd/d2bot/task.go:92) plus `questmem` (internal/d2/questmem.go); campaign steps ran live (g1-log "Go to the Barracks" loops). D: internal/d2dsrc/quests.go, `--goals campaign` in the Gate 1 supervisor |
| 19 | Death handling (detect, corpse run, repeat-site ban, death rule) | PRESENT-TESTED | PRESENT-TESTED | M: many live deaths and autopsies (g1-log attempts 14 to 18, G2, Fresha); corpse run failed once ("unreachable", attempt 15/17). D: same code on `d2.Player`; invariant DEATHS in `d2dsoak` (spec soak-gates). Detection is `combat.IsDead(p)` (internal/bot/observe.go:154), backend-neutral |
| 20 | Save, exit, restart, recycle | PRESENT-TESTED | PRESENT-TESTED | M: save-and-exit `xpLost 0` (g1-log attempts 6, 8, 18), kill recovery 92 s (G2), supervise restart 5 s after OOM (attempt 16). Known: the `restarts` counter counted 1 of 17 recycles on M (STATUS risk 5; whether the `internal/bot/observe.go:220-251` accounting fixed that on M is U, no live run since). D: `d2denv.Cycler`, Gate 1 rerun kill back in about 3 s |
| 21 | Lockstep and virtual time | ABSENT (by design) | PRESENT-TESTED | M: `env.RealClock` (internal/env/env.go:28), no pacer: 2 h gate = 2 h wall. D: `--lockstep`, 120 wall min = 22.6 virtual h (docs/gate-run.md in the ops repo; Gate 1 rerun) |
| 22 | Fault injection used by the gates | PRESENT-TESTED (manual) | PRESENT-TESTED (scripted) | M: `kill -9` of the instance's Game.exe by hand: attempt 6, 8, G2 (92 s, one restart, no orphan winedevice); attempt 15 and 18 injection denied by the agent permission classifier, G2 done with the user's one-off permission (g1-log). D: `inject()` in the gate supervisor kills the d2d under the bot, registry pgid checked (gate-sup.sh:63-73, ops repo) |
| 23 | Character selection and creation | PRESENT-TESTED | PRESENT-TESTED | M: `--char`, OOG `start`/`create`, `tools/soak/fresh.sh` (Fresha, V live; fresh_test.sh PASS). D: `--d2d-class` new character or saved `--char` (internal/d2denv/Cycler; Gate1a) |
| 24 | Kit prep by cheat tool | PRESENT-TESTED | PRESENT-TESTED (different tool) | M: `internal/cheat` memory and `.d2s` edits, ledger (docs/cheats.md); a cheated character must not count toward a gate (tools/soak/ledgercheck.sh). D: `debug` verbs of the lab fork (docs/forkcheats.md, drills). Not interchangeable |
| 25 | Hidden instance | PRESENT-TESTED | PRESENT-TESTED | M: i1..i5 on private Xvfb displays, one Wine prefix each (ops repo docs/multi.md, V run 2026-10-28). D: headless d2d slot (`d2dslot`) |
| 26 | UI interaction (menus, hover, click modifiers) | PRESENT-TESTED | PARTIAL | M: `d2input` helper, window input, modifiers. D: semantic verbs only; `input.ErrUnsupported` for difficulty and modifiers (docs/d2d-bot.md "Launch Flow", beltfill.go:232) |

Counts (rows 1 to 26; a cell with two states counts under its first):
- Memory backend: PRESENT-TESTED 16, PRESENT-UNTESTED 4, PARTIAL 2, ABSENT 3, UNKNOWN 1.
- d2d: PRESENT-TESTED 16, PRESENT-UNTESTED 2, PARTIAL 5, ABSENT 3, UNKNOWN 0.

Reading the matrix: where M is weak the cause is almost always "never run since the feature was written" (rows 2, 3 portal command, 8, 9, 10, 11, 15), not "code missing". The real code gaps on M are rows 5 (merc wiring) and 16 (AI state, absent on both). The real gaps on D are the monster and player state reads (rows 11, 12, 13, 14), which is why Game.exe is the right place to validate anything that depends on curses, minion marks, unique numbers or the stat vector.

## 3. Gate tooling

What each tool assumes (V read):

| Tool | d2d assumption | Evidence |
|---|---|---|
| `gate-run.sh` (ops repo, head box) | `--binary NAME` is a file of the fork's bin dir, a fork hash is mandatory (exit 3 without), a d2d slot (`d2dslot`) is taken, the fresh character name `Gate<N><letter>` comes from d2d saves, wall hours are converted to virtual hours | docs/gate-run.md (ops repo) "Usage", "What it does" 1 to 3 |
| `gate-sup.sh` (ops repo, runs on the worker host) | starts `d2bot --backend d2d --managed --d2d-args "--lockstep" --d2d-linkwait 60s --seed N --goals campaign` and the fault `inject()` kills the d2d child, found by command line, of a registered slot group | gate-sup.sh:99-102, :63-73 |
| `gate-judge.py` (ops repo) | counts `d2d process exited` lines against injected faults (line 179), runs `d2dsoak compare -gate v1 -hours <vmin/60> -baselines baselines-early.json`, refuses fewer than 10 minute lines; the Gate 3 rows read the d2d memory curve | gate-judge.py:126-137, :179-186, :295-299 |
| `d2dsoak` (cmd/d2dsoak) | `run` starts `bot --backend d2d` for N seeds (lockstep); `extract` and `compare` read minute lines from ANY log, "original-game soak logs too" | cmd/d2dsoak/main.go header, V read |
| `tools/d2dsoak/baselines*.json` | the Gate 1 XP baseline (mean 9172 XP/h) is Game.exe's own, only 2 runs, weak (docs/d2d-baselines-early.md section 1); on Game.exe the "at or above baseline" criterion compares the original with itself | S |
| minute lines (`internal/bot/observe.go:957`) | none: emitted by the backend-neutral observer with `packHP`, `packMP`, `beltBoxes`, `xp`, `kills`, `restarts` | V read |
| `soak24.sh` + `tools/soak/*` + `cmd/d2soak` | none (Game.exe tooling): `start iN --gate` = 2 h, profile m1, managed bot, `--supervise`, sidecar alerts, winedevice plateau, d2soak judge. Only the criteria list differs from the d2d gate (spec soak-gates) | tools/soak/soak24.sh header; its four shell tests PASS (V run) |

Gaps to run Gate 1 on the original game in real time (2 h wall):
1. No `--backend game` in gate-run / gate-sup / gate-judge. Cheapest route: keep `soak24.sh start iN --gate --char NAME` as the runner (it worked for attempts 1 to 18) and add what gate-run gives the d2d side: no model polling, a state file, a collected record, one-line result.
2. Fault injection: `kill -9` of ONE instance's `Game.exe` found by the instance marker (`mem.FindGamesFor`), not a d2d under the bot. Recovery expectation from G2: back in game in at most 100 s, `restarts` +1, no orphan winedevice. The attempts of 2026-10-30 show the agent permission classifier refusing an interactive `kill -9`: a scripted supervisor on the worker host needs the user's explicit go for the injection (U whether the classifier sees it).
3. Judge: add the Game.exe exit lines (`Game lost: restarting it`, `recycle done`, g1-log attempt 8) beside `d2d process exited`; use real minutes for `-hours` (`vmin/60` already is: one minute line per minute); keep `d2soak report -profile m1` as the stability/death judge.
4. Character: a FRESH honest character (`tools/soak/fresh.sh iN sorceress NAME`, name from `tools/soak/fresh-names.txt`), never cheated (`ledgercheck.sh`). The cheat tool is for kit prep in the SMOKE only; a cheated character cannot be a gate character.
5. No lockstep: 2 h is 2 h; the fidelity and XP/h windows (NOKILL 10 min, BELT 5 min) are minute-line based and need no change. The d2d "virtual hour" conversions in the record must be skipped.
6. Hidden instance: `tools/game.sh start iN` + helper (private Xvfb, i1..i5; never i0, which opens on the user's desktop); instance lock `d2lod-game-iN.lock`; the host must be idle enough (preflight `D2_SOAK_MAXLOAD`, the old runs needed an override at load 24).
7. Preflight friction: soak24.sh wants a clean tree and a `d2bot --version` equal to HEAD, and WARNs when the branch is not `master` (main now); known Game.exe risks to expect: heap runaway (fixed offline, NAVMEM, attempt 17), `restarts` counter semantics, potions not drunk at death (G2 autopsy tags), "quest step walks a naked caster" death loops.

Cost estimate (I, agent-hours of a Sonnet worker, offline work unless noted):
- Option A (reuse soak24 plus a thin wrapper for no-polling, state file, record; fault injection script for Game.exe; judge accepting Game.exe lines): 5 to 7 h, plus 1 h review.
- Option B (full `--backend game` branch inside gate-run/gate-sup/gate-judge with its offline test `test_gate_run.sh` extended by a fake Game.exe): 10 to 14 h.
- Live: smoke 0.5 h of attention, Gate 1 itself 2 h wall at near-zero tokens when run through the wrapper, plus 1 h review; expect 2 to 3 attempts because Game.exe has not run the current decision code since 2026-10-30 (the whole decision layer has changed since: `git log --since=2026-10-30T20:00 -- internal/combat internal/bot internal/town internal/nav internal/resource`).

## 4. Smoke run plan (runbook only, nothing here was run)

Goal: prove, in 10 minutes of real time on one hidden instance, that the current main bot attaches to the original game, plays, saves and stops cleanly, and read the three unverified inputs (rows 10, 15, and the stat vector). It does NOT count as a gate, and no kill is injected.

Preconditions (all must hold, else wait): the Gate 1 rerun on d2d has ended (state `done`, result file written); `labrun the worker host -- top -bn2 -d1` shows idle + nice at least 25 percent; no `/tmp/d2lod-live.notice` and no running Game.exe or d2bot of the chosen instance; instance lock free; `tools/capacity.sh --add 1` replaced by the labrun top check (capacity.sh hangs); the d2bot to ship is built from the current main tip (preflight checks the stamp).

Instance and character: hidden instance i5 with Fresha (Sorceress, clvl 11 at the 2026-10-30 result, never cheated per g1-log; U that the save is unchanged), or i4 with Soaky (clvl 15) if i5's prefix is broken. First check, read-only: `d2savecheck` of the save (name, class, level, softcore, expansion) and the cheat ledger for that name.

Steps (each from the head box through the stubs; stop at the first FAIL):
1. `tools/soak/soak24.sh preflight i5 Fresha` (read-only; must end with 0 FAIL).
2. `tools/soak/soak24.sh start i5 --char Fresha --for 10m --profile m1 --min-xp-per-hour 0` (managed bot starts the game; goals campaign; bot runs 15 min, harness 10).
3. T+0 to T+3 min: `tools/soak/soak24.sh status i5`; bot status API shows the character in game, area, position.
4. T+3 to T+10: `tools/soak/soak24.sh watch --once --inst i5` every 3 minutes; read the minute lines.
5. T+10: `tools/soak/soak24.sh stop i5` (graceful: walks home, saves, stops the game, cleans the prefix).
6. Afterwards read-only: `d2savecheck` again; `tools/game.sh status i5`; `tools/jobs.sh list`; `tools/agentreap.sh --list`.

Checks and success criteria:
| # | Check | Pass |
|---|---|---|
| S1 | attach | managed game up and state in-game within 3 min; name, class, level equal the save (V run, d2savecheck) |
| S2 | stability | at least 9 minute lines, 0 panics, 0 failsafe rungs above 0, `d2soak report -profile m1 -min-xp-per-hour 0` shows no FAIL for stability |
| S3 | progress | kills greater than 0 and XP rising; no XP/h bar (10 minutes is too short), INFO only |
| S4 | kit | `packHP`, `packMP`, `beltBoxes` present on the minute lines; a belt fill or purchase line, if a town visit happened (INFO) |
| S5 | combat | `casts=` health lines present; count cast commands against mana spent for one fight (docs/fixcastgap.md section 5 recipe) as the first M data for row 8 |
| S6 | melee stats (row 10) | read the player stat vector once with the `cmd/d2snap` read-only subcommands (exact subcommand U, see `d2snap -h` and docs/mem.md) and compare 19, 21, 22, 25 with the character panel and the wielded staff or wand; closes drill R1 |
| S7 | missiles (row 15) | during any fight with a ranged monster the client-table missile list is non-empty at least once (needs a probe line; if no probe exists the result is UNKNOWN and card P2 below covers it) |
| S8 | stop | `.d2s` mtime fresh, saved XP equal the last minute-line XP, no stray Game.exe or wineserver of the prefix, `tools/jobs.sh list` clean, character standing in town, no panel open |
Abort at once (graceful stop, then report): two deaths, load guard trips, `/tmp/d2lod-live.notice` appears, any failsafe rung 2 or higher, the bot leaves the Blood Moor/town area band for a boss area.
Time: about 3 min start, 10 min run, 2 min stop. Leave nothing running; report what was done to the character.

## 5. Ordered backlog of parity work (cards, one line each, by value for validating gates)

1. `gate-run-game`: run Gate 1 on the original game in real time (soak24 wrapper with state file, scripted Game.exe fault, judge accepting Game.exe lines, fresh honest character); option A of section 3.
2. `game-stats-probe`: read-only live probe of the player stat vector (19, 21, 22, 25, 119, 325, resists, `States`) and the client missile table against the panel; closes rows 10, 11, 12, 15 (drill R1 and the missile U).
3. `game-smoke-script`: turn the section 4 runbook into one script with the S1 to S8 pass/fail output, so every later build can be smoke-checked on Game.exe in 15 minutes.
4. `game-cast-live`: live count of commands during a cast and cast length in ticks on Game.exe, plus armor state ids 20, 30, 88 switching; validates the `d2d=false` branch of `castTicks` and the buff state path (rows 8, 9).
5. `game-portal-regression`: live check that `ToTown` on Game.exe still opens and enters a portal after the d2d `enterportal` change and with an empty or absent tome (row 3).
6. `merc-wire-game`: wire the merc policy onto the memory path (row 5); only needed once Act 2 gates require a merc (ROUTEGUARD: merc or clvl 20).
7. `game-teleport-live`: one Teleport hop on Game.exe on a throwaway character that has the skill (row 2); cheat-tool kit prep allowed, never a gate character.
8. `game-monster-ai-fields`: decide whether memory can read the monster AI struct (aware, target, attack in progress) so the threat findings can be checked on the original (row 16); research first.
9. `backend-contract-both`: one scenario test that runs `d2.StateReader` consumers over both a recorded Game.exe fixture and the d2d fake, so contract drift shows offline.
10. `restarts-counter-game`: confirm live that a recycle counts as a restart on Game.exe (STATUS risk 5; spec "A game-lost recovery counts as a restart").

Overlap (declared): the deferred cards `gameexe-retire-soak` and `orig-decision-probes` list the same direction (Game.exe-only criteria, decision probes on the original); this audit feeds them and replaces neither.

## Dead ends and limits
- No live data was collected: every "PRESENT-TESTED" for M rests on docs/g1-log.md and the soak docs (S), not on a run today.
- The ops-repo tools (`gate-run.sh`, `gate-sup.sh`, `gate-judge.py`, `game.sh`, `live.sh`) are read, not run; their public stubs only forward (tools/game.sh is a 3-line stub).
- Rows 11 and 12 on M have no live comparison against the character panel in the docs reviewed (I).
