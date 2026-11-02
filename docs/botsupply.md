# Bot supply: per-kill rates and where the code reads them

> status: current (card botsupply, 2026-11-02). Offline; not verified live. Source: docs/research/treasure-trace.md (V run: exact class odds, 104,000 native rolls) and docs/research/monster-stats-trace.md (V run: real room populations). One player, MF and GF 0, Normal Act 1 / 2. The numbers describe monsters, so nothing here depends on the class or build.

## Rates (internal/data/supply.go `SupplyOf`)

| kill | drops | gold piles | potions |
|---|---|---|---|
| plain (H2H class) | 0.375 | 0.131 | 0.057 (about 1 in 17 kills) |
| plain caster / archer | 0.375 | 0.1125 | about 0.07 |
| champion | 3 | 0.6, coins x5 | 2 |
| unique | 5 | 0 | 4 |
| superunique | 6 | 0 | 4 |

Scrolls 0.015 per plain kill (Cast classes 0.035, Wraith 0.12), keys 1 in 130 (treasure-trace B5).

## Gold is level dependent

A pile is `ilvl + rand(5 ilvl)` coins, ilvl = monster level (at least 1): `ilvl .. 6 ilvl - 1`, mean `3.5 ilvl - 0.5` (`data.GoldPileMean`). Gold per plain kill = 0.131 x that (`data.GoldPerKill`): Blood Moor 0.39, Dark Wood 2.2, Hell Blood Moor 31. The old `GoldPerKill: 6` was about 15x too high and "gold per kill is the same everywhere" was wrong (B1, B2).

## Where the code reads them

| use | code |
|---|---|
| simulator gold | `resource.world.hit` adds `data.GoldPerKill(KillPlain, Scenario.MonLevel)` (fractions carried) |
| simulator monster life | `resource.BloodMoorLife` (mean 4.17 vs the traced 4.2), `resource.LifeCycle` from rows via `data.DB.MeanHP` |
| gold pile bar | `loot.Looter.goldBar`: `min(loot.goldMinStack, data.PileMinWorth(areaMonLvl, loot.goldPileShare / 100))`; Blood Moor bar 3, ilvl 3 and up stays 10. Poor characters (below `poorGoldBelow`) still take every pile |
| recovery area pick | `bot.areaXP` (traced mean XP per plain monster x level-gap share); `normalMLvl` fallback Den 1, Cave 2 2 |

The Levels.txt monster level is an upper bound of the real Normal unit levels (within about 1), so in Normal the bar errs a little high; in Nightmare and Hell it is exact (monster level = area level).

## Not done

`areaXPTable` covers the areas the trace lists, Normal only; other difficulties keep the 10 x level fit (S). Potion and scroll supply are stated here but no code consumes them yet (a soak judge could). docs/items.md still says "above GoldMinStack (10)" (not our file).
