## Why

Gate 1 rerun 2 (run gate1-20261101-1949, bot aa0dbaa with the SOAKFIXES and the HASTPDUMP diagnostic) finished NOT PASSED on XP/h and on progress; it needed a review against run 1: numbers per clvl band, stall windows, guard firings, restart causes, the HasTP root cause (now visible in the `inv-tp-dump` lines), verdict per criterion under the baseline audit's trust levels, and a rerun recipe.

## What Changes

- docs/soaks/gate1r2-20261101.md: the generated record of run 2 plus the section "Review 2" (new file: the generated record of the card gate1run-20261101 overwrites the run 1 record of the same name, so run 2 has its own file).
- docs/probes/soak-logs/gate1r2-20261101-minutes.jsonl.gz: the minute lines of run 2.
- New cards tpglitch, pingpongrate, kitgatelevel (not implemented).

## Impact

No spec delta, no code, nothing run. Overlap (declared): docs/soaks/gate1-20261101.md and docs/probes/soak-logs/gate1-20261101-* belong to gate1run-20261101 and are only read; its branch holds run 2 under the run 1 file names and must not be merged over main's run 1 record.
