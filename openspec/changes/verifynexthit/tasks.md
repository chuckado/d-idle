# Tasks

```yaml
change: verifynexthit
status: in-progress
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
result: ""
```

- [ ] trace justhit set / length / end frame from the disassembly before reading the author's proof
- [ ] own generator and harness (seeds masked to 32 bits on both sides), >= 100k cases, base / tooling alone / claim alone / head
- [ ] CI-class checks (pr_guard per commit, presubmit --quick, upstream tests, gcc sweep, clang-tidy, ruff, replay on upstream main, shared files with open PRs)
- [ ] det state hashes base vs patched on one hidden d2d slot, every difference classified
- [ ] verifier doc and bundle-queue verdict line
