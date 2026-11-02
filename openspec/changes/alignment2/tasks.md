# Tasks

```yaml
change: alignment2
status: in-progress
agent: alignment2
branch: agent/alignment2
worktree: ~/code/d2lod-wt/alignment2
skip_specs: true
owns:
  - docs/research/alignment2-trace.md
  - tools/emu/al2_run.py
  - tools/emu/al2_ctx.py
  - tools/emu/al2_peek.py
  - openspec/changes/alignment2/**
exclusive: []
gates:
  - "game.exe native runs on the worker host: counts in the trace (734 rows, 18, 168, 288 cases; 0 differ)"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Offline, nothing verified live. FUN_006259b0 = alignment: player / merc / Shadow / skeletons / golems / revived / Valkyrie / Hydra / druid summons / sentries / converted all 2 (setter FUN_005543b0, spawn map of MonStats Align 1 -> 2); the undead block gate is player or alignment 2 (V run 288 cases 0 differ), so d2d's not_player exclusion of merc and Shadow is wrong (S d2d side, not re-read). No PR candidate, no ledger row."
```

## 1. Work

- [x] 1.1 Read FUN_006259b0, its setter and callers in game.exe
- [x] 1.2 Native runs (Align byte, setter/getter, getter shapes, gate)
- [x] 1.3 Trace doc with table, rule, bot findings, drill recipe
- [x] 1.4 merge-lines.txt

## 2. Open

- [ ] 2.1 Live read of stat 0xac on a hired merc and a Shadow (drill in the trace)
