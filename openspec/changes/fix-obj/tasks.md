# Tasks

```yaml
change: fix-obj
status: in-progress
agent: fixobj
branch: agent/fixobj
worktree: ~/code/d2lod-wt/fixobj
owns:
  - internal/d2dsrc/items.go
  - internal/d2dsrc/source.go
  - internal/d2dsrc/source_test.go
  - internal/d2dinput/controller.go
  - internal/d2dinput/controller_test.go
  - docs/fixobj.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... plus go test -race ./internal/d2dinput ./internal/d2dsrc ./internal/d2dctl ./internal/d2denv ./cmd/d2bot pass on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof is a drill rerun after merge: H.obj, H.pause, D8, D3.n (docs/fixobj.md)"
merge:
result: ""
live: "verified live 2026-10-31 DRILLS5 (docs/probes/drills/2026-10-31-drills5-ffight.txt, docs/probes/drills/2026-10-31-drills5-h.txt, docs/probes/drills/2026-10-31-drills5-d8.txt, docs/probes/drills/2026-10-31-drills5-d3.txt): H.obj PASS (the bot Interact on the waypoint object opens the panel), H.obj.chest PASS (class 139 listed, mode NU to ON) and H.obj.shrine PASS (class 81, NU to ON), H.obj.cage PARTIAL (class 26 listed and Interact accepted; the cage mode needs the Cain quest state, effect UNVERIFIED), D8.c PASS (the bot ClaimRogue under the game-time Pause grants the Rogue), D8.a PASS (revive), D3.n PASS (5 NPC windows), D3.b PASS (one shop run on lockstep); H.pause FAIL: the game-time deadline works (the failures carry ticks 396 to 434 and 0 wall), but the walk to Kashya from the camp arrival cell stalls and needs 396 to 950 ticks (16 to 38 s) against the callers 15 and 20 s, see docs/probes/drills/README.md DRILLS5 finding F1 (a follow-up card, not this one)"
```

- [x] 1.1 d2dsrc lists placed objects near the player from debug presets
- [x] 1.2 d2dinput Interact for a listed object sends cmd interact <index>
- [x] 1.3 d2denv Pause steps under lockstep, wall sleep otherwise
- [x] 1.4 scenario tests, spec delta, docs/fixobj.md
- [ ] 1.5 live proof: drill rerun (not run here)
