# Tasks

```yaml
change: armorhit
status: in-progress
agent: armorhit
branch: agent/armorhit
worktree: ~/code/d2lod-wt/armorhit
skip_specs: true
owns:
  - openspec/changes/armorhit/**
  - docs/research/pr-armorhit-proof.md
  - docs/research/pr-armorhit-draft.md
  - docs/research/pr-bundle-queue.md
  - tools/emu/armor_hit.py
  - tools/emu/armor_hit_mode.inc
exclusive:
  - "fork-branch:fix/armor-hit"
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
  - "PR claim only with a differential on real builds: before > 0, after 0"
result: "Offline, nothing verified live, no d2d slot, no game window. Local upstream candidate fix/armor-hit on the worker host (claim 36cbe015, tooling d5bd19c1, base upstream main 72132b99): Frozen Armor's striker effect follows game.exe FUN_0057b230 (resisted by cold res less pierce, specials chilled only, plain monsters frozen at coldeffect < 0 for length / MonsterFreezeDivisor). Differential (d2d test_armor_hit through the real Fight::buff_events vs game.exe FUN_005cac50 natively under the emulator): 95,920 -> 0 of 108,506 cases (every damaging case differed before), seed 2 95,986 -> 0. pr_guard 0/0 per commit, presubmit --quick 0 FAIL / 3 WARN answered, upstream tests green, gcc shadow / clang-tidy clean on touched lines, merges clean with #155 and #156 both ways. NOT READY: independent verifier, det hashes."
```

## 1. Work

- [x] 1.1 Worktree fix/armor-hit from current upstream main on the worker host
- [x] 1.2 d2d test mode test_armor_hit (real buff_events) + game oracle armor_hit.py (FUN_005cac50 natively)
- [x] 1.3 Before numbers on upstream main (95,920 of 108,506)
- [x] 1.4 Claim commit; after 0 of 108,506; second seed 95,986 -> 0
- [x] 1.5 CI-class checks; merge check with #155 / #156
- [x] 1.6 Proof, draft, queue line, card, OWNERS
- [ ] 1.7 Independent verifier and det state hashes (another card)
