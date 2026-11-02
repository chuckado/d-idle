## Why

Gate 1 rerun 2 (docs/soaks/gate1r2-20261101.md, Review 2 sections 6 and 7): the `inv-tp-dump` lines show the town-portal verdict dropping to false while the tome still holds scrolls: 95 provable cases (same tome count in the dump before and after, drop-out 2.6 s median, 28 s at the 90th percentile). 32 of 64 kit-gate-on edges and 22 of 48 poverty-on edges follow one. The kit gate then keeps a clvl 14 character in the Blood Moor and Cold Plains (40 percent of its time at clvl 11 and up, about 0.5 XP per kill).

## What Changes

- Fix at the source: internal/d2denv/reader.go:139 builds no stat 70 when the dump row has quantity 0, and internal/combat/portal.go:24 reads that as an empty tome. Keep the last positive tome count across a quantity 0 read until a use (a cast, a portal click) or a shop action is seen, or for a few virtual seconds; one read of 0 must not flip the verdict.
- Move the debounce (internal/bot/tpdebounce.go, now only in leveling.go:136) to the single source of the verdict (internal/bot/resource.go:33) so the kit gate (kitgate.go:31) and the other readers see it.
- Probe the fork side (U): one `dump items` in the first second after a waypoint with a tome of quantity above 0; if the fork reports quantity 0 there, that is a fork patch candidate (private fork rules).

## Impact

No spec delta. Overlap (declared): povertyflip (merged, tpdebounce.go) is replaced or extended; internal/bot/tpdump.go (hastpdump) is only read.
