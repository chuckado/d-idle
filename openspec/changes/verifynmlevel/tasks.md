# Tasks

```yaml
change: verifynmlevel
status: in-progress
agent: verifynmlevel
branch: agent/verifynmlevel
worktree: ~/code/d2lod-wt/verifynmlevel
skip_specs: true
owns:
  - docs/research/pr-nmlevel-verifier.md
  - docs/research/pr-nmlevel-verifier/**
  - openspec/changes/verifynmlevel/**
exclusive: []
gates:
  - own harness run on two seeds before the author's draft is compared
result: pending
```

- [ ] 1 Read FUN_00573cb0 / FUN_0061a1b0 / FUN_0061dca0 and the masks in the disassembly
- [ ] 2 Own harness: game.exe populations (two seeds, every non-town level, three difficulties) against d2d base and head
- [ ] 3 Caller and reader audit of make_monster / monster_stats
- [ ] 4 Det scenarios base and head (one hidden d2d slot)
- [ ] 5 CI parity: g++ / clang++ warnings, clang-tidy Names and includes, suites, merge against upstream main
- [ ] 6 Compare with the author's draft; write docs/research/pr-nmlevel-verifier.md with the verdict
- [ ] 7 changecheck and leakcheck
