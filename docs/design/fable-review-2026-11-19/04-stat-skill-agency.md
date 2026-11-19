# 04. Stat and skill points: how much the player decides, and how the hero copes with it

> status: REVIEW for the owner (design review 2026-11-19, card fablereview, docs only). Nothing here is decided. This is the key design question of the review. Part of [00-summary.md](00-summary.md).
>
> Tags: **[V]** read or run by me or a helper (helper A = public bot sources; helper G = this tree; prototypes are named by branch, their tests ran on 2026-11-19); **[S]** secondary; **[U]** judgment or veteran knowledge not checked against the tables.

## 0. Recommendation in ten lines

1. **Yes: the player decides every stat point and every skill point.** Recommended path: **D, "free points, honest hero"**: free allocation under the game's own rules, a hero whose play is DERIVED from the skills it actually has, capability tiers that say plainly which skills the hero can use today, and guide tracks (shareable build files, several per class) for players who want a suggestion or autopilot.
2. It does NOT add exponential complexity, because nothing enumerates builds. The hero scores its skills one at a time (about 30 per class) and fills a handful of slots (main attack, second element, aura, buffs, summons, curse, movement). Complexity is linear in skills and in about sixteen delivery kinds ([03](03-classes-plan.md) section 2).
3. Stats are almost free to hand over today: the fighter does not branch on strength or energy; only gear requirements, block, life and mana change, and the survival model already reads those.
4. Skills are the real work, and half of it exists: the game's own damage, mana and rate functions for 131 skills (`internal/power/skilldamage.json`), a skill-tree view with a "can add, and why not" verdict (`internal/skilltree`), a reserved `POST /api/skills/allocate` that answers 501, and the power-against-danger margin [V helper G].
5. The prototype (`agent/fable-proto-freebuild`) shows the derivation works and where it does not yet: it agrees with the bot's hand-written attack order in 23 of 30 table checkpoints, and the 7 disagreements are all estimator gaps (multi-strike skills counted as one strike, missing synergy lists, a life cost ignored, no pet damage) [V].
6. So the engine must be **estimate first, measurement second**: the planner proposes; measured kill speed from the fight log corrects it per skill and per hero (the survival design already has a `MeasuredDPS` term, unwired [V docs/survival/power.md]).
7. Unsupported skills are never silently "played badly". The tree shows them greyed with "your hero cannot use this yet"; the player may still buy them (it is their hero, and prerequisites are sometimes the point), with the cost shown.
8. The safety net is the game's own: one Akara respec per difficulty, three per hero [V patch 1.13 notes via helper B], offered by the hero when the viability estimate is red.
9. Idle rule: points are never spent behind the player's back unless the player chose a track with autopilot. Unspent points raise a Needs-you row and the hero keeps playing. A hero that is 5 levels behind on points is weaker, and the return screen says by how much.
10. Phases: P0 one week (stats free, verbs, legality fixes), P1 two weeks (free skills for supported kinds with preview), P2 continuous (capabilities grow with the class roadmap), P3 (import and share). Details in section 7.

## 1. What exists today

| piece | state | source |
|---|---|---|
| who decides | the plan table, entirely; `pointsStep` spends as soon as points are pending and the hero is safe | [V internal/bot/leveling.go:43] |
| how | clicks on the + buttons and skill icons, on d2d too; the fork's `stat` and `skillpt` verbs have no non-test caller | [V helper G] |
| skill page | read-only trees with prerequisites, `canAdd` and `why`, mana now and next | [V internal/skilltree, internal/watch/skillsapi.go] |
| allocate API | `POST /api/skills/allocate` reserved, answers 501; no stat endpoint; no respec | [V] |
| decision F6 (2026-11-07) | skills unlock by quest or find through a per-class table; the player may spend ONE free-choice point on an unlocked skill the build table skips. The unlock tables exist for all seven classes; nothing in production wires them | [V docs/decisions.md; helper G] |
| combat choice | the profile band's ordered attack list; first known, bound, affordable, non-immune skill wins | [V internal/combat fight.go, skills.go] |

