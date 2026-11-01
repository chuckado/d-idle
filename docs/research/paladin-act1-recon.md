# Paladin Act 1 recon (second-class probe)

> status: recon, read-only, current. Card `palrecon`. No game, no code run. Builds on `docs/classes/paladin.md`, `docs/classes/smoke-paladin.md`, `docs/combat.md` (class section) and `docs/sme/class-paladin.md`.

Tags: **V** read in this checkout (file named), **D** data (a repo data file or a table the SME doc computed from the install, tagged T there), **I** inferred from reading, **U** unverified. Classic 1.14d only. The audit BUILDFLEX (`agent/buildflex`) had not written `docs/research/build-flexibility-audit.md` when this was written (V: `git show agent/buildflex:...` fails, the branch is at main), so nothing here uses it and section 5 should be reconciled with it when it lands.

## 0. Plain words

- **The Paladin is not a blank page.** The bot already has a Paladin combat profile, a melee fight loop, aura handling, two Paladin plan tables (`paladin-normal30` and the fresh-character `paladin-normal30-v2`), Paladin weapon/shield gear preferences, a hand-switch binder, and a live record: a clvl 13 Paladin farmed Blood Moor and Cold Plains at 20 to 23 kills per minute with 0 deaths, and a fresh honest-play Paladin reached only clvl 4 in about 70 minutes (V `docs/combat.md`, `docs/classes/paladin.md`). So the "second class" question is partly answered already: the profile and table layer is data-driven and general; the rest (town, merc, drills, d2d) was written and proven on the Sorceress.
- **What is missing is evidence, not code.** Nothing Paladin has run on the d2d backend (the lab oracle all drills use) or in a drill; the d2d backend cannot hold the left button (Zeal) at all (V `docs/d2d-combat.md`, `docs/d2d-backend.md` G5). The honest-play run stalled on class-agnostic bugs (Den of Evil "cleared" without the boss, claim trip with no portal, poverty mode) that the Sorceress path has since fixed or avoided. Merc, repair, weapon swap and the Paladin belt were never run live.
- **How hard.** The offline probe (plan table through the leveler sim and the combat fakes) is small: 1 to 2 days of agent work, mostly tests. The live d2d probe needs one new backend capability (a held or repeated left attack) plus a class parameter in 3 to 4 drill labs. A first real Act 1 run (Den, Blood Raven, Cave, clvl 1 to 13) is feasible with today's profile; expect the early clvl 1 to 5 window to be the dangerous part (55 life, 15 mana).
- **Recommended starter build:** Zeal-first melee ("Zealot-lite", section 3), not Hammerdin. Blessed Hammer needs clvl 18, a ground-cast with wall checks and Concentration, none of which exist before Act 2 (D `docs/sme/class-paladin.md` s2). In Act 1 both builds are the same table anyway; Zeal is the one with a live record.

## 1. What already exists

