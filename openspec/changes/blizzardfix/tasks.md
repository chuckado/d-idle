# Tasks

```yaml
change: blizzardfix
status: in-progress
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
result: "Candidate fix/blizzard-shard-offsets-v2 (46c55621, 3d6f3fe7): cx/cy renamed, claim names Eruption, ruff clean; pr_guard 0/0, presubmit --quick 0 FAIL, differentials 14,926 / 14,957 -> 0 and verifier counts Blizzard 14,804 -> 0, Eruption 1,872 -> 0 re-run; Release build, test_game, test_rules rc 0, shadow and tidy clean. Draft updated, status READY held for the user go. Offline on the worker host, no live game."
live: "docs only, nothing run live"
```

- [x] short-names fix, claim commit names Eruption, ruff clean
- [x] pr_guard, presubmit --quick, differentials, Release build, tidy, shadow sweep
- [x] PR draft updated
