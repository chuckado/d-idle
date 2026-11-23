## What changes

`FUN_00622720` gives a monster that is hit a chance to block (stat 20, MonStats `ToBlock` of the difficulty, at most 75) only through one of two gates. Gate A: the MonStats record's bit 26 (`MonStats.txt` `NoShldBlock`; Mephisto, Diablo and his clone, the Baal crab and clone, the Doom Knight, Andariel, Duriel, the Unravelers). Gate B (`FUN_006225f0`): the unit's look holds a shield on the shield layer. `FUN_00664860(unit, 7, idx)` takes monster data `+0xb` (the look's pick of the MonStats2 `SHv` list), returns its component code, and `FUN_00633640`'s item record for that code must have type word 2, ItemTypes' `shie`. The torch the Fallen can carry, `nil`, `lit`, `med` and `hvy` are not shields. MonStats2 `mBL`, which only says the BLOCK animation exists, is not read.

d2d gave every `mBL` row its `ToBlock` whatever it carried: a Fallen without a shield blocked 9 percent of the landed blows, a skeleton with no shield and a Corrupt Rogue the same, the Skeleton Mage rows (no shield list at all) 9 to 17 percent; and Mephisto, the Baal crab and clone (`mBL` 0) never blocked.

`Monster::block_chance` applies the two gates on the monster's own look (`npc.comp[7]`, which `spawn_monsters` sets from `monster_look`'s layer-7 pick); `MonType::no_shield_block` is the new column; `GameData::shield_code` is the item-code test (`type == shie` over `item_pieces`). `Monster::target` calls it. `docs/research/re/combat.md` records the rule.

Three commits: the oracle (`tools/emu/monster_block.py` and `tests/game/test_monster_block.cpp --monster-block`; it builds and runs alone on the parent), the claim (the rule and the pins), then the finish (the README row, the `MB_SPOT` room control, `D2D_TEST_SKIPPED` in every mode when the data is missing, `compare` reading plain integers only).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00622720` `0x622741..0x62281b` | monster branch: gate A (`record+0xf & [0x6ce270]`, the mask is 4), else `FUN_006225f0`; open: stat 20 (`FUN_00625480(unit, 0x14, 0)`), `cmp 0x4b` caps at 75 |
| `FUN_006225f0` | a four-class special case first (jump table at `0x6226a4`: classes 243, 310 and 333 return 1, class 359 returns 0; none changes an outcome, the first three are `NoShldBlock` rows and the fourth has no shield); then idx = monster data `+0xb`; `FUN_00664860(unit, 7, idx)`; refused when the code is in the list at `0x744580` (one entry, the torch); else `FUN_00633640(code)` and the item record's word `+0x11e` must be 2 |
| `FUN_00664860` | MonStats2 count byte `+0x15+layer` against idx, variant byte `+0x26 + layer*12 + idx`, `FUN_006117d0` for the code; in an Act 2 level a BaseId 0 row takes the table at `0x7489a8` instead (layer 7 lists the same four shields) |

## Counted differential (V run)

`monster_block.py game` boots a game of each difficulty, spawns every MonStats row with the real spawner (`FUN_00555230`), writes each shield-list index into monster data `+0xb` and calls `FUN_00622720` natively. `test_monster_block --monster-block` prints d2d's real `make_monster` and `Monster::target(...).block` for the same cases (the row's shield entry set as `spawn_monsters` sets it). 2,469 cases (every row x difficulty x list entry; 93 of them are 31 placeholder rows the emulator cannot spawn, skipped and counted).

| | cases | differ |
|---|---|---|
| parent | 2,469 | 228 (195 d2d blocks where game.exe does not, 33 game.exe blocks where d2d does not) |
| claim | 2,469 | 0 |

The room control (`MB_SPOT=5`, each unit spawned in another room) also gives 0 of 2,469.

`monster_block.py codes` runs the shield test over all 659 item codes of the tables: 24 shields in game.exe, 0 where the `type` column disagrees. `monster_block.py cap` sets Mephisto's stat 20 to 0, 1, 74, 75, 76, 90, 200: game.exe gives 0, 1, 74, 75, 75, 75, 75.

Mutants of the claim (one text change each on a copy of the tree, `test_monster_block` rebuilt, pins run, compare of the 2,469 cases): 17 mutants that change behaviour all abort the pins: the parent rule, gate A or gate B dropped, no gate, the cap removed or 74, the difficulty ignored, layer 6 for 7, `mBL` ANDed in or required for gate A, the shield read from the `component` column, any item code as a shield, the torch as a shield, gate A read from `boss` or `primeevil`, `shld` for `shie`, gate B by code name. 15 of them also differ on the compare (228, 45, 255, 1,560, 30, 33 and others of 2,469); the two cap mutants are caught by the cap pins only (no row has a `ToBlock` above 75). Two mutants equal the claim on the shipped tables and are not counted: `NoShldBlock` read as any non-empty value (the column holds `1` or nothing) and gate B also accepting `type2 == shie` (no layer-7 code has it). The control, the claim unchanged through the same runner, passes. m10, m11, m12 and m13 were rerun as edits of `ai.hpp` (the first runner could not compile them). They were run on the pre-rebase tree; the rule lines did not change in the rebase (only `gear_res` joined the resistance loop).

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
