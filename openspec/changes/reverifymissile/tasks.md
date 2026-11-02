# Tasks

```yaml
change: reverifymissile
status: in-progress
agent: reverifymissile
branch: agent/reverifymissile
worktree: ~/code/d2lod-wt/reverifymissile
skip_specs: true
owns:
  - docs/research/pr-missilerecord-reverify.md
  - openspec/changes/reverifymissile/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "docs only: independent re-verification of the fixed monster missile record candidate and the Andariel poison split: READY WITH CHANGES for both; nothing posted upstream"
```

- [x] 1 re-verify the record claim, the consumer, the owner-poison add and the skill split with an own harness
- [x] 2 verdicts and the exact changes
