# Tasks

```yaml
change: uniqmods
status: in-progress
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
result: "Docs and offline tooling: docs/research/uniqmods-trace.md (Q1-Q9, D1-D4 read-only, B1-B7, drill G1-G6). No new emulator run (the runner refused remote scripts): Normal numbers are the DECOMPMODS run log re-summarised (S-run), pick rules by disassembly read and tables. Act 2 superunique records not extracted (U). No PR candidate, no coverage ledger row. Not verified live."
live: "docs only, nothing run live"
```

- [x] MonUMod table and pick weights (tables, disassembly read)
- [x] pack sizes, minion counts, Act 1 superunique records (existing run log)
- [x] multipliers and XP / drop links
- [x] d2d comparison (source read)
- [x] bot findings
- [ ] Act 2 superunique records (needs um_extract.sh on the worker host)