The owner's ask today goes well past F6. That is fine; F6's Codex unlocks can sit on top of free allocation as a pacing layer (a skill must be discovered before points can go in) or be dropped. It is question Q2.

## 2. The four options

Effort in agent-cards (60 to 90 min). "Fun" is my judgment of agency per tap for an idle game [U].

| | A: fully free, general engine | B: archetype templates + edits | C: curated library + import | **D: free points, capability tiers, guide tracks** |
|---|---|---|---|---|
| player decides | every point | deviations inside guard rails | which build | every point; a track only suggests |
| engine | derives play from any allocation, all skills | per-template profile | per-build profile, hand-tuned | derives play from the allocation for SUPPORTED delivery kinds; the rest is flagged |
| first usable | never "done": every skill must work before the promise is true | 2 to 3 weeks | exists (plan tables) | P0 in a week, P1 in three |
| effort to seven classes | 150+ before it is honest | 60 to 90, then every new template is work | 20 to 30 per build, forever | 25 to 35 for the planner and UI, then the class roadmap's kind drivers (already needed) |
| risk | promises what it cannot do; players hit dead skills | "why can't I put a point there" | the player watches a build someone else made | estimator wrong for some skills (bounded by measurement and by the tier labels) |
| fun | highest in theory | medium | lowest: no ownership | high, and honest |
| composes with quest board, Ether, vault | readiness per offer changes with every point | same | static | same as A: one margin feeds all of them |
| fails how | silently bad play | frustration | boredom | visible labels and a red estimate |

A is D with the labels removed. B is the worst of both: it needs the general engine to cope with edits AND restricts the player. C is where we are; it is also what every public bot does (kolbot AutoBuild, SoloPlay stages, Koolo ordered lists [V helper A]) because their user is a botter, not a player. **D is the only option where the limit of the engine is shown to the player as part of the game instead of being hidden or imposed.**

## 3. Path D in detail

### 3.1 Three data pieces

1. **Capability table** (per skill, data): role (attack, aura, buff, summon, curse, movement, passive, utility), delivery kind (K1 to K16), tier, element, synergy names, cost notes (life cost, needs a corpse, needs a shield, cooldown). Tier is one of:
   - `supported`: a driver exists and the skill has a green Proving Ground matrix row on the current build;
   - `experimental`: driver exists, not proven; usable with a warning;
   - `unsupported`: no driver; points are legal but the hero will not use the skill (passives and pure prerequisites are `supported` by nature).
   Most of role, kind and element is derivable from `skills.txt` columns and the existing `skills.json` and `skillKinds` tables; the prototype's hand table has 731 lines for two full classes and two partial ones. The tier column must be GENERATED from Proving Ground results, not typed.
2. **The planner** (`Derive(allocation, hero) -> plan`): for every attack-role skill with points and tier above unsupported, estimate sustained damage per second against the current zone's reference pack (emulator skill data x hit count x targets hit x resist of the zone's monsters), discount by mana sustainability, pick the main attack; pick a secondary of another element when the zone has immunes to the first; fill aura, buffs, summons, curse, movement from the skills present; resolve the archetype; attach a one-line "why" to each choice and a list of warnings. Then correct the ranking by MEASURED kill speed per skill once the hero has fought with it.
3. **Guide tracks** (build files): an ordered goal list, shareable, validated, expanded deterministically. A track never forces; it fills the "suggested next point" and, if the player turns autopilot on for it, spends points while they are away.

### 3.2 What the planner prototype showed

Branch `agent/fable-proto-freebuild`, `internal/proto/freebuild` (about 1,200 lines of Go and tests plus the table; `go test` passes) [V]:

