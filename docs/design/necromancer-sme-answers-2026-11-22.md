# Necromancer build answers for the unattended bot (SME, 2026-11-22)

> status: current. Answers the five questions of `docs/probes/necromancer-readiness-2026-11-22.md` section 4 (card necro2). Advice only, no code. Author: d2-sme agent.

Tags: SRC = read in a repo doc (path given); EXP = veteran experience, not looked up this turn; UNVERIFIED = I am not sure, check `skills.txt` or the fork before coding. Target is classic 1.14d. No web source was fetched this turn: every number not tagged SRC needs a check against `skills.txt`/`charstats.txt` before it becomes plan data.

Skill points: 1 per clvl from 2, plus 4 per difficulty (Den 1, Radament 1, Izual 2): about 33 at clvl 30 after Normal, 47 at clvl 40 in Nightmare, about 70 at clvl 60 in Hell (SRC `docs/sme/class-necromancer.md` section 1, EXP for the Hell figure).

## 1. Summoner skill split (verdict: keep the repo path to clvl 24, then buy the Revive line; no bone points)

Cumulative levels. RS = Raise Skeleton, SM = Skeleton Mastery, Mage = Raise Skeletal Mage, BA = Bone Armor, Amp = Amplify Damage.

| clvl | pts | RS | SM | Mage | Clay | Blood+Iron | Revive | Amp | BA | Weaken+Terror+Decrepify |
|---|---|---|---|---|---|---|---|---|---|---|
| 6 | 6 | 3 | 2 | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| 12 | 12 | 5 | 5 | 1 | 1 | 0 | 0 | 0 | 0 | 0 |
| 18 | 19 | 6 | 6 | 2 | 1 | 0 | 0 | 1 | 3 | 0 |
| 24 | 27 | 9 | 9 | 4 | 1 | 0 | 0 | 1 | 3 | 0 |
| 30 | 33 | 10 | 10 | 5 | 1 | 1+1 | 1 | 1 | 3 | 0 |
| 40 | 47 | 14 | 14 | 8 | 1 | 1+1 | 4 | 1 | 3 | 0 |
| 50 | ~59 | 17 | 17 | 10 | 1 | 1+1 | 6 | 1 | 3 | 1+1+1 |
| 60 | ~70 | 20 | 20 | 12 | 1 | 1+1 | 8 | 1 | 3 | 1+1+1 |

- Why: RS and SM move together because count and damage both need them (SRC class-necromancer formulas: count 2+slvl/3 from slvl 4, damage per level from SM); Mage adds the ranged damage; Clay at clvl 6 is the early tank (SRC). Up to clvl 24 this is exactly the repo plan `necromancer_summoner_normal30` (SRC class-necromancer section 4).
- Revive chain costs Clay 1 + Blood 1 + Iron 1 + Revive 1 (Blood needs Clay, Iron needs Blood, Revive needs Mage and Iron: SRC prerequisites table). Start Blood Golem at clvl 26 (after Izual) and Iron Golem right after; Revive 1 at clvl 30. Iron Golem replaces Clay as the one allowed golem (only one golem at a time, SRC) once it exists. Take it at 1 point: it copies the best item mods of the gear in your inventory only if you cube it, which the bot will not do, so it is just a sturdy tank (EXP).
- Golem Mastery: do not buy before clvl 50 (EXP; its prereq is Clay Golem, the value is golem life only, UNVERIFIED number). Summon Resist: never.
- Curses: Amp 1 point at 13 (kept from the repo plan). The Decrepify chain (Weaken, Terror, Decrepify = 3 points, Decrepify needs Terror needs Weaken needs Amp: SRC) only at clvl 50+ when RS/SM are 17+.
- Synergies: in 1.14d RS has none; SM is the damage and life multiplier (SRC formulas). Mage synergy with SM and Clay synergy with Golem Mastery: UNVERIFIED.
- Main risk: the clvl 25-35 hole, where summons do 50% to act bosses in Normal and 35% in Nightmare (SRC class-necromancer section 0), so Andariel/Duriel take minutes. Fallback: the bot may stand and Teeth/Attack with the wand while minions hold, and potions decide the fight (see section 6).

## 2. Bone/Poison Nova caster (verdict: Poison Nova is the best idle caster; Bone Spear is its poison-immune fallback; no Corpse Explosion)

