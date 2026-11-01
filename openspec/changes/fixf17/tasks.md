# Tasks

```yaml
change: fixf17
status: in-progress
agent: fixf17
branch: agent/fixf17
worktree: ~/code/d2lod-wt/fixf17
skip_specs: true
owns:
  - docs/fixf17.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof after merge: the full F group x3 passes DW.m1 in run 1 and no `the Pacer lagged` line above 100 s (docs/fixf17.md)"
result: "offline: cause found in the harness from the DRILLS9 logs (Pacer lag), fixed with a regression test; not verified live"
```

## 1. Work

- [x] 1.1 timeline of the failing DW.m1 run from the DRILLS9 logs
- [x] 1.2 hypotheses ranked, the Pacer lag shown offline (test fails before, passes after)
- [x] 1.3 harness fix (Tick re-syncs the Pacer, waypoint trip syncs and logs the lag)
- [x] 1.4 docs, INDEX, CHANGELOG, BACKLOG, primitives note
- [ ] 1.5 live proof (full F group x3)
