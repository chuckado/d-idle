# Live primitive drills (card m1-liveproof, harness cmd/d2drill + internal/drill)

> status: D0 harness and D1 belt group run live (2026-10-31). Results: `2026-10-31-d1.txt` and `.json`; `-run2` is a second process (identical lines).

Run on the worker host, one hidden slot, lab build, throwaway Sorceress, `--no-save`, lockstep:
`D2D_AGENT=drills tools/agentrun.sh drills --ttl 3600 -- tools/d2dslot.sh run -t 3600 -- d2drill -group D1 [-only D1.P1] [-repeat 2] [-v]`.
Groups: `D1` belt, `D0` gate self-test (leaves the pack open on purpose, expects a P35 FAIL line and a recovery).
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

BLOCKED drills: D1.P1s (shift-click, needs the modifier verb), D1.P3m (merc drink, needs merc hire/kill/revive).
