# Tasks

```yaml
change: armorcut
status: in-progress
agent: armorcut
branch: agent/armorcut
worktree: ~/code/d2lod-wt/armorcut
skip_specs: true
owns:
  - openspec/changes/armorcut/**
  - docs/armorcut.md
  - internal/combat/fighter.go
  - internal/combat/armorcut_test.go
exclusive: []
gates:
  - "go test -p 1 ./internal/combat passes; go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "offline: cut casts retried, no cast mid-hit, tests pass; not verified live"
live: "DRILLS28 (branch agent/drills28): step 0 verified live (d2d refreshes a set state, so candidate d is out); the cut-retry path NOT exercised: no cut line in D26-A on lab-armorhit or lab-nexthit2 (4 runs); no hit-cut reproduced, so the retry is still not verified live"
```

- [x] evidence from D26-A and the hypothesis it supports (docs/armorcut.md)
- [x] fix in buffPass with tests
- [x] live recipe
- [x] step 0 and D26-A rerun on both fork binaries (DRILLS28); the retry itself still unproven live
