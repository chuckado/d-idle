# Tasks

```yaml
change: fix-obj
status: in-progress
agent: fixobj
branch: agent/fixobj
worktree: ~/code/d2lod-wt/fixobj
owns:
  - internal/d2dsrc/items.go
  - internal/d2dsrc/source.go
  - internal/d2dsrc/source_test.go
  - internal/d2dinput/controller.go
  - internal/d2dinput/controller_test.go
  - internal/d2denv/env.go
  - internal/d2denv/fixobj_test.go
  - docs/fixobj.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... plus go test -race ./internal/d2dinput ./internal/d2dsrc ./internal/d2dctl ./internal/d2denv ./cmd/d2bot pass on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof is a drill rerun after merge: H.obj, H.pause, D8, D3.n (docs/fixobj.md)"
merge:
result: ""
live: ""
```

- [x] 1.1 d2dsrc lists placed objects near the player from debug presets
- [x] 1.2 d2dinput Interact for a listed object sends cmd interact <index>
- [x] 1.3 d2denv Pause steps under lockstep, wall sleep otherwise
- [x] 1.4 scenario tests, spec delta, docs/fixobj.md
- [ ] 1.5 live proof: drill rerun (not run here)
