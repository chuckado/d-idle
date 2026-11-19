# 02. Fidelity and gaps: what the idle hero will meet between Act 1 Normal and Baal in Hell

> status: REVIEW for the owner (design review 2026-11-19, card fablereview, docs only). Nothing here is decided. Part of [00-summary.md](00-summary.md).
>
> Tags: **[V]** read by me or by a helper agent of this review from the primary source named (file path, URL, a command that ran); **[S]** another doc's claim, not re-checked; **[U]** unverified or judgment. Helper research is cited as "helper E" (d2d coverage, read from the upstream gap map and a shallow read-only clone on 2026-11-19) and "helper G" (code map of this tree at `989432fcd`). Target is classic LoD 1.14d, offline single player.

## 1. The two lists must be kept apart

"Fidelity" in this project means two different things that need different work:

1. **d2d is not game.exe** (the engine plays differently). Fixed upstream with a differential. The lane works: about 25 merged PRs, an emulator that runs game.exe's own functions, an independent verifier [V docs/upstream/pr-queue.md].
2. **The bot does not know a mechanic** (the engine may be perfect; the hero still stands in front of a cube it cannot use). Fixed in this repo.

The owner's rule for this review is "bot capability first". My finding is that the order of the first list should be DERIVED from the second: a d2d deviation matters when a class, act or mechanic on the roadmap reaches it. Today the fidelity queue is ordered by what the Act 1 Sorceress and Paladin touched (cast ticks, hit recovery, warp arrival). That was right for Act 1. It is the wrong order for seven classes and three difficulties.

## 2. Where d2d stands today (it moved a lot in three weeks)

| measure | 2026-10-30 (our risk map) | 2026-11-19 (helper E) |
|---|---|---|
| functions verified + ported | 22.9 percent | 29.4 percent by count, 45.3 percent by code bytes [V gaps.json at upstream `d5de50ca`] |
| untouched | 66.8 percent | 28.3 percent (mostly client-only code now) [V] |
| listed gaps | 489 | 125 [V] |
| playable content | Act 1 focus | README: all five acts play through, 27 quests to Baal, Nightmare and Hell populate levels, difficulty picker at character select [V upstream README] |

