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
live: "not verified live: rerun drills D3.s, D3.s2, D3.r0, D3.tp (docs/fixbank.md)"
```

- [x] goldMover capability, d2d implementation, typed path kept
- [x] needRepair requires gold for the estimate
- [x] D3.tp analysed (policy, no change)
- [x] scenario tests and spec delta
- [ ] live proof: drill rerun (m1-liveproof)
