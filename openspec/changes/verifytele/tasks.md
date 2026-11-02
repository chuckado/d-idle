# Tasks

```yaml
change: verifytele
status: in-progress
agent: verifytele
branch: agent/verifytele
worktree: ~/code/d2lod-wt/verifytele
skip_specs: true
owns:
  - openspec/changes/verifytele/**
  - docs/research/verify-teleport-room.md
  - tools/emu/vt_run.sh
  - tools/emu/vt_check.py
exclusive: []
gates:
  - none
result: C1 to C4 hold (20000 cases, seed 4242, V run on the worker host); C5 (Teleport 2 line test) read only; not verified live
```

- [x] 1 Read FUN_005ca360, FUN_00554ea0, FUN_00463740 before the author's trace
- [x] 2 Run the native oracle on a fresh seed and classify
- [x] 3 Write the verdict doc
