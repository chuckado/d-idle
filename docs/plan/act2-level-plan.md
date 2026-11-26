# Act 2 level plan (Normal): where the hero farms between the Lut Gholein quests

> status: DATA ONLY (card act2plan, 2026-11-26). Nothing here ran live. Tags: V verified (named file or test), S secondary (the d2-sme answer, from memory, no source fetched), I inferred, U unverified. Every number below is U until the margin model or a seed set moves it (docs/process/threshold-governance.md).

Files: `internal/plandata/act2plan.json` (copy `docs/plan/data/act2plan.json`, equal by `TestMatchesDocs`), loader `internal/plandata/act2plan.go`, tests `internal/plandata/act2plan_test.go`, road rows `internal/plandata/admission.json` (ids `act2_*`).

## Why a separate file (V)
`internal/levelplan` (levelplan/3) cannot carry Act 2 today: its test requires an AREAMON row for every farm area and `internal/levelplan/areaxp.csv` holds Act 1 only (V file); the Normal spine `data/campaign/normal.json` has Act 1 chapters only, so quest slices have no chapter to name (V). The Act 2 slices use the levelplan/3 field names so a later merge into levelplan-normal.json is a copy once the XP rows (an AREAMON run for Act 2 areas, game.exe, lab) and the Act 2 chapters exist. XP per minute is therefore NOT in this plan; entries and exits are the SME ballpark, not an XP-crossing rule.

## Slices (Paladin melee / Sorceress caster)
| slice | areas | Paladin clvl | Sorceress clvl | max-life gate Pal / Sor | way in |
|---|---|---|---|---|---|
| dry_hills | 41,42 | 15-20 | 15-20 | - / - | waypoint 42 |
| far_oasis | 43 | 17-21 | 17-21 | - / - | waypoint 43 |
| lost_city | 44,45 | 19-23 | 19-23 | - / - | waypoint 44 |
| halls_of_the_dead | 56,57,60 | 20-24 | 22-24 | 160 / 90 | waypoint 57 |
| claw_viper_temple | 58,61 | 21-25 | 23-25 | 180 / 100 | walk 44 |
| ancient_tunnels | 65 | 22-27 | 24-27 | 200 / 120 | walk 44 |
| arcane_sanctuary | 74 | 24-29 | 26-29 | 220 / 150 | waypoint 74 |
| tal_rasha_tombs | 66,67,68,69,70,71,72 | 24-30 | 26-30 | 220 / 150 | waypoint 46 |

Without a living merc add 2 levels to every entry from halls_of_the_dead on (`no_merc_add`).

Quests walk through these areas under the journey (data/questboard/act2.json); Radament (Sewers 47-49), the Tainted Sun, the Horadric Staff, the Summoner and Duriel (73) are not farm slices. Duriel's row (`duriel` in the file, road row `act2_tombs`): melee 24, caster 26, life 220 / 150 (floors = Seven Tombs legacyFloor V, life SME S).

Entry rule of the data: melee entries come from the SME table; the Sorceress's dungeon entries follow the caster `legacyFloor` (V, +2 over melee) although the SME suggested one level LESS for a caster with a merc (S). The repo's own Act 1 caster row (admission.json `underground_passage`, minLevel melee 7, caster 8, V) also runs later, so the plan keeps the safer reading until a seed set says otherwise.

## Admission rows (entry floors)
`internal/plandata/admission.json` roads `act2_sewers`, `act2_halls_of_the_dead`, `act2_claw_viper_temple`, `act2_maggot_lair`, `act2_palace_cellar`, `act2_ancient_tunnels`, `act2_arcane_sanctuary`, `act2_tombs` (tombs plus Duriel). Dungeons only; the open fields are plan rows without a floor. Level floors are the questboard `legacyFloor` of the quest that walks there (V data/questboard/act2.json); the Ancient Tunnels floor is plan entry minus 3 (I); life floors for the Arcane Sanctuary, tombs and Duriel are SME (S). Kit: 2 healing potions and a town portal (same as the Act 1 road). All rows `evidence: 1` with ref "needs seed set" (n=1 seeds a row, it may not move one). Tests: `TestAct2PlanShape` (a slice entry is never below a road floor of its areas), `TestAct2AdmissionRoads`.

## Mercenary
Greiz (Desert Mercenary, Act 2): hire from the first Lut Gholein visit, keep through Normal, always revive (`merc` in the file). Aura default Defiance for both archetypes (V code: `quest.Act2Merc` default; hireling.txt is read at the hire screen, docs/act2.md). Revive price: `internal/merc.ReviveCost` = min(int(lvl^2/2)*15, 50000) (V code); the SME's "50-250 gold" guess is NOT used. Hire and revive code exists (`internal/merc`, `task.QMercAct2`); a live hire never ran (U).

## Survival gates
Per slice `min_max_life`, plus the kit and the admission road rows. Where the SME named numbers (Pal 220+ and Sor 150+ before the tombs and Duriel) they are used; earlier rows are interpolated (I).

## SME answer (d2-sme, asked once on 2026-11-26), VERBATIM, tagged S
The SME states it fetched no source: by its own words everything below is from memory. Not checked against levels.txt (U). The repo's V values override it where they differ (for example the revive formula above).

