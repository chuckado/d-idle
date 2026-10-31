# Town basics specification

## ADDED Requirements

### Requirement: Prices follow game.exe's formulas
The bot's repair, sell and buy price helpers SHALL follow the vendor price function of game.exe 1.14d (docs/research/price-trace.md): repair at cost x missing durability / 8 (armor scaled by defence), none for unidentified or ethereal items; sell at cost / 2, a quarter more off for ethereal and class items, capped at the NPC max buy; buy at cost x the vendor multiplier. Terms the bot cannot read SHALL stay unmodelled, making the result a lower bound.

#### Scenario: Live repair price
- **WHEN** a Short Staff of base cost 168 with 1 of 20 durability is estimated for repair
- **THEN** the price is 19 gold, the amount drill D3.r paid at Charsi

#### Scenario: Sell value
- **WHEN** a normal weapon of base cost 1000 is valued, and again as an ethereal copy
- **THEN** the values are 500 and 125, and never above the NPC max buy

#### Scenario: Vendor multiplier
- **WHEN** an item of base cost 100 is priced at Charsi, Gheed and an Act 5 vendor
- **THEN** the prices are 93, 106 and 200
