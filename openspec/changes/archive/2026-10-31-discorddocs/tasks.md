# Tasks

```yaml
change: discorddocs
status: archived
agent: discorddocs
branch: agent/discorddocs
worktree: ~/code/d2lod-wt/discorddocs
owns:
  - openspec/changes/discorddocs/**
  - docs/community/discord/plan.md
  - docs/community/discord/server.md
  - docs/community/discord/roles.md
  - docs/community/discord/feeds.md
  - docs/community/discord/server.json
  - tools/discord/apply_server.py
  - tools/discord/tests/test_apply.py
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge: 7421c4e
result: "docs and layout file updated to the live state; no live Discord call made; tool tests green"
live: "not applicable (docs only; facts were verified live by the coordinator)"
```

- [x] 1.1 plan, server, roles, feeds updated to the live state
- [x] 1.2 server.json STAFF category and news type; tool and test
