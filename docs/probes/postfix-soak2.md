# postfix-soak2: Gate 1 of the d2d gate ladder on current main (2026-10-31)

> status: current. One finished hidden d2d soak (exit 0, 120 wall min, one injected d2d kill), n=1 seed, not a statistical claim. Previous soak: [postfix-soak1.md](postfix-soak1.md). Criteria: soak-gates spec (Gate 1, as modified by ADR 0013), docs/review-strategy.md "d2d gate ladder". Tags: V = counted from the log or read live.

## Verdict
Gate 1 is NOT PASSED on this build (one seed). Of the soak-gates Gate 1 criteria: stability, route guard, portal and fault recovery pass; deaths pass the death rule (1 death, early); XP per hour is below the d2d baseline (indicative fail, INCONCLUSIVE at one seed); the belt window and the 10 minute no-progress rule fail when read in virtual minutes (the reading postfix-soak1 used) and pass in wall minutes. `d2dsoak compare` says `INVARIANTS: PASS`, OVERALL INCONCLUSIVE (exit 2, one seed).

## Decision 2026-11-01
User decision (reply "virtual minutes"): in the d2d gate ladder every minute of the soak-gates criteria is a virtual (game) minute, because the d2d clock runs about 10 times faster than the wall clock and the rules protect in-game behaviour (potion supply, progress). Card `soak-minutes-virtual` (spec delta, numbers unchanged: belt window 5, no kill and no XP 10).
- Gate 1 verdict: NOT PASSED. Failing criteria: belt window (27 virtual minutes at the start, limit 5) and no kill and no XP (26 virtual minutes, limit 10). XP per hour (5437 vs 9172) stays INCONCLUSIVE at one seed.
- A soak log "line" is one `minute` line = one virtual minute (1275 lines for 21.25 virtual h). The two no-progress numbers above differ by definition, not by clock: 26 is the longest run of lines with no new kill and no new XP (the Gate 1 criterion, counter NOKILL); 16 is the longest stretch without any progress event of d2dsoak's wider PROGRESS invariant (also kills, XP, non-town area entered, quest done, gold spent, clvl; limit 30). The 30 was never the Gate 1 number (it matches the stall ladder), so INVARIANTS PASS did not contradict the spec.
- d2dsoak now computes both Gate 1 stretches in virtual minutes (codes BELT and NOKILL) and prints them with the limits; on this log it reports belt 27 and no kill/XP 26 and VIOLATED(BELT,NOKILL) (test `TestInvariantsGate1PostfixSoak2`). The sections below keep the original wording ("FAIL in virtual minutes, PASS in wall minutes"): the wall reading no longer counts.

## Setup (V)
- Bot: origin/main 0b0bdb2 (all fixes merged: fixrun, fixmisc, fixstall3/4, fixmeter, fixbank, fixobj, fixprice, fixmissile and the earlier fix-belt/pickup/loot/portal/wp), built on the worker host, sha256 prefix 387cae79cb81. `d2dsoak` built from the same tree.
- Lab d2d: `d2d.lab-cheats3` (fork build with the click modifier and the test cheats; same binary as the drills), sha256 prefix 5e2e6fe8b204. No debug verb, no cheat of any kind was used in the soak. Soak 1 ran `d2d.lab-port148` (click modifier, no test cheats), so the d2d side differs slightly; treat the comparison as indicative.
- Recipe of postfix-soak1 and navsoak3 (docs/d2d-baselines-early.md section 4): fresh uncheated Sorceress (class 4, character Post2, seed 5), `--backend d2d --managed --goals campaign --toggle autoloot=off,autoparty=off --d2d-args "--lockstep" --d2d-linkwait 60s --run-for <left>s`, `D2D_RUN=on`, nice 10, restart supervisor with up to 3 restarts (never needed), hidden slot 1 (slot 0 is the drill slot), no viewer, no stream.
- Start: `D2D_AGENT=soakg1d2d tools/agentrun.sh soakg1d2d --ttl 8700 -- tools/d2dslot.sh run -n 1 -w 600 -t 8700 -- bash postfix2-sup.sh`. Health guard before the start: `capacity.sh --add 1.2` on the worker host (90 s window) idle+nice 73.8%, iowait 5.2%, predicted 66.3%, verdict OK; no other d2d, slots 0..6 free, the DETHASH agent was gone.
- Times (worker host clock): start 21:01:54, fault 22:01:56 (60 min), supervisor exit 23:01:59 (limit 7200 s). 1275 minute lines = 21.25 virtual h (ratio about 10.6 virtual to wall).

## Timeline
| time | event |
|---|---|
| 21:01:54 | supervisor and bot start, fresh character |
| 21:09:45 | the only death: clvl 4, Burial Grounds, killed by Hungry Dead, belt 0 healing, 1 town portal scroll, step `Quest: fighting (Level up to 5)` |
| 22:01:56 | injected fault: `kill -9` of the slot's d2d (pid taken as the child of the registered bot in the registered process group of the slot, pgid checked before the kill; nothing else touched) |
| 22:01:56.86 | bot logs `d2d process exited ... signal: killed`, `xp at risk`, `game lost, restarting it` |
| 22:01:58 | `d2d: leave: re-entered the game`: back in game 2 s after the kill; kills and XP resumed within seconds |
| 23:01:59 | `--run-for` reached, `d2d: stopped by the bot (kill)`, exit 0, supervisor `time limit reached` |

