# Tasks

```yaml
change: hastpdump
status: archived
agent: hastpdump
branch: agent/hastpdump
worktree: ~/code/d2lod-wt/hastpdump
skip_specs: true
owns:
  - internal/bot/tpdump.go
  - docs/hastpdump.md
exclusive: []
gates:
  - "TestInvTPDump: dump on first look, area change and verdict change, not otherwise"
result: "offline test TestInvTPDump passes; not verified live"
merge: c4e701e7
```

- [x] dump from the HasTP inputs, rate limited
- [x] offline test (area change, verdict change, steady)
- [x] docs/hastpdump.md drill
