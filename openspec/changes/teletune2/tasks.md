# Tasks

```yaml
change: teletune2
status: in-progress
agent: teletune2
branch: agent/teletune2
worktree: ~/code/d2lod-wt/teletune2
skip_specs: true
owns:
  - docs/teletune2.md
  - internal/nav/teletune2_test.go
  - openspec/changes/teletune2/**
exclusive: []
gates:
  - "go test -p 1 ./internal/nav ./internal/env ./cmd/d2bot ./tools/knobdump passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "offline: path-aware hop cell, break-even 13 subtiles (T = 18 ticks), stall escape with its own mana floor, refused-hop bound test; not verified live"
live: "DRILLS25 (agent/drills25): Stony Tomb 211 and Halls 533 ticks beat walking (254, 578) V; border leg 717 against 708 walking (+9, 0 hops, same cause as DRILLS24) V; pocket start leaves by 3 hops in 233 ticks, mana min 88.9 V; stall escape (stalledHere, 15 percent floor) NOT exercised live: the path-aware hop leaves the pocket without a stall; after a stall in one process the navigator roams 7560 ticks with 0 hops (harness poison, UNSURE if stalledHere); no-Teleport control and town leg unchanged V"
```

- [x] docs/teletune2.md (break-even, margin, recipe)
- [x] path-aware hop cell (pathHop)
- [x] stall escape and escape reserve knob
- [x] tests (Halls shape, Stony Tomb, short legs, stall floors, refused-hop bound)
- [x] live proof (partly, see live:)
