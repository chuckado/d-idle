# Build catalogue: ballpark star thresholds and first-draft ratings (SME, 2026-11-22)

> status: current. Answers `docs/design/fable-review-2026-11-22/ui-unification.md` sections D3 and D5 (user decision 2026-11-22: measured thresholds need not be exact, "ask the SME, ballpark numbers"). Advice only, no code. Author: d2-sme agent.

Tags: EXP = veteran experience, not looked up this turn; SRC = read in a repo doc (path given); unverified = I am not sure. No web source was fetched. Target is classic 1.14d. Every number below is a BALLPARK for the first release: the real cut points should be re-set once the Proving Ground has measured two or three builds (D4 `scales`). Stars from a CHEATED or impossible test character never count (test-character provenance rule).

The seventh rating is "Idle-friendliness" (the D3 "Hands-off"): how well the build plays UNATTENDED for the bot, with a merc.

## 1. Ballpark thresholds (MEASURED-LATER ratings)

General rule for all four tables: one stars-row is a band, inclusive of its lower edge. A metric is measured per difficulty and the star carries its difficulty tag. A rating needs at least 3 runs of the scenario; use the median. Hero level for the scenario is "level-appropriate": clvl inside the area's -5..+5 XP window (SRC `docs/sme/pace-and-formulas.md` line 73).

### 1.1 Defense (unattended hero, one hour of farming a level-appropriate area, same gear tier as the guide)

| Stars | Deaths per hour | Chickens/emergency exits per hour | Healing potions per hour | Lowest life seen in a fight |
|---|---|---|---|---|
| 5 | under 0.02 (less than 1 per 50 h) | under 0.2 | up to 10 | 40% or more |
| 4 | 0.02 to 0.05 (1 per 20 to 50 h) | 0.2 to 1 | 10 to 25 | 30% to 40% |
| 3 | 0.05 to 0.1 | 1 to 3 | 25 to 50 | 20% to 30% |
| 2 | 0.1 to 0.25 | 3 to 8 | 50 to 100 | 10% to 20% |
| 1 | 0.25 or more (at least 1 per 4 h) | 8 or more | over 100 | under 10% |

Score = the LOWEST of the four columns (the worst one decides). Normal numbers are expected to be much better than Hell numbers for every build, so the star carries its difficulty tag (section 4). All numbers EXP, unverified; a human Hammerdin dies maybe once per 100 h in Normal (EXP), which is why 5 stars needs under 0.02.

### 1.2 Clear speed (fixed area, full clear incl. unique packs, Proving Ground scenario)

| Stars | Monsters killed per minute (standard pack area) | Relative to the reference build, same area, same gear tier |
|---|---|---|
| 5 | 12 or more | 150% or more |
| 4 | 8 to 12 | 120% to 150% |
| 3 | 5 to 8 | 80% to 120% |
| 2 | 3 to 5 | 50% to 80% |
| 1 | under 3 | under 50% |

Reference build: the Hammerdin (EXP: the most common unattended-style clearer; the user may prefer the Blizzard sorceress). Monsters per minute depends on pack density and walking, so the RELATIVE column is the one to trust; the absolute column is a sanity check (unverified). Count walking time in (a build that needs to run to the next pack is slower).

### 1.3 Leveling (minutes of play per level, fresh character, no setup writes, Normal then Nightmare then Hell as the bot progresses)

| Stars | clvl 1 to 30 | clvl 30 to 60 | clvl 60 to 85 |
|---|---|---|---|
| 5 | 6 min per level or less (about 3 h total) | 15 min or less | 60 min or less |
| 4 | 6 to 9 (3 to 4.5 h) | 15 to 25 | 60 to 120 |
| 3 | 9 to 14 (4.5 to 7 h) | 25 to 40 | 120 to 240 |
| 2 | 14 to 22 (7 to 11 h) | 40 to 65 | 240 to 480 |
| 1 | over 22 (over 11 h) | over 65 | over 480 |

