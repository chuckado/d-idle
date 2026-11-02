# Tasks

```yaml
change: necskill
status: archived
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
result: "docs and emulator tools only; no new PR candidate; coverage ledger updated"
live: "docs and emulator only, nothing run live"
merge: f87ba9c
```

- [x] calc differential (nec_calc.py, nec_test.cpp)
- [x] magic missile differential (nec_missile_hit.py)
- [x] read Corpse Explosion and the summon do
- [x] trace doc, ledger, gaps, rule-coverage regenerated
- [x] merge-lines
