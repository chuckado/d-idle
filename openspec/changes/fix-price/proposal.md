## Why

The decompilation of game.exe's price function (docs/research/price-trace.md, model equal to game.exe on 100,000 emulator cases) showed wrong constants in the bot's economy helpers: the repair estimate is twice the real price (estimate 39, paid 19 in drill D3.r), the sell value uses a quarter instead of a half and ignores class, ethereal and identified terms, the leveler's buy estimate ignores the vendor multiplier and armor defence, and the potion prices ignore the vendor. These are correctness bugs, not policy.

## What Changes

- Shared (game rules): `internal/data/price.go` holds the verified repair, sell and buy formulas, vendor multipliers, max buy and act tables.
- `town.repairEstimate`, `loot.SellValue` (new signature), `leveler.estimatePrice` plus the vendor multiplier in shopAt, `resource.PricesAt` use them. StashGoldCap loses its UNVERIFIED tag.
- No knob or default changes. Policy numbers that depended on the old constants: docs/fixprice.md.
- Delta: town-basics ADDED requirement.
