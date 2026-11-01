# Tasks

```yaml
change: verifyarmorhit
status: in-progress
agent: verifyarmorhit
branch: agent/verifyarmorhit
worktree: ~/code/d2lod-wt/verifyarmorhit
skip_specs: true
owns:
  - openspec/changes/verifyarmorhit/**
  - tools/emu/verify_armor_hit.py
  - tools/emu/test_verify_armor_hit.cpp
  - docs/research/pr-armorhit-verifier.md
exclusive: []
gates:
  - "verify_armor_hit.py vs test_verify_armor_hit on >= 100k cases: base > 0, head 0 or classified; end frames compared"
result: "pending"
```

- [ ] trace the chain and the end / resume frame from the disassembly
- [ ] own generator and harness (seeds masked to 32 bits), >= 100k cases, base / tooling alone / claim alone / head
- [ ] CI-class checks
- [ ] det state hashes base vs patched on one hidden d2d slot
- [ ] verifier doc and bundle-queue verdict line
