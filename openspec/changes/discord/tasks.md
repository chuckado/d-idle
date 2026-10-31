# Tasks

```yaml
change: discord
status: in-progress
agent: discord
branch: agent/discord
worktree: ~/code/d2lod-wt/discord
owns:
  - docs/community/discord/**
  - tools/discord/**
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests passes (no network)"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: "not verified live: nothing was sent to Discord or GitHub; tools tested against local mock servers only"
```

- [x] drafts: plan, server, server.json, welcome, rules, faq, templates, maintainer message, feeds, public-page
- [x] tools/discord/notify.py and tests
- [x] tools/discord/apply_server.py and tests
- [x] INDEX and CHANGELOG lines
