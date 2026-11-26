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
1. DONE (card outcome-task-step): `taskStep` (the non-board quest runner) through `planningBlock`: a quest that is not ready returns the typed refusal (`classify`), and the ban wait returns `area_banned` with a fixed why (the minutes left stay out of the key). The two-minute wait in town for a `not_found` plan (no eligible area) stays a deliberate wait that returns nil; it is not a refusal. Knob 0: nil as before. Tests: `go test -run PlanningBlock ./internal/bot` (two classes).
2. DONE (card outcome-travel-gate): the travel gate. Every `errAreaGated` site (`geargate.go` routeAllowed, `loop.go` gotoDeathArea, `broke.go` levelSlice and the Blood Raven budget) goes through `r.refuse`, which wraps the legacy error in a `*Refused` (need from `needOf`, the sentinel still reachable through `Unwrap`). `do` runs `gateEffects` (ban wait, standing line, skipped target, state) and then `handleRefused` (one line per change, pause by need) instead of `gatedWait`; knob 0 keeps `gatedWait`. Cancel storm of 2026-11-26 (U cause, read from the code: the guard cancelled the no-scroll walk home with a portal or crowd verdict nothing could carry out, so `ToTown` restarted every ~10 s): with no portal source the guard ignores `SafetyTownPortal` and `SafetyCrowdAbandon` for a Resupplying step outside town (`walkHomeNoScroll`; drinks and the chicken exit still cancel). Tests: `go test -run 'GateRefusal|WalkHomeNoScroll' ./internal/bot` (two classes).
3. DONE (card outcome-park-step): `parkStep` (the park returns the typed refusal need `parked`, wake: the park ends or the level rises, pause 5 s) and the stale-target holds in `staletarget.go` (`journeyBlocked` hold and `noteRefusal` go through `handleRefused` instead of `spinGuard` and a fixed 1 s pause; the spin guard stays for knob 0). Tests: `go test -run 'JourneyHoldAndPark' ./internal/bot` (two classes).
4. DONE (card outcome-route-guard): `routeGuardHop` refuses through `r.refuse` (typed, need `route_closed`); `guardedExit` keeps a typed refusal of its walk home as it is (`homeWalkErr`) and joins any other fault with the portal error as before. Tests: `go test -run 'RouteGuardHopIsTyped|HomeWalkErr' ./internal/bot` (two classes).
5. Kit gate holds (`kitgate.go`) and the town chores refusals (`town`).
6. The wording match `guardRefusal` in `internal/watch/hub.go` (reads the goal failure text; becomes the typed need).
7. The `*Until` backoff fields that a wake fact can replace (after the above).

Tags: V unit tests `go test -run 'Classify|Refused|LegacyRefusal' ./internal/bot`; U live behaviour (no game run): the pause values and the claim that the next tick re-plans around a board refusal are read from the code, not measured.
