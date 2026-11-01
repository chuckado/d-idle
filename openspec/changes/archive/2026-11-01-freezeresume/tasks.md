# Tasks

```yaml
change: freezeresume
status: archived
merge: bc86cec
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
result: "local PR candidate fix/freeze-resume proven by differential and independent verifier (READY WITH CHANGES, PR body edits pending, not opened upstream; not verified live). Original: Offline, nothing verified live, no d2d slot, no game window. Local upstream candidate fix/freeze-resume on the worker host (claim c6cc4f12, tooling efacc7f4, base upstream main 72132b99): a frozen monster thinks again at the freeze end + MonStats aidel (state 1 callback 0x57b170; FUN_005a7f80 drops thinks while frozen). Differential (d2d test_freeze_resume through the real Fight::strike and monster_update vs game.exe FUN_0057b230 / FUN_005a7f80 / 0x57b170 natively): frozen cases 12,503 -> 0 of 12,548 and 12,459 -> 0 of 12,516 (two seeds, every aidel value); specials differ by freeze-rules C3 (0 when merged with fix/freeze-rules-v2). pr_guard 0/0 per commit, presubmit 0 FAIL / 2 WARN answered, upstream tests green, gcc / clang-tidy clean on touched lines, ruff clean, merges clean both orders with #155..#158, freeze-rules-v2, armor-hit. NOT READY: independent verifier, det hashes."
```

## 1. Work

- [x] 1.1 Re-derive 0x57b170, FUN_005a7f80, FUN_0057b230's event filing, the queue's file (FUN_005416b0), stun and chill callbacks from the disassembly
- [x] 1.2 Game oracle freeze_resume.py (native) and d2d test mode test_freeze_resume (real Fight::strike + monster_update)
- [x] 1.3 Before numbers on upstream main, per class and per aidel
- [x] 1.4 Claim commit; after 0 on the frozen cases, two seeds; composition run with fix/freeze-rules-v2
- [x] 1.5 CI-class checks; merge check with #155..#158, freeze-rules-v2, armor-hit
- [x] 1.6 Proof, draft, card, OWNERS
- [ ] 1.7 Independent verifier and det state hashes (another card)