Module "done" shares that matter to us [V helper E, gaps.json]: AI 90 percent, QUESTS 67 (but only 89 of 617 functions verified, 323 ported from a read), DRLG 55, SKILLS 51 (163 untouched), ITEMS 51, MISSILES 25 (44 of 59 have an oracle but d2d's code does not cite it), STATS 23.

Our own documents are behind this. `docs/capabilities.md` is dated 2026-11-01 and `docs/research/d2d-fidelity-riskmap.md` 2026-10-30; `tools/riskmap/riskmap.py` still runs but its gap column reads 0 for every area because `dependencies.json` maps gaps by old file and line [V helper E ran it]. **Recommendation:** refresh the risk map as the first fidelity job (2 agent-hours), keyed by class and act instead of by Act 1 metric.

## 3. Ranked list A: d2d deviations the hero meets on the way to Hell

Rank = (how soon the roadmap of [06-roadmap.md](06-roadmap.md) reaches it) x (blocks or only skews). "Who": UP = upstream PR with a differential, FORK = our fork (verbs, flags), BOT = this repo.

| # | deviation | evidence | effect | who | when it matters |
|---|---|---|---|---|---|
| A1 | Player melee chain not in upstream main yet: melee reach, attack-sequence rate (Jab, Frenzy and the other sequence attacks are 30 points slow), Smite damage and reach | `fid/melee-reach-r1`, `fid/seq-ias-r1`, `fid/smite-r1` built, not merged [S pr-queue rows 3h, 3i; helper E] | skews kills per hour for every melee hero (Paladin now; Barbarian, kicker Assassin, melee Amazon later) | UP (already in the verifier queue) | now |
| A2 | Curse "put" rule (slot, refresh, length) differs in 54,685 of 100,000 sequences | covstates2 D1 to D5, queue row 8 NOT READY [S] | Necromancer curses, and every curse a monster casts on the hero in Nightmare and Hell (Amplify Damage, Lower Resist, Decrepify) | UP, needs a new differential | before the Necromancer; before Nightmare |
| A3 | Auras do not reach the mercenary or pets; Prayer heals the player only; mercenary does not follow Teleport | palaura-trace D1, sorc-act-gate T5 [S] | the Act 2 mercenary's aura and the Paladin's party auras are the standard answer to several walls ([03-classes-plan.md](03-classes-plan.md)); summoner Necromancer with Might merc depends on it | UP | before Act 2 mercenary matters (Paladin Act 2) |
| A4 | Starting or continuing a game in Nightmare or Hell through our launcher is unsupported (`debug difficulty` is a lab cheat that moves the hero to town and rebuilds levels) | docs/d2d-backend.md [S]; upstream has a difficulty picker at character select and reads the save's difficulty bits [V helper E, apps/d2d/frontend.cpp, main.cpp] | blocks the difficulty transition | BOT launcher flag or FORK start option | before the first Baal kill |
| A5 | Skill delay not enforced (Blizzard, Meteor, Hydra, Fire Wall); refused Teleport still costs mana | DRILLS15 M6 FAIL; current state unknown [S, U] | inflates every timed-spell Sorceress and Druid build on d2d | UP | before the second Sorceress build (Blizzard, Meteor) |
| A6 | War Cry and Battle Cry hit about 8 times too often (NextHit); Whirlwind radius 5 cells instead of 5 subtiles; Howl fear length | skbarb D1 to D6 [S] | Barbarian crowd damage far too high on d2d; a Whirlwind verdict on d2d would be worthless | UP | before the Barbarian leaves Act 1 |
| A7 | Assassin: 0 of 30 skills run-tested against game.exe (traps, sentries, shadows, charge-ups built from a read only) | rule-coverage ledger skills.ass 0 percent [V helper E read docs/coverage/rule-coverage.md] | the class is unmeasured, not known broken | UP oracle pass, then drills | before the Assassin (last third of the class order) |
| A8 | Druid effects without a verdict: Tornado and Twister flight, Volcano, Arctic Blast, Fire Claws, shapeshift form rules; Rabies spread and werebeast form requirement not built | skdruid-trace "effects NOT run" [S]; upstream skills.md not-built list [S, dated 2026-09] | blocks a wind or shapeshift Druid; the fire Druid needs Firestorm, Fissure, Volcano checked | UP | before the Druid |
| A9 | Pierce (Amazon passive and piercing items) incomplete | skamazon D2; ponytail `missiles.cpp:813` [V helper E] | Amazon points in Pierce are worthless or inflated | UP | before the Amazon |
| A10 | Monster missiles: 371 of 684 missile rows untested; the "no physical damage cannot be blocked" fix built, not merged | rule-coverage missiles 43 percent [S] | unquantified effect on deaths in Hell Acts 3 to 5 (ranged and lightning packs) for all classes | UP | before Hell |
| A11 | Summon specials: Blood Golem life link and Iron Golem item not built; Fenris rage, Valkyrie passives, Shadow gear, Raven hit count untraced | ponytails `fight.hpp:1279`, `summons.cpp:249,280`, `pets.cpp:591` [V helper E] | skews pet builds; Clay Golem and skeletons (the summoner's core) are built and the pet think is oracle-run | UP | Necromancer second build, Druid summoner |
| A12 | Acts 2 to 5 quests "never played end to end by clicks" upstream; only 5 Act 1 area layouts compared in our ledger; boss fights for Andariel, Summoner, Izual, Radament, Council have no per-boss verdict | upstream todo.md row 1 [V helper E]; fidelity-matrix MA04, MA13 [S] | stuck states at quest objects are LIKELY (see the Search for Cain study in [05](05-architecture-and-overhaul-options.md)); this is the largest unknown for "all acts" | our own lockstep drills find them; UP fixes | every act, starting with Act 2 |
| A13 | Treasure classes compared seed-for-seed for Act 1 only; vendor stock and refresh never compared | rule-coverage F16; matrix VN04, VN05 [S] | economy and loot (runes, uniques, what can be bought) may be off past Act 1 | UP oracle | Nightmare and later, where drops become the game |
| A14 | `/players N` does nothing on d2d | docs/probes/pactnative.md [S] | blocks Pact and pacing designs that use it | FORK | when Pacts resume |
| A15 | Poison replacement rule differs (5,417 of 39,994); damage over time in whole points instead of 256ths; player magic resistance missing against monster magic blows | verifycovstates, decomp37, decomp39 [S] | poison builds and Hell poison deaths; Act 5 succubi | UP | Nightmare and later |
| A16 | Open bot shims in `docs/d2d-deviations.md`: R01 Lightning cast length (probably fixed upstream by the sequence table, not re-checked), R03 mana payment tick, R06 waypoint coordinates in Acts 3 to 5, R08 right-skill select | [V docs/d2d-deviations.md; helper E section 4] | shims stay until the lab binary is rebased and each is re-checked | BOT cleanup after a fork rebase | continuous |

Not gaps (so nobody spends time on them) [V helper E, upstream stubs.md]: the Pandemonium (uber) portals are `return 0` stubs in game.exe 1.14d itself for the offline game; `MonItemPercent` is never read by game.exe; the thief unique mod is unreachable in 1.14d.

### What I recommend for the fidelity lane

1. **Tie the queue to class and act gates.** For each step of the class roadmap define a "class gate": the skill matrix of the build's minimum skill set (precedent: `docs/research/pal-skill-matrix.md`, `sorc-skill-matrix.md`) is green or yellow-accepted BEFORE the bot work for that class starts. A13 and A12 get "act gates" the same way.
2. **Stop measuring fidelity by statistical A:B soaks.** The A:B verdict has been INCONCLUSIVE at every attempt (k=2 pilot, single seeds) [V docs/plan/STATUS.md], and seed-for-seed comparison of whole runs is impossible because the full bot loop is not deterministic [V docs/d2d-determinism.md, docs/d2d-loop.md]. The function-level differential (game.exe's own code in the emulator against d2d, tens of thousands of cases with 0 differences) is far stronger evidence per hour. Proposed text for the owner: the G-D A:B verdict is replaced by "every rule on the hero's path for the gated class and act has a 0-difference differential or a listed shim". This is a change to a user decision (ADR 0015 acceptance); it is question Q7 in [06-roadmap.md](06-roadmap.md).
3. **Use our own lockstep drills as the finder for A12.** Upstream says nobody played an act by clicks. We are the project that does. Every quest chain that passes our Proving Ground ([05](05-architecture-and-overhaul-options.md) section 5) is also a fidelity result worth offering upstream.

## 4. Ranked list B: mechanics the bot does not have

Evidence is helper G's grep of this tree unless noted; "code, never live" means files exist and unit tests pass but no run on any backend is recorded.

| # | mechanic | state in this repo | why it matters for Normal to Hell | size (agent-cards of 60 to 90 min) |
|---|---|---|---|---|
| B1 | **Leaving Act 1.** `bot.Config.Acts` defaults to `[1]` with the comment "Act changes need quests, which the bot cannot do yet" [V internal/bot/config.go:24]; Act 2 to 5 quest handlers, boss scripts and board data exist, none drilled | code, never live | everything | the Proving Ground first, then 6 to 10 cards per act of drill-and-fix |
| B2 | **Quest chains with objects** (Cain, Horadric Staff and the orifice, Khalim's Will and the orb, Hellforge, Anya, Ancients). Search for Cain failed in five ways on first contact [V docs/probes/search-for-cain-2026-11-17.md] | Act 1 partly live, rest never | each act boss is behind one | part of B1; the questspec checker of [05](05-architecture-and-overhaul-options.md) |
| B3 | **Difficulty transition** (when to go, how to start the next difficulty, the resist penalty) | `ensureDifficulty` and `progression-nightmare/hell.json` exist; "OFFLINE ONLY, treat as U" [S docs/nightmare-campaign.md]; the bot's resist model caps at 75 BEFORE adding the penalty while game.exe adds then caps [S decomp37 B, helper E] | a hero with raw 90 fire resist has 50 in Nightmare, not 35: the readiness words are wrong until fixed | 3 to 5 cards plus A4 |
| B4 | **Immunities in Hell**: the fighter drops immune targets and logs; `PickByResist` switches element in Nightmare and Hell (code, never live); there is no "this zone is mostly immune to my damage" input to zone choice | partial | the central Hell problem for every single-element build | part of the survival model (danger profiles already carry resists) |
| B5 | **Mercenary of Acts 2, 3, 5**: hire path dormant on d2d (the reader has no `Offers`) [S docs/gear-policy.md]; aura preference table exists; the free Rogue, revive and potions work | partial | the Act 2 mercenary is the standard veteran answer to most walls | 4 to 6 cards |
| B6 | **Status awareness on the hero**: Amplify Damage, Iron Maiden and poison are read; frozen, chilled, slowed, stunned, Decrepify, Lower Resist, Conviction and other monster auras are not ("which aura is not stored") | partial | Koolo ships curse and aura chickens as config [V survey 5.2]; Conviction and Lower Resist void the resist check; Decrepify and Holy Freeze halve a melee hero's output. Note: the Oblivion Knights' Iron Maiden was removed in patch 1.13c, so that classic melee killer is NOT in 1.14d [V helper B, patch 1.13 notes via the Wayback copy of diablo2.diablowiki.net/Patch_1.13]; our code still special-cases it | 3 to 4 cards (states are in the game tables) |
| B7 | **Runewords** | `data.DB.Runewords` is parsed; nothing plans or proposes one; the plan tables cite Spirit and Insight, which are ladder-only and NOT makeable in vanilla offline 1.14d [V docs/sme/builds-gear.md section 0, from this install's runes.txt; `grep -l Spirit internal/plandata` finds five tables] | data error plus missing feature | Stealth, Lore, Ancient's Pledge, Smoke, Rhyme are how a self-found hero gets resists and cast rate. Because the PLAYER equips (gear policy), the feature is an ADVISOR ("you hold Tal and Eth: Stealth in a 2-socket body armor"), not automation | 3 to 4 cards; fix the tables first (1 card) |
| B8 | **Horadric Cube** | gem and rune upgrades only (`internal/cubejobs`); the quest cube exists | rejuvenation potions, socket recipes, Token of Absolution (the only repeatable respec) | 2 to 3 cards |
| B9 | **Quest rewards that are choices**: Larzuk socket, Charsi imbue, Anya resist scroll, Lam Esen, Alkor, personalize | board rows exist for Acts 1 to 5 as data [S]; hand-in is a player tap | resist scrolls (+10 all per difficulty) and the socket are real power | falls out of B1 and B2 |
| B10 | **Shrines** | a decision point `t0.shrine` and the skill-shrine state; no interaction code | experience and skill shrines are free power; some shrines hurt | 2 cards |
| B11 | **Gambling** | vendor ids and a d2d verb field; no routine; dropped as a sink by decision F8 | optional; the late-game gold sink for a player who equips | defer |
| B12 | **Charms** | pickup group and a `charmroom` rule; no evaluation | inventory space against charms is a player decision in an equip-yourself game; needs advisor text only | 1 to 2 cards |
| B13 | **Weapon swap, Call to Arms style prebuffs** | only "undo an accidental swap" | Battle Orders prebuff is a late Hell luxury | defer |
| B14 | **Skill delivery kinds the fighter has no driver for** (details in [04](04-stat-skill-agency.md)): shapeshift, martial-arts charge and release, Whirlwind as a movement attack, stand-in-the-pack casting (Blessed Hammer), thrown-weapon quantity | none | decides which builds a class can offer | per class, in [03](03-classes-plan.md) |
| B15 | **Stat and skill point spending by verb**: the leveler clicks the + buttons and skill icons in the UI on d2d too, although the fork's `stat` and `skillpt` verbs exist with no non-test caller [V helper G: internal/d2dctl/verbs.go:203,206] | works but fragile | a player-facing allocate API needs a dependable primitive | 1 card |
| B16 | **Special events**: Diablo Clone needs Battle.net; the Pandemonium portals are stubs offline [V helper E]; the Cow Level works through the cube | n/a | nothing to build except the cow portal as a board quest | 1 card, late |

## 5. A design gap that is neither list: what the hero is allowed to know

Every injected bot in the survey reads the whole level's preset units the moment it enters an area (D2BS `getPresetUnits`), so "find the Inifuss tree" is a lookup, not a search [V survey 2.1; helper A will confirm the kolbot script]. Our hero learned about the tree only when it walked within about 13 to 18 cells, because the d2d `debug presets` reply lists only rooms that are built around the hero [V docs/probes/search-for-cain-l0-2026-11-18.md]. The result was 17 hours of exploring [V docs/probes/treefix3-2026-11-19.md].

This is a choice the project never made on purpose. We own the fork. Three options:

| option | what the hero knows | consequence |
|---|---|---|
| K0 (today, by accident) | objects near the hero only | every quest object in a large random level needs an explore-to-find primitive with budgets and progress rules; slow and fragile, but it looks like a player searching |
| K1 (recommended) | the level's room graph and quest-object ROOMS on entry (the engine knows them when the level is generated; the l0 probe saw `room <type> x y` rows already) | the hero walks straight to the right room and still has to fight its way there; equivalent to a player who knows the map tiles. Needs a fork verb or flag that lists room rows for the whole level, read-only |
| K2 | everything a maphack shows (all presets, all monsters) | fastest; the watch UI's "information unlocks" (decision F13) lose their meaning for the hero |

K1 is a read, never a write, so the crutch rule (ADR 0014) is not touched; it is instrumentation of knowledge, and it should be recorded as a decision (question Q5).

## 6. What I could not find or verify

- Per-skill built or not-built status for all 210 class skills in d2d: not machine-readable; upstream's lists are mostly dated 2026-09 and the `debug unbuilt` verb needs a running game [helper E]. A 2-hour lab job (run `debug unbuilt` per class on the current fork build) would close this and should be the first step of each class gate.
- Whether the ledger's November deviations (Meteor and Fire Ball bursts, Teleport targets, Chain Lightning length) are merged upstream [U].
- Nightmare and Hell experience and drop numbers against game.exe: only Act 1 Normal is seed-compared [S].
- I did not run any game, emulator or differential in this review.
