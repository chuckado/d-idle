## What changes

`FUN_00622720` gives a monster that is hit a chance to block (stat 20, MonStats `ToBlock` of the difficulty, at most 75) only through one of two gates. Gate A: the MonStats record's bit 26 (`MonStats.txt` `NoShldBlock`; Mephisto, Diablo and his clone, the Baal crab and clone, the Doom Knight, Andariel, Duriel, the Unravelers). Gate B (`FUN_006225f0`): the unit's look holds a shield on the shield layer. `FUN_00664860(unit, 7, idx)` takes monster data `+0xb` (the look's pick of the MonStats2 `SHv` list), returns its component code, and `FUN_00633640`'s item record for that code must have type word 2, ItemTypes' `shie`. The torch the Fallen can carry, `nil`, `lit`, `med` and `hvy` are not shields. MonStats2 `mBL`, which only says the BLOCK animation exists, is not read.

d2d gave every `mBL` row its `ToBlock` whatever it carried: a Fallen without a shield blocked 9 percent of the landed blows, a skeleton with no shield and a Corrupt Rogue the same, the Skeleton Mage rows (no shield list at all) 9 to 17 percent; and Mephisto, the Baal crab and clone (`mBL` 0) never blocked.

`Monster::block_chance` applies the two gates on the monster's own look (`npc.comp[7]`, which `spawn_monsters` sets from `monster_look`'s layer-7 pick); `MonType::no_shield_block` is the new column; `GameData::shield_code` is the item-code test (`type == shie` over `item_pieces`). `Monster::target` calls it. `docs/research/re/combat.md` records the rule.

Two commits: the oracle (`tools/emu/monster_block.py`, `tests/game/test_monster_block.cpp --monster-block`, a README row), then the claim (the rule and the pins).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00622720` `0x622741..0x62281b` | monster branch: gate A (`record+0xf & [0x6ce270]`, the mask is 4), else `FUN_006225f0`; open: stat 20 (`FUN_00625480(unit, 0x14, 0)`), `cmp 0x4b` caps at 75 |
| `FUN_006225f0` | idx = monster data `+0xb`; `FUN_00664860(unit, 7, idx)`; refused when the code is in the list at `0x744580` (one entry, the torch); else `FUN_00633640(code)` and the item record's word `+0x11e` must be 2 |
| `FUN_00664860` | MonStats2 count byte `+0x15+layer` against idx, variant byte `+0x26 + layer*12 + idx`, `FUN_006117d0` for the code; in an Act 2 level a BaseId 0 row takes the table at `0x7489a8` instead (layer 7 lists the same four shields) |

## Counted differential (V run)

`monster_block.py game` boots a game of each difficulty, spawns every MonStats row with the real spawner (`FUN_00555230`), writes each shield-list index into monster data `+0xb` and calls `FUN_00622720` natively. `test_monster_block --monster-block` prints d2d's real `make_monster` and `Monster::target(...).block` for the same cases (the row's shield entry set as `spawn_monsters` sets it). 2,469 cases (every row x difficulty x list entry; 93 of them are 31 placeholder rows the emulator cannot spawn, skipped and counted).

| | cases | differ |
|---|---|---|
| parent | 2,469 | 228 (195 d2d blocks where game.exe does not, 33 game.exe blocks where d2d does not) |
| claim | 2,469 | 0 |

`monster_block.py codes` runs the shield test over all 659 item codes of the tables: 24 shields in game.exe, 0 where the `type` column disagrees. `monster_block.py cap` sets Mephisto's stat 20 to 0, 1, 74, 75, 76, 90, 200: game.exe gives 0, 1, 74, 75, 75, 75, 75.

MUTANT_SUMMARY

## In play

At Normal, game.exe lets Mephisto block 20 percent of the landed blows (40 and 50 in Nightmare and Hell) and the Unravelers, Andariel and Duriel from Nightmare; d2d never did. A Fallen, a skeleton or a Corrupt Rogue without a shield blocked in d2d (9 to 49 percent by row) and never does in game.exe; the Skeleton Mages never block in game.exe. Which share of a game's Fallen carry a shield depends on the region's look sets (`monster_look`, not changed here).

## Limits

- The shield entry is the one `spawn_monsters` sets from `monster_look`; that d2d's pick equals game.exe's per spawn is the existing population port's claim (not re-run here). The oracle sets the entry directly on both sides.
- An Act 2 level's table for BaseId 0 rows was read and agrees with the MonStats2 list on layer 7; an Act 2 level could not be brought up in the emulator, so no unit was run there.
- 31 placeholder MonStats rows could not be spawned by the emulator; they have no shield list and no `NoShldBlock`.
- Gate A was found as MonStats column `NoShldBlock` by matching bit 26 of the compiled table against every column of `monstats.txt`: all 734 rows agree.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cd build && ninja test_monster_block && cd ../tools/emu && uv sync
set -o pipefail
uv run python monster_block.py gen > /tmp/c.tsv
uv run python monster_block.py game /tmp/c.tsv > /tmp/g.tsv
../../build/tests/test_monster_block --monster-block /tmp/c.tsv | grep -E '^[0-9]+ -?[0-9]+$' > /tmp/d.tsv
uv run python monster_block.py compare /tmp/c.tsv /tmp/g.tsv /tmp/d.tsv   # parent: differs, exit 1; claim: 0, exit 0
uv run python monster_block.py codes && uv run python monster_block.py cap
../../build/tests/test_monster_block
```

---

Agent: Claude Code
