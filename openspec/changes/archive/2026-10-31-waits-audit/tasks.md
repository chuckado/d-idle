# Tasks

```yaml
change: waits-audit
status: archived
agent: waitsaudit
branch: agent/waitsaudit
worktree: ~/code/d2lod-wt/waitsaudit
owns:
  - internal/d2dctl/town_verbs.go
  - internal/d2dctl/waits_test.go
  - internal/d2denv/env.go
  - internal/d2denv/waits.go
  - internal/d2denv/waits_test.go
  - cmd/d2dshop/main.go
  - docs/probes/waits-audit.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... plus go test -race on d2dctl, d2denv, cmd/d2dshop pass on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof is a drill rerun after merge: H.pause, D8, D3.n (docs/probes/waits-audit.md)"
merge: 1430083
result: "live PASS except H.pause (F1) which is tracked by fixstall"
live: "verified live 2026-10-31 DRILLS5 (docs/probes/drills/2026-10-31-drills5-d8c.txt, docs/probes/drills/2026-10-31-drills5-d8.txt, docs/probes/drills/2026-10-31-drills5-h.txt): the bot game-time Pause and Now (LateClock.Install) drive ClaimRogue (D8.c PASS), the Akara and Warriv menus (D8.q, D8.w PARTIAL PASS) and the approach deadline (H.pause: the failures carry game ticks, 0 wall); H.pause FAIL only because the approach to Kashya from the camp arrival cell needs 16 to 38 s (finding F1 in the README DRILLS5 section)"
```

- [x] 1.1 audit of every wall wait: docs/probes/waits-audit.md
- [x] 1.2 d2dctl.Now and ApproachNPC game-time deadline
- [x] 1.3 d2denv Install, d2dshop uses it
- [x] 1.4 scenario tests and spec delta
- [ ] 1.5 live proof: drill rerun (not run here)
