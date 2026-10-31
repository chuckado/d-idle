# Tasks

```yaml
change: fix-price
status: archived
agent: fixprice
branch: agent/fixprice
worktree: ~/code/d2lod-wt/fixprice
merge: c43dd3b
owns:
  - internal/data/price.go
  - internal/data/price_test.go
  - internal/town/chores.go
  - internal/town/stashfull.go
  - internal/town/shop.go
  - internal/town/fixprice_test.go
  - internal/loot/pickup.go
  - internal/loot/valuables.go
  - internal/loot/fixprice_test.go
  - internal/leveler/score.go
  - internal/leveler/shop.go
  - internal/leveler/shopkit_test.go
  - internal/leveler/fixprice_test.go
  - internal/resource/potion.go
  - internal/resource/fixprice_test.go
  - internal/town/stash.go
  - docs/fixprice.md
exclusive: []
gates:
  - "go test -short -count=1 ./... passes including the FixPrice tests on the worker host"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "verified live 2026-10-31 DRILLS5; see the live: line"
# note: owns of internal/loot/loot.go released to fix-belt2 (FIXBELT2, code already merged).
live: "verified live 2026-10-31 DRILLS5 (docs/probes/drills/2026-10-31-drills5-repair.txt, docs/probes/drills/2026-10-31-drills5-d3.txt): D3.r repair estimate before the visit 19 = 19 gold paid (Staff 19 of 20 lost; before the fix the estimate was 39), D3.b2 sold an identify scroll for 40 = the price formula 40, D3.u an unidentified magic Cap sold for 32 = the formula for the base (cost/2), each twice, identical"
```

- [x] B1 repairEstimate, B2 SellValue, B3 estimatePrice and vendor mult, B4 PricesAt, B5 stash cap tag
- [x] B6 analysed: the pickit already compares gold to the purse cap (need "gold"); no change
- [x] table-driven tests and spec delta
- [x] live proof: drill D3.r rerun by another agent
