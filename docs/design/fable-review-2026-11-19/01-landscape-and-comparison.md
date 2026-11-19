# 01. The landscape: what public Diablo II bots do, and where d.idle stands against them

> status: REVIEW for the owner (design review 2026-11-19, card fablereview, docs only). Nothing here is decided. Part of [00-summary.md](00-summary.md).
>
> Input: the stage-1 survey `docs/research/public-d2-bots-2026-11-19.md` (branch `agent/botsurvey`, about 46 projects). This page does not repeat its inventory. It (1) corrects the survey where a primary-source re-read disagreed, (2) compares by topic, (3) says what to adopt and what not.
>
> Tags: **[V]** read from the primary source by a helper of this review (helper A: read-only re-read of the shallow clones, with file and line; helper B: about 60 web fetches) or by me; **[S]** secondary; **[U]** unverified. Licence rule: ideas and data formats only; nothing below is copied code. D2R-only facts are marked [D2R].

## 1. Corrections to the survey (verify before relying)

| survey claim | what the source says | consequence |
|---|---|---|
| d2bsng ships a deterministic simulator whose hook list is "a ready checklist" | the simulator is built from a separate PRIVATE repository; the public repo has ten host hooks and one paragraph; clock, seed and stepping are undocumented; "hashes results for determinism" is an inference [V helper A: d2bsng `CLAUDE.md`, `docs/host_hooks.md:33-54`] | it is a precedent for the idea, not a design to study; our lockstep fork and journal are further along in the open ([05](05-architecture-and-overhaul-options.md) section 5) |
| jaenster-kolbot `AutoSkill` picks skills automatically | `AutoSkill.js` only spends points; the automatic CHOICE is `GameData.monsterEffort`, a damage model (credited in its header to another kolbot author) plus a "lowest effort wins" loop [V `modules/GameData.js:941-1073`, `unit/Attack.js:184-358`] | the nearest public precedent for a general skill engine exists and is described in section 4 |
| Kolbot-SoloPlay has 30 builds and respecs "once Enigma is owned" | 37 final-build files; every class has one or two EARLY respecs by level (Paladin 19; Sorceress, Necromancer, Druid 26; Amazon and Barbarian 30; Assassin 32; second respec Sorceress 63, Amazon 64, Barbarian 74) and each needs the Den of Evil reward [V `BuildFiles/<class>/<class>.js:9-10`, `Core/Quest.js:380-430`] | the community's answer to "leveling build differs from the final build" is the free Akara respec, used on purpose |
| Koolo chickens on 4 monster auras and needs the player to bind skills | 7 auras (adds Blessed Aim, Holy Freeze, Holy Shock); leveling characters bind their own skills [V `internal/chicken/chicken.go:13-73`, `character/sorceress_leveling.go:649-700`] [D2R] | minor |
| kolbot town order | heal, identify, clear inventory, town portal tome, potions, identify tome, shop, keys, repair, gamble, revive merc, stash, quest items [V `core/Town.js:184-200`] | minor |

Web pass beyond GitHub [V/S helper B]: the forum scene is "kolbot all the way down"; no Go, Rust or Python bot for 1.13 or 1.14 beyond the survey surfaced; **no reinforcement-learning or LLM agent for Diablo II was found** (only the Diablo 1 projects the survey lists). One browser Diablo-like with auto-battle (Abyssfire, all rights reserved) lets the player decide exactly: skill tree, a skill PRIORITY list, gear, difficulty, mercenary. Not reached: d2jsp, MPGH, Reddit, YouTube, Discord (login walls or blocked).

## 2. Comparison by topic

Legend: kolbot = the maintained trunk; SoloPlay = its auto-leveling layer; Koolo = the newest surviving Go copy [D2R].

