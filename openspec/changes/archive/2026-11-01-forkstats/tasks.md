# Tasks

```yaml
change: forkstats
status: archived
agent: forkstats
branch: agent/forkstats
worktree: ~/code/d2lod-wt/forkstats
skip_specs: true
owns:
  - openspec/changes/forkstats/**
  - docs/forkdbg.md
exclusive: []   # the fork branch lab2-stats is new and written only by this card; no slot, no socket, no lab binary overwritten
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "fork branch lab2-stats built on the builder, test passes, one live read: a clvl 1 Sorceress has only the staff base 1-5; d2d has no player stats 17/18/21/22 (fight code adds item props itself)"
merge: 120e534
```

## 1. Work

- [x] 1.1 Fork branch lab2-stats pushed
- [x] 1.2 Built and tested on the NAS builder
- [x] 1.3 Live read (or the exact commands)
- [x] 1.4 docs/forkdbg.md section, OWNERS
