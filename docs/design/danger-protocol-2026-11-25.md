# Danger protocol and checkpoint portals (2026-11-25)

> status: design proposal, no code. Part 1 reads the code (tags: V = verified in code or a cited live record, S = secondary doc, U = unverified). Part 2 uses the SME answer (`sme-danger-protocol-2026-11-25.md` in the private ops docs, all EXP unless noted) and the owner's framing: a safety portal is a CHECKPOINT between waypoints.

## Today

### Does the bot drop a TP as an anchor today? No.
Every portal the bot casts in the field is an ESCAPE: it is cast when life is already low or the pack is big, and the hero walks into it at once and then does town chores (V: `internal/combat/safety.go` Decide, `internal/resource/abort.go`, `internal/bot/loop.go` handleSafety). Nothing casts a portal in a healthy state to keep a way back, and nothing remembers a portal as a return point. The only reuse of a standing portal is incidental: after a trip the hero walks the portal that is already open (`town.BackToField`).

### (a) In the fight (code values; the brief's 50/22/70 differ, the defaults are these)
- `Decide` (V, safety.go; defaults in `internal/env/tuning_combat.go`): life <= `combat.chickenPct` (25, Hell 30): rejuv if ready, else CHICKEN. Below `drinkRejuvPct` (35): rejuv. Below `drinkLifePct` (55): healing potion; with no potion at all, TownPortal; quiet and short of potions (`townPotionMin` 2): TownPortal for free healing. Champion pack early Normal: +`championEarlyPct` 10; boss pack uses `drinkRejuvBossPct` 50.
- CHICKEN = `Fighter.Leave` = Save and Exit through the menus (`cmd/d2bot/d2dwire.go`: `Cycler.LeaveGame`; V by code). It is not a portal and not a walk.
- Crowd rule (V, `internal/combat/crowd.go`): `crowdN` 8 hostiles within 8 subtiles. In a crowd: at life <= `crowdPortalPct` (70) and a scroll exists and the cast is not blocked: TownPortal; below `crowdDrinkPct` (65): drink; surrounded (adjacent >= 3) with a free step: CrowdRetreat. `crowdAbandonN` 15 on a travel or quest walk (40 for a player-started quest): CrowdAbandon = retreat 20+ subtiles, portal, `skipArea` for 4x navSkip, then a town trip (`handleSafety`).
- Cast guard (V, `resource/castguard.go`, `meleeCastGuard`): a portal cast is replaced by a potion, else CHICKEN, when too many attackers are adjacent. A hit does not interrupt the cast (V run, decomp28 Q5), so the guard is about damage during cast and entry.
- Potion burst (V, `internal/bot/earlyexit.go`): more than `potionBurstN` 5 life drinks in `potionBurstSec` 20 with >= 3 hostiles and a portal source: TownPortal instead of the next drink; the pack is noted as a stale pack (`stalePackHostiles` 4) and `redirectStalePack` keeps the return trip out of that area for `stalePackMinutes` 10 unless the kit is rebuilt.
- `AbortFight` (V, `resource/abort.go`): escalation Retreat, Portal, Leave, Flee for no-potion, no-mana or dead-merc cases; "portal still fast enough" is modelled as about 1.3 s against 1.8 s to leave (comment, U as a measurement).
- Kite (V, fight.go): single melee monster, life < `kiteLifePct` (60) and falling: step back, at most 6 steps.
- Route guard and trip budget (V): the route guard sends the hero home below a potion floor (`routeKit`); `resource.tripBudget` caps finished trips per window so a thin belt does not loop home. Boss runs end with `ErrEscaped / ErrChicken / ErrDead / ErrLeftArena / ErrTooWeak / ErrAborted` (`internal/boss/types.go`); per-script `EscapePct` 30, `ChickenPct` 12, `DrinkPct` 50.
- A portal cancelled before the click by a drink or retreat is retried once after `portalRetryWaitSec` 3 as a town trip (`internal/bot/portalretry.go`).

