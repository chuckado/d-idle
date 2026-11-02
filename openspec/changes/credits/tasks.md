# Tasks

```yaml
change: credits
status: in-progress
agent: credits
branch: agent/credits
worktree: d2lod-wt/credits
skip_specs: true
owns:
  - CONTRIBUTORS.md
  - docs/process/chunker.md
  - README.md
  - openspec/changes/credits/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh"
result: "Docs only; PR facts from gh pr list, upstream count 23 from gh. Nothing run live."
live: "docs only, nothing run live"
```

- [x] CONTRIBUTORS.md
- [x] Credit section, README and INDEX lines
