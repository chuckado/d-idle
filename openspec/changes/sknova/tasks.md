# Tasks

```yaml
change: sknova
status: in-progress
agent: sknova
branch: agent/sknova
worktree: ~/code/d2lod-wt/sknova
skip_specs: true
owns:
  - docs/research/sk-nova-trace.md
  - tools/emu/sk_nova_ring.py
  - openspec/changes/sknova/**
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
  - "any PR claim only with the existing differential harness: before > 0, after 0"
result: "pending"
```

## 1. Work

- [ ] 1.1 Trace the Frost Nova ring in game.exe (do 22, FUN_0056d400, tables, FUN_0059fa30 params)
- [ ] 1.2 Run the 64 flights natively and compare with the reading and with d2d's aim
- [ ] 1.3 Read d2d's nova (fire, skill_missile_hits, the tick loop) and the NextHit rule
- [ ] 1.4 Bot use (novaTick, ranges, mana gate) and the Static Field radius unit
- [ ] 1.5 Trace doc with deviations, bot findings, drill recipe, matrix cells, shared-file lines
