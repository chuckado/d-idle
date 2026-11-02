# Tasks

```yaml
change: reverifymissile
status: archived
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
result: "docs only; nothing upstream; PR drafts to be revised"
merge: a43549d
```

- [x] 1 re-verify the record claim, the consumer, the owner-poison add and the skill split with an own harness
- [x] 2 verdicts and the exact changes
