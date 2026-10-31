# Tasks

```yaml
change: fix-pickup
status: in-progress
agent: fixpickup
branch: agent/fixpickup
worktree: ~/code/d2lod-wt/fixpickup
owns:
  - internal/town/belt/belt.go
  - internal/town/belt/belt_test.go
  - internal/town/beltfill.go
  - internal/town/sim/sim.go
  - internal/loot/export_test.go
  - internal/loot/fixpickup_test.go
  - docs/fixpickup.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... plus go test -race ./internal/loot ./internal/town ./internal/bot pass on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof is drill D1/P7 (not run here)"
merge:
result:
live: "drill D1 rerun 2026-10-31 (docs/probes/drills/2026-10-31-d1rerun.txt): P7 PASS (pickit reports picked 2, belt holds both, no left-the-ground-without-arriving), P7a PASS, P1 PASS (hp columns hold hp, mp columns mp, fill moved 8 of 8), P1s PASS, P2 PASS, P3 PASS; each run twice, identical"
```

## 1. Work

- [x] 1.1 B1/B2: belt.Place and WillBelt, Want and candidates use them
- [x] 1.2 B4: pickBeltBox is belt.Place
- [x] 1.3 B5: panels closed before a ground click (already in PickUp and pickOne), documented
- [x] 1.4 arrival by class count, scenario tests, spec delta, docs/fixpickup.md
