# Tasks

```yaml
change: skstateev
status: in-progress
agent: skstateev
branch: agent/skstateev
worktree: ~/code/d2lod-wt/skstateev
skip_specs: true
owns:
  - docs/research/sk-state-events-trace.md
  - tools/emu/sk_stateev_xref.py
  - openspec/changes/skstateev/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Docs only, static (V read): FUN_005d3ac0 is an attacker hit bonus; dispatcher FUN_005c0c30; event 3 fires on a miss too (FUN_0057d4f0); event 11 before the resistance pass (FUN_0057c1e0). No game code was run; not verified live."
live: "docs only, nothing run live"
```

- [x] find the dispatcher and its call sites
- [x] answer hit/miss and absorb order (V read)
- [ ] native run of FUN_0057d4f0 / FUN_0057c1e0 (open, U)
