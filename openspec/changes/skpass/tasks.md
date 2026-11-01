# Tasks

```yaml
change: skpass
status: in-progress
agent: skpass
branch: agent/skpass
worktree: ~/code/d2lod-wt/skpass
skip_specs: true
owns:
  - docs/research/sk-passives-trace.md
  - tools/emu/sk_pass_*
  - openspec/changes/skpass/**
exclusive: []
gates:
  - "every claim tagged V read / V run / D / S / I / U with addresses; synergies read from the patched Skills.txt, not recited"
  - "changecheck 0 warnings for this card (status in-progress until the merger archives it); leakcheck -mode fail 0"
result: "Offline, nothing verified live (no game, no d2d slot). docs/research/sk-passives-trace.md: passives FUN_00646d60 / 00646f20, masteries (Cold Mastery = pierce, FUN_0057be00 gate res < 100, floor -100), Warmth regen, Telekinesis do FUN_005c98f0, +skills FUN_00644180 / 006442a0 (synergies blvl, rest lvl) V read; synergy table D; plan replay I. d2d D1 Telekinesis knockback always (game.exe 35 %), D2 no item / object branch, D3 entry +0x2c; no PR candidate (no existing differential). Bot B1 leveler spends the clvl 40 Lightning row at 25-27, B2 Effective() = base + stat 107 only. Drill recipe P1-P6."
```

## 1. Work

- [x] 1.1 Dump the patched Skills.txt synergy and passive columns (D)
- [x] 1.2 Trace passive stat application, mastery reads, Warmth regen, Telekinesis do, blvl vs lvl (V read)
- [x] 1.3 Read d2d's and the bot's versions; deviations and findings
- [x] 1.4 Synergy bonus per plan band (I computed)
- [x] 1.5 Trace doc with drill recipe, matrix cells, shared-file lines
