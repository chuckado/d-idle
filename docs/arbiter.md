# The intent arbiter (shadow mode)

> status: current. First slice of roadmap R4 (docs/design/fable-review-2026-11-19/06-roadmap.md, 05-architecture-and-overhaul-options.md section 3.1). SHADOW ONLY: nothing the hero does changes. Tags: V = a test or measurement in this branch, U = unverified.

## Why

`runGoal` (internal/bot/goal.go) is a fixed list of about twenty `if step() { return true }` hooks; each decides alone, and five guards (flail, ping-pong, explore gave up, border loop, arrival bounce) detect afterwards that two hooks undid each other. The arbiter is the place that knows what the hero is committed to and what a change of plan costs. It is built and tested here, run beside the chain, and moved over one hook family at a time (migration plan below).

## Model (internal/arbiter)

- A `Source` proposes zero or one `Intent` per tick: priority class (reflex > player > upkeep > objective > filler), urgency (compared inside a class), where it must happen (`town`, an area name, or `*`), estimated cost, reason, resumable, steps.
- `Arbiter.Decide(View)` picks one. Rules: strict classes; ties go to the running intent, then lower cost, then name. The chosen intent holds a lease of `bot.arbiterDwellSec` virtual seconds; only a higher class preempts it. An intent that fails (`Report(..., false)`) backs off 30 s, 60 s, 120 s ... up to 30 min. An area change that would undo the previous one inside `bot.arbiterTravelSec`, with no progress in between (kill, XP, gold, belt), is refused and logged with both reasons (the intent's own and the budget). Reflexes and player orders are never refused. When an upkeep intent for town wins, every source that implements `TownVisitor` is asked what it wants in the same visit (`Decision.Batch`).
- Sources read only `View`: world facts and the arbiter's previous decision (`View.Prev`, its own trace). They never read each other's state.
- Pure and deterministic: no goroutines, no clock, no I/O. Time and progress come in through `View`.
- `Shadow` (internal/arbiter/shadow.go) wraps it for a live loop: recovers any panic (counted), one short mutex, no I/O, compares the arbiter's hook family with the family the chain acted for.
- `StdSources` (sources.go) are the first nine sources (survive, player, trip, claim, resupply, recovery, goldfarm, quest, level). Build data (belt line, flee line, gold need) arrives as facts, so no class, skill or build literal lives in the package.

## Shadow knobs (docs/knobs.md)

| knob | default | meaning |
|---|---|---|
| `bot.arbiterShadow` | 0 | 1 runs the arbiter beside the chain. 0: nothing runs, nothing is logged |
| `bot.arbiterDwellSec` | 120 | lease length |
| `bot.arbiterTravelSec` | 600 | travel budget window |
| `bot.arbiterLogGapSec` | 60 | minimum gap between two divergence lines of one (chain, arbiter) pair |

To enable on the sandbox: set `bot.arbiterShadow` to 1 in the sandbox's tuning (the same place any `bot.*` knob is set), restart not needed beyond the usual knob reload. To disable: set it back to 0.

## The adapter (internal/bot/shadow.go)

`runGoalShadowed` replaces the one call of `runGoal` in `loop.go`. With the knob 0 it IS `runGoal`. With it on it collects world facts from the snapshot and loop state BEFORE the chain acts (in town, life, hostiles near, a trip running, claim pending, belt potions, gold, recovery, gold farm order, camp goal, player item actions pending), runs `runGoal` unchanged, then asks the shadow. The chain's family is read from the whyline code of the activity the chain started (`chainFamily`). The hooks are not wrapped one by one yet: their predicates have side effects (`needChores` clears `forceTrip`), so the sources use pure facts instead (U: how closely they match each hook's own predicate is exactly what the shadow log measures).

Read-only data: `Runner.Arbiter()` returns the last `Decision` (for the watch API; not wired into a page in this slice).

## How to read the shadow log