| topic | best public practice | d.idle today | verdict |
|---|---|---|---|
| **Game access** | in-process engine API (D2BS family): `getUnit`, `getPresetUnit`, object `mode`, `clickMap`, packets, the game's own pather. Koolo: memory read plus input, map from an external generator. Vision-only bots exist and are slow | control socket to our own d2d fork (dumps, `cmd` verbs, lab verbs), and a frozen memory-and-input backend for Game.exe. We hold ourselves to "what is near the hero" and UI clicks even where the fork could answer directly | **behind where it matters** (level presets, object state, item identity) and it is self-inflicted: see [02](02-fidelity-and-gaps.md) section 5 and [05](05-architecture-and-overhaul-options.md) section 4 |
| **Headless, faster than real time** | d2bsng's private simulator on the real binary, 157 to 175x claimed | lockstep fork, 46 to 62x with the bot attached, documented journal and divergence finder [V docs/d2d-loop.md, docs/d2d-determinism.md] | **ahead in the open**, behind in use: it is not a merge gate |
| **Tests** | d2bsng: 188 unit tests (pathing on captured grids, frame loop on fakes). kolbot, SoloPlay, Koolo: none to speak of | 156,912 lines of tests, fakes for every seam, a drill suite, a differential lane against game.exe | **far ahead** in volume; the tests prove the code against fakes, not the hero against the game |
| **Fidelity discipline** | n/a (they run the real client) | emulator differentials, about 25 upstream PRs | unique |
| **Build definition** | kolbot AutoSkill `[skill, hardPoints, barrier]` with automatic prerequisites, AutoStat `[stat, target]` with an "all" sink and a "block" target; SoloPlay one module per build stage (skills, stats, gear rules, merc, attack slots, respec condition); Koolo an ordered list of skill ids where repeats mean levels [V helper A sections 1, 4, 6.5] | per-level plan table (`internal/plandata/build-*.json`) + combat profile bands + scattered class traits; 20-odd tables for all seven classes | **equal in data, behind in proof**: none of ours beyond two has played. Their goal-list form is shorter and more robust than a per-level table. See [04](04-stat-skill-agency.md) and the build-file prototype |
| **Skill choice in combat** | kolbot: seven configured slots (pre-attack; boss and normal x timed and untimed; secondary if immune), per-class attack files (Necromancer 579 lines of curses, army, corpse explosion; Paladin attack-plus-aura pairs and hammer positioning; Assassin trap layout; Barbarian Whirlwind special case). jaenster: damage-estimated choice | profile band = ordered attack list; first affordable non-immune wins; per-class upkeep code exists for Necromancer, Druid, Assassin, Paladin auras (never live) | **behind**; and the choice is the heart of the agency question |
| **Immunity** | `checkResist` with Lower Resist, Conviction and Infinity arithmetic; `SkipImmune` sets per build and difficulty; "tele-stomp" so the mercenary kills what the hero cannot [V helper A 2.2] | immune targets ignored; element switch by damage factor in Nightmare and Hell (code, never live) | behind; ours is at least data-driven |
| **Item filter** | NIP: a predicate language over item properties and stats, two-stage for unidentified items, a community corpus of about 22,000 lines; SoloPlay adds `[tier]` scoring where the right-hand side is arbitrary JavaScript | loot rules engine: ordered JSON rules, 58 fields, three-valued logic (unknown is not false), scopes ground, pack, stash, per-class overlays, shadow mode, no equip action by construction [V docs/lootrules.md] | **ours is the better design for this product**; do not switch (section 3) |
| **Auto-equip** | SoloPlay scores every item per slot with weights per class and merc (`DynamicTiers`), never replaces what it plans to socket | the bot never equips (user decision 2026-11-10); advice text only | a product choice, not a gap. The scoring idea feeds the ADVISOR |
| **Town routine** | one fixed ordered pass per visit; everything that can be done in town is done in that pass | chores planned per trip reason, with cooldowns, separate repair and belt trips, separate claim trips | **behind in shape**: their single batched visit is exactly the "batching" rule of the arbiter ([05](05-architecture-and-overhaul-options.md) 3.1) |
| **Gold shortage** | SoloPlay: a formula for "broke" by level; when broke, pick up all gold and run a short list of cheap farms; can step DOWN a difficulty for gold [V helper A 4.3] | recovery state with its own area pick, poverty mode, kit gate; went broke with gold in the stash | behind; theirs is simpler and it counts all gold |
| **Survival** | two numbers everywhere (leave at 30 percent life, drink at 75), town chicken, merc chicken, dodge, Koolo's curse and aura chickens | reflex layer plus the power-against-danger model in shadow | **ahead in ambition**, not yet in effect; adopt the curse and aura reflexes |
| **Run list and leveling route** | SoloPlay `SoloIndex`: about 50 scripts in act order, each with `preReq`, `skipIf`, `shouldRun` predicates over quest state, level bands and class; Koolo: `leveling_act1..5.go` | level plan v3 (Paladin active), campaign spine, quest board (player-accepted) | different product: their hero decides, ours asks. Their predicates are the cheapest source for our readiness floors and farm suggestions |
| **Difficulty transition** | SoloPlay: level cap per class (for example Paladin 33 and 65, Amazon 33 and 70) AND fire, cold, lightning resist above 0 after the next penalty, or over-level by 2 (Normal) or 5; Koolo: clvl 41 and 70 plus Hell resist floors, with a fall back to Nightmare when broke [V helper A 4.3, 6.3] | data rows, never run | adopt their gate as the first readiness rule for the Pact of Nightmares |
| **Mercenary** | per-build wanted merc (Holy Freeze guard from Nightmare for most, Might for Necromancer and Barbarian), temporary Defiance guard in Normal, hire rule with a gold formula [V helper A 4.5] | free Rogue, revive, potions; Act 2 hire dormant on d2d | behind |
| **Quests** | all acts scripted; quest objects by preset; success by object state; items by class id | Act 1 partly live; Acts 2 to 5 as data and handlers, never run | far behind in coverage |
| **Multi-class coverage** | kolbot 7 classes at endgame; SoloPlay 7 classes 1 to 99; Koolo 6 leveling classes [D2R] | 2 classes in Act 1 | far behind |
| **Player UI and product** | config files and a manager window | phone watch screen, quest board, vault, return screen, policy sheet | **alone in the field** |
| **Honesty of claims** | README claims | V/S/U tags, dated probes | alone in the field |

