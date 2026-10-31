# Live primitive drills (card m1-liveproof, harness cmd/d2drill + internal/drill)

> status: D0 to D9 (except as listed BLOCKED), D2.d and the hover audit H ran live (2026-10-31, DRILLS3; see the section at the end). Earlier: D0, D1 (rerun after the primitive fixes) and D2 (portals, partly) run live (2026-10-31). Results: `2026-10-31-d1.txt` (first run, before the fixes), `2026-10-31-d1rerun.txt` (after fix-belt and fix-pickup), `2026-10-31-d2.txt`; each `.json` next to it. Findings of the D2 run: section W below.

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

BLOCKED drills in the last runs: D1.P3m, D2.d and DW.a3 were blocked on the cheats now in the lab binary (see the table; docs/forkcheats.md; rerun pending), D2.f (round trip) still needs a working portal walk-in.

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
New missing pieces (exact): a monster to fight (`debug spawn`: D5.c, D5.d, H.monster), `debug merc hurt <dmg>` (D8.m, D1.P3m), a cheat that sets quest records (Blood Raven to 4, Den done with reward pending, Andariel: D8.c, D8.q, D8.w; `debug quest` sets bit 0 only), an unidentified magic item (D3.i), a way to pin the character in a pocket (D6.d). Resolved since the table above: `debug durability`, `debug hit`, `debug merc`, `debug give pack|belt` (used by D2.d, D3.r, D8.a).
Findings for the fix queue (details with file and line in hover-audit.md): F1 gold deposit impossible on d2d (D3.s, D3.s2: `hoverFirst` hides `TypeChar`, the prompt needs digits even for the whole purse); F2 field objects cannot be operated (H.obj: `Interact` unsupported for objects, the click fallback is a walk, d2dsrc lists no field objects); F3 `d2dctl.Pause` is a wall sleep, so the merc claim and revive paths cannot work under lockstep (H.pause); F4 `needRepair` ignores gold (D3.r0); F5 the purse caps at 10000 at clvl 1 and the pickit does not know it (D4.cap).
