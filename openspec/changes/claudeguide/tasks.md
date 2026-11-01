# Tasks

```yaml
change: claudeguide
status: in-progress
agent: claudeguide
branch: agent/claudeguide
worktree: ~/code/d2lod-wt/claudeguide
skip_specs: true
owns:
  - openspec/changes/claudeguide/**
  - docs/agents/claude-tracing-guide.md
gates:
  - "changecheck 0 warnings; leakcheck -mode fail 0"
result: "Docs only, not verified live (no game). Guide written from the repo docs and the maintainer's role decisions; model advice rests on one blind item."
```

## 1. Work

- [x] 1.1 Write the guide
- [x] 1.2 Pointer lines in agent-work.md and AGENTS.md
- [x] 1.3 Checks
