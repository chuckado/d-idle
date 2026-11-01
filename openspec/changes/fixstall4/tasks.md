# Tasks

```yaml
change: fixstall4
status: in-progress
agent: fixstall4
branch: agent/fixstall4
worktree: ~/code/d2lod-wt/fixstall4
skip_specs: true
owns:
  - docs/fixstall4.md
  - internal/d2dctl/approach_path.go
  - internal/d2dctl/approach_path_test.go
  - internal/d2dctl/town_verbs.go
  - internal/d2dctl/fixstall4_test.go
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof after merge: H.pause x3 invocations open Kashya in 6 of 6 first-run tries and 6 of 6 second-run tries (docs/fixstall4.md)"
merge:
result: offline only, not verified live (see docs/fixstall4.md)
```

## 1. Work

- [x] 1.1 F13 cause (a): route legs cut the camp block corner (reproduced offline on the real grid, old code stops at 25.4,22.9 like the live t483)
- [x] 1.2 F13 cause (b): the first route starts late and stalls waste the budget; routing at every stall, mid-round stall detection
- [x] 1.3 F13 (c): final position in the error text; budget from the route length
- [x] 1.4 docs, INDEX, CHANGELOG, BACKLOG, primitives note
