# Tasks

```yaml
change: lockstepp1p2
status: in-progress
agent: lockstepp1p2
branch: agent/lockstepp1p2
worktree: d2lod-wt/lockstepp1p2
skip_specs: true
owns:
  - docs/research/lockstep-p1p2.md
  - openspec/changes/lockstepp1p2/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "preparation only: proxy built and tested offline under Wine; no hidden instance was running and agents may not start one, so P1 and P2 live runs are NOT done"
live: "no game driven"
```

- [x] hook repo read, proxy DLL built against the game's own ijl11.dll copy, offline test suite passes
- [ ] P1 live read (needs a running hidden instance started by the main agent path)
- [ ] P2 k=1 control, k=2 run
