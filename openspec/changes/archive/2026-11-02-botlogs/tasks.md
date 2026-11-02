# Tasks

```yaml
change: botlogs
status: archived
agent: botlogs
branch: agent/botlogs
worktree: ~/code/d2lod-wt/botlogs
skip_specs: true
owns:
  - docs/botlogs.md
  - internal/bot/botlogs_test.go
  - internal/bot/tpdebounce.go
  - internal/bot/tpdebounce_test.go
  - openspec/changes/botlogs/**
exclusive: []
gates:
  - go test -p 1 ./internal/bot passes
result: "offline table tests only; formats in docs/botlogs.md"
merge: bfbd865
```

- [x] explore-giveup line
- [x] tp-debounce-hold and tp-debounce-flip lines
- [x] kitgate-trip line, kitgate-skip condition test
- [x] table tests, docs/botlogs.md
