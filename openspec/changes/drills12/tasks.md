# Tasks

```yaml
change: drills12
status: in-progress
agent: drills12
branch: agent/drills12
worktree: ~/code/d2lod-wt/drills12
skip_specs: true
owns:
  - openspec/changes/drills12/**
exclusive:
  - instance:slot0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
result: "verified live: D-portalmelee PASS x3 (3 zombies within 8 subtiles at 102/300 life, no potion, one scroll: Safety chicken, scroll kept 1); control PASS x3. The harness cannot run the chicken (it leaves the game), so only the decision is observed"
live: "D-portalmelee PASS x3 (decision chicken, scroll 1->1), D-portalmelee.ctl PASS x3 (town-portal, scroll 1->0, portal open)"
```

- [x] 1.1 drill waits for 3 within combat.crowdAdjacent (the knob)
- [x] 1.2 D-portalmelee x3 and the control x3 live
- [x] 1.3 README DRILLS12, primitives rows, live lines on fixcrowd and fixportalmelee

Ownership: no docs/OWNERS line beyond the log glob; the harness files are covered by the standing drills globs of m1-liveproof.
