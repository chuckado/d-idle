# Belt and ground-potion pickup traced in game.exe 1.14d: d2d and our bot compared

> status: research, 2026-10-31, agent DECOMPBELT (card decompbelt; decomp queue item 1, docs/decomp-queue.md). Offline only: no game, no d2d slot; nothing pushed, posted or filed on GitHub. Tags: V = read in game.exe (`tools/emu/disasm.py`) or run in the unicorn oracle by this agent (command given); D = read from the loaded 1.14d tables; S = upstream's `docs/research/re/*.md`; I = inferred from V pieces; U = unverified.

**Answer (P7).** A potion clicked on the ground with the inventory panel CLOSED goes to the BELT first, with no modifier key: into the first column (left to right) whose bottom potion is the same kind (any tier: hp1-5 together, mp1-5 together, rvs/rvl together; other codes only with the same code), at that column's lowest free box within the belt's rows; else, for hp1-5 / mp1-5 / rvs / rvl only (misc.txt `autobelt`), the bottom box of the first EMPTY column; else the pack; no pack room: it stays on the ground ("can't carry"). With the inventory panel OPEN the same click puts it on the CURSOR. The SME was right; our pickit's "keep [C] potion already in the pack" skip assumes the pack. d2d does the same as game.exe for potions; it differs only for Identify/Town Portal scrolls (proven, PR candidate below) and in three client-side gaps (not proven, below).

## Q1 Ground pickup: where the item goes (V unless tagged)

