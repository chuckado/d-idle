# Tasks

```yaml
change: apicookbook
status: in-progress
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
result: wrapper tested offline (stub labrun) and live from the head box (mpq-table monstats first line, hello.py on the worker host); live Haiku job through the runner recorded in the ops repo
```

- [x] 1 Read the stream logs of the seven jobs, list working and denied forms
- [x] 2 Wrapper, run-py, mpq-table, smoke script, offline test
- [x] 3 Cookbook doc
- [ ] 4 Live Haiku job through the runner (ops repo side)
