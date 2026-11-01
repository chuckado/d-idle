# Tasks

```yaml
change: blizzardfix
status: archived
agent: blizzardfix
branch: agent/blizzardfix
worktree: ~/code/d2lod-wt/blizzardfix
skip_specs: true
owns:
  - docs/research/pr-blizzard-draft.md
  - openspec/changes/blizzardfix/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "candidate fix/blizzard-shard-offsets-v2 updated per the verifier (cell_x/cell_y, Eruption named, ruff clean); presubmit 0 FAIL; PR draft READY"
live: "docs only, nothing run live"
merge: acb7daf
```

- [x] short-names fix, claim commit names Eruption, ruff clean
- [x] pr_guard, presubmit --quick, differentials, Release build, tidy, shadow sweep
- [x] PR draft updated
