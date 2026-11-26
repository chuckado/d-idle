# Paladin aura state on d2d, measured live (2026-11-26)

> status: probe result, read-only on the bot; no fix needed for the pick path. Follows docs/probes/paladin-aura-audit-2026-11-26.md (branch agent/auraaudit) which could not tell whether a popup pick starts the aura. Tags: V verified live (command stated), S read in the fork source, U unverified. Card: auralive.

## Verdict
V: on d2d a picked right-hand aura is ACTIVE at once, with no cast. The right skill popup pick (the bot's `SetRight`) sends the SelectSkill command; the world sets the player's active aura to that skill in the same command. A right click, a `cmd skill <aura id>` cast or an attack cast is neither needed nor does it change the state. Selecting a non-aura right skill turns the aura off. The user's rule (an aura works only while it sits on an active skill slot) is therefore true, and the bot satisfies it: the right slot holds the aura, and the left slot is never used for an attack on d2d (attacks go by `cmd skill <id>`, which does not move either hand).

## How the aura state is read
- `state` line field `aura=<skill id>`: the player's active aura (fork `devctl/verbs.hpp` prints `town.view.aura`, which is `actor->aura`, the field the aura pulse and the strike code read). V: it follows every pick below. Not read by the bot yet (U: the d2d dump has no such key in `player.*`).
- `debug states` (`buffs=`) stays EMPTY with an aura on: it lists self states from casts, not the aura. V (every row below). Do not use it for auras.
- Effect check, not only the field: V, a Holy Fire pulse hurt a monster with no cast (section "Effect").
- Stat 350 (game.exe stat list) is a game.exe memory read; d2d has no memory to read, so it was not used.

## Setup and provenance
Throwaway Paladin "Auri" per run, `--no-save`, lockstep, lab binary built from the fork (`d2d.devp27-0fcd512a-linuxbuild`), one hidden lab slot (no game.exe instance, no desktop window). SETUP before the first reading, each ledgered with a reason (actor `auralive`): `debug stat 12 6` (clvl 6), `debug points 98|102|107|106 1` (Might, Holy Fire, Sacrifice, Zeal). Arm `holyfire` also `debug points 102 20`, `debug goto 2`, `debug spawn zombie1 1 near`, `debug mon set 1 life 400 maxlife 400 dmg 0 0`, `debug warp 38 58` before its first reading. In arm `pick` the area change under test is `debug goto 2` and `debug goto 1` (teleports, DIAGNOSIS, declared). The character is CHEATED/IMPOSSIBLE (skill points beyond the level, teleports): mechanics only, never evidence about balance. Not logged in docs/probes/test-characters.md (not in this card's owns; the line is in the report). Nothing is left running, the games are gone.

## Commands
Probe: `cmd/auraprobe` (this branch). On the worker host, inside the worktree:
`D2D_BIN=$HOME/d2d-lab/fork/bin/d2d.devp27-0fcd512a-linuxbuild D2D_DATA=$HOME/d2d-lab/data D2D_AGENT=auralive go build -o /tmp/auralive-probe ./cmd/auraprobe` then per arm
`tools/agentrun.sh auralive --ttl 400 -- tools/d2dslot.sh run -t 380 -- /tmp/auralive-probe -arm pick|cast|cmdskill|left|holyfire`.
Picks use the bot's own `combat.Fighter.SetRight` over `d2dcombat.Attach` with the popup rows loaded (`combat.LoadListRows`), exactly the production path.

## Readings (clock 2026-11-26 13:09 to 13:12 EST, slot 4; `aura` is the `state` field, `r` rskill; every reading also had `debug states` buffs empty)
| arm | step | aura | r | area |
|---|---|---|---|---|
| pick | A baseline | 0 | 0 | 1 |
| pick | B after popup pick Might (13:09:28.684) | 98 | 98 | 1 |
| pick | B +100 ticks | 98 | 98 | 1 |
| pick | F after 3x `cmd skill 107 x y` (Sacrifice, the bot's attack path) | 98 | 98 | 1 |
| pick | B2 pick Holy Fire (switch) | 102 | 102 | 1 |
| pick | B3 pick Might again | 98 | 98 | 1 |
| pick | D `cmd skill 106 x y ... left` (Zeal) | 98 | 98 | 1 |
| pick | E after `debug goto 2` (area change) | 98 | 98 | 2 |
| pick | E +200 ticks | 98 | 98 | 2 |
| pick | E2 after `debug goto 1` (back to town) | 98 | 98 | 1 |
| cast | B pick Might, then `rclick 400 250` (ground), +40 and +100 ticks | 98 | 98 | 1 |
| cmdskill | A, then `cmd skill 98 x y` and `... left` with NO pick | 0 | 0 | 1 |
| left | B pick Might | 98 | 98 | 1 |
| left | D pick Sacrifice (non-aura) as the right skill | 0 | 107 | 1 |

Answers to the brief: (A) 0. (B) set at once, aura equals the picked id. (C) a right click changes nothing, there was nothing to start. (D) persists through attack casts, a left-flagged cast and switching between auras; a NON-aura right skill clears it. (E) persists through an area change both ways and 200 ticks. (F) `cmd skill <attack id>` does not affect it. Extra: `cmd skill <aura id>` without a pick does NOT start the aura (aura stays 0, rskill stays 0): the cast verb is not a way to turn an aura on.

## Effect (arm holyfire, Holy Fire level 21, a 400 life Quill Rat beside the hero, damage 0, no cast at any time)
| 13:11:35 | Quill Rat #1 life |
|---|---|
| control, no aura, start and +100 ticks | 400, 400 |
| after the popup pick, +100 ticks | 363 |
| +200 ticks | 323 |
| after picking Sacrifice (non-aura) as right, start and +100 ticks | 323, 323 |
V: the picked Holy Fire damages monsters with no cast, and stops when the right skill becomes a non-aura skill. (The first run at Holy Fire level 1 showed no loss in 200 ticks; the rat walked 2.6 cells out and the pulse at level 1 is small: U why, the level 21 run is the evidence.)

## Source reading (S, fork wt-devp17, components/game/world/commands.cpp)
`SelectSkill` on the right hand: `actor->aura = skill->aura ? select->skill : 0` (line 485); a weapon swap does the same (line 477). Nothing else assigns the player's aura. U live: a character LOAD (new game with a saved right skill) leaves `actor->aura` 0 until a SelectSkill arrives; the town client (`apps/d2d/town/walk.cpp` line 387) resends SelectSkill when the aura field and the right skill disagree, which should heal it, but this was not measurable here (no save in a `--no-save` lab game). Residual risk for the bot: `SetRight` returns at once when `rskill == id` (bindleft.go:197), so if a load left rskill at the aura with the state off and the client did not resend, the bot would not repair it. Cheap guard if wanted: read `aura=` from the state line and, when it differs from the wanted id, pick another aura and the wanted one again (about 6 clicks).

## What the audit's ranked list becomes
1. HIGH, was U: now V for the pick path. No cast is needed; the audit's section 3 hypothesis is false on d2d.
2. Still true: the bot never reads the aura state (log `aura up` equals rskill). The `state` key `aura=` is a free, exact read; wiring it into the d2d source and the `aura up` check is a small separate card (suggested: `player.aura` in internal/d2dsrc, then `ActiveAura` uses it, so the load case above is covered).
3. Doofer's Might and Holy Fire flips are real aura switches and each takes effect at once (B2, B3).
