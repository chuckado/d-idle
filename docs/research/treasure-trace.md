# Treasure classes, drops and gold (Act 1 and 2): game.exe 1.14d vs d2d vs the bot (coverage gap `treasure`)

> status: card treasure, 2026-11-01, offline (no game, no d2d slot). Coverage gap rank 5 (docs/coverage/gaps.tsv): 0 of 852 TreasureClassEx rows compared. This doc says how game.exe picks the class for a killed monster, how the class rolls (picks, NoDrop, players), what that gives per kill in Act 1 and 2 (potions, scrolls, gold, gear, how many items), how quality and gold are made, chests, and what d2d and our bot assume. Model traces: docs/research/sk-mana-trace.md, docs/research/sk-missile-trace.md. The roll itself was already traced and differenced upstream (the d2d port's `docs/research/re/drops.md`, `tools/emu/drops.py`); this card re-ran it on new seeds against the real d2d build, added the part that proof skips (which class a monster gets) and the Act 1 and 2 numbers.

Tags: **V read** disassembled by this agent with the upstream `disasm.py` (addresses given); **V run** game.exe's own code run natively under the emulator, compared with d2d's real code (the pristine build, see section 8) or with the exact odds; **D** from the patched 1.14d tables as game.exe loads them (class records read out of its memory, MPQ text tables); **S** d2d source or the upstream trace `re/drops.md` (named); **I** inference; **U** not verified.

## Summary

- **Class choice (V read + V run).** A kill rolls `FUN_005a6600`: a superunique takes its SuperUniques `TC` for the difficulty; else a champion (monster flag 4) MonStats `TreasureClass2`, a unique (flag 8) `TreasureClass3`, else `TreasureClass1` (column per difficulty: plain, `(N)` Nightmare, `(H)` Hell). `TreasureClass4` replaces it while the killer's quest for `TCQuestId` has none of flag 15, flag 1 or flag `TCQuestCP` (Andariel, Duriel, Mephisto, Diablo, Baal). In an expansion game past Normal the class then moves on by the monster's level inside its group (`FUN_0055afa0` -> `FUN_00654e00`), except for MonStats rows with `noRatio` or `boss`. d2d's `Loot::drop` makes the same choice: **78,309 of 78,309 selection cases equal** (game side V run, d2d side a transcription, S) and the level move equals on 85,200 of 85,200.
- **The roll (V run).** d2d's `roll_drops` equals game.exe on 100,000 rolls over seeds the upstream proof did not use (plus 1,012 of 1,012 class tables, 20,000 chests and other containers, 20,000 stands, 39,975 of 39,980 made items: the 5 are magic elixirs, a known unported path). **No d2d deviation proven in Act 1 or 2; no PR candidate.**
- **What an Act 1 / 2 kill gives (V run, exact odds checked by 104,000 native rolls).** A plain monster (Normal): 62.5 % NoDrop, so 0.375 drops per kill; per kill 0.131 gold piles (0.1125 for casters and archers), 0.035 healing and 0.020 mana potions, 0.015 scrolls, 0.11 gear. A champion always drops 3 items: exactly 2 potions, 0.6 gold piles (x5 coins) and 0.37 gear. A unique drops 5 (4 potions), a superunique 6 (4 potions, 1.87 gear).
- **Gold is area dependent (V run + D).** A pile is `ilvl + rand(5 ilvl)` coins (mean 3.5 ilvl - 0.5), ilvl = the monster level: Blood Moor Normal (ilvl 1) 0.39 gold per kill, ilvl 5 (Dark Wood) 2.2, Hell Blood Moor (area level 67) 31. **Our bot's `GoldPerKill 6` and its "gold per kill is the same everywhere" are wrong (B1, B2).**
- **Bot findings B1-B5**, d2d deviations D1-D8 (all from upstream's own "not ours yet" list or earlier traces, none proven here), drill recipe T1-T5, ledger row, shared-file lines at the end.

## 1. Which class a monster gets (Q1-Q3)

| Address | Rule | Tag |
|---|---|---|
| `FUN_005a6600` (game, dropper, noNoDrop, killer struct) 0x5a6600..0x5a6822 | difficulty `[game+0x6d]`, MonStats row `[0x744304 + 0xa78]` stride 0x1a8 by `[unit+4]`; item level `FUN_00558200` (below) | V read |
| 0x5a669d..0x5a66b9 | `FUN_005a03a0(unit)`: monster data `[unit+0x14] +0x16` bit 2 (superunique) -> word `+0x26` = SuperUniques index, else -1. An index: `FUN_006556e0(index)` row -> class `word [row + diff*2 + 0x2c]`; **no row: MonStats `+0x8a` (TreasureClass3)** instead | V read |
| 0x5a66c0..0x5a66fa | `FUN_005a0180(unit, 4)` (monster data `+0x16` & 4, champion): `word [row + diff*8 + 0x88]` (TreasureClass2); `(unit, 8)` (unique): `+0x8a` (TreasureClass3); else `+0x86` (TreasureClass1). One u16 per difficulty, 8 bytes apart; TreasureClass4 is `+0x8c` | V read |
| 0x5a6705..0x5a67f1 | quest class: needs MonStats `+0x9e` (TCQuestId) > 0, `word [row + diff*8 + 0x8c]` (TreasureClass4) > 0 and a killer struct; the killer unit `[killer+8]` of type 0 (a player) through `FUN_00543520`: `FUN_0065c310(player, TCQuestId, 15)`, `(.., 1)`, `(.., TCQuestCP [+0x9f])`; any answer 1 keeps the plain class, none uses TreasureClass4. A non-player killer (a pet) goes through its owner first (`FUN_00552f60`, `FUN_0058f0d0`) | V read |
| `FUN_0055afa0` 0x55afa0..0x55afef | level argument for `FUN_00654e00(class, lvl)`: `[game+0x70]` (expansion) set, `[game+0x6d]` (difficulty byte) > 0, unit type 1, MonStats row `FUN_00451f80` flag byte `+0xc` has neither mask `[0x6ce270]` (noRatio, 4) nor `[0x6ce280]` (boss, 0x40): lvl = stat 12 (`FUN_00625480(unit, 0xc, 0)`), else 0. Then `FUN_00654e00`, then the roller `FUN_0055a6d0` | V read |
| `FUN_00654e00(index, lvl)` | while the next row has the same non-zero group and lvl >= its level, move to it (S re/drops.md; run: 85,200 of 85,200 equal d2d's `tc_upgrade`) | V run |
| item level `FUN_00558200` | a monster: stat 12 (its level) at least 1; other units: the area's monster level `FUN_0061dca0` | V read |
| monster level (stat 12) | Normal: the MonStats `Level` column; Nightmare and Hell: the area level (Levels.txt `MonLvl{2,3}Ex`); a champion +2, a unique or superunique +3 | S (`tools/emu/boss_level.py`, docs/research/monster-mods-trace.md) |

Minions of a unique carry neither flag 4 nor 8, so they roll TreasureClass1 (I, from the same flag test). The plain classes: MonStats TC1 is `Act N <family> A/B/C` (H2H, Cast, Miss, Wraith), champions `Act N Champ A/B/C`, uniques `Act N Unique A/B/C`; a few rows have their own (`Quill 1..4`, `Swarm 1`, `Cow`, bosses).

**Differential (V run game side, S d2d side).** `tools/emu/tre_select.py`: game.exe's real `FUN_005a6600` on a fake monster unit (class id, flags 4 / 8 / 2, SuperUniques index, player killer, difficulty, expansion; MonStats row, `FUN_00558200`, `FUN_005a03a0`, `FUN_005a0180`, `FUN_006556e0`, `FUN_0055afa0`, `FUN_00654e00` native; hooked only the room and level id, the stat getter, the killer's player data and the quest flag test; the roller `FUN_0055a6d0` records the class record, item level and killer) against a Python transcription of d2d's rule (`Loot::drop` in `components/game/loot.hpp`, `tc_fixed = noRatio || boss`, `tc_upgrade` over the TreasureClassEx group and level columns). Cases: every MonStats row x 3 difficulties x {normal, champion, unique} x 11 levels (1..99), 8 quest-flag states for every row with a `TCQuestId`, every SuperUniques row x 3 x 11. **78,309 cases, 78,309 equal** (24,948 of them rows with no class on either side: nothing rolls); item level equal, killer passed. Sensitivity: swapping champion and unique in the transcription makes 30,426 differ. Upgrade alone: every TreasureClassEx class x lvl 0..99: 85,200 of 85,200 equal. The d2d side is a transcription, not the real code: the real `tc_upgrade` is covered by the 100,000 rolls of section 8, the selection expression only by this transcription (U for a divergence between the source and my reading of it).

## 2. How a class rolls (Q4, Q8)

All from `FUN_0055a6d0` (S re/drops.md; V run 100,000 + 104,000 rolls): picks = |Picks| (at least 1); each pick draws `rand(nd + total)`, below `nd` nothing, else the entry holding `r - nd` (weights = the entry probabilities of the table); a negative Picks draws nothing and takes entry k for pick k (k from 0, stops at the total: Countess, champions, uniques, superuniques); a class entry recurses (modifiers = the max down the chain); at most 6 items per kill (the count goes up before the gold-find step); an item that cannot be placed is not made and not counted (D6).

**NoDrop and players (V run).** `n = near + (players - near) / 2` with near = 1 (the killer's nearby party, 2..8, when there is one), players = game players or `/players`; for n > 1: `p = nd/(nd + total)`, `nd' = trunc(p^n x total / (1 - p^n))`. Native rolls of `Act 1 H2H A` (NoDrop 100, total 60), 20,000 each:

| n | nd' | no drop, formula | no drop, native | drops per kill, native |
|---|---|---|---|---|
| 1 (players 1 or 2) | 100 | 62.5 % | 62.2 % | 0.378 |
| 2 (players 3 or 4) | 38 | 38.8 % | 38.5 % | 0.615 |
| 3 (players 5 or 6) | 19 | 24.1 % | 24.0 % | 0.760 |
| 4 (players 7 or 8) | 10 | 14.3 % | 14.7 % | 0.853 |

An even `/players` count adds nothing (players 2 = 1, 4 = 3 ...). Champions, uniques and superuniques have NoDrop 0 (D).

## 3. Item type and quality (Q5)

The class entries name an item code, an item type class (the auto classes `weap3`..`armo96`: spawnable bases of a type with qlvl in N-2..N weighted by the type's Rarity) or another class. Quality: `FUN_00558640` (S re/drops.md; run through the 100,000 rolls, qualities included in each line): type Normal -> normal; unique / set / rare / magic by `odds = (base - (ilvl - qlvl) / divisor) x 128`, the magic-find diminishing curve (k 250 / 500 / 600), the class's own Unique / Set / Rare / Magic modifiers out of 1024; then superior, normal, low. Quality shares from native rolls (`tre_act.py quality`, 8,000 kills per row; gear, rings, amulets, jewels and charms only; ilvl as written; one player):

| class, ilvl | MF | items | low | normal | superior | magic | set | rare | unique |
|---|---|---|---|---|---|---|---|---|---|
| Act 1 H2H B, 5 | 0 | 885 | 25.8 | 55.5 | 7.5 | 9.4 | 0.7 | 1.1 | 0.1 |
| Act 1 H2H B, 5 | 100 | 885 | 24.5 | 53.1 | 7.1 | 11.8 | 1.2 | 1.7 | 0.6 |
| Act 1 (N) Champ A, 24 | 0 / 100 / 300 | 3,048 | 0 | 0 | 0 | 91.3 / 83.8 / 76.5 | 2.9 / 5.8 / 8.4 | 4.2 / 8.0 / 12.0 | 1.6 / 2.4 / 3.0 |
| Act 1 (H) Super A, 60 | 0 / 100 / 300 | 15,421 | 0 | 0 | 0 | 73.0 / 51.1 / 26.3 | 3.0 / 5.9 / 8.8 | 22.6 / 40.7 / 61.7 | 1.3 / 2.3 / 3.3 |
| Andariel, 12 | 0 / 100 / 300 | 17,642 | 0 | 5.2 | 0 | 55.1 / 31.2 / 10.2 | 14.1 / 25.0 / 37.2 | 19.1 / 27.7 / 32.3 | 6.5 / 10.9 / 14.9 |

Champion, unique and boss classes carry large Unique / Set / Rare modifiers (the 983 / 1024 style numbers of docs/sme/mechanics-items-drops.md section 3 are the Basin's; V run shows the effect above). The SME's 2.3 quality table is the 1.09 page at another ilvl - qlvl and is not comparable (S, not used here).

## 4. Gold (Q6)

`FUN_00557ab0`: coins = `ilvl + rand(5 ilvl)`, at least 1 (ilvl at least 1), so ilvl .. 6 ilvl - 1, mean 3.5 ilvl - 0.5; an entry `gld,mul=N` (champions: N = 1280 in every difficulty here) scales it `coins x N >> 8` (x5); gold find: `coins x (100 + gf) / 100` unless the drop fills the sixth slot (S re/drops.md; the coins are inside the 39,975 equal items of section 8). Mean coins per pile and per plain-class kill (0.131 piles): ilvl 1: 3.0 / 0.39; 2: 6.5 / 0.85; 3: 10 / 1.31; 4: 13.5 / 1.77; 5: 17 / 2.23; 6: 20.5 / 2.69; 10: 34.5 / 4.53; 15: 52 / 6.83; 20: 69.5 / 9.12; 30: 104.5 / 13.7; 67 (Hell Blood Moor): 234 / 30.7 (D, arithmetic on V). A pile of 10 or more is: ilvl 1 never, ilvl 2 20 %, 3 53 %, 4 70 %, 5 80 %, 8 95 %. d2d's fidelity probe (docs/d2d-fidelity.md 3b, 3c) saw 0.13 to 0.14 piles per kill and 2.9 / 5.3 / 13.7 / 18 coins per pile in Blood Moor / Cold Plains / Stony Field / Dark Wood: the exact odds say 0.131 and 3.0 at ilvl 1 (S, consistent).

## 5. What an Act 1 / 2 kill gives (Q7, Q8)

Exact odds from the class records game.exe loaded (class table `0x96c5ec`, entries stride 0x1c), one player, MF and GF 0, `tre_act.py classes`; checked by native rolls (section 8: 13 classes x 8,000 rolls, largest |z| 1.68 over drops, gold, potions, scrolls and gear). Per kill; potions = hp + mp + rv; scrolls = tsc + isc:

| class (monsters) | P(drop) | piles | hp | mp | rv | scroll | key | gear |
|---|---|---|---|---|---|---|---|---|
| Act 1 H2H A/B/C (fallen, zombie, brute, skeleton ...) | 37.5 % | 0.131 | 0.035 | 0.020 | 0.0022 | 0.015 | 0.0076 | 0.110 |
| Act 1 Cast A (shamans, crow nests) | 37.5 % | 0.1125 | 0.030 | 0.039 | 0.0019 | 0.035 | 0.0065 | 0.102 |
| Act 1 Miss A (archers) | 37.5 % | 0.1125 | 0.030 | 0.017 | 0.0019 | 0.013 | 0.0065 | 0.127 |
| Act 1 Wraith B | 27.9 % | 0 | 0 | 0.120 | 0 | 0.120 | 0 | 0 |
| Act 2 H2H A | 37.5 % | 0.131 | 0.036 | 0.015 | 0.0021 | 0.015 | 0.0076 | 0.110 |
| Act 1 (N) H2H B (Nightmare) | 37.5 % | 0.131 | 0.036 | 0.015 | 0.0025 | 0.015 | 0.0076 | 0.110 |
| Act 1 Champ A (any Normal champion) | 100 % (3.0 items) | 0.6 (x5) | 1.29 | 0.59 | 0.12 | 0 | 0 | 0.37 |
| Act 1 Unique A (the unique, not its minions) | 100 % (5.0 items) | 0 | 2.59 | 1.18 | 0.24 | 0 | 0 | 0.94 |
| Act 1 Super A (Bishibosh, Coldcrow ...) | 100 % (6.0 items) | 0 | 2.59 | 1.18 | 0.24 | 0 | 0 | 1.87 |
| Act 1 Chest A / B, per round | 75.4 % (1.18 items) | 0.42 | 0.113 | 0.063 | 0.007 | 0.049 | 0.024 | 0.314 |

Potion tiers (per kill): Normal Act 1 H2H hp1 0.026, hp2 0.009, mp1 0.011, mp2 0.009 (+ stamina 0.009); Normal Act 2 H2H hp1 0.010, hp2 0.016, hp3 0.009, mp2 0.005, mp3 0.009; **every Nightmare and Hell class** hp3 / hp4 / hp5 (0.001 / 0.015 / 0.020) and mp4 / mp5 (0.002 / 0.0125), rvs 0.0025. A Normal champion's two potions are hp1 0.94, hp2 0.35, mp1 0.35, mp2 0.24, rvs 0.12; a Hell champion (Act 1 (H) Champ C): hp3 0.06, hp4 0.49, hp5 0.59, mp4 0.17, mp5 0.44, rvs 0.12, rvl 0.13. Healing potions are 61 % of a plain kill's potions in Act 1 Normal and 67 % in Act 2 (D, arithmetic on the table above).

Bosses (class by difficulty, quest class only until the quest flag): superuniques of Act 1 and 2 use `Act N Super A/B/C` (6 drops: 4 potions, 1.87 gear; Normal Act 1 has no rune entry, Act 2 Normal 0.018 runes) except Griswold and the Smith (3 drops, 2.94 gear, no potions), the Countess (`Countess`, 2 deterministic picks, 5.75 items, **2.28 runes**), Radament (`pk5, NoDrop 19/48`, 3.6 items) and the Summoner (`pk5, 19/36`). Andariel `pk7`, NoDrop 19 against a total of 48: 5.0 items, 2.1 gear, 0.64 potions, 1.15 gold piles, runes 0.045; **her first kill per difficulty rolls `Andarielq` instead (3.8 items, 3.24 gear, no potions, runes 0.07)**; Duriel `pk5`, NoDrop 0: 6 or more items every kill (the cap holds it at 6; category sums: gear 7.0, potions 2.1, runes 0.15), first kill `Durielq` (gear 10.8, no potions, runes 0.24). In `tre_act.py` the "E[drops]" column is truncated at 7, so it understates boss rows (the category columns are exact and uncapped).

The class a monster gets is MonStats data (D, `tre_act.py monsters`, 115 Act 1 / 2 monster ids from the Levels.txt spawn lists); H2H / Cast / Miss / Wraith / Champ / Unique each come as A, B and C (each carries a non-zero group in the table; I: in Nightmare and Hell A then moves on to B and C by monster level, which changes the item bases' levels, not the counts above; the move itself is the verified `FUN_00654e00` rule).

## 6. Chests (Q9)

Act 1 and 2 chests roll `Act N Chest A / B / C` (S re/drops.md: tier by the area level against the act's two marker levels, rolled off the chest's own seed with ilvl = the tier, items made at the area level, at most 6; a sparkling chest forces quality 4 or 6): `pk4`, NoDrop 100 of 142 per pick, so 75.4 % of a round has something, 1.18 items, 0.42 gold piles (the numbers of the table); tier C adds gear (1.28 items, 0.44 gear). How many rounds an opened container gives (chest 75 %, locked 2, urn 21 %, barrel, casket, corpse, bookshelf scroll or tome) is in re/drops.md (S). **V run:** 20,000 of 20,000 container jobs equal game.exe's real `OperateFn` (drops, extras, both seeds); 20,000 of 20,000 weapon racks and armor stands. The area-to-tier map for each Act 1 / 2 area was not extracted (U, gap for a later card).

## 7. Bot and d2d inputs (summary)

- ilvl = monster level (stat 12), not clvl or area level in Normal (Q10, S).
- d2d reads the same MonStats columns: `tc_fixed`, `tc_quest`, superunique TC (V run, section 1).

## 8. Differential runs (V run, worker host, nothing live)

d2d side: the upstream `drop-dump` built from the pristine upstream sources at merge 03f42254 (upstream main two merges before 72132b99; `components/rules/drops.hpp`, `components/game/loot.hpp` and `tools/drop-dump` have no diff between the two). Game side: `FUN_0055a6d0` and friends natively under the emulator of the 72132b99 read checkout. Runner: `bash tools/emu/tre_run.sh tre_diff.py ...` (the upstream `drops.py` with the d2d side pointed at the pristine build; `tre_all.sh` runs the five below).

| run | cases | result |
|---|---|---|
| `tre_diff.py 100001-200000` (rolls: seed, class, level, ilvl, players 1..8, mf -100..600) | 100,000 | 100,000 equal (not compared: ROP (H), ROP (N), see D1) |
| `tre_diff.py tables` (every class's entries) | 1,012 | 1,012 equal (same two excluded) |
| `tre_diff.py items 1-40000` (seeds, coins, stack, durability, defence, unique and set pick, mods, magic affixes) | 39,980 | 39,975 equal; 5 differ, all `elx` quality 4 (D2) |
| `tre_diff.py objects 1-20000` (chests, caskets, urns, barrels, corpses, bookshelves) | 20,000 | 20,000 equal |
| `tre_diff.py stands 1-20000` | 20,000 | 20,000 equal |
| `tre_select.py` (section 1) | 78,309 + 85,200 | 0 differ (d2d side a transcription) |
| `tre_act.py mc 8000` (exact odds vs native rolls: 13 classes) | 104,000 | largest |z| 1.68 |
| `tre_act.py players 20000` | 160,000 | in step with the formula |

## Deviations d2d vs game.exe (D1..)

None of these is proven by a new run in Act 1 or 2; D1-D6 are the upstream port's own "not ours yet" list (S re/drops.md and the `ponytail` comments of `drops.hpp` / `loot.hpp`), D7 and D8 are earlier traces.

| id | game.exe | d2d | tag | differential |
|---|---|---|---|---|
| D1 | TreasureClassEx entries with flag 1 / 2 (a unique or set item entry: ROP (N), ROP (H), Annihilus) are rolled | not rolled; the harness lists the two classes as known | S; my runs excluded them as upstream does | exists: remove the two names from `KNOWN` in `drops.py`; the lines differ at once. Not in Act 1 / 2 |
| D2 | an Elixir's own magic path (`FUN_00556a60`) | `generate_item` has none | V run: 5 of 39,980 items (all `elx` q4; the item's own seed after the roll differs); S upstream "9994 / 9995" | exists (`items`, run above). Whether any Act 1 / 2 class drops `elx` is U (the per-code table was extracted for consumables only) |
| D3 | players for NoDrop capped by the dropper's stat 100 when `FUN_0044be50` is 1 | not applied | S (`drops.hpp` ponytail); what stat 100 is: U | needs a hook of `FUN_0044be50`; could be added to `tre_select.py`-style harness |
| D4 | magic find (stat 0x50) and gold find (0x4f) of the killer plus its owner (`FUN_005585d0`, `FUN_005589a0`, `FUN_0058f0d0`), any source | gear only, plus the merc's and the wisps' dance | S (`loot.hpp`); no 1.14d skill or state gives either for the planned builds (I) | partly: the mf argument is in the 100,000 rolls (gear-independent); the sources are not |
| D5 | landing spot: rooms' full flags (`FUN_00555da0`, `FUN_0064e810`) | walk grid bit 0x01 plus items lying there; units do not block | S re/drops.md | exists upstream (`diff_drlg.py ... drops`); placement only, not counts or quality |
| D6 | an item `FUN_00555da0` finds no floor for is not made and not counted | every item counts | S (`loot.hpp` comment) | none cheap; sealed-position edge |
| D7 | quest class (TreasureClass4): the killer's owner's quest flags, none without a killer | the acting player's flags | I (read at 0x5a672c..0x5a67e9, not run) | build: `tre_select.py` with a pet killer (hooks for `FUN_00552f60`, `FUN_0058f0d0`); matters for Andariel and Duriel kills by a merc or pet |
| D8 | monster level (stat 12) = row level + 3 unique, + 2 champion, berserker champions +3; stats read at the base level | d2d's berserker champions +0, and boss stats read at the boosted level | S (docs/research/monster-mods-trace.md: 1,746 of 1,746 berserker cases, D1 and B-lvl there) | exists there; for drops it moves ilvl and the level move (NM / Hell) by up to 3 |

## Bot findings (B1..), described, not changed

| id | where | what the bot assumes | the rule | tag |
|---|---|---|---|---|
| B1 | `internal/resource/sim.go:123` and `:140` (`GoldPerKill: 6`, used at `:412`) | 6 gold per kill in the Blood Moor scenarios | Normal Blood Moor (zombie1, fallen1, quillrat1: level 1): 0.131 piles x 3.0 = **0.39 gold per kill**; Act 1 Cast / Miss classes 0.1125 piles; so the sim over-pays about 15x and its potion money is too rich (V run exact odds + D); docs/d2d-economy.md saw 0.3 to 0.4 per kill | V run, D |
| B2 | `internal/bot/recoveryxp.go:10` (comment: "Gold per kill is area independent") and the `recoveryXPPick` description at `internal/resource/tuning.go:127` | the XP is the only difference between the recovery areas | gold per pile is `3.5 ilvl - 0.5`: ilvl 1 (Blood Moor) 0.39 per kill, about ilvl 5 (Dark Wood: 18 coins per pile in the d2d probe) 2.2, **5.7x**; d2d-fidelity's own 3c row (2.9 vs 18 coins per pile) says the same | V run, D |
| B3 | `internal/env/tuning_loot.go:28` (`GoldMinStack 10`, `loot.goldMinStack`) | piles under 10 gold are not worth the walk (once the purse is over `poorGoldBelow` 300) | at ilvl 1 no pile reaches 10, at ilvl 2 20 %, ilvl 3 53 %, ilvl 4 70 %: in Blood Moor and Cold Plains the knob skips essentially every pile | D |
| B4 | `internal/data/loot.go:10` (field comment "Picks: negative means all items, each rolled independently") | negative Picks = every entry once | a negative Picks takes entry k for pick k, so an entry with weight 3 gives 3 drops and Picks -2 can give 3 items (`Act 1 Champ A`: 3.0 items from -2; `Countess`: 5.75 from -2) | V run (the 100,000 rolls and `tre_act.py mc`) |
| B5 | `internal/loot/pickit_starter.json` (rules `town scrolls`, `healing and mana`, `keys`, `spare *`) and the resupply plan | nothing numeric about supply | supply from kills is small and bursty: 0.057 potions per plain kill (about 1 per 17 kills), 2.0 per champion, 4 per unique or superunique, 0.015 scrolls per kill (one scroll per ~65 kills; Cast classes 1 per 29), keys 1 per 130; a Wraith kill is the best scroll / mana source (0.12 each). Useful as the expected supply in any soak judge; today none is stated | D |

## PR candidates

**None.** No d2d deviation was proven in Act 1 or 2 by an existing differential (mismatches before > 0): the only mismatches in the five runs are D1 (excluded by the harness itself) and D2 (elixirs, not a drop in the classes read here, U). Nothing built in the fork checkout.

## Dead ends

- The Basin gold multiplier for champions (1280 / 1536 / 2048 in Normal / Nightmare / Hell, docs/sme/mechanics-items-drops.md section 4, S) is not what the 1.14d class records say: `gld,mul=1280` in all three difficulties for every Act 1 / 2 champion class (D, read from game.exe's memory). The SME number is wrong for 1.14d.
- The first selection run printed about 25,000 "game None" cases: MonStats rows with no class on a column (e.g. towns, NPCs) never reach the roller; the harness now counts them as equal when d2d's class is empty (24,948 cases).
- Chest tier per area (the two marker levels) and the per-area monster mix (spawn weights) were not extracted: the tables above are per class, not per area.
- The SME's drop-rate tables (docs/sme/mechanics-items-drops.md sections 2.3, 4) are 1.09 / 1.13-era Basin pages; not reconciled with the V run numbers, which win.
- `elx` was not traced further: its own magic path is not ported (D2) but no Act 1 / 2 class was shown to hold it.
- A quality table per kill source (champion vs plain) needs the ilvl mix of the area; the table in section 3 is per class and ilvl.

## Drill recipe for the fork verbs (docs/forkdbg.md, docs/forkcheats.md; the lab binary; run later, not here)

A spawned monster is the plain `make_monster` unit (no pack); the player stands on the same level.

- **T1 plain rates.** `debug spawn fallen1 1` at Blood Moor, kill it (`debug mon set <id> life 1` then a cast, or `debug hit`), read `ground` and `dump items`; 2,000 kills. Expect piles 0.131 (+/- 0.016), pile coins 1..5 mean 3.0, potions 0.057 per kill (hp1 0.026, mp1 0.011, hp2 0.009, mp2 0.009), scrolls 0.015, drops per kill 0.375. U until run (V run of the same numbers on game.exe's tables exists, not on a live d2d).
- **T2 champion and unique.** `debug spawn fallen1 1 champion` and `... unique`: expect exactly 3 items (2 potions among hp1 hp2 mp1 mp2 rvs) and exactly 5 (4 potions); if a champion ever drops 0 or 1 the class selection is wrong.
- **T3 players.** With a second party member in the same area (or `/players 3`): the no-drop share of T1 falls from 62.5 % to 38.8 %; `/players 2` must change nothing.
- **T4 gold by area.** The same kill of fallen1 in Blood Moor (ilvl 1) and a `fallen2` / `zombie` in Dark Wood: coins per pile mean 3.0 vs about 17 at ilvl 5.
- **T5 first Andariel kill.** (Not before Act 1 is reachable on the lab character.) First kill rolls `Andarielq` (no potions, 3.2 gear), the second kill `Andariel`.

## Matrix and coverage cell text (proposed)

- `docs/coverage/ledger.tsv` (done in this card): `treasure #850 TESTED-0-DIFF ... 100000`; the generated `rule-coverage.md` row for `treasure` goes from 0 to 850 of 852; the two ROP rows stay untested (D1).
- Gap row 5 of `docs/coverage/gaps.tsv` is replaced by "ROP (N), ROP (H): unique item entries, Annihilus; not Act 1 / 2" if the coordinator wants it kept.
- docs/classes/ranking.md (treasure supply for the Sorceress and Paladin plans): potions and scrolls come mainly from champions, uniques and superuniques (2 to 4 potions each) and Wraith / caster kills; a plain Normal kill is 0.057 potions; gear is 0.11 items per kill, 7 % superior or better in Normal Act 1 at MF 0.

## Shared-file lines for the merger

See `openspec/changes/treasure/merge-lines.txt` (BACKLOG, INDEX, CHANGELOG).

## Limits

Only Act 1 and 2 monster classes, Act 1 / 2 chests and the Act 1 / 2 / quest bosses were tabulated; magic find, gold find and players are 0 / 1 in the tables (players varied only in the differential); the class a unique's minions roll is inferred (I); the d2d side of the selection is a transcription; not run on a live d2d; no live game.exe measurement; D1-D8 are listed from sources, only D2 has a count of my own.
