# Tasks

```yaml
change: fixtelekey
status: in-progress
agent: fixtelekey
branch: agent/fixtelekey
worktree: ~/code/d2lod-wt/fixtelekey
skip_specs: true
owns:
  - docs/fixtelekey.md
  - internal/nav/telecap.go
  - internal/nav/telecap_test.go
  - internal/combat/telekey.go
  - internal/combat/telekey_test.go
  - cmd/d2bot/telewire.go
  - cmd/d2bot/telewire_test.go
  - internal/boss/boss_opts_test.go
exclusive: []
gates:
  - "go test -p 1 ./internal/nav ./internal/combat ./internal/boss ./cmd/d2bot passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "offline: Teleport key and level rule wired from the Fighter bind in both stacks, tests pass; not verified live"
live: "not verified live: see the live proof recipe in docs/fixtelekey.md"
```

- [x] call graph of the finding (docs/fixtelekey.md)
- [x] Navigator capability hooks, Fighter.SkillKey, wireTeleport, boss CanTeleportFn
- [x] tests (key from bind, none without the skill, other key, refused level, bounded refused hops)
- [x] live proof recipe
- [ ] live proof (needs the live slot)
