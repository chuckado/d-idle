# Tasks

```yaml
change: waits-audit
status: in-progress
agent: waitsaudit
branch: agent/waitsaudit
worktree: ~/code/d2lod-wt/waitsaudit
owns:
  - internal/d2dctl/town_verbs.go
  - internal/d2dctl/waits_test.go
  - internal/d2denv/env.go
  - internal/d2denv/waits.go
  - internal/d2denv/waits_test.go
  - cmd/d2dshop/main.go
  - docs/probes/waits-audit.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... plus go test -race on d2dctl, d2denv, cmd/d2dshop pass on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof is a drill rerun after merge: H.pause, D8, D3.n (docs/probes/waits-audit.md)"
merge:
result: ""
live: ""
```

- [x] 1.1 audit of every wall wait: docs/probes/waits-audit.md
- [x] 1.2 d2dctl.Now and ApproachNPC game-time deadline
- [x] 1.3 d2denv Install, d2dshop uses it
- [x] 1.4 scenario tests and spec delta
- [ ] 1.5 live proof: drill rerun (not run here)
