# 06. Roadmap: phases, order, effort, risks, and the decisions only the owner can make

> status: REVIEW for the owner (design review 2026-11-19, card fablereview, docs only). Nothing here is decided or carded. Part of [00-summary.md](00-summary.md).
>
> Units: one **card** = one agent run of 60 to 90 minutes with tests (`docs/agents/rules-short.md`); **host-hours** = time on the builder's long lane or the lab. Estimates are mine [U]; ranges are honest, not padded. The bot-behaviour freeze stands until the owner answers Q1.

## 1. The order in one picture

```
week 1        week 2          week 3          week 4          weeks 5 to 8
PROVE ------> DECIDE CORE --> SWITCH CORE --> WIDEN --------> CLASSES AND ACTS
R1 Proving    R4 arbiter      R4 arbiter      R8 Act 2        R10 Necromancer
R17 input I1-I2  (R17 I3 to I7 run beside R4 and R7)
   Ground L0-L2  in shadow       live by family   through      R11 Hammerdin
R2 Cain e2e   R5 questspec    R6 survival     the board      R12 Barbarian
R3 P0 agency     checker         model on     R9 difficulty  R13 Act 3 to 5
   + legality R7 P1 agency    R7 planner         transition  ...
   fixes         in shadow       drives          flow
```

Nothing in a later column starts before the Proving Ground of column one exists. That is the whole plan.

## 2. Ranked work packages

