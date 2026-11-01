# Tasks

```yaml
change: buildflex
status: archived
merge: c7f140e
agent: buildflex
branch: agent/buildflex
worktree: ~/code/d2lod-wt/buildflex
skip_specs: true
owns:
  - docs/research/build-flexibility-audit.md
  - openspec/changes/buildflex/**
exclusive: []
gates:
  - "changecheck status -write then -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs only: docs/research/build-flexibility-audit.md (39 findings, seams S1-S6, phases 0-9, guardrail text, applied to AGENTS.md). Not verified live: docs only, nothing run live."
live: "docs only, nothing run live"
```

- [x] inventory by layer with severity and seam
- [x] what already works as data
- [x] seams design, phased roadmap, guardrail text
- [x] shared-file lines for the merger
