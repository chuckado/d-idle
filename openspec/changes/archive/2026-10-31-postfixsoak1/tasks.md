# Tasks

```yaml
change: postfixsoak1
status: archived
agent: soakpost
branch: agent/soakpost
worktree: ~/code/d2lod-wt/soakpostdocs
skip_specs: true
owns:
  - docs/probes/postfix-soak1.md
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0"
merge: 853904a
result: "2 h hidden d2d soak on main f8c1b7f, uncheated, one injected d2d kill at 59 min: back in game in 2 s, restarts +1, saved XP not lower; 0 deaths, clvl 13, 20.9 virtual h; verified live (see docs/probes/postfix-soak1.md)"
```

- [x] build and run the soak
- [x] inject the fault and measure recovery
- [x] analysis and doc
