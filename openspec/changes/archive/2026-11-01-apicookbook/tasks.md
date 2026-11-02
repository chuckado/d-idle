# Tasks

```yaml
change: apicookbook
status: archived
agent: apicookbook
branch: agent/apicookbook-pub
worktree: ~/code/d2lod-wt/apicookbook-pub
skip_specs: true
owns:
  - docs/api-agent-cookbook.md
  - tools/emu/run-on-worker*
  - tools/emu/run-py.sh
  - tools/emu/mpq-table.sh
  - openspec/changes/apicookbook/**
exclusive: []
gates:
  - tools/emu/run-on-worker-test.sh passes offline
result: "offline stub tests 26 of 26; live smoke through the ops runner passed"
merge: 7f3c38b
```

- [x] 1 Read the stream logs of the seven jobs, list working and denied forms
- [x] 2 Wrapper, run-py, mpq-table, smoke script, offline test
- [x] 3 Cookbook doc
- [ ] 4 Live Haiku job through the runner (ops repo side)
