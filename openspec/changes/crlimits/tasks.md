# Tasks

```yaml
change: crlimits
status: in-progress
agent: crlimits
branch: agent/crlimits
worktree: d2lod-wt/crlimits
skip_specs: true
owns:
  - docs/research/coderabbit-limits-20261102.md
  - openspec/changes/crlimits/**
exclusive: []
gates:
  - "changecheck 0 warnings; leakcheck 0 hits"
result: "Limits table and bundle budget written; sources fetched 2026-11-02."
live: "not applicable (docs only)"
```

- [x] gather limits (repo config, upstream PR history, CodeRabbit docs)
- [x] write research doc with recommendation
