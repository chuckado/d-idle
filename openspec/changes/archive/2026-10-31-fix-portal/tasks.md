# Tasks

```yaml
change: fix-portal
status: archived
agent: fixportal
branch: agent/fixportal
worktree: ~/code/d2lod-wt/fixportal
owns:
  - cmd/d2bot/d2dwire.go
  - cmd/d2bot/d2dportal_test.go
  - cmd/d2bot/boss.go
  - internal/bot/resource.go
  - internal/bot/resource_portal_test.go
  - internal/combat/portal.go
  - internal/combat/portalpick_test.go
  - internal/combat/fighter.go
  - internal/combat/resource.go
  - docs/fixportal.md
exclusive: []
gates:
  - "go test -p 1 ./internal/combat ./internal/bot ./cmd/d2bot pass including TestPortalSourceSkipsEmptyTome, TestPortalSourceNoneForEmptyTomeAlone, TestHasTPCountsCharges, TestD2DPortalEmptyTomeNoSourceNoClick"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: 9630246
result: "offline only, NOT verified live. combat.PortalSource/PortalQty, d2dPortal and HasTP fixes: an empty tome is no longer clicked or counted. Open: context-canceled portal failures and charged clicks that end in no portal; drills D2 and DECOMPTOME decide."
```

- [x] shared PortalSource/PortalQty, used by combat, d2dPortal, bot resInput, boss view
- [x] scenario tests (empty tome + scrolls, empty tome alone, HasTP, picker agreement)
- [x] spec delta backend-contract
- [x] docs/fixportal.md (further causes and the log line added)
