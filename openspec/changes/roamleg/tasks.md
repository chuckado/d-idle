# Tasks

```yaml
change: roamleg
status: in-progress
agent: roamleg
branch: agent/roamleg
worktree: ~/code/d2lod-wt/roamleg
skip_specs: true
owns:
  - docs/roamleg.md
  - internal/nav/roamleg_test.go
  - openspec/changes/roamleg/**
exclusive: []
gates:
  - "go test -p 1 ./internal/nav ./internal/combat passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "offline: held roam leg ends the episode in 1 leg (old 320 in the sim), no failed target; not verified live"
live: "not verified live: see the live proof recipe in docs/roamleg.md"
```

- [x] design note (docs/roamleg.md)
- [x] roam.go: ErrMouthGate ends the episode, ErrRoamHeld, counter, debug line
- [x] tests (held leg, no-path leg unchanged)
- [x] live proof recipe
- [ ] live proof (DW.m3 x3, needs the live slot)
