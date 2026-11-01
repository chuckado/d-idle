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
result: "Offline, nothing verified live. Frost Nova ring traced (docs/research/sk-nova-trace.md): FUN_005c9b50 -> FUN_0056d400 makes 64 frostnova from the caster aimed at the truncated tables 0x6e1288 / 0x6e1388, calc1 to the velocity, reach 12.0-12.7 subtiles (V read + V run, tools/emu/sk_nova_ring.py, 896 real flights, reading 0 differ). D1 d2d strikes a monster with every touching ring missile (model 6.2x hits per cast, 8-35x point blank), D2 even angles (small), D3 calc1 to range (Howl only); no PR candidate (no stateful harness); bot B1-B6; drills N1-N6."
```

## 1. Work

- [x] 1.1 Trace the Frost Nova ring in game.exe (do 22, FUN_0056d400, tables, FUN_0059fa30 params)
- [x] 1.2 Run the 64 flights natively and compare with the reading and with d2d's aim
- [x] 1.3 Read d2d's nova (fire, skill_missile_hits, the tick loop) and the NextHit rule
- [x] 1.4 Bot use (novaTick, ranges, mana gate) and the Static Field radius unit
- [x] 1.5 Trace doc with deviations, bot findings, drill recipe, matrix cells, shared-file lines
