# Tasks

```yaml
change: sklight
status: open
agent: sklight
branch: agent/sklight
worktree: ~/code/d2lod-wt/sklight
skip_specs: true
owns:
  - docs/research/sk-lightning-trace.md
  - tools/emu/sk_light_rehit.py
  - openspec/changes/sklight/**
exclusive: []
gates:
  - "every claim in the trace tagged V read / V run / D / S / I / U; dead ends written down"
  - "PR candidate only with the existing differential harness (mismatches before > 0, after 0), local only"
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: ""
```

## 1. Work

- [ ] 1.1 Read game.exe: do 26, hit 12, hop search, collision per frame, NextHit, to-hit gate
- [ ] 1.2 d2d's version and the deviations
- [ ] 1.3 The bot's use: order, ranges, cast gap, mana, immune fallback
- [ ] 1.4 Model of the two hit rules (tools/emu/sk_light_rehit.py)
- [ ] 1.5 Trace doc: drill recipe, matrix cells, shared-file lines
