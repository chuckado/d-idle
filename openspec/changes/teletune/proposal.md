## Why

DRILLS22 proved Teleport hops fire live but cost more time than walking (448 against 254 and 851 against 578 ticks): a 17 subtile hop pays about 23 mana and the navigator waited for regeneration. docs/teletune.md derives the break-even.

## What Changes

- internal/nav: a hop is declined (the leg walks) when the walk to its landing cell is shorter than `nav.teleMinLeg` (36 subtiles) or the mana after it falls below `nav.teleReservePct` (40 percent of max mana). No idle wait for hop mana. Cost from the skill data at the learned level (`Navigator.TeleportCost`, wired in cmd/d2bot/telewire.go).
- internal/env: the two knobs in `NavTuning` (docs/knobs.md regenerated).
- Unchanged: which exits are used, border exits, boss teleport stops, the refused-hop bound, level rules, characters without Teleport.

## Impact

Offline only; awaits the live legs in docs/teletune.md. No spec names hop rules (grep openspec/specs), so no delta.

Overlap (declared): cards fixtelekey and drills22 (unmerged) own internal/nav/telecap*.go, cmd/d2bot/telewire*.go and docs/fixtelekey.md; this card edits cmd/d2bot/telewire.go (the cost wiring), internal/nav/telecap_test.go (the shared hop helper), internal/nav/{nav,teleport,teleport_test}.go, internal/env/tuning_nav.go and docs/knobs.md (a standing default of fixcastgap) without listing them in owns.
