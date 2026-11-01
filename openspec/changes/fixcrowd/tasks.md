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
  - internal/combat/crowd_test.go
  - internal/combat/fixcrowd_test.go
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
- [ ] 1.5 follow-up (later drill agent, after drills11 merges): the live drill's wait must change from '3 within 3 subtiles' to '3 within crowdAdjacent' in internal/drill (owned by the drills cards)
