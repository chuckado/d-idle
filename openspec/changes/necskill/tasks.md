# Tasks

```yaml
change: necskill
status: in-progress
agent: necskill
branch: agent/necskill
worktree: ~/code/d2lod-wt/necskill
skip_specs: true
owns:
  - docs/research/nec-skill-trace.md
  - tools/emu/nec_*
  - openspec/changes/necskill/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; coverage.py --check; tools/ci/check.sh"
result: "Offline. nec_calc differential 27,800 cases 0 differ; nec_missile_hit 36,600 cases, 10,509 differ on the real d2d build, 0 with the SKMISS fix; Corpse Explosion and summon base stats read-only or U. Ledger: skills.nec 30 rows, missiles 114/192/193. Not verified live."
live: "docs and emulator only, nothing run live"
```

- [x] calc differential (nec_calc.py, nec_test.cpp)
- [x] magic missile differential (nec_missile_hit.py)
- [x] read Corpse Explosion and the summon do
- [x] trace doc, ledger, gaps, rule-coverage regenerated
- [x] merge-lines
