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
live: "DRILLS7 2026-10-31 (README DRILLS7): H.pause FAIL (3 of 6 bot tries open, 3 fail at 718 ticks: planAround has no clearance-1 route from the stall cell, F10), D6.d FAIL x3 identical to DRILLS6 (F11, walking entry not wired). DRILLS6 2026-10-31 (docs/probes/drills/README.md DRILLS6): H.pause PARTIAL (1 of 2 bot tries opens Kashya in 136 ticks, the other fails at 718 ticks, stall near 25.9,20.5; identical in two invocations), D6.d PARTIAL (escape log line seen both runs, run 1 ends ErrLevelChanged in 493 ticks, run 2 cannot walk into the portal: 62 s stuck)"
```

- [x] 1.1 ApproachNPC: early stall detection, distance-sized game-time budget
- [x] 1.2 nav: pocket escape (town portal) in plain walks, shorter ladder when nothing moves
- [x] 1.3 tests and docs/fixstall.md
- [x] 1.4 live proof: H.pause and D6.d rerun (DRILLS6: PARTIAL, see live:)
