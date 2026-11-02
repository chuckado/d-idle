# Tasks

```yaml
change: monsterhit
status: in-progress
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
result: "docs and one table helper only; no oracle run (no game.exe or emulator on the box, worker host not reachable): nothing V run, no ledger row, no PR candidate; section 8 of the trace is the recipe to run on the worker host"
live: "docs only, nothing run live"
```

- [x] read the earlier melee and missile traces and the d2d source (combat.hpp, monsters.hpp, ai.cpp)
- [x] bot side read (internal/resource/sim.go, internal/combat/safety.go, threat.go, env tuning)
- [x] trace doc with D1..D5, B1..B5, drill recipe, matrix proposals
- [x] tools/emu/mh_rows.py with selftest
- [ ] oracle run for D1 / D2 / D5 on the worker host (open: needs game.exe and the emulator)