| Step | Address | Rule |
|---|---|---|
| client click | `FUN_00480930` case 4 (0x480afd) | sends C->S 0x16 (type 4, id, action); **action = UI state 1 (inventory panel) via `FUN_004538d0(1)`**: open -> 1, closed -> 0. No Shift/Ctrl test on this path (the click caller `0x46163d` not read: I) |
| server | `FUN_00548b00` (0x548c8e) | action 0 -> `FUN_00563560` auto-place; else `FUN_0055cf50` (to the cursor) |
| 1 | `FUN_00563560`: `FUN_0063c1e0` | something on the cursor: refused |
| 2 | `FUN_0055cc90` | may it be taken; not: message 0x13 |
| 3 | `FUN_0062b400` == 4 -> `FUN_0055c850` | gold |
| 4 | `FUN_00560020` | stacking: a scroll into an inventory tome of its kind with room, a tome into a tome, AutoStack (keys) into stacks (S, combat.md "Picking up"; the call V) |
| 5 | `FUN_0055d710` | a free body location it fits (auto-equip; potions have none) |
| 6 | `FUN_0062bad0` | the item's OWN ItemTypes row Beltable (+0x19); no Equiv chain |
| 7 | `FUN_00628ba0` | autobelt (Items +0x131) -> 1; else `FUN_0063c560` and `cmp eax, -1`: 63c560 returns only 0/1 (its isc/tsc and bottom-match test is dead), so **628ba0 returns 1 for every item with a record** (V read + oracle) |
| 8 | `FUN_0063c790` -> `FUN_0063c600` | the box (below); `FUN_0063afd0(inv, box, 0, grid 1, 0)` places it: **the belt item's grid x = box 0..15, y = 0** |
| 9 | `FUN_005600a0` | the inventory; no room: message 0x17 (can't carry) |

`FUN_0063c600(inv, item, &box)`: 1x1 only (`FUN_006286c0`: Items +0x10f/+0x110); belt type = equipped body slot 8 (`FUN_0063bde0`) -> Items +0x130 (`FUN_00621ed0`), none = 2; boxes from belts.txt (`FUN_00660cb0`, buf +4); the belt grid (`FUN_0063add0`, 0x7447b4: 16x1) boxes[0..15]. Columns c = 0..3: when boxes[c] (the bottom) matches (`FUN_00628a40`: same class, or both in one group: 0x744684 hp1-hp5, 0x74466c mp1-mp5, 0x744660 rvl rvs (D, read from the loaded image)), return the first empty of c, c+4, c+8, c+12 below `boxes`. None: if autobelt, the first empty of boxes 0..3. Else fail.

Beltable 1x1 items (D, ItemTypes +0x19 and Items +0x131 read in the oracle): hp1-5, mp1-5, rvs, rvl, hpo, hpf (autobelt 1); vps, yps, wms, elx, mpo, mpf, rps, rpl, bps, bpl, isc, tsc, 0sc (autobelt 0: only into a column whose bottom is the same code). Keys, tomes, gems, rings, jewels: not Beltable (pack). Scrolls: the tome first (step 4), then a same-scroll column, then the pack.

Oracle: `tools/emu/belt_pickup.py` runs 62bad0, 628ba0 and 63c600 natively, NO hooks (fake units {4, class}, fake inventory); 390,000 cases (every belt type x bottom row over {-, hp1, mp1, rvs, isc} x {bottom only, columns full} x 29 picks = 290,000, plus 100,000 random stacked belts, seed 1). Outcome counts per pick in the evidence pack.

## Q2 Belt model (V unless tagged)

- Belt types (D, belts.txt via 660cb0; armor.txt `belt` = Items +0x130): no belt -> row 2 = 4 boxes (1 row); Sash `lbl` (1) and Light Belt `vbl` (4) = 8 (2 rows); Belt `mbl` (0) and Heavy Belt `tbl` (5) = 12 (3 rows); Plated Belt `hbl` (3) and every exceptional/elite belt (6) = 16 (4 rows).
- Boxes: index = row x 4 + column, row 0 = the bottom (the front row on the HUD strip); a column is boxes c, c+4, c+8, c+12. The belt item's grid position is (x = box, y = 0) in game.exe as in d2d (63afd0's arguments).
- Column kind = its bottom item, at pickup time only (628a40); tiers mix (hp1 under hp4), rvs/rvl mix; hp never joins an mp column. A column with another kind is skipped, its free boxes unused.
- Hotkeys Use Belt 1..4 (`FUN_00498c19`..`0x498d47`) -> `FUN_00498a90(box 0..3)`: the bottom item of that column; `GetKeyState(VK_SHIFT) & 0x8000` sets the flag that feeds the hireling (expansion only, `FUN_0044dcc0`; C->S 0x26). A right-click on any belt box (`0x499daf`, MK_SHIFT the same) uses that box.
- After a use, `FUN_00562390` -> `FUN_0055edc0(box)` compacts column box&3: each present item over rows 0..3 drops to the lowest free row. Potions have no quantity in the belt (one item per box).

## Q3 Shift / Ctrl clicks on a pack item (client `FUN_0048ffe0`, V by read; flags from the mouse message: I)

| Click | Condition | Effect |
|---|---|---|
| left, `GetAsyncKeyState(VK_CONTROL)` < 0 | a store's NPC known (`FUN_004b3230`) | **sell** (C->S 0x33, price `FUN_0062fdc0`); not sellable: message 0x13; no store: nothing |
| left, flags & 4 (MK_SHIFT) | Beltable own type (62bad0) and the item in the inventory page (`FUN_00628250` == 0) | C->S **0x63** -> server `FUN_0054d520` -> the same **63c600** box choice as a pickup (no box chosen by the client) |
| left, otherwise | | to the cursor (0x19) |
| right on a pack potion | | use (0x20 `FUN_005bf240`, S combat.md) |

## d2d's version (upstream main f8aab0e8) and the deviations

