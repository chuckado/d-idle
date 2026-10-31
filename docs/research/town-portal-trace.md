# Town portal from the pack: game.exe 1.14d traced vs d2d vs the bot (agent DECOMPTOME, 2026-10-31)

> status: research, current. Decomp lane item 2, driven by P10 (docs/primitives.md) and F4/F5 (docs/primitives-review.md). Tags: V = read in game.exe with upstream's `disasm.py` or run in game.exe's own code under unicorn by this agent (`tools/emu/tome_use.py`); V-d2d = read in d2d (upstream/main f8aab0e8, read-only worktree `trace/tome` on the worker host); D = 1.14d patched tables via `mpq-cat` (docs/process/mpq-data-reading.md); S = upstream `docs/research/re/` (objects.md "Town portals", network.md dispatch); U = unverified; I = inference. Offline only: no game, no d2d slot; nothing pushed, posted or filed.

## Answers in four lines

1. **Q1 empty tome: a silent no-op in both.** game.exe's C->S 0x20 handler refuses a book (type 0x12) whose quantity (stat 0x46 = 70) is <= 0 before anything else: no portal, no charge, no voice, no packet, no error flag (V). d2d: `read_portal` returns at `book && quantity <= 0` (V-d2d). Oracle: 5,743 empty-tome cases, 0 differ. So "no portal after the click" is the correct observation of an empty tome; nothing is cast.
2. **Q2 game.exe opens the portal AT the click, with no cast.** Inventory use runs pSpell 2 straight to the portal object: 0 ticks, any player mode (GH, attack, cast), refused only when dead, under state 54 `uninterruptable`, with an item on the cursor, or in town (V, oracle 108,000 cases). Position: nearest free spot from the caster (size 3, mask 0x3e01, V); twin in the act town (S); one pair per player, the old pair closed first (V). The SC cast (skill 220, srvdofunc 113) is only the skill-bar path.
3. **Q3 d2d deviates:** a right-click starts an SC cast; the portal comes at the SC action frame (sorceress, FCR 0: about 7 ticks, 280 ms); the click is DROPPED while the player is in any mode (pmode >= 0: GH, attack, cast, block); the charge is spent at the click, so a get-hit during the SC loses it with no portal. Empty tome, town and cursor match.
4. **Q4 the bot's failures are not timing.** Of 778 "no portal opened" lines: 525 empty tome (67 %), 134 dead context (17 %), 119 tome still charged (15 %, d2d's busy drop or a cut cast, I). The bot's wait (8 x 200 ms after the click) is 5x the SC action frame.

## 1. game.exe's chain (V unless marked)

| Address | What it does | Tag |
|---|---|---|
| FUN_0054d750 | C->S dispatcher (table 0x6e0d18): packet 0x20 passes unless FUN_0057eec0(player, 1, 0, 0) is 0 | V |
| FUN_0057eec0 | with these args: 0 (refuse) for state 0x36 (States.txt 54 `uninterruptable`, D), mode 0 DT or 0x11 DD; else 1. The player's other modes are not read | V |
| FUN_0054b1e0 | the 0x20 handler: item in the player's inventory (FUN_00549150), the packet's x,y within range of the player (FUN_00548ef0), then FUN_0055e170 | V |
| FUN_0055e170 | item by id (type 4), no cursor item (FUN_0063c1e0), item mode +0x10 == 0, usable (FUN_00628c20); **book: stat 0x46 > 0 else return 0** (`0x55e225`..`0x55e236`); not trading (FUN_005678a0); FUN_005bf240; on success a book's stat 0x46 += -1 (FUN_006272b0 at `0x55e2a2`) and the stat packet; a scroll is removed (FUN_0055e000) | V |
| FUN_005bf240 | pSpell from books.txt (D: tsc/tbk pSpell 2) via FUN_006374b0; table 0x741790 entry 2 = {client 0, server FUN_005be290} | V |
| FUN_005be290 | not a player -> 0; town (FUN_0061ab00) or level 0x88 -> voice event 0x18, return 0; FUN_00535430 closes the old pair; FUN_0056d130 opens; voice event 7 | V |
| FUN_00535430 | player data +0x48 -> the old object 0x3b and its linked twin removed | V |
| FUN_0056d130 | town -> refuse; FUN_0064e810(size 3, 0x3e01, 0xc01) from the caster; object 59 (TP, D); twin in the act town (FUN_0056cf40, spawn index 11) | V (call args), S (twin spot) |
| FUN_005bf3d0 | srvdofunc 113 (skills 219/220, anim SC, checkfunc 5 not in town, D): at the action frame finds the first tome with BookSkill == skill and stat 0x46 > 0 (or a scroll) and calls FUN_0055e170 / FUN_00562390 | V |
| Object 59 | OP 15 frames, FrameDelta 200, then ON; OperateRange 2; OperateFn 15 (FUN_00584870): through to the twin; the owner going back closes the pair | D, S |

No mana, no cooldown and no delay on the inventory path (none read in the chain, V). Portal lifetime: no timer seen in the chain; it lasts until replaced, used back by its owner or the game ends (S, timer U).

## 2. d2d (V-d2d, file:line at f8aab0e8)

