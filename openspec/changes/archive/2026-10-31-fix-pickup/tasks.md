# Tasks

```yaml
change: fix-pickup
status: archived
agent: fixpickup
branch: agent/fixpickup
worktree: ~/code/d2lod-wt/fixpickup
owns:
  - internal/town/sim/sim.go
  - internal/loot/export_test.go
  - docs/fixpickup.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... plus go test -race ./internal/loot ./internal/town ./internal/bot pass on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof is drill D1/P7 (not run here)"
merge: e4bb88d
result: "verified live 2026-10-31 DRILLS5; see the live: line"
# note: owns of internal/town/belt/belt.go, internal/town/belt/belt_test.go, internal/town/beltfill.go, internal/loot/fixpickup_test.go released to fix-belt2 (code already merged)
live: "drill D1 rerun 2026-10-31 (docs/probes/drills/2026-10-31-d1rerun.txt): P7 PASS (pickit reports picked 2, belt holds both, no left-the-ground-without-arriving), P7a PASS, P1 PASS (hp columns hold hp, mp columns mp, fill moved 8 of 8), P1s PASS, P2 PASS, P3 PASS; each run twice, identical | DRILLS5 (docs/probes/drills/2026-10-31-drills5-d1.txt): D1.P7b PASS (healing in the 3 columns 0 to 2, both rows, a healing potion on the ground and column 3 empty: no healing potion enters column 3, the pickit leaves it: potion already in the pack), D1.P1 PASS again"
```

## 1. Work

- [x] 1.1 B1/B2: belt.Place and WillBelt, Want and candidates use them
- [x] 1.2 B4: pickBeltBox is belt.Place
- [x] 1.3 B5: panels closed before a ground click (already in PickUp and pickOne), documented
- [x] 1.4 arrival by class count, scenario tests, spec delta, docs/fixpickup.md
