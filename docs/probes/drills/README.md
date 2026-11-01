# Live primitive drills (card m1-liveproof, harness cmd/d2drill + internal/drill)

> status: DRILLS5 (2026-10-31, last section): live proofs of the merged fixes (fix-bank, fix-obj, fix-price, fix-belt2/pickup, fix-nav, fix-restart, fix-missile) and the drills the fork cheats unblocked. Earlier: D0 to D9 (except as listed BLOCKED), D2.d and the hover audit H ran live (2026-10-31, DRILLS3; see the DRILLS3 section). DRILLS4 (2026-10-31, last section): D10 fault recovery, DW.a3 and acts 3 to 5, the repair price table. Earlier: D0, D1 (rerun after the primitive fixes) and D2 (portals, partly) run live (2026-10-31). Results: `2026-10-31-d1.txt` (first run, before the fixes), `2026-10-31-d1rerun.txt` (after fix-belt and fix-pickup), `2026-10-31-d2.txt`; each `.json` next to it. Findings of the D2 run: section W below.

Size rule: a result log committed here must stay small (CI fails files over 1 MB; `tools/ci/check.sh --hygiene-only` fails probe files over 512 KB, run it before pushing). Cap a pasted `-v` log first: `go run ./cmd/d2drill -cap FILE` (keeps head, tail and verdict/error lines, adds one marker line). See docs/fixbigfiles.md.

Run on the worker host, one hidden slot, lab build, throwaway Sorceress, `--no-save`, lockstep:
`D2D_AGENT=drills tools/agentrun.sh drills --ttl 3600 -- tools/d2dslot.sh run -t 3600 -- d2drill -group D1 [-only D1.P1] [-repeat 2] [-v] [-label NAME] [-probe 'verb; verb']`.
The binary is the lab build WITH the click modifier (`D2D_BIN` or `-bin`); `-label` names the result files; `-probe` runs verbs (debug verbs through the cheat ledger) and prints the replies, to find out what a verb does.
Groups: `D1` belt, `D2` portals and town entry, `D0` gate self-test (leaves the pack open on purpose, expects a P35 FAIL line and a recovery).
A result line is `DRILL <id> <primitive> PASS|FAIL|BLOCKED <reason> | intended | observed | runs`. Every drill runs twice, compared by its
observed end state (`Sig` when the text carries regenerating numbers). The P35 gate (cursor empty, no panel, no dialog) runs before and
after every drill; a dirty UI is its own `P35.pre|post.<drill>` FAIL line, one Escape/ground click recovers, a stuck UI stops the group.

Setup is only through `debug` verbs, each written to the cheat ledger first (kind `d2d-debug`, actor `drills:<drill>`, reason, verb, reply;
ledger on the worker host, `$D2CHEAT_LEDGER`). Used: `debug give <code>`, `debug clearinv`, `debug equip <belt> 8`, `debug stat 6|8 <points*256>`
(life, mana). Game actions (not cheats): `cmd pickup`, `cmd belt <col>` (drain the belt). Bot paths called: `town.Town.FillBelt`,
`loot.Looter.PickUp`, combat `Fighter.Act(DrinkLife|DrinkMana)`; the wiring copies cmd/d2bot d2dwire.go (keep in step when it changes).

