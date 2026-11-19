# Search for Cain, drill L2: free Cain, return, report to Akara (2026-11-18)

> status: result of drill CAIN-L2, follow-up to `search-for-cain-l1-2026-11-18.md` (stones, portal, arrival). Same setup: d2d lab slot, lab binary `d2d.lab3-labverbs`, throwaway Sorceress "Drilly" (CHEATED: warped, `bkd` given, monsters killed in most runs: mechanics only), `d2drill -probe`, every `debug` verb ledgered; NOT Game.exe, no Wine instance. Tags: V verified live (fixture), S secondary (fork source), U unverified. Fixtures `docs/probes/cainl2/testdata/*.txt` (probe transcripts, pure grid rows stripped; `grid-crop.txt` is the walk area); command `docs/probes/cainl2/run.sh LABEL 'verb;...'`. Every run replays the L1 prefix (seed 11, the three stone cycles) and then the steps named below, so each fixture starts with the same ~70 stone lines.

## (1) Where the cage is, and the way to it (V)

- Tristram is level 38. Arrival from the Stony Field portal: (21.5, 24.6) (`a1`). Cain's Gibbet is object class 26, `op10`, at (27.3, 27.7), object index 31 in `debug objects` / `debug presets` / `npcs` (the same index space as the stones: pass it to `cmd interact`), state `NU`. Distance from arrival: 5.8 cells east, 3.1 south. The index (31) was the same in all runs of this seed; as for the stones, resolve it by class 26 from `debug objects` (or `npcs` row "Cain's Gibbet") right before use (U for other seeds: not run).
- Deckard Cain is a hidden preset (row 0, class 0, `hidden=1`, "Deckard Cain", 27.9, 28.3) from the start; he is NOT in `npcs` until the cage is opened (V `a1` vs `h1`).
- The cage tile is an obstacle (`grid-crop.txt`: a blocked patch at subtile cols 134..138, rows 136..143, on the cage), the cells west of it are free.
- Walkability (V): `cmd interact 31` from the arrival does NOT get there: the player moves 1.2 cells (to 22.7, 24.7) and stands, repeating the interact changes nothing (`d1`, `f1`: 600 ticks). `cmd move 25 27` then `cmd interact 31` works: the player arrives at (25.1, 27.1) in 100 ticks, the interact then opens the cage (`f1`, `g1`, `h1`). So: walk to a free cell about 2 cells west of the cage, then interact (the interact itself walks only the last cells). A warp to (26, 27) and interact also works (`e1`).
- Dangers (V, `a1`, `b1`, `v1`): three Champion Skeleton Archers (lvl 7, 39 to 42 hp each) stand at (20.9..21.7, 26.7..27.5), 2 to 3 cells from the arrival, and are awake within 5 ticks (state `A1`). A level 4 Sorceress (43 hp) that walks to the cage without killing them died in under 200 ticks (`v1`: life 0/43). Further out: Griswold (superunique) at (31.7, 39.5), Bile the Mad ("Bile Kill the Mad", unique) with four minions at (27.7, 39.1), Night Clan goatman (24.3, 32.5), a skeleton archer (25.5, 35.7), a carver pack at (9..10, 38..39), Wirt's body (class 42) at (9.7, 35.9), "a Trap" objects (class 250: 18.1,28.1 and others, `op30`). The cage and the archers are the only things near the route; the others are 5 to 12 cells away (a full list is in `b1`/`c1`). I did not run a fight: the lab `debug kill` cleared everything (c1: 30 rows `DT`). A real hero must kill or outpace the three champions first (U how, not tried). The level has no town: the way out is a portal.

## (2) Freeing Cain (V)

