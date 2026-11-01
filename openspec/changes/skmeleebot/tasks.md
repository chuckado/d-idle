# Tasks

```yaml
change: skmeleebot
status: in-progress
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
result: "Offline, no live proof owed. B1 to B3 of docs/research/sk-melee-trace.md in internal/resource, table-driven tests with values from the oracle's reading functions; the stock simulator scenario now rolls to-hit and damage (95 % of a mean 3.30 vs a flat 3). Build-specific data: the Blow and Foe values come from the caller (class to-hit factor, weapon stats, ItemTypes chain); the stock scenario's numbers are its own. Production wiring of Input.Blow is open."
live: "offline only, nothing run live"
```

## 1. Work

- [x] 1.1 Blow model: HitChance, Blow.MeanDamage / Swing, Foe, IsBlunt
- [x] 1.2 B1 simulator melee rolls from the model
- [x] 1.3 B2 weak() with the to-hit term
- [x] 1.4 B3 melee rung pack limit with the Blunt-vs-undead gain
- [x] 1.5 docs/skmeleebot.md (B4 note, before and after, approximations)
