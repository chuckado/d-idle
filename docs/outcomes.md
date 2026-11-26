# Typed step outcome (knob `bot.typedOutcome`)

> status: current, stage 1 (card outcome1, 2026-11-25). Source: P2 of `docs/research/fable-logic-review-2026-11-25.md` (sections 1.2, 3.1, 7.1). One path is converted; the rest still return plain errors.

## The type (`internal/bot/outcome.go`)
| kind | meaning | what the loop does |
|---|---|---|
| `done` | the step finished | nothing |
| `progress` | the step ran, more is left | nothing |
| `refused{need, wake}` | the step may not run now; `need` is a closed code, `wake` the fact that ends it | `handleRefused`: no failure count, one INFO `step refused need= wake= why=` per change of need or facts, a pause that is short on the first look after a change and `needTable[need].PauseSec` while nothing it waits on changed; the planner re-plans around it on the next tick |
| `failed{fault}` | a real fault (game, input, navigation) | the legacy `fail` path (count, ladder) |
| `interrupted{cause}` | cancelled from outside (guard, player, new goal, death) | as before, never a failure |

`Refused` travels as an `error` (`*Refused`) through the legacy `func(ctx) error` actions; `do` unwraps it in ONE place (`errors.As`). `classify(task.Result)` builds the outcome from an executor result.

## Need table (data: `needTable`)
| need | source | wake fact | pause (unchanged facts) |
|---|---|---|---|
| `needs_potions` | board code, or words with potion or kit | potions or portal scroll in the pack | 5 s |
| `needs_gold_farm` | board code, or words with gold | gold in purse or stash rises | 15 s |
| `too_weak` | board code, or words with clvl, level, too weak | character level rises | 15 s |
| `route_closed` | board code, or early road, route guard, tower guard | kit, level or gold change | 10 s |
| `area_banned` | board code, or banned | the ban ends or the level rises | 15 s |
| `not_ready` | fallback | any of level, kit, gold, area changes | 5 s |

The wake fingerprint (`wakeFacts`) is level, healing potions (belt and pack), mana potions, portal scrolls, gold in 50 steps, area. A change logs a new line and resets the pause.

## Converted path (stage 1)
The quest board step (`boardSlice`, `questboard.go`): a result for which `refusedResult` holds (guard says danger or "not ready", not the executor's stuck pause, not a not-found) returns `*Refused`. This is the path of the 2026-11-25 sandbox log (296 `action failed ... not ready: early road` warnings). With `bot.typedOutcome` 0 the legacy error goes through `fail` as before. The invariant monitor rows `refusal-as-failure` and `refusal-storm` (docs/invariants.md) measure the difference.

## Not yet converted (follow-up cards, `convert <path> to typed outcomes`)
1. `taskStep` (`task.go`, the non-board quest runner) and `planningFault` / `planningBlock`.
2. The travel gate: `errAreaGated` in `do` and `gatedWait` (`geargate.go`, `broke.go` Blood Raven errors).
3. `parkStep` and the stale-target holds (`staletarget.go`).
4. `guardedExit` and `routeGuardHop` (`routeguard.go`).
5. Kit gate holds (`kitgate.go`) and the town chores refusals (`town`).
6. The wording match `guardRefusal` in `internal/watch/hub.go` (reads the goal failure text; becomes the typed need).
7. The `*Until` backoff fields that a wake fact can replace (after the above).

Tags: V unit tests `go test -run 'Classify|Refused|LegacyRefusal' ./internal/bot`; U live behaviour (no game run): the pause values and the claim that the next tick re-plans around a board refusal are read from the code, not measured.