> VERDICT: depends. I did not fetch any source this turn. Everything below is EXP (veteran memory), and the mlvl numbers are from memory of the Normal area-level table. Treat them as +/-2 and check them against the repo's extracted levels.txt data or Arreat Summit before using them as hard gates. Normal Act 2 area levels run about 14 to 24. Confidence: medium for the overall span, low for individual areas.
>
> Areas (mlvl; entry clvl Pal / Sor+Desert merc; danger):
> - Rocky Waste 14-16; 15 / 14; Dung Soldiers, Sand Maggots, Carrion Birds.
> - Dry Hills 15-17; 16 / 15; Dark Hunters, Huntress packs.
> - Far Oasis 16-18; 17 / 16; Fangskin (Claw Viper Temple 2 boss) is in the temple. The oasis has Sand Raiders and Maggots.
> - Lost City 18-20; 19 / 18; Dark Shapes, Vipers, Sand Leapers.
> - Valley of Snakes 19-21; 20 / 19; Cobra and Viper packs, Dark Lord.
> - Canyon of the Magi 21-23; 22 / 21; Mummy, Bone Ash/Spirits. Tomb entrances lead off it.
> - Sewers 1-3 (Radament) 14-16; 17 / 16; Radament has Hydra-like "poison/curse" abilities and raises Mummies. He is a quest boss (Book of Skill). The Sewers have dark corridors and many Undead.
> - Harem/Palace Cellar 1-3 17-19; 18 / 17; Council Members hit hard. Blunderbore (Cellar 3) is a tough Unique.
> - Stony Tomb 1-2 17-18; 19 / 18; Beetleburst (Beetle), Mummies.
> - Halls of the Dead 1-3 18-20; 20 / 19; Dark Shapes, Mummy packs, Mummies raise. Hard in the maze Halls 3.
> - Claw Viper Temple 1-2 20-22; 21 / 20; Fangskin boss (Fire/Cold-based Viper). Viper Mages and Venom Lords.
> - Maggot Lair 1-3 20-22; 21 / 20; Coldworm the Burrower (cold, Maggot lair 3). Very slow; avoid melee.
> - Ancient Tunnels 22-24; 22 / 21; Fire Demons, Sand Raiders, Dark Shapes, large packs.
> - Arcane Sanctuary 22-24; 23 / 22; Summoner (quest, Arcane Sanctuary 4); Specters, Arcane Horrors, Horror, Fire Imps. Teleport-heavy; very dangerous below ~23.
> - Tal Rasha's Tomb 1-7 22-24; 24 / 23; Tomb guardians. Only one real tomb is Tal Rasha's, others hold Horrors, Ghosts, Burning Souls; Mummies, Zombie packs; cold/lightning-immune Mummies are common.
> - Duriel's Lair 24; 24 / 24; Duriel has Holy Freeze (cold aura), slows you ~50%, and Crushing Blow. It hits hard in Normal. Duriel mlvl 24. Prepare cold resist and Frost Nova/Frozen Armor mitigation.
>
> (Confidence on individual mlvl values: low-medium. On entry clvl: medium, because it depends on gear.)
>
> Farm bands (medium): Sewers 14-18; Dry Hills/Far Oasis 16-22; Lost City / Valley 20-24; Canyon/Tombs 22-27; Duriel at 24-27.
>
> Life, potions, resists (medium):
> - Minimum life: Pal 220+ before Tombs/Duriel, Sor 150+ (with Frozen Armor or Energy Shield). Below that, one Duriel Crushing Blow or Boss pack is death.
> - Belt: 4 columns, and fill with rejuvenation/full-rejuv potions mixed. Carry 2 rejuvs.
> - Resists: Normal has 0 penalty. Aim for 0%+ all resists at Duriel; cold 20%+ is comfortable and Holy Freeze ignores... hold: Holy Freeze is a cold aura, so cold resist matters most.
> - Pal: Resist cold aura via Holy Shield/Defiance is not available at that level, so use gear.
>
> Greiz (Act 2 merc; medium): hire as soon as you reach Lut Gholein (clvl ~12-14); he pays off immediately. Desert Mercenary cost varies by level, typically ~60-120 gold at hire in Normal. Revive cost is about (merc level) x (a multiplier); expect roughly 50-250 gold in Normal. UNSURE: exact formula, so read it from the repo's merc data. Revive is cheap enough that you should always revive.
>
> Best XP/hr at clvl 18-26 (medium, EXP): Lost City/Valley of Snakes, Claw Viper Temple, Halls of the Dead, Ancient Tunnels (best density, clvl 22+), Maggot Lair is poor for time. Radament is a fast quest XP. Arcane Sanctuary yields good XP but has high death risk; leave for clvl 24+.
>
> Gap: docs/sme/ may not hold a per-area mlvl table for Act 2. Queue research (levels.txt extraction) to turn the low-confidence numbers into SRC.

Note (V): the SME's "Pal: Holy Shield/Defiance not available" and its revive guess are doubtful; not used.

## Not done / gaps
- AREAMON XP rows for Act 2 areas (a lab game.exe run): without them no XP per minute and no simulated trajectory.
- Act 2 chapters in `data/campaign/normal.json` (so levelplan quest slices can name Radament, the Staff, Duriel).
- Maggot Lair (62-64) and Stony Tomb are quest dungeons, not farms (SME: poor); only the Lair has a road row.
- Live: no floor was tested; the lab drill rows (data/stepgrounds) are CHEATED setups.