One-line reading: **the public bots are wide and shallow where we are narrow and deep.** They cover the whole game with almost no verification and no product; we have a product shell, a fidelity lane and a test culture around a hero that has not left Act 1. The comparison says where the next months should go.

## 3. Item filters: NIP against our loot rules

| | NIP (kolbot, Koolo, D2NG) | SoloPlay `[tier]` | d.idle loot rules |
|---|---|---|---|
| form | one text line: properties `#` stats `#` quantity | third section, right side is JavaScript | ordered JSON rules with a condition tree |
| unidentified items | a property match with a failing stat part on an unidentified item means "identify, then ask again" | same | three-valued: an unreadable field makes the rule UNKNOWN, with `onUnknown` per rule |
| actions | keep or not (plus quantity) | rank for auto-equip | pickup, keep, keep for merc, stash, sell, drop, identify, ask, vault, ignore |
| scope | one predicate for every place | equip | ground, pack, stash as separate stages |
| community corpus | about 22,000 lines in the pickits repository, no licence on most files [V helper A 8] | per-build gear lists | none |
| safety | n/a | code execution in data | strict decode, no code |

Recommendation: **keep the loot rules engine; add a one-way NIP IMPORTER** (property and stat predicates with arithmetic map onto our leaf conditions; quantity maps to `owned`). It is about 2 to 3 cards including a parser test on the public grammar guide. That lets a player paste a line they found, without making NIP our format and without touching unlicensed corpora (the player brings the text). Do not adopt `[tier]` as code-in-data; the scoring idea belongs in the gear advisor as a weights table.

## 4. The one public "general skill engine", read closely

jaenster-kolbot's attack loop [V helper A section 3]:

