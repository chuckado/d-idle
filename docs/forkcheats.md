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
| `debug spawn` (a monster near the player) | built in the second set (below) | |

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

# FORKCHEATS2: the second set (card forkcheats2, 2026-10-31)

> status: current. User standing approval 2026-10-31: cheats or functionality may be added to the private fork when documented and vetted, and never reach an upstream pull request without the user's approval. Argument forms, rules and negative cases: docs/cheats.md section "d2d debug verbs".

Same method as above: one hidden slot (lockstep, fresh throwaway Sorceress, `--no-save`), every call a ledgered `d2d-debug` cheat through `d2drill -probe`. Monsters are checked on Blood Moor (`debug goto 2`). State is read after `step 2`.

| Verb | Verdict | Observed live |
|---|---|---|
| `debug spawn <monster> [n] [near] [champion\|unique [mods]]` | DONE | `spawn zombie1 2`: ids #17,#18 on a ring 3 cells from the player, both listed by `monsters` (mode WL, hp 11/11, level 1). `spawn fallen1 1 champion`: "Champion Fallen", boss column `champion`, mod 16, 15 hp; `spawn fallen1 1 unique 5 6`: a named unique with mods 5,6, 36 hp. After `step 200` the lockstep clock answered (tick 297) and the zombies had walked toward the player; `debug kill` then `step 60` left every spawned monster in DD with hp 0 (killable). In town: `err no monster level here`. Negative: unknown name and row 99999 `err no such monster`, `n 0` and `n 99` `err n 1..20`, a stray token `err unknown token bogus`. Not shown: a spawned monster hurting the player (zombies are slow; 200 ticks did not reach). |
| `debug merc hurt <dmg>` | DONE (real path) | merc hired (`dump merc` life 40): `merc hurt 5` gave life 35, `dump merc` 35; `merc hurt 9999` gave mode DT, after 120 steps `dump merc` is `present:false, dead:true`. Negative: `0` and `abc` `err damage > 0`, no live merc (none hired, or already dead) `err no live merc`. |
| `debug quest list\|set\|bit\|join` | DONE | fresh character: `list` empty. `set 2 0x0002` (Blood Raven record, bit 1 = reward pending): `list` prints `2 0x0002 act1 bits=1`, `dump quests` byte 2 = 2; `bit 6 1 on` then `off` toggled the Andariel record; `set 8 0x0001` writes an act 2 record. With record 2 at 0x0002 the player talked to Kashya (`cmd interact`): the record became 0x0001 (reward taken, done), so the quest machine reads the written bits as the game's own. Negative: record 99 `err quest 0..47`, flags 0x1ffff `err flags 0..0xffff`, bit 16 `err bit 0..15`, `join 9` `err act 1..5`, a missing or non numeric argument the usage line. Not shown live: the Den record 1 reward path through Akara (same code path, record 1 set to 0x0002 and listed). |
| `debug give <code> <quality> [unid]` and `give pack\|belt ... q<n> [unid]` | DONE | `give cap pack q4 unid` into the pack: tooltip "Cap, Defense 4, Unidentified" (quality 4 in `items`); `q4` alone gives "Cap of Charged Bolt" identified; `give lea pack q6 unid`: an unidentified rare "Leather Armor"; `q4 l1` a magic cap at item level 1. On the ground `give cap 4` is already unidentified (the game's rule), `give cap 4 unid` the same, `give cap 0` a "Superior Cap" (quality by item level, identified), `give hp1 unid` an unidentified plain potion. Negative: `give cap 9` `err quality 0..8`, `pack q9` `err quality 1..8`, `pack bogus` `err unknown token bogus`, `belt` without a belt `err no belt box`. `debug unid` itself was not changed: it already unidentifies every pack item whatever its quality; what was missing was a magic or better item in the pack. |
| `debug pin on [r]\|off\|status` | DONE | at 28,48 in Blood Moor: before, `blocked 28.9 48.1` 0; `pin on` (r 4): 32 collision cells, the same cell 1, `blocked 27.1 48.1` and `28 49` 1. A walk to x=40 (`cmd move`, 150 steps) ended at 28.5, a walk to x=15 at 27.7, so the character stays in the pocket whichever way it goes. `pin off`: 32 cells freed, the cell 0, the walk to x=40 reaches 40.1. Negative: `pin on 2` and `9` `err radius 3..8 subtiles`, a second `pin on` `err already pinned (pin off first)`, `pin off` unpinned `err not pinned`. The pin stays on the level it was set on if the player leaves it (`pin off` still frees it). |
| `debug help` | DONE | prints one line per registered sub-verb (usage for the lab verbs, the verb name for the rest). |

## Regression of the harness on the new binary
`d2drill` from main (repeat 1, one slot): D0 gives the expected P35 FAIL + recovery; D1 P7a P7 P1 P1s P2 P3 PASS and P3m BLOCKED (the drill itself still lacks the merc step); D2.a PASS (verb 32 ticks, one source spent). The same as with the first set: nothing regressed.

## Drills unblocked (not rerun here)
- D5.c, D5.d (P27, P29), H.monster, the fight half of D5.b (mana drink and Meter totals) and the hit effect of D5.a (P26): `debug spawn`.
- D8.m and D1.P3m (merc drink): `debug merc hurt`.
- D8.c (Blood Raven claim), D8.q (Akara reward), D8.w (Andariel record): `debug quest set` (the act hop to Act 2 is still package main code, not a cheat).
- D3.i (P13 identify) and the unidentified sell prices of D3.g0: `debug give <code> pack q4 unid`.
- D6.d (P25 pocket): `debug pin on`.
- Still not a cheat gap: the gamble drill (D8.g) is a bot gap; D9.c is a soak.

## Notes for the next drills
- Quest word layout: 48 records of 16 bits; act 1 words 0..6 (1 Den, 2 Blood Raven, 3 tools, 4 Cain, 5 tower, 6 Andariel), act 2 words 7..14, act 3 15..22, act 4 23..25, act 5 26 on; bit 0 done, bit 1 reward pending. Setting a record re-runs that act's join, which resets the act's other quest machines from their flags.
- A spawned monster is the plain `make_monster` unit (no pack, no corpse loot rules beyond a normal kill); spawn on a level the player stands on only.
- `debug pin` blocks the walk grid only; a teleport or `debug warp` still leaves the pocket.

# FORKCHEATS3: `debug missiles` (card forkcheats3, 2026-10-31)

> status: current. `debug missiles` exists in the lab d2d binary as of 2026-10-31 (private fork, never in this repo, never in an upstream pull request without the user's go). The d2dctl parser for it is in fix-missile (`d2dctl.ParseMissile`: 10 leading fields, extra trailing fields are ignored). Format and negative cases: docs/cheats.md section "d2d debug verbs".

Read only: it lists `Fight::missiles` (the same list `state missiles=` counts). Same method as above: one hidden slot (lockstep, fresh throwaway Sorceress, `--no-save`), Blood Moor (`debug goto 2`), through `d2drill -probe`.

| Case | Verdict | Observed live |
|---|---|---|
| no missiles | DONE | `debug missiles` prints just `ok` (no rows); `state` `missiles=0`. |
| own bolt | DONE | `debug points 36 1`, `cmd skill 36 <x> <y>` (Fire Bolt, east): one row, missile id 58 `firebolt`, owner type 0, owner unit 1 (the player), dx 1.000 dy 0.002, speed 4.688 cells/s, friendly 1. x advanced 0.75 cells per 4 ticks (speed x 0.04 x 4), ttl 50, 46, 42 ... 2 falling one per tick; the row is gone after the range ends and `state missiles=` is 1 while listed, 0 after, at every sample. |
| hostile missile | DONE | `debug spawn fallenshaman1 4` and `quillrat1 6` beside the player: `shafire1` (id 22) rows with owner type 1, owner unit = the shaman's `#id` from `monsters` (26, 27), friendly 0, dx/dy pointing at the player (a unit vector), speed 1.875, ttl 40 falling; its hit effect `shamanexp` is listed too (speed 0, owner -1, ttl 10 falling). Player life fell from 40 to 3 in the run (the quill rats bite too). |
| negative | DONE | `debug missiles extra` answers `err usage: debug missiles`. 200 ticks with missiles in the air (steps of 2 with a listing each) kept the loop answering; the lockstep clock never stalled. |

## Regression of the harness on the new binary
`d2drill` from main (repeat 1, one slot): D0 gives the expected P35 FAIL + recovery; D1 P7a P7 P1 P1s P2 P3 PASS and P3m BLOCKED (as before: the drill still lacks a merc step); D2.a PASS. Nothing regressed.

## Notes for the bot
- dx dy are a unit vector (zero for a missile that does not move: a hit effect), speed is cells per second, x y are cells like `monsters`. unit is only the place in the list (a missile has no unit id in d2d): do not track a missile across calls by it.
- A monster missile's owner unit is the `#id` column of `monsters`; the player's own missile has owner type 0, owner unit the player's id (1 on a single game). Filter on `friendly` or owner type to drop the bot's own bolts, and on speed > 0 to drop hit effects.
- Live proof of the dodge: drill D5 with `debug spawn` of a ranged monster (`fallenshaman1`, `quillrat1`).
