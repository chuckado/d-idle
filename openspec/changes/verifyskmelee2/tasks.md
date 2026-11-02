# Tasks

```yaml
change: verifyskmelee2
status: in-progress
agent: verifyskmelee2
branch: agent/verifyskmelee2
worktree: ~/code/d2lod-wt/verifyskmelee2
skip_specs: true
owns:
  - docs/research/verify-skmelee2.md
  - tools/emu/vm2_*
  - openspec/changes/verifyskmelee2/**
exclusive: []
gates:
  - own harness run on two seeds before the author's trace is compared
result: D1 holds, D4 holds, D2 / D5 / D6 hold with limits; run on the emulator and a d2d probe build (V run), not live
```

- [x] 1 Read FUN_0057b420 / FUN_0057d9b0 / FUN_0057b660 in the disassembly
- [x] 2 Own harness (vm2_*), seeds 11 and 3, 20,000 cases each
- [x] 3 Write docs/research/verify-skmelee2.md with verdicts and differences from the author
- [x] 4 changecheck and leakcheck
