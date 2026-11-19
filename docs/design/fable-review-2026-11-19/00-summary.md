# Design review 2026-11-19: summary for the owner

> status: REVIEW for the owner (Fable, card fablereview, branch `agent/fablereview`). Docs only. Nothing here is decided; nothing in the bot's behaviour changed; no game was run, nothing was posted, nothing touched the live hero. The questions for the owner are Q1 to Q12 in [06-roadmap.md](06-roadmap.md) section 7.
>
> The owner's ask (2026-11-19): compare all public Diablo II bots with ours; fidelity improvements; gaps; think about all seven classes; let the player decide every stat and skill point if that can be done, or discuss alternatives; weigh fidelity and bot capability, capability first; no cap on findings up to a complete overhaul.
>
> Tags used in all seven files: **[V]** read or run by me or a named helper of this review from a primary source; **[S]** secondary; **[U]** unverified or judgment. Helpers were Sonnet agents (source re-read of the survey's clones, about 60 web fetches, a d2d coverage read, two code maps, two SME consults, three offline prototypes). The SME consults were weak on details and are used for ranking arguments only.

## The files

| file | question it answers |
|---|---|
| [01-landscape-and-comparison.md](01-landscape-and-comparison.md) | what other bots do better and worse; what to adopt and what not; corrections to the stage-1 survey |
| [02-fidelity-and-gaps.md](02-fidelity-and-gaps.md) | ranked: where d2d deviates on the hero's path to Hell; which mechanics the bot lacks; what the hero is allowed to know |
| [03-classes-plan.md](03-classes-plan.md) | seven classes: builds, guide tracks, breakpoints, dangers, engine needs, order and effort |
| [04-stat-skill-agency.md](04-stat-skill-agency.md) | THE KEY DESIGN: free stat and skill points; four options; one recommendation with phases, tests and failure modes |
| [05-architecture-and-overhaul-options.md](05-architecture-and-overhaul-options.md) | keep, rebuild, or start over; the Search for Cain and ping-pong case studies; the Proving Ground |
| [06-roadmap.md](06-roadmap.md) | ranked packages with effort, first week and month, prototypes, risks, the twelve questions |

## The ten findings that matter most

1. **The project builds far faster than it proves.** 20 days, 8,014 commits, about 105 merges a day, 247,771 lines of Go; the proven ground is still Normal Act 1 and the friend sandbox hero stood at one tree for 17 hours after three merged fixes [V `git log`, docs/probes/treefix3-2026-11-19.md]. The quest board (ten slices), the vault (eight) and Ether were all merged "offline, NOT verified live" [V their status lines]. The review of 2026-11-10 already said "verify before building the next one"; as a rule it did not hold. [05](05-architecture-and-overhaul-options.md) sections 0, 8.
2. **The fix is a machine, not a rule: a Proving Ground.** Lockstep d2d scenarios per quest chunk (minutes) as a merge gate, and a nightly act run per class and seed. The engine for it exists and runs 46 to 62 times real time with the bot attached [V docs/d2d-loop.md]; today it is used for survival A/B runs and little else. [05](05-architecture-and-overhaul-options.md) section 5.
3. **No complete overhaul. Rebuild the decision core.** The seams, the data layer, the adapters, the fidelity lane, the loot rules and the survival model are good and rare. The problem is about 25,000 lines: `runGoal` is an ordered if-chain of about twenty hooks with FIVE separate loop guards patched on after incidents (flail, ping-pong, explore-gave-up, border loop, arrival bounce) [V internal/bot]. Replace it with an intent arbiter (commitment, priced travel, batched town visits, a decision trace), in shadow first. [05](05-architecture-and-overhaul-options.md) 3.1.
4. **Quest chains failed because nothing could check them.** Search for Cain broke in seven ways on first contact; five are static faults a checker finds without a game (a fact with no reader on the backend, identity by display name, no postcondition). Public bots finish an interaction when the OBJECT'S STATE changes and identify items by class id; kolbot uses a pointer click, not the interact packet, for quest objects [V helper A]. [05](05-architecture-and-overhaul-options.md) 3.2.
5. **We own the engine fork and use it like a remote desktop.** The hero learns of a quest object only within about 15 cells; points are spent by clicking UI buttons although verbs exist; ground item quality is guessed from label text. Every injected bot reads the level's presets on entry. A read-only "room rows on entry" verb (option K1) removes a whole class of failure and touches no fidelity rule. [02](02-fidelity-and-gaps.md) section 5, [05](05-architecture-and-overhaul-options.md) section 4.
6. **Against public bots we are narrow and deep where they are wide and shallow.** kolbot (about 96,000 lines) plays seven classes at endgame; SoloPlay levels 37 builds from 1 to 99 with staged respecs; none of them has tests that play the game, a fidelity lane, or a player product. No reinforcement-learning or LLM Diablo II player and no D2 idle game on the real rules was found. Two survey claims were wrong: d2bsng's simulator is in a PRIVATE repository and its determinism is undocumented; jaenster-kolbot's automatic skill choice is a damage-ranking model that admits melee does not work. [01](01-landscape-and-comparison.md).
7. **d2d moved fast and our maps are stale.** Upstream now plays all five acts and populates Nightmare and Hell; listed gaps fell from 489 to 125 [V helper E]. The gaps that matter next are per class and per act (melee chain not merged, curse rule, auras not reaching the mercenary, difficulty start through our launcher, Assassin 0 of 30 skills run-tested, Druid effects without a verdict). The fidelity queue should follow the class and act roadmap, and the statistical A:B verdict (never conclusive) should give way to function-level differentials. [02](02-fidelity-and-gaps.md).
8. **Seven classes are sixteen "delivery kinds", not seven projects.** Data for all classes already exists (tables, profiles, upkeep code) and none beyond two has played. Order: Sorceress second build, Necromancer Summoner (the best idle hero, and it forces the summon, corpse and curse drivers three other classes reuse), Hammerdin, Barbarian, ground-target builds with the Druid, Assassin, Amazon. Shapeshifting and martial-arts charges: not for the bot. [03](03-classes-plan.md).
9. **Ladder-only runewords do not exist offline, and our tables assume them.** Spirit, Insight, Fortitude, Grief, Infinity and 18 more cannot be made in vanilla single-player 1.14d [V this install's runes.txt via docs/sme/builds-gear.md; helper B: two archives agree]. No Insight means no free mana. Five endgame plan tables set their stat targets by Spirit, Insight or Fortitude gear. Whether d.idle mode enables them is a product decision (Q4). Also found: the rule that a skill's points cannot exceed character level minus required level plus one is not in our leveler and at least one table violates it [U high confidence; V kolbot enforces it]. [03](03-classes-plan.md) section 1, [04](04-stat-skill-agency.md) 3.4.
10. **The survival model is the right unifier and is half switched on.** One margin (power against danger) should judge zone choice, quest readiness, the difficulty transition and build viability. Its holdout failure came from the recovery state bypassing it, which is finding 3 again. [05](05-architecture-and-overhaul-options.md) 3.3.

## The recommendation on stat and skill points

**Give the player every point (path D: free points, honest hero).** Free allocation under the game's own rules; the hero derives its play from the skills it actually has by scoring them one at a time (no build enumeration, so no exponential blow-up); a generated capability table says per skill whether the hero can use it today (supported, experimental, unsupported) and the tree shows that label; several guide tracks per class as small shareable build files for players who want a suggestion or autopilot; a preview before each point ("about 9 percent faster kills here"); the game's own three Akara respecs as the safety net, proposed by the hero and decided by the player; points are never spent behind the player's back unless a track's autopilot is on.

Why it is feasible: the repo already holds game.exe's damage, mana and rate functions for 131 skills, a skill-tree model with "can add, and why not", a reserved allocate route, and the survival margin. The prototype derives the same main attack as the hand-written profiles in 23 of 30 checkpoints, and its 7 misses name exactly what the estimator still lacks (multi-strike skills, synergy lists, life costs, pet damage). So: hand-written order as the prior, estimate overrules by a clear margin, measured kill speed overrules both. About 30 to 40 cards beyond work that is needed anyway; stats in week one, skills for the two current classes in three weeks. [04](04-stat-skill-agency.md).

## The overhaul verdict

Targeted rebuild of the decision core (Proving Ground, arbiter, checkable quest chains, survival model as the one judge, a small bot API on the fork): about 64 to 90 cards over three to four weeks, mostly in shadow mode beside the old code. A full rewrite would cost several hundred cards, lose a month of fidelity evidence and fix none of the root cause. [05](05-architecture-and-overhaul-options.md) section 7.

## Prototypes (offline, not merged)

| branch | proves |
|---|---|
| `agent/fable-proto-freebuild` | rotation and warnings derived from the hero's actual skill points over `internal/power`; agreement table 23 of 30 |
| `agent/fable-proto-buildfile` | `didle.build/1` goal-list build file; exact round trip of eight shipped plan tables; text form; kolbot-list import; 17 validation codes |
| `agent/fable-proto-arbiter` | arbiter: 180 area changes with no progress under a first-match chain against 9 and a finished quest step; broke-with-stash-gold solved by a batched town visit. questspec: a static checker that reports the documented Search for Cain root causes from data alone |

Details and limits: [06-roadmap.md](06-roadmap.md) section 5.

## What I would do first

1. Build the scenario runner and make the stuck tree step its first red line (R1, R2).
2. In the same week: three five-minute lab facts that are unknown today (pointer click against interact on the tree; the per-level skill cap on d2d; the Akara respec on d2d), and agency P0 (stats to the player, legality fixes).
3. Then the arbiter in shadow and the quest checker in CI; nothing that changes behaviour merges without a green scenario.

## What I challenge

- "The bot is too slow to build" is the wrong worry; it is built at great speed and proven at almost none.
- "Fidelity first, measured by A:B soaks": the A:B has never produced a verdict; the differentials have produced about 25 merged fixes.
- "The hero should only know what is near it": nobody decided that, and it cost 17 hours at a tree.
- "Seven classes need seven tuned bots": they need about sixteen drivers and one honest table.
- "Follow the modern guides": they are written for a game with Spirit and Insight; offline 1.14d is a different, harder game for casters.

## What I could not do or find

- No game, lockstep run or emulator was run by this review; every behavioural claim about d2d is from documents and source reads.
- Forums behind login (d2jsp, MPGH, Discord), Reddit and YouTube were not reached; the maxroll planner's share format was not readable.
- Per-skill built status for all 210 skills on d2d, Hell immunity tables per area, and whether self-found Hell without ladder runewords is playable for each class: all need runs.
- Whether spending points on a hero one does not steer is fun: only a friend session answers that.
