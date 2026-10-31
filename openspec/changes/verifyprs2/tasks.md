# Tasks

```yaml
change: verifyprs2
status: in-progress
agent: verifyprs2
branch: agent/verifyprs2
worktree: ~/code/d2lod-wt/verifyprs2
skip_specs: true
owns:
  - tools/emu/verify_walk_mercpet.py
  - tools/emu/verify_walk_mercpet_mode.inc
  - tools/emu/verify_price_followon.py
  - tools/emu/verify_price_followon_mode.inc
  - docs/research/pr-walk-proof.md
  - docs/research/verify-price2.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - openspec/changes/verifyprs2/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result:
```

## 1. Walk velocity v2 (fix/walk-velocity-v2)

- [ ] 1.1 Replay the six commits onto current upstream main; re-derive the merc and pet movement from the disassembly
- [ ] 1.2 Confirm `test_game --walk-velocity` calls the real d2d functions
- [ ] 1.3 Own generator and harness (merc types, difficulties, pet kinds, edge values), >= 100k cases, real d2d before and after
- [ ] 1.4 Sweep d2d and game.exe for other paths; pr_guard, upstream tests, presubmit --quick, MSVC-class sweep, clang-tidy names and includes
- [ ] 1.5 Verdict and limits in pr-walk-proof.md

## 2. Vendor prices follow-on (fix/vendor-prices-followon)

- [ ] 2.1 Re-derive D9 (multipliers over 0xffff, quest loop) and D10 (FUN_0062e660 repair gate) from the disassembly
- [ ] 2.2 Own generator and harness, >= 100k cases including the quest multipliers, real d2d before and after
- [ ] 2.3 Sweep d2d and game.exe for other price scalers and repair-gate callers; rebase recipe onto upstream main
- [ ] 2.4 pr_guard, upstream tests, presubmit --quick, MSVC-class sweep, clang-tidy names and includes
- [ ] 2.5 Verdict and limits in verify-price2.md

## 3. Close

- [ ] 3.1 pr-bundle-queue.md rows, BACKLOG line, OWNERS, changecheck, leakcheck, openspec validate
- [ ] 3.2 Remove the throwaway worktrees and branches on the worker host; agentreap empty
