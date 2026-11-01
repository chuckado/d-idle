# FORKDBG: lab-only debug verbs for the live skill drills (card forkdbg, 2026-11-01)

> status: current. The verbs exist in the lab d2d binary built from the private fork branch `lab-debug-skills` (patches never live in this repo, GPL, docs/decisions.md). Older verbs: docs/cheats.md section "d2d debug verbs", docs/forkcheats.md. Never part of an upstream pull request without the user's approval (standing approval 2026-10-31: cheats or functionality in the private fork when documented and vetted).

Why: the SKMISS drill recipe (docs/research/sk-missile-trace.md section 8, M1-M5) listed four missing verbs. This card adds them and the client calls; it does not run the drills.

## Fork
- Branch `lab-debug-skills` on the private remote (the name `lab/debug-skills` is not possible: the branch `lab` exists, git refuses a ref below it). Built from the `lab-cheats3` branch (all earlier lab verbs, `debug missiles`).
- Commit `d9b96c2ba67b6c4772b62db2f7702876f464b323` (pushed to the private remote the same day, before this doc cites it): new `apps/d2d/devctl/debug_skills.cpp`, the `states` verb moved out of `debug_fight.cpp` into it (no behaviour change for a bare `debug states`), the `help` list and usage line, `CMakeLists.txt`.
- Lab binary: `bin/d2d.lab-debug-skills` on the worker host (Release, `D2D_WERROR=OFF` like the other lab builds: GCC 16 rejects `ui.cpp` under -Werror, not touched here). Pass it as `D2D_BIN`. Not a replacement for the cheats3 binary other agents use.
- Fork-side tests: none added. The fork's tests are headless-with-game-data ctest scripts that need the game archives and a loaded game; the verbs were vetted live instead (below).

## Verbs
All four refuse with `err refused: not an offline --no-save game` unless the character store is off (`--no-save`) and no host is joined. Every reply says what changed. They are cheats or test actions: use them only through `d2dctl.SkillDebug` (below) or `d2drill -probe`, which write the cheat ledger first.

