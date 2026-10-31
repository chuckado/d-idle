# Tasks

```yaml
change: staticopen
status: in-progress
agent: staticopen
branch: agent/staticopen
worktree: ~/code/d2lod-wt/staticopen
skip_specs: true
owns:
  - docs/research/pr-static-evidence-pack.md
  - docs/research/pr-wp-evidence-pack.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - openspec/changes/staticopen/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result: ""
```

## 1. Work

- [x] 1.1 Replay fix/static-field onto upstream main as fix/static-field-v2 (already on upstream 72132b99; identity correct)
- [x] 1.2 CI-class checks, push to the fork, open the PR (#155)
- [x] 1.3 Poll CI, evidence-pack posted lines, bundle-queue row, BACKLOG line
