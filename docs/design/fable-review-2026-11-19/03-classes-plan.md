# 03. All seven classes, Normal to Hell: builds, what the engine needs, and the order to bring them online

> status: REVIEW for the owner (design review 2026-11-19, card fablereview, docs only). Nothing here is decided. Part of [00-summary.md](00-summary.md).
>
> Tags: **[V]** read from the source named (helper A = public bot sources with file and line; helper B = web pages fetched; helper E = d2d coverage; helper G = this tree); **[S]** secondary (the repo's SME notes under `docs/sme/`, community guides); **[U]** veteran knowledge or judgment, not checked against the tables. **Every skill level, prerequisite and number tagged U must be read from the player's `skills.txt` before it becomes data.** The two d2-sme consults of this review were written from memory and were weak on details (one of them wrongly held Spirit and Insight to be available offline); I used them for ranking arguments only and say so where I do.
>
> Setting: classic LoD 1.14d, offline single player, players 1, self-found. The PLAYER equips gear (decision 2026-11-10). The hero has mediocre reflexes: it casts, holds range, drinks, retreats, uses a mercenary.

## 1. Ground rules that change every class plan

| rule | fact | source |
|---|---|---|
| Ladder-only runewords are NOT makeable offline | 23 words: Spirit, Insight, Fortitude, Grief, Infinity, Faith, Harmony, Ice, Dream, Dragon, Phoenix, Pride, Edge, Obedience, Oath, Death, Destruction, Brand, Last Wish, Lawbringer, Rift, Wrath, Voice of Reason. Available: Stealth, Lore, Ancient's Pledge, Rhyme, Smoke, Leaf, Malice, Strength, Treachery, Duress, Lionheart, Heart of the Oak, Call to Arms, Chains of Honor, Enigma and the other 1.10 non-ladder words | [V docs/sme/builds-gear.md from this install's runes.txt; V helper B: Arreat Summit archive and Amazon Basin wiki; V helper A: kolbot marks the same 23] |
| Consequence | nearly every modern guide (Maxroll is written for D2R) and Kolbot-SoloPlay assume Spirit and Insight. No Insight means NO free mana from the mercenary: casters pay with potions, energy points, Warmth-type skills and mana gear. Five of our endgame plan tables still set stat targets by Spirit, Insight or Fortitude [V the `tags` text of the hammerdin, summoner, tornado, warcry and blizzard endgame tables in `internal/plandata`] | question Q4: does d.idle mode enable ladder words in the fork? |
| Respecs | one free respec per difficulty from Akara after the Den of Evil (patch 1.13), so three per hero; Token of Absolution (cube four essences from Hell act bosses) exists since 1.13 for repeatable respecs | [V helper B: patch 1.13 notes; "three total" S]; d2d support for the Akara respec is U and must be drilled |
| Points | 1 skill point per level plus 4 quest points per difficulty (Den 1, Radament 1, Izual 2): 110 at clvl 99, about 85 to 90 at a realistic Hell hero of clvl 75 to 80. 5 stat points per level plus 5 per difficulty from Lam Esen | [U standard; the repo's `quest-points.json` holds it] |
| Difficulty penalties | resists -40 in Nightmare, -100 in Hell; experience lost on death 5 and 10 percent of the level; Hell monsters commonly carry one immunity, sometimes two | [V helper B; V helper E for d2d wiring] |
| No Iron Maiden from Oblivion Knights | removed in 1.13c | [V helper B] |
| No Diablo Clone, no Pandemonium event offline | portals are stubs in 1.14d itself | [V helper E] |
| Difficulty gate (community practice) | SoloPlay: level cap per class, about 33 for Normal and 65 to 75 for Nightmare, AND fire, cold, lightning resist above 0 after the next penalty; Koolo: clvl 41 and 70 with Hell floors fire 15, lightning -10 | [V helper A] |
| Mercenary | the Act 2 guard is the default for every class; aura by hire difficulty and type (Normal: Prayer, Blessed Aim, Defiance; Nightmare: Holy Freeze, Might, Thorns). SoloPlay wants Holy Freeze for Sorceress, Amazon, Assassin, Paladin; Might for Necromancer and Barbarian | [V helper A 4.5]; on d2d auras do not reach the merc's owner or pets correctly yet ([02](02-fidelity-and-gaps.md) A3) |

## 2. What "the engine needs" means: delivery kinds

A build is playable when every skill it leans on has a DRIVER for its delivery kind in the fighter. Classes share drivers; that is why the class order below is not alphabetical. State today is from helper G's read of `internal/combat` and the freebuild prototype's capability table.

| id | delivery kind | examples | state today |
|---|---|---|---|
| K1 | targeted missile | Fire Bolt, Ice Blast, Fire Ball, Bone Spear, Teeth, Holy Bolt, arrows, thrown javelins | live (Sorceress); others code only |
| K2 | self-centred nova | Frost Nova, Nova, Poison Nova, Static Field, War Cry, Shock Wave | live for Static Field and Frost Nova as two special rules |
| K3 | ground target, with delay or cooldown | Blizzard, Meteor, Fire Wall, Fissure, Volcano, Firestorm | kind exists in data; never live; cooldown handling U (and d2d skill delay is suspect) |
| K4 | melee swing | Zeal, Sacrifice, Bash, Double Swing, Frenzy, Concentrate, Berserk, Jab, Dragon Talon | live (Paladin) |
| K5 | aura selection | Might, Holy Fire, Concentration, Fanaticism | live for the Paladin's ladder |
| K6 | timed self or party buff upkeep | Frozen Armor, Holy Shield, Bone Armor, Cyclone Armor, Shout, Battle Orders, Fade, Burst of Speed, Energy Shield | live for Frozen Armor; state ids for five armors only |
| K7 | summon upkeep | skeletons, mages, golem, Valkyrie, wolves, spirits, Shadow | code for Necromancer and Druid; a short live check on Necromancer (2026-10-29) [S code comments] |
| K8 | corpse target | Corpse Explosion, Revive, Find Potion, Death Sentry needs corpses | code (Necromancer) |
| K9 | curse application | Amplify Damage, Decrepify, Lower Resist | code; d2d put rule deviates |
| K10 | trap placement and cadence | Wake of Fire, Lightning Sentry, Death Sentry | code (`trap.go`), never live; d2d Assassin unmeasured |
| K11 | cast standing inside the pack | Blessed Hammer | none (a band lists Hammer; no positioning rule) |
| K12 | movement attack | Whirlwind, Charge, Leap Attack, Dragon Flight | none |
| K13 | charge-up and finisher | martial arts | none |
| K14 | shapeshift state | Werewolf, Werebear | none |
| K15 | ammunition and thrown quantity | arrows, bolts, javelin stacks | bow restock path exists; thrown stacks none |
| K16 | Teleport movement | Teleport | live (Sorceress) |

## 3. The classes

For each: builds ranked for THIS product (few delivery kinds, safe when played sloppily, tolerant of self-found gear, and fun to watch), a guide track (leveling order and the respec the community uses), stats, gear without ladder words, breakpoints, dangers, what is needed. Breakpoint rows are from `docs/sme/mechanics-breakpoints.md` and agree with helper B's fetch [V both].

### 3.1 Paladin (online in Act 1 today)

| build | kinds | notes |
|---|---|---|
| **P1 Zealot** (Zeal + Fanaticism or Might + Holy Shield) | K4, K5, K6 | what the hero plays now; no positioning; physical immunes in Hell are the wall (answer: mercenary, Vengeance or Berserk-like element is not available, so skip rules or P2) |
| **P2 Hammerdin** (Blessed Hammer + Concentration + Vigor + Holy Shield) | K11, K5, K6 | the strongest self-found Hell clearer in 1.14d: magic damage, almost nothing immune [S]; needs K11 and mana without Insight |
| P3 Holy Fire start | K5, K4 | the community's Normal opener: Holy Fire aura does the killing to about clvl 19 to 24, then respec [V helper A: SoloPlay Paladin respecOne 19; Koolo Holy Fire to 24] |
| not for the bot | | Smiter (boss-only), Charge (K12), Fist of the Heavens (cooldown, weak alone in 1.14d) |

- Guide track (community) [V helper A, SoloPlay `paladin.StartBuild`, `LevelingBuild`]: Start = Might 1, Sacrifice 1, Holy Fire rising to 6 with Resist Fire as its synergy, Zeal 4. Respec at 19. Then Blessed Hammer to 20 with Concentration and Vigor alternating, Holy Shield 1 at 24, Blessed Aim and Holy Shield last. A no-respec Zealot track is what `paladin-zealot-lite` holds.
- Stats rule: strength for the worn gear, dexterity for block with Holy Shield (the "block" target of kolbot AutoStat), rest vitality, no energy [V helper A: SoloPlay BlockChance 75 for the Paladin].
- Gear without ladder words [S docs/sme/builds-gear.md]: Stealth armor (Tal Eth) at 17, Ancient's Pledge shield at 21, Lore helm, Rhyme or a resist paladin shield, Smoke; a fast one-hand weapon for Zeal; +skills scepters from vendors. Later: Heart of the Oak is not for him; Hand of Justice and Exile are far.
- Breakpoints: FCR 0/9/18/30/48/75/125 (Hammer wants 75); FHR 0/7/15/27/48/86; block frames 5 without and 2 with Holy Shield.
- Dangers: Normal Duriel (cold, chill); Nightmare and Hell packs with Decrepify or Holy Freeze (halve a Zealot); physical immunes (Zealot), a handful of magic immunes (Hammer); mana burn for Hammer.
- Needed: nothing for P1 beyond leaving Act 1; K11 for P2 (2 to 3 cards: stand at the pack's edge, cast, hold; kolbot's `getHammerPosition` is the idea [V helper A]); Smite and melee reach fidelity (A1); the aura-to-merc gap (A3).

### 3.2 Sorceress (online in Act 1 today)

| build | kinds | notes |
|---|---|---|
| **S1 Fire** (Fire Bolt then Fire Ball, Fire Mastery, Meteor optional) | K1, K6, K16 | the simplest caster: one targeted missile with splash from clvl 12; Koolo levels this way to 24 [V helper A 6.5] |
| **S2 Blizzard** (Ice Blast or Glacial Spike, Blizzard, Cold Mastery) | K1, K3, K6, K16 | the standard Hell farmer; needs K3 and correct skill delay on d2d (A5); cold immunes are common in Hell [U] |
| S3 Lightning or Nova | K1, K2 | what the bot plays today as Charged Bolt; Lightning and Chain Lightning were broken on d2d (last-unit fix merged 2026-11-12; cast length not re-checked) and sit on every profile's avoid list [V freebuild README; helper E] |
| two-element endgame | | Meteorb (Fire Ball or Meteor with Frozen Orb), Blizzard plus Fire Ball: the community's answer to Hell immunities [V helper A: SoloPlay finals Blizzballer, Meteorb, Blova] |

- Guide track [V helper A SoloPlay; V helper B Maxroll]: Normal on Charged Bolt and Nova or on Fire Bolt and Fire Ball, Static Field 1 to 6, Teleport at 18, Frozen Armor 1, Warmth 1; respec at about 26 into Blizzard with Ice Blast; second respec in the 60s into the two-element final.
- Stats rule: strength for gear, energy early (SoloPlay goes to energy 40 to 60 before vitality in the Start stage because there is no other mana), then all vitality. Without Insight keep some energy; this is a real difference from D2R guides.
- Gear: Stealth, Lore, Leaf staff (Tir Ral, +3 fire skills) at 19 for S1, Ancient's Pledge or Rhyme, +skills staves and orbs from Drognan, Ormus and Akara; a Staff of Teleportation is no longer needed once Teleport is learned.
- Breakpoints: FCR 0/9/20/37/63/105/200 (63 by Hell, 105 the goal); Lightning skills have their own slower table; FHR 0/5/9/14/20/30/42/60/86.
- Dangers: lowest life of all classes: anything that reaches her; ranged packs in Act 2 tombs and Act 3; Hell immunities to her one element.
- Needed: the second build as the proof of "builds are data" (the audit's phase 7, never run); K3; Energy Shield left unsupported on purpose until mana is solved.

### 3.3 Necromancer (first new class)

| build | kinds | notes |
|---|---|---|
| **N1 Summoner** (Raise Skeleton, Skeleton Mastery, mages, Clay Golem, Amplify Damage, later Decrepify, Corpse Explosion, Revive) | K7, K8, K9, K6 | the safest idle hero there is: the army takes the hits, failure is slow not fatal; the only class whose immunity answer is built in (mixed minion damage, Amplify breaks physical immunity on many monsters) [S docs/classes/ranking.md] |
| N2 Bone (Teeth, then Bone Spear, Corpse Explosion, Bone Armor) | K1, K8, K6 | the community's LEVELING build: SoloPlay starts every Necromancer on Teeth and Bone Spear and respecs at 26; Maxroll the same [V helper A 4.2; V helper B] |
| N3 Poison Nova | K2, K9 | needs Lower Resist and mana; later |

- Guide track: N2 to clvl 26 or straight N1 from level 1 (slower early, no respec). For a watched idle game N1 from the start is the better show and needs no respec; offer both.
- Stats: strength for gear, rest vitality, energy only for N2 and N3.
- Gear: +skills wands and heads from vendors (Koolo even shops 2-socket Bone Wands for the White runeword [V helper A 6.4]), Stealth, Lore, Rhyme, Smoke; the mercenary with Might is the army's damage aura (A3 must be fixed on d2d first).
- Breakpoints: FCR 0/9/18/30/48/75/125; FHR 0/5/10/16/26/39/56/86. Neither matters much for N1.
- Dangers: Duriel and other bosses that delete the army (golem recast is the tool); narrow corridors (Maggot Lair) where minions block; no corpses at the start of a fight.
- State: profile, upkeep code (`necro.go`, `necro_corpse.go`), three tables, a smoke note and one early live check exist; d2d has skeletons, mages, Revive, Clay Golem and an oracle-run pet AI; curse put rule and golem specials deviate [V helper E]. The freebuild prototype shows the power estimator has no notion of pet damage yet (it picks the staff swing as "main") [V].
- Needed: 8 to 12 cards: live K7 loop (count, re-raise, golem), K9 once per pack with the state read back, K8 Corpse Explosion, pet damage in the power model, the summoner archetype's margin checked in the Proving Ground.

### 3.4 Barbarian

| build | kinds | notes |
|---|---|---|
| **B1 Frenzy** (Double Swing to 24, then Frenzy; one weapon mastery; Battle Orders, Shout; Berserk 1 for physical immunes; Natural Resistance, Iron Skin) | K4, K6 | reuses the Paladin's melee driver; the community's leveling line: SoloPlay Start is Double Swing and mastery, Frenzy at 24 [V helper A 4.2] |
| B2 Whirlwind | K12, K6 | the classic Hell build; needs the movement-attack driver, and d2d's Whirlwind radius deviates (A6): do not start here |
| B3 Concentrate (shield) | K4, K6 | the safe variant: uninterruptible, defence bonus, sword and shield; slow but hard to kill, good for an unattended hero [U] |
| not for the bot | | Throw (K15), Singer (War Cry, K2 with positioning; and War Cry is 8 times too strong on d2d) |

- Stats: strength for weapons and armor, dexterity for attack rating as needed, rest vitality.
- Gear: the most gear-dependent class (weapon damage is everything); without Grief and Fortitude the self-found plan is: Malice, Steel or Strength early, the best rare or unique one-hander found, Lionheart or Duress armor, life leech on rings or gloves. The upgrade advisor matters most here.
- Breakpoints: FHR 0/7/15/27/48/86; attack speed comes from the weapon tables (in `skilldamage.json`).
- Dangers: he is always in the pack: curses, monster auras, Fire Enchanted death, Hell elemental damage (Natural Resistance is why the class works); physical immunes (Berserk); mana for skills is tiny, leech is the supply.
- Needed: 6 to 9 cards after the Paladin is solid: dual-wield awareness, warcry upkeep (K6 with party range), Berserk as the immune fallback through the profile, the A1 and A6 fidelity items first.

### 3.5 Druid

| build | kinds | notes |
|---|---|---|
| **D1 Fire, Normal** (Firestorm, Molten Boulder, Fissure) | K1 or K3 | the community's Normal opener: Fissure from 12 clears Normal; SoloPlay Start stage, respec at 26 [V helper A] |
| **D2 Wind** (Tornado, Hurricane, Cyclone Armor, Oak Sage) | K1 (erratic missile), K6, K7 | the Hell build: physical and cold, few double immunes; Tornado must be cast close [V helper A: kolbot aims it with a random offset]; d2d has no verdict on Tornado flight (A8) |
| D3 Summoner | K7 | wolves and Grizzly; safe and slow |
| not for the bot | | shapeshifters (K14, and d2d form rules are incomplete) |

- Stats: strength for gear, rest vitality, some energy early.
- Gear: Stealth, Lore, pelts with +skills, Heart of the Oak late.
- Breakpoints: FCR human 0/4/10/19/30/46/68/99/163; FHR 0/5/10/16/26/39/56/86.
- Needed: 6 to 10 cards, mostly reuse (K3 from the Sorceress, K6 and K7 from the Necromancer); blocked on the d2d verdict for Tornado, Fissure, Volcano.

### 3.6 Assassin

| build | kinds | notes |
|---|---|---|
| **A1 Traps** (Fire Blast, Wake of Fire in Normal; Lightning Sentry and Death Sentry after a respec at about 32; Burst of Speed, Fade, Shadow optional) | K10, K1, K6, K8 (Death Sentry), K7 optional | safe: the traps fight; the hero stands back. SoloPlay and Maxroll both level this way [V helper A 4.2; V helper B] |
| A2 Kicker (Dragon Talon) | K4 | simple melee, strong early; needs good boots; a cheap second build once K4 is shared |
| not for the bot | | martial arts with charges (K13), Whirlwind assassin (K12) |

- Breakpoints: FCR 0/8/16/27/42/65/102/174; FHR 0/7/15/27/48/86; trap-laying speed is an attack-speed table and was not sourced [U].
- Dangers: lightning immunes in Hell (Death Sentry's corpse explosion is the answer, it needs a first corpse); low life.
- Needed: 8 to 12 cards AND the largest fidelity debt: 0 of 30 skills run-tested on d2d (A7). The fidelity class gate comes first.

### 3.7 Amazon

| build | kinds | notes |
|---|---|---|
| **Z1 Bow, fire and cold arrows early, Multiple Shot and Strafe later, Valkyrie** | K1, K15, K7 | ranged and easy to aim; the bow restock path already exists in town code; gear-hungry in Hell (no Faith) |
| Z2 Javelin (Jab and Power Strike, then Charged Strike, Lightning Fury, Plague Javelin) | K4 then K1, K15 | SoloPlay's choice: melee javelin to 30, respec, lightning to Hell [V helper A]; needs thrown-stack handling and melee on a fragile hero |
| Z3 Poison or Plague Javelin | K1, K15 | few immunes, slow |

- Stats: strength and dexterity for the weapon, rest vitality.
- Breakpoints: FHR 0/6/13/20/32/52/86; bow speed from the weapon tables.
- Dangers: fragile; lightning immunes for Z2; arrows run out; Pierce is incomplete on d2d (A9).
- Needed: 8 to 12 cards. I rank her last: two delivery problems of her own (ammunition, thrown stacks), the most gear-dependent ranged class without ladder words, and a d2d gap in her signature passive. This disagrees with nobody in the repo; `docs/classes/ranking.md` also has her near the bottom.

## 4. Order, effort and risk

"Playable" below means: the class has at least one guide track that passes the Proving Ground act run for Normal Act 1 and 2 on three seeds. Effort is in agent-cards of 60 to 90 minutes and excludes the act work that is shared by all classes (B1, B2 of [02](02-fidelity-and-gaps.md)).

| order | class and first build | why here | new kinds | bot cards | fidelity gate first | risk |
|---|---|---|---|---|---|---|
| 0 | Paladin Zealot, Sorceress (current) | finish Act 1, enter Act 2; they prove the acts | none | part of the act work | A1, A3 | low |
| 1 | Sorceress second build (Fire) | the cheapest proof that a build is data and that the planner of [04](04-stat-skill-agency.md) picks the attack | none (K1) | 3 to 5 | Fire Ball burst check | low |
| 2 | **Necromancer Summoner** | the best idle hero; forces K7, K8, K9, which Druid, Assassin and Amazon reuse; makes the third archetype (screened by pets) real | K7, K8, K9 | 8 to 12 | A2 (curse rule), pet matrix | medium |
| 3 | Paladin Hammerdin | the Hell answer for the default class; adds K11 | K11 | 2 to 3 | Blessed Hammer hit logic (read only today) | low |
| 4 | Barbarian Frenzy, then Concentrate | K4 reuse; the first hero with no ranged option and no mana, tests the survival model's contact archetype hard | warcry upkeep | 6 to 9 | A1, A6 | medium: gear dependence |
| 5 | Sorceress Blizzard, Druid Fire | K3 once, used twice | K3 | 4 to 6 + 6 to 10 | A5 (skill delay), A8 | medium |
| 6 | Assassin Traps | K10 exists as code; fidelity unknown | K10 | 8 to 12 | A7 (whole class) | high until measured |
| 7 | Amazon Bow | last | K15 | 8 to 12 | A9 | medium |
| later | Whirlwind (K12), Wind Druid, javelins, kicker, Poison Nova, two-element Sorceress finals | each is one more kind or one more track | | | | |
| never (for the bot) | shapeshift, martial-arts charges | high cost, low watch value; the capability table says "the hero cannot use this" | | | | |

Reading: after step 2 the three archetypes that matter (caster, contact melee, screened summoner) are all live and three of seven classes have a build. Steps 0 to 3 are about a month if the Proving Ground exists; without it I would not start a third class at all.

Normal to Hell per class is not a separate project: it is the act work plus the difficulty transition plus immunity handling, shared by all. The class-specific part of Hell is one row of data each: the element set, the fallback attack, the wanted mercenary, and the guide track's final stage.

## 5. Data and code each class needs under "builds are data"

| piece | exists | to add |
|---|---|---|
| plan tables (`internal/plandata/build-*.json`) | all seven, 2 to 5 each | re-author as guide tracks in the build-file form ([04](04-stat-skill-agency.md)); remove ladder runewords from tags and gear milestones; add the respec stage where the community uses one |
| combat profile bands (`classdata/profiles.json`) | all seven, mostly U | bands become a tie-break under the planner; `avoid` lists become capability tiers |
| class traits | scattered over about ten places (prefs, mana need, caster triple written out three times, readiness classes for two classes only) [V helper G] | one traits table; this is audit phase 3, still open |
| power archetype | derived from the build | pet damage; multi-strike and synergy inputs (prototype findings) |
| danger profiles | generated for all difficulties | measured unit counts beyond Normal Acts 1 and 2 |
| readiness classes (`readiness/checks.json`) | Sorceress, Paladin | every class (an unknown class is an error today) |
| level plan v3 | Paladin active; Barbarian and Sorceress inactive copies | replace per-class slices by the survival model's zone choice; keep the plan only as the quest order |
| mercenary wanted per build | aura preference table | one row per track (section 1) |
| skill fixtures and matrices | Sorceress and Paladin skill matrices | one per class as the fidelity class gate |

## 6. What I could not verify

- Hell immunity lists per area and per class (who is immune where): the repo's `docs/progression/difficulty.md` and the danger profiles carry resists from the tables; I did not tabulate them. The class rows above say "common" or "few" from experience [U].
- Whether the Akara respec and the Token recipe work on d2d [U].
- Exact skill levels and prerequisites in the build sketches [U]; SoloPlay's ordered lists are V as that project's choices, not as optimal play, and they assume ladder words for gear.
- Self-found viability of each Hell build without Spirit, Insight, Grief and Fortitude is the largest open game-design unknown in this plan. Only play (the Proving Ground at Hell area levels, with gear sets the owner considers "self-found typical") can answer it.