### (b) After death or chicken
- Death autopsy (`internal/bot/autopsy.go`, `deathlog.go`): kit, potions in the last minute, whether a portal cast was started, life trail. Spiral breaker and deaths-per-hour pause (`deathPolicy`, `pauseDeathsPerHour` 8).
- Corpse policy (V, `death.go` `retrievalDecision`, `finishRecovery`): retrieve only when the gear is worth it and the area is not judged lethal; at most `corpseRetries` attempts (the brief's #1291: 3). Each try walks to the death area and clicks the corpse (`retrieveCorpse`, `gotoDeathArea`). If it fails and `corpseWorthChasing`: `resetForCorpse` goes home, `startFreshGame` (Save and Exit, new game): the game moves a field corpse to the town of the next game (V d2d, per code comments), and `corpseTownStep` clicks it, max 3 tries. `corpseSweep` (#1299) re-checks every minute in town for an untracked corpse.
- Going back (V, `town/trip.go` BackToField): wait out the portal settle, take the portal object that stands in town (dropped after `portalClickFails`), else the waypoint of the level, else walk. A level only a portal leads to (Tristram) fails fast with `ErrPortalStale` if the standing portal came from another area. IMPORTANT (V, comment citing fork `travel.cpp:363`): game.exe and d2d open NO portal from town, so after death the only way back through a portal is one cast in the field BEFORE the death. Today that exists only by accident (an earlier escape cast).
- `preferWaypointRun` (V, `loop.go`): from town, the portal leads where the last scroll was read; when that is not the corpse area and an activated waypoint shortens the route, waypoint plus walk wins. Waypoint choice is by route length: the last path node with a plan waypoint at hop >= `nav.wpMinDistance` that `wpTripWorthTrying` accepts (`levelplan` also uses `ViaWP`). So "closest waypoint to the objective" is roughly what exists for corpse runs and travel (V by reading: it scans from the target end of the hop list, so it is the closest in hops to the target, not necessarily in walking distance).

### (c) How a portal is used
- Cast: `Fighter.openPortal` right-clicks the scroll or tome; the control channel also has a cast path in `d2dctl` (`town_verbs.go`, polls `Portals` every 300 ms for up to 3 s until the count grows; V by code).
- Entry: `d2dctl.EnterPortal` = fork verb `cmd interact -2000-which` through `TakePortal` (V live, `docs/d2d-portal-entry.md`: from next to the portal 1 tick; from 4.5 cells 34 ticks while d2d walks there). The bot wraps it in `approach` (stall detection, 15 s budget), because `cmd interact` stands still at the first object on a straight line (V lab probe portalfar).
- So the engine verb already needs no mouse. What is not instant: the cast (about 1 s per SME, EXP; code assumes up to 2 s in the crowd comment, U), the settle wait on arrival (`portalSettle`), and the walk to the portal (distance dependent).
- Latency measurement from the sandbox logs: NOT done. The container log API returns only the newest 1000 lines unfiltered and I did not pull them into context; the lines to grep are `town portal`, `back to field: portal in town`, `town portal cast started`. Left as a checklist item below.

## Proposal: checkpoint portals

### Where SME and our data disagree
1. SME: "after death walk out of the old portal puts you at the old spot where the monsters stand; drop the portal route to the corpse." Our data: the corpse is not in the field after a rejoin (V d2d), so a corpse run through a checkpoint is only needed when the bot does NOT reset. The bot already retries 3 times then resets. Keep the reset as the strong fallback; use the checkpoint for the FIRST try only, and only if the checkpoint area is not the death area's crowd (the stale-pack rule).
2. SME: rejoin removes all monsters. Our data (V, loop-breaker doc, areathrash): a new game respawns the area, so the checkpoint is gone after a reset (agrees: the checkpoint record dies with the game).
3. SME: second death before pickup loses the first corpse. Our corpse policy already caps retries and refuses a repeat crowd spot (`crowdSpotAgain`). Rule kept: no corpse try is allowed to start from a checkpoint that failed once.
4. SME: cast takes about 1 s and entry about 1 s. The code treats the cast as up to 2 s under attack (U both). This is why the anchor is cast BEFORE the fight.

### State machine (per game; one record)
```
NONE --proactive trigger, healthy, safe--> SET(area, cell, time, scrolls)
NONE --reactive: danger and no checkpoint, after retreat to a safe spot--> SET
SET  --new cast anywhere / game restart / waypoint taken out of the area / area left by exit--> NONE   (the engine allows one portal: a new TP replaces the old)
SET  --retreat trigger (life < retreatPct, or boss run abort)--> USE (walk to the portal, enter) --> TOWN chores
TOWN --chores done--> RETURN: standing portal if SET else waypoint else walk
DEAD --respawn--> TOWN; if SET: first corpse try through the portal (area not stale), else waypoint+walk; try N; failure or tries exhausted --> fresh game (corpse moves to town) --> waypoint nearest to objective
```
Triggers:
- PROACTIVE (healthy, life >= `checkpoint.minLifePct`, no hostiles within the fight radius, last checkpoint older than `minGapSec` 60 and farther than `minDistance` cells or in a different area): entering an area marked dangerous (boss area, area with a death or chicken in the log, monster level above clvl by `checkpoint.levelDelta`), at the start of a boss room approach (SME: "portal at the door"), after clearing a chokepoint, and at an arrival point of a new area when the previous checkpoint is more than one area away. Cast spot: arrival cell or the last cell of a cleared room, never mid-crowd.
- REACTIVE: `Decide` would choose TownPortal or CrowdAbandon for danger, but no checkpoint exists within `minGapSec`: first retreat toward the last cleared cell (existing CrowdRetreat), cast there, THEN continue the escape. Never cast when `meleeCastGuard` blocks it. SME: at life < 25% with no potion CHICKEN, do not cast.
- USE: retreat when life < `checkpoint.retreatPct` (35 to 40) with a checkpoint present, before the chicken line, so the fight is joined at full resources afterwards. The existing drink lines stay above it.
Records: `{area, cell x y, castAt, scrollsLeft, gameId}`; invalidate on every item above; the autopsy prints it.

### After death and the resume rule
Order (owner): checkpoint, waypoint, walk. Checkpoint try only if the record survives, `crowdSpotAgain` and the stale-pack rule allow it, and the kit is rebuilt (potions) first. Resume from the waypoint closest (hop count to the objective, as `preferWaypointRun` scans) when no checkpoint exists. A try that dies, or `corpseRetries` exhausted, goes to `resetForCorpse` unchanged.

### Instant portal entry (feasibility)
- Entry is already a pure control verb (`cmd interact -2000-which`, V). Latency pieces: walk distance and the 15 s `approach` stall logic. Spec for a thin verb `cmd portal enter` (fork, control channel, not a memory write, so ADR 0014 stays satisfied; it is the same World command a click sends): `portal enter` = TakePortal for the nearest portal of this level owned by the player, reply `ok level=<new>` when the move happened or `err` with a reason (no portal, too far, blocked), no client-side walk polling. Because the checkpoint is cast standing next to the spot, the entry distance is under 2 cells, 1 tick (V probe row 3). Gain over today: it removes the portal list read and the approach loop, a few hundred ms of 60 ms polling; the real wait is the cast (about 1 s) which no verb removes unless a `cmd cast townportal` also exists. A cast verb would be a control verb too (the bot's own input), U for whether the engine allows a scroll-less cast; not recommended because scrolls are the cost control.
- Use the checkpoint proactively and the instant case is simple: cast, walk away; entry later from a distance is one `cmd interact` that d2d walks.

### Risks
- Scroll supply: one cast per `minGapSec` can drain scrolls; cap by `checkpoint.maxPerHour` and stock floor `checkpoint.minScrolls` (keep 2 for escapes). Tome charges cheaper (S, SME: scroll about 100 gold).
- Portal closed or replaced by the bot's own escape cast: record invalidated on any cast (log line).
- Monsters at the portal after death (SME): stale-pack rule; skip the checkpoint try.
- Quest-state areas (Tristram, Cain gibbet) may reset on rejoin (SME UNSURE); `data.PortalOnly` handling stays.
- Proactive casts cost game time (2 to 3 s) and a hit during the cast: only when no hostile is within the fight radius.

### Data knobs (new group `checkpoint.*`, defaults off first)
`enabled` 0, `minLifePct` 80, `retreatPct` 38, `minGapSec` 60, `minDistance` 40, `levelDelta` 3, `maxPerHour` 12, `minScrolls` 2, `bossDoor` 1, `staleSkipMin` 10 (reuses stalePack). Per-area danger and chokepoint cells live in the plan data (`internal/plandata`), not in Go. Class is irrelevant to the protocol; the exposure test (`CastGuard`) already reads damage, so no class check is added.

### Tests
Proving Ground scenarios (`internal/scenario`, `mem.Fake`, two builds: Sorceress and a melee class): (1) proactive: arrive healthy in a boss area, expect one cast and a record; (2) rate limit: second trigger within 60 s casts nothing; (3) reactive: crowd, no record: retreat, cast, escape; (4) danger and USE below `retreatPct`; (5) death with a live record: first try via portal, then waypoint, then reset on repeat failure; (6) a new cast invalidates; (7) no scrolls: no cast, existing escapes unchanged; (8) cast guard blocked: no cast.

### Effort
- Checkpoint record and invalidation in the Runner (S): 1 agent, about 150 lines plus tests.
- Proactive triggers and spot choice with plan data (M): 1 agent, 250 lines, needs the cleared-cell source.
- Reactive and USE wiring into `Decide` / `handleSafety` (M): 150 lines; touches the survival path, review with Opus singly.
- After-death order (S): 80 lines in `finishRecovery` / `BackToField`.
- Fork verb `portal enter` (S): fork side, about 60 lines, lab binary and a drill; upstream no.
- Live checks (cannot be done offline): latency of cast and entry from the sandbox logs (grep `town portal`, `back to field: portal in town`, `town portal cast started` with timestamps); one drill: cast at a door, walk away, die or Save-and-Exit, return through the portal; confirm that the portal survives the death in the lab (U).

## Built in phase 1 (card checkpoint1) and deferred

Built, all behind `bot.checkpointEnable` (default 0; knob keys are `bot.checkpoint*`, not a new `checkpoint.*` group, to stay inside the existing Bot tuning table and its generated docs/knobs.md):
- `internal/bot/checkpoint.go`: record `{area, x, y, at, scrolls}`, states none / set / use / town / return / dead, the pure rule `cpDecide`, guard hook (`errCheckpoint`) and the main-loop `checkpointStep`.
- PROACTIVE cast: danger area from DATA (`internal/plandata/checkpoint_areas.go`, `CheckpointArea`), life >= `bot.checkpointMinLifePct` 80, no hostile within `checkpointSafeRadius` 30, within `checkpointEntryDist` 40 of where the hero entered the area, gap `checkpointGapSec` 60, more casts in the pack than `checkpointMinScrolls` 2 (scrolls plus tome charges), no record in this area yet, never while a step forbids portals (Ancients).
- RETREAT: life <= `checkpointRetreatPct` 40, record in the same area within `checkpointReach` 60, fewer than 3 adjacent; the potion and crowd rules still run first (the guard asks Safety before the checkpoint). The hero walks to the cell, then the existing emergency trip enters the open portal (`ToTown`) and returns through it.
- AFTER DEATH (`cpRoute`, `cpArrival`, in the corpse run): a usable record in the corpse area (no stale pack, area allowed by the route guard) makes the standing town portal the way back (`portalFrom` = area); otherwise the record is dropped and the waypoint then the walk go first (the stale portal of an earlier trip is no longer trusted while the feature is on). Arrival check: `checkpointArrivalHostiles` 4 or more hostiles near the exit send the hero straight back to town, note the stale pack and fail that corpse try. The corpse policy (#1291 retries, #1299 sweep, the reset) is untouched.
- Invalidation: a safety portal cast, a new game (`forgetGame`), or going home alive from a place far from the portal (`cpEntryNear` 12).
- Tests: `internal/bot/checkpoint_test.go` (decision table, cast, off by default, scroll reserve, hostiles near, retreat, invalidations, death route, arrival crowd) on a Sorceress and a Barbarian rig. They were written with the code, so there is no RED run against the old tree; the old tree cannot express the knob.

Deferred: the REACTIVE anchor (retreat to a safe spot and cast when danger meets no checkpoint), proactive casts "after clearing a chokepoint" and for death/chicken-history areas, a scenario in the Proving Ground lab (needs the lab and `debug town portal`; checklist: cast at the Crypt door, walk away, die or take Save and Exit, return through the portal, confirm the field portal survives the death), the instant `portal enter` fork verb (phase 2), a Catacombs 4 / Andariel door cast (the route guard's kit gate closes that area for the rig hero, so the data lists it but the test uses the Crypt). U: whether the portal object stays after the owner dies (the whole after-death route depends on it; if it does not, `cpArrival` finds no portal and BackToField falls back to the waypoint).