## Fault recovery (V)
Back in game in 2 s (limit 3 min). Minute-line `restarts` 9 to 10 (+1; the counter also counts routine d2d recycles, 9 before the kill; the log has exactly one `d2d process exited` and one `game lost, restarting it`). Saved XP: the save held 80503 (`xp lost` 1330 of 81833 in memory); after the restart the character had 80503, not lower. `agentreap --list` held only the run's two entries before and nothing of soakg1d2d after the end; no process of the slot left.

## d2dsoak output (scrubbed)
```
metric                     d2d mean  n | baseline (min .. max)   |  gap    verdict
xp_per_hour                    5437   1 | 9172 (6600 .. 11743)    | -40.7%  INCONCLUSIVE [too few seeds]
kills_per_hour                  314   1 | 398 (389 .. 408)        | -21.3%  INCONCLUSIVE [too few seeds]
deaths_per_hour              0.0471   1 | 0.405 (0.29 .. 0.52)    | -88.4%  INCONCLUSIVE [too few seeds]
deaths_per_1000_kills          0.15   1 | 1.02 (0.7 .. 1.34)      | -85.3%  INCONCLUSIVE [too few seeds]
invariants: vmin 1274, hang 0, stall 0, wedge 0m/57, noprog 16m, deaths/1000k 0.2 (1/6664k),
            fresh22/chk59/bot0, exit limit, unrecovered 0: PASS
COVERAGE: class unknown (tuning group), act 1, engine d2d, clvl 1-13, 21.2 virtual h
INVARIANTS: PASS     OVERALL: INCONCLUSIVE (exit 2)
```
(`d2dsoak compare -hours 21.25 -baselines tools/d2dsoak/baselines-early.json bot-0.log`; `compare` has no `report` subcommand, its invariants block is the report.)

## Gate 1 criteria
| criterion | result | evidence |
|---|---|---|
| Stability: only the injected exit, 0 panics, alive at 2 h | PASS | exit 0, 0 panics, 0 bot restarts, 1 `d2d process exited` (the fault), no watchdog line. `saveXp == charXp` after a graceful stop: UNVERIFIED (the end is a d2d kill by the bot; `xpRisk` 2313 at the last minute line) |
| Deaths: death rule | PASS | 1 death in 6664 kills (0.15 per 1000 kills), invariants PASS, no death-rule code. One death at clvl 4 with an empty belt (finding F1) |
| True XP/h at or above the d2d baseline | FAIL (indicative) | 5437 vs 9172 (-40.7%), INCONCLUSIVE because one seed; clvl 13 vs the baseline's level is not like for like |
| Belt: no window over 5 min with no healing outside town | FAIL in virtual minutes, PASS in wall minutes | longest window 27 virtual minute lines (87 s wall, 21:09:48 to 21:11:16, the start and the death); 2 of 9 windows over 5 virtual min; 46 of 1275 minutes (3.6%) without healing outside town |
| Portal: at least one trip by town portal | PASS | 98 `arrived in town through an open portal`, 53 waypoint-to-town fallbacks |
| Route guard | PASS | 0 `blocked: not ready`, no boss run |
| Injected fault recovery | PASS | 2 s, restarts +1, saved XP not lower, no orphan |
| No single state over 10 min without progress (Gate 1 deaths and progression) | FAIL in virtual minutes, PASS in wall minutes | d2dsoak no-progress max 16 virtual min (progress includes area changes; the invariant limit is 30, so INVARIANTS stays PASS); by kills/XP alone 26 virtual lines (21:36:30 to 21:39:01, 2.5 wall min, `recovery: farming gold` then `Quest: fighting` in area 4) |

Because the spec says Gate 1 passes only if every criterion passes, the verdict is not PASS. The unit question is settled by the decision above: virtual minutes.

## Comparison with postfix-soak1
| | soak 1 (f8c1b7f, port148) | soak 2 (0b0bdb2, cheats3) |
|---|---|---|
| length | 119 wall min, 20.9 virtual h | 120 wall min, 21.25 virtual h |
| clvl / XP / kills / deaths | 13 / 119024 / 6331 / 0 | 13 / 115637 / 6668 / 1 |
| XP per vh (d2dsoak) | 5719 (raw 5705) | 5437 (-4.7%) |
| kills per vh | 304 | 314 |
| restarts counter (fault +1) | 19 (8 to 9) | 22 (9 to 10) |
| fresh games / chicken lines | 19 / 64 | 22 / 65 |
| minutes with belt potions 0 | 2.2% of 1253 | 4.2% of 1275 (54) |
| `no healing potion in the belt` | 92 (4.4/vh) | 463 (21.8/vh) |
| `belt: moved 0 potions` of `belt: moved` | 47 of 190 | 0 of 166 |
| `potion already in the pack` skips | 650 | 645 |
| `recovery: loot` lines | 249 | 347 |
| `no portal opened` / `portal did not lead` | 11 / 8 | 7 / 7 |
| town arrivals through a portal / waypoint fallbacks | 88 / 33 | 98 / 53 |
| nav stuck | 1797 | 1820 |
| combat `move failed` | 1375 | 1210 (1185 are `nav: no path: empty path`) |
| re-plans | 1113 | 1198 |
| `nav stalled by monsters` / `failsafe ladder` | 635 / 17 | 674 / 23 |
| walking share of kill time (`killtime` lines) | 68% | 70.3% (53865 of 76607 s; cast 7.1%, wait 22.5%; 6427 kills) |

