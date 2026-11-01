# Tasks

```yaml
change: soakstall
status: in-progress
agent: soakstall
branch: agent/soakstall
worktree: ~/code/d2lod-wt/soakstall
skip_specs: true
owns:
  - docs/soakstall.md
  - internal/bot/goal.go
  - internal/bot/runner.go
  - internal/bot/emptyarea_test.go
  - openspec/changes/soakstall/**
exclusive: []
gates:
  - TestExploreGaveUp (internal/bot) passes
result: offline analysis of the run logs plus a unit test (TestExploreGaveUp, go test -p 1 ./internal/bot ok); NOT verified live.
```

- [x] 1 Timeline and root cause from the run logs
- [x] 2 Fix with table test
- [x] 3 docs/soakstall.md with the live drill
