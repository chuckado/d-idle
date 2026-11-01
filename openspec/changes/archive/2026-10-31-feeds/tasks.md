# Tasks

```yaml
change: feeds
status: done
agent: feeds
branch: agent/feeds
worktree: ~/code/d2lod-wt/feeds
skip_specs: true
owns:
  - tools/discord/feed.py
  - tools/discord/tests/test_feed.py
  - docs/community/discord/feeds.md
  - docs/community/discord/server.md
  - docs/process/feed-after-events.md
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result:
```

## 1. Work

- [x] 1.1 feed.py with five subcommands, dry run default
- [x] 1.2 offline tests
- [x] 1.3 docs, rule, INDEX, CHANGELOG, BACKLOG
- [x] 1.4 one real post to #progress (verified live: posted)
