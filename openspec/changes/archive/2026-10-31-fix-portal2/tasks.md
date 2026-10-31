# Tasks

```yaml
change: fix-portal2
status: archived
agent: fixportal2
branch: agent/fixportal2
worktree: ~/code/d2lod-wt/fixportal2
owns:
  - cmd/d2bot/d2dwire.go
  - cmd/d2bot/d2dportal_test.go
  - docs/fixportal.md
exclusive: []
gates:
  - "go test -p 1 ./cmd/d2bot passes including TestD2DPortalDeadContextNoClick, TestD2DPortalBusyWaitsForIdle, TestD2DPortalStaysBusyNoClick, TestD2DPortalCutCastRetriesNextSource, TestD2DPortalEmptyAfterRereadNoRetry"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: agent/fixportal2 (code merged earlier; live proof closed by DRILLS7)
result: "verified live: D2 group 11 of 11 in one process twice (DRILLS7); see live:"
live: "drill D2 2026-10-31 (docs/probes/drills/2026-10-31-d2.txt): D2.a verb PASS (1 click, 1 portal, 1 source, 32 ticks), D2.b empty tome PASS (0 clicks, 0 keys, ErrNoPortal at tick 0), D2.c PASS, D2.e busy PASS (waits: click +37 vs +22 idle), D2.h one pair PASS; D2.d cut cast BLOCKED (needs debug hit); the walk-in after the verb FAILS (D2.t, D2.g: town.ToTown never enters the portal, finding W1), so no fix-portal2 defect, a new fix card is needed | DRILLS5: D3.tp PASS unchanged (docs/probes/drills/2026-10-31-drills5-d3.txt); the D2 group was not rerun this round, the earlier results above stand | DRILLS6 2026-10-31: D2.a t a2 b c g d d3 e h f all PASS run as subsets (d3: 2 clicks, 0 scrolls left; e: click +37 vs +22); the full group in one process is order-dependent FAIL (D2.a2 stale portal, D2.d3 DIFF, character dead for e h f: harness order, not a portal defect) | DRILLS7 2026-10-31 (docs/probes/drills/README.md DRILLS7): the full D2 group passes 11 of 11 in ONE process, twice (harness fixes ensureAlive and longer stale-portal walk-away)"
```

- [x] ctx.Err() guard in every try and wait
- [x] click only when idle (bounded wait)
- [x] success only by portal object 59; cut cast retries with the next source
- [x] re-read source and charges before each try
- [x] cause in each failure line
- [x] scenario tests and spec delta
- [x] live proof: drill D2 (DRILLS7: full D2 group 11 of 11 in one process, twice)
