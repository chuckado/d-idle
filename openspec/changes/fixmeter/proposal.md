## Why

The drills (DRILLS5, docs/probes/drills/README.md) list primitive P28 Meter totals as BLOCKED: `d2dcombat.Attach` built no Meter, so the d2d backend could not report damage dealt or taken and kills to a drill or probe.

## What Changes

- internal/d2dcombat/meter.go: the Meter can wrap an existing clock (Attach) and derives damage dealt and kills from the hostile monsters' life in the snapshot; `Totals.String` prints `n/a` for what d2d cannot supply (hits, misses, and dealt/kills when no monster source is wired). Nothing is fabricated.
- internal/d2dcombat/rig.go: `Attach` builds the Meter, off by default (no extra reads in the bot until `Meter.Enable()`), so the bot and soak path is unchanged.
- docs/fixmeter.md: derived vs unavailable, VERIFIED vs UNVERIFIED, live drill recipe.
- The soak report is not touched: it reads the per-fight telemetry lines of `combat.FightTelemetry` (kills, damage taken), not this Meter, and that code is outside this card's files.
- No numeric gate changes (no delta).

Overlap (declared, not edited beyond one note): docs/primitives.md and docs/probes/drills/README.md are owned by the open card m1-liveproof (drills); the edit is a one-line P28 status note, disjoint from that card's open tasks. internal/d2dcombat/** and the shared docs are standing `main` defaults.
