# Tasks

```yaml
change: pingpongrate
status: proposed
agent: pingpongrate
branch: agent/pingpongrate
worktree: ~/code/d2lod-wt/pingpongrate
skip_specs: true
owns:
  # code paths (goal.go, emptyarea_test.go) are added when an agent takes the card
  - openspec/changes/pingpongrate/**
exclusive: []
gates:
  - go test -p 1 ./internal/bot passes, including the cluster C replay
result: not started
```

- [ ] 1 Test first: cluster C replay fires, A and B fire, a field with real kills does not
- [ ] 2 noteArea resets on progress, not on a kill; the knob is listed in docs/knobs.md by the generator