| clvl | Poison Dagger | Poison Explosion | Poison Nova | Teeth / CE / Bone Spear | Main attack |
|---|---|---|---|---|---|
| 6-17 | 1 then to 10 (+Teeth 1 for CE) | 0 | 0 | Teeth 1-3 | Poison Dagger melee-range or Teeth |
| 18-29 | 10 to 15 | 1 (needs CE 1) | 0 | CE 1, Bone Spear 1-5 | Bone Spear (range, no aim problem) |
| 30 | 15 | 1 | 1-3 | CE 1, BS 5 | Bone Spear, Nova on packs |
| 40 | 20 | 1 | 10 | CE 1, BS 5 | Poison Nova main (mana 20 flat) |
| 60 | 20 | 1 | 20 | CE 1, BS 10 | Poison Nova main, Bone Spear on immunes |

- Why: Poison Nova costs 20 mana at every level (SRC table), needs no aim, and gets +20% per Poison Dagger level, +15% per Poison Explosion level (SRC synergies). It is the standard 1.14d guide Necromancer (EXP; the guide source to cite is Maxroll "Poison Nova Necromancer" and Amazon Basin Necromancer pages, UNVERIFIED this turn).
- When it becomes the main attack: clvl 36-40 (Nightmare), once Nova is about 8-10 with Dagger 20. Before that a level-1 Nova does not kill a pack faster than Bone Spear (EXP; at clvl 30 only 1-3 points are free after Dagger).
- Mana: Nova 20 each. At clvl 40 unbuffed mana is about 109 plus gear and Energy (SRC mana table), so 5 casts: the bot needs a mana belt column and Energy per section 5. Rule: cast only while mana >= 40 or a pack is on screen; drink mana at 25%.
- Main risk: poison-immune monsters (more of them in Nightmare and Hell, EXP), and the caster has about 320 life at clvl 30 (SRC table), no tank. Fallback: Bone Spear (magic, immunes are rare) and a Clay/Iron Golem point (2 points at clvl 26) so something is hit first. Do not buy Corpse Explosion points beyond 1: it is a prerequisite, not an attack here.
- Hell: a pure caster is the less safe unattended class (EXP); the summoner is the primary recommendation, the caster the build to offer second.

## 3. Curse rotation

| target class | summoner | caster | recast / why |
|---|---|---|---|
| regular pack | Amplify Damage on the middle of the pack | none until Lower Resist, then Lower Resist on packs with a 75%+ poison-resist type | Amp is 4 mana flat (SRC); it only doubles PHYSICAL damage, so it helps skeletons and golems, not mages, Bone Spear or Poison (EXP) |
| unique, champion | Amplify; Decrepify from clvl 50 | Lower Resist (clvl 30+) | once per target group, not per monster; curses are area on cast (EXP) |
| Andariel, Duriel, Mephisto | Amplify first, Decrepify at clvl 50+ | Lower Resist | bosses are not immune to Amp/Decrepify/Lower Resist; AI curses (Terror, Confuse, Attract, Dim Vision) mostly fail on bosses (EXP, UNVERIFIED) |
| immune to physical | skip Amp (does not break immunity); rely on mages, Lower Resist cannot help | n/a | leave or let mages/golem work; log and move on |
| curse-resistant / curses resisted | none | none | UNVERIFIED which 1.14d monsters resist: confirm with `debug skillstate` and stop casting after one failed cast per monster |

