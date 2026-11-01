# Tasks

```yaml
change: fixmouth3
status: archived
agent: fixmouth3
branch: agent/fixmouth3
worktree: ~/code/d2lod-wt/fixmouth3
skip_specs: true
owns:
  - internal/nav/roam.go
  - internal/nav/fixmouth3_test.go
  - internal/combat/fixmouth3_test.go
  - docs/fixmouth3.md
exclusive: []
gates:
  - "live: DW.m3 x3 and DW.m4 rerun, zero level-change errors in Clear and Explore (docs/fixmouth3.md)"
merge: db32c75
result: "den bounce closed live in DRILLS19 (DW.m3 0 bounces x3, DW.m4 0); earlier iterations improved it; trigger bound and long-click behaviour remain UNVERIFIED as principles"
live: "verified live (drills18, hidden slot 0, build of 2a9234a plus main, lab build cheats3): stayLevel set at the first step-away of every stay and held to the level change (V: stay=true on every later mouth check); DW.m3 x3 identical: 4 entries, 1 real bounce to Blood Moor (Fighter.move target (37,150) 15.1 from the exit, step-away replanned from 3.2 after a basic-attack fight pulled the character to (36,138), first click (37,140) fired) plus 1 town-portal escape from a pocket stall, 18 level-change errors (16 are one stale move retried after the escape); DW.m4 0 bounces in 4 entries; Clear finishes; the single long click: U (no step-away click fired in the clean cases, the failing one was a replan with a normal click); evade cap never fired; D2 11/11 D3 12/12 | drills19 (fixmouth4 build, tip 80bf381, lab build cheats3): DW.m3 x3 identical, 0 level-change errors, 0 bounces in 4 entries; DW.m4 x3 0 bounces, 0 level-change errors; the same DRILLS18 chain (step-away replan from 3.2) was caught by the gate; D2 11/11 D3 12/12; live defect: a held gate livelocks Den cycle 2 (117 holds, 117 step-aways), see DRILLS19"
```

- [x] root cause from the DRILLS17 logs (docs/fixmouth3.md)
- [x] stay guard, trigger radius, away-path check, one-click step-away, roam target filter
- [x] table tests and an Explore sweep sim
- [x] live proof (DW.m3 x3, DW.m4): ran in drills18, FAIL (one bounce per DW.m3 run, see README DRILLS18)
