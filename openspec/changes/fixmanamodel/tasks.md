# Tasks

```yaml
change: fixmanamodel
status: in-progress
agent: fixmanamodel
branch: agent/fixmanamodel
worktree: ~/code/d2lod-wt/fixmanamodel
skip_specs: true
owns:
  - docs/fixmanamodel.md
  - internal/resource/manamodel.go
  - internal/resource/manamodel_test.go
exclusive: []
gates:
  - "go test -p 1 ./internal/resource ./internal/combat passes; whole-repo go vet and go test -short on the worker host pass"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "offline: regen and booking models fixed, tests pass; not verified live"
live: "not verified live: see the proof idea in docs/fixmanamodel.md"
```

- [x] regen model (per-frame whole 256ths)
- [x] tracker booking per backend
- [x] potion over time in the sim
- [x] tests per backend
- [ ] live proof (modelled mana vs the dump over a fight)
