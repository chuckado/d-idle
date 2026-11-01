# Tasks

```yaml
change: crfix156
status: in-progress
agent: CRFIX156
branch: agent/crfix156
worktree: ~/code/d2lod-wt/crfix156
skip_specs: true
owns:
  - docs/research/pr-walk-evidence-pack.md
  - docs/research/pr-bundle-queue.md
  - openspec/changes/crfix156/**
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "CodeRabbit on upstream PR #156: finding 1 (merc.cpp) INVALID, answered in thread; finding 2 (walk_velocity.py compare) VALID, fixed and pushed as a fast-forward (3f418788), answered in thread"
```

## 1. Work

- [x] 1.1 judge both findings against the code and the differential
- [x] 1.2 fix finding 2 in the tooling commit (autosquash), re-run the gates
- [x] 1.3 reply in the finding 1 thread
- [x] 1.4 push finding 2 as a fast-forward and reply in its thread
