# Tasks

```yaml
change: pingpongrate
status: in-progress
agent: pingpongrate
branch: agent/soakfixes2
worktree: ~/code/d2lod-wt/pingpongrate
skip_specs: true
owns:
  - openspec/changes/pingpongrate/**
  - internal/bot/goal.go
  - internal/bot/emptyarea_test.go
  - docs/soakstall.md
exclusive: []
gates:
  - go test -p 1 ./internal/bot passes, including the cluster C replay
result: implemented and tested offline (mem.Fake rig); not verified live
```

- [x] 1 Test first: cluster C replay fires, A and B fire, a field with real kills does not
- [x] 2 noteArea resets on progress, not on a kill; the knob is listed in docs/knobs.md by the generator
