# Tasks

```yaml
change: fixmouth
status: in-progress
agent: fixmouth
branch: agent/fixmouth
worktree: ~/code/d2lod-wt/fixmouth
skip_specs: true
owns:
  - internal/nav/walk.go
  - internal/nav/nav.go
  - internal/nav/fixmouth_test.go
  - internal/combat/combatspin_test.go
  - docs/fixmouth.md
exclusive: []
gates:
  - "live: DW.m3 x3 and DW.m4 rerun, no bounce (docs/fixmouth.md)"
merge: ""
result: "offline tests pass; live proof FAILED (drills16): the arrival step-away works, bounces remain from other walks"
live: "verified live (drills16, hidden slot, lab build cheats3): the arrival guard fired at radius 16 on every first walk (dist 11.2) and the Explore first walk stepped away (no bounce there); DW.m3 x3 identical: 23 level-change errors from Fighter.move (engage approach and evade), never finishes; DW.m4: 2 explore bounces of 4 entries from roam first walk 13 subtiles from the exit after arrival was consumed; the Cave Level 1 step-away click did NOT fire the trigger (U resolved)"
```

- [x] decide the option (first-walk arrival radius) from the logged geometry
- [x] arrival state in nav walk, mouthStep radius, NearMouth after arrival
- [x] table test, roam test, Fighter.move test
- [x] live proof run (DW.m3 x3, DW.m4): FAIL, see docs/probes/drills/README.md DRILLS16
