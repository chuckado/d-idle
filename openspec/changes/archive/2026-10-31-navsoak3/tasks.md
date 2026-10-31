# Tasks

```yaml
change: navsoak3
status: archived
agent: navsoak3
branch: agent/navsoak3
worktree: ~/code/d2lod-wt/navsoak3
owns:
  - docs/probes/navsoak3/**
exclusive: []
gates:
  - "the soak is registered (tools/agentreap.sh --list shows navsoak3) and its logs advance"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: 64e6be9
result: soak finished clean (exit 0, clvl 13, 9406 kills, 0 deaths, pre-fix baseline for the primitives); numbers in docs/probes/navsoak3-summary.md; verified live on the hidden d2d, analysis read-only
```

- [x] Build d2bot from origin/main 35ce4c2 to a new binary and start the soak with the supervisor
- [x] Verify healthy for about 20 minutes
- [x] Write docs/probes/navsoak3/README.md
