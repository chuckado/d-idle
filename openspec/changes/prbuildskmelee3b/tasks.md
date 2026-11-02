# Tasks

```yaml
change: prbuildskmelee3b
status: in-progress
agent: prbuildskmelee3b
branch: agent/prbuildskmelee3b
worktree: d2lod-wt/prbuildskmelee3b
skip_specs: true
owns:
  - docs/research/pr-skmelee3-d1-draft.md
  - docs/research/pr-skmelee3-d2-draft.md
  - tools/emu/sk_melee2.py
  - tools/emu/sk_melee2_test.cpp
  - tools/emu/sm3_probe.py
  - tools/emu/sm3_run.sh
  - openspec/changes/prbuildskmelee3b/**
exclusive:
  - fork-branch:local/skmelee3-d1
  - fork-branch:local/skmelee3-d2
gates:
  - "suites test_rules test_game_rules test_monsters test_monster_thinks test_monster_ai test_combat test_game green on each head"
  - "counted differential base vs head over at least 100,000 cases"
result: "Fork branches local/skmelee3-d1-final 46ced89a and local/skmelee3-d2-final 283c7bce (private remote only) on upstream main cb716f10; verifier findings applied; two generators 0 differences vs game.exe on armed blows; 7 suites green on both heads; CI parity 0; PR 174 merge resolution prepared; det hashes not run. Not verified live (no game)."
live: "offline: no game; game.exe's own functions run under the emulator on the worker host"
```

- [x] 1 Fork branches local/skmelee3-d1 and -d2 from upstream main (squashed, noreply identity)
- [x] 2 Verifier findings applied (gate, missile test, comment, stat 122 scope stated)
- [x] 3 Seven suites on each head
- [x] 4 Counted differential base vs head, both harnesses
- [x] 5 CI parity (g++, clang++, tidy, pr_guard, presubmit)
- [ ] 6 Det hashes (not run: needs two lab builds and a slot while the gate runs; checklist in the drafts)
- [x] 7 Throwaway merge with PR 174, resolution prepared
- [x] 8 Drafts, changecheck, leakcheck