| Item | d2d | Verdict |
|---|---|---|
| Empty tome | world.cpp:2162 `if (book && scroll->quantity <= 0) return;` | SAME |
| Town, dead, cursor | :2160 `in_town` / `fight.dead()`; apps/d2d/town.cpp:1596 `!held` sends nothing | SAME (d2d plays no voice 0x18 in town) |
| Busy | :2160 `actor->pmode >= 0` and fight.cpp:4439 return: the UseItem is dropped in GH, A1/A2/S1..S4, SC, BL | DIFFERS: game.exe opens |
| Cast | :2163 `cast_scroll(220/219)` sets SC; :2164 `--quantity` / erase at the click; fight.cpp:2409 `portal_due` at the action frame; world.cpp:3203 `open_portal` | DIFFERS: game.exe opens at once, no SC |
| Cut cast | fight.cpp:4973: a hit sets GH over SC (SC is not `attack_mode`): the charge is gone, no portal | DIFFERS (game.exe has nothing to cut) |
| No free spot | :2205 logs and returns: charge spent, old pair kept | DIFFERS: game.exe keeps the charge, old pair closed |
| State 54 | no test | DIFFERS (rare: 278 of 108,000 oracle cases had it) |
| Position, twin, one pair | `portal_spot` (size 3, 0x3e01, 0xc01), `nearest_free(portal_spot)`, `portal[pair]` overwritten | SAME (V-d2d read; upstream deviations.md row 8) |
| Going through | `use_portal` :2214: within 2 cells; the owner back through closes the pair (:2253) | SAME shape (S) |

Oracle (`tools/emu/tome_use.py --exe game.exe 100000 7`, 108,000 cases: book/scroll x quantity -2..20, 255, 2^20 x town x cursor x modes 0..19 x free spot, then random incl. state 54): same 85,605 (all "nothing", the empty tome 5,743 of them); game.exe now / d2d at the SC frame 3,109; game.exe now / d2d nothing (busy) 7,930; no free spot 3,116 + 7,962 (busy); state 54 278. Hooks are listed in the script's docstring.

## 3. The bot (Q4)

`d2dPortal` (cmd/d2bot/d2dwire.go:210-279): picks the FIRST Portal item with no charge check (F4), presses `i`, sleeps 700 ms, right-clicks the `debug itemat` pixel, polls 8 x 200 ms for object 59 within 25 or the item's id gone; three tries. `HasTP` (internal/bot/resource.go:57) is true for any "Town Portal" item (F5). The failure lines (`tools/emu/tome_soaklines.py`, read-only stream over navsoak, navsoak2, navsoak3 (still running), m2early, goldab bot logs, run 2026-10-31):

| Cause (bucket) | navsoak | navsoak2 | navsoak3 | m2early + goldab | Total |
|---|---|---|---|---|---|
| Empty tome: tp = 0 at the last shop visit | 206 | 9 | 81 | 27 | 323 |
| Empty tome, estimate: tp minus trips since the shop <= 0 | 83 | 1 | 118 | 0 | 202 |
| Dead context: `context canceled (... i/o timeout)`, three tries in ~10 ms, nothing clicked | 44 | 12 | 65 | 13 | 134 |
| Tome still charged (estimate) | 40 | 23 | 53 | 3 | 119 |

After a click failure the next trip event (navsoak*, 520 with one) is the waypoint fallback (`no town portal`: Town.ToTown saw no object 59 either) in 425, an arrival in town in 25 (I: a late portal or a walk), a field level in 39, abandon or chicken in 31: the portal was never created, not late and not elsewhere. In the charged bucket 50 had hostiles near (d2d's busy drop or a cut cast fits, I) and 69 none (I: the estimate, or the `i` toggle closing an already open panel so the click went to the world, U).

**What the drill D2 (P10) should expect on d2d today:** idle player, charged tome, inventory open, right-click -> SC starts, tome quantity -1 at once, object 59 at the nearest free spot by the player about 7 ticks later (FCR 0), its twin in town; empty tome -> nothing at all; in GH / attacking -> nothing (retry when idle); a hit during the SC -> charge gone, no portal. On game.exe the portal exists on the click's tick in every live mode.

## 4. Findings for FIXPORTAL (bot side)

1. Pick by charges: a tome with stat 70 > 0, else a loose scroll; never click an empty tome (V both engines). Same count for `HasTP` and the kit gate.
2. `d2dPortal` must check `ctx.Err()` before its tries and the caller must hand it a live context (134 lines; which context dies is U).
3. On d2d click only when the player is idle (not GH / attacking / casting; stop the fight loop first) and judge success by object 59 only: a scroll that left the pack counts as success today (d2dwire.go:261-268) though a cut SC spends it with no portal.
4. The wait is long enough (1.6 s vs ~0.3 s); no timeout change needed. Re-read the quantity before each retry: a cut cast spends a charge per try.

## 5. Upstream PR candidate

d2d's inventory TP (read_portal) differs from game.exe in four ways (cast vs instant, busy drop, charge on a cut cast or no spot, state 54). Proof of the game.exe side is a native run; the d2d side is a line copy (allowed as "before"). Not ready: no patch, no d2d-binary differential, no verifier, upstream tests not run. See [pr-tome-proof.md](pr-tome-proof.md), [pr-tome-evidence-pack.md](pr-tome-evidence-pack.md) (PROOF: incomplete), body draft [pr-tome-draft.md](pr-tome-draft.md). The local branch `trace/tome` on the worker host has no commits.
