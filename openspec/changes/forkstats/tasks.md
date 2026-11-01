# Tasks

```yaml
change: forkstats
status: in-progress
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
result: "Fork branch lab2-stats pushed (d68ff038), binary d2d.lab2-stats-d68ff038-linuxbuild built on the NAS builder (319 s, tests rc 0). Verified live on one hidden slot: dump stats shows gear and weapons for a fresh Sorceress. d2d has no player stats 17/18/21/22; gear is the item-prop sums, see docs/forkdbg.md."
```

## 1. Work

- [x] 1.1 Fork branch lab2-stats pushed
- [x] 1.2 Built and tested on the NAS builder
- [x] 1.3 Live read (or the exact commands)
- [x] 1.4 docs/forkdbg.md section, OWNERS