| finding | meaning |
|---|---|
| 23 of 30 checkpoints (six shipped plan tables x clvl 6, 12, 18, 24, 30) pick the same main attack as the hand-written profile band | the derivation is right where the skill is a plain missile or swing |
| Paladin at 18 and 24: it prefers Sacrifice (93) over Zeal (37) and Blessed Hammer (47) | Zeal is counted as one strike; Hammer gets no synergy; Sacrifice's life cost is not a cost. The band order holds knowledge the estimator lacks |
| Necromancer summoner: "main attack" is the staff swing over Teeth | pets have no damage in the model; the archetype is still resolved as summoner |
| Chain Lightning won at clvl 30 until marked unsupported | tiers are not decoration: a skill that is broken on the backend must be out of the ranking |
| `power`'s kind table lacks Fire Bolt, Ice Bolt and Fire Ball, so the archetype resolver fails on the commonest Sorceress attacks | a real bug in the survival model's data, found by trying off-plan builds |
| off-plan heroes behave sensibly: all-Fire-Ball Sorceress (warning: single element), twelve skills at one point (tourist), Whirlwind-maxed Barbarian (falls back to Bash and names Whirlwind as unplayed), zero attack skills (no attack) | the warnings a player needs fall out of the same pass |
| the "tourist" threshold (35 percent in main skill and synergies) fires on our own plan tables; the mana budget guess fires too often | thresholds must be calibrated on tracks that are known good, and mana needs the real regeneration rule |

Conclusion: **keep the hand-written band order as a prior and a tie-break, let the estimate overrule it only by a clear margin, and let measurement overrule both.** That is cheap, and it degrades to today's behaviour for a hero on a stock track.

Public precedent: jaenster-kolbot ranks every known skill by hits-to-kill divided by targets hit, with a x5 penalty when mana is short and for cooldown skills, hand-kept synergy and multiplier tables, and class code for curses, corpses, traps and hammers; it admits melee does not work [V helper A, [01](01-landscape-and-comparison.md) section 4]. Same shape, weaker data, no measurement loop.

### 3.3 The viability estimate (what the player sees before a tap)

One model, the survival margin, answers all of these; nothing new is invented:

| question | answer shown | from |
|---|---|---|
| "what does this point do?" | Preview: "Fire Ball 7 to 8: kills in Dark Wood about 9 percent faster; mana per kill up 4 percent" | planner before and after, zone = where the hero farms |
| "can this hero do the next quest?" | the quest board's readiness word and sentence | margin against the boss pack (exists as `Assess`) |
| "is this hero broken?" | a hero-level line: "Main attack: Zeal (strong). Second element: none. 9 points in skills your hero cannot use. Hell readiness: fire resist -23" | warnings + margin at the next act's area level |
| "why can't the hero kill X?" | "Cold immune: your only damage is cold. Options: put points in a second element, hire a mercenary, skip" | immunity cover check (exists in `internal/readiness` for two classes) |
| "should I respec?" | offered once the hero-level line is red for a set time: "Akara can reset your points (1 of 1 in this difficulty). Suggested: track X" | respec count from the quest state |

Rules for honest numbers: every preview carries the quality flag the power model already has (measured, derived, default) and says "estimate" until the hero has fought with the skill; a term that is unread (no weapon data) gives no number rather than a guess.

### 3.4 Guard rails that are fair

- The game's rules are the only hard walls: required level, prerequisites, 20 hard points, points available. **And one the repo does not know yet [U, high confidence; V that kolbot enforces it in `AutoSkill.js:54-62`, helper A]: a skill's hard points cannot exceed character level minus the skill's required level plus one.** A level-18 skill takes one point at 18, two at 19. `leveler.rule` does not check it [V internal/leveler/leveler.go:268-290], the build-file prototype found that our hammerdin table "banks" five points and spends them all at clvl 18, and kolbot refuses exactly that. Either d2d enforces the rule and the table stalls, or d2d does not and that is a fidelity deviation. A five-minute lab check decides; it affects every "bank then dump" table we ship.
- Everything else is a warning with a reason and a number, never a lock ("warn, never wall" is already the readiness rule).
- One confirmation only where a choice is hard to undo and clearly harmful: the first point into an `unsupported` attack skill, and dumping energy on a melee hero. The text says what the hero will do instead.
- No hidden correction. The hero never "fixes" a build by spending banked points on its own.

