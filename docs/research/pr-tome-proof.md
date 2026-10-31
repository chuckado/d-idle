# Proof: a Town Portal read from the pack opens at once, in any live mode (fix/tome-use)

> status: current (agent TOMEPR, 2026-10-31; game.exe trace and first oracle by DECOMPTOME). Applies docs/process/upstream-pr-standard.md. Branch `fix/tome-use` in the tome worktree of the PR checkout on the worker host, three commits on upstream/main 27113a31: `d72b97a7` (tooling: `tools/emu/tome_use.py`, `tests/test_game.cpp --tome-use`, a `tools/emu/README.md` row), `112f5de0` (claim A), `c56670b9` (claim B). Local only; nothing pushed or posted. Trace: [town-portal-trace.md](town-portal-trace.md).

## Claims (two, one commit each)

**A (112f5de0).** game.exe's `FUN_0055e170` refuses the use when the player holds an item on the cursor (`FUN_0063c1e0`); d2d's World had no such test (its client sends nothing then, apps/d2d/town.cpp). Fix: `read_portal` returns when `actor->held`.

**B (c56670b9).** A Scroll of Town Portal, or a Tome of Town Portal with stat 0x46 > 0, used from the pack (C->S 0x20) opens the pair on that tick with no SC, whatever live mode the player is in; game.exe closes the old pair first (`FUN_00535430`) and spends the charge or the scroll only when `FUN_0056d130` made the portal. d2d cast SC (`Fight::cast_scroll`), dropped the use while `pmode >= 0`, opened at the SC action frame, spent the charge at the click and kept the old pair when no spot was free. Fix: `read_portal` closes the pair, calls `open_portal_at` (now `bool`), spends on success; `cast_scroll` and `item_cast` (no other user) removed.

The empty tome is a no-op in both before and after (not claimed). State 0x36 and level 0x88 also refuse in game.exe: out of scope (d2d's player has no state 0x36; Uber Tristram's not compared), `// ponytail:` in `read_portal`.

## The chain, tagged

| Link | What it shows | Tag |
|---|---|---|
| `FUN_0054d750` | 0x20 passes unless `FUN_0057eec0(player, 1, 0, 0)` is 0 | V (read, run) |
| `FUN_0057eec0` | 0 for state 0x36, mode 0 or 0x11; no other mode read | V (read; the oracle runs modes 0..19) |
| `FUN_0054b1e0` | owner (`FUN_00549150`), range (`FUN_00548ef0`), then `FUN_0055e170` | V (read, run; owner and x, y hooked) |
| `FUN_0055e170` | cursor (`FUN_0063c1e0`), item mode 0, usable (`FUN_00628c20`), tome stat 0x46 > 0 (0x55e225..0x55e236), not trading (`FUN_005678a0`), `FUN_005bf240` at 0x55e25d (`je 0x55e3b0` on 0); after success stat 0x46 -1 (`FUN_006272b0`, 0x55e2a2) or the scroll removed (`FUN_0055e000`) | V (disasm.py read this card, run) |
| `FUN_005bf240` | books.txt pSpell 2 -> table 0x741790 entry 2 server `FUN_005be290` | V (read, run), D (books.txt) |
| `FUN_005be290` | not town (`FUN_0061ab00`), not level 0x88; `FUN_00535430` then `FUN_0056d130`; no mode word written | V (read, run) |
| `FUN_005bf3d0` | srvdofunc 113: the skill bar's SC; not on this path | V (read) |

No U link. The hooks (the script's docstring) are the item lookup, owner, range, the cursor test (case input), usable (met), item type and stat 0x46 (case inputs), trade, rows, flags, room and level (Blood Moor), the town test (case input), `FUN_0056d130`'s result (spot / no spot: case input), packets and voice. Recorded: `FUN_00535430`, `FUN_0056d130`, `FUN_006272b0`, `FUN_0055e000`.

## The differential (V)

`tome_use.py gen 100000 7`: 108,000 cases = 8,000 grid (tbk / tsc x charges -2..20, 255, 2^20 x town x item in hand x modes 0..19 x free spot) + 100,000 random (20 % charges up to 2^24). `game` runs `FUN_0054d750` natively; `test_game --tome-use` runs d2d's `World::apply(cmd::UseItem)` in the Blood Moor (no spot: a copy with every subtile walled), the old pair set before each case, then up to 25 world ticks. Each line `id opened spent kept`.

| Build | Lines differ (of 108,000) | Classes (game / d2d) |
|---|---|---|
| d72b97a7 (before) | 29,551 | busy + spot: open now / nothing 8,360; busy, no spot: old pair closed / nothing 8,337; idle + spot: open now / at tick 9 3,194; idle, no spot: charge kept, old pair closed / charge spent, old pair kept 3,236; item in hand: nothing / cast (6,424) |
| 112f5de0 (claim A) | 23,127 | the item-in-hand class gone; the rest as above |
| c56670b9 (claim B) | 0 (md5 6c9d11e0... equal) | same: nothing 84,873, open now 11,554, old pair closed only 11,573 |

Second seed, `gen 200000 31` (208,000 cases): before 56,616, claim A 44,392, claim B 0.

Their checks on c56670b9: `ninja test_game test_combat test_rules` (Release, gcc 16, `D2D_WERROR` on: no warning), `test_game` rc 0 (29 OK incl. the two new blocks: an item in hand does nothing; a read in get-hit opens on the tick, the second read replaces the pair, the empty tome and a dead reader do nothing), `test_combat` rc 0, `test_rules` rc 0; `pr_guard.py upstream/main` 0 FAIL 0 WARN; `presubmit.py --quick` 0 FAIL 4 WARN (ruff absent; build by hand; siblings unit kinds and hit paths: answered in the body).

## Callers (nothing else moved)

- `read_portal`: one caller, `World::apply` `cmd::UseItem` (world.cpp, tsc / tbk). Local co-op: each player's own `portal_pair`, the same function. Joined game: the use goes to the host (a d2d host's World runs the same `read_portal`).
- `open_portal_at`: void -> bool; the other callers (world.cpp `open_portal` for `portal_due`, the two quest portals at world.cpp ~832 / ~1956, world_quests_act2.cpp ~100) ignore the result: unchanged.
- `Fight::cast_scroll`, `Hero::item_cast`: no other user (grep); the ItemCastSound test in `set_pmode` keeps `from_charges` (item_cast was only true inside cast_scroll's `set_pmode`).
- The skill bar's SC (`portal_due` at the SC action frame): unchanged.

## Limits

- State 0x36, level 0x88: not compared (ponytail).
- Hooked: `FUN_00628c20` (requirements always met), `FUN_0056d130`'s spot search (only whether a spot exists is compared; where the portal stands is d2d's existing `portal_spot`, not part of the claim).
- The item is always in the pack (item mode 0); belt and stash items, and the client side of game.exe (what its client sends), are outside the generator.
- A dead player is not ticked in the d2d mode (its death would run); nothing pends in either build there (no SC starts while dead).
- Sounds (ItemCastSound, voice event 7 / 0x18) not compared.
- `tools/d2ddet` scenario hashes not run (no scripted scenario reads a tome).

## Verifier

PENDING.
