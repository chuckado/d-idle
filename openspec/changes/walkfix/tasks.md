# Tasks

```yaml
change: walkfix
status: in-progress
agent: walkfix
branch: agent/walkfix
worktree: ~/code/d2lod-wt/walkfix
skip_specs: true
owns:
  - docs/research/pr-walk-proof.md
  - docs/research/pr-walk-draft.md
  - docs/research/pr-walk-evidence-pack.md
  - docs/research/pr-bundle-queue.md
  - tools/emu/walk_velocity.py
  - tools/emu/walk_velocity_mode.inc
  - openspec/changes/walkfix/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result: "offline (no live game, nothing pushed): branch fix/walk-velocity-v2 head 10338345 on upstream 72132b99; merc and pet claims added; real-d2d differential 122,646 cases, 69,437 -> 0 (monster 38,543, player 15,146, merc 8,287, pet 7,461); CI-class checks clean; READY after the det state-hash run (not run, needs a d2d slot)"
```

## 1. Branch

- [x] 1.1 fix/walk-velocity-v2 from the author's branch, replayed on current upstream main (72132b99)
- [x] 1.2 merc commit (merc.cpp) and pet commit (pets.cpp), each with its own differential
- [x] 1.3 tooling commit: test_game --walk-velocity (real monster_update, walk_percent, merc_turn, pet_walking) and tools/emu/walk_velocity.py

## 2. CI-class checks

- [x] 2.1 clang-tidy names and includes, gcc shadow sweep, ruff, pr_guard, presubmit --quick, the four test binaries
- [ ] 2.2 det state hashes (needs a d2d slot: not run here)

## 3. Docs

- [x] 3.1 pr-walk proof, draft, evidence pack, bundle-queue walk row
- [x] 3.2 changecheck, leakcheck, openspec validate