### 3.5 Stats

Free from P0. The only consumers are requirements (the advisor already says "Hand Axe: usable after +5 strength"), block (a dexterity target the planner can state: "block 62 percent with Holy Shield"), life and mana (the margin). A track's stat goals use the community form: ordered targets with a "rest" sink and a "block" target [V helper A: kolbot AutoStat]. One guard: strength below the hero's own equipped gear is impossible in the game anyway.

### 3.6 The build file

Prototype `agent/fable-proto-buildfile`, `internal/proto/buildfile` (`go test` passes) [V]:

- Format `didle.build/1`: `skills` = ordered goals `{skill, to, when?, hold?}`; prerequisites are learned automatically; `stats` = ordered `{stat, to | "rest", per?}`; optional `phases` and `respec_at` for a staged track; strict decode; names, not ids.
- `Expand` turns it into the per-level rows the leveler already consumes; `FromRows` compresses an existing table: **eight shipped tables round-trip exactly at every row**, and 89 rows become 17 to 25 skill goals and 6 to 13 stat goals (2 to 3 KB).
- A one-line text form for chat (`didle1:pal:Zeal20,Holy Shield20,...|str=60,dex=block,vit=rest`) and an importer for kolbot-style `[[skill, points, barrier], ...]` lists.
- 17 validation codes (unknown skill, wrong class, over 20, unreachable by 99, prerequisite cycle, stat budget, and so on).
- Findings: the normal30 tables quietly assume quest points; there is no dexterity "block" target yet; `main_attack` and hotkeys are deliberately NOT in the file (the engine derives them), which removes the "two sources of truth" finding of the build-flexibility audit.

Comparison of notations [V helper A, helper B]:

| notation | order of points | prerequisites | respec | validation | who writes it |
|---|---|---|---|---|---|
| kolbot AutoBuild | per level, exact | must be listed; throws on error | no | strict | botter, by hand |
| kolbot AutoSkill and AutoStat | goal list with barriers | automatic | no | none; stalls silently | botter |
| SoloPlay build module | goal list per stage + code | automatic | stages with levels, conditions as code | none | maintainer |
| Koolo leveling character | ordered id list, repeats are levels | automatic | a "should reset" function | none | maintainer [D2R] |
| community planners (URL forms) | none: end state only | n/a | n/a | n/a | players |
| our plan table | per level, exact | must be listed | flag only | checker | agent |
| `didle.build/1` | goal list, optional phases | automatic | one level per phase | strict, 17 codes | a player |

No community format records the ORDER of points in a shareable way except kolbot's two; players share end states. An importer for "end state only" builds is therefore useful: the player pastes a target allocation, the expander orders it by the planner's own preview (best next point first). That is a P3 nicety.

### 3.7 Respec policy

- Three free respecs per hero (one per difficulty), plus Tokens in late Hell. The community's leveling builds USE them: every SoloPlay class respecs once between clvl 19 and 32 [V helper A].
- Product rule: a respec is a player event at Akara ("Akara offers to reset your points"), never automatic; the hero proposes it when the hero-level line is red, with a track preselected and the preview of the result. An unused respec is shown as an asset.
- A staged track declares its respec level; with autopilot off the hero asks at that level.
- To verify before promising any of it: the Akara respec and the Token recipe on d2d [U].

## 4. How it composes with what was already decided

| existing piece | effect |
|---|---|
| Quest board (quests offered by readiness) | unchanged API; the readiness word now moves when the player spends points. A good loop: "Blood Raven: risky. +2 Zeal makes it safe" |
| Policy sheet and return screen | "points waiting: 3" is a Needs-you row; the return screen shows what the unspent points cost ("about 12 percent slower kills since level 14") |
| Ether | a natural sink that does not touch fidelity: extra preview depth, a second saved track slot, or nothing; never points or respecs for currency (F7 forbids permanent power) |
| Vault and gear policy | the same advisor voice: "this staff gives +2 to a skill your hero uses" becomes computable from the capability table |
| Codex unlocks (F6) | optional pacing layer on top; if kept, an undiscovered skill is greyed with "not discovered" before "cannot use" |
| Pacts | a "no respec" or "tourist" pact is one line of data later |
| Automation ladder (ADR 0016) | autopilot on a track is a rung the player can raise or lower per hero |
| Measured runs and gates | a measured run names its track; free allocation is a d.idle-mode feature and does not enter gates |

