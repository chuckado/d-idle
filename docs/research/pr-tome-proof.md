# Proof: a Town Portal read from the pack opens at once, in any live mode (fix/tome-use)

> status: current, PROOF complete (agent TOMEPR, 2026-10-31; game.exe trace and first oracle by DECOMPTOME). Applies docs/process/upstream-pr-standard.md. Branch `fix/tome-use` in the tome worktree of the PR checkout on the worker host, three commits on upstream/main 03f42254: `29e75c76` (tooling: `tools/emu/tome_use.py`, `tests/game/test_game.cpp --tome-use`, a `tools/emu/README.md` row), `c5e2a0b1` (claim A), `e35ab5ca` (claim B). The proof and the verifier ran first on 27113a31 (`d72b97a7` / `112f5de0` / `c56670b9`, branch `fix/tome-use-27113a31`); see the rebase section. Local only; nothing pushed or posted. Trace: [town-portal-trace.md](town-portal-trace.md).

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
- Sounds (ItemCastSound, voice event 7 / 0x18) not compared; game.exe plays event 7 even when no spot is free.
- Not modelled: the item having to be stored (item mode 0), its requirements, the 50-subtile range test (`FUN_00548ef0`); the joined-game path not run.
- `tools/d2ddet` scenario hashes not run (no scripted scenario reads a tome).

## Verifier

Independent second agent (Opus, its own detached worktree, read-only on the author's), report verbatim (it checked the branch on 27113a31, before the rebase below):

> VERDICT: holds with limits
>
> I checked branch fix/tome-use (d72b97a7, 112f5de0, c56670b9 on 27113a31) in my own detached worktree on the worker host. Every step below is tagged V: I either read it in the disassembly or ran it.
>
> 1. V, disassembly, the gate. For 0x20, the dispatcher FUN_0054d750 looks up byte table 0x54d848 and gets index 1. That index calls FUN_0057eec0 with arguments (1,0,0). Inside it, argC==0 jumps straight to "return 1" once three tests are done: state 0x36, unit NULL, mode 0 or 0x11. No other mode is refused. Most other packets go through the same call.
> 2. V, disassembly, the use-item path. FUN_0055e170 has no mode test, no delay and sets no SC. Its order is: held item (0x63c1e0), item mode +0x10==0, requirements (0x628c20), then for a tome stat 0x46 > 0. Next come the trade test FUN_005678a0 and then FUN_005bf240. The stat 0x46 -1 (FUN_006272b0) or the scroll removal (FUN_0055e000) happens only on the branch where FUN_005bf240 returned non-zero.
> 3. V, disassembly, the portal path. The pSpell table at 0x741790, entry 2, holds {0, 0x5be290}, so the item flag test at 0x6280a0 is never reached for TP. FUN_005be290 tests town (0x61ab00) and level 0x88 first, then calls FUN_00535430 before FUN_0056d130, and returns the result of 0x56d130. A spot that fails therefore closes the old pair, spends nothing, and returns. FUN_0056d130 has no mode test.
> 4. V, run. The author's numbers reproduce exactly. Seed 7 (108,000 cases): 29,551 differ before, 23,127 after claim A, 0 after claim B. Seed 31 (208,000 cases): 0. The new seed 1234 (108,000 cases): 0. In the outcome table each class gives one result. Mode matters only as dead or alive. Town, held, empty tome or dead means nothing happens and the pair is kept. A failed spot closes the pair and spends nothing.
> 5. V, run, a scratch driver of my own using the real World::apply path:
>    - A tome going from 1 to 0 opens; reading it again does nothing and keeps the pair.
>    - Two scrolls read in a row: the second replaces the pair.
>    - An old pair with no free spot: both portals close and the scroll is kept. The log says "no room", so the walled copy really means no spot.
>    - Modes 0..19: pmode, mode_ms and swing_skill are untouched after a read. Only 0 and 17 refuse.
>    - A read during a skill-bar TP cast opens at once, the SC cast continues, and its action frame replaces the pair.
>    - Co-op: player B (pair 12) reads while in get-hit, and player A's pair is untouched.
> 6. V, read. cast_scroll was the only place that set item_cast. The other open_portal_at callers ignore the new bool. The skill-bar srvdofunc 113 path (portal_due, then open_portal) is unchanged.
> 7. V, run. Plain test_game passes (29 OK lines, including both new ones); test_combat and test_rules pass. pr_guard against 27113a31 reports 0 FAIL, 0 WARN.
>
> Limits to disclose
> - The oracle hooks 0x56d130, so the portal's position and its twin are not compared. Only whether it opened, as a per-case input, is compared.
> - Sound and voice are not compared. game.exe sets event 7 (FUN_00553380) even when the spot fails, and voice 0x18 on a town refusal. d2d plays its cue only on success and stays silent in town (that silence was already there before the patch).
> - Not modelled: state 0x36, level 0x88, the item having to be stored (+0x10==0), requirements, the 50-subtile range test in FUN_00548ef0, and any client-side gating in game.exe's client. Nothing was verified live; there is no game run behind this.
> - The joined-game path in apps/d2d (local World plus host) was not run.
> - upstream/main has moved to 03f42254 ("Split the big files into a directory per subsystem"). The branch now gives modify/delete conflicts on world.cpp and fight.cpp and needs a rebase before the PR.
>
> Left behind: my worktree was removed with worktree remove --force. No processes are left (procs registry for tomepr-verify is empty). The author's worktree is untouched (clean, at c56670b9). Scratch files (disassembly dumps, cases, outputs, before binaries, the scratch driver source) are kept in ~/d2d-lab/pr/tome-verify-scratch/. I ran one git fetch of upstream; it changed no refs (the move to 03f42254 at 12:44 came from an earlier fetch).

Acted on: the rebase (next section). Limits added from the report: game.exe plays voice event 7 even when the spot fails (d2d's cue only on success); the item-stored test, requirements and the 50-subtile range test are not modelled; the joined-game path was not run.

## Rebase onto upstream/main 03f42254 (the file split)

Upstream split world.cpp / fight.cpp into directories after the verifier's run. The three commits were replayed hunk by hunk onto 03f42254 as `fix/tome-use`: `29e75c76` (tooling), `c5e2a0b1` (claim A), `e35ab5ca` (claim B); the pre-split branch stays local as `fix/tome-use-27113a31`. The changed lines are identical (the sorted `+` / `-` lines of both diffs compared: equal); the files are now `components/game/world/travel.cpp`, `fight/tick.cpp`, `fight/setup.cpp`, `tests/game/test_game.cpp`. Re-run on the rebased commits (V): seed 7 29,551 -> 23,127 -> 0 (md5 6c9d11e0... equal to game.exe's), seed 31 56,616 -> 44,392 -> 0; `test_game` rc 0 (20 OK in this file after upstream's split, both new blocks among them), `test_combat` rc 0, `test_rules` rc 0; `pr_guard.py upstream/main` 0 FAIL 0 WARN; `presubmit.py --quick` 0 FAIL, the same 4 WARN.

**PROOF: complete** (differential 29,551 -> 23,127 -> 0 over 108,000 author cases, 0 over 208,000 more and the verifier's 108,000 fresh cases; chain V; checks green; verifier holds with limits).
