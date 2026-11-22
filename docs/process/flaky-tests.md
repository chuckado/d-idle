# Flaky tests: the waiting pattern

> status: current. Load-sensitive tests cost 45-minute gate re-runs; this is the rule that keeps new ones out.

- A test that waits for something to HAPPEN uses an event or a bounded poll with a generous deadline (30 s; it costs nothing when the thing happens at once). Never `time.After(2 * time.Second)` for a positive wait, never `time.Sleep(n)` then assert.
- A rate assertion ("N parts in 1 s") is a wall-clock assumption: count the events instead and assert on the count, plus a lower bound on elapsed time when the point is "it lasted long enough".
- Asynchronous consumers (the event ring, the bus bridge, a build goroutine) are waited for by their own progress marker (`last >= bus.Seq()`), not by a helper's short internal settle bound.
- A NEGATIVE check ("nothing happened in 100 ms") may stay short: load only makes it weaker, never flaky.
- Never weaken the assertion; only the waiting mechanism changes.
- Reproduce before fixing: `labrun <worker host> -- go test -race -count=300 -cpu 1,4 -run <Test> ./internal/<pkg>/` with a CPU burner running (one job at a time).
