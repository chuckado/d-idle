# Tasks

```yaml
change: skbarb
status: in-progress
agent: skbarb
branch: agent/skbarb
worktree: ~/code/d2lod-wt/skbarb
skip_specs: true
owns:
  - docs/research/skbarb-trace.md
  - tools/emu/skbarb_*
  - openspec/changes/skbarb/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; coverage.py --check; tools/ci/check.sh"
result: "docs and emulator tools only; no PR candidate; coverage ledger updated (skills.bar calc and physical aspect, 27,800 cases)"
live: "docs and emulator only, nothing run live"
```

- [x] calc and skill-physical differential (skbarb_calc.py, skbarb_test.cpp, skbarb_run.sh)
- [x] Whirlwind target search, Howl readers, war-cry ring flights (skbarb_ww.py, skbarb_howl.py, skbarb_ring.py)
- [x] read the whirl timer, Frenzy, Howl hit, do 68
- [x] trace doc, ledger, rule-coverage regenerated
- [x] merge-lines
