# Tasks

```yaml
change: auditfix
status: in-progress
agent: auditfix
branch: agent/auditfix
worktree: ~/code/d2lod-wt/auditfix
skip_specs: true
owns:
  - openspec/changes/auditfix/**
  - tools/ci/gamecode_guard.sh
  - tools/ci/gamecode_allow.txt
  - tools/ci/gamecode_test.sh
  - tools/ci/scrub_paths.py
  - docs/process/gamecode-guard.md
  - docs/fixtures.md
  - cmd/leakcheck/scan.go
  - cmd/leakcheck/leakcheck_test.go
exclusive: []
gates:
  - "tools/ci/gamecode_test.sh passes (fake listing and pseudo-C fail) and tools/ci/check.sh passes on the whole tree"
  - "go test -p 1 ./cmd/leakcheck passes incl. TestLabPathScopedToProbes; go run ./cmd/leakcheck -mode fail finds 0"
  - "go run ./cmd/changecheck -cli has 0 warnings; tools/openspec.sh validate --all --strict passes"
result: ""
live: "n/a: offline docs and CI"
```

- [x] README line, docs/fixtures.md (fixtures inspected)
- [x] gamecode guard, allow list, counter-test, wired into check.sh and scripts.sh
- [x] leakcheck labpath class with tests; scrub_paths.py; 801 paths scrubbed
- [x] docs/process/gamecode-guard.md, INDEX, CHANGELOG, BACKLOG