| area | what is there | evidence | tag |
|---|---|---|---|
| Plan tables | `internal/plandata/build-paladin-normal30.json` (legacy, Maxroll order: Might, Sacrifice, Resist Fire x2, Holy Fire x4, Smite, Zeal+Charge) and `build-paladin-normal30-v2.json` (fresh Paladins only: Might x8, Sacrifice 1, Prayer 1, Smite 1, Zeal 4, then Holy Fire to 19, Holy Bolt, Blessed Hammer, Blessed Aim, Concentration to clvl 30). Both 29 rows (clvl 2 to 30), `respec` disabled. Endgame tables `build-paladin_hammerdin_endgame90.json`, `..._zeal_endgame90.json`; gear files `internal/gear/builds/paladin_*.json` | files read | V |
| Table choice | `leveler.SelectBuild`: BUILDV2 table for a fresh character, legacy when the character already owns a skill the v2 table dropped; wrong-class table is `ErrWrongClass` and pauses | `internal/leveler/build.go:196-260` | V |
| Combat profile | `internal/combat/classdata/profiles.json` entry `paladin` (no build field): bands clvl 24 + Blessed Hammer: [Blessed Hammer, Zeal, Smite, Attack]; clvl 12 + Zeal: [Zeal, Smite, Attack]; else [Smite, Attack]. Buff Holy Shield. Aura ladder Prayer (rest only, below 70% life), Holy Fire (fire element), Might, Thorns. Layout F1 Might, F2 Holy Fire, F3 Concentration, F4 Vigor, F5 Holy Shield, F6 Thorns, F7 Resist Fire, F8 Prayer. Melee knobs: reach 4, holdMs 600, retreatLifePct 40, surround 4 at 65%, manaReserve 4. Sorceress has the same file with per-build profiles; Barbarian has the same `melee` block; Amazon, Necromancer, Druid, Assassin entries exist and are U | file read; `docs/combat.md` s"Class profiles" says the five non-Sorceress non-Paladin profiles are U | V, U |
| Skill kinds | `classdata/skills.json`: Sacrifice, Smite, Zeal (`hold`), Charge (range 20) are `melee` left-hand; Holy Bolt `bolt` left; Blessed Hammer `ground` left `hold`; all auras `aura` range 0; Holy Shield `buff`. Ranges are Kolbot's, not game data | file read; `docs/combat.md` | V, U (ranges) |
| Skill ids | Verified live: Sacrifice 96, Might 98, Holy Fire 102, Zeal 106 (`docs/classes/smoke-paladin.md` finding 2). The rest of the Paladin block 96 to 125 (Smite 97, Prayer 99, Resist Fire 100, Holy Bolt 101, Thorns 103, Defiance 104, Resist Cold 105, Charge 107, Blessed Aim 108, Cleansing 109, Resist Lightning 110, Vengeance 111, Blessed Hammer 112, Concentration 113, Holy Freeze 114, Vigor 115, Conversion 116, Holy Shield 117, Holy Shock 118, Sanctuary 119, Meditation 120, Fist of the Heavens 121, Fanaticism 122, Conviction 123, Redemption 124, Salvation 125) is the stock Skills.txt order from memory. The bot reads skills.txt through `internal/data/skills.go` (`Skill.ID`, `Class`, `ReqLevel`, `ReqSkills`, `LeftSkill`); the extracted tables are NOT in this checkout, so I could not print the ids | U for ids other than the four live ones |
| Hands and binder | `Fighter.SetLeft` via the left popup (cell grid verified for Sacrifice, Smite, Holy Bolt, Zeal, Attack, about 1.1 s per switch); hotkeys select the right skill only | `docs/combat.md` | V (live, Palli) |
| Auras | State read from the player stat list: aura stat 350 = skill id, state 33 Might, 35 Holy Fire, 36 Thorns; `KeepAura` re-applies after death or area change; aura switch is a key press, 161 to 229 ms, no mana | `internal/combat/aura.go` header, smoke-paladin finding 6 | V (live). The state ids come from the live probe; `data.StateByID` reads `states.txt` at runtime, the table is not in this checkout (U for other aura state ids) |
| Melee loop | `Fighter.cast`: left skills click the target; `hold` skills use `HoldClick` (button down `holdMs`) when the input supports it. `meleeGo` lets the game walk the character on a click in sight (walk only when off screen or beyond reach), `meleeGuard` backs out when `surround` monsters are within 3 subtiles and life is under `surroundLifePct` or at `retreatLifePct`. No kiting for `KindMelee`; no low-mana wait or pre-pack rest for melee profiles | `internal/combat/fight.go:595-660`, `fighter.go:654-681`, `docs/combat.md` | V |
| Leveler data | Stat plan is Str-heavy (2 Str + 3 Vit to clvl 13), not all Vit; `prefs.go` Paladin: DmgWeight 0.8, two-hander x0.5, `WantsShield`; early gear policy: Paladin takes any usable melee weapon and any shield | `internal/leveler/prefs.go`, `docs/sme/early-gear-policy.md` | V |
| Merc | `internal/merc/policy.go`: free Rogue at Kashya, Act 2 Desert Mercenary by aura. `AuraPrefs` for a non-caster: Normal Defiance then Prayer; Nightmare Might first. The Paladin counts as non-caster (`caster` is Sorceress, Necromancer, Druid) | file read; run live only for the Sorceress | V code, U Paladin run |
| Potions and belt | `resource.PlanFor` and `belt.PlanForClass` are class-aware (the smoke doc read the Paladin plan as hp hp hp mp); `ManaNeed(Paladin)` is 1; `cheapestCast` ignores melee skills so a melee class never waits for mana | `internal/resource/types.go`, `internal/town/belt/belt.go:257`, smoke finding 7 | V code (I that the columns are as stated: the switch body was not printed) |
| Death loop | Class-agnostic and run live on Palli: respawn, equipped items to the corpse, recovery by walking | smoke finding 10 | V (live) |
| Tests | `TestPaladinBands`, `TestPaladinFightsWithLeftHand`, `TestPaladinFallsBackToAttackWithoutMana`, `TestPaladinDoesNotRestForMana` (combat); `TestPaladinBuildLoads`, `TestPaladinStatPlanIsNotAllVitality`, `TestPaladinSkillOrder`, `TestBuildV2PaladinSequence`, `TestPaladinEarlyPoints` (leveler); `TestWeaponProblemPaladin` (gear) | grep | V |
| Other classes | Amazon (smoke-amazon, ammo model `internal/town/ammo.go`), Necromancer (smoke-necromancer, `necro.go`, `necro_corpse.go`), Druid (`druid.go`), Assassin (`assassin.go`), Barbarian (research and table, no live run). Each has a plandata table and a profile entry | `docs/quickstart.md` s4, `docs/classes/` | V |

