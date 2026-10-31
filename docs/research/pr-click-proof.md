# PR proof: shift-click a beltable inventory item into the belt (C->S 0x63)

> status: current, agent DECOMPCLICK (on Opus), 2026-10-31. Applies docs/process/upstream-pr-standard.md. Branch `trace/click` in the click worktree of the PR checkout on the worker host, two commits on upstream/main 0dc4c62a (after #147 merged): `c4038d15` (the claim), `ff0d8b2f` (tooling: `tools/emu/auto_belt.py`, `test_game --auto-belt`, a `tools/emu/README.md` row). Local only; nothing pushed or posted. Trace: [click-handlers-trace.md](click-handlers-trace.md).

## Claim

A shift + left click on an item stored in the inventory, nothing in hand, moves it into the belt when its own type is Beltable and the belt has a box for it, at the box a pickup would choose (game.exe client `FUN_0048ffe0` 0x49029d sends C->S 0x63; server `FUN_0054d520` -> `FUN_0063c600`). No box: nothing moves (the item stays in the pack, nothing in hand). d2d had no such command: the shift was ignored and the item went into the hand.

## Causal chain (all V: read with upstream `tools/emu/disasm.py` on game.exe 1.14d; the three rule functions also run in the oracle)

| Link | Address | What |
|---|---|---|
| client | `FUN_0048ffe0` 0x49029d..0x4902dd | no item in hand, the player's own item, not ctrl (0x4901f2), flags & 4, `FUN_0062bad0`, `FUN_00628250` == 0 -> `FUN_00478680(0x63, id)`; else 0x4902e8 the plain pickup to the hand (0x19) |
| server gate | `FUN_0054d520` 0x54d526..0x54d5ac | size 5; `FUN_00549150` (item found, mode 0, in the player's inventory); `FUN_0063c1e0` hand empty (else resync, return 2); `FUN_0062bad0`; `FUN_00628250` == 0 (else 3) |
| box | 0x54d5bf `FUN_0063c600(inv, item, &box)`; 0x54d5c6 / 0x54d5d0 | 0 or box < 0 -> `xor eax, eax` return: nothing moves |
| move | 0x54d5db `FUN_00535610`, 0x54d628 `FUN_0063ad90`, 0x54d6e3 `FUN_0063c4f0(inv, item, box)` | out of the inventory, into the belt box |
| box rule | `FUN_0063c600` (belt-pickup-trace.md Q1) | 1x1, matching bottom (`FUN_00628a40`: same code or hp / mp / rv group) column's lowest free box below the belt's boxes; else autobelt the first empty bottom box |

d2d change (`c4038d15`): `rules::beltable` (`FUN_0062bad0`), `rules::belt_box` (`FUN_0063c600`, moved out of `rules::pick_up`, which now calls it: no change to pickups), `rules::auto_belt` (`FUN_0054d520`'s page test and move); `cmd::AutoBelt{item}` (opcode 0x63) in `protocol.hpp`; the World applies it with nothing in hand (`world/commands.cpp`); the client sends it for a shift-click on a Beltable inventory item (`apps/d2d/cursor.cpp`, shift from `SDL_GetModState` in `town/update.cpp`); a joined game sends game.exe's 0x63 (`c2s::auto_belt`, `town/net.cpp`). Tests: `test_rules` (six cases), `test_protocol` (round trip).

## Differential

`tools/emu/auto_belt.py` runs game.exe's `FUN_0062bad0`, `FUN_00628250` and `FUN_0063c600` natively in `FUN_0054d520`'s order, no hooks, on fake units (item {4, class, item data with the page at +0x45}; the belt inventory as in `belt_pickup.py`). d2d side: `test_game --auto-belt` calls `rules::auto_belt` with the belt's boxes from `belt_index`. Cases: page 0 x every belt (8) x bottom row over {-, hp1, mp1, rvs, isc} x {bottom only, full} x 29 items = 290,000, plus 100,000 random stacked belts (10 % on page 3 / 4), seed 1: 390,000.

| | cases | moves (game) | mismatches |
|---|---|---|---|
| before (`upstream/main`: no 0x63 command; d2d leaves every item where it is, or in the hand) | 390,000 | 129,283 | 129,283 |
| after `c4038d15` | 390,000 | 129,283 | **0** (md5 equal, 31c81b34...) |

Pickups unchanged by the `belt_box` move: `belt_pickup.py` (from #147) against `test_game --belt-pickup` on the branch: see the Checks section.

## Not exercised (limits)

- `FUN_0054d520`'s unit lookup, `FUN_00549150`'s owner / mode test and the cursor test are not run; d2d's World finds the item by id among the player's own items (stored only) and the handler tests `held`.
- `FUN_00535610` (V, verifier) refuses during a player trade (playerdata +0x4c / +0x50 > 4) and with some per-player NPC interaction states (`FUN_00573140`); not modelled (`// ponytail:` on `auto_belt`).
- The client's pending-item flag (`FUN_004c2240` / `FUN_004c21f0`: I) is not modelled.
- A joined d2d applies the move locally before the host confirms; whether the host's item packets reconcile a refusal is U.
- The client's sound 0xeb on the click (`FUN_004b9a00`) is not played (U what it is).
- The client path is read, not run; the shift-click input is a UI path a reviewer checks by hand.

## Checks (upstream tests, pr_guard)

On the worker host, `/tmp/click-build` (Release, FETCHCONTENT from the lab's deps, local gcc 16 with `-Wno-error=maybe-uninitialized -Wno-error=array-bounds`), branch at `ff0d8b2f`: `ninja test_rules test_game test_protocol test_combat d2d` ok; `test_rules` ok, `test_protocol` ok, `test_combat` ok, `test_game` rc 0 (18 OK, MPQs + LODPatch_114d.exe); `auto_belt.py` 390,000 cases 0 mismatches; `belt_pickup.py gen 100000 1` (390,000) vs `test_game --belt-pickup` 0 mismatches (pickups unchanged by the `belt_box` move); `pr_guard.py upstream/main` 0 FAIL 0 WARN; `presubmit.py --quick` 0 FAIL 2 WARN: "net roles: the diff touches host, not client" (answered in the body: the World is the host's; d2d's client sends `cmd::AutoBelt`, a joined d2d sends game.exe's 0x63, so both roles are covered) and "build skipped (--quick)" (built and tested by hand above). Diff: 13 files, +259 -23 (claim 10 files, tooling 3).

## Verifier

Independent Opus verifier (read-only, 2026-10-31), report verbatim:

> VERDICT: the claim holds, with limits. I found no counterexample to the rule. The proof doc's limits text needs tightening (below).
>
> **What I ran** (worker host, branch trace/click ff0d8b2f, tree clean, build current; all scratch files are /tmp/click-verify-*)
> - Adversarial generator (/tmp/click-verify-gen.py, seed 4242): each box filled at random, so there are gaps in columns, bottom boxes empty under filled ones, and rows above the belt. Kinds the author's generator lacks: hpf, mpf, rpl, bps. Non-belt picks: tbk, ibk, box, cm1, r01, gpw. 20% on page 3/4. Result: 150,000 cases, 53,199 moves, 0 mismatches (md5 56a51794 on both sides).
> - Author's generator, seed 7: 390,000 cases, 129,218 moves, 0 mismatches (md5 89ec58ec).
> - belt_pickup.py seed 9 against test_game --belt-pickup: 390,000 cases, 0 mismatches (md5 6b32703c). The refactor is also equivalent by reading: same info_of, type and types lookup.
> - test_rules and test_protocol: rc 0. agentreap: nothing of mine left, the list is empty.
>
> **Re-derived (V, disasm)**
> - Server 54d520:
>   - `549150` refuses (returns 1) unless the item unit exists, has mode 0 ([+0x10]) and `63ad50` finds it in the player's inventory. A belt item (mode not 0) is refused there.
>   - `63c1e0` reads inv+0x20, the cursor item.
>   - Mode 0 and the page are checked again at 0x54d60e/0x54d61c.
>   - `63c600` failing or giving box < 0 goes to 0x54d731 and returns 0.
>   - `63c600` refuses anything that is not Beltable or not 1x1 (`6286c0`). The belt size comes from body slot 8. The autobelt fallback has no column < boxes test, which is harmless because there are always at least 4 boxes. d2d matches all of this.
>   - `63c4f0` failing after `63ad90` would leave the item out of the inventory, but it repeats only the Beltable and 1x1 tests, so it cannot fail after `63c600` succeeded.
> - `535610` is now V, no longer I. It refuses during a player trade (playerdata +0x4c != 0 or +0x50 > 4). During an NPC interaction it refuses only if a per-player NPC list entry is >= 1 (`573140`). It returns 0 if the interaction unit is missing.
> - Client: ctrl is tested before shift. Before both, `4c2240(item)` != 0 eats the click (a pending-item flag: I), and `486b30` / `486d10` must pass. The owner test is at 0x4901b6. After sending 0x63 the client calls `4c21f0`, which sets the pending flag: I.
>
> **d2d side**
> - Only update.cpp:433 passes shift. The IdentifyWith, Imbue, right-click and pad callers default to false, so they still get ToCursor.
> - In a joined game, net_items sends 0x63 and then the local World applies the move optimistically. Whether the host's item packets reconcile a refusal is U.
>
> **Suggested limits text**
> "`FUN_00535610` (V) refuses during a player trade (playerdata +0x4c / +0x50 > 4) and with some per-player NPC interaction states; d2d has no such states. The client's pending-item flag (`4c2240` / `4c21f0`) is not modelled. The joined-game path applies locally before the host confirms."
>
> Nothing committed (read-only brief).

PROOF: complete. The limits below take the verifier's text.
