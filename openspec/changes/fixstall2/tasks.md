# Tasks

```yaml
change: fixstall2
status: in-progress
agent: fixstall2
branch: agent/fixstall2
worktree: ~/code/d2lod-wt/fixstall2
skip_specs: true
owns:
  - internal/d2dctl/town_verbs.go
  - internal/d2dctl/approach_path.go
  - internal/d2dctl/approach_path_test.go
  - internal/d2dtown/adapter.go
  - internal/town/trip.go
  - internal/town/enterportal_test.go
  - docs/fixstall2.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof after merge: H.pause x3 tries and D6.d x3 (docs/fixstall2.md)"
merge:
result: ""
live: ""
```

- [x] F7: grid route in ApproachNPC (planAround) with side-step fallback
- [x] F8: enterPortal skips GoTo when Interact walks (WalkingInteractor)
- [x] offline tests
- [ ] docs, INDEX, CHANGELOG, BACKLOG
