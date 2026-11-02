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
result: ""
```

- [ ] 1 poison units fix and monster_blow test on the fork
- [ ] 2 presubmit, ruff, clang-tidy findings
- [ ] 3 differentials, suites, det scenarios re-run
- [ ] 4 rebase on upstream main
- [ ] 5 doc update and PR draft