## 5. Cost of getting it wrong

| failure | how likely | what limits the damage |
|---|---|---|
| the estimate ranks a weak skill first | certain for some skills at first (prototype) | band order as prior; measured kill speed; tiers generated from the Proving Ground; a per-skill override row in data with evidence and expiry (the survival design's override budget) |
| a player bricks a hero | by design possible | three respecs; warnings before; the hero keeps farming a zone its margin allows (the survival model already steps down) |
| preview numbers feel like lies | if shown as exact | always "about", always with the weakest input named |
| capability labels frustrate ("why can't my hero Whirlwind") | yes | the label is a roadmap: "not yet"; the class plan says in which order drivers arrive; the player can still play the class with supported skills |
| scope creep toward option A | the main risk | the tier column is generated; a skill is `supported` only with a green matrix row; no promise beyond the table |
| per-class tuning burden returns | moderate | archetype data plus measured correction instead of per-class knobs ([05](05-architecture-and-overhaul-options.md) 3.3) |

## 6. How to test it

| level | test | where |
|---|---|---|
| unit | planner agreement table: every shipped track at clvl 6, 12, 18, 24, 30 against the profile band (golden; a change must be explained) | CI; the prototype has the first version |
| unit | off-plan zoo: 20 deliberately odd allocations per class with expected warnings | CI |
| static | build files validate against the player's skill tables; capability table covers every class skill | CI and on game-data load |
| scenario (L3 matrix) | each `supported` skill is cast in a scripted lockstep fight and its effect is read back; the row turns the tier green | Proving Ground |
| A/B (L2) | "planner on" against "band order" on the same seeds, per class: kills per minute, potions per kill, deaths; the survival A/B harness and `abreport` already do this shape | long lane, nightly |
| off-plan soak | three odd but legal heroes per class play Act 1: the bar is "no stall, warnings match what happened" | long lane |
| live | one friend spends points by hand for an evening | the only test of fun |

## 7. Phases

| phase | content | cards | exit |
|---|---|---|---|
| P0 "stats are yours" (week 1) | spend by verb instead of clicks; stat allocation API and sheet; autopilot switch per hero (default: ask); fix the per-level skill cap in `leveler.rule` and in the tables after the lab check; remove ladder runewords from table tags; add the missing skill kinds to `power` profiles | 6 to 8 | a friend can place stat points from the phone; tables are legal |
| P1 "skills are yours" (weeks 2 and 3) | capability table generated for Paladin and Sorceress; planner in SHADOW next to the band order (log both, count differences on sandbox logs); allocate API with `CanAdd` and Preview; hero-level viability line; guide tracks for the two classes as build files (2 to 3 each) | 14 to 18 | planner agrees or beats the band in the L2 A/B for both classes; then it drives |
| P2 "more of the tree" (continuous) | each class roadmap step adds drivers and turns rows green; measured kill speed wired into the ranking; synergy and multi-strike data completed in the skill-damage table; pet damage | per class, in [03](03-classes-plan.md) | per class gate |
| P3 "share" | text notation, import of kolbot-style lists and end-state builds, a track gallery in the watch UI, respec proposal flow | 6 to 9 | a build is passed between two friends |

Total new work that would not happen anyway: about 30 to 40 cards. The drivers, the survival inputs and the Proving Ground are on the roadmap regardless.

## 8. What I could not settle

- Whether the per-level cap is enforced by d2d (lab check).
- Whether the Akara respec works on d2d.
- How wrong the damage estimate is for area skills against real packs: only the L2 A/B will say.
- Whether players enjoy spending points on a hero they do not steer. It is the owner's bet; the first friend session decides more than any analysis here.
