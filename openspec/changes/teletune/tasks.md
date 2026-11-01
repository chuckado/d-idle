# Tasks

```yaml
change: teletune
status: in-progress
agent: teletune
branch: agent/teletune
worktree: ~/code/d2lod-wt/teletune
skip_specs: true
owns:
  - docs/teletune.md
  - internal/nav/teletune.go
exclusive: []
gates:
  - "go test -p 1 ./internal/nav ./internal/env ./cmd/d2bot ./tools/knobdump passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "offline: break-even (minimum leg, mana reserve, walk instead of wait) with two nav knobs, tests pass; not verified live"
live: "not run: group T legs of docs/teletune.md with the new build"
```

- [x] break-even derivation (docs/teletune.md)
- [x] hopWorth, min leg, mana reserve, no idle wait, skill-data cost
- [x] knobs in NavTuning, docs/knobs.md regenerated
- [x] tests (DRILLS22 geometry, reserve, short leg, no wait, no Teleport unchanged)
- [x] live recipe
- [ ] live proof
