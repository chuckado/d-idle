# Tasks

```yaml
change: uniqmods
status: archived
agent: uniqmods
branch: agent/uniqmods
worktree: ~/code/d2lod-wt/uniqmods
skip_specs: true
owns:
  - docs/research/uniqmods-trace.md
  - tools/emu/um_*
  - openspec/changes/uniqmods/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "docs and emulator tools only; credits runner; census is game.exe against itself, no d2d side"
live: "docs only, nothing run live"
merge: 2ae7673
```

- [x] MonUMod table and pick weights (tables, disassembly read)
- [x] pack sizes, minion counts, Act 1 superunique records (existing run log)
- [x] multipliers and XP / drop links
- [x] d2d comparison (source read)
- [x] bot findings
- [ ] Act 2 superunique records (needs um_extract.sh on the worker host)
