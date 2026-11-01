# Tasks

```yaml
change: modelab
status: archived
agent: modelab
branch: agent/modelab
worktree: ~/code/d2lod-wt/modelab
skip_specs: true
owns:
  - docs/research/model-ab-skresist.md
  - openspec/changes/modelab/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "three blind arms compared against the Opus trace: Sonnet found the decisive curse rule and the real d2d bug, Haiku missed it; docs only"
merge: 058ec1c
```

## 1. Work

- [x] 1.1 Read the three arm results, the verifier, the usage ledger
- [x] 1.2 Rule-by-rule table, deviations, errors, blindness, effort
- [x] 1.3 Conclusions and limits; merge-lines.txt; checks
