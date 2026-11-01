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
result: "offline verification (no live game, no d2d slot): fix/walk-velocity-v2 merc and pet claims HOLD WITH LIMITS, 123,904 own cases, 89,544 -> 0 through the real merc_turn / pet_walking (upstream 72132b99 before, each commit, head 10338345), NOT READY (pet follow paths pets.cpp:246 and fight/tick.cpp:446 at RunVelocity x 1.1 need a ponytail or body line; det hashes); fix/vendor-prices-followon D9 and D10 HOLD WITH LIMITS, 104,422 own cases through the real item_price / repair_cost, D9 21,259 -> 0 of 47,156 incl. the quest multipliers, D10 2,235 -> 0 of 41,663, NOT READY (price_claims.py:41 B905 fails presubmit ruff; waits for #152, rebase recipe simulated clean); pr_guard 0/0, upstream test binaries rc 0, gcc shadow and clang-tidy names/includes 0 on added lines for both"
```

## 1. Walk velocity v2 (fix/walk-velocity-v2)

- [x] 1.1 Replay the six commits onto current upstream main; re-derive the merc and pet movement from the disassembly
- [x] 1.2 Confirm `test_game --walk-velocity` calls the real d2d functions
- [x] 1.3 Own generator and harness (merc types, difficulties, pet kinds, edge values), >= 100k cases, real d2d before and after
- [x] 1.4 Sweep d2d and game.exe for other paths; pr_guard, upstream tests, presubmit --quick, MSVC-class sweep, clang-tidy names and includes
- [x] 1.5 Verdict and limits in pr-walk-proof.md

## 2. Vendor prices follow-on (fix/vendor-prices-followon)

- [x] 2.1 Re-derive D9 (multipliers over 0xffff, quest loop) and D10 (FUN_0062e660 repair gate) from the disassembly
- [x] 2.2 Own generator and harness, >= 100k cases including the quest multipliers, real d2d before and after
- [x] 2.3 Sweep d2d and game.exe for other price scalers and repair-gate callers; rebase recipe onto upstream main
- [x] 2.4 pr_guard, upstream tests, presubmit --quick, MSVC-class sweep, clang-tidy names and includes
- [x] 2.5 Verdict and limits in verify-price2.md

## 3. Close

- [x] 3.1 pr-bundle-queue.md rows, BACKLOG line, OWNERS, changecheck, leakcheck, openspec validate
- [x] 3.2 Remove the throwaway worktrees and branches on the worker host; agentreap empty
