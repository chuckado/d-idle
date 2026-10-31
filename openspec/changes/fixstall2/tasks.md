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
live: "DRILLS7 2026-10-31 (README DRILLS7): FAIL. H.pause x3: 3 of 6 bot tries open Kashya (136 ticks), 3 fail at 718 ticks at (25.9,20.5), identical to DRILLS6; routeAround ran but planAround returns no route from the stall cell (clearance 1), clearance 0 returns one (F10). D6.d x3: run 1 escapes (493 ticks), run 2 stuck 62 s, identical to DRILLS6; enterPortal still pre-walks because hoverBelt does not implement WalkingInteractor (F11)"
```

- [x] F7: grid route in ApproachNPC (planAround) with side-step fallback
- [x] F8: enterPortal skips GoTo when Interact walks (WalkingInteractor)
- [x] offline tests
- [ ] docs, INDEX, CHANGELOG, BACKLOG
