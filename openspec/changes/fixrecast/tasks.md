# Tasks

```yaml
change: fixrecast
status: in-progress
agent: fixrecast
branch: agent/fixrecast
worktree: ~/code/d2lod-wt/fixrecast
skip_specs: true
owns:
  - docs/fixrecast.md
  - docs/knobs.md
  - internal/combat/necro.go
  - internal/combat/buffrecast_test.go
  - internal/data/skills.go
  - internal/data/data_test.go
  - internal/env/tuning_combat.go
exclusive: []
gates:
  - "go test -p 1 ./internal/combat ./internal/data ./internal/env ./tools/knobdump passes including TestFrozenArmorTimerLeavesGap, TestFrozenArmorRecastBeforeExpiryByClock, TestFrozenArmorRecastBeforeExpiryByState, TestAuraFrames"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: "not verified live: see the live recipe in docs/fixrecast.md"
```

- [x] buff recast logic found (Fighter.buffDue, Buff)
- [x] real durations verified from Skills.txt rows
- [x] state read (Frozen Armor 10) and duration-based limit with margin
- [x] tests: gap before, no gap after, levels 1 and 2, margin, fallback
- [x] docs, primitives note, backlog, changelog
- [ ] live check (hidden slot)