## Run state (new fields, V from the minute lines)
- `runBelieved` true in 84.7% of minutes (the run share, a belief: the toggle is not readable on d2d). `runKeys` 5171 presses, `runResets` 59 (one per game start, chicken, recycle). Stamina below its maximum in 88.2% of minutes, which is consistent with the character running most of the time. `walkingRaw` is "1" in 84.3% and "0" in 15.7% of minutes: its meaning is still UNVERIFIED (docs/fixrun.md), the value tracks `runBelieved` closely enough that "1" probably means moving in run mode, but nothing here proves it.
- So the walking share of kill time (70%) is not a walk-versus-run problem: the bot runs and still spends 70% of each kill walking.

## Where the walking time goes (V counts, S reading)
Nav counts per virtual hour are unchanged from soak 1 (stuck 86/vh vs 86, replans 56 vs 53, stalled by monsters 32 vs 30). The new breakdown: `nav stuck` rungs 0..5 = 1374 / 362 / 67 / 7 / 6 / 4 (most self-clear at rung 0); 98% of combat `move failed` (1185 of 1210) say `nav: no path: empty path`, i.e. the fight asked for a move to a target cell the nav grid has no path to (monster behind a wall or in an unlinked pocket) and the bot gives up the approach, which costs a re-plan each time. The long tail is the walk between kills (areas of 100 to 170 subtiles) rather than one wedged state.

## Findings (fix-queue candidates)
- F1 Early death by pushing on without potions: the only death was clvl 4, area Burial Grounds (mlvl 2), belt 0 healing, 1 portal scroll, quest step `Level up to 5`; `no healing potion in the belt` fired 463 times in the run (5x soak 1). The early game runs out of healing for the first 27 virtual minutes (the belt window FAIL, 21:09:48 to 21:11:16, ending with the death); a policy gate "do not enter a field level with 0 healing and gold for potions" is the candidate (early-game-policy spec).
- F2 `recovery: farming gold` poverty loops at clvl 12 (22:30 to 22:45): the bot farmed gold in Blood Moor and Cold Plains for about 20 virtual min to afford mana potions (`mp belt columns below 5`, `Going to town (low on mana potions)`), the stretch 22:26 to 22:41 (15 wall min) gained 2685 XP (100608 to 103293) against 8843 in the next 15 min (to 112136). The XP/h gap to the baseline is mostly this poverty plus walking. Candidate: farm gold where the XP is (the current level's own area) and cap mana-potion demand.
- F3 No-progress stretch of 26 virtual lines (21:36 to 21:39) in area 4 with `Recovery: farming gold` then `Quest: fighting (Level up to 17)`: no kill for 2.5 wall min; look at the recovery to quest hand-over.
- F4 Combat `no path: empty path` (1185): the approach target is not reachable; candidate: choose another target or skip the monster; evidence in the log excerpt.
- F5 Waypoint-to-town fallbacks rose 33 to 53, `no town portal and no waypoint: walking back` 47, `d2d town portal: no portal source (empty)` 180: town portal scrolls run out under poverty; same root as F1 and F2.
- F6 `walkingRaw` meaning still unproven; a one-off drill with the run toggle in both states would settle it.
- Improved against soak 1 (V): `belt: moved 0 potions` 47 of 190 to 0 of 166 (the belt fill now always moves a potion), `no portal opened` and `portal did not lead` stay at about 7 each, `recovery: loot` recovers more often.

## Second seed
Run the same recipe with 2 more seeds (and the same d2d binary) in parallel on slots 1 and 2 to turn the XP/h number into a verdict (the compare gate needs 5 seeds for a PASS/FAIL, `docs/d2d-baselines-early.md` section 3 for the sample size). Compare per seed: the clvl at which healing runs out (F1), the share of minutes in `recovery`, the poverty loops (F2) and the first 30 virtual minutes (belt window), since those decide the three FAILs. Add a save-and-exit at the end to settle `saveXp == charXp`.

## Files
`docs/probes/soak-logs/postfix2-minutes.jsonl.gz`: the supervisor log, every `minute` line, the death lines and the fault lines of the run (the full 12 MB bot log stays on the worker host in the run directory). Build on the worker host: `go build ./cmd/d2bot ./cmd/d2dsoak` at 0b0bdb2.
