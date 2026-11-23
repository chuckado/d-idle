# Cached whole-repo gate: measurement, and the nightly uncached race run (card mgcache, TASK-747)

> status: current, 2026-11-21. Design: `docs/design/fable-review-2026-11-21/proving-ground-gate.md` sections 4.2 and 4.3; owner answers Q3 and Q4 (ADR 0027). Code in the private ops repo (merge script, queue tool, two user timers); described by role.

## The gate today and what changed

The merge gate still runs `go vet ./... && go test -short -count=1 -p 4 ./...`. `-count=1` is NOT dropped: the owner's switch waits for the measurement (card AC 2 stays open). What exists now:

1. Per-merge gate stats (observation only, `MERGE_TIMINGS=1` is already on in the queue worker's settings): the merge result's `timings` gets `go-whole-secs`, `go-whole-cached` (packages answered from the test cache), `go-whole-pkgs`, `go-whole-count1` (1 while `-count=1` is in the gate), `go-race-secs` and `load1x100` (builder 1-minute load times 100). `python3 tools/gate_phases.py` (ops repo) prints the table and median/p90 per mode from the queue file; `--tsv` for a spreadsheet.
2. The switch itself: `MERGE_GATE_CACHED=1` in the queue worker's settings drops `-count=1` from the whole-repo test only (default 0, nothing changes). With it on, the stats above record the cached merges and the same script compares the two modes once there are 20 of each.
3. A replay measurement that needs no change to the live gate: `tools/gate_cache_measure.sh 20` takes the last 20 code merges of main (oldest first), syncs each commit to a scratch directory on the builder (rsync keeps unchanged files' mtimes like the labrun mirror) and runs the whole-repo vet and short tests WITHOUT `-count=1` on the long lane (own lock, nice 12) with a separate build and test cache. One TSV line per commit: seconds, packages, cached packages, builder load. The first line is a cold cache. It runs from a user timer at 01:30 New York time (inside the night window) and resumes where it stopped.
4. The nightly uncached whole-repo `-race` run: `tools/nightly_race.py`, a user timer at 23:10 New York time. It refuses to run outside 23:00 to 07:00 (an all-day grant the user states is `NIGHTLY_AD_HOC=1` for one run), skips when the second worker host has under 25 percent idle plus nice CPU (`top`, second sample), runs from a scratch worktree of origin/main as `nice -n 19 go vet ./... && go test -race -count=1 -p 3 ./...` with a deadline of 06:55, and never touches the host's containers. Result lines go to a ledger file (`green`, `skipped-busy`, `incomplete`, `flake`, `red`). A red result re-runs only the failing packages once: a pass is a flake candidate (ledger line, no hold); a reproduced failure records a nightly red that HOLDS THE MERGE QUEUE like a red CI (`mergeq status` prints `NIGHTLY RED holds the queue`); a green night or a strictly later green head of main lifts it; `mergeq nightly-clear` lifts it by hand.

## Data so far (V, queue file, 405 code merges with a gate phase, all with `-count=1`)

Gates phase median 280 s, p90 1041 s (the design review read 435 s and 1990 s over a window of heavy agent load). The per-package split (cached hit rate, whole-repo seconds) exists only for merges after this change and for the replay; the replay has not run yet when this page is written (first run is the 01:30 timer): read `gate_phases.py` and the replay TSV, and decide Q3 with 20 merges of each.

## Verification

- V: ops `tools/tests/test_nightly_race.py` (5 tests: outside the window nothing runs and the ad hoc flag lifts it, busy host skips, green is niced uncached race and clears a red, flake passes on the re-run without a hold, reproduced red sets the hold) and `test_mergeq_watchdog.py` (nightly red shows in status and clears); `tools/tests/test_merge_sh.sh` (69 checks) and `test_mergeq.py` pass with the merge script change; `--dry-run` of the replay lists the 20 commits and the nightly dry run reports the window.
- U: the cache hit rate on the builder (needs the replay); whether the live mirror keeps file times stable enough (the replay assumes it does); behaviour of the nightly run on a loaded second host (it skips under 25 percent free, so it may skip most nights).
- The nightly hold lives in the queue worker: it takes effect when the worker is restarted.
