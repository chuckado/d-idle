# Tasks

```yaml
change: skabrecheck
status: archived
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
result: "offline emulator differentials on the worker host; local PR candidate for the physical roll NOT ready (needs independent Opus verifier)"
live: "docs and offline emulator only, no game"
merge: b94cfc1
```

- [x] claim 1 tables and differential
- [x] claim 2 read and differential
- [x] trace doc
