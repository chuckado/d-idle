# Tasks

```yaml
change: discordtoken
status: archived
skip_specs: true
agent: discordtoken
branch: agent/discordtoken
worktree: ~/code/d2lod-wt/discordtoken
owns:
  - tools/discord/common.py
  - tools/discord/apply_server.py
  - tools/discord/tests/test_token.py
  - docs/community/discord/server.md
  - docs/community/discord/plan.md
  - docs/community/discord/roles.md
  - openspec/changes/discordtoken/**
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: 88e2cf68
result: "806ff4e merged; 40 offline unit tests pass (fake values only); not verified live (no network, no real token touched)"
```

## 1. Work

- [x] 1.1 load_bot_token with fail-closed file rules, used by apply_server.py
- [x] 1.2 unit tests (env/file/keyring order, mode, symlink, owner, parent, whitespace, empty)
- [x] 1.3 server.md, roles.md, plan.md updated for the long-lived ops bot
