# Tasks

```yaml
change: fixstall3
status: in-progress
agent: fixstall3
branch: agent/fixstall3
worktree: ~/code/d2lod-wt/fixstall3
skip_specs: true
owns:
  - docs/fixstall3.md
  - internal/d2dctl/testdata/camp-grid.txt
  - internal/d2dtown/wiring_test.go
  - internal/combat/fight.go
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof after merge: H.pause x3 (6 of 6 open), D6.d x3, D-sync waypoint x3 (docs/fixstall3.md)"
merge:
result:
```

## 1. Work

- [x] 1.1 F10: adaptive clearance in planAround, test on the real camp grid (TestPlanAroundRealCampStallCells)
- [x] 1.2 F11: d2dtown.Input in bot and drills, TestInputIsWalkingInteractor
- [x] 1.3 F12: diagnosis (UNVERIFIED cause), geometry in the ignore reason
- [x] 1.4 docs, INDEX, CHANGELOG, BACKLOG
