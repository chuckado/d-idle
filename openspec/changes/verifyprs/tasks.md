# Tasks

```yaml
change: verifyprs
status: in-progress
agent: verifyprs
branch: agent/verifyprs
worktree: ~/code/d2lod-wt/verifyprs
skip_specs: true
owns:
  - tools/emu/verify_static_field.py
  - tools/emu/verify_static_field_mode.inc
  - tools/emu/verify_walk_velocity.py
  - tools/emu/verify_walk_velocity_mode.inc
  - docs/research/pr-static-proof.md
  - docs/research/pr-walk-proof.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - openspec/changes/verifyprs/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result: "offline verification (no live game): fix/static-field HOLDS WITH LIMITS, 232,152 own cases, 100,624 -> 0 through the real Fight::spot (upstream 72132b99 before, the branch after), READY; fix/walk-velocity HOLDS WITH LIMITS, 121,216 cases, 70,171 -> 0 on the real builds (replayed on 72132b99), NOT READY (mercs and pets still run at Run, no reviewer-runnable d2d mode, rebase); pr_guard 0/0, upstream test binaries green, sweeps clean for both"
```

## 1. Static Field (fix/static-field)

- [x] 1.1 Re-derive the rule from FUN_005c9800 / FUN_005c96a0 / FUN_0056c8e0 / FUN_00483360 / FUN_0057c1e0 / FUN_0057be00 / FUN_0057bf80 / FUN_0063ee90
- [x] 1.2 Own generator and harness, >= 100k cases, real d2d before (upstream 72132b99) and after
- [x] 1.3 Sweep d2d and game.exe for other paths and readers
- [x] 1.4 pr_guard, upstream tests, presubmit --quick, MSVC-class sweep, clang-tidy names and includes
- [x] 1.5 Verdict and limits in pr-static-proof.md

## 2. Walk velocity (fix/walk-velocity)

- [x] 2.1 Replay onto current upstream main; re-derive from FUN_00623f50 / FUN_00621740 / FUN_00621360 / FUN_0057af80
- [x] 2.2 Own generator and harness, >= 100k cases, real d2d before and after
- [x] 2.3 Sweep for sibling paths (pets, mercs, skill moves, effective_speed)
- [x] 2.4 pr_guard, upstream tests, presubmit --quick, MSVC-class sweep, clang-tidy names and includes
- [x] 2.5 Verdict and limits in pr-walk-proof.md

## 3. Close

- [x] 3.1 pr-bundle-queue.md rows, BACKLOG line, OWNERS, changecheck, leakcheck, openspec validate
- [x] 3.2 Remove the throwaway worktrees and branches on the worker host; agentreap empty

