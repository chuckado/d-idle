# Tasks

```yaml
change: fixstall
status: in-progress
agent: fixstall
branch: agent/fixstall
worktree: ~/code/d2lod-wt/fixstall
skip_specs: true
owns:
  - internal/d2dctl/fixstall_test.go
  - internal/nav/pinned_test.go
  - docs/fixstall.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof after merge: rerun H.pause and D6.d (docs/fixstall.md)"
merge:
result: "offline only; early stall detection and distance-sized budget for the NPC approach, pocket escape in plain walks; not verified live (awaits H.pause and D6.d reruns, so the card stays open at merge)"
live: ""
```

- [x] 1.1 ApproachNPC: early stall detection, distance-sized game-time budget
- [x] 1.2 nav: pocket escape (town portal) in plain walks, shorter ladder when nothing moves
- [x] 1.3 tests and docs/fixstall.md
- [ ] 1.4 live proof: H.pause and D6.d rerun
