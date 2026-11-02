# Tasks

```yaml
change: levelsareas
status: in-progress
agent: pa-levelsareas
branch: agent/levelsareas
worktree: d2lod-wt/levelsareas
owns:
  - docs/research/levelsareas-trace.md
  - tools/emu/lv_*.py
  - openspec/changes/levelsareas/**
skip_specs: true
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0 hits; coverage.py --write"
merge:
result: "Emulator run (lv_area.py, 1,088 cases equal) and table census; not verified live."
credit:
  author: owner
  agents: [Claude Code]
```

## 1. Work

- [x] 1.1 Levels table census (lv_tbl.py)
- [x] 1.2 FUN_0061dca0 vs Levels columns (lv_area.py)
- [x] 1.3 trace doc, ledger row
- [ ] 1.4 density / unique count differential (D3, not started)
