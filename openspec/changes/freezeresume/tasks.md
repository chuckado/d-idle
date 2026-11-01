# Tasks

```yaml
change: freezeresume
status: in-progress
agent: freezeresume
branch: agent/freezeresume
worktree: ~/code/d2lod-wt/freezeresume
skip_specs: true
owns:
  - openspec/changes/freezeresume/**
  - docs/research/pr-freezeresume-proof.md
  - docs/research/pr-freezeresume-draft.md
  - tools/emu/freeze_resume.py
  - tools/emu/freeze_resume_mode.inc
exclusive:
  - "fork-branch:fix/freeze-resume"
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
  - "PR claim only with a differential on real builds: before > 0, after 0"
result: "RESULT_PLACEHOLDER"
```

## 1. Work

- [x] 1.1 Re-derive 0x57b170, FUN_005a7f80, FUN_0057b230's event filing, the queue's file (FUN_005416b0), stun and chill callbacks from the disassembly
- [x] 1.2 Game oracle freeze_resume.py (native) and d2d test mode test_freeze_resume (real Fight::strike + monster_update)
- [x] 1.3 Before numbers on upstream main, per class and per aidel
- [x] 1.4 Claim commit; after 0 on the frozen cases, two seeds; composition run with fix/freeze-rules-v2
- [x] 1.5 CI-class checks; merge check with #155..#158, freeze-rules-v2, armor-hit
- [x] 1.6 Proof, draft, card, OWNERS
- [ ] 1.7 Independent verifier and det state hashes (another card)
