# Waypoint travel: game.exe 1.14d traced vs d2d vs the bot (agent DECOMPWP, 2026-10-31)

> status: research, current. Decomp lane item 3, driven by P23 (docs/primitives.md). Tags: V = read in game.exe with upstream's `disasm.py` or run in game.exe's own code under unicorn (`tools/emu/wp_refuse.py`); V-d2d = read in d2d (upstream/main 03f42254, read-only worktree on the worker host); D = 1.14d Levels.txt via `mpq-cat`; S = upstream `docs/research/re/`; I = inference; U = unverified. Offline: no game, no d2d slot; nothing pushed, posted or filed.

## Answers in four lines

1. **Q1 panel:** rows are Levels.txt `Waypoint` order per act (9 / 9 / 9 / 3 / 9; our `wpLevels` is identical, D); the panel opens on the tab of the player's current act; one click on an activated row sends 0x49 and travels, no confirm (V).
2. **Q2 the 10 s window is real but keyed on PvP hostility, not on a level change.** The server refuses 0x49 while `GetTickCount() < pcdata+0x160 + 10000`; the only writer of pcdata+0x160 is the hostile-declaration path (V, whole-.text scan). In single player it never fires. **d2d refuses for 10 s after every arrival: a proven deviation** (132,964 oracle cases, 59,309 differ; 0 after the patch).
3. **Q3 transition:** same act: the warp runs inside the packet handler (0 server ticks, V call chain; client draw delay I); arrival at the destination waypoint's x+3, y+3 subtiles, which is lit (V call args); the portal pair is not touched (V for the handler, I below it).
4. **Q5 none of the 295 "nav waypoint failed" lines reached the panel:** all but 3 are `unit not found`: the d2d state reader files the waypoint as an object with class -1, so the bot never sees a waypoint (bot finding W1). Zero "wp: used" successes in all four soaks. Refusal, row and timing are not the cause.

## 1. game.exe's chain (V unless marked)