- Never use in the unattended bot: Iron Maiden, Life Tap (Life Tap heals only on the necromancer's own hits, not minions: UNVERIFIED), Attract, Confuse, Dim Vision (Hell cuts AI curse effect to a quarter: SRC class-necromancer curses note).
- Recast: read the curse from `debug skillstate` (SRC readiness S3); if that is not available, recast every 20 s while the pack lives (repo guess is 30 s, UNVERIFIED; the real duration scales with skill level, so the 20 s margin is cheap at 4 mana). Cast the curse before the army engages, never while HP below 40% (the cast frame is a free hit for the monster).
- Risk: casting at a unique that is already at melee range. Fallback: skip the curse, drink, back up.

## 4. Corpse policy

| situation | rule |
|---|---|
| before Revive (clvl < 30) | always Raise first. Corpses are free and Skeleton count is the whole damage budget |
| army below target count | Raise first (RS count, then Mage) |
| army at cap | Revive the best corpse: champion/unique pack members first, then highest-life regular monster |
| Corpse Explosion | do not cast (not trained beyond the prerequisite point); never at corpses you want |
| boss corpses | do not try: act bosses and quest monsters cannot be revived (EXP, UNVERIFIED list) |

- Limits: Revive max count equals skill level and revived monsters time out after about 2 minutes; mana 45 per cast (SRC table for mana; count and duration UNVERIFIED, read them from `skills.txt`). Revive gets +5% life and +10% damage per SM level (SRC class-necromancer).
- Policy: reserve nothing. Use a `tried` set (exists in `necro_corpse.go`, SRC readiness S6) so a corpse that fails twice is dropped. The Revive reserve rule only adds value after clvl 40; before that keep it simple.
- Corpse Explosion risk to minions: in 1.14d CE does not hurt your own minions (EXP, UNVERIFIED); the real cost is it destroys the corpse a Revive wants and it makes `debug minions` counts jump.
- Risk: spending 45 mana on a revive when 20 mana is needed to re-raise skeletons. Fallback: rule "re-raise before Revive, always"; Revive only when mana >= 45 + 25 + 20.

## 5. Stat points

Starting Str 15, Dex 25, Int(Energy) 25, Vit 15 (SRC). 5 points per clvl. Life +2 per Vit, mana +2 per Energy (SRC).

| clvl | summoner | caster |
|---|---|---|
| 2-10 | 3 Energy at clvl 2, then all Vit | 3 Energy at clvl 2, then all Vit |
| 11-24 | 5 Str per level until Str 25 (leather / hard leather / studded), the rest Vit, plus 1 Energy per level | same, Energy 2 per level (Nova and Bone Spear are cast) |
| 25-40 | 4 Vit + 1 Energy | 3 Vit + 2 Energy |
| 40+ | all Vit, Energy only if mana runs dry in a soak | 4 Vit + 1 Energy |
| Dex | never | never |

- Why: the necromancer has 45 life at the start and +1.5 per clvl (SRC), so Vit is the survival stat; summoners need mana only for re-raising (Raise 15-20 + Mage about 19 + Golem), roughly 120-150 per full re-summon at clvl 40 (SRC costs), so 1 Energy per level is the floor (EXP). The caster casts all fight, so it needs more.
- Strength 25 is for armor (Studded Leather 27 needs one more point, UNVERIFIED), stop there. Skull/helm and wand are light (EXP).
- Resist targets (EXP; Nightmare -40, Hell -100 all resist, SRC from the difficulty penalties noted in `docs/sme/mechanics-progression.md`, not re-checked): Normal none needed; Nightmare entering: 0 net minimum, 40 gear resist each by Act 3; Hell entering: 60 gear resist each at least, goal 100 each before Act 3 (net 0 is the floor, 75 is the human target and is not reachable without runewords for a bot). Verify live from the character stats read, not from the plan.
- Main risk: Vit-heavy caster runs dry of mana. Fallback: belt 2 mana columns and the "mana empty events" counter in the soak log; raise Energy only when it exceeds 3 per 10 minutes.

## 6. What kills an unattended Necromancer in Normal Act 1-2 (one rule each)

| killer | what the player sees | safety rule |
|---|---|---|
| Blood Raven (Burial Grounds), fire arrows, 60 life at clvl 8 | fire arrow hits every second, skeletons die first, then you | clvl >= 10 with Clay + 3 skeletons; stay behind the minions, retreat if life < 50%; never fight with < 2 healing potions |
| Cold Plains packs (clvl 5-8, fallen and gore packs) | 8 monsters, the 3 skeletons die and the necromancer is next | only after clvl 7 and RS >= 3; fight at a door or the bridge; re-raise before the next pack, never after the first hit |
| Countess (her keep) | fire bolt hits, the level below has unique packs | not before clvl 14 with Clay up and 4+ potions; skip the keep runes if the minion count is 0 |
| Andariel poison (Catacombs 4) | green life bar, no regeneration, the minions all die in 3 s | antidote potions or town after the first poison tick below 60% life; clvl >= 20 and Iron/Clay up; leave the fight when potions < 3 |
| (extra) Act 2 packs and Radament | packs of 6+ with melee ranged mix | same re-raise-first rule; do not clear the lair at < 8 skeletons |

- Verify live: log lines for min life, minion count at fight start, potions used, and town-trip reason. A death under these rules points at the rule, not at the class.

## Gaps for the main agent

- `docs/sme/class-necromancer.md` has no curse chain, Revive or Poison Nova numbers: queue a research card to read `skills.txt` for Revive count and duration, Lower Resist, Decrepify and Poison Nova damage rows.
- No source for the guide-standard caster was fetched here (see section 2).
