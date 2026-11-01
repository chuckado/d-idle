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
result: ""
live: "docs and offline emulator only, no game"
```

- [ ] claim 1 tables and differential
- [ ] claim 2 read and differential
- [ ] trace doc