| Verb | Forms | Effect and reply |
|---|---|---|
| `debug mana full` | no other argument | the lead player's mana := max mana. `ok mana 5 -> 35 (max 35)`. Other words: usage line. |
| `debug states set <skill> [level]` | skill 0..1000 with a state (`aurastate` in Skills.txt; Frozen Armor is 40, States.txt row 10), level 1..99 (default: the skill's level, at least 1) | the player's self state exactly as a cast starts it (`Fight::start_state`: duration from auralen, Bone Armor's absorb pool). `ok state frozenarmor of skill 40 Frozen Armor lvl=3`. Errors: `no such skill, or it has no state`, `level 1..99`. |
| `debug states clear <skill>\|all` | | removes that state (`ok cleared state of skill 40`, `err skill not in states`) or all (`ok cleared 1 states`); Bone Armor's pool goes to 0. |
| `debug states` | no argument | unchanged read-only listing: `absorb= maxlife= maxmana= def= buffs=<skill>:<level>,` and cursed monsters. |
| `debug mon set <monster id> [life <n>] [frac <0..255>] [maxlife <n>] [dmg <lo> <hi>] [res <phys\|magic\|fire\|light\|cold\|poison> <-100..100>]...` | the id is the `#id` of `monsters`; at least one change; any order; `res` may repeat | `life`/`maxlife` points (1..1000000; a `life` above the max raises the max; the edit does not trigger the life-dropped effects), `frac` the 256ths already taken off life, `dmg` the A1, A2 and S1 range (0..100000), `res` the ABSOLUTE resistance (the mod bonus is set to make it; 100 is immune). `ok #24 Fallen maxlife 4 -> 50 life 4 -> 50 dmg 5-5 res.fire 100`. Errors: `no live monster with that id`, one `err <field> range` line per bad value, `unknown token`. Lowering `maxlife` below the current life does not lower life: set both. |
| `debug cast <skill> <x> <y>` / `unit <monster id>` / `near` | skill 0..1000 with a level (`debug points <id> <n>`), x y in cells (\|v\| <= 4096), `near` the nearest targetable monster | queues `cmd::UseSkill` (the same message `cmd skill` and a click send), so the start check, mana at the action frame, the get-hit interrupt and the missiles are the real path. `ok queued skill=36 Fire Bolt lvl=1 at=30.20,46.70 unit=24 mana=35` (mana in points before the cast). Errors: `no such skill`, `skill has no level (debug points <id> <n>)`, `no such monster`, `x y: cells, |v| <= 4096`. The cast starts at the next tick: `step`. |

`debug help` lists all of them. Not in the verbs, on purpose: no verb sets a monster's mana or AI, and nothing changes any state outside the lab game.

## Client and ledger (`internal/d2dctl`)
`d2dctl.SkillDebug{C, Ledger, Actor, Instance, Character}` has `ManaFull`, `StatesSet`, `StatesClear` (skill -1 = all), `MonSet` (`MonEdit{Life, LifeFrac, MaxLife, Dmg, Res}`), `CastAt`, `CastAtUnit`, `CastNear`. Rules, all tested against a fake socket (`internal/d2dctl/debug_skills_test.go`):
- Every call needs a non-empty reason, an actor, a client and a ledger. It validates the ranges above first (a bad call sends nothing and writes nothing).
- The ledger entry (kind `d2d-debug`, params `id` and `verb`, the actor, the reason) is written BEFORE the verb is sent; an unwritable ledger stops the call. A second line follows with success and the reply (`after.reply`) or the error, also for a server refusal.
- `debug cast` is marked `test_action: true` in the params: it changes no state itself but spends mana and starts a fight; filter it apart from the state cheats when needed.
- A character that appears in the ledger never counts toward a gate (docs/cheats.md).
Suggested reasons: `M1: known mana for the action-frame read`, `M1: Frozen Armor state for the recast read`, `M4: monster at 6 life, no lightning resist`, `M1: cast through the normal path`.

## Vetting (live, one hidden slot, lockstep, fresh throwaway Sorceress `--no-save`, `d2drill -probe`, every call ledgered with the probe reason)
Method: `D2D_AGENT=forkdbg D2D_BIN=<lab-debug-skills binary> tools/agentrun.sh forkdbg --ttl 600 -- tools/d2dslot.sh run -n 2 -- d2drill -group D1 -bin $D2D_BIN -probe '<verbs>'` on the worker host (idle+nice 66 percent beforehand; `tools/capacity.sh` was not used). Blood Moor (`debug goto 2`), `dump stats` raw[8] is mana in 256ths, raw[6] life. Lockstep is deterministic: the spawned monster was `#24` in every run.

| Verb | Observed |
|---|---|
| `mana full` | `debug stat 8 1280` then `dump stats` raw[8] 1286 (5 points); `debug mana full` replied `mana 5 -> 35 (max 35)`, `dump stats` raw[8] 8960. |
| `states set/clear` | `debug states` buffs empty, def=6; `states set 40` -> `state frozenarmor of skill 40 Frozen Armor lvl=1`; `debug states` `def=7 buffs=40:1,`; `states set 40 3` -> `buffs=40:3,`; `states clear 40` -> def=6, buffs empty; a second clear `err skill not in states`; `set 36` `err no such skill, or it has no state`; `set 40 100` `err level 1..99`; `set 40 7` then `clear all` -> `cleared 1 states`. |
| `mon set` | spawned `fallen1` (#24, 4/4): `mon set 24 life 50 maxlife 50 res fire 100 res light -100 res cold 0` replied `... maxlife 4 -> 50 life 4 -> 50 res.fire 100 res.light -100 res.cold 0`, `monsters` showed 50/50. Resist A/B (separate runs, one Fire Bolt, same seed): fire 100 stayed 50/50, fire 0 went 45/50. `dmg 5 5`: the player's life 40 -> 35 after the Fallen's swing (raw[6] 10240 -> 8960); without it 38. `life 70` / `maxlife 30` / `life 6 frac 128` -> `monsters` 6/30. Negatives: unknown id, `life 0`, no field, `res foo 5`, `dmg 9 3`, `frac 300` all errors with the messages above. |
| `cast` | `cast 36 near`: `queued skill=36 Fire Bolt lvl=1 at=30.20,46.70 unit=24 mana=35`; 14 ticks later `debug missiles` listed a `firebolt` (owner 0, speed 4.688) and mana was 8354 raw (32.6 points: cost paid; regen is on); `cast 36 unit 24` hit the monster (A/B above). Negatives: `unit 9999` `err no such monster`, skill 999 `err no such skill`, `cast 37 near` `err skill has no level (debug points <id> <n>)`, `cast 36 12 x` `err x y: cells, |v| <= 4096`. A fresh Sorceress already reported Fire Bolt `lvl=1` before any `debug points` (the build gives it; not investigated). |
| not tested | the refusal outside an offline game (the lab is always `--no-save`; the guard is `town.world.characters` and `town.net_game`, by code reading); `debug cast` of a charged-item skill (no item argument: `-1`); co-op (only the lead player is touched). |

Clean-up: the probes ran under `tools/agentrun.sh`; `tools/agentreap.sh forkdbg` reported nothing registered and `tools/agentreap.sh --list` was empty; slots 0..3 free. No game.exe was started and no character exists after the probes (`--no-save`).

## The M1-M5 recipe, runnable now
Use one hidden slot, a throwaway Sorceress, `debug goto 2` (a field level), the lab binary above. Call the verbs through `SkillDebug` (or `d2drill -probe`) so they are ledgered; read with `dump stats` (mana raw[8], 256ths), `monsters`, `debug missiles` (one row per missile), `step <n>`.
- M1 (mana timing): `debug points 36 1`, `debug mana full` or `debug stat 8 <256 x n>` for a known mana, `debug cast 36 <far x> <far y>`, then `step 1` and `dump stats` every tick: game.exe drops 640 on the action-frame tick, d2d on the command tick. Interrupted cast: `debug spawn fallen1 1` next to her, `debug mon set <id> dmg <big> <big>` (more than max life / 12 for a get-hit; or `debug hit <dmg>`), cast, and watch whether mana is spent without a missile.
- M2 (Charged Bolt count): `debug points 38 1|5|20`, `debug cast 38 <x> <y>`, `debug missiles` on the action frame: 3, 7, 22 rows, start = the caster's subtile, all gone by 77 frames.
- M3 (reach): `debug cast` Fire Bolt (36) / Ice Bolt / Lightning at a point 60 subtiles away, `debug missiles` every frame, last position and frame count.
- M4 (roll): `debug mon set <id> life 6 res light 0` (add `maxlife 6`), Charged Bolt 1, many single casts after `debug mana full`; refill life with `debug mon set` again each round. Statistical only.
- M5 (Lightning pass-through, Chain Lightning hops): `debug spawn` two monsters in a line, `debug mon set` each to a large `life`, `debug cast <skill> unit <id>`, `debug missiles` and `monsters` life before and after.
- Armor recast and buff reads: `debug states set 40` (Frozen Armor) then `debug states`, a recast cast, `debug states clear 40`.