### 1.1 What the "other classes: short smoke runs only" means exactly (`docs/quickstart.md` s4, `docs/capabilities.md` row "Classes other than the Sorceress: Experimental")

- Paladin CLASSPAL 2026-10-28 (`docs/classes/smoke-paladin.md`): Palli on i1, cheated to clvl 6 then 13 with Zeal 3, Smite 3, Holy Fire 5, Str 60; live probes of hands, auras, melee, death; report-only. 14 findings: 3 BLOCK (no Paladin skills in combat, hands, Sorceress-only leveler), all since addressed by COMBATCLASS and the class-aware leveler (V `docs/combat.md`, `docs/leveler.md`). Still open from that table: 8 weapon swap (`w` flips body 4/5 with 11/12, rebuy would see empty slots), 9 gear score caster-shaped (partly fixed by `prefs.go`), 11 no merc hire or revive code, 12 cheat allow-list and sentinel map are per character name, 13 Teleport assumed by route data.
- Paladin COMBATCLASS (same doc set): 63 to 91 kills in 3 to 4 minutes per zone, 0 deaths, with cheated skills.
- Paladin PALALIVE 2026-10-29 (`docs/classes/paladin.md` "Live results"): fresh clvl 1 honest play, 4 runs of about 10 minutes, no cheats: clvl 4 after about 70 minutes, 19k xp/h, `melee: backing off surrounded` fired in a Fallen pack, drinks at life 35 to 40 of 88, 0 deaths. Stalls: Den of Evil "cleared" fired without a boss kill so the Akara claim failed; claim trip with no portal in poverty mode; planner oscillation Blood Moor / Cold Plains. Not Paladin specific (owner: town/quest).
- Amazon and Necromancer: smoke docs; Druid, Assassin: profile names checked against skills.txt only (U).
- None of these ran on the d2d lab backend or in a DRILLS drill (V grep: drill labs use skill ids for Fire Bolt, Ice Bolt, Lightning, Charged Bolt, Frost Armor: `cmd/d2drill/d13_lab.go`, `d15b_lab.go`, `d16_lab.go`).

## 2. What a Paladin Act 1 needs that the Sorceress does not

Normal, clvl 1 to about 13: Blood Moor, Cold Plains, Den of Evil (one reward point), Burial Grounds optional, Blood Raven, Cave, Tristram optional (I from `docs/sme/class-paladin.md` s10).

