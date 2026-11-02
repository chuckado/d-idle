# Tasks

```yaml
change: kitstall
status: in-progress
agent: kitstall
branch: agent/kitstall
worktree: ../d2lod-wt/kitstall
skip_specs: true
owns:
  - docs/kitstall.md
  - openspec/changes/kitstall/**
exclusive: []
gates:
  - "go test -p 1 ./internal/nav ./internal/gear ./internal/town; changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Dead-gate memory, caster bow no longer a weapon problem, ammo chore Amazon-only; offline tests pass. Not verified live; how the bow got equipped is U (docs/kitstall.md)."
live: "not verified"
```

- [x] timeline from the bot log
- [x] dead-gate memory in nav with test
- [x] gear and town ammo fixes with tests
- [x] docs/kitstall.md with the open U items and the read-only probe
