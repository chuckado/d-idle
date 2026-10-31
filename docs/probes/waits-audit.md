# Waits audit: wall-clock waits for the game (card waits-audit)

> status: current. Offline audit of every wait in the bot, the d2d wiring and the d2d backends; fixes in the same card. Live proof is a drill rerun after merge (bottom).

Lockstep: the game moves only when the Pacer steps it (`d2dinput.Pacer`, `LateClock`). OK = steps the game, or wall by design. BROKEN = waits for the game on the wall clock. RISK = too short in real time. Scope: d2d only unless stated.

| Where | Waits for | Duration | Lockstep | Real time | Class |
|---|---|---|---|---|---|
| internal/bot, nav, town, combat, boss, loot, leveler, quest, d2dcombat (`e.Clock.Sleep`, `n.clk.Sleep`, boss `World.Sleep`, `Meter`), cmd/d2bot (`d2dwire`, `objects`, `task`, `cain`, `adapters`, `boss` Sleep) | game state, settle times, polls | 0.1 s to 8 s | `LateClock.Sleep` steps the Pacer | wall sleep | OK (about 150 call sites, none raw) |
| internal/bot/loop.go:52,859,907 main loop idle (`pause`), guard poll, reaction | next tick | TickMs, guard period | stepped (guard runs through `Watch` on the acting goroutine) | wall | OK (pacing is deliberate) |
| internal/bot/loop.go:137,279,297,336,473 step and restart timeouts | hang detector | 15 s to 5 min wall | wall ceilings; a slow lockstep box makes them tighter in game time | hang guard | OK (not a game wait; see note 1) |
| internal/d2dctl/town_verbs.go `Pause` (HoverClick, shop, repair, waypoint, EnterPortal, sidestep), d2dtown/claim.go:67, d2dmerc/act.go:72 | menu or panel after a click | 0.15 s to 1.5 s | FIXOBJ `LateClock.Pause` steps (installed by `Backend.Start`) | wall sleep | OK |
| internal/d2dctl/town_verbs.go:245,262 `ApproachNPC` deadline | NPC menu | 15 s (d2dtown), 6 s to 15 s | was `time.Now`: 15 s WALL while Pause steps about 11x real time, so about 160 game seconds of retries | wall | BROKEN, fixed (`d2dctl.Now`) |
| cmd/d2dshop/main.go:156-276 `d2dctl.Pause` | game after click | 0.6 s to 3 s | no `Install`: raw wall sleep, game frozen | wall | BROKEN, fixed (`LateClock.Install`) |
| internal/d2dtown/adapter.go:288 `hoverFirst.wait` fallback | hover settle | 0.12 s, 0.35 s | uses `Controller.Wait` (Clock); raw sleep only when the wrapped controller has no `Wait` (d2dinput's has) | wall | OK |
| internal/d2dinput/controller.go `Wait`, `MoveTo`, `TickDur` | walk arrival | per tick | `Clock` (Pacer) | wall | OK |
| internal/d2dworld/world.go:233-251 `waitLinked`, `LateClock.Poll` | neighbour link by a d2d worker thread that runs in WALL time | 90 s wall, 80 ms step plus 10 ms wall per poll | steps 2 ticks and sleeps 10 ms wall per poll | wall | OK (wall deadline by design, docs/d2d-economy.md) |
| internal/d2dctl/script.go:58 `sleep` step | CLI script | as written | the script says `step N` itself | wall | OK (tool) |
| internal/d2dctl/client.go:177,237, d2dsrc/source.go:90, d2denv reader/quest/beat, d2dtown/adapter.go:56,271-327, d2dmerc:43, d2dloot:46, internal/town/town.go:434, chores.go:470 context timeouts | socket round trip of one verb | 3 s to 30 s wall | hang ceilings; a step inside a verb is fast | same | OK (note 1) |
| internal/d2dinput/pacer.go:113 | dead channel: also waits the real time | lost ms | by design | n/a | OK |
| internal/d2dinput/lifecycle.go:161-261, d2denv/proc.go:183-345, cmd/d2bot/d2dbackend.go:111,232 | d2d process start, socket, in-game screen, quit | 0.25 s to 10 s | before a game exists; soaks pass through it (V) | wall | OK |
| internal/d2denv/env.go (`traceLoop`, `cpuProfile`, `wallStacks`), cmd/d2bot `logPace`, `stopAtVirtual` (1 s wall ticker on virtual `Now`) | profiling, logs, soak length | wall | n/a | n/a | OK |
| cmd/d2bot supervise, heapguard, watchdog, safego, live.go, serve, internal/api, fleet, d2dmesh (leases), d2dreplay (tap, lock), internal/launch (process and OOG flows), internal/input (Game.exe, real time only), internal/party | supervision, rate limits, network, Game.exe | wall | not d2d game waits | wall | OK |
| internal/d2denv/seams.go:47 | log throttle | 1 min | n/a | n/a | OK |
| cmd/d2drill, internal/drill | drill harness | | owned by m1-liveproof, keeps its own `Pause` override | | not audited |
| cmd/*live, cmd/inputcal, townlive, navlive, combatlive | Game.exe probe tools, real time only | | not a d2d path | | out of scope |

Counts: OK 16 rows, BROKEN 2 (both fixed), RISK 0, not fixed 0.

Notes:
1. Wall ceilings (round trips, step timeouts) hold at lockstep speed at or above real time. A box where lockstep runs far slower than real time (`logPace`: ratio below 1) makes them fire early in game time. Not changed: they are hang guards, and virtual ceilings would need a goroutine on the Pacer (breaks determinism, docs/d2d-determinism.md). RISK only on a heavily loaded box; watch the pace line.
2. `d2dctl.Pause` and `Now` are process globals: one d2d per process. `Install` says so; no process builds two backends today.
3. Not audited as game waits: Game.exe paths (shared code, real time by definition).

Fix: `d2dctl.Now` (default `time.Now`), read by `ApproachNPC`; `d2denv.NewLateClock` and `LateClock.Install` set `Pause` and `Now`; `Backend.Start` and cmd/d2dshop call it. Tests: `TestApproachNPCDeadlineIsGameTime`, `TestNowDefaultsToWall` (d2dctl), `TestInstallLockstepAndWall` (d2denv).

Drills to rerun after merge: H.pause (Kashya approach with the bot's pause, now with the 15 s game-time deadline), D8 (merc claim and revive), D3.n (NPC menus), and one `d2dshop` run against a lockstep d2d.