| # | package | content | cards | host-hours | depends on | risk if skipped |
|---|---|---|---|---|---|---|
| R1 | **Proving Ground** ([05](05-architecture-and-overhaul-options.md) section 5) | L1 scenario runner on the lockstep fork from scripted saves; L2 act run (fresh hero to Andariel, 2 classes x 3 seeds) as a nightly long-lane job; report page that feeds "verified"; merge-queue hook for behaviour merges; finish loop determinism | 10 to 14 | 1 to 2 per night | fork build with lab verbs off; builder long lane | everything else repeats the Cain story |
| R2 | **Search for Cain end to end** as the first L1 set (five chunks) | lab drill of the tree click (pointer click against interact verb, object state read back); fix what it shows; level room rows on entry (K1) if the owner agrees (Q5) | 4 to 6 | 2 | R1 runner; Q5 | the sandbox hero stays stuck |
| R3 | **Agency P0 and legality** ([04](04-stat-skill-agency.md) section 7) | spend by verb; stat sheet and API; autopilot switch; per-level skill cap check and table fixes; ladder runewords out of table tags; missing skill kinds in the power profiles; bot resist cap order (add penalty, then cap) | 6 to 8 | 0.5 (one lab check) | Q1, Q2 | tables that cannot be played; wrong readiness in Nightmare |
| R4 | **Arbiter** ([05](05-architecture-and-overhaul-options.md) 3.1) | wrap the existing steps as sources; shadow log beside `runGoal`; read differences on sandbox logs; switch families one at a time (town batching first, then travel and quest, then recovery and gold farm); delete the five loop guards when their scenarios stay green | 14 to 20 | nightly L2 | R1 | every new class and act multiplies interactions |
| R5 | **Quest chains as checkable data** ([05](05-architecture-and-overhaul-options.md) 3.2) | questspec schema beside `data/questboard`; static checker in CI over every chain and backend capability set; convert Act 1 chains; interaction primitive "done when a named fact changes, identity by id" | 8 to 12 | 0 | R1 for the scenarios | Acts 2 to 5 arrive one 17-hour stall at a time |
| R6 | **Survival model as the one judge** | real inputs (weapon, defence, block read live); every area choice through the model (recovery, gold farm, quest travel) via the arbiter; class matrix A/B in L2; then `bot.zoneModel` default on per archetype; retire the level floors and bans it replaces | 8 to 12 | A/B sets on the long lane | R4 | per-class knob tuning forever |
| R7 | **Agency P1** | capability table generated for Paladin and Sorceress; planner in shadow, then driving; allocate API with preview; hero viability line; guide tracks as build files | 14 to 18 | L2 A/B | R1, R3, R6 inputs | the owner's central ask does not ship |
| R8 | **Act 2 through the quest board** | cube, staff, amulet, Arcane Sanctuary, tombs, Duriel as L1 scenarios; Act 2 mercenary hire on d2d; `Acts` default opens per act as its scenarios turn green | 10 to 16 | 4 to 6 | R1, R4, R5; fidelity A3 | the hero's world ends at Andariel |
| R9 | **Difficulty transition** | start in the next difficulty through the launcher; gate = level and resists after the penalty; immunity cover in zone choice; Pact of Nightmares hook | 5 to 8 | 2 | R6, R8 to R13 for the acts; fidelity A4 | Normal only |
| R10 | **Necromancer Summoner** ([03](03-classes-plan.md) 3.3) | summon, corpse, curse drivers live; pet damage in the power model; class gate matrix | 8 to 12 | 3 | R1, R7; fidelity A2 | no third archetype |
| R11 | Hammerdin, Sorceress Fire | stand-in-pack driver; second Sorceress build as data only | 5 to 8 | 2 | R7 | no Hell answer for the default class |
| R12 | Barbarian Frenzy and Concentrate | warcry upkeep; dual wield; Berserk fallback | 6 to 9 | 2 | R10 or R11; fidelity A1, A6 | |
| R13 | Acts 3, 4, 5 | per act as R8 | 10 to 16 each | 4 to 6 each | R8 | |
| R14 | Runeword and cube ADVISOR; shrines; status reflexes (curses, monster auras) | data and advice, the player acts | 8 to 10 | 0 | R7 | self-found heroes stay weak |
| R15 | Agency P3 (share, import), NIP importer | | 8 to 12 | 0 | R7 | |
| R16 | Blizzard and Druid (ground target), Assassin traps, Amazon bow | per [03](03-classes-plan.md) section 4 | 26 to 40 total | | fidelity A5, A7, A8, A9 | |
| R17 | **Input layer** ([07](07-input-layer.md)) | fork: `cmd select` sets the right skill, `cmd move` without the priming click, message-equivalent verbs for buy, sell, repair, waypoint, menu choice, use item; bot: aura and skill selection, town portal, points, town chores on verbs; `SemanticInput` as the only interface of behaviour packages; generated hotkey map from the planner slots; one differential per promoted verb | 10 to 14 + a fork branch of 6 to 8 verbs | emulator and lab time for the differentials | Q13; I1 and I2 can start in week 1 (they remove the popup clicks and the kick), the rest with R4 town batching | click failures keep feeding the loop guards; every new class adds popup geometry |
| F | **Fidelity lane, re-pointed** ([02](02-fidelity-and-gaps.md) section 3) | refresh the risk map by class and act; class gates in roadmap order: melee chain (A1), aura reach (A3), curse rule (A2), difficulty start (A4), then per class | continuous, 1 to 2 cards a day as today | emulator runs | the class order | numbers measured on d2d mean less with every class |

Totals: R1 to R7 (the core) about 64 to 90 cards; at the current pace of several agents a day that is three to four weeks of focused work, with the WIP limit of 3 implementation agents kept. R8 to R13 about 55 to 85 more over the following month.

## 3. First week, day by day (if Q1 is yes)

| day | work | proof at the end of the day |
|---|---|---|
| 1 | L1 runner skeleton on the builder: scripted hero to "one step before" a quest fact, run the bot under lockstep with a virtual budget, read the fact, exit code; Den of Evil as the first scenario | one green line for a quest that already works |
| 1 to 2 | lab drill on the tree (pointer click against interact, object state before and after, `debug presets` hidden column); the per-level skill cap check on d2d; Akara respec check | three facts that are U today |
| 2 to 3 | Cain chunks as scenarios, red first; fixes until green; friend sandbox gets the build only then | the stuck hero moves |
| 3 | L2 nightly: fresh Paladin and Sorceress, 3 seeds, to Andariel or a 6 virtual-hour cap; report table | the first "act run" numbers the project has had |
| 3 to 4 | agency P0: verbs, stat sheet, legality fixes | a friend places stat points from the phone |
| 4 to 5 | arbiter sources and shadow log; questspec checker in CI for Act 1 | a diff of "chain chose, arbiter would choose" on a real night's log |
| 5 | review of the week against the L2 table; next week's order | |

