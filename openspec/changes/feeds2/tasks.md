# Tasks

```yaml
change: feeds2
status: in-progress
agent: feeds2
branch: agent/feeds2
worktree: ~/code/d2lod-wt/feeds2
skip_specs: true
owns:
  - .github/workflows/ci.yml
  - tools/discord/feed.py
  - tools/discord/tests/test_feed.py
  - docs/community/discord/feeds.md
  - docs/community/discord/server.md
  - docs/community/discord/plan.md
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result:
```

## 1. Work

- [ ] 1.1 CI job notify-failure and feed_test dispatch
- [ ] 1.2 repo webhook without workflow_run (after the feed test succeeded)
- [ ] 1.3 #d2d-upstream webhook (private file) and feed.py generic fail-closed text
- [ ] 1.4 one real #d2d-upstream pr-status post (verified live: posted)
- [ ] 1.5 docs, CHANGELOG, BACKLOG
