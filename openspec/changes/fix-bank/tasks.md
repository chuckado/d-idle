# Tasks

```yaml
change: fix-bank
status: in-progress
agent: fixbank
branch: agent/fixbank
worktree: ~/code/d2lod-wt/fixbank
owns:
  - internal/d2dtown/adapter.go
  - internal/town/fixbank_test.go
  - docs/fixbank.md
exclusive: []
gates:
  - "go test -p 1 ./internal/town ./internal/d2dtown passes including TestFixBank*"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: "verified live 2026-10-31 DRILLS5 (docs/probes/drills/2026-10-31-drills5-d3.txt, docs/probes/drills/2026-10-31-drills5-repair.txt, lab binary cheats3, each drill twice, identical): D3.s PASS (purse 1500 deposited whole, 300 withdrawn: purse 300 stash 1200), D3.s2 PASS (1000 of 1500 typed into the prompt: purse 500 stash 1000), D3.r0 PASS (gold 0: repair skipped, 0 ticks, no smith opened) and D3.r1 PASS (gold 5: skipped, gold and durability unchanged), D3.tp PASS unchanged (tome and 5 scrolls, 800 gold = the report)"
```

- [x] goldMover capability, d2d implementation, typed path kept
- [x] needRepair requires gold for the estimate
- [x] D3.tp analysed (policy, no change)
- [x] scenario tests and spec delta
- [ ] live proof: drill rerun (m1-liveproof)
