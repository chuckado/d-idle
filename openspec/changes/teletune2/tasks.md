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
live: "awaiting the three legs and the pocket start (docs/teletune2.md)"
```

- [x] docs/teletune2.md (break-even, margin, recipe)
- [x] path-aware hop cell (pathHop)
- [x] stall escape and escape reserve knob
- [x] tests (Halls shape, Stony Tomb, short legs, stall floors, refused-hop bound)
- [ ] live proof
