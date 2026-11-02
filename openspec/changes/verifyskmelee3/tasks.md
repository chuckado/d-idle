# Tasks

```yaml
change: verifyskmelee3
status: active
agent: verifyskmelee3
branch: agent/verifyskmelee3
worktree: d2lod-wt/verifyskmelee3
skip_specs: true
owns:
  - docs/research/pr-skmelee3-verifier.md
  - docs/research/pr-skmelee3-verifier/**
  - openspec/changes/verifyskmelee3/**
exclusive: []
gates:
  - own harness run before the author's draft is compared
result: ""
```

- [ ] 1 Read the candidate diff and the disassembly for each rule, address and mask
- [ ] 2 Own harness: game.exe oracle vs d2d base and head
- [ ] 3 Reader audit of Blow::frac, the floor and the new Fighter fields
- [ ] 4 Det scenarios base vs head
- [ ] 5 CI parity (g++/clang++ warnings, include-cleaner, suites, pr_guard, presubmit, merge)
- [ ] 6 Claims vs counts; verdict; changecheck and leakcheck
