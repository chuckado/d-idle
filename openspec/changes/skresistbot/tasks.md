# Tasks

```yaml
change: skresistbot
status: in-progress
agent: skresistbot
branch: agent/skresistbot
worktree: ~/code/d2lod-wt/skresistbot
skip_specs: true
owns:
  - openspec/changes/skresistbot/**
  - docs/skresistbot.md
  - internal/combat/rules.go
  - internal/combat/target.go
  - internal/combat/skresist_test.go
exclusive: []
gates:
  - "go test -p 1 ./internal/combat passes; whole-repo vet and test on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; tools/ci/check.sh"
result: "B1 floor, B2 pierce, B4 comment done with table tests from the trace rules; B3 left (S); not live-verified, none owed"
```

- [x] B1 floor -100
- [x] B2 pierce in EffRes / DamageFactor / ImmuneWorkaround
- [x] B3 reviewed, left (trace tag S)
- [x] B4 comment
- [x] tests and docs/skresistbot.md
