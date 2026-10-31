# FIXPRICE: the bot's price helpers follow game.exe

> status: current (agent FIXPRICE, 2026-10-31). Source of the formulas: [research/price-trace.md](research/price-trace.md) (V: game.exe disassembly and the emulator, model equal on 100,000 cases). Shared code: the formulas are game rules, not d2d-only.

Code: `internal/data/price.go` (`SellPrice`, `RepairPrice`, `BuyPrice`, `VendorBuyMult`, `MaxBuy`, `ActOfArea`). Tests take their cases from the verified model.

| Finding | Change | Test |
|---|---|---|
| B1 `town.repairEstimate` /4 | cost x (max-dur)/max x 128/1024, armor x defence/maxac, unidentified and ethereal 0, whole skipped; floor of 1 when anything is damaged | `TestFixPriceRepairEstimate`, `TestRepairPrice` (D3.r: 19) |
| B2 `loot.SellValue` /4 x guessed factors | cost/2, ethereal /4, class /4, identified-only extras, max buy cap (`MaxBuyOf`), min 1; new signature (db, base, item, maxBuy) | `TestSellValueFollowsGameFormula`, `TestSellPrice` |
| B3 `leveler.estimatePrice` | quality heuristic only when identified, armor defence scaling, `shopAt` applies the vendor mult | `TestEstimatePriceFollowsGameRules` |
| B4 potion prices | `resource.PricesAt(npc, questA)`; `DefaultPrices` stays the 1.0 price | `TestPricesAtVendorMult` |
| B5 `StashGoldCap` | UNVERIFIED tag dropped (FUN_00623460) | none (comment) |
| B6 purse cap | no change: the pickit's `gold` need already is `Gold < clvl x 10,000` (loot.go needs). The cap is tag I (game.exe's own check not traced); a sale over it is U. | none |

## Policy numbers that rested on the old constants (not retuned)
- `loot.poorMinSell` 6, `minValuePerCell` 100, `valuableMinGold` 1500 compare `SellValue`, which was cost/4 x a guessed factor (magic x3, rare x6, unique x8). Now a normal item is worth 2x as before; an unidentified magic or rare one is cost/2 (0.67x and 0.33x of the old value); identified magic, rare, set and unique are a LOWER bound (stat cost unmodelled). Low-gold pickup and valuables will shift: review the three knobs.
- `town.*` repair thresholds: the gold check in `needRepair` now needs about half the gold it did.
- Leveler budget: vendor stock at Charsi is 6% cheaper, Gheed 6% dearer, Act 5 2x than estimated before.

## Unmodelled (stay UNVERIFIED, no guess)
Stat cost of identified magic, rare, crafted, set, unique items (set/unique rows), staff skill mods, missing charges, socketed item cost/2, throwing-weapon stack terms, reduced-price stat, the item's own defence for ground armor (SellValue takes the base maximum: an upper bound), quantity of stacks in SellValue (1).

## Live proof expected (drill D3.r rerun, another agent)
Short Staff at Charsi, 1/20 durability: `repairEstimate` logs 19 and the price paid is 19 (before: estimate 39). Armor pieces: estimate = cost x defence/maxac x missing / 8, paid within the unmodelled terms (a magic item pays more than the estimate, never less).
