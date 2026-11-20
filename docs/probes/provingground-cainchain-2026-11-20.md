# Proving Ground: Search for Cain chain (card cainchain, 2026-11-19/20)

> status: current. Tags: V verified (run named), S secondary, U unverified. Host: the builder main lane, lab d2d slots, bot of branch agent/cainchain, `tools/provingground/run.sh -id cain.chain -class <C> -seed <n> -parallel 1`. Every setup is CHEATED (scenario `provenance`): mechanics evidence only, never balance or play evidence.

## Verdicts (V, builder)

| scenario | verdict | job |
|---|---|---|
| cain.chain Paladin 11 | GREEN, 80 virtual min, 18 min wall | 20261119-232655-7217 |
| cain.chain Paladin 5 | GREEN, 80 virtual min, 24 min wall | 20261120-000256-0f09 |
| cain.chain Sorceress 3 | GREEN, 80 virtual min, 34 min wall | 20261120-002704-6dae |
| cain.chain Paladin 11 (repeat, same host) | see "Determinism" | 20261120-010053-a896 |

## Root causes found (V from bot logs of the red runs)

1. Stale town portal: after a walk home from Tristram the standing town portal still led to the Dark Wood. `Town.portalFrom` now records where the last portal trip began; `BackToField` refuses a portal-only level (`data.PortalOnly`, Tristram) with `ErrPortalStale` when the standing portal leads elsewhere.
2. No walkable way into Tristram: the Stony Field to Tristram nav hint is gone (a walk roamed 5 minutes). The cage step is needed anywhere outside Tristram and splices in the walk to Stony Field plus the Tristram portal (`quest.search_for_cain/reenter/stony`, `reenter/portal`). U: that the red portal stays open after the stones (it did in the green runs).
3. The level planner ran between the stones chunk and free_cain while the hero stood in Tristram: every farm target read "no route", all Act 1 areas were skipped as unreachable for 150 s, the hero went to town, and the re-entry step was then refused ("area banned after repeated deaths", wrongly worded; deaths were 0). `LevelTo` now waits (2 s per slice, 30 slices) in a portal-only level; nav-fault skips (`unreachable`, `nav-stuck`) no longer stop the steps or the walk of a player-started chunk (`TestBannedAreaNoteAndWaiver` follows). This race between the board accept and the planner is the most likely run-to-run difference of the earlier flaky verdicts (U: no pair was run before the fix).
4. Frozen virtual clock: a refused board step ran nothing, so under lockstep no time passed and the chunk spun for the rest of the wall budget. A danger refusal now takes the idle path and an idle tick in town sleeps (`boardFailGap / 3`).

## Determinism

See the table row above; both runs on the builder (the Orion lab d2d may differ). Runs are not bit-identical; the requirement is the same verdict.

## Not done

- The cancelled town-portal cast of a potion trip is not retried (its context is dead); the walk home plus re-entry covers it.
- The 4 fixes in the bot (3, 4) have no unit test beyond the waiver test: the bot package has no `boardRun` harness; the scenario is the test.
- `-race -count=20` ran on internal/task only; internal/nav ran `-count=3` (the 20 repeats exceeded 10 minutes under nice 19); the town tests need the game tables (skipped on Orion; passed with them on the builder, without -race).