3 stars is meant as "a normal non-rushed human guide pace" (EXP, unverified: no source gives clvl-per-hour for 1.14d, SRC `docs/sme/pace-tables.md` says the same). A bot wastes time on town trips and mistakes, so the absolute stars will skew low; therefore RANK builds against each other too (a build is Leveling 5 only if it is also in the top third of the measured builds). Deaths on the way: each death in the band lowers the star by 1 (floor 1). Fresh-start only; never from a CHEATED character.

### 1.4 Bossing (time from first hit to kill on Normal; level-appropriate hero; walk-in and Town Portal time excluded)

| Stars | Andariel | Duriel | Mephisto | Diablo | Baal (the boss himself, waves excluded) |
|---|---|---|---|---|---|
| 5 | under 20 s | under 15 s | under 20 s | under 30 s | under 45 s |
| 4 | 20 to 40 s | 15 to 30 s | 20 to 45 s | 30 to 60 s | 45 to 90 s |
| 3 | 40 to 90 s | 30 to 75 s | 45 to 100 s | 60 to 120 s | 90 to 180 s |
| 2 | 90 to 180 s | 75 to 150 s | 100 to 200 s | 120 to 240 s | 180 to 360 s |
| 1 | over 180 s | over 150 s | over 200 s | over 240 s | over 360 s |

