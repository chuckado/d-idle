# Tasks

```yaml
change: dethash
status: in-progress
agent: dethash
branch: agent/dethash
worktree: ~/code/d2lod-wt/dethash
skip_specs: true
owns:
  - docs/research/det-hashes-2026-11-01.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - openspec/changes/dethash/**
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0"
result: "det run done (headless d2d slot, no game.exe): #155 DET-CLEAN, chill DET-CLEAN with coverage gap, walk3 DET-CHANGES-EXPLAINED; docs/research/det-hashes-2026-11-01.md. Not run on game.exe."
```

## 1. Work

- [x] 1.1 Build d2d: upstream main + lab patch set (baseline), and each candidate head with the same lab patch set
- [x] 1.2 Run det6.py (passive and FIGHT) 3x per build; diff hashes
- [x] 1.3 Explain each change against the claim; verdicts
- [x] 1.4 Doc, queue rows, BACKLOG, OWNERS
