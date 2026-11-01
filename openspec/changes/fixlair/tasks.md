# Tasks

```yaml
change: fixlair
status: in-progress
agent: fixlair
branch: agent/fixlair
worktree: ~/code/d2lod-wt/fixlair
skip_specs: true
owns:
  - docs/fixlair.md
  - internal/data/levels.go
  - internal/data/teleportrule_test.go
  - internal/boss/act2.go
  - internal/boss/act2_test.go
  - docs/act2.md
  - openspec/changes/fixlair/**
exclusive: []
gates:
  - "go test -p 1 ./internal/data ./internal/boss passes; whole-repo vet and go test -short on the worker host"
result:
```
