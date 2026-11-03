# Tasks

```yaml
change: seeddist
status: in-progress
agent: seeddist
branch: agent/seeddist
worktree: ~/code/d2lod-wt/seeddist
skip_specs: true
owns:
  - openspec/changes/seeddist/**
  - docs/research/seeddist-trace.md
  - tools/emu/sd_*
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
  - "every V run number has stated cases, seeds and the test used"
result: ""
```

## 1. Work

- [ ] 1.1 Read seed-study, d2d life / drop seeding, game.exe unit seeding
- [ ] 1.2 Game-side sampler (sd_game.py) and d2d-side sampler (sd_d2d_test.cpp)
- [ ] 1.3 Tests (chi-square, KS) per quantity, 8 classes x >= 20,000
- [ ] 1.4 Trace doc, ledger rows for proven rows, merge-lines.txt
