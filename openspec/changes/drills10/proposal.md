## Why

Three merged fixes (fixf17, fixtalk, fixmeter) were proven offline only. DRILLS10 runs them live on the hidden slot 0 with the lab build cheats3 and records the numbers.

## What Changes

- Harness only (cmd/d2drill, internal/drill): drill D-meter (group M) with an independent per-unit sampler, FightSpec.Meter.
- Results in docs/probes/drills (drills10 files, README section DRILLS10), primitives rows, `live:` lines on the three fix cards.
- No product code changes: a failure is a finding with logs and a diagnosis.
- Ownership: the paths this change also writes (docs/probes/drills/**, docs/primitives.md, cmd/d2drill/**, internal/drill/**, the `live:` lines of the fixf17, fixtalk and fixmeter cards) are the m1-liveproof globs of docs/OWNERS; only the card directory is listed in `owns:` to avoid overlap warnings, and no OWNERS lines are added for those paths.
