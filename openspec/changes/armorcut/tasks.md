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
live: "not verified live: see the recipe in docs/armorcut.md"
```

- [x] evidence from D26-A and the hypothesis it supports (docs/armorcut.md)
- [x] fix in buffPass with tests
- [x] live recipe
- [ ] live proof (D26-A rerun on both fork binaries)
