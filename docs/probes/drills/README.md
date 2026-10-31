# Live primitive drills (card m1-liveproof, harness cmd/d2drill + internal/drill)

> status: D0, D1 (rerun after the primitive fixes) and D2 (portals, partly) run live (2026-10-31). Results: `2026-10-31-d1.txt` (first run, before the fixes), `2026-10-31-d1rerun.txt` (after fix-belt and fix-pickup), `2026-10-31-d2.txt`; each `.json` next to it. Findings of the D2 run: section W below.

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
| item into pack or belt directly | MISSING: `debug give <code> [q]` only drops on the ground at the feet (V); pickup then goes belt first, pack when the belt is full | `debug give <code> pack\|belt [box]` |
| gold | EXISTS: `debug stat 14 N` (purse), `15` (stash) (V live) | none |
| level, life, mana | EXISTS: `debug stat 12\|6\|8` (6 and 8 take 24.8 fixed point, read one tick later) (V live) | none |
| durability | PARTIAL: `debug wear` halves every worn item (S, source); no arbitrary value, a live read after it was not conclusive | `debug dur <slot> <cur>` |
| kill, hire or revive the merc | MISSING: no `debug merc`/`hire` (V: unknown verb); `debug die` kills the player only | `debug merc hire\|kill\|revive` |
| belt removal | MISSING: no verb takes a belt off; drills wear a Sash (`lbl`) and drain with `cmd belt` | `debug equip none 8` |

BLOCKED drills: D1.P3m (merc drink, needs merc hire/kill/revive), D2.d (cut cast, needs `debug hit`), D2.f (round trip, needs a working portal walk-in).

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
