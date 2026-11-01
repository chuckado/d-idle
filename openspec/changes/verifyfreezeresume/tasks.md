# Tasks

```yaml
change: verifyfreezeresume
status: in-progress
agent: verifyfreezeresume
branch: agent/verifyfreezeresume
worktree: ~/code/d2lod-wt/verifyfreezeresume
skip_specs: true
owns:
  - openspec/changes/verifyfreezeresume/**
  - tools/emu/verify_freeze_resume.py
  - tools/emu/test_verify_freeze_resume.cpp
  - docs/research/pr-freezeresume-verifier.md
exclusive: []
gates:
  - "verify_freeze_resume.py (real event queue) vs test_verify_freeze_resume on >= 30k cases: base > 0, head 0 or classified, per case class"
result: ""
```

- [ ] trace the queue (FUN_005416b0, FUN_005414d0), 0x57b170, FUN_005a7f80 before reading the author's proof in depth
- [ ] own oracle with the real queue code, own generator (seeds masked to 32 bits)
- [ ] own d2d probe; builds base / claim alone / head
- [ ] differential by case class, off-by-one check
- [ ] side claims (kinds frozen, stun / chill callbacks, d2d freeze vs chill skills)
- [ ] verifier doc, checks, commit
