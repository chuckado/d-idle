# Tasks

```yaml
change: fix-belt2
status: in-progress
agent: fixbelt2
branch: agent/fixbelt2
worktree: ~/code/d2lod-wt/fixbelt2
owns:
  - internal/town/belt/belt.go
  - internal/town/belt/belt_test.go
  - internal/town/beltfill.go
  - internal/town/beltfix_test.go
  - internal/loot/loot.go
  - internal/loot/fixpickup_test.go
  - docs/fixbelt2.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... plus go test -race ./internal/loot ./internal/town ./internal/bot pass on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof is a drill rerun (docs/fixbelt2.md), not run here"
result: ""
live: "verified live 2026-10-31 DRILLS5 (docs/probes/drills/2026-10-31-drills5-d1.txt, each twice, identical): D1.P1b PASS (Sorceress sash, pack of 6 healing + 4 mana, empty belt: the belt holds 4 healing and 4 mana, every kind in its planned column, 2 healing stay in the pack, 0 failed moves, no belt: marked full line), D1.P1 PASS unchanged"
```

## 1. Work

- [x] 1.1 log analysis of post-fix soak 1 (docs/fixbelt2.md)
- [x] 1.2 no belt column for a kind is no room, never the sticky full mark
- [x] 1.3 belt.PlaceFor: healing may not open the last empty column kept for a planned kind with no column; fill candidates and ground pickups use it
- [x] 1.4 scenario tests, spec delta
- [ ] 1.5 live proof: drill rerun by another agent
