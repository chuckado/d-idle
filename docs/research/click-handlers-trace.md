# Shift-click, ctrl-click and inventory-open pickup traced in game.exe 1.14d: d2d compared

> status: research, 2026-10-31, agent DECOMPCLICK (card decompclick; decomp queue item 9, docs/decomp-queue.md), on Opus. Offline only: no game, no d2d slot; nothing pushed, posted or filed on GitHub. Follows [belt-pickup-trace.md](belt-pickup-trace.md) (gaps D2-D4). Tags: V = read in game.exe 1.14d with upstream's `tools/emu/disasm.py` or run in the unicorn oracle by this agent; D = loaded 1.14d tables; S = upstream `docs/research/re/*.md`; I = inferred from V pieces; U = unverified.

**Verdicts.** D3 (shift-click to the belt) is a PROVEN world-level gap: game.exe's server has a rule (C->S 0x63, `FUN_0054d520`) that d2d's World has no command for; PR candidate `pr-click-*` (below). D4 (ctrl-click sell) is CLIENT-ONLY: d2d's World already has the sell (`cmd::Sell`, C->S 0x33), the shortcut is input handling; no PR. D2 (pickup with the inventory open) is CLIENT-SIDE: the client picks the packet's to-cursor flag from the panel state and the server's cursor branch is a plain move; no rules oracle worth a claim; no PR.

## Q1 Shift + left click on a pack item (V unless tagged)

