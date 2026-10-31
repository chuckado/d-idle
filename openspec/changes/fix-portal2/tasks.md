# Tasks

```yaml
change: fix-portal2
status: in-progress
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
result: ""
```

- [x] ctx.Err() guard in every try and wait
- [x] click only when idle (bounded wait)
- [x] success only by portal object 59; cut cast retries with the next source
- [x] re-read source and charges before each try
- [x] cause in each failure line
- [x] scenario tests and spec delta
- [ ] live proof: drill D2 (m1-liveproof)
