# Tasks

```yaml
change: verifycurseimmune
status: archived
agent: verifycurseimmune
branch: agent/verifycurseimmune
worktree: ~/code/d2lod-wt/verifycurseimmune
skip_specs: true
owns:
  - openspec/changes/verifycurseimmune/**
  - tools/emu/verify_curse_immune.py
  - tools/emu/test_verify_curse_immune.cpp
  - docs/research/pr-curseimmune-verifier.md
exclusive: []
gates:
  - "verify_curse_immune.py vs test_verify_curse_immune on >= 100k cases per class: base > 0, head 0 or classified"
merge: 024c03e
result: "independent verifier: READY (own harness, 0 differ on head across 5 classes, hirelings and curse calcs checked); offline, not verified live"
```

- [x] read FUN_005c3540, FUN_005c35c0, FUN_005cedc0, FUN_005a1370's setter, FUN_00645b20
- [x] own generator and harness (seeds masked to 32 bits), >= 100k cases per class, base / claim alone / head
- [x] side claims: D2 untouched and unreachable, no other caller of apply_target_stats
- [x] merge-tree both orders with #155..#158 and the local candidates; the new upstream main
- [x] verifier doc with verdict; changecheck, leakcheck, ci check, openspec validate
