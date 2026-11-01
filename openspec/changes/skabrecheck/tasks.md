# Tasks

```yaml
change: skabrecheck
status: in-progress
agent: skabrecheck
branch: agent/skabrecheck
worktree: ~/code/d2lod-wt/skabrecheck
skip_specs: true
owns:
  - docs/research/skabrecheck.md
  - tools/emu/skab_*
  - openspec/changes/skabrecheck/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "docs/research/skabrecheck.md: claim 1 REFUTED (data) and equal on the missile path (V run, 0 differ); melee pierce UNCOVERED; claim 2 CONFIRMED (V run 84,271 of 101,626), local candidate branch in the fork checkout not pushed. Offline emulator only, nothing live."
live: "docs and offline emulator only, no game"
```

- [x] claim 1 tables and differential
- [x] claim 2 read and differential
- [x] trace doc
