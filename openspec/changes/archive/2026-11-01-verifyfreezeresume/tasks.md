# Tasks

```yaml
change: verifyfreezeresume
status: archived
merge: f570e02
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
result: "independent verification complete (READY WITH CHANGES; not verified live). Original: READY WITH CHANGES (body only): real queue run, no off-by-one, claimed frozen classes 0 of 34,299 (base 34,200), claim alone = head; residuals prior stun, zero-length, d2d tick order; det hashes not run; offline, not verified live"
```

- [x] trace the queue (FUN_005416b0, FUN_005414d0), 0x57b170, FUN_005a7f80 before reading the author's proof in depth
- [x] own oracle with the real queue code, own generator (seeds masked to 32 bits)
- [x] own d2d probe; builds base / claim alone / head
- [x] differential by case class, off-by-one check
- [x] side claims (kinds frozen, stun / chill callbacks, d2d freeze vs chill skills)
- [x] verifier doc, checks, commit
