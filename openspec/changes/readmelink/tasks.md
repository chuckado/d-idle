# Tasks

```yaml
change: readmelink
status: in-progress
agent: readmelink
branch: agent/readmelink
worktree: ~/code/d2lod-wt/readmelink
owns:
  - openspec/changes/readmelink/**
  - README.md
  - docs/community/discord/plan.md
  - docs/community/discord/public-page.md
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; go build ./..."
result: ""
live: "not applicable (docs only)"
```

- [x] 1.1 README line
- [x] 1.2 plan.md and public-page.md live state