## 4. First month: what "done" looks like

1. A fresh Paladin and a fresh Sorceress finish Act 1 through the quest board in the nightly act run on 3 of 3 seeds, no stall over 10 virtual minutes, and the report says so without a human reading logs.
2. `runGoal`'s chain is replaced by the arbiter for town, travel and quest families; the flail guard has not fired in a week of nightly runs.
3. The player spends stat and skill points for both classes; the planner drives combat choice; the viability line exists.
4. Act 2 is open behind its scenarios; the Necromancer has started.
5. `docs/capabilities.md` is generated from the Proving Ground.

## 5. Prototypes built in this review (offline; not merged; nothing ran against a game)

| branch | package | what it proves | what it does not |
|---|---|---|---|
| `agent/fable-proto-freebuild` (last commit `7141e916b`) | `internal/proto/freebuild` | a capability table plus `Derive` over `internal/power` picks a rotation from the hero's ACTUAL skill points with reasons and warnings; 23 of 30 checkpoints agree with the hand-written bands; the 7 that differ expose estimator gaps (multi-strike, synergies, life cost, pets); off-plan heroes get the right warnings; `Validate`, `CanAdd`, `Preview` | tiers are typed from code comments, not measured; prerequisites are a fixture from memory; mana budget is a guess |
| `agent/fable-proto-buildfile` (`5b3b2867c`) | `internal/proto/buildfile` | `didle.build/1`: goal-list build file with automatic prerequisites, phases and a respec level; `Expand` and `FromRows` round-trip eight shipped tables exactly at every row; one-line text form; kolbot-list importer; 17 validation codes | never run through the real leveler or real skill tables; no "block" target yet |
| `agent/fable-proto-arbiter` (`3a5f96c98`) | `internal/proto/arbiter` | scenario simulation: the first-match chain changes area 180 times in 60 virtual minutes with no progress, the arbiter 9 times and finishes the quest step; broke-with-stash-gold becomes a withdraw step in a batched town visit; reflex and player-order preemption with resume; back-off and starvation; golden decision trace | no real sources wrapped; a three-area cycle still needs a guard; constants are placeholders |
| same branch | `internal/proto/questspec` | Search for Cain as 14 steps of data; a static checker that, given the backend capability set of 2026-11-17, reports root causes RC1, RC2, RC3 of the probe docs and nothing on the fixed set; resumption from facts at all 15 restart points; eight kinds of broken chain caught | no generator from `data/questboard`; capability sets are hand-written |

Each has a README with its V and U lines. All three test suites passed on 2026-11-19 (`go test -p 1 ./internal/proto/...` in each worktree) [V, re-run by me].

## 6. Risks to the plan itself

| risk | sign | answer |
|---|---|---|
| the Proving Ground becomes one more unverified layer | a week passes without a red-to-green story | its first job is R2 (a real stuck hero); if it cannot reproduce the stall it is not done |
| arbiter shadow never ends | differences are "analysed" | dated switch per family; the scenario set is the judge |
| the fork drifts from upstream while it grows bot verbs | rebase pain | verbs are read-only and small; offer them upstream behind the existing flag |
| self-found Hell without ladder runewords is not fun or not possible for some classes | L2 runs at Hell area levels stall with typical gear | Q4 (enable ladder words in d.idle mode) is the lever; decide with data from R9 |
| agents keep producing breadth | cards without a scenario | the merge rule of [05](05-architecture-and-overhaul-options.md) section 5; the WIP limit |
| estimator trust | a player follows a preview that was wrong | "about", quality flags, measurement loop, tiers from results |
| builder load | merge gates wait | L1 sets chosen by touched files (minutes); L2 only on the long lane at night |

