# Tasks

```yaml
change: missilefix
status: in-progress
agent: missilefix
branch: agent/missilefix
worktree: ~/code/d2lod-wt/missilefix
skip_specs: true
owns:
  - docs/research/missilediff.md
  - docs/research/pr-missilerecord-draft.md
  - openspec/changes/missilefix/**
exclusive:
  - fork-branch:local/missilediff-record
  - fork-branch:local/missilediff-record-up
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "fork candidate fixed and pushed to the private remote (c301eaef, rebased ec09c73f, sibling 70fa479d); differentials 0 differ on both heads; suites and 14 det scenarios as base; NOT verified live (no game); independent re-verification owed"
```

- [x] 1 poison units fix and monster_blow test on the fork
- [x] 2 presubmit, ruff, clang-tidy findings
- [x] 3 differentials, suites, det scenarios re-run
- [x] 4 rebase on upstream main
- [x] 5 doc update and PR draft
