# Tasks

```yaml
change: fixemptyplan
status: in-progress
agent: fixemptyplan
branch: agent/fixemptyplan
worktree: ~/code/d2lod-wt/fixemptyplan
skip_specs: true
owns:
  - internal/input/client.go
  - internal/nav/nav.go
  - internal/nav/walk.go
  - internal/nav/fixemptyplan_test.go
  - internal/combat/fight.go
  - internal/combat/reach.go
  - internal/combat/fixemptyplan_test.go
  - docs/fixemptyplan.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof after merge: low-mana pack next to a wall, log has no 'empty path' and the bot falls through to its next option (docs/fixemptyplan.md)"
merge:
result:
```

## 1. Work

- [x] 1.1 reproduce: blocked goal beside the player, plan empty, PathTo, GoTo
- [x] 1.2 ErrAlreadyThere; PathTo/walk typed; valid vs validApproach; move quiet tick
- [x] 1.3 tests (reproduction, approach vs back-off table, snapped-start edge cases, bounded candidates)
- [x] 1.4 docs/fixemptyplan.md, primitives note, INDEX, CHANGELOG, BACKLOG
- [ ] 1.5 live proof (not run: offline card)
