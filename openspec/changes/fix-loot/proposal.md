## Why

docs/primitives-review.md F2: `looterAdapter` (cmd/d2bot/adapters.go) did not forward `SetRecovering` or `RecoveryTotals`. The runner reaches the Looter only through `bot.Looter` and asserts the optional `recoveryLooter` interface (internal/bot/broke.go), so the assertion failed silently: the recovery pickup rules (every gold pile, sellable gear up to the kit line, sell pickups may use the last pack cells) never switched on, and 0 of about 20,000 `recovery: loot` lines were non-zero.

## What Changes

- `looterAdapter` forwards `SetRecovering` and `RecoveryTotals`.
- Compile-time assertions in cmd/d2bot/behaviors.go for the three optional Looter interfaces the runner asserts on (recovery, `leftBehinder`, `UnitEvaluator`), so a dropped method breaks the build.
- Scenario test through the adapter (cmd/d2bot/looteradapter_test.go).
- Not changed: the Looter itself and its poor/recovery rules (internal/loot/poor_test.go already covers them directly).

## Impact
Shared by both backends (the adapter wraps the same Looter for Game.exe and d2d). Recovery now actually collects gold piles and sellable items: this is the documented design (docs/recovery.md), so the recovery gold economy improves; the size is not measured. Delta: town-basics ADDED requirement "Recovery pickup is wired".
