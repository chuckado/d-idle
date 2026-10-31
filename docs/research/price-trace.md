# Vendor prices: repair, sell, buy and the gold caps (game.exe 1.14d vs d2d vs the bot)

> status: current (agent DECOMPREPAIR, Opus, 2026-10-31; decomp-queue item 5). Proof and tooling: [pr-price-proof.md](pr-price-proof.md), `tools/emu/price_oracle.py`, `price_model.py`, `price_terms.py`.

Tags: V = verified (disassembly plus the emulator running game.exe's own function), I = read from disassembly only, U = unverified. Addresses are game.exe 1.14d.

## The price function (V)

`FUN_0062efb0(ecx player, edx item; difficulty, quest data, npc hcIdx, mode)`: mode 0 buy, 1 sell, 2 gamble, 3 repair. It builds three bases from items.txt `cost` (item record +0xe0) at once: B (buy), S (sell), R (repair), then applies the same steps to each. `price_model.py` re-implements it as written below and equals game.exe on 100,000 random cases (`model 100000/100000`: all item kinds except body parts and tomes, qualities 1-8, identified or not, ethereal, every durability state, stack sizes, all 17 vendors, 3 difficulties). Outside that space (no stat list in the oracle): magic stat costs, staff mods, charges, sockets, quests, reduced prices are I.

1. Base: B = S = R = cost. Armor (ItemTypes 0x32 "armo" via `FUN_00629bb0`): all three = cost x defence / maxac (`0x62f0f2`; the item's base defence stat 31, items +0xd0 maxac). Quiver: B = S = cost x qty / 1024, R = cost x maxstack / 1024. A stacking weapon divides its quality extras by its max stack (`[ebp-0x18]`). Ear: level x cost; body parts: + monster row x 8 (not traced further).
2. Quality extras, ONLY when identified (item flag 0x10, `0x62f13c`): low quality -base/2; set (`FUN_00483440`, SetItems cost mult/add) and unique (`FUN_00483470`, UniqueItems cost mult/add) mult x base / 1024 + add; magic: prefix 1 + suffix 1, rare / crafted: all 6 affixes (MagicPrefix multiply / add; every MagicSuffix row is 0 in 1.14d); magic, rare, crafted, superior also add `FUN_00628e70` (cost from the item's stats through Skills / ItemStatCost, I), and `FUN_0062edd0` adds staff mods (stat 107 skills: Skills cost mult / add x (2 x lvl - 1), I). An automagic affix (data +0x36) adds the same way.
3. Sockets: each socketed item adds its cost / 2 to B, S and R (`FUN_006292f0`, I).
4. Sell only: ethereal / 4 (`0x62f888`); class-specific item (ItemTypes Class < 7, `FUN_0062c060`) / 4 again.
5. Mode 1 (sell): an ethereal item with durability at 0 sells for 0 (then the floor 1). Mode 3 (repair, `0x62f9c2`): not quiver, not throwable, has durability (`FUN_00629930`: items durability > 0, nodurability 0, max > 0, not indestructible stat 152): R = (max - dur) x R / max, 0 when whole; with stat 252 (replenish) R = (max - 1) x R / max whatever the durability (I).
6. NPC (`FUN_00656900`, npc.txt): B x "sell mult" / 1024, S x "buy mult" / 1024, R x "rep mult" / 1024; then per quest flag A..C done (`FUN_0065c310`) x quest mults (A: buy 922, the -10% after the act quest). Above 0xffff the base is divided first ((base / 1024) x mult).
7. Quantity (not books, not quivers): B x qty; a repairable stacking weapon: S = maxstack x S - missing x R and R = missing x R; other items S x qty.
8. Mode 3 adds missing charges (`FUN_00628d30` with 10000). Caps: S <= npc "max buy" of the difficulty; reduced prices (stat 87, max 99%) come off B and R (`FUN_00483360`); B, S and R are at least 1.

npc.txt 1.14d (V): sell mult 1024 (Charsi 960, Gheed 1088, Act 5 2048), buy mult 512, rep mult 128 for every vendor; max buy Normal 5,000 / 10,000 / 15,000 / 20,000 / 25,000 by act, Nightmare 30,000, Hell 35,000.

## Q1 repair (V)

`FUN_0062e660` gates it: identified, not ethereal, and either charges below max or ItemTypes Repair with durability or a restockable stack; otherwise the price is 0. Repair All (`FUN_0062fe60`) sums mode 3 over the 13 body slots for items that pass the gate AND are damaged, missing charges or a short stack. Normal identified weapon or armor: **cost x (max - dur) / max / 8** (armor: cost scaled by defence / maxac first). Unidentified and ethereal items cannot be repaired. The server's refusal for no gold was not traced (U); in the drills "repair not confirmed" with gold 0 is the bot's own check (durability unchanged).

## Q2 sell (V)

**S = cost / 2** (x 1/4 ethereal, x 1/4 class item), quality extras only when identified, capped at max buy, at least 1. An unidentified magic item sells for its base / 2 (low quality unidentified: no halving). Stacks: x qty (potions, scrolls, keys per unit).

## Q3 buy (V)

**B = cost x sell mult / 1024 x qty** with the identified extras, -10% per quest A mult, minus reduced prices. Potions / scrolls / keys: cost x qty (hp1 30, tsc 100, isc 80, key 45 at 1024). Tomes: 250 / 200 + 50 / 40 per scroll (PR #127). Gamble: out of scope.

## Q4 gold caps

Stash: `FUN_00623460` returns 2,500,000 (0x2625a0) whatever the level or difficulty (V by disassembly). Purse: clvl x 10,000 (d2d's pickup and withdraw, `loot.hpp`, `commands.cpp`; live D4.cap: 9990 -> 10000, the pile stays, V on d2d; game.exe's own check not traced: I). A sale over the purse cap: d2d clamps (the excess is lost); game.exe's sell handler not traced (U).

## Q5 d2d (`components/rules/rules.hpp` item_price / repair_cost, upstream 0dc4c62a)

Differential, 100,000 cases (`price_terms.py`): buy 39,626, sell 38,939, repair 42,705 differ; weapons of normal quality identified match. Deviations, each proven by switching one term:

| # | term | game.exe | d2d | cases that one term explains |
|---|---|---|---|---|
| D1 | armor defence | cost x def / maxac | cost | buy 19,122, sell 12,395, repair 13,525 |
| D2 | class item sell | / 4 | none | sell 9,762 |
| D3 | unidentified | no quality extras | extras applied | buy 2,939, sell 2,520 |
| D4 | unidentified repair | refused (0) | repaired | repair 2,736 |
| D5 | undamaged repairable | 1 | 0 | repair 10,841 (Repair All filters it; single repair hover only) |
| D6 | stack quantity | B x qty for stackables; throw stacks S = max x S - missing x R | no multiply / missing x x / max | sell 35, repair 213; the rest in "several" |
| D7 | ethereal at 0 durability | sells 0 -> 1 | full | sell 107 |
| D8 | stats, staff mods, sockets in repair | added (I) | none (`ponytail: no socket terms`) | not in the oracle |

The rest (buy 17,565, sell 14,049, repair 15,390) are cases with two or more terms at once. d2d's repair of normal identified weapons matches (the D3.r live price 19 for the Short Staff at Charsi: game.exe and d2d both 19).

## Q6 the bot

- B1 `internal/town/chores.go:187` repairEstimate = cost x missing% / 4: game.exe and d2d charge cost x missing / max / 8 (rep mult 128), so the "lower bound" is about TWICE the price (Short Staff 95%: estimate 39, paid 19, drill D3.r); a character with gold between the two skips a repair it can pay. Use cost / 8 (armor: x def / maxac), 0 for unidentified or ethereal items, 1 for whole ones.
- B2 `internal/loot/loot.go:373` SellValue = cost / 4 x quality factor: the vendor pays cost / 2 for a normal item, / 8 for a class item or ethereal, / 32 for an ethereal class item; quality factors apply only when identified; capped at max buy (5,000 in Act 1 Normal).
- B3 `internal/leveler/score.go:329` estimatePrice multiplies quality factors on unidentified items too and ignores the vendor mult (Act 5 x 2) and armor defence scaling.
- B4 `internal/resource/potion.go:173` DefaultPrices are cost x 1024 / 1024: right at Akara, 6% low at Gheed, 6% high at Charsi, half at Act 5 vendors (x 2); -10% after quest A.
- B5 `internal/town/stash.go:126` StashGoldCap 2,500,000: correct (FUN_00623460, V); the UNVERIFIED tag can go.
- B6 the pickit does not know the purse cap clvl x 10,000 (drills F5): piles over it stay on the ground.

Table the bot should use (gold, 1.14d, Normal, before quests): repair = cost x missing / max / 8 at every smith (Charsi, Fara, Hratli, Halbu, Larzuk); sell = cost / 2 (x 1/4 class, x 1/4 ethereal, identified extras), max 5,000 Act 1 .. 25,000 Act 5; buy = cost x {Akara, Lysander..Asheara, Jamella, Halbu 1.0; Charsi 0.9375; Gheed 1.0625; Malah, Larzuk, Drehya, Nihlathak 2.0} x qty.

## PR candidate state

Deviations D1-D7 proven (100k differential, model equal to game.exe 100,000 / 100,000); local tooling commit `f38cb9af` on branch `trace/price` (upstream 0dc4c62a) in the worker host's PR checkout. Fix commits, upstream tests, pr_guard and the independent verifier: NOT done (time box). Next: one claim per commit, D1 first (largest), then D2 + D7 (sell), D3 + D4 (identified gate).
