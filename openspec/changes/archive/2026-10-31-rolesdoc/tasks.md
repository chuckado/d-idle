# Tasks

```yaml
change: rolesdoc
status: archived
agent: rolesdoc
branch: agent/rolesdoc
worktree: ~/code/d2lod-wt/rolesdoc
skip_specs: true
owns:
  - docs/community/discord/roles.md
  - docs/community/discord/server.md
  - docs/community/discord/server.json
  - docs/community/discord/plan.md
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests passes (no network)"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: 4d53842
result: "sourced Discord roles and safety guideline (roles.md) plus aligned server.md, server.json, plan.md; docs only, not verified live"
live: "not verified live: documents only; Discord support pages returned 403 to the fetch tool, official claims marked unconfirmed"
```

- [x] roles.md
- [x] reconcile server.md, server.json, plan.md
- [x] INDEX and CHANGELOG lines