| Piece | d2d | Same as game.exe? |
|---|---|---|
| auto-place order | `components/rules/rules.hpp:1907` `pick_up` (Loot::take, loot.hpp:230) | yes, except the body step (upstream ponytail) |
| belt box choice | pick_up's column loop + autobelt fallback; groups `kGroups` | **yes for every potion** (390,000 cases: 0 potion mismatches) |
| isc / tsc | excluded from the belt unless autobelt (`rules.hpp:1936`) | **no: D1**, 3,111 of 390,000 (2,859 isc, 252 tsc), 0 after the fix |
| pickup with the inventory open | `cmd::Pickup` always auto-places (world.cpp:2759, 3114) | **no: D2** (game: cursor). Client UI; not oracle-run |
| shift-click pack -> belt (0x63) | none (only `cmd::ToBelt{box}` with a held item, cursor.cpp:80; our fork's `cmd belt put`) | **no: D3** (feature gap) |
| ctrl-click sell | not found (grep apps/d2d) | **no: D4** (feature gap, U for d2d) |
| hotkeys, shift -> merc, compaction | `drink_belt(col)` bottom box; `drink_at` shifts boxes +4 down; town.cpp:1387 shift | yes (drink_at = 55edc0 for contiguous columns) |

D1 is the PR candidate: `docs/research/pr-belt-{proof,draft,evidence-pack}.md`, branch `trace/belt` (2 commits) in `~/d2d-lab/pr/belt` on the worker host, local only. D2-D4 are client behaviour, not a function differential: written as limits/leads, no PR.

## Our bot's model vs both (findings for the primitives fix queue)

| id | where | bot assumes | game.exe and d2d | effect | fix |
|---|---|---|---|---|---|
| B1 | `internal/loot/loot.go:471` Want "potion" | a picked potion lands in the pack; beyond the plan's fill target it is wanted only while `PackCount < PotionBuffer` | it lands in the belt when its kind's column has a free box (or, hp/mp/rv, a column is empty) | the 939 "potion already in the pack" skips include free belt potions (any BeltFillPct < 100 leaves free matching boxes) | wanted if `beltTakes(kind)`: same-group bottom with a free box in rows, or autobelt kind and an empty bottom box; then no pack cell needed |
| B2 | `internal/loot/pickup.go:345` | any free belt box (`Free() > 0`) means no pack space is needed | only a matching column or an empty column takes it | hp picked while only mp columns have room: it goes to the pack; full pack: "can't carry", retried until "failed before" | use the same `beltTakes` test |
| B3 | `internal/combat/safety.go:33` BeltOf (= review F7) | GX is the column | GX = box 0..15 on BOTH backends (63afd0) | rows 1-3 invisible to combat | column = GX % 4, row = GX / 4 |
| B4 | `internal/town/beltfill.go:141` freeBeltBox (= F8) | lowest free box, any column | pickup and shift-click choose by bottom kind (63c600) | hp above mp | mirror 63c600; on game.exe prefer shift-click (0x63): the game picks the box |
| B5 | real-game input path (P35) | not modelled | pickup with the inventory open goes to the cursor | the next click drops or swaps it | close the inventory before ground pickups; drill it |

## Verified vs not

V: every address above (read), 62bad0/628ba0/63c600/628a40/6286c0/621ed0/660cb0/63add0/63bde0 run in the oracle; the differential (3,111 -> 0, md5 equal); upstream `test_rules`, `test_combat`, `test_game` (27 OK) green on the branch; pr_guard 0/0. D: belts.txt boxes, armor `belt`, misc `autobelt`, ItemTypes Beltable, the three groups. I: MK_SHIFT is bit 4 of the click flags (the mouse message wParam); the left-click caller of 0x480930. U: the cursor -> belt explicit-box rules (0x23, `FUN_0054b3e0` not read), whether d2d's put_in_belt (Equiv chain) matches it; the 1.14d client reaction when a belt hotkey hits an empty column; d2d ctrl-click (only grepped). Not run live (no game).

Drill D1 (m1-liveproof, P7) should expect: belt empty, inventory closed, click 2 hp on the ground -> the first takes box 0 (empty column, autobelt); the second matches column 0 and takes box 4 when the belt has 2+ rows (box 1 with no belt); pack count unchanged, nothing on the cursor. With an mp in box 0 and a sash: hp -> box 1, a second mp -> box 4. With the inventory open: the potion is on the cursor.