- Action: stand within about 2 cells and `cmd interact 31` (the gibbet's object index). The reply is empty; after 100 ticks:
  - object 31: `op10 26` state `NU` -> `ON` (V `e1`, `f1`, `g1`). This is the "cage is open" signal for `debug objects`; for Game.exe the object's mode (U).
  - quest record 4 (`Search for Cain`, `debug quest list`): `0x0010` (bits=4) -> `0x2012` (bits 1, 4, 13) at the same moment, no delay (`g1`, `h1`). Bit 4 (0x10) was already set by arriving in Tristram (V `a1`: `4 0x0010` right after the portal). Bits 1 and 13 are new. In this build bit 1 (0x0002) is the "reward pending" bit of the other records (docs/forkcheats.md), so it means: cage opened, the quest is waiting to be turned in. Bit 13 (0x2000): U meaning (S: not named in the fork notes).
  - `npcs` shows "Deckard Cain" (row `x y 1 0 27.9 28.3`) about 75 ticks after the interact (not at 25, `h1`) and he is gone again by tick 175 without any talk (`h1`, `g1`): he walks off to town on his own. The player need not talk to him in the cage (there is no dialog to open). No XP, no item, no gold (V: `exp` equal before and after, `h1`, `i1`).
- So the bot's "free Cain" success signal can be the quest record 4 value `0x2012` (or bit 1 set), or the gibbet `ON`; NOT the NPC row (short lived).

## (3) The way back (V)

- Town portal: the character starts with a Scroll of Town Portal (item id 6, `items`); `key i`, `debug itemat 6` (433 329), `move 430 329`, `move 433 329`, `rclick 433 329`, `key i` (exactly the existing `CastTP`) opens `debug portals` rows `0 38 25.1 27.1` (in Tristram) and `1 1 16.7 25.7` (the town twin) (V `k1`). `cmd interact -2000` (`TakePortal(0)`) arrives in the Rogue Encampment, level 1 at (16.7, 26.3) after 200 ticks (V `l1`). So the existing TP logic works from level 38; no special case.
- The permanent portal back to Stony Field also works: `debug portals` row `3 38 21.5 24.0` in Tristram, `cmd interact -2003` goes to level 4 at (25.1, 58.3) (V `w1`). From there the bot's normal return applies (tp scroll or walk).
- Quest flags do not change when leaving (V `l1`, `w1`: record 4 stays `0x2012`).

## (4) The quest state and the report (V)

- After the cage: record 4 = `0x2012`. In town, Cain (npcs index 39 here; class 39) is standing at (13.3, 19.3). His menu: `Deckard Cain`, `talk`, `cancel`; after "talk": `talk`, `gossip`, `The Search for Cain`, `cancel`; "The Search for Cain" plays speech only (`speech=3`, `voice=3663`): record 4 stays `0x2012` through all of it (V `n1`, `o1`, `p1`). `cmd interact 39` stops short (about 1.5 cells away) at the first calls; repeat until `menu` lists the lines (three calls, `n1`).
- Talking to Akara (npc index 9, class 9 at 24.5, 15.5) completes the quest: record 4 `0x2012` -> `0x2011` (bits 0, 4, 13: bit 0 = done, the reward-pending bit 1 cleared) and the Den record 1 shows `0x0004` (bits=2, probably the "talked to Akara" bit) at the same time (V `r1`: the change came on the 3rd `cmd interact 9` after the walk (the first calls only walk, the greeting is spoken once the player is next to her); `t1`: the same without talking to Cain first, on the 4th to 5th call). So the Cain dialog is NOT needed: the report to Akara alone does it. Akara's menu then lists `talk`, `trade`, `cancel` (`q1`); no click was needed. Akara's greeting is spoken (`speech`) and the `menu` verb may be empty during it: poll `debug quest list` (or the save) instead of the menu.
- Reward: none seen (gold 0, exp unchanged). The identify service of Cain does not appear in his menu in these runs (U whether it needs more, S: in D2 it comes with the quest).

## What the bot needs (CAIN-5b inputs)

1. In Tristram (level 38): re-resolve the gibbet by class 26 (`debug objects`/`npcs` "Cain's Gibbet"), kill the three Champion Skeleton Archers at the arrival first (or do not enter below the safe level; U level), `cmd move` to a free cell about (25, 27), then `Interact(gibbet)`.
2. Success = quest record 4 bit 1 set (`0x2012`) and/or gibbet state `ON`. Do not wait for the Cain NPC row.
3. Return: `CastTP` then `TakePortal(0)` (or the Tristram portal `-2003` back to Stony Field and the bot's normal way); verify level 1.
4. In town: walk to Akara and talk (`cmd interact 9`, repeat until the greeting; poll the quest record); success = record 4 `0x2011` (bit 0 set, bit 1 clear). Cain's own dialog is optional and has no quest effect.
5. Board evidence (the quest board's done chunks derive from game state, `data/questboard/act1.json`): `free_cain` done = record 4 bit 1 (reward pending) or bit 0; `report` done = record 4 bit 0 with bit 1 clear. On Game.exe the same record is the quest word of "Search for Cain" (docs/d2-1.14d-memory.md); U that the bit numbers are the same as in d2d (d2d writes the game's own words, forkcheats.md).
6. Not covered: the same on game.exe (U), a second player, a seed other than 11 (U: object index 31 is only a measured value), how a real hero beats the champion archers.

## Dead ends / deviations
`b1`: `cmd interact 1` (the `npcs` list position of the gibbet) does nothing; the index is the object index (31). `d1`, `f1`: a bare `cmd interact 31` from the arrival stalls at (22.7, 24.7). `debug grid` prints the whole level (215x239 subtiles), the crop is in the fixtures. The d2d's own log was not collected this time (only the probe transcripts).
