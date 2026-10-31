# Tasks

```yaml
change: fixstatus
status: in-progress
agent: fixstatus
branch: agent/fixstatus
worktree: ~/code/d2lod-wt/fixstatus
skip_specs: true
owns:
  - tools/agentstatus/**
  - docs/fixstatus.md
exclusive: []
gates:
  - "sh tools/agentstatus/stub-test.sh prints ALL OK"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: "board update accepted from the worker host"
```

- [x] root cause (shell stub in a .py file)
- [x] stub rewritten in Python
- [x] regression test tools/agentstatus/stub-test.sh
- [x] docs and changelog
