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
- `FUN_00535610` (a trade / NPC interaction check before the move: I) is not modelled: `// ponytail:` on `auto_belt`.
- The client's sound 0xeb on the click (`FUN_004b9a00`) is not played (U what it is).
- The client path is read, not run; the shift-click input is a UI path a reviewer checks by hand.

## Checks (upstream tests, pr_guard)

On the worker host, `/tmp/click-build` (Release, FETCHCONTENT from the lab's deps, local gcc 16 with `-Wno-error=maybe-uninitialized -Wno-error=array-bounds`), branch at `ff0d8b2f`: `ninja test_rules test_game test_protocol test_combat d2d` ok; `test_rules` ok, `test_protocol` ok, `test_combat` ok, `test_game` rc 0 (18 OK, MPQs + LODPatch_114d.exe); `auto_belt.py` 390,000 cases 0 mismatches; `belt_pickup.py gen 100000 1` (390,000) vs `test_game --belt-pickup` 0 mismatches (pickups unchanged by the `belt_box` move); `pr_guard.py upstream/main` 0 FAIL 0 WARN; `presubmit.py --quick` 0 FAIL 2 WARN: "net roles: the diff touches host, not client" (answered in the body: the World is the host's; d2d's client sends `cmd::AutoBelt`, a joined d2d sends game.exe's 0x63, so both roles are covered) and "build skipped (--quick)" (built and tested by hand above). Diff: 13 files, +259 -23 (claim 10 files, tooling 3).

## Verifier

PENDING_VERIFIER
