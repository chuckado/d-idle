# Tasks

```yaml
change: verifycurseimmune
status: in-progress
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
result: ""
```

- [ ] read FUN_005c3540, FUN_005c35c0, FUN_005cedc0, FUN_005a1370's setter, FUN_00645b20
- [ ] own generator and harness (seeds masked to 32 bits), >= 100k cases per class, base / claim alone / head
- [ ] side claims: D2 untouched and unreachable, no other caller of apply_target_stats
- [ ] merge-tree both orders with #155..#158 and the local candidates; the new upstream main
- [ ] verifier doc with verdict; changecheck, leakcheck, ci check, openspec validate
