# Tasks

```yaml
change: verifyskmana
status: in-progress
agent: verifyskmana
branch: agent/verifyskmana
worktree: ~/code/d2lod-wt/verifyskmana
skip_specs: true
owns:
  - openspec/changes/verifyskmana/**
  - docs/research/verify-skmana.md
  - tools/emu/vsm_*
exclusive: []
gates:
  - tools/ci/check.sh
result: C1 C2 C3 C5 C6 hold (20,000 cases, seed 20261102, V run on the worker host), C4 not tested; no live game
```

- [x] 1 Read FUN_0056c160 and FUN_0056bfe0 and write an own oracle
- [x] 2 Run own harness vsm_diff.py against game.exe
- [x] 3 Write docs/research/verify-skmana.md
