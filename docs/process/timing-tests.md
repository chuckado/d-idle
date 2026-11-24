# Timing tests: no wall-clock assertions in unit tests under `-race`

> status: current (card pipefix2, 2026-11-24, user-approved audit fix 3). Source: pipeline audit row 3 ([pipeline-audit-2026-11-24.md](pipeline-audit-2026-11-24.md)): `TestVaultPollCost` failed the merge gate under `-race` (12 to 14 ms against a 10 ms ceiling) and `TestPlayerResetStartsOnTheNextTick` turned main red from a goroutine-versus-fake-clock race.

## Rules for tests

1. A test that runs by default (`go test -short ./...`, the merge gate's `go-whole` and `go-race`) asserts no wall-clock ceiling (`time.Since`, `NsPerOp`, a `Millisecond` bound) unless it is already insensitive to a 10 times slowdown and a loaded builder.
2. Prefer, in this order:
   - the fake clock (`vclock` in internal/bot, `Clock` in env): freeze it at the act (`clock.freeze()`) and assert it did not move, or count its ticks; never compare fake time with a real number when other goroutines also advance it;
   - a ratio against a baseline measured in the same test (a poll at 2400 records against a poll at 240);
   - a ceiling that is at least 10 times the plain-build number and far below the thing being excluded (a 1.5 s slow verb against a 1.2 s ceiling, not a 400 ms verb against a 350 ms one);
   - `if raceEnabled { skip the ceiling }` for the genuinely wall-clock ones (below).
3. A lower bound ("the pump cannot be faster than 20 parts in 300 ms") is fine: a slow box never breaks it.

## The race guard

A package that needs it gets a pair of test files, `race_on_test.go` (`//go:build race`, `const raceEnabled = true`) and `race_off_test.go` (`//go:build !race`, `const raceEnabled = false`); existing in internal/watch, cmd/d2dview (name `raceOn`), internal/loot, internal/nav, internal/bot. Use it as `if !raceEnabled && d > ceiling` so the measurement still logs. The race detector slows code 5 to 10 times (V: vault poll, builder, 2026-11-24).

## Flaky list and the gate

`tools/ci/flaky.txt` lists only tests that cannot be fixed yet (TestName, package, owner, measured rate). The merge gate (`tools/merge.sh` in the ops repo) re-runs ONCE when every failing test is on the list, in both `go-whole` and `go-race` (pipefix2 added the race half). The list is empty today: a test that flakes is fixed or guarded, not listed, unless a card owns the fix.

## State after pipefix2 (the audit's 17 tests)

| Test | Action |
|---|---|
| internal/bot TestPlayerResetStartsOnTheNextTick | fake clock frozen at the press; asserts it did not move (no wall bound, no race with the loop's own sleeps) |
| internal/watch TestVaultPollCost | already skipped under race (#1261) |
| internal/watch TestVaultPagingPerformance (150 ms) | ceiling off under race |
| internal/loot TestShadow... (1 ms per decision) | ceiling off under race |
| internal/nav plan 1500x1500 (10 s) | ceiling off under race |
| internal/bot leveling kit-trip loop (10 s) | ceiling off under race |
| cmd/d2dview pacing part latency (15 ms) | ceiling off under race |
| cmd/d2dview view_test viewer behind the pacing sleep | ceiling 500 ms to 1.2 s (the sleep is 2 s) |
| internal/d2dtown HoverFirst wall wait (100 ms) | 400 ms (the waits it must not sleep are 470 ms) |
| internal/d2dctl TestTimeoutAndContext | slow verb 1.5 s, ceiling 1.2 s, redial wait 1.7 s |
| internal/d2denv freeze, waits, fixobj | frozen stretch 100 ms with a 100 ms bound on the leak; lockstep pause bound 1 s to 3.5 s (a 4 s real pause is the thing excluded) |
| internal/art speed_real | needs the real install (skips on the builder); left, U |
| internal/nav walksmooth (500 ms), internal/d2dctl fixstall4 (2.5 s), internal/d2denv virtual Now | simulator or virtual time, not wall time; no change |
| internal/watch videoresume | lower bound only; no change |
