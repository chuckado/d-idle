# Tasks

```yaml
change: verifynexthit
status: archived
merge: 0674406
agent: verifynexthit
branch: agent/verifynexthit
worktree: ~/code/d2lod-wt/verifynexthit
skip_specs: true
owns:
  - openspec/changes/verifynexthit/**
  - tools/emu/verify_nexthit.py
  - tools/emu/test_verify_nexthit.cpp
  - tools/emu/verify_nexthit_compare.py
  - docs/research/pr-nexthit-verifier.md
  - docs/research/pr-bundle-queue.md
exclusive: []
gates:
  - "verify_nexthit.py vs test_verify_nexthit on >= 100k cases: base > 0, head 0 or classified, per claim slice"
result: "NOT READY: justhit end frame one short (f+5 in game.exe); head 25,252 of 100,624 differ, fix variant 0; det DET-CHANGES-EXPLAINED; see docs/research/pr-nexthit-verifier.md; offline + one hidden d2d slot, not verified live"
```

- [x] trace justhit set / length / end frame from the disassembly before reading the author's proof
- [x] own generator and harness (seeds masked to 32 bits on both sides), >= 100k cases, base / tooling alone / claim alone / head
- [x] CI-class checks (pr_guard per commit, presubmit --quick, upstream tests, gcc sweep, clang-tidy, ruff, replay on upstream main, shared files with open PRs)
- [x] det state hashes base vs patched on one hidden d2d slot, every difference classified
- [x] verifier doc and bundle-queue verdict line
