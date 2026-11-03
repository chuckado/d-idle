# Tasks

```yaml
change: hirereader
status: in-progress
agent: hirereader
branch: agent/hirereader
worktree: ~/code/d2lod-wt/hirereader
skip_specs: true
owns:
  - docs/research/hirereader-trace.md
  - docs/research/hirereader-design.md
  - tools/emu/hrd_pay.py
  - tools/emu/hrd_d2d_test.cpp
  - tools/emu/hrd_build.sh
  - tools/emu/hrd_run.sh
  - openspec/changes/hirereader/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
  - "a V run claim needs counts: purse differential 200,000 cases, 0 differ"
result: "Docs and emulator tooling only, nothing run live, no d2d slot, no Go change. Purse differential (game.exe FUN_00576d90 and the revive chain FUN_006637b0 + FUN_00576d90 vs rules::hire and rules::resurrect_merc): 200,000 cases (8 runs of 25,000), 0 differ (V run). Bot findings B1..B5 (B2: a dead merc whose corpse is gone reads as not owned, so the revive never fires on d2d), no PR candidate, no ledger row (needs a new denominator line, proposed in the trace). Design: bot-side fixes first, then four keys in dump merc."
live: "docs only, nothing run live"
```

- [x] read the fork and bot sides (dump merc, state, menu, hire and resurrect commands)
- [x] purse differential (hrd_*): 200,000 cases, 0 differ
- [x] trace and design docs
- [x] card, OWNERS, merge lines, checks
