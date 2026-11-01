# Tasks

```yaml
change: feedtick
status: in-progress
agent: feedtick
branch: agent/feedtick
worktree: ~/code/d2lod-wt/feedtick
skip_specs: true
owns:
  - openspec/changes/feedtick/**
  - tools/discord/**
  - docs/process/feed-after-events.md
  - docs/community/discord/feeds.md
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge: ""
result: "Offline only: dry run and unit tests (mock webhook); nothing posted to Discord."
```

## 1. Work

- [x] 1.1 feed.py tick subcommand and tests
- [x] 1.2 docs, CHANGELOG, INDEX, BACKLOG
