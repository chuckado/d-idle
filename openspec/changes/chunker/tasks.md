# Tasks

```yaml
change: chunker
status: in-progress
agent: chunker
branch: agent/chunker
worktree: ~/code/d2lod-wt/chunker
skip_specs: true
owns:
  - tools/contrib/chunker.py
  - tools/contrib/tests/test_chunker.py
  - docs/process/chunker.md
  - docs/community/groom-proposal.md
  - docs/community/withdrawn.tsv
  - openspec/changes/chunker/**
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/contrib/tests -t tools/contrib passes"
  - "go run ./cmd/changecheck -cli reports 0 warnings; go run ./cmd/leakcheck -mode fail 0 hits"
merge: ""
result: "pending"
```

## 1. Tool
- [x] 1.1 sync mode (create, close, body drift, stale claim, write cap, dry run)
- [x] 1.2 groom mode (schema and path validation, privacy and game-code checks, proposal file)

## 2. Tests and docs
- [x] 2.1 tests with a fake gh and a fake HTTP layer
- [x] 2.2 docs/process/chunker.md
- [x] 2.3 real groom dry runs (3 calls) and a read-only `sync` dry run
