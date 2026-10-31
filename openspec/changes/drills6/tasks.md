# Tasks

```yaml
change: drills6
status: in-progress
agent: drills6
branch: agent/drills6
worktree: ~/code/d2lod-wt/drills6
skip_specs: true
owns:
  - openspec/changes/drills6/**
  - docs/probes/drills/2026-10-31-drills6*
  - internal/drill/drun.go
  - cmd/d2drill/run_lab.go
exclusive:
  - instance:slot0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge:
result: ""
live: "H.pause PARTIAL, D6.d PARTIAL, D-sync PASS x3, D-run PASS x3, D2 PASS in subsets (full group order-dependent), D-skill not run"
```

- [x] 1.1 fixstall: H.pause and D6.d live (PARTIAL, findings F7 and F8)
- [x] 1.2 fixmisc: D-sync live (PASS)
- [x] 1.3 fixrun: D-run live (PASS, the bot runs after a fresh game)
- [x] 1.4 fix-portal2: D2 group live
- [ ] 1.5 D-skill P26 observables (recipe not on main; not run)
