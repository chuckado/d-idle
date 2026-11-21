# Town planner scenarios (card town1)

> status: current (2026-11-21). Source: docs/design/fable-review-2026-11-21/town-planner-b-model.md section B4.

`tools/provingground/scenarios/townplan.json` holds the 12 sketches of B4 as `didle.scenario/1` scenarios (id prefix `town.plan.`), each on the Paladin and Sorceress matrix with seeds 11 and 3 and `no_writes`. The one new assertion kind is `belt_at_least {code: hp|mp|rv, min}`: potions of that family in the belt alone (`internal/scenario/state.go`, V by unit test). The knob `town.plan` they set arrives with card town3; before it merges, a run reports an unknown knob.

## Status: nothing was run

The builder was not used (rule: heavy runs only when free, one job at a time). All setups are SETUP-write CHEATED states (mechanics only). So the RED/green column is the design's expectation, not a measurement:

| id | expected on today's main | green needs |
|---|---|---|
| guard-blocked-restock | RED (I1, run 20261120-200952-46d2) | stage 2 (card town4) |
| mana-kit-satisfiable | RED for the Sorceress (I7) | stage 2 |
| broke-after-bank | RED expected (I9), U | stage 3 |
| malus-errand-gold | RED (I10) | stage 3 |
| belt-refill-from-pack | U, expected RED on the keep count | stage 2 (belt fill before sell) and stage 4 |
| sell-keeps-need | RED (I3) | stage 4 |
| scroll-hoard | green in town; the ground half U | stage 4 |
| gold-short-priorities | partly green | stage 7 |
| short-gold-handover | RED | stage 7 |
| one-visit, determinism | U | stage 5 |
| no-cast-in-town | U (cannot force the entry) | unit test of the cast function |

U: the ground-item quality (2) and the 10 cap items that fill the pack are guesses; a run that cannot reach the state is a scenario fault, not a finding.
