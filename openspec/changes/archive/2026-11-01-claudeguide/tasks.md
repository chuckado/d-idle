# Tasks

```yaml
change: claudeguide
status: archived
agent: claudeguide
branch: agent/claudeguide
worktree: ~/code/d2lod-wt/claudeguide
skip_specs: true
owns:
  - openspec/changes/claudeguide/**
  - docs/agents/claude-tracing-guide.md
gates:
  - "changecheck 0 warnings; leakcheck -mode fail 0"
result: "public guide for Claude users added (docs/agents/claude-tracing-guide.md): evidence tags, independent verifier standard, model per role (Claude only, other models untested), sanitized briefs; pointer lines in agent-work.md and AGENTS.md"
merge: f933d0f
```

## 1. Work

- [x] 1.1 Write the guide
- [x] 1.2 Pointer lines in agent-work.md and AGENTS.md
- [x] 1.3 Checks