| need | what the Sorceress path does | what the Paladin needs | state of the code | tag |
|---|---|---|---|---|
| Approach and reach | Cast from up to 10 subtiles, kite, `castFirst` watchdog | Close to melee; click walks the character (`meleeGo`); profile reach 4, swing no-walk 2.5. The cast-first loop treats a melee skill's range as 7 (`defaultMeleeRange`); `castRange` uses the longest non-self attack range (Charge 20 once learned would stretch it, see 5) | works live; `crowdAdjacent` 8 for the portal cast rule is monster reach, not ours | V, I (Charge) |
| Melee left-hand skill | Right-click bolt on a hotkey | Left skill chosen in a popup (1.1 s), aura on the right; the ladder flips Zeal/Smite/Attack when mana drains | live-verified | V |
| Hold-to-attack | none | Zeal needs button held (600 ms per tick, 19 kills in 55 s held vs 7 in 30 s single clicks); Smite and Sacrifice are one click each. Sacrifice costs life per hit and is deliberately off the ladder in v2 ("never cast before Zeal") | live; **d2d has no hold** (`ErrUnsupported`, gap G5), so on d2d Zeal degrades to a single click | V, V (d2d) |
| Auras | Armors as buffs, recast by state/duration | Aura is a state toggled by selecting the skill on the right hand; costs 0 mana; `KeepAura` rechecks every `AuraCheckMs`, switches Holy Fire to Might against fire-immune targets; Prayer only while resting under 70% life. Charge/Vigor movement not used (Charge planned but "the bot does not use it yet" per the table note) | live (Might, Holy Fire, Thorns state ids) | V |
| Attack speed / animation | `docs/d2d-casts*.md`, `cast-trace-gameexe.md` and `fixcastgap.md` give per-skill cast gaps | No melee animation table in our docs: d2d measured base 15 ticks (Paladin, 0 WSM sword), WSM scaling matches the documented formula, IAS never tested (V `docs/d2d-fidelity.md` s7b, "U base"). So the melee swing gap, Zeal's 5-hit sequence length and IAS are unmodelled in `castgap.go` (cast gaps are per spell) | gap | V (d2d), I (castgap) |
| Mana | Pool is the economy; rest before packs | Pool 15 at start, +1.5 per clvl; Zeal and Smite 2 mana. Melee profile skips mana rests and ends the ladder in free Attack. Live: Zeal 33 -> 25 in 22 s, refills in 45 s | works | V |
| Life and potions | Low life 40 hp start | Melee takes hits: 55 life at clvl 1, 88 observed at clvl 4. The live run drank at 35 to 40 of 88; `retreatLifePct` 40 | works; early deaths U | V, U |
| Hit recovery and block | not modelled | A shield blocks only when Dex and block chance allow (Paladin block factor 30, D SME s1); no code reads block or FHR. Smite is a shield attack: damage from the shield, so a missing shield makes it weak (I; U) | not modelled; affects survival numbers only | I, U |
| Resists | Irrelevant in Normal Act 1 | Same (Normal has no penalty); Resist Fire is an unneeded prerequisite in the legacy table | none | D |
| Gear | Wand/orb +8, ignore armor | Any usable one-hander plus shield, Str for gear; `earlyGearMax` makes weapons matter for Paladin below clvl 40; rebuy must not buy a two-hander over a shield (`WantsShield`); weapon-swap read missing | partly done | V, I |
| Repair | Rarely needed | Melee loses durability on the weapon and armor: repair matters from the first Act 1 hours. The repair verb exists (Repair All button (369,452), D3.r drill: restores 20/20 for 19 gold on d2d, price model = game.exe 100,000/100,000, repair skipped at gold 0 after fix-bank) but live repair was never confirmed ("repair effect UNVERIFIED: nothing could be damaged", `docs/town.md`) | formula V, effect live U | V, U |
| Merc | Rogue then Act 2 Desert merc by aura | Same machinery; `AuraPrefs` gives Defiance then Prayer in Normal. The Act 1 Rogue is the tank-less archer (`pets-merc-trace.md`: the Rogue is no tank), so a melee Paladin takes more hits than a Sorceress behind it | untested for Paladin | V, I |
| Teleport | native, routes use it | none: Act 1 routes need no Teleport; Act 2 `task` steps with `NeedsTeleport` have a walk fallback only partly (smoke finding 13) | fine for Act 1 | V |
| Town travel | Waypoints, same | same; Paladin has no walk speed boost unless Vigor/Charge | none | I |

## 3. Starter build as data (PROPOSAL, text only)

### 3.1 Choice

| option | Act 1 table | extra code | bot risk | verdict |
|---|---|---|---|---|
| Zeal-first melee ("Zealot-lite") | Might, Sacrifice (1, prerequisite), Prayer, Smite, Zeal 4 at clvl 12 to 15 | none beyond what runs live | melee deaths in packs at clvl 1 to 8 (U, SME flags "weakest at clvl 1-6") | **recommended**: left-hand melee skills have a live record, no ground aiming, no wall check, no respec |
| Hammerdin-lite | Might to clvl 17, then Holy Bolt, Blessed Hammer (clvl 18), Concentration | ground-cast aimed at a free cell, hold, wall check, Concentration state, Holy Shield timer (smoke doc s8) | the first 17 clvls are plain Attack and Smite anyway; hammers vanish on walls | **not for the probe**: BH is not reachable in Act 1; both builds equal until clvl 17 |

