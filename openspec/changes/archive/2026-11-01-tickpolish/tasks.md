# Tasks

```yaml
change: tickpolish
status: archived
agent: tickpolish
branch: agent/tickpolish
worktree: ~/code/d2lod-wt/tickpolish
skip_specs: true
owns:
  - tools/discord/tick_post.py
  - tools/discord/tests/test_tick_post.py
  - openspec/changes/tickpolish/**
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests passes; tools/ci/check.sh passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "tick wrapper repairs guard runs (separators only) and renders an empty waiting list without the model; unit-tested with the real guard; not verified live"
merge: 8422bcc
```

- [x] separator repair for guard runs
- [x] empty waiting accepted
- [x] tests
