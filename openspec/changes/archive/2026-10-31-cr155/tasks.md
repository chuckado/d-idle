# Tasks

```yaml
change: cr155
status: archived
agent: CR155
branch: agent/cr155
merge: 06c1f1a
worktree: ~/code/d2lod-wt/cr155
owns:
  - docs/research/pr-static-evidence-pack.md
  - docs/research/pr-static-chill-proof.md
  - docs/BACKLOG.md
  - docs/plan/STATUS.md
  - openspec/changes/cr155/**
exclusive: []
gates:
  - "changecheck status -write then -cli: 0 warnings; leakcheck -mode fail: 0 hits; tools/ci/check.sh --hygiene-only: no FAIL; tools/openspec.sh validate --all --strict passes"
result: "answered CodeRabbit on #155: seed masks (83f35558) and the oracle modes moved to their own test files (ddc8281d), pushed as a fast-forward; verified live: not applicable (offline tests and differentials on the worker host)"
```

## 1. Work

- [x] 1.1 investigate the findings against the code, prove the negative-seed divergence before and the match after
- [x] 1.2 fixup commits, CI-class checks, fast-forward push to the PR branch
- [x] 1.3 reply on each CodeRabbit thread and one PR comment
- [x] 1.4 posted lines in the two docs, BACKLOG line