1. For the target, compute for EVERY skill the hero has: damage from the skill table (staged per-level growth), times synergies (a hand-kept table), times mastery, with skill-specific multipliers (Blizzard x2, Frozen Orb x5, five traps, a full Static Field model), against the monster's resist after Amplify, Conviction and pierce.
2. Count the targets the skill would hit (neighbours within the skill's radius, capped per skill).
3. Rank by `ceil(monster life / damage) / targets hit` (hits to kill, lower is better); multiply the effort by 5 when mana is short and by 5 for cooldown skills; exclude cooldown skills while on cooldown.
4. Cast the best one whose mana cost is affordable; class code handles what a formula cannot (Necromancer: Bone Armor, Lower Resist on everything, corpse explosion; Paladin: Holy Shield, Conviction when the damage is elemental, hammer position; Assassin: keep five traps).

Its own README says it does not work for melee; summons, warcries, bows, martial arts and shapeshifting are not modelled; mana efficiency is "not fully functional"; it ranks by hits and ignores cast time. It has no tests and is a 2021 snapshot.

What this tells us:

- **A damage-ranked choice over the hero's actual skills is feasible and was done by one person in about 1,600 lines of script.** The combinatorics never appear: skills are scored one at a time.
- **We already own the better half.** Their damage model is hand-written arithmetic from the skill table with a hand-kept synergy list; ours is game.exe's own functions run in an emulator for 131 skills with synergy sets and cast and attack frame tables, with melee weapon skills included when the weapon is read [V internal/power/skilldamage.go].
- **The other half is not a formula.** Everything their model lacks is a DELIVERY problem (how to use a summon, a trap, a warcry, a hammer) and they solved three of those with class code. That is the "capability" list of [04](04-stat-skill-agency.md): about ten delivery kinds, each written once.

## 5. Adopt, and do not adopt

| # | adopt (idea, in our own words and code) | from | where it lands |
|---|---|---|---|
| 1 | postcondition-by-state interactions and identity by class id or item code | kolbot `openChest`, Koolo `InteractObject` | quest steps, loot ([05](05-architecture-and-overhaul-options.md) 3.2) |
| 2 | level preset lookup on entry (rooms first) | D2BS API | fork read verb, decision K1 |
| 3 | one batched town pass | kolbot `Town.doChores` | arbiter batching rule |
| 4 | goal-list build notation with automatic prerequisites, an "all" sink and a "block" target for dexterity | kolbot AutoSkill and AutoStat | build file `didle.build/1` ([04](04-stat-skill-agency.md)) |
| 5 | staged leveling builds joined by the free Akara respec, with the respec levels as data | SoloPlay | guide tracks per class ([03](03-classes-plan.md)) |
| 6 | damage-ranked skill choice with neighbour counting and a mana and cooldown penalty | jaenster-kolbot | the rotation planner ([04](04-stat-skill-agency.md)) |
| 7 | difficulty gate = level cap and positive resists after the next penalty, with an over-level escape | SoloPlay, Koolo | readiness for the difficulty transition |
| 8 | curse and monster-aura reflexes as data | Koolo | reflex layer |
| 9 | "broke" as a formula over ALL gold, with a short list of cheap farms | SoloPlay | recovery, inside the arbiter |
| 10 | merc wanted per build as data, temporary merc before the wanted one | SoloPlay | merc policy rows |
| 11 | a scored, repeatable run as the public quality number | d2bsng leaderboard | Proving Ground L2 report |
| 12 | per-stat gear weights for the ADVISOR (never auto-equip) | SoloPlay `DynamicTiers` | upgrade advisor |

| do NOT adopt | why |
|---|---|
| NIP as our rule format | weaker logic on unknowns, one scope, text predicates; ours is the better base (importer only) |
| code inside data (`[tier]` expressions, `Update()` functions in build templates, respec conditions as functions) | not checkable, not shareable safely, breaks "builds are data" |
| per-class attack FILES as the unit of class support | this is how kolbot got seven hand-written class files; the delivery-kind drivers of [04](04-stat-skill-agency.md) are the same work organised so a new build needs no code |
| silent stalls (AutoSkill's barrier entries simply stop spending) | our rule is an error or a Needs-you row that names the cause |
| auto-equip and never-ask autonomy | the product is the player's decisions |
| anything from the D2R anti-detection lineage (mouse humanisers, obfuscated builds), leaked paid-bot code, packet casting | irrelevant offline; some of it is stolen code the survey already ruled out |
| moving the bot inside the engine process | the fork can answer the same questions over the socket; GPL boundary (ADR 0003) |
| ladder-only runewords assumed by nearly every guide and by SoloPlay (Spirit, Insight, Fortitude, Grief, Infinity and 18 more) | not makeable in vanilla offline 1.14d [V docs/sme/builds-gear.md from this install's runes.txt; helper B: Arreat Summit archive and the Amazon Basin wiki agree on 23]. Whether d.idle mode enables them is the owner's call (question Q4) |

## 6. What I could not find

- Any D2 bot whose correctness is measured against the original game, any D2 reinforcement-learning or LLM player, any D2 idle game built on the real rules. On those three axes there is no prior art to copy; d.idle is the first attempt in public as far as about 60 fetches and the survey's 46 projects show.
- The maxroll planner's share format (the page is rendered by script and was not readable) and a community notation that records the ORDER of points other than kolbot's two forms [helper B section 3].
- Forums behind login.
