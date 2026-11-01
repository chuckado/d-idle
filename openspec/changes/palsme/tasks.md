# Tasks

```yaml
change: palsme
status: in-progress
agent: palsme
branch: agent/palsme
worktree: ~/code/d2lod-wt/palsme
skip_specs: true
owns:
  - docs/palsme.md
  - openspec/changes/palsme/**
exclusive: []
gates:
  - "gofmt, go vet, go test -p 1 per touched package, whole repo on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Zealot-lite table revised per the SME; install legality check run on the worker host and passes. Policy items recorded as ready cards only. Nothing run live."
live: "offline only"
```

- [x] Zealot-lite table (Prayer 1, Might, Defiance 1, Holy Fire 2)
- [x] legality tests, install check on the worker host
- [x] docs/palsme.md with ready cards
