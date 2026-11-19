# decomp40: the Cairn Stones order, the Tristram portal and the scroll drop, game.exe 1.14d vs d2d vs our bot

> status: current (2026-11-18). Docs only: no game started, nothing posted. Trigger: a d2d contributor reported that the translated scroll does not show the pillars' order, and that the Tristram portal appears OUTSIDE the pillars and BLUE, where the original shows a RED portal in the MIDDLE; our drill L1 (`docs/probes/search-for-cain-l1-2026-11-18.md`, on the cain branches not yet on main; fixtures `docs/probes/cainl1/testdata`) saw the portal at StoneLambda + (6,-3). Consumer: the CAIN-4/5 slices of the Search for Cain work. Tags: V read (read in game.exe's disassembly, this run, through the lab disassembler), S (shown by the fork's own notes or source, which cite game.exe), D (derived from d2d lab output), I (inferred), U (unverified). NO claim here is V run: no emulator differential was run (plan in section 6), so nothing here is PR-proof yet.

## 1. Questions and answers

**Q1a. How does game.exe decide the order?** A random permutation of the five stone classes 0x11..0x15, drawn from the game's quest rng, once per game. V read `FUN_00592e90` (called with record data +0x4a as the "generated" latch): seed = `lo * 0x6ac690c5 + hi` (64 bit), slot = `lo % 5`, the n-th class goes to the first free slot; the 64-bit state is the one `FUN_005438b0` returns, `game+0x10f4 + 0x18` (V read, 0x5438b0). Different per game, constant within it. The d2d port (`CainQuest::stone_order`, `components/rules/quests.hpp:1543`) is the same loop on `fight.spawning.quests.low/high` (S; same recurrence read by eye, equal rng state per game is U).

**Q1b. Is the order shown to the player?** Yes, once, on the DECIPHERED scroll. V read `FUN_00593cb0` (0x593cb0), reached when a player reads (uses) a `bkd`: it makes the order if +0x4a is still 0, then builds a 0x50-type message (subtype 4, five words `class - 0x11`, bound-checked, error 0x87f when a word is above 4) and sends it to that player with `FUN_0053d7e0` (S->C). The client shows them as a five-stone pictogram replay (section 9), not text. Nothing else shows it: Akara's decipher dialog (`FUN_00592250`, dialog 0x70) only swaps `bks` for `bkd` (V read through the fork's notes, S); the stones carry no text. Corollary: because the order is made at the FIRST read of the `bkd` (or the first stone touch), the order is fixed before the player ever clicks a stone.