Final Bossing star = the average of the five rows (round half down). A death in a boss scenario caps that row at 2; running out of potions or not finishing within 10 minutes is a 1. Level-appropriate means the human median level at the boss (EXP, unverified: about clvl 22 to 26 at Andariel, 24 to 28 Duriel, 28 to 32 Mephisto, 30 to 34 Diablo, 33 to 38 Baal; `internal/plandata/boss_levels.go` has the repo's own table, use it instead). The kill times are unverified guesses; boss life is in `monstats.txt`/`bossdata`.

### 1.5 Not measurable yet (OPINION)

- Farming (XP): MEASURABLE (XP per hour in a named area) but I give no table: use the 1.2 pattern (relative to the reference build, 150/120/80/50 percent) until you have a floor. The repo's current XP floor is 1000 XP/h (SRC `docs/sme/pace-tables.md`).
- Farming (items): OPINION. Proxy only: boss and area runs per hour of a farm order (needs endgame tables most builds lack).
- Idle-friendliness: OPINION. Proxies later: needs-you rows, held points and stall minutes (mana, corpses, dead merc, minion upkeep) per hour from the soak log.

## 2. First-draft star ratings

Columns: Lv Leveling, It Farming (items), XP Farming (XP), Bo Bossing, De Defense, Cl Clear speed, Id Idle-friendliness. Scale 1-5. MEASURED-LATER ratings: Lv, XP, Bo, De, Cl. OPINION: It, Id (and every number is OPINION today, since nothing is measured). All unverified. "N/H" means Normal / Hell where the gap is big (section 4).

Builds are grouped by REAL build, from `internal/plandata/build-*.json` (SRC). The six Paladin tables (`paladin-normal30`, `paladin-normal30-v2`, `paladin_hammerdin_guide`, `paladin-zealot-lite`, `paladin_hammerdin_endgame90`, `paladin_hammerdin_zeal_endgame90`) are two real builds: Hammerdin (hammer from 18) and Zeal start (Zeal bridge to hammers); the `v2`, `guide` and `normal30` tables are variants of the same leveling plan.

### 2.1 Core builds (the Sorceress, Paladin and Necromancer set)

| Build (plandata tables) | Lv | It | XP | Bo | De | Cl | Id | One-line reason |
|---|---|---|---|---|---|---|---|---|
| Paladin Hammerdin (normal30, normal30-v2, hammerdin_guide, hammerdin_endgame90) | 3 | 5 | 5 | 4 | 5 (N 5 / H 4) | 4 | 3 | Slow before Blessed Hammer at 18; magic damage few monsters resist; shield and block make it the sturdiest caster; hammers miss in corridors and drink mana. |
| Paladin Zeal start (zealot-lite, hammerdin_zeal_endgame90) | 4 | 2 | 2 | 3 | 4 | 3 | 4 | Melee with Zeal and cheap auras levels steadily and needs little mana; weak if kept past hammers. |
| Sorceress Blizzard (sorceress_normal30 cold line, blizzard_static, sorceress_blizzard_endgame90) | 4 | 5 | 4 | 4 | 3 (N 3 / H 2) | 4 | 3 | Ground cold from range with Teleport; thin life, cold-immune packs must be skipped, and the bot cannot swap element. |
| Sorceress Fire Ball (sorceress_fire_normal30, experimental) | 5 | 3 | 3 | 4 | 2 | 4 | 4 | Fire Bolt then Fire Ball is the simplest strong early caster; fire immunes late cut farming. |
| Sorceress Lightning (fallback-lightning, v2, v3 = lightning_chain) | 3 | 4 | 4 | 4 | 2 | 5 | 2 | Highest clear when it works, uneven damage, biggest mana bill; the fighter cannot use its attacks well yet (SRC the D5 note). |
| Necromancer Summoner (necromancer_summoner_normal30, _endgame90, _guide, _sme) | 3 | 3 | 3 | 2 | 4 (N 4 / H 3) | 2 | 3 | Skeletons and mages take the hits; slow on bosses at clvl 25 to 35 (SRC `docs/design/necromancer-sme-answers-2026-11-22.md` section 1); minion upkeep each area. |
| Necromancer Bone and Corpse Explosion (necromancer_leveling_normal30) | 3 | 2 | 3 | 3 | 2 | 3 | 2 | Teeth and Bone Spear from range, CE needs a corpse and aim; no tank. |
| Necromancer Poison Nova (necromancer_poison_nova_sme, experimental) | 2 | 4 | 4 | 3 | 2 (N 2 / H 1) | 4 | 3 | No aim, flat 20 mana, strong on packs from clvl 36 to 40 (SRC the same SME answers, section 2); poison immunes, about 320 life at clvl 30. |

### 2.2 Other classes (second pass; lower confidence, not reviewed against any class doc this turn)

| Build (plandata tables) | Lv | It | XP | Bo | De | Cl | Id | One-line reason |
|---|---|---|---|---|---|---|---|---|
| Amazon Bow (amazon_bow_normal30 Exploding Arrow, amazon_multishot_endgame90) | 4 | 4 | 4 | 3 | 3 | 4 | 3 | Strong from range with a merc; arrows and quiver upkeep; Multishot endgame needs real gear. |
| Amazon Javelin (amazon_javelin_normal30, Poison Javelin and Lightning) | 4 | 3 | 3 | 3 | 3 | 3 | 3 | Great Normal leveler; needs good javelins, lightning immunes late. |
| Assassin Traps (assassin_fireblast_normal30 Wake of Fire, assassin_sentry_endgame90) | 3 | 4 | 4 | 4 | 3 | 3 | 2 | Sentries do the damage but need placement and aim; the bot must lay traps. |
| Barbarian Frenzy (barbarian_melee_normal30, barbarian_frenzy_guide) | 4 | 3 | 3 | 4 | 4 | 3 | 4 | Tanky melee with no mana problem; needs gear, merc and good life. |
| Barbarian War Cry (barbarian_warcry_endgame90) | 2 | 4 | 3 | 4 | 3 | 3 | 3 | Late stun and area damage; mana hungry; endgame table only. |
| Druid Fire (druid_fire_normal30, druid_fire_guide) | 4 | 3 | 3 | 3 | 3 | 3 | 3 | Fissure, Firestorm and Molten Boulder are easy early; fire immunes. |
| Druid Tornado (druid_tornado_endgame90) | 2 | 3 | 3 | 3 | 3 | 4 | 2 | Random hits at a distance; mana heavy; endgame table only. |

## 3. Identity, pros and cons (plain words; all unverified opinion)

Core builds:
- **Hammerdin**: the magic-hammer Paladin. Pros: almost nothing resists its damage; very hard to kill with a shield; scales to the end with cheap gear. Cons: weak and slow before level 18; hammers miss in narrow corridors; uses a lot of mana.
- **Zeal start**: melee Paladin that holds on to Zeal with cheap auras. Pros: steady early; little mana; good block. Cons: weak once monsters resist physical; slow clear; needs a switch to hammers.
- **Sorceress Blizzard**: Teleporting cold caster. Pros: kills at range and slows; Teleport saves time; the classic Baal and Hell runner. Cons: thin life; cold-immune packs; mana and Frozen Armor upkeep.
- **Sorceress Fire Ball**: fire caster. Pros: strong and simple early; fast single-target damage; cheap points. Cons: many fire immunes in Nightmare and Hell; thin life; experimental in the repo.
- **Sorceress Lightning**: lightning caster. Pros: fastest clear; ranged; Static Field helps on bosses. Cons: uneven damage; big mana bill; lightning immunes.
- **Necromancer Summoner**: army caster. Pros: minions take hits; strong with Iron Golem; hero rarely dies in Normal. Cons: slow on bosses; minions die more in Hell; per-area upkeep.
- **Necromancer Bone and CE**: ranged bones. Pros: Teeth and Bone Spear work early; magic damage; cheap start. Cons: CE needs a corpse and aim; no tank; weak late.
- **Necromancer Poison Nova**: poison caster. Pros: no aim; fixed 20 mana; very strong on packs late. Cons: poison immunes; fragile; weak before clvl 36.

Other classes (second pass):
- **Amazon Bow**: Pros: range; strong with a merc; good single-target. Cons: quiver upkeep; physical immunes; needs gear. 
- **Amazon Javelin**: Pros: strong early; Poison Javelin helps vs bosses; cheap. Cons: needs good javelins; lightning immunes; fades if unlucky.
- **Assassin Traps**: Pros: strong damage; safe from range; shadow helps. Cons: trap placement; fire/lightning immunes; fragile.
- **Barbarian Frenzy**: Pros: tanky; no mana issue; Iron Skin. Cons: needs gear; slow clear; Whirlwind aim.
- **Barbarian War Cry**: Pros: stuns and area damage; strong late; good vs bosses. Cons: mana heavy; endgame only; positioning.
- **Druid Fire**: Pros: simple early; strong area; cheap points. Cons: fire immunes; mana; fragile.
- **Druid Tornado**: Pros: ranged random hits; good sustained damage; strong clear. Cons: random aim; mana heavy; endgame only.

## 4. Difficulty tags (only where the gap is big)

Only Defense changes enough to tag, and only for four builds (the Normal number first): Hammerdin 5 / 4, Blizzard sorceress 3 / 2, Summoner 4 / 3, Poison Nova 2 / 1 (in the table above). The reason is the same for all (EXP, unverified): Hell resists and damage hit thin or minion-based heroes harder. Leveling and Clear shift with gear, not with difficulty, so they are not tagged. Bossing in Hell is a different scenario table (Hell boss life and resists); do not reuse the Normal times above for Hell.

## 5. Gaps, risks and recommendations for the main agent

- docs/sme has no `class-*` entry for the "Idle-friendliness" criteria; I used the D3 definition. Queue research if a proper checklist is wanted. Also no source gives clvl-per-hour for 1.14d (SRC `docs/sme/pace-tables.md`), so section 1.3 is the least reliable table.
- The D5 draft rated Summoner Defense 5 and Sorceress Blizzard Idle 4; I lowered them (Normal 4, H 3; Idle 3) because the hero has about 320 life at clvl 30 and the bot cannot swap elements. Both are EXP.
- Recommend `scales` data keeps both the absolute and the relative columns of 1.2 so the table can switch to relative-only without a schema change.
- Rate each build again after its first measured scenario; a star becomes `measured` only from a named scenario (D3).
