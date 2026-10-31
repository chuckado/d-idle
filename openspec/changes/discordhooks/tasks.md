# Tasks

```yaml
change: discordhooks
status: in-progress
skip_specs: true
agent: discordhooks
branch: agent/discordhooks
worktree: ~/code/d2lod-wt/discordhooks
owns:
  - tools/discord/common.py
  - tools/discord/notify.py
  - tools/discord/tests/test_notify.py
  - tools/discord/tests/test_webhook.py
  - docs/community/discord/feeds.md
  - docs/community/discord/server.md
  - openspec/changes/discordhooks/**
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
```

## 1. Work

- [x] 1.1 load_webhook_url (env, file, keyring; fail-closed; shape check; channel validation; D2IDLE_DISCORD_HOME)
- [x] 1.2 notify.py uses it; tests isolated from the real home
- [x] 1.3 feeds.md, server.md, INDEX, CHANGELOG
- [ ] 1.4 feed test posts to the three channels
