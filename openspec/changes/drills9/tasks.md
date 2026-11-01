# Tasks

```yaml
change: drills9
status: in-progress
agent: drills9
branch: agent/drills9
worktree: ~/code/d2lod-wt/drills9
skip_specs: true
owns:
  - openspec/changes/drills9/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge:
result: ""
live: ""
```

- [ ] 1.1 fixstall4: H.pause x3 live
- [ ] 1.2 D10 full drill, two runs
- [ ] 1.3 harness: group-run isolation (D1.P1s, H.obj.chest, D-sync)
- [ ] 1.4 regression pass, one table
