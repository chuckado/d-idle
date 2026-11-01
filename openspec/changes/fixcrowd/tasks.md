# Tasks

```yaml
change: fixcrowd
status: in-progress
agent: fixcrowd
branch: agent/fixcrowd
worktree: ~/code/d2lod-wt/fixcrowd
skip_specs: true
owns:
  - docs/fixcrowd.md
  - docs/knobs.md
  - internal/combat/crowd.go
  - internal/combat/fighter.go
  - internal/combat/resource.go
  - internal/combat/crowd_test.go
  - internal/combat/fixcrowd_test.go
  - internal/env/tuning_combat.go
  - internal/resource/types.go
exclusive: []
gates:
  - "go test -p 1 ./internal/combat ./internal/env ./internal/resource ./tools/knobdump passes; whole-repo vet and go test -short on the worker host"
  - "live proof after merge: D-portalmelee x3 passes (Safety is chicken with 3 zombies in reach, scroll unspent) and D-portalmelee.ctl still passes (docs/fixcrowd.md)"
merge:
result:
```

## 1. Work

- [x] 1.1 derive the reach from the DRILLS11 data and the code (docs/fixcrowd.md)
- [x] 1.2 knob combat.crowdAdjacent (default 8), Crowd takes the reach, knobs.md regenerated
- [x] 1.3 tests: drill geometry, single adjacent control, far monsters, knob override
- [ ] 1.4 live proof (not verified live)
