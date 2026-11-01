# Tasks

```yaml
change: hastpdump
status: in-progress
agent: hastpdump
branch: agent/hastpdump
worktree: ~/code/d2lod-wt/hastpdump
skip_specs: true
owns:
  - internal/bot/tpdump.go
  - docs/hastpdump.md
exclusive: []
gates: []
result: "offline test TestInvTPDump passes; not verified live"
```

- [x] dump from the HasTP inputs, rate limited
- [x] offline test (area change, verdict change, steady)
- [x] docs/hastpdump.md drill
