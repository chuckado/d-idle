# Tasks

```yaml
change: verifybundle
status: in-progress
agent: verifybundle
branch: agent/verifybundle
worktree: d2lod-wt/verifybundle
skip_specs: true
owns:
  - docs/research/pr-bundle-1-verifier.md
  - docs/research/pr-bundle-1-verifier/**
  - openspec/changes/verifybundle/**
exclusive: []
gates:
  - own harness run before the author's ground draft is compared
result: "pending"
```

- [ ] 1 Ground: own game.exe oracle and own d2d probe, base vs head counts
- [ ] 2 Integration on the bundle head (shared files, suites, readers)
- [ ] 3 CI parity (g++/clang++ warnings, MSVC by reading, include-cleaner, pr_guard, presubmit)
- [ ] 4 Commit hygiene and credit
- [ ] 5 PR body claims vs counts; size vs budget; merge with current upstream main
- [ ] 6 Verdict; changecheck and leakcheck
