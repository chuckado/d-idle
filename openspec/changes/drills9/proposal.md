## Why

fixstall4 (town verbs follow line of sight, route at every stall) is proven offline only. DRILLS9 runs H.pause and the full D10 fault drill live, isolates the group-run flakiness DRILLS8 found (F14) and repeats the regression pass.

## What Changes

- Harness fixes only (cmd/d2drill, internal/drill) so each case sets up its own state.
- Results in docs/probes/drills (drills9 files, README section DRILLS9), primitives rows, a `live:` line on the fixstall4 card.
- No product code changes: failures are findings.
