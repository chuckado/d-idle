# Tasks

```yaml
change: fix-belt
status: in-progress
agent: fixbelt
branch: agent/fixbelt
worktree: ~/code/d2lod-wt/fixbelt
owns:
  - internal/town/beltfill.go
  - internal/town/beltfix_test.go
  - internal/town/chores.go
  - internal/town/sim/sim.go
  - internal/bot/goal.go
  - internal/bot/kitgate.go
  - internal/bot/kitgate_test.go
  - internal/bot/observe.go
  - internal/combat/safety.go
  - internal/combat/beltbox_test.go
  - internal/loot/loot.go
exclusive: []
gates:
  - "go test ./internal/town ./internal/combat ./internal/bot passes including TestBeltPutItemNotFoundDoesNotStick, TestBeltPutRespectsColumnKind, TestShopPurchaseIsBelted, TestBeltOfSeesEveryBox, TestKitCountsPackPotions"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge:
result:
```

## 1. Work

- [x] 1.1 F1 sticky mark: item-not-found never sets it, retry after a fresh read, reset on new game and on pack change
- [x] 1.2 F3 belt step after the shop
- [x] 1.3 F7 BeltOf box mapping (box = GY*4+GX, column = box%4)
- [x] 1.4 F8 column kind in the box choice
- [x] 1.5 F6 kit counts pack potions
- [x] 1.6 F9 diagnostics: put line, mark set/clear, minute line packHP/packMP/beltBoxes, potion skip line (portal line: fix-portal)
- [ ] 1.7 changecheck, leakcheck, validate
