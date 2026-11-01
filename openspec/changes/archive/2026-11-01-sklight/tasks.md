# Tasks

```yaml
change: sklight
status: archived
merge: ea3e136
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
result: "Offline, not verified live (docs and a Monte Carlo model only), no V run. game.exe read: Chain Lightning do 26 / hit 12 / hop search, per-frame collision, NextHit as unit state justhit; D1 (justhit missing in d2d) and D2 (one unit a frame) read on both sides, magnitude by a model (sk_light_rehit.py, I); D3 existing held candidate; bot B1-B5; no PR candidate (no flight harness)."
```

## 1. Work

- [x] 1.1 Read game.exe: do 26, hit 12, hop search, collision per frame, NextHit, to-hit gate
- [x] 1.2 d2d's version and the deviations
- [x] 1.3 The bot's use: order, ranges, cast gap, mana, immune fallback
- [x] 1.4 Model of the two hit rules (tools/emu/sk_light_rehit.py)
- [x] 1.5 Trace doc: drill recipe, matrix cells, shared-file lines
