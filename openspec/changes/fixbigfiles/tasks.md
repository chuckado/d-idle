# Tasks

```yaml
change: fixbigfiles
status: in-progress
agent: fixbigfiles
branch: agent/fixbigfiles
worktree: ~/code/d2lod-wt/fixbigfiles
skip_specs: true
owns:
  - openspec/changes/fixbigfiles/**
  - docs/fixbigfiles.md
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge:
result: "offline; five logs 1.0-1.1 MB -> about 218 KB; hygiene 0 FAIL"
```

- [x] 1.1 cap the five logs
- [x] 1.2 CapLog, d2drill -cap, test
- [x] 1.3 check.sh probe limit, README rule, docs