**Q1c. A wrong click.** V read, `FUN_00593710` (OperateFn 9): the touched class is compared with `data[idx]`; if different the function returns. No reset, no sound, no mode change, `idx` and the lit stones stay. A stone already lit (mode != 0) returns before counting (step 9 of the fork's notes, S; the compare/return V read at 0x593710..0x59398x). So the lit set only grows. d2d: same (`CainQuest::stone`, `quests.hpp:1556..1565`; our L1 runs `b1`, `c2` D: after 15 clicks the lit set only grew). NO deviation on the rule; the only difference is the display (D3).

**Q2a. The Tristram portal object.** The fifth right stone makes `FUN_0056ede0(game, room, 0, 1, 0x120, x+6, y-3)` (V read: at 0x593a16..0x593a2f the code loads StoneLambda's x/y from `FUN_00620870`, adds 6 to x (`add eax,6`) and -3 to y (`add ecx,-3`), pushes 0x120 and calls 0x56ede0). That is missile 288 (`cairnstones`); its server function reaches `FUN_005a9930`, which takes the missile's own position and calls `FUN_0056d130(room, x, y, level, 0, 0x3c, 1)`: object 0x3c = 60, the permanent RED portal, level 38 (S: the fork's notes; the level value is inferred from the missile parameter, the extracted Missiles.txt lacks the func columns: I). The rejoin path (`FUN_00592d50`, stone InitFn 6) makes the same object 60 at StoneAlpha + (4,4) (S). Object 60 vs the blue town portal 59: V read of the call constant 0x3c (S for the two object names).

**Q2b. Where, in what unit.** game.exe unit positions are subtile integers (a tile is 5x5 subtiles); `FUN_00620870` returns the unit's integer subtile position (V read: it calls the path getter 0x6488c0/0x648900 and stores ints; that these are subtiles, not tiles, is I, matching every other offset in the fork, e.g. `portal_spot (subtile+0.5)/5`, Duriel's portal `orifice->x - 13.f/5`). So the portal stands 1.2 tiles right and 0.6 tiles up of StoneLambda. With our L1 numbers (StoneLambda 19.1,60.7; stones at x 18.1..21.9, y 59.1..63.1; the class 61 preset, the "Stone Sigil", at 20.5,60.5, D from `search-for-cain-l0`) the game.exe spot is (20.3, 60.1): INSIDE the ring, 0.4 tile from the sigil. The contributor's "middle of the pillars" is exactly this (I).

**Q2c. d2d.** `World::cain_operate`, `components/game/world/quests/shared.cpp:722` (fork HEAD, and the same line on `upstream/main`): `level->nearest_free(lambda->x + 6, lambda->y - 3)`. `Npc::x`/`y` and `nearest_free` are in TILE floats (`components/game/gamedata/gamedata.hpp:299`, ring step 0.2 = one subtile). So d2d adds 6 and -3 TILES = 30 and 15 subtiles. Result (25.1, 57.7), D by L1 (`debug portals` row `2 4 25.1 57.7`; no `nearest_free` move: free cell): 6.0 tiles right and 3.0 up of StoneLambda, outside the ring. Exactly 5x the game.exe offset on both axes (D1). The same mistake is on two more lines: the rejoin portal `setup.cpp:477` (`alpha->x + 4, alpha->y + 4`, whose own comment says "StoneAlpha + 4 subtiles") and the Gibbet fallback portal `shared.cpp:705` (`open_portal_at(object.x + 6, object.y + 6, ...)`, game.exe `FUN_0056d130` at x+6, y+6 subtiles, S).

**Q2d. Red vs blue.** In the fork source the stones' portal slot 2 is drawn with object 60 (`World::portal_object`, `world.hpp:471`: slots 2..5 and 8..11 give 60, so `portal_look` in `apps/d2d/town/view.cpp:99` picks `scene.red_portal`), since commit 20d49580 (2026-11-07), which `upstream/main` (2026-11-17) contains (S: `git merge-base --is-ancestor`). So current d2d is RED and the blue the contributor saw is most likely an older build (I), unless object 60's art failed to load on his install (U; the loader falls back to nothing, not to blue, `gamedata_load.cpp` `portal(60, red_portal)`). NOT a deviation of HEAD; ask which build (section 7).

**Q3. The tree and the scroll.** game.exe, `FUN_00593af0` (OperateFn 12, S + V read of the quest record offsets): requires the tree untouched (mode 0), state < 6, the player's bits 0 and 1 clear and the player holding NEITHER `bkd` NOR `bks`; then voice event 0x2d, forces `bks` as the dropped item (`obj+0xb8 = 'bks '`), drops it at the tree (`FUN_00559a30`), tree mode 1. One drop per game (mode stays 1; re-click does nothing); a player who already has either scroll gets the "impossible" voice (0x13). The scroll is a quest item: Items.txt `quest` +0x12a is non-zero, `FUN_00555d20` binds it to quest record a1q4 as it is made, and `FUN_0055ca40` refuses a second of the same pair `bks`/`bkd` (S: the fork's `itemmode.md`). d2d: `CainQuest::tree` (`quests.hpp:1535`: `!tree_used`, no bits, `!has_scroll`) then `loot.put({.code="bks"}, ...)` (`shared.cpp:679`): same rule. Our drills: L0 did not click the tree (read-only listing); L1 gave a `bkd` by `debug give bkd` (no tree click); no probe doc clicked the tree. So the tree-to-scroll drop on d2d is U in the lab (rule equal by code); whether the d2d scroll can be dropped or traded depends on the `quest` column handling in the item code path (I: same table, not traced; D5).

## 2. What the contributor saw, mapped

| Report | Cause | Tag |
|---|---|---|
| the translated scroll shows no pattern | d2d has no reader for `bkd`: no `FUN_00593cb0` equivalent, no S->C 0x50/4 message (grep of the fork for `bkd` use: only item bookkeeping, the stone operate and Akara's decipher) | D2, V read for the game side, D (grep) for d2d |
| only one pillar lights at a time, try every one | a correct rule, d2d = game.exe (Q1c); the order is simply invisible to the player (D2) | V read |
| portal outside the pillars | D1, offset applied in tiles | V read + D |
| portal blue | older build (Q2d) | I |
| tree fixed, scroll dropped | nothing to do | S |

## 3. Deviations d2d vs game.exe

| ID | game.exe | d2d | Tag | Differential |
|---|---|---|---|---|
| D1 | portal at StoneLambda + (6,-3) SUBTILES = + (1.2,-0.6) tiles | + (6,-3) TILES (`shared.cpp:722`); same for the rejoin portal (+4,+4, `setup.cpp:477`) and the Gibbet fallback (+6,+6, `shared.cpp:705`) | V read (game side), D (d2d side, L1) | yes, buildable (section 6); the arithmetic is already proven by two lab runs |
| D2 | reading a `bkd` sends the order (`FUN_00593cb0`, message 0x50/4), the client replays the five stones lighting in order (section 9) | nothing sent; the order exists only after the first stone touch and only in the d2d log | V read, D | UI, not a numeric oracle: unit test that reading the item emits the message with `order[i]-17`; game side V read. Needs the client text (U: which UI string draws the five words) |
| D7 | the order is shown as a pictogram replay of the five stones (section 9) | nothing | V read, D | two oracles in section 9 |
| D3 | the order is made at the first read of the `bkd` OR the first stone touch (`+0x4a` latch) | only at the first stone touch | V read | consequence: the rng position at which later rng users see the game differs when the scroll is read first; U if any other code consumes `game+0x10f4` between the two times |
| D4 | stone InitFn/touch also drives the "Stone Sigil" (object 61, record +0x28): sigil mode = idx+1 per lit stone, 6 on completion (`FUN_00593710` steps 5, 10, 11; S) | no sigil handling (grep: no object 61 use); the class 61 preset is only listed | S (game), D (grep) | visual only; U what the sigil shows (invisible object with a light/anim?) |
| D5 | quest item (`bks`, `bkd`): the +0x12a flag decides what the item path refuses | d2d's handling of the flag not traced here | U | read `ItemCanDrop`/trade code for `quest` in the fork |
| D6 | portal red | red at fork HEAD and upstream/main | S | none; the reported blue = build age (I) |

## 4. Bot findings

- B1. `cmd/d2bot/cain.go:68-71` on main (`solveCairn`, backtracking, "a wrong stone resets the lit ones (U)"): wrong per Q1c (V read). The cain branches (`agent/cain5b`, `cain.go:81-84`) already replaced it with a cycle plus a portal watch. Not changed here.
- B2. cain branches `cmd/d2bot/cain.go:181` and `:268` find the portal by object class `quest.ObjRedPortal` (60) in the snapshot; in d2d the way to Tristram is not an object in `debug objects` or presets, only in `debug portals` (L1, D). Whether the live bot sees it through the snapshot at all is the CAIN-5 question; the class is right for game.exe (Q2a).
- B3. The bot assumes nothing about the portal offset (it takes the portal by class and `-2000-which`, `objects.go` row `"Tristram portal"`), so D1 does not break it today; it would break a hard-coded `Lambda + (6,-3)` (none found).
- B4. The bot cannot read the scroll's order (D2), so the cycle solver is the right design for both d2d and game.exe: at most 15 clicks.

## 5. What the bot must do under either d2d behaviour

1. Never compute the portal spot from the stones. Find it by the portal list (d2d) or the red portal object class 60 (game.exe) and walk to it by pathing; keep a generous interaction radius, because after the D1 fix it stands INSIDE the ring (about 1.2 tiles from StoneLambda), next to stone footprints (`nearest_free` may nudge it), whereas today it stands 6 tiles away in the open.
2. Keep the lit-state loop (`ON` per accepted click, one `debug objects` read) and "never re-click a lit stone": valid for both builds (Q1c).
3. Once d2d sends the order (D2 fixed), an optional fast path may read the five words from the item/message and click in that order; keep the cycle as the fallback because the bot has no UI-text reader today.
4. After a fix in d2d the portal position changes between builds: do not cache a spot across restarts or fixtures (`debug portals` row in `docs/probes/cainl1/testdata/c2.txt` is build specific).
5. Class-61 sigil lights (D4) are not a signal; use stones' `ON` and the portal.

## 6. PR candidate: yes for D1 (plan only, NOT proven), no for the rest

- D1 is a one-line family (three lines): divide the subtile offsets by 5 (`lambda->x + 6.f / 5`, `lambda->y - 3.f / 5`; `alpha->x + 4.f / 5`; `object.x + 6.f / 5`), the way `shared.cpp:296` does for Duriel. The fork's own test `tests/game/test_cairn_missile.cpp` only checks that the flash stands at `portal[2]`, so it passes either way.
- Proof plan under the existing standard (differential, then a before/after count):
  1. V run, game side: run `FUN_00593710`'s tail (0x593984..0x593a2f) under the emulator (`tools/emu`, like `quest_cain_rewind.py`, which hooks the callees) with a hooked `FUN_00620870` returning a chosen StoneLambda position and a hooked `FUN_0056ede0` recording its stack arguments; the oracle is the recorded (x, y) = (lambda_x + 6, lambda_y - 3) in game units; then drive the NEXT unit conversion: the object's position in the room (`FUN_0056d130` / the unit placer) to confirm the unit is the subtile (check: the same placer's output for a known preset object whose DS1 tile position is known).
  2. d2d side: a unit test (no MPQ needed for the arithmetic, MPQ for the real Stony Field) that makes the fifth stone touch and compares `portal[2]` with `lambda + (1.2,-0.6)` after `nearest_free`: mismatches before the fix > 0 (1 of 1 at L1: the row `2 4 25.1 57.7` against `20.3 60.1`), after the fix 0.
  3. Independent check: Opus verifier (`d2-reviewer`), own harness.
  Draft text stays in the repo; nothing is posted; no branch is made here (rule: only with a proven differential).
- D2 would be a feature PR (new S->C message and client text), not a bug; needs the client's string table (U).

## 7. Drill recipe (docs/forkdbg.md, docs/forkcheats.md; do not run without a slot)

1. Build check for the contributor's report (no game.exe needed): lab slot, lab binary `d2d.lab3-labverbs` (or the next one), throwaway Sorceress, `debug give bkd pack`, `debug warp 4`, click the five stones by cycle (L1 `run.sh`), then `debug portals`, `debug objects`, and `debug presets` (class 61 row). Expected today (D1): portal at StoneLambda + (6.0,-3.0); expected after the fix: + (1.2,-0.6), within 0.5 of the class 61 preset. Confirms D1 on a second seed and the "is red" claim: the portal slot 2 object is 60 in `portal_object` (a screenshot of the portal is the only way to confirm the colour; U in the lab).
2. D3: read `bkd` before any stone touch, then `debug` dump of the quest rng if the fork has a verb (U: none known); otherwise compare the order printed after reading against the log line at the first touch.
3. D5: with `bks` in the pack try `drop` and a trade/stash move in the lab; expected refusal in game.exe (S).
4. U on the game.exe side stays U until the emulator differential in section 6 runs.

## 8. Matrix and queue notes

No quest matrix doc exists; proposed first rows for one: `A1Q4 stones order | game.exe = random per game, shown on bkd read | d2d: rule equal, display missing (D2)`; `A1Q4 Tristram portal | red object 60 at Lambda+(6,-3) subtiles | d2d: red, 5x offset (D1)`.

Missing glossary terms (report only): "subtile" vs "tile" (5 subtiles = 1 tile) for coordinates in fork prose.

## 9. Translated scroll pattern (added on the coordinator's request, 2026-11-18)

Verdict: d2d does NOT show the pattern. V read for game.exe, D (source grep, no run) for d2d. It is a PICTOGRAM, not text.

Game.exe chain, all V read this run unless tagged:
1. Akara (dialog 0x70, `FUN_00592250`, S): removes `bks`, makes `bkd` ("Key to the Cairn Stones") with `FUN_005466b0`; the item carries no per-item sequence. The permutation lives only in the quest record (`data[0..4]`, +0x4a latch).
2. Selecting the `bkd` (inventory right-click) runs `FUN_0049ff90` (S, notes): sets `0x7bf254 = 1` and sends client packet 0x3e (`FUN_00478680`, S) asking the server for the order. While `0x7bf254 != 0` the item text is the ordinary item text (`FUN_0049fa10`).
3. The server `FUN_00593cb0` (0x593cb0) makes the order if `+0x4a == 0` (calls `FUN_00592e90`), builds a 15-byte message: byte 0x50, word 4, then five words `data[i] - 0x11` (0..4), and sends it with `FUN_0053d7e0` (0x593ce6..0x593d34). A word above 4 aborts (error 0x87f).
4. The client handler at 0x49ff60 checks word `[+1] == 4`, copies the five words to `0x7bf098..0x7bf0a0` and clears `0x7bf254`. From then on `FUN_0049ff10` (item text dispatch: code `bkd ` with `0x7bf254 == 0`) calls `FUN_0049fba0(1)`.
5. `FUN_0049fba0` (0x49fba0) draws, per step i 0..4 of the sequence, the picture of stone `0x7bf098[i]` at the screen position from the table at 0x722eb8 (x,y per stone) and animates by a 50 ms counter (`0x7bf243`, `0x7bf247`) against per-step delays at 0x722f08: a step is dim, then bright, then (counter >= 0x15) in its final colour, so the five stones light one after another in the order of the sequence. It loads two art files named by format strings at 0x6da174 and 0x6da184 (U: which files; not read), draws with `FUN_004f6480`. So the player SEES the five stones lighting in order, a replay, not a written list.

d2d: no `bkd` use handler (no `bks`/`bkd` in `apps/d2d` or `components/net`; only item bookkeeping and the stone operate in `components/game`), no message 0x50/4 built, no client reader for it, no pictogram. `apps/d2d/netgame.cpp:426` handles only `0x89`. Right-clicking the key shows the plain item text, so the translated scroll shows nothing about the order (exactly the report). Our `debug` log prints the order only on the first stone touch (D2).

PR candidate (feature, plan only; nothing run, no branch):
- Server: on the read request (C->S 0x3e from the selected `bkd`), call `cain.stone_order` if `!ordered` and send the 15-byte message (0x50, 4, order[i]-17 as u16). Client: store the five words, draw the pictogram as `FUN_0049fba0(1)` does (needs the two art files and the tables 0x722eb8 / 0x722f08, U: read them from the exe or the MPQs).
- Differential (existing standard: oracle runs game.exe code): emulator run of `FUN_00593cb0` with hooked `FUN_00543640` (record with +0x18 data: `+0x4a` = 1, five order bytes for each of the 120 permutations of 0x11..0x15) and a hook on `FUN_0053d7e0` that captures the 15 bytes; d2d side builds the same message from `CainQuest::order`; compare byte for byte (expected: 120 of 120 equal after the fix, 0 of 120 before because d2d sends nothing). Second oracle for the client: run `FUN_0049ff60` then `FUN_0049fba0(1)` with the draw call `FUN_004f6480` hooked to log (art, x, y, colour index) per counter value; compare with d2d's pictogram draw list for a sample of orders and counter values. Also fix D3 together (order made at the read).
- Needs the coordinator: no lab run was made; a lab slot is needed only to confirm that d2d sends nothing today (expected, by source), and an Opus verifier (`d2-reviewer`) before any PR.

Matrix row proposal: `A1Q4 scroll pattern | game.exe: 0x50/4 message, five-stone replay pictogram | d2d: missing (D2, D7)`.

## Shared-file lines for the merger
Generated from the task card at merge (title and result); no merge-lines file.
