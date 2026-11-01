# Tasks

```yaml
change: skarmor
status: in-progress
agent: skarmor
branch: agent/skarmor
worktree: ~/code/d2lod-wt/skarmor
skip_specs: true
owns:
  - docs/research/sk-armor-trace.md
  - tools/emu/sk_armor.py
  - openspec/changes/skarmor/**
exclusive: []
gates:
  - "sk_armor.py game vs direct (freeze_rules.py): 0 mismatches on every case"
  - "go run ./cmd/changecheck: 0 warnings for this card; go run ./cmd/leakcheck -mode fail: 0"
result: ""
```

## 1. Work

- [ ] 1.1 Trace do 18 / do 23, state creation and group replacement, aura events 1 / 2 / 3 / 24 in game.exe (tags)
- [ ] 1.2 Read d2d's start_state, buff_events, return_fire, absorb; the bot's buffDue / buffLimit / Buff
- [ ] 1.3 Game-side oracle for Frozen Armor's event 2 (sk_armor.py), run natively
- [ ] 1.4 Trace doc: Q, deviations, bot findings, PR decision, drill recipe, matrix cells, shared-file lines
- [ ] 1.5 changecheck, leakcheck

## 2. Open

- [ ] 2.1 Live drill (recipe in the trace; needs the lab binary and a hidden slot)
- [ ] 2.2 d2d-side test mode for buff_events (stateful; described in the trace)
