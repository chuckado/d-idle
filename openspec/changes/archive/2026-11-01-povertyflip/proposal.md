## Why

In the Gate 1 run the poverty verdict flipped with the area (Cold Plains: `no town portal`, poverty on; Blood Moor: `kit complete`, poverty off) and made the bot alternate between two emptied fields for 61 and 312 virtual minutes (docs/soaks/gate1-20261101.md, Review; 82 of 86 poverty-on lines said no town portal while 131 scrolls were bought). The cause of the flip is unknown until the `inv-tp-dump` line from a rerun is read, so this card only stops it from steering the route: no observable behaviour change to any spec number, only fewer area flips.

## What Changes

- internal/bot/resource.go, the poverty check input: once poverty mode switched on (or off), the opposite verdict needs two consecutive reads in the same area or a town visit, so a single per-area read cannot flip it.
- If the rerun's `inv-tp-dump` shows the real cause (stale tome count, wrong container, stat read), fix it at that source instead and drop the hysteresis.

## Impact

No spec delta. Overlap (declared): none in flight; tpdump.go belongs to the merged card hastpdump and is only read.