Why Zeal is simpler: one left skill plus one aura, no aiming, no timers; the game walks the character on a click in sight (V `fight.go:651`). Zeal has 100% weapon damage over up to 5 hits, level 4 is the cap for hit count (D SME s2), 2 mana.

### 3.2 The table (clvl 1 to 24; rows 25 to 30 copy the v2 rows)

Assumptions: 1 skill point per clvl from 2, plus 1 for the Den of Evil (carried on the row where it normally arrives, as `build-fallback-lightning-v2.json` does with Frozen Armor at clvl 5); +5 stat points per clvl. Prerequisites and `reqlevel` below come from `docs/sme/class-paladin.md` s2 (D, the SME's tables read from the install: Sacrifice, Smite, Might, Prayer reqlvl 1; Zeal reqlvl 12 needs Sacrifice; Charge 12 needs Smite; Cleansing 12 needs Prayer; Vigor 18 needs Cleansing and Defiance; Defiance 6). I could not re-read `skills.txt` here (extracted tables not in this checkout), so every reqlvl and prerequisite below is **UNSURE until `TestPaladinSkillOrder` and a loader run confirm it** (the existing test already checks the v2 sequence against the loader).

Differences from v2 on purpose: Prayer 3 instead of Might 8 (the sustain aura for the thin early game, used only while resting), Smite taken at clvl 4 instead of 10 (U: whether the earlier Smite helps; costs 2 mana of a 15 pool; flagged for the SME), Vigor line added after Zeal 4 for the walking speed (clvl 18 and up, U value).

```json
{
 "build": "paladin_zealot_lite",
 "class": "paladin",
 "schema": "same as build-fallback-lightning-v2.json: per-clvl allocation applied on level-up",
 "respec": null,
 "rows": [
  {"clvl": 2,  "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Might", "points": 1}]},
  {"clvl": 3,  "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Sacrifice", "points": 1}]},
  {"clvl": 4,  "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Smite", "points": 1}]},
  {"clvl": 5,  "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Prayer", "points": 1}, {"skill": "Might", "points": 1}]},
  {"clvl": 6,  "strength": 1, "dexterity": 1, "vitality": 3, "energy": 0, "skills": [{"skill": "Prayer", "points": 1}]},
  {"clvl": 7,  "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Might", "points": 1}]},
  {"clvl": 8,  "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Might", "points": 1}]},
  {"clvl": 9,  "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Prayer", "points": 1}]},
  {"clvl": 10, "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Might", "points": 1}]},
  {"clvl": 11, "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Might", "points": 1}]},
  {"clvl": 12, "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Zeal", "points": 1}]},
  {"clvl": 13, "strength": 2, "dexterity": 0, "vitality": 3, "energy": 0, "skills": [{"skill": "Zeal", "points": 1}]},
  {"clvl": 14, "strength": 1, "dexterity": 0, "vitality": 4, "energy": 0, "skills": [{"skill": "Zeal", "points": 1}]},
  {"clvl": 15, "strength": 1, "dexterity": 0, "vitality": 4, "energy": 0, "skills": [{"skill": "Zeal", "points": 1}]},
  {"clvl": 16, "strength": 1, "dexterity": 0, "vitality": 4, "energy": 0, "skills": [{"skill": "Defiance", "points": 1}, {"skill": "Cleansing", "points": 1}]},
  {"clvl": 17, "strength": 1, "dexterity": 0, "vitality": 4, "energy": 0, "skills": [{"skill": "Might", "points": 1}]},
  {"clvl": 18, "strength": 1, "dexterity": 0, "vitality": 4, "energy": 0, "skills": [{"skill": "Vigor", "points": 1}]},
  {"clvl": 19, "strength": 1, "dexterity": 0, "vitality": 4, "energy": 0, "skills": [{"skill": "Vigor", "points": 1}]}
 ]
}
```

Rows 20 to 30 would be the v2 rows 20 to 30 (Holy Bolt, Blessed Hammer, Blessed Aim, Concentration, Blessed Hammer x7; 1 Str + 2 Dex + 2 Vit): Act 1 never reaches them (I). Point count check by clvl 13: 12 level points plus 1 Den point = 13 = Might 6 + Sacrifice 1 + Smite 1 + Prayer 3 + Zeal 2 at clvl 13 (the row-5 Den point is the second skill of that row). U: the Den point lands on row 5 only if the Den is cleared at clvl 5; the loader applies it when it arrives, so the order matters, not the clvl (V `build-fallback-lightning-v2.json` note).

Stats: Str 2 + Vit 3 per level to clvl 13 (the v2 stat plan, V), because Str gates armor and shields (smoke finding 9: Str 60 was enough in the test); Dex 1 at clvl 6 as Maxroll (S, `docs/classes/paladin.md`). Block needs Dex; skipped in Act 1.

`hotkeys` block (copy of v2): F1 Might (right), F2 Holy Fire (right, from clvl 6: this table has no Holy Fire, so the profile layout keeps F2 unused: harmless, I), F4 Zeal (left, from clvl 12), F5 Charge, F6 Blessed Hammer, F7 Concentration. The hotkey list is "plan for the combat agent, never pressed" (V `docs/leveler.md`); the live binder uses the profile layout.

Open points in the table (UNSURE): Prayer 3 vs Might (does Prayer matter if the code uses it only while resting); Smite at clvl 4; Defiance and Cleansing at clvl 16 solely as Vigor prerequisites; whether Charge belongs at all before Holy Shield.

## 4. The generality metric

Question: how much of the bot is class-general? Score = number of Go files outside `internal/plandata` and class profile data (`internal/combat/classdata/*.json`) that had to change to run the table in section 3 on a Paladin.

Counting rules (so the score cannot be gamed):
1. Run A, offline sim: the Paladin table through the leveler (`leveler` plan tests, `TestPaladinSkillOrder` pattern), one combat fake fight with `combat` fakes (`scenario_test.go` style), `goals/sim` with `Class: "Paladin"`, `resource` sim. A changed file counts once. Changed test files count separately from production files (report both numbers).
2. Run B, one drill on d2d (hidden instance, throwaway Paladin, `--start-class 2`): one fight drill plus one town drill (repair at the smith). The drill labs count as files.
3. A file counts if it needed any edit to run the Paladin end to end, even a one-line class check. Files only touched to add Paladin test cases do not count toward the product score.
4. Record, for each counted file, the cause: **literal** (a Sorceress name or id), **model** (a mechanism that does not exist for melee: hold, aura, swing animation), **backend** (d2d cannot do it), **bug** (class-agnostic bug the Paladin exposed).
5. Report the score as `product files / model files / backend files / bug files`. A change request is "general" when it is needed by the Paladin run too, "bespoke" when the Paladin run passes without it.

### 4.1 Prediction (to compare with reality)

Production files expected to change (I, from reading; count in brackets):

| file | why | class |
|---|---|---|
| `internal/input/d2d` or `internal/d2dinput/*` (held left attack; `HoldClick` or an attack-continue mode) | d2d has no hold: Zeal degrades to one click per tick (`docs/d2d-combat.md`, G5) | backend |
| `internal/d2dctl/combat_verbs.go` (`Melee`, left-hand `CmdSkill` with a monster id) | `Interact(monster)` path is "U" in the d2d backend table | backend |
| `cmd/d2drill/*_lab.go` (3 to 4 files: the fight/portal/meter labs assume Fire Bolt, Charged Bolt, Frost Armor) | skill ids hard-coded to Sorceress | literal |
| `cmd/d2dsoak/lab.go` and `run.go` (class flag exists; the coverage and gate lines name the class) | probably 0 to 1 | literal |
| `internal/combat/castgap.go` (cast gap per skill: no melee swing or Zeal sequence length) | melee animation gap unmodelled | model |
| `internal/town/panel/panel.go:74-87` (class row for the inventory layout; default returns `"Sorceress"`) | smoke finding 14: same pack grid, so only if the Paladin grid differs | literal |
| `internal/bot/geargate.go`, `internal/bot/task.go` (Caster flags; BRHigh gate floors for Blood Raven) | Paladin is not Caster: floors differ; a melee gate needs checking | model |
| `internal/merc/policy.go` (AuraPrefs: Paladin gets Defiance/Prayer like a Sorceress in Normal) | likely fine, listed to confirm | none expected |
| `internal/town/rebuy.go` / leveler gear score for a one-hander plus shield | smoke findings 8 and 9 (weapon swap read, rebuy) | bug |
| `internal/cheat/policy.go` (`Allowed` names) and `internal/sentinel` instance map | smoke finding 12 | literal |
| `internal/goals/sim/sim.go`, `internal/resource/sim.go`, `internal/town/sim/*.go` (default `Sorceress`, world starts a Sorceress) | class param exists in goals/sim; town sim hard-codes the class | literal |

Predicted score: **product files 5 to 8, of which backend 2, literal 3 to 4, model 1 to 2, bug 1 to 2**. Zero changes expected in `internal/leveler/*` (class-aware already), `internal/combat/fight.go` and `fighter.go`, `internal/plandata` loader, `internal/merc`. Surprise candidates: (a) `bot/plan.go` areas or poverty rules because a melee class earns less gold and gets hit more (V PALALIVE oscillation), (b) the Den of Evil "cleared without the boss" bug (not class related, would be found again), (c) `castFirst` watchdog treating Zeal's hold as "no cast for 6 s" (I from `castfirst.go`: casts are noted per `cast()` call, so fine).

Interpretation: a count at or below the prediction supports "the class layer is general"; a count much above it with `literal` dominating says the Sorceress build is baked into labs and sims rather than the bot proper.

## 5. Dependencies and risks

- BUILDFLEX audit is not available yet (see header); reconcile this prediction with it when it merges. Anything it lists that I did not predict is a finding.
- Charge (planned at clvl 16 in the legacy tables, range 20 in `skills.json`) would lengthen `castRange`, the distance within which `castFirst` forces a cast: with Charge learned the Paladin might "cast" at monsters 20 subtiles away. The ladder omits Charge, so `attackSkills` does not include it (V profiles.json), so I rate this low; confirm.
- d2d fidelity is "unvalidated" for melee: monster damage, attack rate and AR were not measured (`docs/d2d-fidelity.md` s258), so a d2d drill proves mechanics (hold, switch, repair), not survival.
- Live on the real game: the Den of Evil claim path and no-portal claim trip were open bugs at PALALIVE (V `docs/classes/paladin.md`); check they are fixed on main before a live probe (I could not run anything).

## 6. Questions for the d2-sme (not asked)

1. Zealot-lite or Hammerdin-lite for an unattended Normal Act 1 to 13 run with no respec: is Zeal-first at clvl 12 really the sturdiest, and is the clvl 1 to 11 plain Attack plus Might phase survivable with 55 life, or should Sacrifice/Smite come earlier?
2. Smite at clvl 4 vs 10: does Smite (a shield attack) out-damage plain Attack with a Buckler at clvl 4 to 11, and is the 2 mana per use worth a 15 mana pool and a 1.1 s hand switch?
3. Auras in Act 1: Might vs Holy Fire vs Prayer. How many points in Might before Zeal (our v2 gives Might 8)? Is Prayer useful while fighting, or only resting? Should the bot ever run Resist Fire vs Blood Raven's fire arrows?
4. Attack skill: Zeal vs Sacrifice at clvl 12 to 17 (Sacrifice costs life, we leave it off the ladder): confirm, and say what Zeal level and how many IAS-style gear stats make the hold worthwhile.
5. Belt: our plan treats the Paladin as hp hp hp mp. For a melee Act 1 run, is hp hp hp hp (no mana) the right plan, and how many health potions should the bot carry into Den, Blood Raven and the Cave?
6. Repair: at what durability should a melee Paladin repair in Normal Act 1, and when does a repair trip cost more than it saves (gold per hour at clvl 8 to 13)?
7. Merc: hire the Act 1 Rogue for a melee Paladin (it is no tank), or skip it and save gold for potions and the Act 2 Desert Mercenary with Defiance or Prayer?
8. Block and shield: is a Buckler enough in Act 1, and does the bot need any Dex for block before clvl 20?

## 7. Shared-file lines for the merger

- BACKLOG, ROADMAP: `second class probe: Paladin Act 1 (user go 2026-11-01); recon docs/research/paladin-act1-recon.md (card palrecon): Paladin profile, plan tables and live smoke exist; missing d2d held attack and a class parameter in drills; build after the BUILDFLEX audit`
- INDEX: `- Paladin Act 1 recon: [research/paladin-act1-recon.md](research/paladin-act1-recon.md) (what exists for the Paladin, Act 1 melee needs, Zealot-lite starter table as a proposal, generality metric and predicted file list, 8 SME questions; card palrecon)`