- Divergence line (rate limited): `arbiter shadow diverge="chain chose quest, arbiter would choose resupply/refill (belt 1 under 4) because quest/step: class objective below upkeep; ..."`. Format: chain family, then the arbiter's intent as `source/kind (reason)` with ` [lease]` when held, then the first two reasons it rejected something.
- Minute line attribute `arb="chores=68/103 quest=258/310 panics=0"`: per chain hook family `diverged/ticks` since the previous minute line, then the panic count. Families: chores, claim, quest, goldfarm, recovery, player, survive, level. Ticks where the chain family is unknown (no whyline code) are not counted.
- A high divergence rate for a family means either the arbiter source is wrong (fix the source) or the chain's hook is (the reason is in the line). Never both unread.

## Replay (offline)

`go run ./cmd/d2arbiter [-follow] [-trace] LOG.jsonl` feeds a JSON-lines log (`minute` lines, or `tick` lines: area, kills, xp, chain, objectiveArea, facts) to the arbiter. Without `-follow` the arbiter sees the logged hero (divergences and counters). With `-follow` the simulated hero obeys the arbiter, facts and progress still come from the log, and an intent that ran its estimated cost in place with its reason still holding is reported failed.

Fixture `internal/arbiter/testdata/darkwood_pingpong.jsonl`: the Dark Wood, camp, Dark Wood, Cold Plains cycle of 2026-11-18 over 12 minutes. V: the area sequence and gaps are the real ones copied from the sandbox log into `internal/bot/flail_test.go`. U: facts (belt 1 under 4, 350 pack gold, a quest step in Dark Wood, no progress) are assumed from docs/probes/treefix3-2026-11-19.md, not logged. Result (V, `TestReplayPingPong`): the chain made 412 area changes; the arbiter, obeying itself, makes 2 (one to town, one back after the back-off), refusing 359 undo trips, each with both reasons in the trace (`testdata/pingpong_replay.golden`). Limit: the log never says whether the town visit helped; a chore that stays wanted is treated as failed after its cost.

## Measured (V, this branch)

- `go test -race -count=20 ./internal/arbiter` passes; the prototype scenarios are all here: first-match chain 180 area changes in 60 virtual minutes vs arbiter 9, broke with stash gold, preemption and resume, starvation and back-off, golden decision trace.
- `BenchmarkShadowTick`: about 2.6 microseconds and 19 allocations per tick with the standard sources, against a goal tick of about a second.
- Shadow on, off: `TestShadowOffByDefault`, `TestShadowOnLogsAndKeepsTheChain`; panics: `TestShadowSurvivesPanicAndKeepsCounting`.

## Not done (U)

- Live: no sandbox run yet. Whether the divergence rates are sane on a real log is the first thing to read.
- `ObjectiveArea` is empty in the adapter (where the quest step must be done is not available as a pure fact), so the travel budget bites only for the trip and replay cases live. Wiring it is the first adapter follow-up.
- The chain's family is the whyline code and can lag one activity (U).
- No executor: the arbiter chooses, it does not run or abort steps; `EstCost` is a tie-break and trace field.

## Migration plan per family

Order is by risk. Each family moves from the chain to the arbiter behind its own knob only when its exit rule is met; until then the chain's hook stays. The flail guard, ping-pong and border-loop breakers stay until the shadow log shows they never fire for 14 days after the last family moved. Under the project freeze only shadow and scenario-backed work merges.

| order | family | exit rule (all of) | earliest |
|---|---|---|---|
| 1 | survive, player | divergence under 2 % of ticks over 3 sandbox days, every remaining line explained | 2026-11-26 |
| 2 | chores, claim | under 5 % over 3 days; the replay of every flail log line since 2026-11-18 shows no new area change the chain did not make | 2026-12-03 |
| 3 | goldfarm, recovery | under 5 % over 3 days; the stash-gold case has a shadow line or a scenario | 2026-12-10 |
| 4 | quest (board and task) | `ObjectiveArea` wired; under 5 % over 7 days | 2026-12-17 |
| 5 | level (filler) | last: everything else moved and a 7-day sandbox run with the arbiter acting shows no flail guard hit | 2026-12-24 |

If a date passes without the rule met, the family stays on the chain and the reason goes into this table; the dates are earliest, not promises.