| Address | What it does | Tag |
|---|---|---|
| FUN_0049cf90 | client, S->C 0x63 (via 0x45e670): opens the panel; tab = act of the player's level (FUN_006427f0 -> FUN_0049c760), 0 without a room; object id to 0x7bf07d; known-waypoint bits (FUN_00661030) | V |
| FUN_0049c490 | tab hit, panel-local: y <= 30, x < 320; expansion 64 px per tab (0..4), classic 80 px | V |
| FUN_0049c510 | row hit: rows with a non-zero flag at 0x7bf040 + 5i (activated) only; x in (17, 297) and y in (top, top + 30), exclusive; table 0x7224e8 hit tops 60, 96, 132, 168, 205, 241, 277, 313, 349 (+80 x, +60 y on screen, S) | V |
| Levels.txt | `Waypoint` 0..38 by act: A1 1 3 4 5 6 27 29 32 35; A2 40 48 42 57 43 44 52 74 46; A3 75..81 83 101; A4 103 106 107; A5 109 111 112 113 115 123 117 118 129 | D |
| FUN_0054d750 | C->S dispatcher: 0x49 goes to its handler (busy test FUN_0057eec0 as for 0x20) | V (run) |
| FUN_0054c5d0 | 0x49 handler: length 9; **FUN_0055b6c0(player) + 0x2710 > GetTickCount (import 0x6cc260) -> voice event 0x13, FUN_00554190 (end interaction), return 1**; FUN_00549570; error -> end interaction; else FUN_00584f60 | V (run) |
| FUN_0055b6c0 | with no second unit: player data (FUN_006221a0) +0x160 | V |
| pcdata +0x160 | written only at 0x55b76b (FUN_0055b720: `= GetTickCount()`); FUN_0055b720's one caller is 0x5a5f51 in FUN_005a5e50, the hostile declaration (both level >= 9, stat 0xc; per-pair timer now + 60 s). Other `+0x160` writes (0x421e77, 0x4b83d5, 0x4b83fb) are client structs | V (scan of every `mov [reg+0x160]`) |
| FUN_00584870 | town portal operate: same field + 0x1388 (5 s) -> voice 0x13, refused (upstream objects.md already calls it hostility) | V, S |
| FUN_00549570 | object (type 2) exists else 1; object's act == player's act else 2; range \|dx\|,\|dy\| <= 10 subtiles (22 when the player's class id is 1, sorceress: telekinesis, I) via FUN_00548ef0; level 0 -> 0; level >= count or no waypoint index -> 3; bit clear -> 2 (FUN_00660e00, FUN_00547f50, FUN_00660e50) | V |
| FUN_00584f60 | object's OperateFn entry == FUN_00584e30 (a waypoint); end interaction; target == current level or 0 -> nothing; bit set; FUN_0053aec0(target, 0xd for a town or levels 46, 74, 133..136, else 0) | V |
| FUN_0053aec0 | same act: FUN_0061b060 (room) + FUN_00554ea0 (warp) at once; other act: FUN_00537340 + FUN_0053acc0 (act change, load) | V (calls), I (timing) |
| after the warp | FUN_005809d0 (mode 2: lights the target waypoint) and FUN_0053b4b0 at the target's x+3, y+3 (0x5850e9); no FUN_00535430 (the portal pair stays) | V (args), S (+3 = tile + 3 subtiles) |

Oracle (`tools/emu/wp_refuse.py`, runs from upstream's tools/emu; FUN_0054d750, FUN_0054c5d0, FUN_0055b6c0, FUN_00554190 and the setter FUN_0055b720 native; hooks in its docstring): 32,964 grid cases (ms since the last level change 0..20,000 x since a hostile declaration none / 0..20,000 x active) plus 100,000 random timelines. game.exe travels 2 s after a level change; refuses with voice 0x13 2 s after a hostile declaration; travels at 10.0 s.

## 2. d2d (V-d2d, upstream/main 03f42254)

| Item | d2d | Verdict |
|---|---|---|
| Refusal | components/game/world/commands.cpp:267 `now - actor->arrived_at < 10000`; arrived_at set on every arrival (world/setup.cpp:363, world.hpp:209 "pcdata+0x160") | **DIFFERS**: 47,157 cases d2d refuses, game.exe travels (within 10 s of a level change); 12,152 game.exe refuses (hostile; d2d has none) |
| Active, same level, waypoint object | commands.cpp:267-268 | SAME |
| Panel | apps/d2d/panels.cpp:930-950 (tabs 64/80 px, y <= 30; rows kWpHitTop = game.exe's table, x 17..297, 30 px), town/walk.cpp:450 tab = view_act() | SAME |
| Arrival | commands.cpp:278-282: preset tile + 3 subtiles (+0.6 cell), lights a dark one; `arrive` at once | SAME (cross-act load I) |
| Range / act checks | interaction walks within 2 cells first (tick.cpp:281); no separate act test | not a bot risk |

PR candidate: local branch `fix/waypoint-hostile-window` in the upstream read worktree (see pr-wp-proof.md, pr-wp-evidence-pack.md). Not pushed, not posted.

## 3. The bot (Q5)

`Navigator.TakeWaypoint` (internal/nav/waypoint.go:89): walk to the nearest waypoint unit, open the panel (3 tries), `waitCooldown` (WpCooldownSec 10 s + 1 s from the first look at the level), click the act tab, then the row text or icon, `waitArea` (3 s, then LoadWait), two rounds. Active waypoints come from the quest/waypoint flags (internal/bot/waypoints.go:50; questmem `WaypointBit`). The failure lines (`tools/emu/wp_soaklines.py`, read-only stream over the four soaks' bot logs, 2026-10-31):

| Cause | navsoak | navsoak2 | navsoak3 | m2early | Total |
|---|---|---|---|---|---|
| `waypoint unavailable: unit not found` (in town, area 1, to 3/4/5/6) | 71 | 36 | 129 | 53 | 289 |
| same, area already the target (status line lag, I) | 2 | 0 | 1 | 0 | 3 |
| `no path: ... unreachable` (walk to the preset) | 0 | 2 | 1 | 0 | 3 |
| row not found / not activated / click did not travel / refusal window | 0 | 0 | 0 | 0 | 0 |

(73 + 38 + 131 + 53 = 295; the P23 count of 220 was taken before navsoak3 finished.) The same `unit not found` also ends 216 TP fallbacks ("waypoint to town failed"), 31 activations, 32 task steps and 21 actions; 369 more TP fallbacks are `none on level` (a level with no waypoint: correct). "wp: used" lines: 0 in all soaks.

**What drill D-wp should expect.** game.exe and d2d: the panel opens on the current act's tab; a click inside an activated row's hit box travels at once; the character stands by the target waypoint (+3 subtiles). d2d today only: within 10 s of any arrival the click closes the panel and nothing happens (no voice). After the d2d patch: travel at once.

## 4. Findings for the primitives fix queue (bot side)

| id | Finding | Where | Fix |
|---|---|---|---|
| W1 | d2d's `npcs` row "Waypoint" becomes an object with Class -1; `nearestWaypointUnit` keeps only `IsWaypointObject(class)`: the waypoint is never found on d2d (292 nav + 216 TP-fallback + 84 activation/step lines; one failure may log at two layers, I) | internal/d2dsrc/items.go:343, internal/nav/waypoint.go:204, :261 | map the name to a waypoint class in d2dsrc (or accept Extras.Names "Waypoint" in nav) |
| W2 | nav row pitch 34 px from y 135: rows 7 (373, the exclusive edge) and 8 (407, the gap) miss, row 6 is 2 px inside. The "rows 7-8 react only on the icon" note is this pitch error. d2dctl's 138 + 35r hits all nine | internal/nav/waypoint.go:48-51 | row centres 135, 171, 207, 243, 280, 316, 352, 388, 424, x 97..377 |
| W3 | `WpCooldownSec` / `waitCooldown` model d2d's level-change window; game.exe has none in single player. Right on d2d today, up to 11 s lost per use after the d2d fix | internal/nav/waypoint.go:58-84, internal/env/tuning_nav.go:34 | keep while d2d refuses; drop (knob 0) when the patch lands; the knob comment says "game.exe", it is d2d |
| W4 | the TP fallback's own 10 s and "panel did not open" never showed up: W1 hides every later step; rerun the soak after W1 before judging rows | - | re-measure |

Not covered: cross-act load time (I), the sorceress' 22-subtile range (I), the hostile path beyond the setter.
