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
result: "READY after draft text edits: claim holds, 0 of 134,400 chill / freeze end frames differ (upstream main 109,573); shared deviation found: game.exe resumes a frozen monster at end + aidel (FUN_0057b170), d2d at end; merc wearer unchanged; det 14/14 identical; see docs/research/pr-armorhit-verifier.md; offline + one hidden d2d slot, not verified live"
```

- [x] trace the chain and the end / resume frame from the disassembly
- [x] own generator and harness (seeds masked to 32 bits), >= 100k cases, base / tooling alone / claim alone / head
- [x] CI-class checks
- [x] det state hashes base vs patched on one hidden d2d slot
- [x] verifier doc and bundle-queue verdict line