| Step | Address | Rule |
|---|---|---|
| client grid click | `FUN_0048ffe0`, no item in hand (0x490183), the item at the cell (`FUN_0063bd10`), owner = the player (0x4901b6) | ctrl tested first (Q2); then `test byte [ebp+0x18], 4` (0x49029d: the click flags, MK_SHIFT: I) |
| client gate | 0x4902a3 `FUN_0062bad0` (the item's OWN ItemTypes Beltable) and 0x4902ae `FUN_00628250` == 0 (item data +0x45 = page 0: the inventory; stash 4 / cube 3 / belt 0xff refused) | pass: C->S **0x63** (u32 item id) + sound 0xeb (`FUN_004b9a00`); fail: the plain click, item to the cursor (0x19, 0x4902e8) |
| server | `FUN_0054d520` | size 5 else 3; `FUN_00549150` (the item is the player's, mode 0 stored, in its inventory) else 1; item in hand (`FUN_0063c1e0`) -> resync, nothing; `FUN_0062bad0` again; `FUN_00628250` != 0 -> 3 |
| box | 0x54d5bf `FUN_0063c600(inv, item, &box)` | the SAME function as the pickup's belt step: 1x1, matching bottom column's lowest free box within the belt's rows, else (autobelt) the first empty bottom box. **No box: return 0, nothing moves** (the item stays in the pack; nothing on the cursor) |
| move | `FUN_00535610` (trade / NPC check: I), `FUN_0063ad90` out of the inventory, `FUN_0063c4f0(inv, item, box)` into the belt, update packets | |

Kinds: every 1x1 item whose own type is Beltable (D, see belt-pickup-trace Q1): hp1-5, mp1-5, rvs, rvl, hpo, hpf (autobelt: also into an empty column) and vps, yps, wms, elx, mpo, mpf, rps, rpl, bps, bpl, isc, tsc, 0sc (only into a column whose bottom is the same code or group). Unlike a pickup there is **no stacking step** first (a scroll does not go into a tome) and no `FUN_00628ba0` (it returns 1 anyway). Non-Beltable items (keys, tomes, gems, charms) with shift are a plain click: to the cursor.

## Q2 Ctrl + left click (V unless tagged)

- Client `FUN_0048ffe0` 0x4901f2: `GetAsyncKeyState(VK_CONTROL)` < 0 (after `FUN_00486d10`), on an item the player owns, no item in hand. Store known (`FUN_004b3230` -> NPC id, `FUN_00463990` the unit): `FUN_0062a130` sellable -> price `FUN_0062fdc0` and C->S **0x33** (NPC id, item id, item mode, price); not sellable: message 0x13 ("impossible"). **No store: nothing at all** (0x4902f8: the click is eaten; the item does NOT go to the cursor).
- Sellable (`FUN_0062a130`): item flag 0x1000 clear (item data +0x18; the "no sell" flag: I, D2MOO naming) and `FUN_00629d40`: Items +0x12a (quest) < 1 and type +0x11e != 0x27 (quest items refused).
- Server `FUN_0054bb20` (size 17) -> `FUN_00579510`: the NPC must be the player's interaction target (`FUN_00554d00`), the trade open (`FUN_00557ff0`, else message), the packet's mode = the item's mode, `FUN_0062a130` again; the price is the server's own `FUN_0062fdc0` (0x57961b, 0x5798cd: the lower of two); the client's cost field is not used in the parts read (I). So the normal sell price.
- Sources: any grid `FUN_0048ffe0` handles (the inventory while trading; stash and cube cannot be open with a store). **Belt items: no.** The belt's left-click handler (0x498870) has no modifier test: an item there goes to the cursor (0x24) with or without ctrl (V).

## Q3 Pickup click with the inventory open (V)

Client `FUN_00480930` case 4 (0x480afd): 0x16's third dword = `FUN_004538d0(1)` (the inventory panel open). Server `FUN_00548b00` (0x548c8e): 0 -> `FUN_00563560` (auto-place: belt / pack); non-0 -> `FUN_0055cf50`: hand not empty -> nothing; item on the ground (mode 3), may it be taken (`FUN_0055cc90`, else message 0x13); gold is added as gold; anything else **to the cursor** (`FUN_0063c180`). Pack space is never checked: a full inventory still takes it into the hand.

## Q4 d2d upstream main (03f42254, after #148), read

| Gap | d2d today | Layer | Verdict |
|---|---|---|---|
| D3 shift-click | no 0x63 command in `components/game/protocol.hpp`; upstream's own `docs/research/re/net-packets.md` lists 0x63 auto-belt as "none" (S); the client never reads shift on item clicks (`apps/d2d/cursor.cpp`; only belt hotkeys and stat points read SDL shift) | World (missing command) + client | **proven gap**; oracle-checkable (the box is a rules function) |
| D4 ctrl-click | `cmd::Sell` exists (store grid, held item, pad X; `apps/d2d/town/update.cpp`, `input.cpp`); no ctrl test | client only | client feature, NOT a rules deviation; no PR |
| D2 open-panel pickup | `cmd::Pickup{item}` has no to-cursor flag; `town/input.cpp:75` sends it with any panel open; World auto-places (`rules::pick_up`) | client choice + a missing flag | client-side; the server branch is a move to the hand; not proven by an oracle; no PR |

The lab fork implements D3 and D4 (and a shared belt-slot rule); this trace did not copy it. The candidate below is written fresh on upstream main.

## Q5 Our bot

- Belt fill (`internal/town/beltfill.go`): on game.exe it shift-clicks (the game picks the box: matches Q1); on d2d it uses the lab's `belt put` verb or drags (shift gives `ErrUnsupported`). Selling uses the shop UI (drag to the store), not ctrl. Pickups close panels first (belt-pickup-trace B5).
- So none of the three gaps blocks a bot verb; they matter for live-play parity and for any future bot path that wants one input model on both backends (shift-click belt fill on d2d would drop the lab-only verb). Recommendation: D3 build the PR candidate, low priority, bundle it when a bundle has room; D2 / D4 note only.

## PR candidate (D3)

Branch `trace/click` in the click worktree of the PR checkout on the worker host, local only. State, numbers and commits: [pr-click-proof.md](pr-click-proof.md), [pr-click-draft.md](pr-click-draft.md), [pr-click-evidence-pack.md](pr-click-evidence-pack.md); oracle `tools/emu/click_auto_belt.py` (upstream's copy: `tools/emu/auto_belt.py`).

## Verified vs not

V: every address above (read with disasm.py on game.exe 1.14d); `FUN_0062bad0`, `FUN_00628250`, `FUN_0063c600` run in the oracle. I: MK_SHIFT = bit 4 of the click flags; flag 0x1000 = no-sell; the sell cost field unused; `FUN_00535610`'s trade gate. U: what sound 0xeb is; whether the 1.14d client greys out anything on a refused shift-click. Not run live (no game).
