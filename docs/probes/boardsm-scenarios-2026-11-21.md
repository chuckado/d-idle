# Quest board state machine scenarios, first live run (2026-11-21)

> status: current. Tags: V verified (run id named), S secondary, U unverified. Host: the second worker host (Orion, 16 cores), lab d2d `d2d.survtune-base-fabae935-linuxbuild` (the survtune-base build of the Nov 16 tuning waves, not the K1 build), bot of branch agent/boardsmrun (origin/main 33e6ea8aa plus the tooling fix below), `tools/provingground/run.sh -id <id> -class <C> -seed <n> -parallel 1`, `quests.board` 1 and `quests.boardSM` 1 from each scenario's `trigger.knobs`. Two lanes at a time. EVERY setup is CHEATED (level, gold, waypoints, belt, stash are SETUP writes; scenario `provenance`): mechanics evidence only, never balance or play evidence. `writes after start: 0` in every run that has a ledger.

Run id = the `provingground-out/<stamp>` directory on the host (kept in the run dir, not in the repo). Classes: Paladin seed 11, Sorceress seed 3.

## Verdicts

| scenario | Paladin 11 | Sorceress 3 | note |
|---|---|---|---|
| board.sm.one-head | GREEN 084607, GREEN 112213 | GREEN 100639, GREEN 115026 | V; first Paladin run 083706 was BLOCKED (tooling, fixed below) |
| board.sm.stop-idempotent | GREEN 090818, 104223 | GREEN 105450, 113009 | V |
| board.sm.start-idempotent | GREEN 091336, 104711 | GREEN 110124, 113404 | V |
| board.sm.busy-second-quest | GREEN 091833, 105912 | GREEN 110750, 113652 | V |
| board.retry.runs-a-step | GREEN 094612, 111146 | GREEN 112108, 114339 | V, weak assertion (a step line after minute 3) |
| board.auto.chain-safe-only | GREEN 095220 (vacuous) | GREEN 115927 (vacuous) | no auto multi-part quest in act 1, nothing is tested |
| board.sm.stop-cascades | RED 090226 | RED 104820 | reply has no `changed` (card 6) |
| board.view.reason-always | RED 091239 | RED 112646 | rows carry no reason code (card 6) |
| board.handin.field-tap | RED 092917 | RED 113313 | no awaiting chunk at the start (no setup verb, U); tap gets 409 not_awaiting |
| board.handin.without-run | RED 094559 | RED 114019 | hand-in booked at once; no `handin_ready` offer (card 6) |
| board.hold.restock | RED 085748 | RED 102600 | see finding 2 |
| board.hold.restock-stash-gold | RED 084607 | RED 101834 | see finding 2 |
| board.hold.broke | RED 085106 | RED 103630 | no hold fired; no reason in the view |
| board.hold.too-weak | RED 085451 | RED 104214 | no hold fired in the budget; Sorceress ended clvl 10 |
| board.hold.timeout | RED 090305 | RED 105137 | parked by the old 3-failure counter, no `hold_timeout` |
| board.sm.regress-proxy | BLOCKED 093435 | BLOCKED 111511 | the setup cannot leave the stones done, so the step runs (U); a field start has no write ledger |
| board.sm.restart-mid-chain | BLOCKED 100217 | not run | runner has no `restart_at_s` |
| board.park.persist-restart | BLOCKED 094213 | not run | needs `restart_at_s` and a hero that dies on the chunk |

Determinism pair (same host twice, same seed): all 10 greens repeat GREEN. Wall times differ (about 4 to 19 minutes), so runs are not bit-identical; V for the pair only (two runs each).

Scenario file after this run: the 5 real greens are `active`; the other 13 stay `pending` with the live reason in `pending_reason`.

## Tooling defects fixed here (V)

1. `cmd/d2scenario/run.go`: the watch server and the board sampler started only when a scenario had `trigger.api` or the tap driver, so `board` and `board_heads_le` assertions of a scenario without api calls were BLOCKED ("the quest board not observed"; run 083706). Now `scenario.NeedsBoardSamples` starts them too. Verified: board.sm.one-head GREEN afterwards.
2. `board.sm.stop-cascades`: `log_line` with `max 1` and no `min` demanded at least one line (min defaults to 1); now `max 0` (nothing restarts after the stop). Verified on the Sorceress run: that assertion passes, only `changed` is left red.

## Findings (defects in the merged cards, not fixed here)

1. Card 6 `boardsm-api` is not built: no `changed` on the stop reply (stop-cascades), no `reason` on chunk rows (view.reason-always, hold.broke, hold.too-weak), no handin codes or `handin_ready` offer (handin.*). These REDs are the scenarios doing their job; flip them when card 6 lands. Card: `boardsm-api` (exists as the plan's card 6).
2. Potion-kit hold does not engage (bsm5). In hold.restock, restock-stash-gold and hold.timeout the tower guard sends a chores trip; the trip ends with 3 healing potions and 1 mana potion (`shop: potion plan ... slots 4`, belt column taken by mana) while the guard wants 4 healing (`route guard: blocked tower (3 healing potions < 4)`, internal/task/route.go:220, kit minima clamped by `KitPlan`), so the step ends `guard/ftower/chores/home ran but is still needed (evidence 2)` three times and the chunk is parked (internal/bot/questboard.go, the parked line in the Paladin timeout run 090305). `needs_potions` is a `Fault.Code` without `Hold` (internal/task/gate.go:326, 282), so no `quest board: hold` line is ever written for it, and bug 4 (a potion shortage must be a town trip with no count) is not met with a 4-slot belt. Root cause U (belt capacity versus kit minima). Card: `boardsm-kit-belt`.
3. In broke and too-weak the hero never reached a refusing guard within the budget (it trained in areas 2, 3, 10), so no hold fired; the scenarios need a shorter path to the guard or a longer budget (scenario defect, U). Only one hold line was ever seen: `needs_gold_farm: not ready: gold 100 < 250` in handin.without-run (Paladin 094559), so the hold path itself works.
4. Runner has no `restart_at_s` trigger: restart-mid-chain and park.persist-restart cannot run. Card: `boardsm-restart` (runner trigger plus a state-dir reuse).
5. `no_writes` reports "write ledger not observed" on a field start (regress-proxy), so it can never be green there. Card: `boardsm-fieldledger` (tooling).

## Repeat

```
tools/provingground/run.sh -id <scenario> -class Paladin -seed 11 -parallel 1     # PG_D2D_BIN=survtune-base d2d, PG_DATA, PG_DIR
```
