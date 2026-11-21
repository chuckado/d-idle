# The Loop order and its unlock

> status: built offline (task uib7), not verified live. Design: `docs/ui/redesign-2026-11-14/C-orders.md` section 8. Sequencer pattern: `docs/unlock-orderqueue.md`. Repair trip: `docs/repair.md`.

Tags: V verified (by test or by reading code, named), U unverified (needs the live game).

## Tiers (user answer 5, strict order)

| Tier | Player can | Opens when | Ladder row |
|---|---|---|---|
| 0 | One zone order at a time (a new one replaces), Repair now / Farm until repair by hand | always | - |
| 1 | The Loop: area for X min, home, repair, repeat | a repair the player asked for ended `done` after a zone order was given (event `manual_repair_cycle`) | `orderloop` |
| 2 | The queue (slot rows) | three Loop rounds finished in all (event `loop_rounds_3`) | `orderqueue`, then `orderqueue2`, `orderqueue3` unchanged |

Changed against before: `orderqueue` used to open at the Den of Evil (`quest_stage den_of_evil 2`). It now opens on `seen loop_rounds_3`. A hero that already has `orderqueue` open keeps it (V: `TestOrderQueueOpenedByTheDenStaysOpen` opens it with the old trigger, reopens with the shipped ladder). DECISION FOR THE COORDINATOR: keep the Den as an `alt` trigger of `orderqueue` (one data line) if heroes who clear the Den should still get slot one; strict order says no.

## Spec

Fields: `area` (an orderable area), `minutes` 5 to 240 (default 20), `repair` `damaged` (default) or `always`, `rounds` 1 to 20 (default 5) or `forever`, `approve` (charter of an above-range area).

One round: place a `go` order through `Hub.place` (charter, 5-open limit and the bot's route guard unchanged). When it ends `done`: `damaged` asks the repair desk only when something worn is under the warn line or broken, else places `home`; `always` asks every time. A desk answer `undamaged` or `cooldown` counts as done and places `home`. A request that is open is awaited; `done` finishes the round, then the next round starts. State `<state dir>/orderloop.json`; a restart resumes at the start of the round (V test).

Stops (state, reason on the row): rounds done (`done`); player Stop, or a zone order with replace (`cancelled`); death, the guard refuses the area or the farm leg fails, repair failed or dropped by the bot, a Pact forbids repairs, the `Halt` hook (Pact or low-life pause) (`failed`). All V by `TestLoopStopConditions` for Sorceress and Paladin.

Never loops forever on failure: a busy hero (another order open, orders box full) is retried after 30, 60, 120 seconds, then the loop stops (`gave up`); three farm legs in a row that end `done` in under 60 seconds stop it (`ended at once`); a refused area stops it at once. Journey blocked (`goals.Status.JourneyBlocked`) is shown on the view as `journeyBlocked` and does not stop the loop: the loop is a player order and goes before the plan.

Priority: the loop is player-order level, behind safety. Between its steps `Fun.Settle` does not tick the queue (`Loop.Holding`). A finite loop therefore holds the place and queue entries wait. `forever` is refused with `queue-busy` while entries wait (V). The other direction (adding to the queue during a `forever` loop) is NOT refused: the queue just never runs while the loop holds. Making it refuse needs a line in `unlockorderqueue.go`, which this task may not touch (U, open item).

## API (additive)

`GET /api/orderloop` (viewer, `schema: watch.orderloop/1`): `open`, `title`, `teaser` (locked), `state` (`locked|idle|active|done|cancelled|failed`), `phase` (`farming|going home|repairing|waiting`), `round`, `rounds`, `forever`, `area`, `minutes`, `repair`, `text`, `finished`, `stoppedWhy`, `waitingFor`, `journeyBlocked`. Goldens: `internal/watch/testdata/orderloop/view-{locked,idle,active,cancelled}-{sorceress,paladin}.json`.

`POST /api/orderloop {action: "start" | "stop", area, minutes, repair, rounds | forever, approve}` on the answer listener (origin list, token, JSON, small body, unknown fields refused). Codes: 409 `locked`, 409 `active`, 409 `none`, 409 `queue-busy`, 402 `charter-needed`, 400 `bad-order`, 503 `off` (repair desk not bound).

For `/api/commands` (S4): `OrderLoop.Command()` returns one row `{id: "loop:1", text: "Loop: Cold Plains 20 min, repair, repeat (round 3 of 5)", state: "active", phase}` while a loop is active.

## Where the code is

`internal/watch/orderloop.go` (controller, routes), `cmd/d2bot/loopwire.go` (config and the repair desk adapter on `repair.Controller`), ladder rows in `data/unlocks/ladder.json`. Small additive hooks in `internal/watch/funwire.go` (field, routes, Tick, death) and `cmd/d2bot/funwire.go`, `repairwire.go` (one line each). No `internal/bot` change. Build flexibility: no class, skill or build literal; the area list is the orders list; tests run Sorceress and Paladin.

## U (needs the live game)

1. The `home` order ends `done` by itself (it is a one-minute idle goal, `OrderToGoal`; V by reading, U live).
2. The repair trip, started from the loop right after a farm leg, finds its safe moment and reaches `done` with the hero back in the field; a farm leg for gold works for a loop request.
3. After `done` of a repair request the next `go` order starts cleanly (the hero walks out again).
4. The `manual_repair_cycle` event: detected from the repair desk's request state while the row is shut (a stale `done` also counts once a zone order exists; the zone-order flag is persisted).

## Live checklist (hidden instance, throwaway character)

1. Fresh hero: `GET /api/orderloop` says locked with the teaser; `POST start` answers 409 `locked`.
2. Give a zone order, damage gear, press Repair now, let it finish: the row opens (reveal in the journal).
3. Start a Loop of 2 rounds of 5 minutes at a near area, `repair: always`: go, town, smith, out again, second round, `done`.
4. Repeat with nothing damaged and `repair: damaged`: `home` order, no smith.
5. Start a loop and give a zone order with Go now: loop `cancelled`. Start one and Stop: order cancelled.
6. Three finished rounds in all: the queue row opens.
7. Kill the bot during a round, restart: the loop resumes at the start of the round.

## Gates

Run from inside the worktree after the merge of origin/main, when the worker host is free:
`labrun <worker> -- sh -c 'go vet ./... && go test -short -count=1 -p 4 ./... && go test -race -count=1 ./internal/watch/...'`, then `python3 -I tools/contracts/check.py`, `go run ./cmd/leakcheck -mode fail`, `tools/ci/check.sh`, then `~/code/d2lod-ops/tools/mergeq.sh add agent/uib7 --card TASK-425`. Done on the small box: per-package `go vet` and `go test -p 1` of internal/watch, internal/unlock, cmd/d2bot (selected tests).
