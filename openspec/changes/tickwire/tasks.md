# Tasks

```yaml
change: tickwire
status: in-progress
agent: tickwire
branch: agent/tickwire
worktree: ~/code/d2lod-wt/tickwire
skip_specs: true
owns:
  - tools/discord/tick_post.py
  - tools/discord/tests/test_tick_post.py
  - openspec/changes/tickwire/**
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests -t tools/discord -p test_tick_post.py passes; tools/ci/check.sh passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "tests pass with fake HTTP and sender; verified live once: one real Haiku call and one real #progress post (wiring test)"
```

- [x] tick_post.py and tests
- [x] one real dry run and one real post
- [x] docs and scripts.sh
