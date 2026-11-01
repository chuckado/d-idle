# Tasks

```yaml
change: skpass
status: open
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
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: ""
```

## 1. Work

- [ ] 1.1 Dump the patched Skills.txt synergy and passive columns (D)
- [ ] 1.2 Trace passive stat application, mastery reads, Warmth regen, Telekinesis do, blvl vs lvl (V read)
- [ ] 1.3 Read d2d's and the bot's versions; deviations and findings
- [ ] 1.4 Synergy bonus per plan band (I computed)
- [ ] 1.5 Trace doc with drill recipe, matrix cells, shared-file lines
