## Why

Gate 1 rerun 2 (docs/soaks/gate1r2-20261101.md, Review 2 section 6): at clvl 11 and up 40 percent of the minutes were in the Blood Moor and Cold Plains because the kit gate (internal/bot/kitgate.go) allows no other field while the kit is missing. The gate was written for clvl 1 to 7. The user decided 2026-11-02 to adopt the veteran advisor's kit gate rule.

## What Changes

- internal/bot/kitgate.go: `kit` (the one minimum-kit definition), `kitTrip` (town trip from clvl 8, cap, `kitgate-skip` log), `kitHolds` (the hold only without portal and healing from clvl 8); knobs `resource.kitGateLevel`, `kitMinHP`, `kitMinMP`, `kitTripCap` (default on, 0 for the level restores the old gate).
- internal/bot/goal.go (`choresDue`), resource.go (`logTrip` counts kit trips), observe.go (counters).
- Rules and drill checklist: docs/kitgate.md.

## Impact

No spec delta. Based on agent/soakfixes2 (tpglitch debounce, pingpongrate); merge it first.

Overlap (declared): internal/bot/goal.go and docs/OWNERS lines with soakfixes2, soakstall, pingpongrate and tpglitch (same area, different hunks).
