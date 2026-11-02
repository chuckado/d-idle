# Tasks

```yaml
change: skias
status: in-progress
agent: skias
branch: agent/skias
worktree: ~/code/d2lod-wt/skias
skip_specs: true
owns:
  - docs/research/sk-ias-trace.md
  - tools/emu/skias_bp.py
  - openspec/changes/skias/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Partial desk trace: d2d attack_speed / attack_ticks read and breakpoints reproduced by skias_bp.py; game.exe attack branch not read, nothing V run, no ledger row. Not verified live."
live: "docs only, nothing run live"
```

- [x] read d2d attack_speed / attack_ticks, breakpoint script
- [x] trace doc with deviations D1-D4 (U) and bot finding B1
- [ ] game.exe attack branch of 00623f50 under the emulator (open)
