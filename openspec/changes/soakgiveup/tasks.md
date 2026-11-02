# Tasks

```yaml
change: soakgiveup
status: in-progress
agent: soakgiveup
branch: agent/soakgiveup
worktree: ~/code/d2lod-wt/soakgiveup
skip_specs: true
owns:
  # code paths (goal.go, emptyarea_test.go) are added when an agent takes the card and soakstall is archived
  - openspec/changes/soakgiveup/**
exclusive: []
gates:
  - go test -p 1 ./internal/bot passes, including the new clear-run case
result: implemented: exploreGaveUp guards on the run kind (keepRoaming) plus a ping-pong guard (noteArea); TestExploreGaveUp, TestPingPongGuard; offline only, not verified live
```

- [x] 1 Test first: a LevelRun{Kind: RunClear, Area: a} gives up after two failed explores in a with no kill (fails on main), a RunQuest or RunBoss run in a does not
- [x] 2 Change the guard in exploreGaveUp to the run kind; gofmt, go vet and go test -p 1 ./internal/bot
- [x] 3 Live drill checklist from docs/soakstall.md, run on the Gate 1 rerun: `area abandoned` for the run area within two roam timeouts
