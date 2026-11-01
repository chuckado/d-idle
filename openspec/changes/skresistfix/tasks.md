# Tasks

```yaml
change: skresistfix
status: in-progress
agent: skresistfix
branch: agent/skresistfix
worktree: ~/code/d2lod-wt/skresistfix
skip_specs: true
owns:
  - openspec/changes/skresistfix/**
  - docs/research/pr-curseimmune-draft.md
  - docs/research/pr-curseimmune-proof.md
exclusive:
  - "fork-branch:fix/curse-immune"
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
  - "PR claim only with a differential on real builds: before > 0, after 0"
result: "Offline, nothing verified live, no d2d slot, no game window. Local upstream candidate fix/curse-immune on the worker host (claim 6a04cb3e, tooling 557d14e3, base upstream main 72132b99): a curse or aura lowering a resistance counts a fifth (toward 0) when the monster's base resistance is 100 or more (game.exe FUN_005c3540). Differential (real Fight::target_of and rules::missile_blow vs game.exe natively): adj 88,101 -> 0 of 147,628 (seed 7: 88,224 -> 0), dealt 14,355 -> 1,958 (seed 7: 14,289 -> 1,977); the 1,958 left are all D2 (magic total under -100, not in the claim, unreachable in 1.14d data); a line-for-line port of the fixed rule equals the real build on every case. pr_guard 0/0, presubmit --quick 0 FAIL, five test binaries rc 0, gcc shadow and clang-tidy clean on touched lines, ruff clean, merges clean with #152 #153 #155 #156 #157 #158, fix/armor-hit, fix/freeze-resume, fix/frozen-armor-and-resume-2c. NOT READY: independent verifier, det state hashes, the user's go."
```

## 1. Work

- [x] 1.1 Worktree skresistfix, branch fix/curse-immune from upstream main 72132b99 (worker host)
- [x] 1.2 Claim commit; test mode test_sk_resist (real target_of) and oracle sk_resist.py as the tooling commit
- [x] 1.3 Differential before and after on two seeds (147,628 cases each)
- [x] 1.4 CI-class checks and merge check with the open PR branches and local candidates
- [x] 1.5 Draft, proof, card, OWNERS
- [ ] 1.6 Independent verifier and det state hashes (another card); the user's go before any push or post
