# Paladin aura audit (2026-11-26)

> status: audit, read-only; no fix made. Question (owner): auras only work when selected on an active skill slot (left or right); is the Paladin scuffed because of that? Mechanics side: docs/sme/paladin-auras-2026-11-26.md (branch agent/smeaura). Tags: V verified (how), S read in code, U unverified.

## 1. Flow in the code (d2d backend, what runs in production)

1. Left skill: on d2d the bot never selects it. The fighter attacks by skill id (`cmd skill`), so `SetLeft` is skipped when the input is a `swinger` (internal/combat/fight.go:768-771, comment cites WALKSMOOTH3). LeftSkill is read (d2dsrc/player.go:187) but nothing sets it. S.
2. Right skill = the aura. `KeepAura` (aura.go:135-182) is called from the fight tick (fight.go:593, hostile target present, interval `combat.auraCheckMs` 1500 ms), from `Buff` (fighter.go:706, new game, back from town, death edges) and, as the resting aura Prayer, from the rest loop (resource.go:215).
3. Which aura: `wantAura` (aura.go:85) walks the profile ladder (classdata/profiles.json, Paladin zealot profile: Prayer if resting and life under 70 percent, Holy Fire needsCrowd with at least `combat.holyFireMinMonsters` 3 hostiles and a non-fire-immune target, Might, Thorns) and requires the skill to be learned (reachable = right skill already selected, no hotkeys on d2d).
4. How it is applied: `applyAura` on d2d calls `SetRight` (aura.go:191-199) which clicks the right-skill popup cell and reads `rskill` back (bindleft.go:189-216). `cmd select` does not change rskill (V live, docs/palready.md section 1). If `SetRight` finds `RightSkill == id` it returns at once without any click (bindleft.go:197).
5. How it verifies: `auraActive` / the 3-try loop ask `ActiveAura` (aura.go:30), which reads the player stat-list child chain, stat 350, through `e.Mem`. On d2d `e.Mem` is `noMem` (d2denv/env.go:112, misc.go:14-18) so every read fails, `ok=false`, and the code falls back to "`RightSkill == id` and not dirty" (aura.go:72-77, 173). The d2d dump does not carry `player.states` or `player.hotkeys` (Questy and Doofer boot line: unknown fields list). So the only evidence of an active aura is the selected right skill.
6. Re-selection triggers: `auraDirty` is set at construction (fighter.go:144), on death (fight.go:246), on any area change (aura.go:142, which includes the town portal both ways). Game restart = new process = dirty. Merc revive and merc state: nothing touches the aura (a merc aura is the merc's own). Failure: 3 tries then `auraRetryAt` backs off 10 s (aura.go:179-181) and the fight goes on with no aura and an Info line `aura did not come up`; with no popup rows (d2dwire.go:63-70) `SetRight` errors `ErrBindFailed` at every try and the bot plays WITHOUT an aura, logging only a Warn at boot plus `aura popup pick failed`. Neither line is a task failure or an alarm.

## 2. Live evidence (docker logs, last 12 h and 72 h, read-only)

| | Questy (Paladin clvl 12, plan Sacrifice) | Doofer (Paladin clvl 21) |
|---|---|---|
| `combat aura up` | 42 (41 Might, 1 Prayer) over about 10 h | 123 (62 Might, 61 Holy Fire) in the 72 h window (log starts at its boot, about 20 h) |
| `aura did not come up`, `popup pick failed`, `skill list rows unavailable` | 0, 0, 0 | 0, 0, 0 |
| attack skill at kills | Sacrifice 780, Zeal 890, Attack 21 | Zeal 729, Attack 3 |
| planner shadow | `aura=Might auraProfile=Might` | `aura=Might auraProfile=Holy Fire` in 58 of 58 lines (disagree) |
| merc | none (`merc=none` 108 zone entries) | Rogue (class 271) alive, level 19, `aura=-` |

Findings. (a) The popup pick works mechanically: rskill reads back as the wanted skill. (b) The logs do NOT show left or right skill ids anywhere (status, minute and zone-entry lines carry no skill fields; `skill selected` and `left skill set` are never logged on d2d). (c) Whether the aura is ACTIVE in fights is unknown from logs: `aura up tries=1` is also logged when no click happened (rskill already Might, `SetRight` no-op, fallback check passes). V that the right hand holds the aura; U that the aura state exists in the game.

## 3. The central gap (owner's suspicion, partly right)

Everything about d2d aura state rests on rskill. The hotkey path on game.exe was V live: a hotkey switch shows the aura in the stat-list chain within 20 ms (docs/classes/smoke-paladin.md finding 6, aura.go:12-16). The d2d popup path was verified only by reading rskill (docs/palready.md; `aura up` counted 4 in 2 min). Nobody has read the aura STATE on d2d after a popup pick. In the real game picking an aura in the menu may only arm the skill and need one cast (right click) to start the state: U, for the SME note. The bot never right-clicks (`cmd skill` casts at a point and leaves rskill unchanged, V palready), so if that holds the Paladin has been fighting without Might, Holy Fire, Prayer for the whole run, and the planner and the log both would say "aura up".

## 4. Ranked defects and gaps

1. HIGH, U: aura state never read on d2d; "aura up" is rskill only (aura.go:72-77, 173; d2denv/misc.go:14-18; d2dsrc/player.go:206 unknown `player.states`). A popup pick that does not start the state is invisible.
2. HIGH: "aura up" log is self-fulfilling when `SetRight` is a no-op (bindleft.go:197 then aura.go:173); also clears `auraDirty` after area change and death without any action, though the game may drop auras on death (U).
3. MED: no cast ever starts the aura on d2d. `applyAura` only selects (aura.go:191-199); a click or `cmd skill <aura>` is the missing step if section 3 holds.
4. MED: the plan hotkeys (F1 Might, F2 Holy Fire, F4 Zeal, ...; plandata/build-paladin-normal30-v2.json `hotkeys`) are "not executed by the leveler" and d2d has none, so the slot layout the owner describes (left Zeal, right aura) does not exist on d2d; left is whatever the character had (Attack).
5. MED: Doofer plan vs profile disagree, `aura=Might` vs `Holy Fire` (agency/planner.go:74-87); only shadow, but the 61 flips Might/Holy Fire are real popup picks each time the crowd count crosses 3.
6. LOW: failure is silent in the sense that no counter reaches the status line or minute line; `skill list rows unavailable` is a Warn at boot only (d2dwire.go:65-67).
7. LOW: merc aura is not read on d2d (`merc.aura` unknown); the Paladin merc aura policy (merc/policy.go:101-117: Defiance, Prayer, Might on Nightmare) matters only for an act 2 hire, which neither sandbox has (Rogue at Doofer, none at Questy).

## 5. Smallest read-only probe

`debug states` is a read-only lab verb (docs/forkdbg.md; `Client.Buffs`, d2dctl/verbs.go:305-317) returning `buffs=<skill>:<level>,`; the bot already parses it but keeps only five armor ids (d2dsrc/source.go:358-380). U that it lists aura states. On a lab instance (never the production containers) with a Paladin that knows Might: (1) `state` shows rskill; `debug states` (baseline); (2) pick Might through the popup exactly as the bot does and run `debug states` again; (3) one right-click cast (or a `cmd skill` cast of the aura id read from the skill table) then `debug states` and `state`. Expected readings decide section 3: buffs entry after (2) = picking starts the aura; only after (3) = the bot needs a cast. No memory write; no cheat ledger entry. Follow-up if confirmed: log rskill, lskill and the aura state in the minute line (d2d dump `player.states`).