## 7. Decisions for the owner

Each is a concrete question with my recommendation (Q13 was added with the input audit and sits before Q12 in the table). None is recorded anywhere but here; the guard files (`docs/decisions.md`, ADRs) are not touched by this review.

| # | question | recommendation | what a "no" changes |
|---|---|---|---|
| Q1 | **Lift the bot-change freeze only for work that comes with a Proving Ground scenario, and build the Proving Ground first (R1, R2)?** | yes | without it I would advise against starting any new class or act |
| Q2 | **Stat and skill agency: path D** (free points, capability tiers, guide tracks, hero asks and never auto-spends unless a track's autopilot is on)? And: does the Codex unlock rule (F6: skills unlock by quest or find, one free-choice point) stay as a pacing layer on top, or go? | D; drop F6's "one free point" (it is superseded); keep discovery only if it tests as fun | B or C keeps today's tables and adds choices inside them |
| Q3 | Default when the player is away and points are waiting: **bank and ask** (hero is weaker meanwhile, shown on the return screen), or follow the chosen guide track? | bank and ask by default; autopilot is a per-hero switch | |
| Q4 | **Ladder-only runewords** (Spirit, Insight and 21 more) are not makeable in vanilla offline 1.14d. In d.idle mode, enable them in the fork (as PlugY players do), or stay vanilla? Gates and fidelity runs stay vanilla either way | decide after the first Hell-level act runs; lean to enable in d.idle mode as an explicit, visible rule, because every modern guide and the player's memory assume them | vanilla: casters have no Insight; class plans in [03](03-classes-plan.md) already assume vanilla |
| Q5 | **What the hero may know**: K1 (level room layout and quest-object rooms on entry, read-only fork verb), instead of today's accidental "only what is near"? And the general rule "the fork may expose any engine fact as a READ; reads never make a run CHEATED"? | K1 and the rule | K0: an explore-to-find primitive for every quest object, slower and more fragile |
| Q6 | **Arbiter**: replace `runGoal`'s chain by the intent arbiter, shadow first, family by family? | yes | keep patching; I expect the guard count to grow with every act |
| Q7 | **Fidelity acceptance**: replace the statistical A:B verdict (6 game.exe and 6 d2d lanes at 20 percent, ADR 0015) by "every rule on the hero's path for the gated class and act has a 0-difference differential or a listed shim", and stop planning to PLAY on the Game.exe backend? | yes | the A:B stays INCONCLUSIVE and the Wine lanes stay a cost |
| Q8 | **Class order**: Sorceress second build, then Necromancer Summoner, Hammerdin, Barbarian, ground-target builds and Druid, Assassin, Amazon? | yes | the friend-preferred class can jump the queue at the cost of its new delivery kinds |
| Q9 | **Unit of work**: vertical cards ("the hero does X on d2d, scenario green") instead of layer slices, even when a card is larger than today's 400-line guide? | yes | |
| Q10 | **Respec as a player event**: the hero proposes the Akara reset when viability is red, the player decides; staged guide tracks may plan one? | yes | no-respec tracks only: slower Normal for most classes |
| Q11 | Unsupported skills: **allowed with a clear "your hero cannot use this yet" label** (my recommendation), or refused until supported? | allowed, one confirmation for attack skills | refused: simpler, less the player's hero |
| Q13 | **Input rule**: "the fork's action verbs are the game's own client messages; the bot prefers verb, then key, then verified click", and the item, gold and identify verbs move from lab to input after a differential each? | yes | hotkeys only: then d2d needs real key bindings and a bind bootstrap, and town stays click-driven |
| Q12 | May the review's prototypes be turned into real cards (arbiter, questspec, freebuild planner, build file) by the coordinator, on the order of section 2? | yes | |

## 8. What I would do first, in one sentence

Build the scenario runner, make the stuck hero's tree step the first red line on it, and do not merge anything that changes behaviour until that line is green.
