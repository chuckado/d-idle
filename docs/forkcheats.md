# FORKCHEATS: test cheats added to the lab binary (card forkcheats, 2026-10-31)

> status: current. The verbs exist in the lab d2d binary as of 2026-10-31 (the patches live in the private fork, not in this repo). Argument forms and rules: docs/cheats.md section "d2d debug verbs".

User go 2026-10-31: add whatever is needed for testing. Each verb below was exercised live on one hidden slot (lockstep, fresh throwaway Sorceress, `--no-save`) through `d2drill -probe`, so every call is a ledgered `d2d-debug` cheat with a reason. State is read after `step 2` (the client copy that `dump` reads follows the world one tick later).

| Verb | Verdict | Observed live |
|---|---|---|
| `debug act <1-5>` | DONE | with the waypoint panel open, `dump ui` `wp_tabs` has 1 rect, then 2 after `debug act 2`, 5 after `debug act 5`, 1 after `debug act 1`; `debug act 7` is an error. Only the four quest flags the panel reads change (the same ones `waypoint_act_open` tests). |
| `debug hit <dmg>` | DONE (real path) | `debug hit 5` at full life 40: life 40 -> 35, player mode `GH` at once, back to idle after 60 steps. Routed through the game's own hurt path (`hurt_side`: life, armour wear, death at 0), then the get-hit mode is forced (a small blow is normally only a probable get-hit). Works in town. Not run: a hit that kills (death path is the existing one). |
| `debug merc hire <type>` | DONE | `hire 0` on a character without merc: `dump merc` present, code RG; `hire 12` code GU; an unknown type lists the known ids 0 to 29. No gold taken. |
| `debug merc kill` | DONE | the merc dies (life -1 after 3 steps) and after 120 steps `dump merc` is `present:false, dead:true`; killing twice is an error. |
| `debug merc revive` | DONE | after the kill: `present:true, dead:false`, life 40, beside the player; without a dead merc it is an error. No gold taken. |
| `debug give <code> pack` | DONE | pack empty: `hp1` went to cell 9,3 then 9,2 (the player's search order), `lea` (2x3) to 0,0, `cap` to 8,0; pack full gives `err no room` (by code reading). |
| `debug give <code> belt [box]` | DONE | with a light belt: `hp1` box 0, `mp1` box 1 (the pickup's `belt_slot` choice), `hp1 belt 7` box 7, a taken or out-of-range box answers `no belt box`. A box above an empty box is allowed. |
| `debug durability <slot\|all> <v>` | DONE | `durability 3 5` armour 13 -> 5; `1 9999` helm capped at its max 12; `all 1` three worn items at 1 (the staff included); `6 3` on an empty slot is an error. A hit afterwards did not change a 1-durability item in that run (wear is a roll). |
| `debug belt clear` | DONE | 4 belt potions removed (`removed=4`), belt boxes empty in `dump items`. |
| `debug unequip <slot\|belt>` | DONE | belt worn then removed, the second call errors; the item is gone (belt boxes stay: `belt clear` first). |
| gold (`debug stat 14 N`) | exists already | confirmed in the README table (V live). |
| `debug spawn` (a monster near the player) | NOT BUILT | no drill is blocked on it; `debug hit` covers the busy/hit-recovery drills. |

## Regression of the harness on the new binary
`d2drill` from main (repeat 1, same slot): D0 gives the expected P35 FAIL + recovery; D1 P7a P7 P1 P1s P2 P3 PASS and P3m BLOCKED (the drill still lacks a merc step); D2.a PASS (verb 32 ticks, one source spent). Nothing regressed.

## Drills unblocked (not rerun here)
- D2.d (cut cast): `debug hit` while the player casts.
- DW.a3 (cross-act waypoint rows): `debug act 2` (or higher) before opening the panel.
- D1.P3m (merc drink): `debug merc hire 0`, `debug merc kill`, `debug merc revive`.
- Repair and belt drills: `debug durability`, `debug give pack|belt`, `debug belt clear`, `debug unequip belt`.
- D2.f (round trip) stays blocked on the portal walk-in (W1 in docs/probes/drills/README.md), not on a cheat.

## Notes for the next drills
- `debug give <code> pack` skips the tome merge: a scroll given while a tome is in the pack stays a scroll.
- `debug hit` also wears armour like a monster blow does (a roll), and always flinches unless it kills.
- The merc verbs act on the lead player's merc; `merc hire` uses name 0 and experience 0 and replaces any merc present.
