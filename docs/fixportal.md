# FIXPORTAL: portal primitive fixes (card fix-portal)

> status: current. Findings F4 and F5 of docs/primitives-review.md; live proof is the m1-liveproof drill D2 (P10).

## Fixed (shared code, offline-tested)
- F4 (d2d verb): `d2dPortal` clicked the first Portal item, an empty tome included. It now uses `combat.PortalSource`: a scroll or a tome with charges; none means `ErrNoPortal`, no click, and `d2dTown.ToTown` falls back to the walk/waypoint at once.
- F5: `HasTP` (bot kit check, combat resource input, boss view) uses the same pick, so an empty tome is not a portal and the Tome refill rule (town scrollCount) buys scrolls.
- Combat's `portalItem` is the same function: picker and town portal agree by construction (TestFighterAndPortalSourceAgree).

## How tome charges are read
- d2d: `d2denv.Reader.UnitStats` sets stat 70 only when the dump's item `qty` > 0 (reader.go), so an empty tome has no stat 70. V (code); live behaviour of the dump verb for an empty tome: V per internal/town/shop.go note (i4).
- Game.exe: `d2.Reader.UnitStats` stat list, stat 70. Absent at 0 charges is UNVERIFIED (the town code relies on the same live note from i4).
- An unreadable stat list (stock d2d without the dump verb) counts as empty, like the town's `scrollCount`; with no stats source at all a tome counts as usable.

## Further causes (read from navsoak, navsoak2, navsoak3 bot logs)
1. PROVEN, not fixed: 5-10 percent of failures read `context canceled (read ... d2d.sock: i/o timeout)` on tries 1, 2 and 3 within about 10 ms of each other. The verb ran with an already-cancelled context (the step deadline), so the three tries cannot work. Cause is upstream (who cancelled); effect is only wasted retries. A guard `ctx.Err() != nil` in the retry loop is the fix, left for a card that also finds the canceller.
2. PROVEN: the `[no town portal scroll]` trip reason comes from town `scrollCount` (empty tome counts 0) while `d2dPortal` still clicked the tome: the F4 mismatch, fixed above.
3. UNPROVEN: clicks with a tome that has charges or a scroll that still end in `no portal after the click` (about 165 of 246 fallbacks in navsoak3 show tp above 0 at the last shop visit). Candidates: the "i" panel not open after 700 ms, `debug itemat` pixel off, the portal object read gap, a hostile interrupting the cast. Log line added: `d2d portal: click without a portal item=<id> class=<c> portalObjNear=<n> portalObjAny=<n>`; if portalObjAny > 0 the object read or the 25-subtile test is wrong; if both are 0 the click did not cast. Still to add if it recurs: the panel flag after "i" and the `itemat` pixel.
