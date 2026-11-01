# Tasks

```yaml
change: nexthit
status: in-progress
agent: nexthit
branch: agent/nexthit
worktree: ~/code/d2lod-wt/nexthit
skip_specs: true
owns:
  - docs/research/pr-nexthit-proof.md
  - docs/research/pr-nexthit-draft.md
  - tools/emu/next_hit.py
  - tools/emu/next_hit_mode.inc
  - tools/emu/next_hit_summary.py
  - openspec/changes/nexthit/**
exclusive:
  - "fork-branch:fix/nexthit"
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
  - "PR claim only with a differential on real builds: before > 0, after 0"
result: "Offline, nothing verified live, no d2d slot used. Local upstream candidate fix/nexthit on the worker host (claim 8abab0ed, tooling 7901fda2, base upstream main 72132b99): NextHit hit sets unit justhit for NextDelay frames, LastCollide last unit, one unit a frame. Differential (d2d test_next_hit real Fight::fire / Fight::world contacts replayed through game.exe FUN_005adf10 / FUN_005adb00 under the emulator): Nova / Frost Nova 84,779 -> 0 of 100,198 cases; d2d 4.2-4.4 hits per touched monster per cast vs game.exe 1.04. pr_guard 0/0 per commit, presubmit --quick 0 FAIL, four upstream test binaries green, shadow / clang-tidy clean. NOT READY: verifier, det hashes; Chain Lightning hop (C2) and Lightning per frame (C3) designed, not built. One docs line in pr-bundle-queue.md (shared file, owned by no open card)."
```

## 1. Work

- [x] 1.1 Worktree fix/nexthit from current upstream main on the worker host
- [x] 1.2 Stateful differential: d2d test_next_hit mode + game.exe replay oracle
- [x] 1.3 Before numbers on upstream main (84,779 of 100,198 cases differ)
- [x] 1.4 Claim commit (justhit, LastCollide, one unit a frame); after 0 of 100,198
- [x] 1.5 CI-class checks (pr_guard, presubmit, tests, shadow, clang-tidy, ruff)
- [ ] 1.6 Chain Lightning hop search (C2) and Lightning per frame (C3): designs only
- [x] 1.7 Proof, draft, queue line, card; bot-measurement note
- [ ] 1.8 Independent verifier and det state hashes (later, another agent)
