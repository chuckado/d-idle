# Map knowledge: would revealing the whole map fix the Countess class of stuck steps?

> status: current (2026-11-21, Fable design review, small; docs only, no Go or behaviour change). Card `fablemap`. Proposals and questions only: the user decides (section 6); nothing here is an ADR.

Tags: **V** = I read the cited file and line in this run (a V on a doc line means the doc says so; "V live" is the doc's own live claim), **S** = secondary or inferred from several sources, **U** = unverified. Nothing in this review was run against a game.

The question (the user's words): "if we reveal the entire map to the bot and fake it exploring, would that fix the pathfinding issues with bosses like the Countess, since it can plan the route and make it messier".

## 1. Verdict

**No, because the bot already has it.** On d2d the collision grid of the current level arrives whole on entry and the bot already "fakes exploring" on top of it. The stuck steps of the last week were not caused by missing walls. They were caused by four other gaps: where the TARGET is, what shape an exit TRIGGER has, how LEVELS LINK, and guards and wiring that have nothing to do with maps. A new "whole collision grid" verb would fix none of them.

What the proposal gets right is the direction: read what the engine knows on arrival instead of searching for it. Three such reads landed this week (object layout, monster presets, object presets of all rooms) and they are what turned the Countess, the Dark Wood waypoint and the Cain tree from searches into lookups. Two reads of the same kind are still missing (trigger areas, level exits) and those are the first slice. The "messier" wander is a feel feature, not a fix; I recommend deferring it and I disagree that it helps pathfinding (section 4.2).

## 2. What the bot knows today (A), with evidence

| Fact | Evidence | Tag |
|---|---|---|
| `debug grid 5` answers the collision of the WHOLE current level in one reply; the world source fetches it on every arrival | `internal/d2dworld/world.go:292`, `:224-230` | V |
| The first fetch knows every cell: town `58425 = 285*205`, Blood Moor `134400 = 480*280`; `Level::walk` is filled for the whole level when it is built; "nothing in `debug grid` is unknown" | `docs/d2d-backend.md:684`, `:686` | V live per doc, for the town and the Blood Moor ONLY |
| The same for maze levels (Tower Cellar, Jail, Catacombs) and acts 2 to 5 | no probe found | U |
| The world source tells nav "every level map is whole" as a hard-coded `true`, not a measured property | `internal/d2dworld/world.go:575-577` | V |
| So `Explore` has nothing to reveal and clearing is a ROAM over coverage points (spacing 16, visit radius 10 subtiles) laid on the whole reachable level | `internal/nav/roam.go:16-18`, `:241`; `internal/env/tuning_nav.go:43-44`, `:53` | V |
| Room CONTENTS are not whole: objects and monsters are made when a room comes up near the hero; `debug presets` lists built rooms only | `docs/d2d-backend.md:662`; `docs/process/layout-read.md:25` | V (S for the engine reason) |
| Three K1 reads fill that gap behind knob `quests.layoutRead`: `debug levelmap` (objects of built rooms), `debug levelmonsters` (monster presets of all rooms), its `objpresets` (object presets of all rooms) | `docs/process/layout-read.md:7`, `:26`; task-770 `result:` | V; task-770 says "NOT verified in the live sandbox" |

So proposal B(1) describes, for collision, the status quo. "Fake exploring" also exists: the roam is exactly a walk over a map the bot already holds.

**One conflict I could not settle.** The nav test for the Countess fix says her stand (21,70) was "100 subtiles west and outside the known map" and that the planner's next waypoint was "the known grid's edge (121,70)" (`internal/nav/mouthgoal_test.go:10-12`, V as a comment). The probe doc of the same day says "every d2d level map is whole" (`docs/probes/provingground-countess-2026-11-20.md:8`, V). The test's own fixture is an open field that is fully known, goal included (`mouthgoal_test.go:23-25`, `internal/nav/sim_test.go:354`), so the test does not model a partial grid and passes either way. Either the Tower Cellar grid really was partial at arrival (then maze levels differ from the Blood Moor and `FullyKnown() == true` is wrong for them) or (121,70) was simply the first corner of a normal path and the comment mislabels it. This is the one place where the proposal could still be right, and one cheap probe decides it (card 1).

If the grid can be partial, today's combination is the worst one: unknown cells are not walkable (`world.go:6-7`), and nav is told there is nothing to explore (`roam.go:16-18`), so a goal behind unknown cells is a silent dead end. That is worth fixing whatever the probe says (card 2).

## 3. Stuck-step classes: what a whole map fixes and what it does not

| Class | Root cause on record | Whole collision map fixes it? | What does |
|---|---|---|---|
| Countess, attempt 1: chest anchor | object 580 is not in Tower Cellar 5 (`provingground-countess...md:7`, V) | no | target knowledge |
| Countess, attempt 2: no presets | `Env.World` was nil on d2d, so the boss code saw no presets at all (`:8`, `:35`, V) | no | wiring, fixed |
| Countess, attempt 3: `Anchor.Spawns` | coordinates copied from one live game; the cellar is random per game; every target resolved to the same refused click (`:22`, V) | no | `debug levelmonsters` names her preset by NAME (`:33`, `:36`, V) |
| Countess, attempt 4: found, not reachable | the mouth gate refused the first click west of the arrival stairs, 13.5 from them (`:36`; task-769 `result:`) | **no** | the gate rule changed; the real trigger shape is still U (`internal/nav/mouthgate.go:21-22`, `:42`) |
| Countess, all attempts: held in town | potion guards and the scenario setup (`:17-21`, V) | no | town and chores cards |
| Dark Wood waypoint, over an hour | the waypoint object sits in a room not built; `levelmap` lists built rooms only (task-716 `result:`, V) | no (the outdoor grid was already whole) | `objpresets` (task-770; builder green, live U) |
| Cain tree | the preset is listed only within about 13 to 18 cells (`docs/design/fable-review-2026-11-19/05-architecture-and-overhaul-options.md:77`, V) | no | the same read |
| Monastery Gate, Outer Cloister, Jail: "nav no usable gate", "no exit to that level: level 27 from 26" | a level-to-level link, not a wall: neighbours are built on a worker thread and linked late; a gate needs the neighbour's collision across the edge (`internal/d2dworld/border.go:16-17`, `world.go:64-68`, `internal/nav/exit.go:35`, `:96`; `docs/probes/provingground-act1more-2026-11-20.md:29`). Root cause U | **no**: the current level's grid says nothing about the other side | an exits read (card 4) |
| Mouth gates in general | five generations of fixes around constants 15 and 10 because the trigger area's "real bound and shape ... are UNVERIFIED" (`mouthgate.go:13-23`, `:39-43`, V) | **no** | read the trigger area from the engine (card 3) |
| Combat danger on the way | nav has no threat cost: only an exit penalty (`internal/nav/nav.go:569-573`) and a stop with `ErrDanger` (`internal/nav/walk.go:698-723`). The survival model prices ZONES, not cells (`internal/survival/ledger.go:16`, `:294`) | no | nothing in this proposal: "A* with the survival model's threat costs" does not exist to plug in, and the packs of rooms not yet up do not exist to cost (`layout-read.md:25`) |
| Doors, walls that open | closed-door collision in the grid is U (`world.go:470-472`; `docs/d2d-backend.md:666`) | no; a whole-route plan is MORE exposed to it | door state read and replanning |
| Quest-state evidence, hand-ins, board offers | records and readiness (`provingground-act1more...md:29`) | no | quest and board cards |

Score: of the classes above the whole collision grid fixes zero, with one U (the cellar edge of section 2). Target knowledge fixed three. Two are open and are about exits.

## 4. Costs and risks of the proposal's three parts

### 4.1 (1) Whole-level grid read and whole-route planning

- Fork work: probably none for collision; the verb exists. The cost is VERIFICATION of the claim it rests on: a differential of the grid at arrival against the grid after every room is up, for N levels with maze levels included, 0 differences apart from unit stamps. Unit stamps are real: monsters and critters are in the grid, three Blood Moor grids differed by 5 to 10 cells (`docs/d2d-backend.md:684`, V live per doc), so the differential must subtract them.
- Size and refresh: 134 KB of text for the Blood Moor, up to about 1 MB for a big level (`docs/d2d-backend.md:668`); refetched after 60 subtiles of movement and 2 s (`world.go:117-122`). Fine per level, not per tick.
- Determinism: a refetch that happened in one run and not the other was the first divergence of two same-seed runs (`docs/d2d-determinism.md:84-85`, `:158`, V). Any new periodic read needs a schedule keyed on game ticks or position, never wall time.
- Dynamic obstacles: the plan is only as good as the last fetch. A long route planned once over doors (U) fails late; today's short legs fail early. Keep replanning on the way.
- Game.exe backend: it reveals rooms as the hero walks, so the two backends differ here. ADR 0023 stops planning to PLAY on Game.exe (`docs/adr/0023-fidelity-by-rule-differentials.md`), so I would not build whole-map for it.
- Teleport builds are where a whole map changes play most: hops are chosen from the known grid by distance (`internal/nav/teleport.go:14`), so with the target known the Sorceress goes straight to the boss. That is what kolbot-style bots do (survey: they read presets on entry and path on the game's own collision, `docs/design/fable-review-2026-11-19/02-fidelity-and-gaps.md:85`, S). Fast, and visibly not a search.

### 4.2 (2) Seeded route wander ("messier")

Where I disagree. Messier routes do not fix pathfinding; they add clicks, and every stuck class in section 3 that involves a click (mouth gates, doors, borders) gets more chances to fire. Three facts:

- In FARM mode the hero is already as messy as it can be: it visits every coverage point of the zone until `bot.exploredPct` is reached (`internal/bot/goal.go:1376-1377`, `internal/bot/exploredpct.go:53`). Farming does not depend on fog or on explore-to-find for collision; it depends on the roam. So whole-map routing cannot starve XP or loot in farm mode, and wander adds nothing there.
- On QUEST and TRAVEL legs the hero goes straight (`internal/bot/task.go:743` explores only until the target is seen). This is the only place wander would change anything. There a detour costs time and potions, and the Countess runs were already held by potion guards after the walk in (`provingground-countess...md:17`, `:21`). A 10 to 20 percent longer path through a monster zone is not 10 to 20 percent more risk; it is more fights.
- It is a FEEL feature (the stream, the idle-game look), not a fidelity feature. Fidelity is per-rule engine differentials (ADR 0023); the hero's route style is not an engine rule and no claim depends on it.

If the user wants it for feel, the honest design is not random noise but a priced choice: on a leg through a zone whose blended yield (`internal/survival/ledger.go:385`) beats the hero's current best, route through up to K nearby coverage points; otherwise go straight. Parameters as data, per build profile and per act: `curiosity` (share of extra path, default 0), `maxSideLooks`. Seed: a hash of game seed, level id and leg index; no wall clock. In Proving Ground it stays 0: two runs of one seed still differ in the full digest for our own reasons (`docs/adr/0027-proving-ground-gate-and-night-window.md:12`; `provingground-act1more...md:21`), and wander would hide real drift behind intended variation.

### 4.3 (3) Fog of war in the viewer

Mostly built. Card mapfog (task-382, Done, offline and headless render, not live) withholds coverage points, targets, events and heat server-side in blocks the hero has not explored (owns `internal/senses`, `internal/watch/senses.go`, `internal/watch/web/panels/map`). The data it needs is the roam coverage (`roam.go:61`). The gap: `NotePosition` "does nothing without a coverage of this level" (`roam.go:208-211`), and a coverage exists only once a level is roamed, so a level crossed on a quest or travel leg has no walked trail to reveal. Small card (card 5). No fork work.

### 4.4 Fidelity and the idle-game feel

- Runs: reads never make a run CHEATED (`docs/adr/0022-hero-knowledge-reads-never-cheat.md`). No A:B-style claim changes, because there is none left to change (ADR 0023).
- Product: ADR 0022 decided K1, "the level room layout and the quest-object rooms". The monster-preset read already goes past that for superuniques: the hero knows the Countess' cell on arrival, which the 2026-11-19 review called K2 ("everything a maphack shows", `02-fidelity-and-gaps.md:91-93`). I think that is the right call for quest targets and bosses and the wrong one for ordinary packs; it should be said out loud (Q2), because "what the hero may ACT on is a product decision" (`05-architecture-and-overhaul-options.md:127`).
- Feel: a hero that walks straight through a random maze to the boss reads as omniscient. The fog in the viewer is the cheap mitigation; wander is the expensive one.

## 5. Recommendation and minimal first slice

Do not build a whole-map verb. Prove the map is already whole, stop nav from trusting that blindly, then read the two things that are actually unknown.

| # | Card | Owns | Scenario | Effort (agent-days) |
|---|---|---|---|---|
| 1 | `gridwhole`: probe. `debug grid 5` at arrival against the grid after all rooms are up, 12 levels (4 outdoor, Den, Tower Cellar 5, Jail 1, Catacombs 2, 4 from acts 2 to 5), 3 seeds, unit stamps subtracted; pass = 0 differences. Also one log line per arrival: known cells of rectangle cells | `docs/probes/gridwhole-*.md`, the fork's private tests | arrival of `act1.forgotten.countess` | 1 |
| 2 | `knownedge`: `FullyKnown` becomes measured (known cells equal the rectangle); a goal behind unknown cells plans optimistically (the `fOpt` mode exists: `internal/nav/field.go:13`, `nav.go:426`) and refetches on approach; must-fail test on a REAL partial grid, Paladin and Sorceress | `internal/d2dworld`, `internal/nav` | `act1.forgotten.countess`, `.sorc` | 1 |
| 3 | `warptrigger`: read-only fork verb listing each warp's trigger cells; the mouth gate uses the shape instead of 15 and 10. Differential: stepping on a listed cell fires, a neighbour does not, N warps, 0 differences | fork branch, `internal/d2dctl`, `internal/nav/mouthgate.go` | `act1.den.clear`, `act1.forgotten.countess` | 2 to 3 |
| 4 | `exitsread`: probe first. Dump layout, nearby, warps and the border strip at the Monastery Gate and the Outer Cloister, name the cause; the fix is a second card | `docs/probes/exitsread-*.md` | `act1.andariel.slaughter` | 1 |
| 5 | `fogtrail`: a coverage for every level on arrival, so the Map panel's fog follows travel legs too | `internal/nav/roam.go`, `internal/senses` | a new `watch.map.fog-travel` | 0.5 to 1 |

About 6 agent-days. Cards 1 and 4 are probes and can run first and in parallel; 2 waits for 1; 3 needs the user's go for a new fork read (Q3). Wander is not in the slice.

**Interim, before card 2 lands:** keep the #1055 rule (a waypoint inside the 15 bound passes when the goal lies beyond and it stays outside the 10 core). Add nothing else to the gate. If a walk again ends on a "grid edge" waypoint, the arrival log line of card 1 is the first thing to read; do not answer it with another gate exception.

## 6. Open questions (the user decides)

1. **Drop the whole-map verb and treat "collision is whole on arrival" as the working truth, pending card 1?** Recommended: yes. If card 1 finds differences on maze levels, the verb comes back as a real card with that evidence.
2. **How much may the hero act on?** Recommended: quest objects, waypoints, exits and named bosses (superuniques) on arrival: yes. Ordinary monster packs of rooms not yet up: no, even if a read could list their presets. It keeps danger and discovery honest and keeps the farm roam meaningful.
3. **A read verb for warp trigger areas (card 3)?** Recommended: yes. It replaces two guessed constants and the fixmouth series with an engine fact, under the ADR 0022 rule.
4. **Wander.** Recommended: defer until cards 1 to 4 are green; then, if wanted for the stream, the priced version of 4.2 with `curiosity` 0 by default and always 0 in Proving Ground. Not random detours.
5. **Fog in the viewer: walked ground only, or the level outline dimmed with the objective marked?** Recommended: walked ground only, plus the objective marker once the hero has it; the outline gives away the layout the fog is there to hide.
6. **Teleport builds going straight to a boss: accepted?** Recommended: yes for quest steps; farming stays a roam, so yield does not change.
7. **Game.exe backend: any whole-map work there?** Recommended: none (ADR 0023).

## 7. Dead ends and limits of this review

- I did not read the logs of job 20261121-005917, so the section 2 conflict stands as U.
- The fork source was not read; every statement about the engine is from repo docs (S).
- ADR 0021 and 0026 were not needed for the verdict: the waypoint, portal and pickup verbs (#1083) change how a step ACTS at its target, not how it gets there, and the ground-target grammar is a combat matter.
