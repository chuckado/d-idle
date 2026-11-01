# Tasks

```yaml
change: feedtick2
status: archived
agent: feedtick2
branch: agent/feedtick2
worktree: ~/code/d2lod-wt/feedtick2
skip_specs: true
owns:
  - openspec/changes/feedtick2/**
  - tools/discord/feed.py
  - tools/discord/tests/**
  - docs/process/feed-after-events.md
  - docs/community/discord/feeds.md
exclusive: []
merge: 015b95c
gates:
  - "python3 -m unittest discover -s tools/discord/tests; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "Offline only (dry-run tested, 80 tests pass), nothing posted, not verified live. tick cap 3800 chars with a final '(cut: N chars)' line."
```

- [x] 1.1 raise cap, cut marker, tests
- [x] 1.2 docs and shared-file lines (in proposal.md)
