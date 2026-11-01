# Tasks

```yaml
change: skmeleebot
status: archived
merge: 015f4c1
agent: skmeleebot
branch: agent/skmeleebot
worktree: ~/code/d2lod-wt/skmeleebot
skip_specs: true
owns:
  - internal/resource/melee.go
  - internal/resource/melee_test.go
  - internal/resource/types.go
  - internal/resource/attack.go
  - internal/resource/sim.go
  - docs/skmeleebot.md
  - openspec/changes/skmeleebot/**
exclusive: []
gates:
  - "go test -p 1 ./internal/resource ./internal/combat ./cmd/d2bot; whole-repo vet and test on the worker host"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "melee model in the simulator matches the SKMELEE traces (unit-level; production not yet fed: SKMELEEWIRE); not verified live"
live: "offline only, nothing run live"
```

## 1. Work

- [x] 1.1 Blow model: HitChance, Blow.MeanDamage / Swing, Foe, IsBlunt
- [x] 1.2 B1 simulator melee rolls from the model
- [x] 1.3 B2 weak() with the to-hit term
- [x] 1.4 B3 melee rung pack limit with the Blunt-vs-undead gain
- [x] 1.5 docs/skmeleebot.md (B4 note, before and after, approximations)
