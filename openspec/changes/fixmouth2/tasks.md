# Tasks

```yaml
change: fixmouth2
status: in-progress
agent: fixmouth2
branch: agent/fixmouth2
worktree: ~/code/d2lod-wt/fixmouth2
skip_specs: true
owns:
  - internal/combat/fight.go
  - internal/combat/fighter.go
  - internal/combat/fixmouth2_test.go
  - docs/fixmouth2.md
exclusive: []
gates:
  - "live: DW.m3 x3 and DW.m4 rerun, no bounce, Clear finishes (docs/fixmouth2.md)"
merge: ""
result: "offline tests pass; live proof FAILED (drills17): bounces remain, see the DRILLS17 section of docs/probes/drills/README.md"
live: "verified live (drills17, hidden slot 0, lab build cheats3): (a) arrival flag set on the level change and cleared after the first step-away (walk starts 28.7 from the exit), so later walks run at mouthR 9 and NearMouth is off (mouthGuard off): FAIL; (b) the evade cap never fired (no spin: 14 step-aways per run, was about 25k); DW.m3 x3 identical 3 bounces in 4 entries, DW.m4 2 bounces in 4 (explore); the Clear phase finishes (no 200 s TTL)"
```

- [x] decide the change from the DRILLS16 geometry
- [x] arrival held for the stay, NearMouth 12, evade cap
- [x] table tests and a Clear sim
- [ ] live proof (DW.m3 x3, DW.m4)
