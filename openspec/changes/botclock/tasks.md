# Tasks

```yaml
change: botclock
status: in-progress
agent: botclock
branch: agent/botclock
worktree: d2lod-wt/botclock
skip_specs: true
owns:
  - internal/gclock/**
  - docs/timescale.md
  - openspec/changes/botclock/**
exclusive: []
gates:
  - "gofmt, go vet and go test of the touched packages; changecheck 0 warnings; leakcheck 0"
result: "offline tests only; not verified live (needs the user's go, i4 and the hook)"
```

- [x] inventory and classification (docs/timescale.md)
- [x] internal/gclock with tests and the guard test
- [x] move game-clock call sites
- [x] docs/timescale.md
- [ ] live k=2 and k=4 re-measure (not part of this card)
