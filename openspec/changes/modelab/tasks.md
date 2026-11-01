# Tasks

```yaml
change: modelab
status: in-progress
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
result: "Docs only, nothing verified live and no emulator run repeated: every count is quoted from the arm documents, the Opus trace, the independent verifier and the usage ledger. Verdict: Opus and Sonnet found the curse 1/5 rule (the real bug), Haiku missed it and had only a port of d2d; Sonnet has one doubtful claim (Fire / Lightning Mastery as pierce) and one extra (physical whole points) to re-check."
```

## 1. Work

- [x] 1.1 Read the three arm results, the verifier, the usage ledger
- [x] 1.2 Rule-by-rule table, deviations, errors, blindness, effort
- [x] 1.3 Conclusions and limits; merge-lines.txt; checks
