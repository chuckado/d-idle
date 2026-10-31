## Why

DRILLS6 (docs/probes/drills/README.md, findings F7 and F8) showed fixstall only partly works live: H.pause opened Kashya in 1 of 2 tries (the other stalled 28 s near the camp objects), and the D6.d pocket escape opened a town portal but then walked to it with the pinned character (62 s stuck).

## What Changes

- internal/d2dctl town_verbs.go, approach_path.go: after a stall `ApproachNPC` plans a route over `debug grid 5` (other NPCs blocked), walks it leg by leg, and only then falls back to the blind side step.
- internal/town trip.go, internal/d2dtown adapter.go: `enterPortal` does not `GoTo` the portal when the input's Interact walks itself (d2d `EnterPortal` verb), so a pinned character takes the portal it opened.
- No numeric gate changes (no delta).
