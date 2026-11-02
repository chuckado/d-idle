# Tasks

```yaml
change: propenmissile
status: in-progress
agent: propenmissile
branch: agent/propenmissile
worktree: ~/code/d2lod-wt/propenmissile
skip_specs: true
owns:
  - docs/research/openpr-log.md
  - openspec/changes/propenmissile/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Upstream PR #164 opened (bundled record + poison units, 8 commits, head ee8a2093); logged in docs/research/openpr-log.md. Checks run on the worker host (offline build and tests), not against a live game."
live: "docs only, nothing run live"
```

- [x] push the combined branch to the fork (plain) and open the PR
- [x] log the PR in docs/research/openpr-log.md
