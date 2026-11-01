# Tasks

```yaml
change: fixmeter
status: archived
agent: fixmeter
branch: agent/fixmeter
worktree: ~/code/d2lod-wt/fixmeter
skip_specs: true
owns:
  - internal/d2dcombat/meter.go
  - internal/d2dcombat/meter_test.go
  - internal/d2dcombat/rig.go
  - docs/fixmeter.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof after merge: spawn a pack, fight, compare the Meter totals with the monsters' life deltas and the kill count (docs/fixmeter.md)"
merge: f795c60
live: "DRILLS10: D-meter PASS, kills 13 = 13, dealt 121 = 121, taken 106 = 106, hits and misses n/a"
result: "live PASS (DRILLS10)"
```

## 1. Work

- [x] 1.1 Meter wraps a clock, derives damage dealt and kills from monster life, n/a for the rest
- [x] 1.2 `Attach` builds the Meter (off by default)
- [x] 1.3 offline tests (fake fight totals, n/a, off forwards only, Attach wiring)
- [x] 1.4 docs/fixmeter.md, P28 note, INDEX, CHANGELOG, BACKLOG
- [x] 1.5 live proof (DRILLS10)
