# Town planner, stages 1 and 2 (2026-11-21)

> status: current. Design: docs/design/fable-review-2026-11-21/town-planner.md and appendix B. Nothing here was run on the builder or live: every verdict below is an offline unit test unless it says otherwise. Stopped after stage 2, as asked.

## Cards and commits (merge heads on main)

| card | task | branch tip | merge head | what |
|---|---|---|---|---|
| towncards | TASK-746..765 | `c9d49843d` | `35960ca36` | the 12 cards (11 of B5 plus the SME review card town12) |
| town1 | TASK-746 | `a57287bcd` | `f333ddef6` | 12 B4 scenarios in `tools/provingground/scenarios/townplan.json`, assertion `belt_at_least` |
| town2 | TASK-748 | `ae5d79b8e` | `a05fa2d8c` | `internal/supply` counter; bot and town counters call it |
| town3 | TASK-753 | `d3babf679` | `d3190e47a` | `NeedFor`, pure `Solve`, `profiles.json`, knob `town.plan`, stage 1 shadow trace and minute-line field |
| town4 | TASK-755 | `144106a48` | `98d75df55` | stage 2: guards, `StepChores.Need` and the task-chores shop read the plan |

The chores2 fix (belt and pack counting, restock chores for a kit-only guard refusal) merged first (`cf4ce1207`); stage 2 extends it: `guardFault` now takes the World and the plan.

## What changed at `town.plan` 0, 1, 2

- 0 (default): no behaviour change, no new log or minute-line field. The only code on this path is the counter, which returns the old numbers (belt-only mana stays belt-only for `kitRead`).
- 1 (shadow): every `Chores` visit logs `town plan (shadow): for=.. profile=.. verdict=.. needs ..; has ..; would buy ..`; the minute line carries `townPlan=1`. Nothing acts on it (unit test `TestPlanStage1ChangesNoBehaviour` needs game tables, so it skips here; the stage-1 code only logs).
- 2: guards read `World.KitPlan` (clamped minima, the line the plan cannot close), mana counts belt plus pack, a kit-only refusal the gold cannot pay is a danger fault naming the line, a fixable one inserts the restock chores; the task chores buy `KitHealing` and `KitMana` to the plan minima; a belt-fill step runs before the sell chore.

## Verdicts

| check | result | how |
|---|---|---|
| `Solve` priority, clamp, blocked, deterministic (20 repeats) | pass, Paladin and Sorceress bands | `go test ./internal/supply`, `./internal/town -run SolveShadow` (V, offline) |
| guard opens at a clamped minimum, names an unclosable line, inserts chores when fixable | pass on both classes | `internal/task/kitplan_test.go` (V, offline) |
| bot hands the guards the belt-plus-pack mana count and the plan only at stage 2 | pass on both classes, stages 0 to 2 | `internal/bot/kitplan_test.go` (V, offline) |
| `-race -count=20` | supply, task, town pass | `TestRefusedGearUntilLevelUp` fails with `-count>1` on main too (pre-existing, skipped) |
| scenarios `guard-blocked-restock`, `mana-kit-satisfiable` | NOT RUN | offline twins only; setups are CHEATED mechanics |
| shop buys mana into the pack, belt-first order | NOT VERIFIED | the town rig tests skip on every host without the game tables (`/tmp/d2mpq`) |
| determinism pair of the scenario run | NOT RUN | |

## Interface decisions

- Profiles live in `internal/supply/profiles.json` (embedded, like the potion floor table), not `data/supply`: `//go:embed` cannot reach outside the package.
- `Town.SupplyPlan(snapshot, profile, reserve)` is the one entry point; the bot reaches it through an optional interface (`townAdapter` and the merc town wrapper forward it; a wiring test guards that).
- Guards use profile `monastery_road` for the route and the tower; the boss gates still use their own constants.
- Reconcile note: the bsm5 branch changes the same `fieldReady` refusals (`refusal(...)` with codes); on its merge it must keep `guardFault(w, ...)` and the plan's `Short` text on the danger path.

## Before purse, sell and pickup, route (stages 3 to 5)

1. Run the two stage-2 scenarios and the existing town scenarios with `town.plan` 2 on the builder when free; read the `town plan (shadow)` lines of a stage 1 run first.
2. Decide whether to flip the default to 1 (log only) now.
3. Known gaps: boss gate kits, the Andariel antidote line, repair, merc and key lines are not modelled; `kitRead` mana is belt-only below stage 2.