## Missing cheats and verbs (lab6 build, read-only check of the fork 2026-10-31)
| Need | State | How to add (fork change, needs the user's go) |
|---|---|---|
| click modifier (shift/ctrl-click) | EXISTS in the lab binary as of 2026-10-31: `click|rclick <x> <y> [shift\|ctrl\|alt]`; shift (pack potion to the belt box a pickup would use) and ctrl (sell with a store open) verified live on a hidden slot; alt is parsed only (no game rule), shift with a full belt (no box, no move) is by code reading only | none (D1.P1s can be unblocked in a later change) |
| waypoint activation (`debug wp`) | EXISTS: `debug wp <i>`, `all`, `level <id>` (V live: mask 2^39-1) | none |
| item into pack or belt directly | EXISTS in the lab binary as of 2026-10-31: `debug give <code> pack` (first free cell by the player's search) and `debug give <code> belt [box]` (the pickup's box, or the box asked for); `debug give <code> [q]` still drops on the ground (V live) | none |
| gold | EXISTS: `debug stat 14 N` (purse), `15` (stash) (V live) | none |
| level, life, mana | EXISTS: `debug stat 12\|6\|8` (6 and 8 take 24.8 fixed point, read one tick later) (V live) | none |
| durability | EXISTS in the lab binary as of 2026-10-31: `debug durability <slot\|all> <value>` (capped at the item's maximum; V live); `debug wear` still halves | none |
| kill, hire or revive the merc | EXISTS in the lab binary as of 2026-10-31: `debug merc hire <type>\|kill\|revive` (no gold; V live: hire, death to `dead:true`, revive) | none (D1.P3m can be unblocked in a later change) |
| belt removal | EXISTS in the lab binary as of 2026-10-31: `debug belt clear` (empties the boxes) and `debug unequip belt` (takes the worn belt off) (V live) | none |
| act unlock | EXISTS in the lab binary as of 2026-10-31: `debug act <1-5>` (waypoint panel tabs 1..n; V live: `wp_tabs` 1, 2, 5, 1 rects) | none (DW.a3 can be unblocked in a later change) |
| hit the player | EXISTS in the lab binary as of 2026-10-31: `debug hit <dmg>` (the game's own hurt path, then the get-hit mode; V live: life 40 to 35, mode GH) | none (D2.d can be unblocked in a later change) |
| spawn a monster to fight | EXISTS in the lab binary as of 2026-10-31 (second set, docs/forkcheats.md): `debug spawn <MonStats id\|row> [n] [near] [champion\|unique [mods]]` on a field level (V live: zombies walk, are killable, 200 ticks keep the loop answering) | none (D5.c, D5.d, H.monster, the fight half of D5.b and D5.a's hit effect can be unblocked in a later change) |
| hurt the merc | EXISTS (second set): `debug merc hurt <dmg>` (V live: life 40 to 35 to dead) | none (D8.m, D1.P3m can be unblocked) |
| quest record bits | EXISTS (second set): `debug quest list`, `set <q> <flags>`, `bit <q> <b> on\|off`, `join <act>` (V live: record 2 to 0x0002 and Kashya's talk took it to 0x0001) | none (D8.c, D8.q, D8.w can be unblocked; the act hop is still package main) |
| unidentified magic item | EXISTS (second set): `debug give <code> <quality> [unid]` on the ground and `debug give <code> pack\|belt q<n> [unid]` into the pack | none (D3.i and the unidentified sell prices of D3.g0 can be unblocked) |
| pin in a pocket | EXISTS (second set): `debug pin on [r]\|off\|status` (a collision ring round the player) | none (D6.d can be unblocked) |
| list the verbs | EXISTS (second set): `debug help` | none |

BLOCKED drills in the last runs: D1.P3m, D2.d, DW.a3, D5.c, D5.d, H.monster, D8.m, D8.c, D8.q, D8.w, D3.i and D6.d were blocked on the cheats now in the lab binary (see the table; docs/forkcheats.md; rerun pending), D2.f (round trip) still needs a working portal walk-in.

## D1 rerun (after fix-belt F1/F3/F6/F7/F8, fix-pickup B1/B2/B4 and the arrival fix)
`2026-10-31-d1rerun.txt`: P7a PASS, P7 PASS (pickit reports picked 2, no "left the ground without arriving"), P1 PASS (hp columns hold hp, mp columns mp: the F8 FAIL of the first run is gone), P1s PASS (new), P2 PASS, P3 PASS, P3m BLOCKED. Each twice, identical. The first run's file stays as the "before".

## D2 portals and town entry (`D2`, drills of docs/primitives.md P10, P11)
The drills call a verbatim copy of the bot's `d2dPortalWith`, `d2dTown.ToTown` and `townAdapter.BackToField` (`cmd/d2drill/portal_copy.go`; package main cannot be imported; `TestPortalCopyInStep` fails when the originals change). A spy in front of the input counts the clicks and keys the bot sends. Setup cheats (all ledgered): `debug clearinv`, `debug give tbk|tsc`, `debug goto 2`. Items: scrolls are picked up first and alone (a scroll picked up while a tome is in the pack goes into the tome, V), the tome gets 1 charge from `debug give` (plus 1 per absorbed scroll). An empty tome is made by raw casts (harness, not the bot).
| drill | verdict | observed |
|---|---|---|
| D2.a | PASS | verb from the field: 1 right click, 32 virtual ticks from the call to its return (the right click at +22: panel open, 700 ms wait, item pixel, 150 ms hover), 1 portal object, 1 source spent, `debug portals` shows the pair |
| D2.a2 | PASS | raw cast: portal object 8 ticks after the right click, 3 of 3 casts (trace research: about 7) |
| D2.b | PASS | empty tome, no scroll: error is `ErrNoPortal`, 0 clicks, 0 keys, 0 ticks; log line `d2d town portal: no portal source (empty)` |
| D2.c | PASS | empty tome + loose scroll: scroll spent, tome untouched, 1 right click |
| D2.d | BLOCKED | cut cast: no verb hits the player (missing: `debug hit <dmg>`) |
| D2.e | PASS | cast mode (mode 10) at the call: click at tick +37 vs +22 idle: the verb waited for idle, 1 click, portal opened |
| D2.h | PASS | one-pair rule: second cast 12 cells away closes the first (1 object, 60 subtiles from the old place), 3 scrolls to 1 |
| D2.t | FAIL | see W1 |
| D2.g | FAIL | see W1 |
| D2.f | BLOCKED | round trip needs a walk-in that works (W1) |
Observed failure-line texts: `combat: no town portal scroll or tome` (error), `d2d town portal: no portal source (empty) try=1` (log), `town: area -1 not reached within 15s` (D2.t). The cut-cast and dropped-click texts (`no object after the click ...`) did not occur in these runs.

## W Findings of the D2 run
- W1 (FAIL, real, not fixed here): no portal walk-in works in the lockstep lab. (a) `cmd/d2bot` wires the town package with `d2dtown.HoverFirstBelt` (cmd/d2bot/d2dwire.go:62), not `d2dtown.Input`, so `Interact` on the portal object reaches `d2dinput.Controller.Interact` -> `UsePortal`, which uses the uncalibrated screen transform (internal/d2dinput/controller.go UsePortal, marked U); `townAdapter.ToTown` waits 15 s, `town.ToTown` then re-casts (it spends every remaining scroll, 3 of 3 here) and finally walks (cmd/d2bot/adapters.go:109, internal/town/trip.go:57 and :164). (b) even the d2dctl `EnterPortal` verb (internal/d2dctl/town_verbs.go:633, the one verified by a click on the sprite) did not change the level here: the click moved the player instead (probe: the player walked away from the portal). Not known whether (b) is a lockstep or hover artefact of this lab or a real defect; the earlier proof was not run under lockstep. Suspected fix card: wire `d2dtown.Input` (or a `cmd use <portal>` verb, docs/d2d-town.md) and prove it outside the harness first; also stop `town.ToTown` from re-casting while a portal stands next to the player.
- W2 (harness, fixed): `d2dctl.Pause` waits wall time, which does nothing under lockstep; the lab now pauses in virtual ticks. Without it EnterPortal, HoverClick and the NPC walks act on a frozen world.
- W3 (note): the verb's success test is any portal object within 25 subtiles, so a stale portal from an earlier cast makes it return success without a cast; the drills walk off stale portals first (`clearSpot`). In the field the bot casts again only after the old pair closed or moved, so this is only a risk when the player stands near its own old portal.
- Timings: verb 32 ticks (click +22), raw cast 8 ticks to object (trace ~7), busy player +15 ticks later click, empty tome 0 ticks.

## DW waypoints (card fix-wp, `-group DW`, result `2026-10-31-dw.txt`)
Shared nav code, no bot copy needed: the drills call `nav.Navigator.TakeWaypoint` (internal/nav) on the lab's navigator. Cheats (ledgered): `debug goto 1|40`, `debug wp all`. Verdicts and findings: docs/fixwp.md.

## DRILLS3 (2026-10-31): D3 to D9, D2.d and the hover audit H
Groups and files (`2026-10-31-<label>.txt/.json`, each drill twice, identical unless stated): `D3` town (label d3: starting gold, NPC windows P14, heal P17, shop and sell P15, scrolls P12, repair P16, stash P5, identify P13), `D4` pack (d4: pickit P9, pack full and MakeRoom P4, gold pile and cap P8, cube P6), `D5` combat (d5: BindSkills and cast P26, survival thresholds P28; P27 and P29 BLOCKED), `D6` movement (d6: 10 walks in the camp and in Blood Moor P22, 5 exits P24; P25 BLOCKED), `D7` build (d7: P30 P31 P32), `D8` NPC (d8: revive P18; claim, Akara reward, Warriv, gamble, merc drink BLOCKED), `D9` lifecycle (d9: death and corpse P34, 10 leave and enter cycles P33; P36 is a soak), `D2` follow-up (d2d: `D2.d` and `D2.d3`, the cut cast with `debug hit`), `H` hover audit (h, written up in hover-audit.md). Run: `d2drill -group D3` etc. on the worker host as above, with the lab binary that has the test cheats (docs/forkcheats.md).
Harness code: `internal/drill/dtown.go`, `d4.go`, `dgame.go`, `h.go` (drills and the lab interfaces), `cmd/d2drill/town_lab.go`, `pack_lab.go`, `game_lab.go`, `hover_lab.go` (the live side: it calls `town.Town` chores, `OpenForTest`, `SellForTest`, `DepositForTest`, `MakeRoom`, `Transmute`, `loot.Looter.PickUp`, `combat.Fighter.Safety` and `Respawn`, `leveler.Leveler.Do` and `Equip`, `nav.Navigator.GoTo` and `ToArea`, `d2dmerc.Reviver`, `d2denv.Cycler`). Flags added: `-repair-below` (override the repairBelowPct knob; the default now reaches the repair path with `debug durability all 1`).
Harness notes (not bot defects): the starting state of a repeat is not the state of the first (belt potions, items the pickit refuses stay on the ground, there is no verb that clears the ground), so the drills compare differences and keep their determinism signature to end-state facts; `town.Town.OpenForTest` and `PickEntryForTest` do not reset the vendor tab cache that `openShop` resets (town/npc.go:534), so D3 buys through the shop chore and uses the hooks for selling only; `debug clearinv` reaches the reader one tick later (D4.b run 2 saw the old pack).
New missing pieces (exact): a monster to fight (`debug spawn`: D5.c, D5.d, H.monster), `debug merc hurt <dmg>` (D8.m, D1.P3m), a cheat that sets quest records (Blood Raven to 4, Den done with reward pending, Andariel: D8.c, D8.q, D8.w; `debug quest` sets bit 0 only), an unidentified magic item (D3.i), a way to pin the character in a pocket (D6.d). All five now exist in the lab binary (second cheat set, docs/forkcheats.md; the drills are unblocked, rerun pending). Resolved since the table above: `debug durability`, `debug hit`, `debug merc`, `debug give pack|belt` (used by D2.d, D3.r, D8.a).
Findings for the fix queue (details with file and line in hover-audit.md): F1 gold deposit impossible on d2d (D3.s, D3.s2: `hoverFirst` hides `TypeChar`, the prompt needs digits even for the whole purse); F2 field objects cannot be operated (H.obj: `Interact` unsupported for objects, the click fallback is a walk, d2dsrc lists no field objects); F3 `d2dctl.Pause` is a wall sleep, so the merc claim and revive paths cannot work under lockstep (H.pause); F4 `needRepair` ignores gold (D3.r0); F5 the purse caps at 10000 at clvl 1 and the pickit does not know it (D4.cap).

## DRILLS4 (2026-10-31): D10 fault recovery, DW cross-act rows, repair prices
Result files `2026-10-31-d10`, `-dw3`, `-repair`, `-repair-poor` (.txt and .json). Run as above on one hidden slot with the lab binary that has the test cheats; new groups `D10`, `DR` and the new `DW` drills (`d2drill -group D10 -d10-kill-after 4m -d10-observe 8m`, `-group DR -repair-below 100`, `-group DW`).

### D10 fault recovery (primitive P36's gate-ladder rung: Gate 1 fault injection, docs/review-strategy.md, soak-gates "Injected fault recovery on d2d")
`internal/drill/d10.go` (verdict) and `cmd/d2drill/d10_lab.go` (processes). Three stages on one throwaway character in the slot's d2d home: A (a harness d2d creates and saves it, P35 gate before), B (the bot started the way the soak starts it, `d2bot --backend d2d --managed ... --goals campaign --lockstep`, SIGKILL of the slot's d2d only at 4 wall minutes, observed 8 minutes, then SIGTERM of the bot and the process list), C (the saved character loaded in a harness d2d, P35 gate after, saved experience read from the save file). The d2d pid is found by its `--devctl <the slot's socket>` argument; the drill records that the pid belongs to the process group registered to its own agent in `tools/agentreap.sh --list`; `--list` and `--orphans` are read after the bot stopped. Pass condition: the minute log shows the bot back in game within 3 min, the minute line's `restarts` rises by exactly 1 (read 10 s after the first minute line back in game, with the restart events of the log attributed: a counted rise caused by another new game does not count), saved XP not lower, the loop resumes within 5 min (a kill, XP or area change), no process of the kill left, P35 clean before and after.
Result (FAIL on one criterion, the recovery itself works; two runs, identical verdict; the first run is the evidence): the managed bot noticed the dead d2d at once (log `d2d process exited ... signal: killed`, `game lost, restarting it`), started a new d2d and was back in game 1.4 to 1.7 s after the kill (new d2d pid seen +2 s, first minute line +6 s and +7 s), no d2bot process exit (the soak supervisor's restart count stays 0), the loop resumed (kills 218 to 223 within the first minute line and 223 to 233 after; run 2 240 to 251), no leftovers (the killed pid gone from the process table, `--list` and `--orphans` name nothing of it, no d2d left on the slot after the graceful stop), P35 clean before (fresh character) and after (the saved character loaded). Saved XP: the save file held 3429 before the kill (both runs; the last save point) and 19989 (run 1) and 18194 (run 2) after the graceful stop, so saved XP did not fall; the in-memory XP since the last save was lost, as expected for a kill -9 (xpLost 1566 in run 1, 1974 in run 2: about 1.5 minutes of play). FAIL: the minute line's `restarts` counter did not rise (1 before, 1 at +13 s and +16 s; delta 0, wanted exactly 1). The rises seen later (to 3) belong to the planned save point that follows every re-entry and to later fresh games, not to the kill. A pilot run before the attribution code existed saw +1 in one of two runs, caused by an unrelated fresh game 3 s after the recovery (that is why the verdict attributes restart events).
Findings (not fixed here, exact places):
- F-D10-1: the `restarts` counter does not count the failsafe game-lost recovery on d2d: internal/bot/loop.go:487 (`gameLost` ends in `noteReentry`, not `noteRecycle`), internal/bot/observe.go:241 (`restartCount` takes unplanned `newGameRun` games only, comment observe.go:219). The review-strategy and soak-gates text ("restarts rises by exactly 1") is not satisfiable on d2d as the counter is wired; either count the recovery (one `noteRecycle` in `gameLost`) or change the criterion to the log event `game lost, restarting it` (the drill already reads it). Fix card needed.
- F-D10-2: `xpRisk` on the minute line is always the whole XP on d2d (xpRisk equals xp in every line, also right after a save): `runD2D` builds `bot.Deps` without `SavedExp` (cmd/d2bot/d2dbackend.go:254; only the Game.exe path wires it, live.go:139) and `savedExp` is only taken on a menu-observed game entry (internal/bot/observe.go:386), which d2d never shows. The soak's "xp at risk" alert reads nonsense on d2d; the drill reads the save file itself.
- F-D10-3 (note): after every d2d re-entry the bot runs a save point at once (a second Save and Exit 1.5 s later), and a fresh game follows some save points (a third leave); each costs about 2 s and a game start. Not a defect of the recovery, but it makes any counter-based check noisy.
Run evidence: bot logs stay on the worker host (they hold private paths); the result lines carry the pids, offsets and counters.

### DW cross-act waypoint rows (DW.a3, DW.a3b, DW.a6 to DW.a11)
`debug act N` (ledgered) before each trip that crosses an act. All PASS, each twice, identical: DW.a3 Rogue Encampment to Lut Gholein (tab 2 row 0), DW.a3b Lut Gholein to Cold Plains (act 1 row 1, from an act 2 area), DW.a6/a7/a8 Rogue Encampment to Kurast Docks, the Pandemonium Fortress and Harrogath (tabs 3, 4, 5 row 0), DW.a9/a10/a11 back to the Rogue Encampment from acts 3, 4, 5. Each trip is 2 clicks (tab, row) and 193 virtual ticks (the 10 s refusal wait), arrival 0 to 1 subtiles from the target waypoint, except target 1: d2d puts the character at the town start (141,103), 66 subtiles from the Rogue Encampment waypoint (84,69), identical from every act; whether Game.exe arrives beside the waypoint there is UNVERIFIED, so the drill records the distance and passes on the area id for target 1.
Tab geometry measured with `dump ui` while the panel is open (`wp_tabs` rects at the tab click, in the result lines): act N lists N tabs, each `[80+64(i-1), 60, 64, 31]`, so the tab centres 112, 176, 240, 304, 368 of `internal/nav/waypoint.go` are exactly the rect centres for acts 1 to 5 (act 4 at 304, act 5 at 368), and the click on each of them travelled.
Still missing for acts 3 to 5: `knownWaypointPos` (internal/nav/act2hints.go) has only Lut Gholein. After `debug goto 75|103|109` on a level the character has not visited yet, the preset has not streamed in and the verb fails with `none on level 75` (V live, run alone: DW.a9 failed that way before DW.a6 had visited the level); once the level was visited (the full group order) the waypoint object is there. Level-relative waypoint positions seen on arrival (d2d, origin not subtracted): Kurast Docks (159,49), Pandemonium Fortress (44,19), Harrogath (114,69); for comparison Lut Gholein (69,84), Rogue Encampment (84,69). The absolute table entries need each level's `Origin` added (as Lut Gholein 5069,5084 = 69,84 + 5000).

### Repair (DR group, P16): the price table and the poor purse
`D3.rt` damages one worn item of each base alone (`debug unequip`, `debug equip`, `debug durability all 9999`, `debug durability <slot> N`, all ledgered), runs the bot's repair chore at the Act 1 smith (`-repair-below 100` so half-worn rows are repaired too) and records the gold paid (purse before minus after, equal to the chore's own report) for two losses per base. Normal-quality bases, Normal difficulty, clvl 1 purse 10000; both runs identical. `debug equip` sometimes makes an ethereal item (maximum durability half plus one: cap 7, hax 15, V live in an earlier run), which the smith does not repair (paid 0, durability unchanged): the drill equips again until the maximum is the normal one.
| base | code | max dur | lost -> gold paid | gold per point lost |
|---|---|---|---|---|
| Cap | cap | 12 | 11 -> 7, 6 -> 4 | 0.64, 0.67 |
| Leather Armor | lea | 24 | 23 -> 57, 12 -> 30 | 2.48, 2.50 |
| Quilted Armor | qui | 20 | 19 -> 16, 10 -> 8 | 0.84, 0.80 |
| Splint Mail | spl | 30 | 29 -> 1871, 15 -> 968 | 64.5, 64.5 |
| Full Plate Mail | ful | 70 | 69 -> 5814, 35 -> 2949 | 84.3, 84.3 |
| Hand Axe | hax | 28 | 27 -> 20, 14 -> 10 | 0.74, 0.71 |
| Short Sword | ssd | 24 | 23 -> 8, 12 -> 4 | 0.35, 0.33 |
| Long Sword | lsd | 44 | 43 -> 185, 22 -> 95 | 4.30, 4.32 |
| Staff | sst | 20 | 19 -> 19, 10 -> 10 | 1.00, 1.00 |
| Buckler | buc | 12 | 11 -> 7, 6 -> 4 | 0.64, 0.67 |
| Kite Shield | kit | 30 | 29 -> 257, 15 -> 133 | 8.86, 8.87 |
| Leather Boots | lbt | 12 | 11 -> 9, 6 -> 5 | 0.82, 0.83 |
| Leather Gloves | lgl | 12 | 11 -> 9, 6 -> 5 | 0.82, 0.83 |
Reading: the price is proportional to the durability lost (per-point cost constant per base, rounded to whole gold), strongly base dependent (Full Plate 5814 for a full repair, a Short Sword 8), so the bot's repair needs the price before the visit and a purse check (finding F4 stays). The fresh Sorceress's staff at 1 of 20 cost 19 (D3.r). For the decomposition item on price formulas: the table is the measurement to fit (item base price and the repair factor are not in the table; Charsi's price modifier at clvl 1 Normal is included).
Poor character (D3.r0 gold 0, D3.r1 gold 5, pre-fix-bank main): the chore still opens the smith once and fails ("repair not confirmed (durability 5% unchanged)", 101 and 168 ticks), gold and durability unchanged, UI clean, then the latch (internal/town/basicsfix.go) stops the retries until durability changes or gold grows. fix-bank was not on origin/main at the time of the run, so this is the pre-fix behaviour; D3.r1 (`-only D3.r1`) is the drill to rerun after it merges (expected: `repair skipped` in the first call). Repeats use a different durability each (the latch is per process).


## DRILLS5 (2026-10-31): live proofs of the merged fixes, the drills the fork cheats unblocked
Result files `2026-10-31-drills5-<label>.txt/.json` (each drill twice, P35 gate around each; `2026-10-31-drills5.txt` is every DRILL line of the final pass in one file): `d3` (D3.s, s2, tp, n, b, b2, i, u, h), `repair` (D3.r, r0, r1), `d1` (D1.P1, P1b, P7b), `d8c` (D8.c, one run: the Rogue can only be claimed once per character), `d8` (D8.a, m, q, w), `ffight` (D5.dodge, p27, p28, p29, H.obj.chest, .shrine, .cage), `fnav` (DW.m1, m2, a12), `dw-a9`, `dw-a10`, `dw-a11` (each alone in a fresh process), `d6pin` (D6.d), `h` (H.pause, H.obj), `d10` (two runs). Run as in DRILLS4 on one hidden slot with the newest cheats build; new group `F` (`-group F`, `internal/drill/dfield.go`, `cmd/d2drill/field_lab.go`), D8, D3, D1, D6 extended (optional lab interfaces `beltGiver`, `unidGiver`, `QuestLab`, `PinLab`, `repairEstimator`, `sellExpecter`, so the offline fakes are unchanged). The d2drill process sets the lockstep Pause itself (`newLab`); the drills that prove the bot's own wiring (H.pause, D8.c, D8.w) install the bot's (`LateClock.Install`, game-time Pause and Now) around the call.

### Verdicts (headline numbers)
| Group | Drills | Verdict | Evidence |
|---|---|---|---|
| A1 fix-bank | D3.s, D3.s2, D3.r0, D3.r1, D3.tp | PASS x5 | full deposit 1500 then withdraw 300: purse 300 stash 1200; partial 1000 of 1500: purse 500 stash 1000; gold 0 and 5: repair skipped in 0 ticks, no smith opened; D3.tp unchanged (800 gold) |
| A2 fix-obj, waits-audit | H.obj, H.obj.chest, H.obj.shrine, D8.c, D8.a, D3.n, D3.b | PASS | the bot's Interact on field objects: chest (class 139) and shrine (class 81) NU to ON, the town waypoint opens; ClaimRogue under the game-time Pause grants the Rogue; revive PASS; one shop run on lockstep PASS |
| A2 | H.obj.cage | PARTIAL PASS | Cain's gibbet is listed with class 26 (`quest.ObjCainGibbet` is right) and Interact is accepted; its mode stays NU (the cage needs the Cain quest state: effect UNVERIFIED) |
| A2 | H.pause | FAIL (finding F1) | the game-time deadline works (failures at 396 to 434 ticks, 0 wall); the walk to Kashya from the camp arrival cell (25.5,22.5) stalls and needs 396 to 950 ticks (16 to 38 s) |
| A3 fix-price | D3.r, D3.b2, D3.u | PASS | repair estimate 19 = 19 gold paid (Staff 19 of 20 lost); identify scroll sold for 40 = formula 40; an unidentified magic Cap sold for 32 = the base formula (cost/2) |
| A4 fix-belt2, fix-pickup | D1.P1b, D1.P7b, D1.P1 | PASS | sash, 6 healing + 4 mana in the pack, empty belt: belt 4 healing + 4 mana in the planned columns, 2 healing stay in the pack, 0 failed moves, no `belt: marked full` line; healing in columns 0 to 2 (both rows) and one on the ground: column 3 stays free, the pickit leaves it |
| A5 fix-nav | DW.m1, DW.m2, DW.a12, DW.a9/a10/a11 | PASS x6 | standing in the Cave Level 1 mouth trigger of Cold Plains: 1 bounce (the next step-away took the other direction) and 0 bounces, town in 497 and 554 ticks; a waypoint touched in passing is handed over (DrainActivated [3]) and works as a trip target; towns 75, 103, 109 never visited, each run alone: no `none on level N` |
| A6 fix-restart | D10 | PASS | kill -9 of the slot's d2d: restarts 0 to 1 and 1 to 2 (delta exactly 1 in both runs), back in game within 8 s, xpRisk after the kill 108 = the unsaved XP only (xp 1863, saved 1755; 588 before), saved XP not lower, no leftovers, P35 clean |
| A7 fix-missile, fix-missile2 | D5.dodge | PASS | Fallen Shamans spawned near a Sorceress (400 life): 8 dodges and 1.00 cast commands per cast over 17 casts (other runs: 4 to 17 dodges, 5 of 5 and 7 of 7 commands per cast); the idle character sees 1 to 2 hostile moving missiles in `debug missiles` and in `Snapshot.Missiles` |
| B1 | D5.p27, D5.p29 | PASS (recorded) | a pack with a champion Fallen at clvl 1: kills and ignores recorded, no avoid (the champion is not named in the log: the danger table calls it danger 1 to 2 at clvl 1); life 9 of 40 with a pack in reach: Fighter.Safety decides `chicken` (not run: it leaves the game); with potions it decides `drink-rejuv` |
| B2 | D5.p28 | PASS | a rich character (5000 gold) in a fight with 20 zombies: 2 mana drinks and 12 kills (the cast effect on a target, P26); the poor control (gold 0) drinks none: the policy rests and uses melee (finding F3); Meter totals stay BLOCKED (see below) |
| B2 | D8.m | PASS | `debug merc hurt 20` then the belt verb with the merc flag: merc life 20 to 32 with one belt potion |
| B2 | D3.i, D3.u | PASS | `give rin pack q4 unid`: the identify chore identifies it with a scroll (scroll -1); the unidentified magic Cap sells for 32 = the base price |
| B2 | D8.q, D8.w | PARTIAL PASS | `debug quest set 1 0x0002` then Akara's menu: +1 skill point, record 0x0001; `set 6 0x0001` then Warriv's menu row `go east`: area 40. The bot's own talk nodes are package main (cmd/d2bot task.go:285), so these prove the cheat and the game rule, not the bot's code path |
| B2 | D6.d | FAIL (finding F2) | `debug pin on`: the stall ladder climbs rungs 0 to 5, a `hop` is logged as the recovery at rung 6, the character moved 0.4 cells in 66 s, GoTo ends in `nav: stuck ... after 6 recoveries`; the Escape hook (town portal) is not on this path |

### Findings (not fixed here, exact places)
- F1 (H.pause): `d2dctl.ApproachNPC` (internal/d2dctl/town_verbs.go:239) from the arrival cell (25.5,22.5) of the camp takes 396 to 950 ticks to reach Kashya (16 to 38 s): the character stalls at (25.5,22.5) and the sidestep needs several rounds. The callers give it 15 s (internal/d2dtown/adapter.go:333) and 20 s (internal/d2dtown/claim.go:56), so with the game-time deadline (waits-audit) the approach reports `not reached in 15s`. Before waits-audit the wall deadline never fired under lockstep, which hid the long approach; on a real run the same stall costs the same game time. Cause not isolated (stall at 25.5,22.5; D8.a and D8.c passed because they start next to Kashya). Fix card: longer or distance-based timeouts and a look at the camp stall.
- F2 (D6.d): a plain `Navigator.GoTo` that is pinned returns `ErrStuck` after `MaxRecoveries` rungs (internal/nav/walk.go:233) in 66 s of game time; the Escape hook (town portal) runs only from `exitUnstick` (internal/nav/stuckmem.go:167), so a pocket reached during a plain walk is left to the caller. The primitive row P25 expects "waypoint or TP". Without the warm-up walk the map knows the pin and GoTo answers `no path` at once (0 s).
- F3 (P28): `resource.PotionPolicy` (internal/resource/potion.go:100) lets a poor character (gold below PoorGold) with a melee fallback rest and back off instead of drinking mana while the pack danger is below 2; it drank in every fight of a character with gold and in crowd fights with natural uniques. As designed; a soak of broke characters will show many `out of mana: backing off` lines.
- F4 (note, UNVERIFIED in the bot): the first `Fighter.Clear` right after a `debug goto`, before the navigator has walked on the level, calls every spawned monster unreachable (internal/combat/fight.go:646, `no path and no firing position`) and returns in about 4 s; the drills take a short warm-up walk first. A real arrival has the navigator's map from the first step.
- F5 (note): the Fighter's `casts` counter (fight telemetry) can exceed the `cmd skill` commands sent (0.4 to 1.0 commands per counted cast) because the input layer's cast gap drops the repeats; resends (more than 1.0) never occurred in any run.
- F6 (D8.a, unchanged): revive is free at merc level 1 (price 0).

### Still BLOCKED (exact missing piece)
- P28 Meter totals: `d2dcombat.Attach` leaves `Rig.Meter` nil (internal/d2dcombat/rig.go, only `New` builds it): a drill that wants damage and mana totals needs `New` or an `Attach` that builds the Meter; a code change in an area the drill group does not own. Fixed offline by fixmeter ([fixmeter.md](../../fixmeter.md)): `Attach` builds the Meter (off until `Meter.Enable()`), live recipe there.
- D8.q and D8.w as the bot's own code path: the talk nodes live in package main (cmd/d2bot task.go, acts.go) and cannot be imported; they need a testable extraction.
- D8.g (P19 gamble): the bot has no gamble code. D9.c (P36): a soak, not a drill.
- H.obj.cage effect: the Cain quest state (portal opened by the Tree of Inifuss scroll chain) is not set by `debug quest set` alone in a way the cage honours (U).
- D1.P3m: not rerun (D8.m covers the merc drink by the same belt verb).

### Harness notes
- The lab fights spawn the monsters with `debug spawn <id> <n> near` and clear the level first with `debug kill` where a crowd would trigger the chicken (the kill's experience levels the character: the drills set the life they need by cheat and go back to town with `debug goto 1` after a fight). A drill that fights twice with the same character sees a higher clvl in the second repeat; the signatures use the facts that do not move (dodge count above 0, a mana drink seen, an action decided).
- `Fight` takes a short walk first (map), then counts `cmd skill` verbs from the client's verb trace, the casts and dodges from the Fighter's `combat fight area=` telemetry lines and `combat dodging` lines, the potions from `combat potion kind=` lines.
- Mouth drills stand the character beside the exit tile with `debug warp` (the exit comes from `debug warps`) after a real waypoint trip in, so the navigator knows the level's waypoint.

## DRILLS6 (2026-10-31): live proofs of fixstall, fixmisc, fixrun, fix-portal2
Result files `2026-10-31-drills6-<label>.txt/.json` (each drill twice per invocation, P35 gate around each; lockstep is deterministic, so a second invocation gives the same numbers): `h`, `hb` (H.pause), `d6pin`, `d6pinb` (D6.d), `sync`, `syncb` (D-sync), `run` (D-run, last invocation; three invocations agreed), `d2` (D2 subset after the full group, see below), `d2a2`, `d2d`; `2026-10-31-drills6-d2-fullgroup.txt` is the first full D2 group run (its json was overwritten by the subset); `2026-10-31-drills6.txt` has every DRILL line. Run through `labrun` on the worker host: `D2D_AGENT=drills6 tools/agentrun.sh drills6 --ttl 900 -- tools/d2dslot.sh run -n 0 -t 900 -- d2drill -group F -only D-run -label drills6-run -v` (lab build cheats3). New drills (cmd/d2drill and internal/drill only): `D-sync` and `D-run` in group `F`; D6.d now also accepts ErrLevelChanged after the escape.

| Card | Drill | Verdict | Numbers |
|---|---|---|---|
| fixstall | H.pause | PARTIAL | bot Pause: try 1 FAIL at 718 ticks (`not reached in 28.4s`, budget 15 s + distance), try 2 opens in 136 ticks; lab Pause 147 and 1229 ticks. DRILLS5 was 396 to 434 ticks failing, 0 of 2 opened this way in one run, 1 of 2 in the other |
| fixstall | D6.d | PARTIAL | log line `nav: stalled in a pocket, escaping by town portal` seen in both runs; run 1 GoTo ends `ErrLevelChanged` (escaped from level 2 to 1), 493 ticks (20 s); run 2 FAIL: portal opened but unreachable from the pin, 1554 ticks (62 s), `stuck ... after 6 recoveries`, character left in Blood Moor (put back by `debug goto 1`) |
| fixmisc | D-sync | PASS x3 | first Clear on a cold arrival (no warm-up walk): kills 3 of 3, cast commands 9, 0 `no path and no firing position` lines, 9 s. Town portal and waypoint arrivals: not run (the lab only has the cheat goto as arrival) |
| fixrun | D-run | PASS x3 | see below |
| fix-portal2 | D2.* | PASS alone | see below |
| P26 | D-skill | NOT RUN | recipe not available |

### D-run numbers (fixrun)
Leg of about 40 subtiles in Blood Moor from the same start and direction, one tick at a time, positions from `debug level` (cells x5): fresh game 0.5561 subtiles per tick (path 49.5 in 89 ticks, a bend), same game again 0.5628, after a restart 0.5628 and 0.5628 (ratio after/before 1.01: the reset works), stamina 19200 to 15680 and 18944 to 16264 (drain = running). Baseline fresh game with the bot's run policy off (no run key): 0.3753 per tick, stamina unchanged, ratio 0.67 = walking. Controller counters: `runBelieved=true`, `runKeys` 1 after the first game and 2 after the restart, `runResets` 1, 2, 3 (restart, restart, baseline restart). `walkingRaw` is `0` while standing and `walking=1` in a state read while moving, in BOTH running and walking: it means "moving", not "walking instead of running" (V live). ANSWER: the bot runs after a fresh game and after a restart (one key press each time, belief reset works); the toggle is not why walking is 68 percent of each kill in the soak; look at kills (casting stands still), stalls and the stamina policy instead (U). The `Step`-count of this harness is the game tick of the lockstep: 0.5628 and 0.375 match the expected 0.5625 and 0.375 per tick.

### D2 (fix-portal2)
All of D2.a (32 ticks), D2.t, D2.a2 (portal at 8 ticks), D2.b, D2.c, D2.g, D2.d (2 clicks, 1 scroll left), D2.d3 (2 clicks, 0 scrolls left, the cast cut with the source spent, the retry opens the portal), D2.e (click at +37 vs +22), D2.h, D2.f PASS in subsets, each twice identical. The full group in one process FAILS in an order-dependent way: D2.a2 (`setup: a stale portal stays within 25 subtiles after walking away`, left by D2.t) then D2.d3 DIFF (3 clicks, no object), after which the character is dead and D2.e, h, f fail their stage setup (`tome -1, scrolls 0`). The `debug hit` drills should run last or alone (harness fix, not a fix-portal2 defect).

### Findings (not fixed here)
- F7 (H.pause): the first bot approach after the lab's own approach stalls: the failing try ends at (25.9,20.5), 718 ticks, with a budget of 28.4 s. The log reports the position of the previous poll (`prev`), not the final one. The stall is next to Kashya's wander (UNVERIFIED); the sidestep (3.5 cells, up to 3 candidates) is not enough there. Not the old 396 to 950 stall at 25.5,22.5, but the same class.
- F8 (D6.d): the pocket escape runs but `town.enterPortal` walks to the portal with the pinned character (`GoTo` to the portal cell, ladder again, `open portal did not lead to town`), so the escape works only when the portal is within the ring: it opened in run 1 and not in run 2.
- F9: `debug revive` does not exist in the lab build as written (usage reply); a dead character is put back by the Fighter's Esc respawn.

## DRILLS7 (2026-10-31): live proof of fixstall2, D-sync arrivals, the full D2 group in one process
Result files `2026-10-31-drills7-<label>.txt` (stdout of each d2drill invocation: DRILL lines, table, the H.pause diagnostics; `2026-10-31-drills7.txt` has every DRILL line; no json: the labrun sync does not bring the result directory back, the stdout is the record): `h1`..`h3` (H.pause), `d6d1`..`d6d3` (D6.d), `dsync-tp`, `dsync-wp`, `d2full1`, `d2full2`, `p26`, `camp-grid` (the Rogue Encampment collision grid around the stall, from `debug grid 5`). Build `d2drill` on the worker host and run: `D2D_AGENT=drills7 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh drills7 --ttl 900 -- tools/d2dslot.sh run -n 0 -t 900 -- <built>/d2drill -group H -only H.pause -label drills7-h1 -v` (H.pause is group `H`, D6.d is group `D6`, D-sync, D-sync.tp and D-sync.wp are group `F`, D2 is group `D2`). Lab build cheats3 (fork copy cheats3, the same binary as DRILLS5 and DRILLS6), code of origin/main at ea07758 (fixstall2 merged).

| Card | Drill | Verdict | Numbers |
|---|---|---|---|
| fixstall2 | H.pause x3 | FAIL (finding F10) | 6 bot tries: 3 open Kashya (136 ticks each), 3 fail at 718 ticks (`not reached in 28.4s`, player at 25.9,20.5); the three invocations are identical to DRILLS6 (lockstep is deterministic), so the map route changed nothing; lab Pause 147 and 1229 ticks |
| fixstall2 | D6.d x3 | FAIL (finding F11) | all three invocations: run 1 ErrLevelChanged (escaped from level 2 to 1, 493 ticks = 20 s, in town), run 2 stuck 1554 ticks (62 s), 0.4 cells, `open portal did not lead to town`, `nav: stuck ... after 6 recoveries`; identical to DRILLS6 |
| fixmisc | D-sync.tp | PASS | first Clear right after a town portal arrival (portal verb, ToTown, the bot's BackToField), no warm-up walk: 3 of 3 zombies, 13 cast commands, 0 unreachable lines, 16 s, both repeats identical |
| fixmisc | D-sync.wp | FAIL 1 of 2 (finding F12) | first Clear right after the navigator's waypoint trip to Cold Plains: run 1 kills 3 of 3 (11 casts, 15 s); run 2 kills 0 (3 casts, 19 s, `ignoring target why=approaches 12 approaches, 7 to 8 subtiles away` for all three), 0 `no path and no firing position` lines in both |
| drill harness | D2 full group, one process | PASS x2 | 11 of 11 drills PASS and identical in both repeats, in two invocations (D2.a 32 ticks, a2 [8 8 8], d 2 clicks 1 scroll, d3 2 clicks 0 scrolls, e +37 vs +22, h one object 59 subtiles away, f arrival 4 subtiles from the origin) |
| P26 | probe | PARTIAL record | see below |

### Harness changes (cmd/d2drill and internal/drill only)
- `lab.ensureAlive` (called by `Goto` and `ClearPack`, so at the start of every drill that stages items or arrives): a dead character takes the Fighter's Esc respawn (the game's own, `debug revive` does not exist), a hurt one gets its life topped up by the logged `debug stat 6` cheat. Per-case setup no longer depends on the previous case.
- `clearSpot` (stale portal reset between cases): each of the 4 directions is walked up to 4 times while a portal object is within 25 subtiles, instead of one 8-cell step per direction. In the DRILLS7 full-group runs no stale portal setup failure and no dead character occurred (D2.d3 was not lethal here; the DRILLS6 death came from the earlier failure chain).
- D2.h's signature carried the `debug portals` text, whose coordinates move by one subtile between repeats (59 vs 60): the signature now holds the row count and whether the portals are 20 or more subtiles apart (the DIFF of the first full-group run was this and nothing else).
- New drills `D-sync.tp` and `D-sync.wp` (`FightSpec.Arrive`, `lab.arrive`), and H.pause prints, when a try fails, the position timeline, the verb counts, the NPC list and `debug grid 5` to stderr.

### Findings (not fixed here)
- F10 (H.pause, fixstall2 does not work for this stall): the failing try sends 9 `cmd interact`, 5 `cmd move` and 2 `debug grid` calls (so `routeAround` ran twice) and still ends at (25.9,20.5): the player is jammed against the east face of a large blocked block of the camp (`2026-10-31-drills7-camp-grid.txt`), in a gap that is 1 to 2 subtiles wide. `planAround` (internal/d2dctl/approach_path.go) needs clearance 1 around EVERY cell of the route, including the first step, so from the stall cells (25.9,20.5), (25.5,22.5) and (26.4,20.5) it returns NO route (checked offline against the real grid, scratch test, not committed); the same BFS with clearance 0 returns a 4 to 5 waypoint route from all three, and from the arrival cell (27.7,20.6) the clearance-1 route exists (it is not used there: the first interact walks until it is stuck). The fix is a clearance fallback (try clearance 1, then 0) or a start-cell relaxation. Whether the walker can follow the clearance-0 route is UNVERIFIED.
- F11 (D6.d, fixstall2 F8 fix is dead code in the wiring): the log of run 2 shows `GoTo < town.enterPortal < ToTown < walkEscape`: `enterPortal` still pre-walks. `town.WalkingInteractor` is implemented only by `d2dtown.townInput` (built by `d2dtown.Input`), but both the bot (cmd/d2bot/d2dwire.go:62) and the drill harness build the town on `d2dtown.HoverFirstBelt(be.Ctl, be.Client)`, a `hoverBelt` that does not implement `InteractWalks` and whose Interact is the controller's one-shot. So the walking entry never runs. Fix: give `hoverBelt` an `InteractWalks` that matches the verbs it uses, or build the bot's town input on `d2dtown.Input`; one line in internal/d2dtown/adapter.go plus a wiring test. The harness copy of the wiring (cmd/d2drill/live.go:78) has to follow the bot's.
- F12 (D-sync.wp, run 2): not the DRILLS6 map-less symptom (no unreachable lines): the Fighter walks toward zombies it cannot get closer to (12 approaches at 7 to 8 subtiles) and ignores them. The first repeat on the same level kills them. Cause not isolated (zombies spawned in a ring round the waypoint arrival cell with an obstacle between, or the repeat's character level, U); needs the geometry (`debug grid 5` at the spawn) the next time.

### P26 observables (record only, lab verbs of the fork, recipe in docs/research/skill-bind-trace.md)
`debug skill` and `debug skillpts` exist (`debug help`). Fresh clvl 1 Sorceress in town: `debug skill right 54` (Teleport, 0 points) replies `err d2d: not usable` and `debug skill right 9999` too, `debug skill right 36` (Fire Bolt) is accepted: the lab verb follows the picker rule (V live, step 3 of the recipe). Step 1, `cmd select 54 0` then `dump`: the dump read at tick 0 without a step shows rskill 0 (the select is queued, not yet applied: D1 stays UNVERIFIED, would need a step between). Steps 2, 4 and 5 (cast mana costs in the field and in town) not run.

## DRILLS8 (2026-10-31): live proof of fixstall3, the regression pass of the whole suite
Result files `2026-10-31-drills8-<label>.txt` (stdout of each d2drill invocation; no json, the labrun sync does not bring it back): `h1`..`h3` (H.pause), `d6d1`..`d6d3` (D6.d), `dsyncwp1`..`dsyncwp3` (D-sync.wp before the harness fix), `dsyncwp-diag` (with the new `clear-diag` geometry dump), `dsyncwpmana1`..`3` (after it), `d3`, `d8`, `d1`, `d2`, `d4`, `d5`, `d6`, `d7`, `d9`, `dw`, `dr`, `dr60`, `h`, `f`, `d10` (one invocation per group, in that order), and the single reruns `d8c`, `d1p1s`, `hobjchest`, `dsync1`. Run as in DRILLS7 (build `d2drill` on the worker host, lab build cheats3, code of origin/main at 1628eeb with fixstall3): `D2D_AGENT=drills8 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh drills8 --ttl 900 -- tools/d2dslot.sh run -n 0 -t 900 -- <built>/d2drill -group H -only H.pause -label drills8-h1 -v`. The harness has no `-group all`: one invocation per group (D0 is the gate self-test and leaves the pack open on purpose: not run).

| Card | Drill | Verdict | Numbers |
|---|---|---|---|
| fixstall3 | H.pause x3 | PARTIAL (finding F13) | per invocation run 1: 1 of 2 bot tries opens Kashya (144 ticks), the other fails at 796 ticks (`not reached in 28.4s`); 3 of 6 first-run tries fail again (DRILLS7: 3 of 6 at 718); run 2 of every invocation: 2 of 2 bot tries open (144). The stall is now left (see F13), the budget runs out 3 cells short |
| fixstall3 | D6.d x3 | PASS x3 | `ErrLevelChanged` (escaped from level 2 to 1 after a stall in a pocket), 225 ticks = 9 s, character in town, both repeats identical, all three invocations identical (DRILLS7 run 2: stuck 62 s) |
| fixstall3 | D-sync.wp x3 | FAIL 1 of 2, cause isolated: the drill (F12 closed) | run 1 kills 3 of 3 (11 casts, 15 s); run 2 kills 0, 3 ignores, in all three invocations. `clear-diag` of the failing try: mana 1/35 (run 1 spent it, the lab does not refill mana), melee `Attack` fallback, open floor around the player (no obstacle in `debug grid 5`). With `Mana: 35` in the drill spec: 3 invocations x 2 repeats, 3 of 3 killed, 11 casts, 15 s |
| fixstall3 | D3 and D8 smoke (NPCs through the bot's `d2dtown.Input`) | PASS | D3: g0, n (5 NPCs open and close clean), h, b, b2, tp, r (19 gold), r0, s, s2, u, i all PASS; D8: a, m, q, w PASS; D8.c FAIL in the group (the Rogue was claimed by D8.a first: harness order) and run alone: run 1 granted, run 2 FAIL (the reward can be claimed once per character, so a repeat 2 cannot pass) |
| regression | the suite | see the table below | |

### Regression table (one invocation per group)
| Group | Verdicts |
|---|---|
| D1 | P7a P7 P1 P1b P7b P2 P3 PASS; P1s FAIL in the group (setup: pack hp1x4, a leftover of the previous drill), PASS alone x2 identical; P3m BLOCKED (unchanged) |
| D2 | 11 of 11 PASS identical (the full group, as DRILLS7) |
| D3 | 12 of 12 PASS |
| D4 | a b c g cap x PASS (6 of 6) |
| D5 | a, b PASS; c, d BLOCKED with the stale text "no debug verb spawns a monster" (the fork cheat exists, the F-group fight drills cover it: PASS below) |
| D6 | a, b, c PASS; d PASS (16 s in the group, 9 s alone) |
| D7 | a, b, c PASS |
| D8 | a, m, q, w PASS; c see above; g BLOCKED (no gamble code) |
| D9 | a, b PASS; c BLOCKED (a soak) |
| DW | 14 of 14 PASS (a1..a11, b, c, a3b) |
| DR | r, r0, r1 PASS; rt FAIL at the default repair knob (the half-worn rows are not repaired at knob 30 by design: 13 of 26), PASS with `-repair-below 60` (26 rows measured, restored, paid) |
| H | pause FAIL (F13), obj, ground PASS, monster BLOCKED (stale text, as D5.c) |
| F | H.obj.shrine, H.obj.cage (partial), D-run, D-sync.tp, D-sync.wp (mana fixed), DW.m1, m2, a12, D5.dodge, p27, p29, p28 PASS; D-sync 2 of 3 (one ignore, repeats identical) in the group and 3 of 3 alone; H.obj.chest FAIL in the group (class 144 chosen, mode stays NU), PASS alone (class 139) |
| D10 | run 1 PASS (d2d killed and restarted by the bot, counters and saves as required, the bot built from this tree with the d2dtown.Input wiring); run 2 cut by the 1200 s TTL I gave (the group needs about 2 x 12 min), no table line |

### Harness changes (cmd/d2drill and internal/drill)
- `D-sync.tp` and `D-sync.wp` and `D-sync` now start with 35 mana (`Mana: 35` in `fSyncArrive`): the second repeat started with the mana the first spent (1 of 35) and fought with the melee fallback. Closes F12 as a drill artefact.
- `lab.clearDiag` (cmd/d2drill/hover_lab.go): after a cold first Clear that killed nothing, prints the player (life, mana), the monsters and `debug grid 5` to stderr (`clear-diag:` lines).

### Findings (not fixed here)
- F13 (H.pause, fixstall3 clearance fallback works but not fast enough): the failing try now leaves the stall: the timeline shows (25.9,20.5) at t855, then a walk along row 18.6 at t933 to t1013 ((26.1,18.8) to (20.1,18.4)) with 3 cells to Kashya (17.1,18.3) when the 28.4 s budget ends (796 ticks). So `routeAround` with clearance 0 reaches the NPC, but only after about 700 ticks of earlier stalls and sidesteps (7 `cmd move`, 2 `debug grid`: two routes were planned, the first did not free the player; why is UNVERIFIED). The error text still reports the position of the previous poll (25.9,20.5), not the final one (20.1,18.4). Possible fixes for a card: plan the clearance-0 route at the FIRST stall (not stall 1 after a full round of interacts), or extend the budget while the player makes progress.
- F14 (note, drill order): in a whole-group process D1.P1s, D8.c and H.obj.chest (group F) differ from their alone runs (leftover pack potions, the Rogue already claimed, another chest preset chosen); D8.c cannot pass at repeat 2 even alone. The groups are not independent; for verdicts read the alone runs.
- F15 (note, UNVERIFIED): with 1 mana the Sorceress's melee `Attack` fallback did not kill three zombies in 60 s on open floor (12 approaches without closing, three ignores, DRILLS8 `dsyncwp-diag`). The product question is whether a mana-less Sorceress should be able to melee a weak pack.
- F16 (harness text): D5.c, D5.d and H.monster still say BLOCKED "no debug verb spawns a monster"; the spawn verb exists and the F group covers them. The texts should be removed (not done here, only the BLOCKED lines are stale).

## DRILLS9 (2026-11-01): live proof of fixstall4, the D10 fault drill twice, group-run isolation, regression pass
Result files `2026-11-01-drills9-<label>.txt` (stdout of each d2drill invocation, no json): `h1`..`h3` (H.pause), `d10a`, `d10b` (D10, `-repeat 1` each, one background job per repeat, TTL 1800 s), one file per group (`d1`, `f`, `h`, `d3`, `d8`, `d2`, `d4`, `d5`, `d6`, `d7`, `d9`, `dw`, `dr`), the group reruns after harness fixes (`f-r2`..`f-r5`, `d1-r2`, `d1-r3`, `d8-r2`, `dr60` = DR with `-repair-below 60`) and the DW.m1 narrowing runs `dwm1-a`..`dwm1-i`. Run as in DRILLS8 (build on the worker host, lab build cheats3, code of origin/main at 912086a with fixstall4): `D2D_AGENT=drills9 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh drills9 --ttl 900 -- tools/d2dslot.sh run -n 0 -t 900 -- <built>/d2drill -group H -only H.pause -label drills9-h1 -v`; D10 additionally `-repeat 1 -bot <built>/d2bot` with `--ttl 1800 -t 1800` (a repeat takes about 13 minutes: 4 min to the kill, 8 min observed).

| Card | Drill | Verdict | Numbers |
|---|---|---|---|
| fixstall4 | H.pause x3 | PASS | all three invocations: the bot's own Pause opens Kashya in 2 of 2 tries of run 1 AND 2 of 2 of run 2 (248 ticks each; lab stepping Pause 147 and 238), so 6 of 6 first-run and 6 of 6 second-run tries (DRILLS8: 3 of 6 at 144 ticks with 3 failing at 796); no failure, so no `H.pause-diag` timeline; character in town afterwards |
| m1 fault | D10 run 1 | PASS | kill -9 of the slot's d2d under the managed bot: restarts 1 to 2 (delta 1), new d2d pid +2 s, first minute line +4 s, loop resumed (kills 148 to 154), saved XP not lower, P35 clean before and after, no leftover process, bot exits 0 |
| m1 fault | D10 run 2 | PASS | the same checks: delta 1 at +15 s, first minute line +6 s, kills 201 to 207, no leftovers, P35 clean. Both runs completed inside the 1800 s TTL (DRILLS8's second run was cut at 1200 s) |
| harness | group isolation | FIXED (3 of 4 causes), 1 open | see below |

### Harness changes (cmd/d2drill and internal/drill only; no product code touched)
- D1.P1s (F14, setup said pack hp1x4): `stagePack` now takes any loose item at the character's feet first (`PickupRaw`), then ClearPack empties the pack, so potions an earlier drill left on the ground are not counted into the staged pack. D1 group: 9 of 9 rows as before, P1s PASS in all three group runs (`d1`, `d1-r2`, `d1-r3`).
- H.obj.chest (F14, class 144 chosen in the group, mode stays NU, class 139 alone): the first chest the preset list names depends on which rooms earlier drills left streamed in, and some chest classes do not change mode. `OperateObject` now tries up to 4 objects of the kind (skipping the ones whose mode did not change, each attempt after a fresh `debug goto`), and the result text names the classes tried. Group F: PASS in `f-r2` to `f-r5` (classes 140 and 144 chosen, 139 alone).
- H.obj.shrine (new in an F subset: NU to NU after D-sync.wp and the chest): `OperateObject` runs `debug kill` (logged, with a reason) before it looks for the object, so a pack a fight drill left in the level does not interrupt the walk to the object. After it the subset `D-sync.wp,H.obj.chest,H.obj.shrine,DW.m1` and the full F group pass shrine.
- D8.c (F14, FAIL because D8.a ran first and the Rogue is claimable once per character): D8.c now runs FIRST in the group and returns BLOCKED, not FAIL, when the character already owned a merc before the claim (`drill.MercOwnedBefore`). Live in `d8-r2`: run 1 granted (14.3,21.9:RG), run 2 BLOCKED (owned), as the game's rule requires.
- D-sync (F14, 2 of 3): the `Mana: 35` setup of DRILLS8 holds: D-sync, D-sync.tp and D-sync.wp pass in every group run (D-sync kills 2 of 3 or 3 of 3, both pass).

### Regression table (one invocation per group, code of 912086a plus the harness fixes above)
| Group | Verdicts |
|---|---|
| D1 | 8 PASS (P7a P7 P1 P1b P7b P1s P2 P3), P3m BLOCKED (unchanged); three runs identical |
| D2 | 11 of 11 PASS |
| D3 | 12 of 12 PASS |
| D4 | 6 of 6 PASS |
| D5 | a, b PASS; c, d BLOCKED (stale text, F16) |
| D6 | a, b, c, d PASS |
| D7 | a, b, c PASS |
| D8 | a, m, q, w PASS; c granted in run 1 and BLOCKED in run 2 (once per character); g BLOCKED (no gamble code); before the harness fix D8.c FAIL in the group |
| D9 | a, b PASS; c BLOCKED (a soak) |
| DW | 14 of 14 PASS |
| DR | r, r0, r1 PASS; rt FAIL at the default knob (13 of 26 rows are half worn and not repaired at knob 30, by design), PASS with `-repair-below 60` (26 rows) |
| H | pause, obj, ground PASS; monster BLOCKED (stale text) |
| F | D-run, D-sync, D-sync.tp, D-sync.wp, H.obj.chest, H.obj.shrine, DW.m2, DW.a12, D5.dodge, p27, p29, p28, H.obj.cage PASS in `f-r2` to `f-r5`; DW.m1 BLOCKED DIFF (run 1 only), see F17 |
| D10 | 2 of 2 PASS |

### Findings (not fixed here)
- F17 (DW.m1, group F, run 1 only, UNVERIFIED cause): in the full F group (4 of 4 runs, `f`, `f-r2`..`f-r5`) the first repeat of DW.m1 ends BLOCKED `the waypoint trip in failed: nav: timeout: walk did not finish in 2m0s` (the navigator's walk from the camp arrival to the waypoint stops after its first stuck recovery at (139,103), clicks at the next waypoints, then no progress) and the second repeat passes (494 ticks, 1 bounce, 2 step-aways). Narrowing: DW.m1 alone and after any of D-run, D-sync + D-sync.tp, or H.obj.chest + H.obj.shrine, or D-sync.wp alone: PASS (`dwm1-a`..`dwm1-g`, `dwm1-f`); it fails after `D-sync.wp,H.obj.chest,H.obj.shrine` (`dwm1-i`, also `dwm1-h` before the `debug kill` fix, where the shrine failed too). So the trigger needs the waypoint arrival of D-sync.wp followed by the Blood Moor object drills; the navigator is rebuilt by `Goto` (wp_lab.go), so it is not a stale navigator; a leftover in `be.World` or in the camp (a monster, a dropped item) is the next suspect. Product or harness is open: a walk that stalls for 2 game minutes in the camp is the F13 family and may be a product stall.
- F16 stays open (D5.c, D5.d, H.monster stale BLOCKED text).

## DRILLS10 (2026-11-01): live proof of fixf17, fixtalk, fixmeter; P26 observables; regression pass
Result files `2026-11-01-drills10-<label>.txt` (stdout of each d2drill invocation, capped with `d2drill -cap`, scrubbed): `f1`..`f3` (full F group, the fixf17 proof), `f4` (F group again at the end), `d8` (D8, fixtalk), `m` and `m2` (group M, D-meter; `m` is the first run before the harness fix below, kept as the record), `p26a`..`p26c` (probe sequences), `d2`, `d6`, `dw`, `h` (regression). Run as in DRILLS9 (build on the worker host through `labrun`, lab build cheats3, code of origin/main at a7f5bc6): `D2D_AGENT=drills10 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh drills10 --ttl 900 -- tools/d2dslot.sh run -n 0 -t 900 -- <built>/d2drill -group M -label drills10-m2 -v`. Every invocation starts a fresh throwaway (`NoSave`), so D8.c is first-run in each. A whole F group took under 1 minute of wall time in lockstep.

| Card | Drill | Verdict | Numbers |
|---|---|---|---|
| fixf17 | full F group x3 (`f1`..`f3`) | PASS | 14 of 14 drills PASS in each of the three invocations; DW.m1 PASS in both repeats of each (483 ticks, 1 bounce, 2 step-aways, identical); `the Pacer lagged` lines: 0 in all three (the log line is written only above 1 s of lag, so the lag never exceeded 1 s). F17 does not reproduce. The fourth F group at the end (`f4`) is the same 14 of 14, 0 lag lines |
| fixtalk | D8.q | PASS | the bot's own `talk.Adapter.Talk(Akara, reward)` after `debug quest set 1 0x0002`: skill points 0 to 1, quest 1 record now 0x0001, identical in both repeats |
| fixtalk | D8.w | PASS | `debug quest set 6 0x0001`, `Talk(Warriv, go east)`: menu `[bot path Go East]`, area after 40, identical in both repeats |
| fixmeter | D-meter (new, group M) | PASS | 20 zombies spawned at Blood Moor, life 1000, `Fighter.Clear`: Meter kills 13, units seen dead by the independent sampler 13, Fighter.Clear kills 13; dealt 121 = 121 (per-unit first alive hp minus last alive hp, plus the last alive hp of the dead; the drops alone are 67), within the spawn hp sum 222 (26 units listed: 20 spawned plus 6 others that stayed untouched); taken 106 = 106 (life drops of `state`); mana 68 spent, deaths 0, min life 894, 1214 samples; hits and misses print `n/a`; identical in both repeats (`m2`) |
| P26 | probe | RECORDED | see below |
| regression | D2, D6, DW, H, F | see table | |

### D-meter notes
- Harness fix (not product): the first run (`m`) FAILED in repeat 2 because Meter totals are cumulative across `Enable()` calls (Enable forgets the last sample and the seen set, not the totals): repeat 2 read kills 17 = 13 + 4, dealt 198 = 121 + 77, taken 246 = 106 + 140, each equal to the sum of the two windows. The drill now subtracts the totals at the start of its window. Product note for a later change: `Meter.Reset` and `Enable` do not zero `Totals`, so a reader needs deltas.
- The independent sampler (`cmd/d2drill/meter_lab.go`) reads the raw `monsters` and `state` verbs per unit id at the same sleeps (it wraps the clock under the Meter), so the totals are computed twice by different code on the same game ticks. It does not catch a sampling gap (damage between samples), both see the same samples.
- Repeat 2 of the first run (`m`) killed only 4 of 20 within 150 s while the sampler agreed with the Meter; in `m2` both repeats were identical at 13 kills. Combat variance between invocations is not a meter matter and was not chased.

### P26 observables (record only; `p26a`..`p26c`, lab verbs, recipe of docs/research/skill-bind-trace.md; fresh clvl 1 Sorceress)
- (1) `cmd select 54 0` then 3 steps: state `rskill=0` (the select is not applied for a 0-point skill; the recipe expected rskill 54 on d2d), UNVERIFIED which side is right (d2d refuses by the picker rule or the select is not visible in `state`).
- (2) `debug points 36 1` and `debug points 42 1`: accepted, `skillpts=0` before and after, `BindSkills` not run in the probe.
- (3) `debug skill right 54`: `err d2d: not usable`; `debug skill right 36`: accepted, `rskill=36` after 2 steps (as DRILLS7).
- (4) in Blood Moor, one zombie (hp 11): `cmd skill 36 <x> <y> <unit>` mana 35 to 32 (whole points as `state` shows them), the zombie hp 11 to 8 in mode GH (the hit effect). `cmd skill 42` (Static Field, no points): mana unchanged 32 to 32.
- (5) in town: `cmd skill 42` and `cmd skill 36` leave the mana unchanged (33 to 33): no cast, as expected.
- Not recorded: the 8.8 mana value and the not-in-town cue text (the probe prints `state` only).

### Regression table (code of a7f5bc6 plus the D-meter harness)
| Group | Verdicts |
|---|---|
| D2 | 11 of 11 PASS |
| D6 | a, b, c, d PASS |
| DW | 14 of 14 PASS |
| H | pause, obj, ground PASS; monster BLOCKED (stale text, F16) |
| F | 14 of 14 PASS in four invocations |
| D8 | a, m, q, w PASS; c granted in run 1 and BLOCKED in run 2 (once per character); g BLOCKED (no gamble code) |

### Findings
- F17 closed: not seen in 4 of 4 full F group invocations after fixf17 (DRILLS9: run 1 failed in 4 of 4).
- F16 stays open (H.monster stale BLOCKED text).

## DRILLS11 (2026-11-01): live proof of fixemptyplan and fixportalmelee; the cut-cast data; regression pass
Result files `2026-11-01-drills11-<label>.txt` (stdout of each d2drill invocation, capped with `d2drill -cap`, scrubbed): `emptyplan-new1`..`3` (code of origin/main at 1a47b45 with both fixes), `emptyplan-old1`..`3` (control: the parent of the fixemptyplan merge, ea5e2ae^1 = 8ac245e, plus the same harness), `portalmelee1`..`3`, `portalctl1`..`3`, `castcut`, regression `d2`, `d5`, `d6`, `dw`, `f`. New group `P` (cmd/d2drill and internal/drill only): D-emptyplan, D-portalmelee, D-portalmelee.ctl, D-castcut. Run as in DRILLS10 (build on the worker host through `labrun`, lab build cheats3): `D2D_AGENT=drills11 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh drills11 --ttl 900 -- tools/d2dslot.sh run -n 0 -t 900 -- <built>/d2drill -group P -only D-emptyplan -label drills11-emptyplan-new1`. Without `-v` the log is a few lines (with `-v` one invocation is 270,000 lines). The control binary is built from a second worktree at the pre-fix commit with `internal/drill` and `cmd/d2drill` copied over it (the harness does not use the new `input.ErrAlreadyThere`).

| Card | Drill | Verdict | Numbers |
|---|---|---|---|
| fixemptyplan | D-emptyplan x3 | PASS | per invocation (2 repeats, 2 levels each): 58 out-of-mana back-offs, `empty path` lines 0, in 3 of 3 invocations (identical). The Fighter fell through to melee 29 times (back-off cap). 10 failed walks, all "the level changed during the walk: escaped from level 9 to 1 after a stall in a pocket" (nav's pocket escape in Cave Level 1, repeat 2): not the planner's empty plan |
| fixemptyplan | control, the pre-fix binary x3 | FAIL as intended (the drill can fail) | per invocation: 47 back-offs, 6 failed walks, all 6 `no path: empty path` (5 of 8 at the wall-backed cell (29,50) of Blood Moor, 1 of 14 in Cave Level 1): 13 percent of the back-offs, in 3 of 3 invocations. Soak 2 had 28 percent. The same geometry with the fix: 0 |
| fixportalmelee | D-portalmelee x3 | FAIL (PRODUCT, see below) | 3 zombies in melee reach (within 8 subtiles), 2 within 3 subtiles at the read, life 102 of 300 (34 percent), belt empty, one scroll: Safety is `town-portal`, not `chicken` (the scroll stays only because the harness does not run the action). The rule's adjacency (3 subtiles) is never reached by 3 monsters at once |
| fixportalmelee | D-portalmelee.ctl x3 | PASS | ONE zombie adjacent, same life and pack: Safety `town-portal`, the bot's Act casts: scroll 1 to 0, portal object listed |
| fixportalmelee | D-castcut | RECORDED | see below |
| regression | D2, D5, D6, DW, F | see table | |

### Findings (product)
- F18 (fixportalmelee, rule unreachable live): measured in Blood Moor with 6 spawned zombies lured to the character, 142 reads of the monster list: the nearest, second and third nearest monster were at best 1.0, 2.4 and 5.7 subtiles away (centre to centre); never 3 within 3 subtiles. Monsters in an attack mode (A1, A2) stood 1 to 9 subtiles away (histogram in the log, 1 to 3: 37 of 59 samples). `crowdAdjacent = 3` (internal/combat/crowd.go, marked U there) is smaller than the melee reach of the monsters met, so with 3 attackers at 34 percent life and no potion `meleeCastGuard` does not fire and the bot still casts the portal, the case the card meant to stop. Diagnosis only (product code is not touched here): the adjacency radius needs a measured value (the data says at least 8 subtiles for a Zombie; other monsters not measured); 3 monsters within 8 subtiles WAS reached in all three runs. UNVERIFIED: the soak death had 3 to 5 adjacent by the analyst's count, not by this radius.
- D2.d3 stays as it was (PASS, the `debug hit 5` cheat cuts the cast): that is a forced hit, see the next point.

### D-castcut: does a hit cut a town portal cast on d2d (the D2.d unknown)
Setup: one Zombie in melee reach, scroll in the pack, cast by the harness's own right click, one `Step 1` at a time from the click, 14 trials (life 300 and 40, a distinct open cell each so the previous portal never hides the new one; some clicks timed to the start of the zombie's blow animation, so its hit lands inside the cast), identical in both repeats. Units: one Step call is 2 world ticks of 40 ms (a measured 320 ms from the click to the portal object after 4 steps).
- Cast timeline, every trial: the scroll leaves the pack at step 1, the player is in mode 10 (cast) from step 1 to step 6, the portal object is listed at step 4 (8 ticks, 320 ms; D2.a2 saw about 7 to 8 ticks), the player is back at rest at step 7.
- Cut by a monster hit: NO in 14 of 14. Five trials took a zombie hit at step 2 or 3, before the portal object appeared (damage 2 to 3, at life 300 and at life 40, 5 to 8 percent); the cast still completed, the player mode stayed 10 (never 4, hit recovery), the scroll was spent once, one portal. A hit outside a cast (step 18) does put the player in mode 4 for 8 steps, so these hits do stagger a player who is not casting.
- The `debug hit 5` cheat (D2.d3, hit 3 ticks after the click) DOES cut the cast (2 clicks, both scrolls spent, one portal; passes again in `d2` of this run). So a forced hit and a real zombie blow differ on d2d: the cheat is not evidence for the rule.
- What the data decides: for weak hits (2 to 3 damage) a cast under melee is NOT interrupted and takes 320 ms, so for zombies the premise of fixportalmelee (the cast is cut, the scroll lost) is not supported; the soak death (cast 1.3 s by the analyst, hits cut it) does not match 320 ms. Not tested: harder hits (the only way to deal them is a stronger monster; a Zombie hit 2 to 5), other monster kinds, a cast at low life where the same hit is larger. UNVERIFIED either way for a big blow. Until that is measured the rule is cheap insurance at best, and per F18 it does not fire at the live distances anyway.

### Regression table (code of origin/main at 1a47b45 plus the group P harness)
| Group | Verdicts |
|---|---|
| D2 | 11 of 11 PASS (D2.d and D2.d3 included) |
| D5 | a, b PASS; c, d BLOCKED (stale text "no debug verb spawns a monster", superseded by D5.p27 and D5.p29 in F) |
| D6 | a, b, c, d PASS |
| DW | 14 of 14 PASS |
| F | 14 of 14 PASS (DW.m1 PASS) |

### Harness notes
- D-emptyplan places the character with `debug warp` on the nearest cell with a wall 3 subtiles behind and 15 free subtiles in front (grid rows are level-relative subtiles, row = y), spawns 12 plain zombies (a pack over 8 is not a weak pack, so the out-of-mana ladder reaches Retreat; a champion pack is avoided and returns at once), sets mana to 0 before each 6 s `Fighter.Clear` round and life to 1000. Zombies never died in these runs (0 kills at mana 0): the drill measures the back-off, not the kill.
- A first try with 8 plain zombies never backed off (melee fallback applies to a weak pack); a first try with champions never fought (`combat ignoring target ... avoid`, a loop that returned at once: a spin guard was added).
- The runner prints run 1's reason when the repeats differ; D-emptyplan therefore also prints one `D-emptyplan call N` line per repeat.
- Monster positions from `monsters` are level-relative cells, the snapshot's are act subtiles: compare in cells (an early version compared screen coordinates).
- A spawned Zombie idles until the character comes close (most stayed 14 to 20 subtiles away for 1,200 ticks): the drills walk toward the nearest one to wake it.
- Scroll setup picks up only the `tsc` item (PickupRaw picked up potions dropped by earlier kills, so the belt was not empty).

## DRILLS12 (2026-11-01): live proof of fixcrowd (F18)
Result files `2026-11-01-drills12-<label>.txt` (stdout of each d2drill invocation, scrubbed; one DRILL line plus the table): `portalmelee1`..`3` and `portalctl1`..`3` (the proof), `blocked-verbcells`, `fail-verbcells`, `fail-botview` (the harness iterations, kept as the record). Run as in DRILLS11 (build `d2drill` on the worker host, slot 0, lab build cheats3, code of origin/main at 2f34d65 plus the harness changes below): `D2D_AGENT=drills12 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh drills12 --ttl 900 -- tools/d2dslot.sh run -n 0 -t 900 -- <built>/d2drill -group P -only D-portalmelee -label drills12-portalmelee1`; the control with `-only D-portalmelee.ctl`. Capacity: worker host idle+nice 62 percent before the first run. Each invocation takes about 4 s of wall time (lockstep).

| Card | Drill | Verdict | Numbers |
|---|---|---|---|
| fixcrowd | D-portalmelee x3 (after: crowdAdjacent 8) | PASS | 3 zombies within 8 subtiles (the knob) at the read, life 102 of 300, belt empty, one scroll: Safety `chicken`, scroll count 1 to 1. Both repeats of each invocation identical; the bot's snapshot distances at the read were 5, 6.4, 6.4 (then 18.4, 24.4, 34.1) |
| fixcrowd | before (DRILLS11, crowdAdjacent 3) | FAIL x3 | `town-portal` at the same life, same pack, the rule never fired |
| fixportalmelee | D-portalmelee.ctl x3 | PASS | one zombie in reach: Safety `town-portal`, the Act casts, scroll 1 to 0, portal object listed (unchanged) |

### What the harness could observe (honesty point)
- The decision (`Fighter.Safety` on the live snapshot) and the scroll count: observed. The chicken ACTION is not observed: `Chicken` leaves the game, and a drill must not close the game it runs in (AGENTS.md), so `acted` stays false for the chicken case. The proof is that the pure decision on a live snapshot with live monsters is `chicken` instead of `town-portal`, with the scroll still in the pack; "the scroll is not spent on a cut cast" holds in the sense that no cast is started. UNVERIFIED live: the leave-game itself (covered by the offline chicken tests) and any hit cutting a cast (no hard-hit verb on main).

### Harness findings and changes (cmd/d2drill and internal/drill only)
- The wait is now "3 within combat.crowdAdjacent", read from the tuning handed to the bot (`l.tune.Combat.CrowdAdjacent`); the control keeps "1 within 3".
- The first run (`blocked-verbcells`) did not reach the setup: the drill measured with the `monsters` verb (level cells, rounded) and found 1 within 8 at the read. The loop returned as soon as 3 were within 8, but the pack had spread by the read after the life cheat.
- `fail-verbcells`: the pack was re-waited at low life, 3 within 8 by the verb, and the decision was still `town-portal`: not a product bug. `fail-botview` (new bot-view print) shows why: the same monsters in the snapshot's act subtiles (the bot's measure, `combat.Hostiles`) were 4.5, 5.8 and 8.1 away; the verb's cell distance put the third at 7. The 8.1 monster is outside the knob's 8. So the harness now counts reach with the snapshot coordinates (what the guard sees); the verb stays for the attack-mode histogram only. Lesson: a cell-based distance is 1 to 3 subtiles coarser than the bot's; compare in the bot's unit.
- Product note (no change): the margin is thin: the third monster was 6.4 (PASS runs) and 8.1 (the earlier read) away. With the default 8 a pack that stands 8.1 away is not counted; the knob range is 3 to 12.

### Shared-file lines for the merger
- BACKLOG: close fixcrowd item 1.5 (drill wait fixed) and its live proof (task 1.4): verified live (D-portalmelee PASS x3, control PASS x3).
- INDEX / CHANGELOG: "drills12: live proof of fixcrowd: the 3-adjacent portal drill passes at crowdAdjacent 8 (Safety chicken, scroll kept); the control still casts; the drill measures in the bot's own unit".

## DRILLS13 (2026-11-01): live proof of fixrecast (Frozen Armor), M1 mana timing
Result files `2026-11-01-drills13-<label>.txt` (stdout of each d2drill invocation, scrubbed): `fro12` (D-fro1 and D-fro2, the proof), `fro1-stalecache` (the same drill on the lab's old tables cache, the failing control), `m1a`, `m1b` (D-m1, two invocations; identical). The first D-fro2 invocation (BLOCKED, "did not reach level 2": `debug points <skill> <n>` SETS the level, the harness added one point per call) was not kept. New group `Q` (internal/drill/d13.go, cmd/d2drill/d13_lab.go; overlap declared: three lines in cmd/d2drill/main.go for the group switch and the end-of-run goto town). Run on the worker host, one hidden slot (slot 1), `lab-debug-skills` binary, lockstep, throwaway Sorceress `--no-save`: `D2D_AGENT=drills13 D2D_BIN=<fork>/bin/d2d.lab-debug-skills tools/agentrun.sh drills13 --ttl 900 -- tools/d2dslot.sh run -n 1 -t 900 -- <built>/d2drill -group Q -only D-fro1,D-fro2 -repeat 1 -cache <fresh tables dir> -label drills13-fro12`. Capacity: worker host idle+nice 54 percent before the first run. A whole invocation takes under 10 s of wall time (600 virtual seconds). Skill-drill verbs go through `d2dctl.SkillDebug` (ledgered with reason and actor `drills:<drill id>`); `debug points`, `debug hit`, `debug stat`, `debug goto`, `debug kill` through the logged lab cheats.

| Card | Drill | Verdict | Numbers |
|---|---|---|---|
| fixrecast | D-fro1 (level 1) | PASS | no bot: the state (`debug states` buffs=`40:1`) is on 8 ticks (320 ms) after `cmd skill 40`, gone 120.32 s after the command: the real duration is 120.0 s from the on-state (table 3000 frames, 25 fps). Bot: `Fighter.Buff` called every sample for 600 virtual s: casts at 0.0, 114.0, 228.0, 342.0, 456.0, 570.0 s (buffLimit = min(150, 120 - 6) = 114 s; the bot's own clock and game time agree to 0.1 s), armor absent in 0 of 819 samples (samples at most 2.08 s apart, 120 ms near the recast and the real end) |
| fixrecast | D-fro2 (level 2) | PASS | state `40:2`, gone 132.32 s (table 3300 frames); casts at 0.0, 126.0, 252.0, 378.0, 504.0 s; absent in 0 of 775 samples |
| fixrecast | D-fro1 on the stale cache | FAIL | casts at 0, 150, 300, 450 s; armor absent in 567 samples (118 s of 600): `auralencalc=""` in the loaded Skills row (see F1) |
| sk-mana-trace D1/D3 | D-m1 | RECORD (x2, identical) | below |

### What the recast proof answers (the fixrecast "unverified live" list)
- State switches on after a cast: YES, 8 ticks after the command (the action frame, the same tick Fire Bolt's missile is made), read with `debug states` (buffs list). Bit 10 of `Player.States` is NOT what was read: on d2d the player snapshot never fills `States` (`d2dsrc` marks `player.states` unknown, "no state list in dump"; the drill read `Player.States.Has(10)` at every up sample: never true). So on d2d the state-readable path of `buffDue` never runs; the clock path does the work.
- Real duration: 120.0 s at level 1, 132.0 s at level 2 (state on to state off), equal to the table; the timer is not shorter on d2d. Cast cost 7 points (6.9 read, regen in between).
- A recast while the armor is up refreshes the full duration: VERIFIED LIVE on d2d (5 recasts at 114 s left the armor on for the whole 600 s; without a refresh it would have dropped at 120 s).
- Recast time: `buffLimit` = 114.0 s (level 1) and 126.0 s (level 2); observed casts exactly there. PASS = armor never absent in a simulated 600 s: yes, at the sampling above. The harness cannot see a gap shorter than its sample spacing (120 ms to 2 s); the state expires only at 120 s, so a 6 s margin is far above that.

### Findings (product, not fixed here)
- F1 (stale tables cache silently disables fixrecast): `data.Open` re-extracts only when the cache was made from a different install (`CacheCurrent`), not when the column list changed. fixrecast added `auralencalc` to `skillCols`; the lab's old cache (`drills/tables/skills.tsv`, made 2026-10-31, 71 columns, no `auralencalc`) loads `AuraLenCalc=""`, `AuraFrames` is false, and the bot falls back to the flat 150 s timer: Frozen Armor down 118 s of 600. A fresh cache directory (`-cache`) gives the 114 s recast. Any machine with a pre-fixrecast cache (the soak host, the user's) keeps the old behaviour until its cache is rebuilt. Candidate fix: a schema stamp (hash of the column lists) in the cache check. UNVERIFIED: that the production bot's cache is stale (not looked at).
- F2 (Buff is not called on a timer): `Fighter.Buff` is called from `internal/bot/goal.go` sweepPass only when `needBuff` is set (new game, new goal, back from town, after death) and from the boss `BeforeAttack` hooks. Nothing re-arms `needBuff` during a long sweep, so even with the 114 s limit the armor can run out during one sweep longer than 120 s: the drill calls Buff every sample, which is the assumption of the fix, not the bot's measured behaviour. S (read the call sites); not measured live.
- F3 (`Player.States` empty on d2d): see above; `docs/fixrecast.md` and `Fighter.buffDue` describe a state-readable path that only a memory-reading backend has. The lab verb `debug states` is the only read.

### M1: when does d2d pay mana (Fire Bolt level 1, `debug cast 36` east, 40 ms ticks, one tick per sample)
Mana = `dump stats` raw[8] (256ths), full 9472 before the cast; cost 640 (2.5 points); regen is +3 per tick afterwards. Player mode from `players` (10 = cast, 4 = get-hit, -1 = none), missiles = player-owned rows of `debug missiles`.
- Undisturbed: command queued at +0 (mana 9472); +1: mana 8835 (paid, 637 = 640 - 3 regen), mode 10; +2..+7 regen only; +8: a firebolt missile is listed (the action frame); mode 10 through +10. Mana leaves 1 tick after the command and 7 ticks (280 ms) BEFORE the action frame.
- Compare with sk-mana-trace (read, not run): D1 says game.exe pays at the action frame (FUN_0056f7f0) and d2d at the command: confirmed, the gap is 7 ticks. The action frame tick (+8) matches the trace's "event around tick 8-9 (13-tick cast)".
- Interrupted (`debug hit 1`, a CHEAT, applied after k samples; the cast is cut to mode 4 on the next tick): k = 0 (hit before the cast started): mode 4 at once, no cast, mana unchanged (9472), no missile (not a cut cast: the cast never began). k = 1..7: mana already paid at +1 and NOT refunded (final 8862 = 9472 - 640 + 10 ticks of regen), no missile ever. k = 8 and 9 (hit applied at or after the action frame): the missile exists (+8), mana paid. So a hit up to tick +7 cuts the cast and, on d2d, costs the mana: D3 confirmed (game.exe would charge nothing, trace section 2).
- What the harness did not observe: game.exe itself (this is d2d only; the comparison is to the trace's reads); a real monster blow as the interruption (castcut, DRILLS11: weak zombie hits do not cut a cast; only the cheat did); start-paying skills (Telekinesis, Hydra, Thunder Storm, Inferno: not run); M2 (Charged Bolt count), M3 (Teleport floor, needs level 25), M4 (regen), M5 (potions), M6 (skill delay): not run.

### Harness notes
- Raw `Client.Step(n)` is n ticks; `lab.Tick(n)` is n+1 (it ends with a one-tick Pacer sync): earlier notes that "a Step call is 2 ticks" are `Tick`.
- The bot's own clock (the Pacer) and the game clock agreed to 0.1 s over 570 s once `lab.Tick` synced them after every sample.

### Shared-file lines for the merger
- BACKLOG: fixrecast live proof done (verified live, 114 s / 126 s recasts, 120 s / 132 s real duration); NEW items F1 (cache column stamp in data.Open / CacheCurrent), F2 (Buff only on needBuff: re-arm needBuff by the armor's recast time), F3 (Player.States empty on d2d); sk-mana-trace D1 and D3 confirmed live on d2d (pay at +1, action frame +8, a cut cast stays paid): PR candidate "pay at the action frame" now has its live evidence.
- INDEX / CHANGELOG: "drills13: live proof of fixrecast (Frozen Armor recast at 114 s / 126 s, real duration 120 s / 132 s, recast refreshes the state) and the M1 mana trace (d2d pays 1 tick after the command, 7 ticks before the action frame, a cut cast is not refunded); finds a stale tables cache disables the fix and that Buff is only called on needBuff".


## DRILLS14 (2026-11-01): DENBOUNCE, the call that leaves the cave

Question (docs/fixdenbounce.md, branch agent/fixdenbounce): which call walks the character off the cave arrival cell and fires the return trigger. Candidates: `sweepPass` (touchWaypointHere, Fighter.Clear, Explore), nav `mouthStep`'s skip, the d2d mover. Tags: V = read from the live logs below.

New drills in group F (`internal/drill/dfield.go` fDen, `cmd/d2drill/field_lab.go` DenPhase): DW.m3 (Blood Moor 2 to Den of Evil 8) and DW.m4 (Cold Plains 3 to Cave Level 1 9). Each does the bot's own `ToArea` into the cave (the exit walk), then the phases of `Runner.sweepPass` one at a time: idle 30 ticks, `Fighter.Clear` (natural cave monsters, invincible character), `Navigator.Explore` (the bot's `Mover.Explore`), 4 entries per drill. A phase that ends outdoors is the bounce; the drill prints the log of that phase. The only cheat is the setup `debug goto` (ledger). `touchWaypointHere` is excluded by the code: Blood Moor and the caves are not in `Plan.Waypoints` (internal/bot/waypoints.go:115, returns before any walk). One small diagnostic added to nav: `nav mouth check` and `nav mouth exit` debug lines in `mouthStep` (internal/nav/walk.go, no behaviour change, test `TestMouthStepDebugLines`); the `nav walk callers` line already existed (walk.go:69).

Files `2026-11-01-drills14-<label>.txt`: `m3a` (before the mouth lines), `m3b`, `m3c`, `m3d` (DW.m3, identical), `m4a` (DW.m4, cut by hand to 117 KB: head 400 lines plus the DRILL, callers, mouth, step-away, arrival and failure lines of the rest). Run: `D2D_AGENT=drills14 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh drills14 --ttl 900 -- tools/d2dslot.sh run -n 0 -t 600 -- <built>/d2drill -group F -only DW.m3 -label drills14-m3b -v`. About 3 s of wall time each.

### Result 1: Den of Evil (DW.m3, 4 of 4 invocations, 4 of 4 entries each, identical)
V: exit walk lands the character at (40,145) in level 8; `idle` and `Fighter.Clear` keep it there (area 8); the first `Navigator.Explore` walk leaves: log of every cycle is
`roam: coverage points level=8 exits=1` then `nav walk callers=nav.(*Navigator).roam < explore < Explore < DenPhase`, `nav mouth check pos={40 145} level=8 exits=1 wantExit=0`, `nav mouth exit dest=2 exit={35 135} dist=11.2 skip=true`, `nav plan from={40 145} to={34 146} wps=1`, `nav click pos={40 145} wp={36 146}`, then 12 ms later `d2dworld: arrived in level 2` (the single click fires the trigger), error `the level changed during the walk: 8 -> 2`. No `stepping away` line, no `nav exit`: exactly the soak 2 signature.

Call chain (V): `Runner.sweepPass` goal.go:1102 `Mover.Explore` -> `Navigator.Explore` explore.go:18 -> `explore` explore.go:46 -> `roam` roam.go:150 -> `walk` roam.go:207 (first target (34,146), one click toward (36,146)). The step-away that should have caught it is `mouthStep` (walk.go:618), skipped by `if d > mouthR || d < 0.5` (walk.go:631) because the arrival is 11.2 subtiles from the listed exit position and `mouthR = 9` (walk.go:587). `roam`'s own filter `dropNearExits` (roam.go:123) keeps the target (Chebyshev distance 11 from the exit, `roamExitClear = 7`, roam.go:121). So the verdict between the candidates: (b), the `mouthStep` radius, together with roam's target choice; not touchWaypointHere, not the d2d mover (it just obeys one click), not Fighter.Clear in this level (V: clear phases stayed in 8 in 4 of 4).
U: why the trigger is 11 subtiles from the listed exit position: the arrival is at offset (+5,+10) from the exit position in BOTH caves (Den: (40,145) vs (35,135); Cave Level 1: (260,155) vs (255,145)), so the exit position is probably a corner of the warp rectangle and the trigger area is around the arrival cell, not around the listed point. Not measured.

### Result 2: Cave Level 1 (DW.m4)
V: 3 of 4 entries bounce in one run: the other path is the fight. `Fighter.Clear` (combat/fight.go: `fight` 169, `engage` 449, `move` 815, `f.mv.GoTo` 822) decides `approach why=no shot dist=7` toward a Dark Ranger at (258,153) beside the mouth; `nav walk callers=GoTo < combat.move < engage < fight`; here `mouthStep` DOES fire (`nav mouth exit dest=3 exit={255 145} dist=7.3 skip=false`, `nav: stepping away from a cave mouth from={257 152} to={262 171}`) but its first click `wp={257 156}` fires the trigger anyway (`arrived in level 3`, `combat move failed ... the level changed during the walk: 9 -> 3`). This is candidate (a): the approach; the step-away direction was wrong on the first try (FIXNAV turns it only after a bounce). In the third entry of that run the roam bounce of Result 1 recurs. Run 2 of the drill differed from run 1 (Den was identical), so the Cave Level 1 numbers are 2 clears and 1 explore bounce of 4 entries (run 1), not a stable count.

### Proposed fix (not implemented)
Smallest: make `mouthStep` measure from where d2d puts the character, not only from the listed exit position. Treat "this walk is the first of the level after a warp arrival" (the navigator already sees the level change; n.walkSeq/startLevel) as being in the mouth whatever the distance up to about 16 subtiles (Den 11.2, later bounce positions 16.2 to 18 in m4a did not fire), or raise `mouthR` from 9 to 14 and use the exit position plus the (+5,+10) offset; then the step-away goes first and `roam` starts from a cell outside the trigger. Offline test: `internal/nav` with `newKnown` on an open field, the exit at (2048,2050) and the character at (2053,2060) (distance 11.2): `mouthStep` must return a step (today `ok` is false); plus a second case at distance 20 that must stay false. Second part (candidate a): `Fighter.move` should not approach a target within `mouthR` of an unwanted exit unless `MouthGuard` is on (combat config.go:56 default off; DW.m4 is the live case), test with the fake navigator `NearMouth`. Whether the step-away then actually avoids the trigger is UNVERIFIED (Result 2 shows it can fire it on the first step); do not claim the fix before a DW.m3/DW.m4 rerun at debug level.

### Closing block: shared-file lines for the merger
- docs/BACKLOG.md: FIXDENBOUNCE: cause named by drills14 (nav mouthStep mouthR 9 vs arrival at 11.2 subtiles; Explore/roam first walk fires the trigger; Fighter.Clear approach in Cave Level 1); fix proposed in the DRILLS14 section, not implemented.
- docs/INDEX.md: DRILLS14 section in docs/probes/drills/README.md.
- CHANGELOG: none (no product change).

## DRILLS15 (2026-11-01): first live skill-smoke drills (teleport, mana, missile, armor traces)
Result file `2026-11-01-drills15-full.txt` (stdout of the one d2drill invocation, scrubbed, 15 KB; earlier partial runs while the drills were being fixed were not kept). New group `R` (internal/drill/d15.go, cmd/d2drill/d15_lab.go, d15b_lab.go; overlap declared: the group switch and the end-of-run goto town in cmd/d2drill/main.go). Run on the worker host, hidden slot 0, `lab-debug-skills` binary, lockstep, throwaway Sorceress `--no-save`: `D2D_AGENT=drills15 D2D_BIN=<fork>/bin/d2d.lab-debug-skills tools/agentrun.sh drills15 --ttl 600 -- tools/d2dslot.sh run -n 0 -t 580 -- <built>/d2drill -group R -repeat 1 -label drills15-full`. Capacity: idle 45 percent (44 percent iowait) on the worker host before the first run. A whole invocation takes under 10 s. All state changes go through `r.Cheat` and the ledgered `d2dctl.SkillDebug` (`debug cast`, `mana full`, `states clear`); reads are `dump`, `debug players`, `monsters`, `debug missiles`, `debug blocked`. PASS means d2d equals the trace's game.exe expectation (a read of game.exe, V; game.exe itself was NOT run); FAIL means d2d differs (a deviation, not a harness error).

| Trace item | Drill | Verdict | Numbers (V: read from the log) |
|---|---|---|---|
| sk-teleport T1 / T7-town | D15-T1 | PASS | in town two casts (3 and 5 cells): no hop (the character walks 1.1 cells toward the target in 14 ticks), mana 35.0 before and after |
| T2 room limit | D15-T2 | FAIL | hops of 3, 6, 9, 12, 15, 18, 24 cells (15 to 120 subtiles) along one clear line: all 7 land on the target (y drifts 0.1), each costs 23.8 points (level 1 cost 24, regen included). game.exe refuses outside the room list: d2d has no room limit (trace D1). Which of the near hops game.exe allows depends on the caster's place in its room (U: the drill cannot say) |
| T3 blocked target | D15-T3 | FAIL | 3 blocked cells at 2.7, 3.2, 4.3 cells: the character does not move (0.00 cells) and 23.8 points of mana are spent each time. game.exe snaps to the nearest free spot and pays at the action frame (trace D2); d2d refuses AND pays (also part of D4) |
| T4 monster | D15-T4 | FAIL | zombie 3 cells away: a point target and a unit target both put the character 0.20 cells from the monster, on it (trace D3). game.exe lands beside it |
| T5 merc | D15-T5 | FAIL | merc hired with `debug merc hire 0`, readable in `dump` (merc x y). A 6 cell hop: the merc moved 0.09 cells in the landing tick (its own walk step) and stayed 4.5 to 5 cells behind; it walked 3.06 cells in 45 ticks. d2d does NOT move the merc with the hop (game.exe: moved with the owner, a read of FUN_00650be0, UNVERIFIED by play). Pets: a Sorceress has none, `pets` lists nothing |
| T7 mana on a refused hop | D15-T7 | FAIL (not reachable) | exactly 24 points, 25 cells: the hop is not refused on d2d (moves 25.30 cells), mana 24.06 -> 0.23. There is no refusal to price at that distance; the priced refusal is T3 (23.8 paid) |
| T6 Duriel's lair | not run | | unreachable |
| sk-mana M1 | done by DRILLS13 | | |
| M2 interrupted cast | D15-M2 | FAIL | a real Fallen blow (`debug mon set dmg 20 20`, life 5000, the cast started 0 to 7 ticks after its A1 begins), 48 casts: 7 cast cut by the blow before the missile (blow at +1, +4, +5): 7 of 7 had the mana paid and no missile (game.exe pays nothing, trace D3); 1 blow after the missile; 3 casts started in hit recovery. The castcut result of DRILLS11 (weak zombie blows of 2 to 3 do not cut a cast) did not repeat for a 20 damage blow: a big blow cuts it |
| M3 Teleport floor | D15-M3 | FAIL | level 25 (`debug points 54 25`, cost formula 0): mana 0.55 point and 0.05 point: no hop (0.00 cells), mana only regenerates; mana 1.04 point: hops 4.3 cells and pays 1 point. d2d needs a whole point (trace D2); game.exe casts at any mana |
| M4 regen | D15-M4 | PASS | max mana 45 (`debug stat 9`), mana 0.04 -> 8.8 points over 750 ticks: +2250/256, exactly game.exe's 3 per frame; Warmth 1 (`debug points 37 1`) the same 2250. The bot's model of 13.5 points over 30 s is 4.7 points too high (trace B2) |
| M5 potions | not run | | no mana potion in the lab pack and no item-give verb |
| M6 skill delay | D15-M6 | FAIL | Blizzard (level 1, 23.0 points): first and second cast 25, 40, 60 and 100 ticks apart (1 to 4 s) all enter cast mode, pay 23.0 points and make 4 missiles (the 100 tick one 1). `debug states` shows no delay state. game.exe refuses the early ones (trace D6): d2d has no skill delay |
| sk-missile K1 reach | D15-K1 | PASS | level 1 at a point 12 cells away on a clear line: Fire Bolt farthest 46.5 subtiles (trace 46.9), 50 ticks of flight, first row at +8 ticks; Ice Bolt 28.1 (28.1), 50 ticks; Lightning 34.3 (35.2), 25 ticks. Sampled per tick, so the figure is the last seen position (0.4 to 0.9 under the trace) |
| K2 Charged Bolt count | D15-K2 | PASS | levels 1, 5, 20 (`debug points 38 n`): 3, 7, 22 distinct bolts, all alive at once from +8 ticks, all starting on the caster's cell (0.0 subtiles), last alive 77 frames after the first |
| M4 roll, M5 Lightning pass-through / Chain Lightning | not run | | statistical (thousands of hits) / not scripted |
| sk-armor E1 | D15-E1 | FAIL | Energy Shield (`58:1` after the cast), 3.02 points of mana, `debug hit 50` on 100 life: life lost 49.00 (absorbed 1.0), mana 3.02 -> 1.05, state `58:1` stays (game.exe: 48.5 lost, mana 0, state removed, trace D3). After `debug mana full` a second 50 blow: 40.0 life lost, mana 35.0 -> 15.0, shield still on (game.exe: no shield until recast). The numbers of d2d match the trace's description of d2d |
| armor A2 to A5 | not run | | time |

### Findings
- All thirteen are d2d observations; the PASS or FAIL against game.exe is the comparison to the traces, no game.exe run (a game.exe read on a hidden instance is the open step). Trace deviations D1 (room limit), D2 (blocked target, minmana floor), D3 (lands on monsters, cut cast keeps mana, shield state), D4 (mana on a refusal), D6 (skill delay), D7 (merc not moved) are all CONFIRMED live on d2d. Deviations that were open now have numbers; none is a bot bug.
- For the bot: Teleport hops of 18 cells and shorter all land on d2d, so the bot's nav hops work there; a cast at a blocked cell costs a full 24 points on d2d for nothing (trace B2), and the bot's 13.5 point regen model is 4.7 points high (M4).
- The reach numbers (K1) match the trace within 1 subtile for all three, so the bot's cast ranges can be checked against 46.5 / 28.1 / 34.3 subtiles (the bot's own cast ranges were not read here).

### Harness notes
- `dump stats` raw[] shows a cheat's new life or mana only after the next tick (V: the `debug hit` reply said life 51 while raw[6] still read 100 on the same tick): drills step one tick after every set. The first versions of D15-M6 and D15-E1 read stale values and gave wrong numbers; the final log is after the fix.
- A cast the engine refuses can still start a walk toward the target (town, T1: 1.1 cells in 14 ticks): classify a hop by the distance covered (at least 80 percent of the way), not by "moved".
- `debug blocked x y` takes level-relative cells and returns 1 for blocked; the long clear lines are found by scanning it at half-cell steps, the start moved to the far end of the line.

### Shared-file lines for the merger
- BACKLOG: SKTELE/SKMANA/SKMISS/SKARMOR first live runs done on d2d (drills15): teleport D1 to D4, D7, mana D2, D3, D6, armor D3 confirmed with numbers; missile reach and Charged Bolt count equal the trace; not run: T6, M5, M4-roll, missile M5, armor A2 to A5 and every game.exe-side read.
- INDEX / CHANGELOG: "drills15: first live skill-smoke drills (group R): Teleport, mana, missile and armor traces compared with d2d; deviations of the traces confirmed live".

## DRILLS16 (2026-11-01): live proofs of fixmouth (FAIL) and fixbuff (PASS)

Run on the worker host, hidden slots, current main (75bc85a) plus the new group R. Files `2026-11-01-drills16-<label>.txt`: `m31`..`m33` (DW.m3 x3; m31 is the full filtered log, m32/m33 the first 300 filtered lines: the three runs are identical), `m41` (DW.m4), `bs14` (R-bs1, R-bs4), `bs3` (R-bs3). The -v logs are 10 to 16 MB (the Den Clear phase spins until the 200 s TTL): the m3/m4 files keep only the nav mouth, step-away, arrival, level-change and caller lines (consecutive repeats collapsed). Commands: DW: `D2D_AGENT=drills16 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh drills16 --ttl 300 -- tools/d2dslot.sh run -n 0 -t 200 -- <built>/d2drill -group F -only DW.m3 -label drills16-m3-N -v` (and DW.m4); R: same with `d2d.lab-debug-skills`, `-n 1`, `-group R -only R-bs1,R-bs4` / `R-bs3`. Capacity before: worker idle+nice 44 percent.

### A. fixmouth: FAIL (a bounce remains), not fixed here
- The arrival guard works (V, every arrival of every run): `nav mouth check ... arrival=true r=16` then `nav mouth exit dest=2 exit={35 135} dist=11.2 skip=false` in Den (Cave Level 1: `exit={255 145} dist=11.2`), `nav: stepping away from a cave mouth` from (40,145) to (49,163) (Cave Level 1 (260,155) to (269,173)), 4 clicks, no trigger. The first Explore walk of Den (entry 1) steps away, then roams on in level 8 with no bounce: the DRILLS14 Result 1 signature is gone.
- DW.m3 (3 of 3 invocations identical, 26 arrival walks each): the drill does not finish (the Clear phase never ends; killed by the 200 s TTL, no DRILL line). 23 `the level changed during the walk` errors per run. Chains (V, counted from the nav walk line before each error): `GoTo < combat.(*Fighter).move < combat.(*Fighter).engage < combat.(*Fighter).fight` (7 times, 8 -> 1 after an `escaped ... after a stall in a pocket`), `walk < GoTo < Fighter.move < Fighter.evade` (6, 8 -> 2), `walk < GoTo < Fighter.move < Fighter.engage` (2 + 5), `GoTo < Fighter.move < Fighter.evade < Fighter.engage` (2). First bounce, log line 621: Clear's `engage` approach walks to a Zombie at (43,141): `nav mouth exit dist=9.4 skip=true` (mouthR 9), the next walk is at 8.9 and steps away to (61,148), then `Fighter.move` walks back to (43,141) and fires the trigger (`arrived in level 2`). After that the loop repeats; one run logs 25,565 `stepping away` lines, 3.2 from the exit at (36,138) (evade alternating with the step-away: a spin).
- DW.m4 (1 invocation, 8 s; the drill's own verdict `FAIL 2 bounce(s) out of the cave, first in explore`, entries 1 and 2 of 4): the bounce is Explore's roam first walk: `roam < explore < Explore`, position (267,150), `arrival=false` (the Clear phase's first walk had consumed the arrival), `nav mouth exit dest=3 exit={255 145} dist=13 skip=true` (13 > 9), first click wp (265,151) fires the trigger. Also in this log: `Fighter.move` bounces 9 -> 3 / 9 -> 1 from `engage` and `evade` (5+3+3+2).
- The U of the card is resolved: the Cave Level 1 step-away click (arrival radius 16, to (269,173) / (262,167) / (265,160)) did NOT fire the trigger in 4 of 4 entries.
- Reading: the arrival radius only covers the first walk. Two cases remain: (1) any later walk that starts 9..16 subtiles from the listed exit (roam after a Clear walk consumed `arrival`, Cave Level 1: 13), (2) Fighter.move approach/evade targets within a few subtiles of the mouth in Den, which also oscillates against the step-away. Candidate: keep `mouthArrivalR` for the whole stay (until the character is 16+ from the exit once) instead of one walk, and filter approach/evade targets within `mouthR` of an unwanted exit unconditionally. Not implemented (brief: stop as FAIL).

### B. fixbuff: PASS on all four live items
1. Raw reply on the production lab build (V): before a cast `absorb=0 maxlife=40 maxmana=35 def=6 buffs=`; after `cmd skill 40`: `absorb=0 maxlife=40 maxmana=35 def=7 buffs=40:1,`; after `debug states clear 40`: back to `def=6 buffs=`. `d2dctl.Client.Buffs` parses it to [40:1].
2. `Player.States` is filled (V, snapshot through d2denv with BuffStates on): bit 10 set in 12 of 12 samples over 30 s of game time with the armor up (R-bs1), clear in 5 of 5 after the clear. In the long fight, 66 of 66 up samples had bit 10 set. No refusal note (`debug states` accepted).
3. Long fight (R-bs3, V): level 1, Blood Moor, one Zombie held with `debug mon set` (maxlife 1000000, life 1000000, dmg 0-0), the bot's own `Fighter.Clear` in 66 rounds of 6 s (1 s from 100 s to 130 s after a cast), 300 game seconds, no need-buff edge. First cast manual (`Fighter.Buff`) at 0; the fight loop recast at 115.1 s and 229.2 s (limit 114 s plus the round length), armor absent in 0 of 66 samples (longest gap 6.08 s; the armor lasts 120.3 s). Without the change the armor would have dropped at 120 s.
4. Ids (V, `debug states set <skill> 1` and, for 50/58/60, real casts after `debug points`): skill 50 Shiver Armor reports `buffs=50:1,` (def 8) and Player.States bit 88; skill 58 Energy Shield `58:1,` bit 30; skill 60 Chilling Armor `60:1,` bit 20; skill 68 Bone Armor `68:1,` (absorb=20) bit 14; 40 -> bit 10. All match the new d2.State* ids. `debug points 68 1` refuses on the Sorceress ("not a class skill"), so Bone Armor was only set, not cast. Verb replies: `state shiverarmor of skill 50 Shiver Armor lvl=1` etc.
5. Stale-cache loud failure: offline test only (not run live). The recast at 115 s (not the 150 s clock fallback) shows the AuraFrames path was active with the lab's tables cache; whether that cache had been rebuilt before this run was not checked (U).

### What was done to the character
Drilly (throwaway Sorceress, `--no-save`): `debug goto 1/2/3`, `debug kill`, `debug points 40/50/58/60 1`, `debug stat 6/7` life, `debug spawn zombie1`, SkillDebug `states set/clear`, `mana full`, `mon set` (all in the ledger). Ends alive in Rogue Encampment, panels closed (`END character Drilly ... in town: true`).

### Shared-file lines for the merger
- docs/BACKLOG.md: FIXMOUTH2: fixmouth live proof FAILED (DRILLS16): the arrival step-away works but later walks bounce (roam first walk 9..16 from the exit after arrival is consumed; Fighter.move engage/evade targets near the Den mouth); fix proposed in DRILLS16 A, not implemented. fixbuff live proof done, PASS.
- docs/INDEX.md: DRILLS16 section in docs/probes/drills/README.md.
- CHANGELOG: the armor recast inside long fights and the d2d armor states are verified live (fixbuff); none for fixmouth.

## DRILLS17 (2026-11-01): live proof of fixmouth2 (FAIL)

Run on the worker host, hidden slot 0, lab build cheats3, `d2drill` built from origin/main 2d1b4ba (fixmouth2 merged as b2c88b8). Capacity before: idle 3.7 plus nice 84.4 percent. Files `2026-11-01-drills17-<label>.txt`: `m31`..`m33` (DW.m3 x3; m31 the filtered log, m32/m33 the first 400 filtered lines; the three runs are identical), `m41` (DW.m4), `d3` (town group, DRILL lines and mouth lines), `d2` (regression, full stdout). The -v logs are filtered to the nav mouth, step-away, arrival, level-change, walk-caller and DRILL lines (consecutive repeats collapsed). Command: `D2D_AGENT=drills17 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh drills17 --ttl 300 -- tools/d2dslot.sh run -n 0 -t 200 -- <built>/d2drill -group F -only DW.m3 -label drills17-m3-N -v` (and DW.m4; groups D3 and D2 without `-only`). Each invocation took about 6 s of wall time.

### Result: fixmouth2 FAILS, but with real progress
| Drill | DRILLS16 | DRILLS17 |
|---|---|---|
| DW.m3 x3 | Clear never ends (200 s TTL), 23 `the level changed during the walk` per run, about 25,565 `stepping away` lines | all 3 runs identical and FINISH in about 6 s: verdict FAIL `3 bounce(s)` (run 1: entries 2, 3, 4 bounce, first in clear; run 2: 4 of 4, first in explore); 18 level-change errors per run, 14 `stepping away` lines, no `evade: capped` line (no spin, so the cap never had to fire: U whether it works live) |
| DW.m4 | 2 bounces in 4 entries (first in explore) | 2 bounces in 4 entries (first in explore): unchanged |

### Callers (V, from the m31 and m41 logs)
1. DW.m3, Den of Evil, the bounce at log line 620 (entry 2 clear): `GoTo < combat.(*Fighter).move < combat.(*Fighter).engage < combat.(*Fighter).fight`, approach to a Zombie at (43,141): `nav mouth check arrival=false r=9`, `nav mouth exit dist=10 skip=true`, then dist 9.4 and 8.9 (`skip=false`, `stepping away` to (61,148), 3 clicks), then the engage approach plans (60,147) to (43,141) (25 cells) and its last click at (42,146) fires `arrived in level 2`, `combat move failed ... level changed during the walk: 8 -> 2`. The Fighter did NOT skip (43,141), which is 10.0 from the exit (35,135): well inside `mouthApproachR` 12.
2. Why: `NearMouth` answers only when `n.mouthGuard || n.avoiding() || n.arrival`. The drill build leaves mouthGuard off (the bot's default too: cmd/d2bot/d2dwire.go sets it from config key `mouthGuard`, default off). `n.arrival` was already false: the first walk after the arrival stepped away to (49,163), the next `mouthStep` found the start 28.7 from the exit (beyond 16), and the held flag ended there (`nav mouth check pos=(37,151) arrival=false r=9` is the next line). So the hold ends after ONE step-away in Den, and every approach after it (Clear's engage) is unguarded and walks back into the 9.4 to 12 subtile band. The arrival flag is a one-step guard in practice here; the card's "whole stay" holds only when no walk starts 16 away.
3. DW.m4 bounce 1 (entry 2, explore): `roam < explore < Explore`, position (267,150), `arrival=false r=9`, `nav mouth exit dest=3 exit={255 145} dist=13 skip=true`, first click wp (265,151) fires the trigger: the same cause (flag ended after Clear's first walk cleared 16, the character wandered back to 13, outside mouthR 9).
4. DW.m4 bounce 2 (entry 3, explore): `arrival=true r=16`, position (256,148), 3.2 from the exit, `stepping away` to (262,167) and its second click wp (260,156) fires the trigger. The step-away itself walks too close to the trigger when the arrival lands inside it (DRILLS16 saw 4 of 4 step-aways safe: here 1 of 3 fired; the start was 3.2 from the exit rather than 11.2).

### Arrival flag across Clear and Explore (V)
Set on every `arrived in level` (`arrival=true r=16` on the first `nav mouth check` of a stay: Den at (40,145), 11.2 from the exit, Cave Level 1 at (260,155)); it ends on the next walk start beyond 16 of every unwanted exit (28.7 to 31.3 in Den after the step-away), after which every walk logs `arrival=false r=9` through the rest of Clear and Explore until the next level change. It does not carry between stays. Proposal (not implemented): keep NearMouth guarded for the stay in a cave level whose arrival was inside the mouth (a separate `stayGuard` that stays true until the level changes), or turn `mouthGuard` on for d2d by default; the DRILLS16 note that a bigger radius widens fight filtering still applies.

### Side effects and regression (V)
- Town: D3 12 of 12 PASS (`d3`: g0, n, h, b, b2, tp, r, r0, s, s2, u, i). Rogue Encampment lists `exits=0` tile exits, so the arrival flag (set on the level change, `arrival=true wantExit=2` seen once on the way back) guards nothing there and has no effect; every D3 `nav mouth check` shows `arrival=false`. Not checked: a town with tile exits.
- D2 (DRILLS11 group): 11 of 11 PASS (`d2`).

### What was done to the character
Drilly (throwaway Sorceress, `--no-save`): `debug goto 1/2/3`, `debug stat 6/7` life (512000) in DW.m3/DW.m4 (ledgered cheats `drills:<drill id>`); portal and town drills of D2/D3 as in DRILLS11. Ends in Rogue Encampment, panels closed.

### Shared-file lines for the merger
- docs/BACKLOG.md: FIXMOUTH3: fixmouth2 live proof FAILED (DRILLS17): DW.m3 x3 3 bounces in 4 entries and DW.m4 2 in 4 remain. The arrival flag ends after the first step-away (start 28.7 from the exit), then NearMouth is off (mouthGuard default off) so engage approaches to (43,141) in Den and roam at 13 in Cave Level 1 bounce; a step-away from 3.2 also fires the trigger once. Clear now finishes and the spin is gone. Fix proposed in DRILLS17, not implemented.
- docs/INDEX.md: DRILLS17 section in docs/probes/drills/README.md.
- CHANGELOG: none.

## DRILLS18 (2026-11-01): live proof of fixmouth3 (FAIL, large progress)

Run on the worker host, hidden slot 0, lab build cheats3, `d2drill` built from branch agent/fixmouth3 (tip 2a9234a) merged with origin/main. Capacity before: idle 29 plus nice 33 percent. Files `2026-11-01-drills18-<label>.txt`: `m31`..`m33` (DW.m3 x3; m31 the filtered log, m32/m33 the first 400 filtered lines; the three runs are identical), `m41` (DW.m4), `d3`, `d2` (DRILL lines). Filter: nav mouth, step-away, level-change, walk-caller, stall and DRILL lines, repeats collapsed. Command as DRILLS17 (`-group F -only DW.m3 -label drills18-m3N -v`, `-only DW.m4`; groups D3 and D2 without `-only`). Each invocation took about 7 s of wall time.

### Result
| Drill | DRILLS17 | DRILLS18 |
|---|---|---|
| DW.m3 x3 | 3 of 4 entries bounced, 18 level-change errors per run | all 3 runs identical: verdict FAIL `2 bounce(s)` of 4 entries; 1 real mouth bounce (cycle 2 clear) plus 1 town-portal escape from a pocket stall (cycle 3 clear, area 1); 18 errors per run, but 16 are one stale move retried after the escape (`8 -> 1`), 1 the real bounce (`8 -> 2`), 1 the escape; 11 `arrived in level` lines, no `Entered` cycle within 1 s except the one bounce |
| DW.m4 | 2 of 4 bounced | 0 bounces in 4 entries, 0 level-change errors; the drill then timed out `walk back out` (2 min) and left the inventory open (P35 dirty UI, group stopped; the later D3 group started clean) |

Clear still FINISHES in both drills (no TTL). `stayLevel`: V. The first `nav mouth check` on arrival shows `arrival=true stay=false r=16`, the first step-away sets it (`stay=true` from the next check), and every later check in the level logs `stay=true r=16` (arrival goes false after the first walk clear of the exits) until the level changes. 11 `nav: stepping away from a cave mouth` lines per DW.m3 run (DRILLS17: 14); DW.m4 8 (DRILLS17: 6). The evade cap never fired. `combat move skipped: the point is in a cave mouth` fired 7 times in m3 and once in m4: the Fighter filter works.

### The remaining bounce (V, m31 log)
Caller chain (V): `nav.walk < nav.GoTo < combat.Fighter.move < combat.Fighter.engage < combat.Fighter.fight`. Engage targets a Fallen and calls `move` to (37,150): 15.1 from the exit (35,135), just outside `mouthApproachR` 15, so not skipped. The walk's `mouthStep` at (37,148), 13.2 away, steps away to (40,168) (`stay=true r=16`). The step is blocked by a monster (`nav stuck`, `stalled by monsters, yielding`), the fight loop runs inside the step-away, and its basic-attack swings (no bound attack skill in the drill) walk the character toward the Fallen at the mouth: 7 `approach` decisions, 7 `move skipped` lines, and the character ends at (36,138), 3.2 from the exit. The step-away walk then replans from there (29 cells, 5 waypoints) and its FIRST click wp (37,140) fires the trigger 16 ms later (`arrived in level 2`). So: click distance 5.4 from the exit, start 3.2; the replan path first goes along the corridor toward the exit, so the `leavesExit` check (done once, at the original start 13.2) did not cover the replan. The other failure in the same run is not a mouth bounce: a pocket stall at (90,171) (`nav stuck` rung 0..2, `escaping by town portal`), which the drill counts as leaving the cave. DW.m4 had neither.

### The long single click (UNVERIFIED, partly)
No step-away's single long click fired the trigger in the 11 plus 8 step-aways of these runs: every step-away that ran without interference completed. The 3.2 start of DRILLS17 was not reproduced in a clean case. The one failure above is a REPLAN of the step-away after interference (oneClick only applies while the goal is in line of sight; the path from (36,138) was cornered, so ordinary waypoints were used). U: whether the long click itself is safe at a start of 3.2 in Den.

### Coverage (V, weak evidence)
`nav plan` lines: 664 in m3, 519 in m4; roam walks in the filtered m3 log 92 (DRILLS17 m31: 41, m41: 130; the filters differ, so only the order of magnitude is comparable). The sweep did not starve: Explore still logs roam walks in the Den after the step-away (m31, cycles 0 to 3), and cave Level 1 in DW.m4 reached `walk back out` after four full cycles. No per-point visited count exists in the logs: U.

### Regression (V)
D3 12 of 12 PASS (`d3`), D2 11 of 11 PASS (`d2`).

### What was done to the character
Drilly (throwaway Sorceress, `--no-save`): `debug goto 1/2/3`, `debug stat 6/7` life (ledgered cheats `drills:<drill id>`), portal and town drills of D2/D3. Ends in Rogue Encampment (`END character Drilly ... in town: true`), panels closed (the D3 and D2 groups cleaned the inventory left open by DW.m4).

### Shared-file lines for the merger
- docs/BACKLOG.md: FIXMOUTH4: fixmouth3 live proof FAILED narrowly (DRILLS18): DW.m3 x3 down to 1 mouth bounce per run, DW.m4 0 bounces. The remaining bounce is a step-away replan after a fight inside it (monster stall) pulled the character to 3.2 from the exit; the replan's first click (37,140) fired. Also a Fighter.move target 15.1 from the exit (just outside mouthApproachR 15) and a pocket-stall town portal escape. Fix proposed in DRILLS18, not implemented: re-run `leavesExit`/mouthStep on every step-away replan and refuse fight swings toward a mouth during the step-away.
- docs/INDEX.md: DRILLS18 section in docs/probes/drills/README.md.
- CHANGELOG: none.

## DRILLS19 (2026-11-01): live proof of fixmouth4 (PASS, with a hold-loop finding)

Run on the worker host, hidden slot 0, lab build cheats3, `d2drill` built from branch agent/fixmouth4 (tip 80bf381, nothing else changed; origin/main merged as a no-op). Capacity before: idle 52 plus nice 0.4 percent. Files `2026-11-01-drills19-<label>.txt`: `m31`..`m33` (DW.m3 x3; m31 and m41 are filtered to the gate, step-away, arrival, exit, stuck, portal and DRILL lines with similar repeats collapsed, 1500 lines at most; m32/m33/m42/m43 the first 300 to 400 lines of a wider filter), `m41`..`m43` (DW.m4 x3), `d3`, `d2` (regression). Command as DRILLS17/18 (`-group F -only DW.m3 -label drills19-m31 -v`, `-only DW.m4`; groups D3 and D2 without `-only`; build `d2drill` on the worker through `labrun`). Each run took 6 to 7 s of wall time. The three DW.m3 runs have the same shape (same line count 8,741 and same counts); so do the DW.m4 runs (3,722 or 3,723 lines).

### Result: PASS
| Drill | DRILLS18 | DRILLS19 (fixmouth4) |
|---|---|---|
| DW.m3 x3 | 1 real bounce per run plus 1 pocket-stall town-portal escape; 18 level-change errors | `DW.m3 P23 PASS no bounce in 4 entries` x3: 0 `level changed during the walk` errors, 0 extra `arrived in level` lines (5 of level 8 and 5 of level 2 = the first entry plus 4 exits and 4 re-entries as the drill drives them), no `Entered` cycle within 1 s, no town-portal escape |
| DW.m4 x3 | 0 bounces in 4 entries | 0 bounces, 0 level-change errors in all 3 runs; verdict FAIL only on the dirty-UI check (`open: inventory`), the same harness issue as DRILLS18 (P35) |

Clear and Explore both FINISH in every cycle (every trace phase `ok` in DW.m3). Coverage: `roam: coverage points` 425 (Den), 551 (Cave Level 1); `nav plan` lines 559 in m31 (DRILLS18: 664) and 188 in m41 (DRILLS18: 519, but the DW.m4 character dies earlier here, see below); no per-point visited count exists in the logs: U.

### Gate activity per run (V, counted in the raw logs)
| Line | m31..m33 (each) | m41..m43 |
|---|---|---|
| `nav mouth gate: click rewritten` (all `how=away`) | 351 | 1 (`how=clip`, from (258,162) to (264,156) as (261,159)) |
| `nav mouth gate: click held` (`why=rewrite cap`, cap 3) | 117 | 0 |
| `ErrMouthGate` text | 0 (returned 117 times, no caller logs it) | 0 |
| `nav: stepping away from a cave mouth` | 129 (DRILLS18: 11) | 4 |
| `combat swing held: the click would walk into a cave mouth` | 16 | 0 |
| `combat move skipped: the point is in a cave mouth` | 7 | 0 |

### The DRILLS18 chain recurred and the gate caught it (V, m31 lines 1380 to 1450)
Den cycle 2 (the third entry): Fighter.engage approaches a Fallen, the step-away is blocked by monsters, 7 `approach` decisions are skipped by the Fighter filter, 16 basic-attack swings toward (37,141) are held (`swing held`, from (36,138)), then the step-away replan starts at 3.2 from the exit (35,135): `step-away replan heads to the mouth, new goal` (old (40,168), new (42,157)) and the first click wp (37,140) is rewritten `how=away` to (42,157). No `arrived in level 2` follows: DRILLS18 fired exactly here. The proof therefore covers the real failing case, not only an easy one.

### Finding: hold loop (V) - a defect, not a bounce
After the 3rd rewrite in the walk the gate holds (`click held ... cap=3`), the walk fails, and the caller (Fighter.engage via move, then the nav step-away) starts a new walk that is rewritten the same way and held again: 117 hold lines and 117 `stepping away` lines at the same spot (36,140), 5.1 from the exit, over roughly 0.9 s of wall time, until the drill's phase ended the cycle (`nav: timeout: walk did not finish in 2m0s` 173 times in m31). The character does not move because monsters block the away route (`nav stalled by monsters, yielding` in the same window), so the walk fails over and over. The brief's expected outcome (the walk fails with `ErrMouthGate`, hold) happens, but nothing ends the stay at the mouth: the caller retries instead of fighting in place or giving up. Cost: one Den cycle burned its time budget (the other three cycles were clean); the step-away count per run rose from 11 to 129. The spec of fixmouth4 says "at most twice" re-checks on replan; the 117 repeats are fresh walks by the caller, not replans. Proposal (not implemented): when the gate holds, let the caller fight in place (swing gate already holds the swing) or treat ErrMouthGate as a reason to ignore the target for this engage; and log the error once where it surfaces.

### Real trigger bound (U)
No click fired from beyond 15 in these runs (no bounce at all), so the trigger bound is unconfirmed by a firing; every unwanted-exit click the gate allowed was at least 15 away or outward. m41 shows one `clip` of an Explore click: from (258,162) toward (264,156) (14.2 from the exit (255,145)), cut to (261,159), 15.2 from it. What would show it: a drill that clicks at 15 to 17 on purpose; not done.

### The pocket stall (V): a fight, not a nav defect
This time there was no town-portal escape in DW.m3 (the character passes (90,171) with a single click, `nav click pos (90,171)`, and goes on). The long stall is at (108,178), cycle 3, after the cleared cycles: 993 `nav click` lines at one position and a 288 s fight summary with `kills=0 casts=1` against 28 hostiles (target a Fallen Shaman), `combat pinned: evasive moves go nowhere, fighting in place`, `nav stalled by monsters, yielding near=6`, then `combat: left the game to survive` (36 times, no town portal scroll in the pack of a fresh character). The character has no bound attack skill (`swinging the basic attack`), so the swarm cannot be killed and nothing frees the path: an unfinished fight against 6 adjacent monsters, which matches the author's reading. In DW.m4 the character died in cave level 1 cycle 1 (23 monsters, life 70 to 0, fight window 167 s, one town portal cast that had nothing to carry it, `crowd-abandon unavailable: no town portal scroll or tome`), so cycles 2 and 3 of DW.m4 ran on a dead character (`nav: character is dead`); DRILLS18 m4 also logged a death. Only the first two DW.m4 entries are live evidence (0 bounces). The next drill should give the character a bound attack skill or cheat in a full belt.

### Regression (V)
D3 12 of 12 PASS (`d3`), D2 11 of 11 PASS (`d2`). Final `END character Drilly ... in town: true` in both (D3 life 40/40, D2 34/40).

### What was done to the character
Drilly (throwaway Sorceress, `--no-save`, fresh in every invocation): `debug goto 1/2/3`, `debug stat 6/7` life in DW.m3/DW.m4 (ledgered cheats `drills:<drill id>`), portal and town drills of D2/D3. It died inside DW.m4 (each invocation starts a new throwaway); the last run (D2) ends in Rogue Encampment, alive, panels closed.

### Shared-file lines for the merger
- docs/BACKLOG.md: FIXMOUTH5 (optional): fixmouth4 live proof PASSED (DRILLS19): DW.m3 x3 and DW.m4 x3 zero mouth bounces and zero level-change errors, including the DRILLS18 chain (replan from 3.2 caught, 351 rewrites and 117 holds per DW.m3 run). Defect: after the 3-rewrite cap the held walk fails and the Fighter restarts a step-away at once, 117 times at (36,140) in one Den cycle (monsters block the away route); make ErrMouthGate end the engage or fight in place, and log it. DW.m4 drill character dies in cave level 1 (no attack skill, no scrolls): fix the drill setup.
- docs/INDEX.md: DRILLS19 section in docs/probes/drills/README.md.
- CHANGELOG: fixmouth4 (cave-mouth click gate) verified live; nothing else.

## DRILLS20 (2026-11-01): live proof of fixmouthloop (FAIL: the loop is inside nav, not in the Fighter)

Run on the worker host, hidden slot 0, lab build cheats3, `d2drill` built from branch agent/drills20 (agent/fixmouthloop tip 27285b3 plus origin/main merged). Capacity before: idle 47 plus nice 0.8 percent. Files `2026-11-01-drills20-<label>.txt`: `m31` (DW.m3, 1200 filtered lines), `m32`/`m33` (first 300 lines), `m41` (DW.m4, first 400), `d3`, `d2` (DRILL lines). Command as DRILLS19 (`-group F -only DW.m3 -label drills20-m3N -v`, `-only DW.m4`; groups D3 and D2 without `-only`). Every DW.m3 run is identical to the DRILLS19 run, count for count.

### Result: FAIL (the restart storm is unchanged)
| Item | DRILLS19 | DRILLS20 (fixmouthloop) |
|---|---|---|
| DW.m3 x3 verdict | PASS no bounce in 4 entries | same, 0 bounces, 0 `level changed during the walk` (V) |
| `click rewritten` / `click held` per run | 351 / 117 | 351 / 117 (V) |
| `stepping away from a cave mouth` per run | 129 | 129 (V; target was about 20 or fewer) |
| `walk did not finish in 2m0s` per run | 173 | 173 (V) |
| `combat move held by the mouth gate, backing off` | not applicable | 0 lines (V): the Fighter back-off never ran |
| DW.m4 | 0 bounces, 1 rewrite, 4 step-aways | the same (V); verdict FAIL only on the dirty-UI check as before |
| Fights | kills 0 everywhere | kills 0 in all 284 fight summaries; no new stall seen, no improvement |
| Regression | D3 12/12, D2 11/11 | D3 12/12, D2 11/11; ends in town, life 34/40 (V) |

### Why it does not fire (V, m31 lines 1437 to 1470)
No `combat move failed` line mentions ErrMouthGate: the hold is not seen by `Fighter.move`. The repeats are inside ONE `Navigator.walk`: `walk.go:74` increments `n.walkSeq` on every call, including the nested step-away walk started at `walk.go:127`. That nested walk fails with `ErrMouthGate` (cap 3), the outer loop does `continue` (only `ErrLevelChanged` and a cancelled context return), and at `walk.go:124` `mouthSeq != n.walkSeq` is true again, so it steps away again, 3 more rewrites, held, and so on at about 7 ms of wall time per round until the outer 2 minute timeout (173 per run). Caller chain on every `nav walk` line: `nav.walk < nav.GoTo < combat.Fighter.move < combat.Fighter.engage`, one GoTo call per episode. The DRILLS19 reading (fresh walks by the caller) was wrong; fixmouthloop's design and offline test (a mover that returns ErrMouthGate) assume it. Fix direction (not done): `walk.go:124` should compare against a walk sequence captured before the nested walk (or the step-away should return `ErrMouthGate` / set a per-walk flag), then the Fighter back-off would take over. U whether the Fighter then resolves the fight.

### What was done to the character
Drilly (throwaway Sorceress, `--no-save`): drill cheats only (`debug goto`, `debug stat`), as DRILLS19. Last run (D2) ends in the Rogue Encampment, alive, panels closed. No process of mine left.

### Shared-file lines for the merger
- docs/BACKLOG.md: FIXMOUTHLOOP2: fixmouthloop FAILED live (DRILLS20): counts equal DRILLS19 (129 step-aways, 117 holds, 173 timeouts per DW.m3 run), 0 Fighter back-off lines. The loop is inside one Navigator.walk: walk.go:74 bumps walkSeq in the nested step-away walk, so walk.go:124 re-triggers the step-away after every failed one. Fix in nav (capture the sequence before the nested walk), keep the Fighter back-off.
- docs/INDEX.md: DRILLS20 section in docs/probes/drills/README.md.
- CHANGELOG: nothing (fixmouthloop not merge-ready).

## DRILLS21 (2026-11-01): live proof of fixmouthloop2 (PARTIAL: the storm halves, a roam loop remains)

Run on the worker host, hidden slot 0, lab build cheats3, `d2drill` built from branch agent/drills21 (agent/fixmouthloop2 tip 293c816 plus origin/main merged). Capacity before: idle 47 plus nice 0.7 percent. Files `2026-11-01-drills21-<label>.txt`: `m31` (DW.m3, filtered, repeats collapsed), `m32`/`m33` (first 250 filtered lines; the three invocations are identical, line for line in count), `m41` (DW.m4, first 400), `d3`, `d2` (DRILL lines). Command as DRILLS19 (`-group F -only DW.m3 -label drills21-m3N -v`, `-only DW.m4`; groups D3 and D2 without `-only`). Each invocation took 3 to 7 s of wall time. Each DW.m3 invocation runs the drill twice: run 1 FAIL (one level change), run 2 PASS, verdict DIFFER (DRILLS19/20: identical runs).

### Result: PARTIAL
| Item | DRILLS19/20 | DRILLS21 (fixmouthloop2) |
|---|---|---|
| Mouth bounces | 0 | 0 (V): no `arrived in level` from a gate click |
| `level changed during the walk` | 0 | 22 per invocation (V), all from ONE pocket-stall town-portal escape (`nav: stalled in a pocket, escaping by town portal`, level 8 pos (102,183), 74 from the exit, run 1, explore phase of cycle 3) plus 21 stale retries of the same error in `Fighter.move` over 30 ms. Not a mouth event (DRILLS18 saw the same escape); the drill verdict counts it as a bounce |
| `click rewritten` / `click held` | 351 / 117 | 154 / 51 (both runs of the drill together) |
| `stepping away from a cave mouth` | 129 | 62 (about 20 wanted; not met) |
| `walk did not finish in 2m0s` | 173 | 0 (DW.m4: 30, all after the character died, see below) |
| `combat move held by the mouth gate, backing off` | 0 | 1 (V): holds=1, backoff 3 s, long 30 s, to (37,150) |
| `combat swing held` / `combat move skipped` | 16 / 7 | 24 / 14 |
| Kills | 0 | 0 in every fight summary: the throwaway has no attack skill (`swinging the basic attack`), so nothing here shows whether a fight ends |
| Den cycle time | burned the phase limit | U: wall time is not comparable (lockstep game clock, 3 to 7 s per invocation in both); the cycle no longer ends by `walk did not finish` timeouts, and Clear and Explore finish in every cycle (every trace phase `ok`) |
| DW.m4 | 0 bounces, 4 step-aways | 0 bounces, 1 rewrite, 4 step-aways; 341 `character is dead` lines (the character dies in cave level 1 as before), 30 walk timeouts after that; dirty-UI verdict as before |
| Regression | D3 12/12, D2 11/11 | D3 12/12, D2 11/11; D2 ends life 34/40, D3 40/40, in town (V) |

### What the fix did (V, m31)
The Fighter path now works: the DRILLS19 chain (Den, step-away blocked, 3 rewrites then `click held`) ends in `combat move held by the mouth gate, backing off` (once, at (36,140)), and the nested `ErrMouthGate` no longer re-arms the step-away inside one walk. No 2 minute timeout remains. The count of step-aways fell from 129 to 62.

### The remaining repeats: a roam loop (V, m31 lines 1429 to 1585, 4583 to 4739, 4896 to 5057, 5227 to 5383, 5540 to 5696)
Five episodes of ten step-aways each at the same spot (36,140), 5.1 from the exit. Chain on every `nav walk` line: `nav.walk < nav.roam < nav.explore < nav.Explore < DenPhase`. The step-away walk is rewritten 3 times and held (`click held ... cap=3`), the outer roam leg fails with `roam leg failed ... no path: the click would cross the trigger area of an unwanted exit`, roam.go:222 treats that as `ErrNoPath` (it marks the target failed and picks another one), and the next target starts the same way: 50 `roam leg failed` lines in all, exactly 10 per episode, so roam stops after about 10 failed targets and Explore ends. The loop is bounded (about 8 ms of wall time per leg) and no longer burns the phase, but each failed leg costs a step-away. Fix direction (not done): let roam return `ErrMouthGate` (or stop after the first one) instead of folding it into the failed-target case, or pick the next roam target away from the mouth. U whether that frees a fight in place.

### What was done to the character
Drilly (throwaway Sorceress, `--no-save`): drill cheats only (`debug goto`, `debug stat`), as DRILLS19. Last run (D3) ends in the Rogue Encampment, alive (40/40), panels closed. Cleanup: `tools/agentreap.sh drills21` and `--list` run at the end (see the report).

### Shared-file lines for the merger
- docs/BACKLOG.md: FIXMOUTHLOOP3: fixmouthloop2 PARTIAL live (DRILLS21): 0 mouth bounces, 0 `walk did not finish` (was 173), the Fighter back-off line fires, step-aways 129 to 62 per DW.m3 invocation (target about 20). Remaining: Navigator.roam folds ErrMouthGate into the failed-target case (roam.go:222), 10 failed legs and 10 step-aways per episode in Explore; and one pocket-stall town-portal escape (22 level-change errors, 21 stale Fighter.move retries) makes the drill verdict FAIL in run 1. Merge-ready as an improvement; the roam fix is a follow-up.
- docs/INDEX.md: DRILLS21 section in docs/probes/drills/README.md.
- CHANGELOG: fixmouthloop2 (one step-away per walk) verified live: no walk timeouts, back-off runs.

## DRILLS22 (2026-11-01): live proof of fixtelekey (PASS, with findings)

Run on the worker host, hidden slot 0, `lab-debug-skills` binary (D2 and D3 and DW.m3 on `lab-cheats3`), `d2drill` built from branch agent/drills22 (agent/fixtelekey tip f16e229 plus origin/main merged at 45dd1b1). Capacity before: idle 42 plus nice 18 percent. New group `T` (internal/drill/d22.go, cmd/d2drill/d22_lab.go; overlap declared: the group switch and the end-of-run goto town in cmd/d2drill/main.go). Files `2026-11-01-drills22-<label>.txt`: `t` (group T, the proof), `town` (D22-T and D22-R after the town leg went through a tile exit), `d2`, `d3` (regression), `m3` (DW.m3, filtered to the nav, mouth, level-change and DRILL lines, repeats collapsed). Command: `D2D_AGENT=drills22 D2D_BIN=<fork>/bin/d2d.lab-debug-skills tools/agentrun.sh drills22 --ttl 500 -- tools/d2dslot.sh run -n 0 -t 400 -- <built>/d2drill -group T -repeat 1 -label drills22-t -v`. The navigator under test is the lab's own (`Backend.Navigator()`, with the combat rig attached); `wireTeleport` is in package main of cmd/d2bot, so `d22wire` repeats its four lines (`TeleportKeyFn = Fighter.TeleportKey`, `TeleportOK` from `data.CanTeleportIn` and `TeleportRule`): the d2bot call site itself is NOT exercised (U). Cheats (ledgered, actor `drills:<drill id>`): `debug goto`, `debug kill`, `debug stat 12 18` (clvl 18), `debug stat 9` (max mana 150), `debug stat 6/7` (life 2000), `debug points 54 1` (Teleport level 1), `debug mana full`. Hops are counted from the navigator's own log lines (`nav teleport` at Debug, `nav teleport did not move`). Every leg is the bot's `Navigator.ExitTo`, started from the same cell (the arrival of `debug goto` from Lut Gholein, V: identical start in both passes).

### Result (V, from `2026-11-01-drills22-t.txt`)
| Item | Drill | Verdict | Numbers |
|---|---|---|---|
| 1 control | D22-C | PASS | no Teleport learned: key empty, 0 hops, 0 errors; ticks Rocky Waste to Stony Tomb 254, Dry Hills to Halls of the Dead 1 578, border Rocky Waste to Dry Hills 708 |
| 1 Teleport | D22-A | PASS | key sk54, 12 hops landed, 0 refused, 0 errors: Rocky Waste to Stony Tomb 7 hops, 448 ticks, mana 150.0 to 3.0; Dry Hills to Halls of the Dead 1 5 hops, 851 ticks, mana 150.0 to 69.9; the border leg 0 hops, 687 ticks (708 walking) |
| 2 bind | D22-B | PASS | `TeleportKey()` = `sk54`, `HasTeleport()` true, player skills `36:1 54:1`; a cast through that key at a free cell 12 cells away moved 11.9 cells and paid 23.3 mana. `debug states` has no bind field (buffs= empty): the bind readback is the Fighter and the input map plus the effect of the cast |
| 3 boss | D22-T (part) | not run | `debug goto 74` reaches the Arcane Sanctuary (area 74, `TeleportOK(74)` true, HasTeleport true), but the Summoner stand point is a boss-runner path (cmd/d2bot bossKill); no boss fight forced |
| 3 rules | D22-T | PASS | `TeleportOK`: Lut Gholein false, Rocky Waste true, Duriel's Lair (73) false (clear-line level left to walking), Arcane Sanctuary true. A leg Lut Gholein to Sewers level 1 (a tile exit, so the hop code is reached) with Teleport learned: 0 hops, 0 refused, 191 ticks |
| 4 refused bound | D22-R | PASS | the unwired navigator (key but no level rule) in the same town leg: 3 refused hops (`did not move`, equal to `teleportBadHops`), 0 landed, 0 mana paid (d2d refuses in town for free, DRILLS15 T1), then the leg walked on and arrived (235 ticks): no spam |
| 5 other class | not run | | the harness has one throwaway class; the no-Teleport control (item 1) is the same code path, and `internal/combat` tests cover a class without the skill (key empty) |
| 6 regression | `d2`, `d3`, `m3` | PASS / PASS / as baseline | D2 11 of 11, D3 12 of 12 (`END ... in town: true`). DW.m3 with the default 2 runs: verdict DIFFER, run 1 one bounce in the explore phase (a pocket stall escape, level 8 to 1), run 2 pass: the same shape as DRILLS21 (PARTIAL). The DW.m3 character has no Teleport (40 mana), so nothing of fixtelekey runs there; the gate "0 bounces" is NOT met, as on main before this card |

### Findings
1. Teleport legs were NOT faster than walking (V). Same start, same exit: Stony Tomb 448 ticks with 7 hops against 254 walking; Halls of the Dead 1 851 against 578. A hop moves about 17 to 18 subtiles (3.4 cells; the navigator's `defaultTeleportRange` 18) and pays about 21 to 24 mana (23.3 in D22-B), so 7 hops drained 147 of 150 mana and the navigator then waited for regeneration (`waitMana`, no `Refill` installed in this rig). d2d would land hops of 15 to 24 cells (DRILLS15 T2), so the cap is the navigator's range, not d2d. At clvl 18 with level 1 Teleport and a pool of 150 mana the claim "hops fire" holds, the claim "hops save time" does not.
2. Outdoor border exits never hop (S exit.go: `hasTeleport() && !g.border`; V: border leg 0 hops, 687 against 708 ticks). On d2d the whole level map is known on arrival, so Explore finds no frontier either (the first runs of the rig, Rocky Waste to Dry Hills, hopped 0 times). So in Act 2 the card changes only tile exits (caves, stairs) and levels with unexplored frontier.
3. The d2d deviations (DRILLS15) did not show: 0 refused hops on 12, no landing on a monster (the legs were cleared with `debug kill`), the merc was not in play. The mana paid per hop is real on d2d and game.exe alike (the bind cast cost 23.3).
4. Harness: the arrival cell of `debug goto` depends on the level the character came from (Rocky Waste arrivals (155,213), (171,229), (165,220) in different runs); (165,220) is a pocket that traps a walk (control and Teleport both timed out there). The drills arrive from Lut Gholein each time. `debug warp` lands several subtiles off the asked cell.

### What was done to the character
Drilly (throwaway Sorceress, `--no-save`, a fresh one per invocation): the cheats listed above on levels 40 to 42 and 74, `debug points 54 1`. The last invocation (group T, D22-B) ends in the Rogue Encampment, life 2000 of 2000, panels closed. Nothing left running (`tools/agentreap.sh drills22`: nothing registered).

### Shared-file lines for the merger
- docs/BACKLOG.md: TELETUNE (optional, from DRILLS22): fixtelekey proven live (hops fire at clvl 18 in Act 2 on tile exits, key sk54, no hop in town, refused hops bounded at 3, no-Teleport control walks) but Teleport is slower than walking at level 1 (448 against 254 and 851 against 578 ticks): 18 subtile hops cost about 21 to 24 mana each and the navigator waits for regeneration; consider a mana budget (hop only when the walk is longer than N hops or mana is above a floor), a larger range on d2d, and note border exits never hop.
- docs/INDEX.md: DRILLS22 section in docs/probes/drills/README.md.
- CHANGELOG: nothing (harness and docs only); fixtelekey verified live (see its card).

## DRILLS23 (2026-11-01): live proof of roamleg (PASS, with a finding on the Explore end)

Run on the worker host, hidden slot 0, lab build cheats3, `d2drill` built from branch agent/drills23 (agent/roamleg tip 9859164 plus origin/main merged). Capacity before: idle 48 plus nice 0.1 percent (iowait 49). Files `2026-11-01-drills23-<label>.txt`: `m31`..`m33` (DW.m3, three invocations; filtered: first 500 matching lines, then the roam, held, DRILL and END lines; raw logs were 5,635 lines each), `m41` (DW.m4), `d3`, `d2` (DRILL lines). Command as DRILLS21 (`-group F -only DW.m3 -label drills23-m3N -v`, `-only DW.m4`; groups D3 and D2 without `-only`). Each invocation took 1 to 7 s of wall time. The three DW.m3 invocations agree on every count below (m31, m32, m33).

### Result: PASS
| Item | DRILLS21 | DRILLS23 (roamleg) |
|---|---|---|
| Mouth bounces | 0 | 0 (V): every `arrived in level` line is a planned entry of a cycle (2 or 8) |
| `level changed during the walk` | 22 (one pocket-stall escape) | 0 (V); no `stalled in a pocket` escape in these runs, both drill runs PASS and are identical (verdict `same`) |
| `roam leg failed` | 50 (16 in the filtered file) | 0 (V) |
| `roam: held by the mouth gate, ending the episode` | not present | 1 per invocation (V), at (36,140) after 83 legs, `held=1` |
| `stepping away from a cave mouth` | 62 | 11 per invocation (V; target about 20 or fewer, met) |
| `click rewritten` / `click held` | 154 / 51 | 6 / 2 |
| `combat move held by the mouth gate` | 1 | 1 |
| `nav plan` lines | 559 to 664 | 425 per invocation |
| Clear and Explore finish | yes, every phase `ok` | yes: 15 of 16 phases `ok`; the 16th (cycle 2 explore) ends with the new result, `nav: nothing left to explore: roam leg held by the mouth gate`, which the drill records as `error:` text but does not count as a bounce |
| DW.m4 | 0 bounces, 4 step-aways, 341 dead lines | 0 bounces, 4 step-aways, 0 roam lines, 341 `character is dead` lines, 30 walk timeouts after the death, dirty-UI verdict (`open: inventory`): unchanged from DRILLS21 |
| Regression | D3 12/12, D2 11/11 | D3 12/12 (ends 40/40 in town), D2 11/11 (ends 34/40 in town) (V) |

### Sweep completeness (U)
The first hold ends the episode after 83 legs (the first roam in cycle 2) instead of after about 10 failed targets: the sweep is not shortened by the old 10-target rule, but the hold now stops the episode at the first leg that the gate refuses, so targets beyond that point are not visited in that episode (the next episode retries them). The throwaway has no attack skill and kills nothing (`kills=0` in every fight line), so the drill cannot show whether the Den is cleared; visited coverage is not printed by the logs. U: whether the Den is cleared with a real character. Plan lines fell from 559 to 664 down to 425.

### Finding: `walk did not finish in 2m0s` returned (50 in m31, 30 in m41)
DRILLS21 had 0 of these in DW.m3. Here 50 `combat move failed ... walk did not finish in 2m0s` lines (m31 lines 5367 to 5490, the last cycle, after the roam hold) at (98,189), 83 from the exit, far from the mouth. Chain: `nav stalled by monsters, yielding` (near=6), `nav stuck`, a fight at 29 percent life with `combat: left the game to survive`; no mouth gate line is involved (V from the log). It looks like a stalled walk among six monsters with a throwaway that cannot kill (the DRILLS19 clock was lockstep, so the 2 minutes are game time), not a roamleg effect, but it was absent in DRILLS21, so U: whether the shorter episode leaves the character in that fight more often. The run ends with life 591 of 2000, not in town (`END ... in town: false`; the harness's final goto town is not reached when the phase ends this way); D3 and D2 afterwards were fresh characters.

### What was done to the character
Drilly (throwaway Sorceress, `--no-save`): drill cheats only (`debug goto 2`, `debug stat`), as DRILLS21. Last run (D2) ends in the Rogue Encampment, alive (34/40), panels closed. Cleanup: `tools/agentreap.sh drills23` and `--list` run at the end.

### Shared-file lines for the merger
- docs/BACKLOG.md: ROAMLEG live proof PASSED (DRILLS23): DW.m3 x3 0 bounces, 0 `roam leg failed` (was 50), 1 `roam: held by the mouth gate` per run, step-aways 62 to 11, plan lines 559-664 to 425, D3 12/12, D2 11/11. Open: 50 `walk did not finish` in the last Den cycle (monsters stall, character at 29 percent) absent in DRILLS21; Den clear completeness needs a real character.
- docs/INDEX.md: DRILLS23 section in docs/probes/drills/README.md.
- CHANGELOG: roamleg verified live: a mouth-gate hold ends the roam episode in one leg, no failed targets, step-aways 62 to 11.
