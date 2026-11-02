# Tasks

```yaml
change: prbuildskmelee3
status: in-progress
agent: prbuildskmelee3
branch: agent/prbuildskmelee3
worktree: d2lod-wt/prbuildskmelee3
skip_specs: true
owns:
  - docs/research/pr-skmelee3-draft.md
  - tools/emu/sk_melee2.py
  - tools/emu/sk_melee2_test.cpp
  - tools/emu/sm3_probe.py
  - openspec/changes/prbuildskmelee3/**
exclusive:
  - fork-branch:local/skmelee3-v1
gates:
  - "suites test_rules test_game_rules test_monsters test_monster_thinks test_monster_ai test_combat test_game green on the head"
  - "counted differential base vs head over at least 100,000 cases"
result: ""
live: "offline: no game; game.exe's own functions run under the emulator on the worker host"
```

- [ ] 1 Fork branch local/skmelee3-v1 from upstream main: D1 and D2 commits with tests
- [ ] 2 Differential base vs head over >= 100,000 cases per claim
- [ ] 3 Det scenarios on lab builds of base and head (if the tool allows)
- [ ] 4 CI parity (gcc and clang flags, include audit, MSVC read, pr_guard, presubmit --quick)
- [ ] 5 Draft docs/research/pr-skmelee3-draft.md
- [ ] 6 changecheck and leakcheck
