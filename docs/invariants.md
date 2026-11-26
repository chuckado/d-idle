# Invariant monitor (knob `bot.invariants`)

> status: current (card invmon, 2026-11-25). Logging only. Source: P1 of `docs/research/fable-logic-review-2026-11-25.md` (section 3.1, 3.3).

## What it does
- Once per tick `Runner.invTick` (called from `step` right after the poll) builds an `invFacts` value from the snapshot and loop state and runs the table `invRows` (`internal/bot/invariants.go`). A hit logs ONE WARN line `invariant <name>` with `detail=...`, rate limited per name (60 s, `GapSec` in the row), and counts per name.
- The minute line gains `inv=name=n,name=n` (hits since the last minute line); nothing when none, so the default line is unchanged. `Runner.InvariantHits()` returns lifetime counts per name.
- Nothing is acted on. Knob `bot.invariants` 1 (default) logs, 0 is off.
- Thresholds are rows of the table (Secs, Count, Pct). The per-hero numbers (drink line, chicken line, kit minimum, belt) come from the facts and the tuning, so no row names a class or build.

## Rows
| name | fires when |
|---|---|
| `potion-not-drunk` | in the field under the drink line, healing potion in the belt, none drunk for 8 s |
| `portal-not-cast` | in the field at or under the chicken line, hostiles near, a town portal scroll, same area after 6 s |
| `field-entry-hurt` | first tick outside town with life under 50 percent |
| `field-entry-no-kit` | first tick outside town while `kit.missing()` (portal, healing, mana minimum) is not empty |
| `resupply-in-field` | state `resupplying` outside town for 10 s |
| `refusal-as-failure` | a refusal (`not ready`, `errAreaGated`) reached `fail` |
| `refusal-storm` | 10 refusals (`noteStepRefused`) with no kill, xp or area change |
| `life-zero-alive` | alive (not dying) with life 0 for 3 s |
| `xp-frozen` | in the field, not paused, no xp change for 30 min |
| `mode-no-exit` | the same state and step for 20 min, not paused, parked or idle |
| `corpse-unresolved` | `corpseTown` pending for 15 min |
| `town-thrash` | 17 field entries within an hour |
| `economy-stall` | purse and stash gold under 100 for 2 h while kills continue |

Not yet rows (need the reflex lane, P3): safety latency (verdict to potion) and cancellation latency.

## Oracle use
`Deps.OnInvariant(name, detail)` is called for every logged hit. A lockstep or Proving Ground runner sets it (or reads `InvariantHits()` at the end) and fails the run when anything fired. The hook point is in place; wiring it into `cmd/d2scenario` is a follow-up in the ops-side runner (U: not run).

## Tags
- V: unit tests `go test -run 'Invariant|InvTick' ./internal/bot` (two heroes with different maxLife, drink and chicken lines, kit lists).
- U: threshold values are the review's first numbers (section 7.1) and have not been tuned on live logs; `resupply-in-field` may fire on the portal-back leg of a trip (state is resupplying until the portal is taken). Read the counts for a few days before trusting a row.
