# Tasks

```yaml
change: fixlair
status: archived
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
result: offline-complete, NOT verified live (live drill T6 needs the Duriel lair, unreachable early); TestRealLairTeleport passed on the worker host with a real install; data and comments only, no behaviour change.
merge: 19ca7c447c568be98d0390bf3880369f96be2f55
```
