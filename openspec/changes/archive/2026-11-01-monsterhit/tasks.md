# Tasks

```yaml
change: monsterhit
status: archived
agent: monsterhit
branch: agent/monsterhit
worktree: ~/code/d2lod-wt/monsterhit
skip_specs: true
owns:
  - docs/research/monsterhit-trace.md
  - tools/emu/mh_rows.py
  - openspec/changes/monsterhit/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "docs and emulator tools only; credits runner; player-defender damage rules not run"
live: "docs only, nothing run live"
merge: abed092
```

- [x] read the earlier melee and missile traces and the d2d source (combat.hpp, monsters.hpp, ai.cpp)
- [x] bot side read (internal/resource/sim.go, internal/combat/safety.go, threat.go, env tuning)
- [x] trace doc with D1..D5, B1..B5, drill recipe, matrix proposals
- [x] tools/emu/mh_rows.py with selftest
- [ ] oracle run for D1 / D2 / D5 on the worker host (open: needs game.exe and the emulator)
