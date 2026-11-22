# Appendix A. The combat engine today, per skill kind

> status: REVIEW (card fablecombat, docs only). Part of [combat-engine.md](combat-engine.md). All file and line references are to this tree at `5b12ce5e7`.
>
> Tags: **[V]** = I read the code at the place named; it says what the engine DOES, not that a game ever showed it. **[V live: ...]** = a live observation recorded in the repo doc named (S for this review). **[U]** = unverified.

## 1. How a skill reaches the game today

| step | what happens | where |
|---|---|---|
| table | a skill has `kind` (bolt, ground, self, buff, melee, aura, summon, charged), `range`, `hand`, `hold` | `internal/combat/classdata/skills.json`; parser `classdata.go:148-175` |
| fallback | a skill a profile names but the table lacks is classified from skills.txt: aura, summon, melee by `h2h`, bolt when it has a missile, else ground, else buff | `classdata.go:181-205` |
| spec | mana from the skills.txt formula rounded up, cooldown = `delay` frames x 40 ms, prerequisites; one skill-name special (Static Field range by level) | `skills.go:87-122`, special at `:112-114` |
| order | the active band of the profile (first band whose level and `requires` are met), minus `avoid`, minus life-cost skills when not allowed, minus unbound skills, minus skills the gear rules out | `classdata.go:257-282`, `fighter.go:826-860` |
| pick | first in order that the target is not immune to, that is affordable and off cooldown; in Nightmare and Hell reordered by resist | `skills.go:175-204`, `rules.go:193` |
| aim | the target unit's current position | `fight.go:646` |
| reach | walk until in range, on screen and in line of sight; give up after N approaches | `fight.go:711-751` |
| send | left hand on a swinger backend: `Swing(id, point)`; left hand else: click or held click; right hand with a key: `CastAt(key, point)`; right hand selected: right click | `fighter.go:778-805` |
| after | time and skill noted; nothing is verified | `fighter.go:800-803` |

Which kinds change the loop's behaviour [V, `grep -n 'Kind(Bolt|Ground|Self|Buff|Melee|Aura|Summon|Charged)' internal/combat/*.go`, tests excluded]:

| kind | read by | effect |
|---|---|---|
| `KindMelee` | `fight.go:648`, `:662`, `:695`, `:712`, `:721`, `:723` | melee guard, no kite, reach from the profile, "click walks there" shortcut |
| `KindSelf` | `fight.go:652`, `:695`, `:723`; `castfirst.go:54`, `:158-161` | no kite, range not extended, excluded from the watchdog and from cast range |
| `KindBuff` | `castfirst.go:54`, `:158-161`; `rules.go:202` | excluded from attacks |
| `KindAura`, `KindSummon` | `skills.go:109`; `rules.go:202`; `necro.go:191`; `druid.go:125` | aura costs nothing; excluded from attacks; a summon must be this kind to be cast by the side loops |
| `KindBolt`, `KindGround`, `KindCharged` | nowhere outside the parser | none |

## 2. The map, per kind

### 2.1 Melee (K4)

- Works. Left-hand loop: on the product backend the blow is sent by skill id at the unit near the point (`fighter.go:783-784`, `internal/d2dcombat/input.go:99-120`), because a bare click there acts on the previous frame's hover unit [V live: docs/palmelee.md, S].
- Positioning is a `melee` block in the profile: reach, hold time, retreat and surround lines, mana reserve (`classdata.go:67-74`), used by `meleeGuard` (`fight.go:1016-1051`). The block decides "is this a melee class" (`classdata.go:294`), which then switches target ranking (`fight.go:388-389`), the mana ladder (`fight.go:636`) and the reach rule.
- Life-cost skills are a profile list with three knobs named after one skill (`classdata.go:85-87`, `lifeprice.go:54-60`, `internal/env/tuning_combat.go:86-89`).
- Dual wield is a gear filter (`gearcheck.go:82-94`); one class literal sits there (`gearcheck.go:95`).
- Live proof: Paladin Zeal and Sacrifice [V live: docs/classes/smoke-paladin.md]. Barbarian, Amazon Jab, Assassin, Druid melee rows: code only [U].

### 2.2 Single-target missile (K1)

- Works; it is the default path. Live proof: Fire Bolt [V live per `classdata.go:20-21`], Charged Bolt and Ice Bolt by the same path [S capabilities.json reasons].
- Cast length is modelled for the Sorceress only, by a name list (`castgap.go:17-24`, `:30-46`); every other class uses the flat `CastGapMs` (`castgap.go:51-61`).
- On the product backend the cast is sent at the living monster within one cell of the point (`d2dcombat/input.go:63-85`), so a "missile" is unit-targeted there.

### 2.3 Self-centred nova (K2)

- A `self` skill in a band would be cast when the target is within its range (`fight.go:722-725`); no band lists one [V profiles.json].
- Two special rules by skill name: Static Field as an opener on elites and bosses and as the immune fallback (`fight.go:808-833`, also `affix.go:142`), with a name test for one boss (`fight.go:813`); Frost Nova on a crowd of 3 within 4 subtiles (`castfirst.go:98-126`).

### 2.4 Aura upkeep (K5)

- Works and is data-driven: a ladder of `{name, elem, restLifePct, needsCrowd}` per profile or band, one generic chooser (`aura.go:85-126`), re-applied after death or area change.
- The active aura is read from the player's stat list in game memory (`aura.go:30-68`) [V live: Might, Holy Fire, per the comment at `aura.go:14-16`]; without that it trusts the selected right skill.
- Leftover: the crowd threshold is a knob named after one aura (`aura.go:105`, `tuning_combat.go:90`). No "while walking" slot (Vigor) [S hammerdin-support H4].
- On the product backend the switch went through popup clicks until the `input.select` verb (devp14, loaded, not switched) [S docs/process/sandbox-devp14.md].

### 2.5 Timed buff upkeep (K6)

- Works by clock for any skill in a profile `buffs` group: one cast per group, first learned and bound member (`fighter.go:685-731`), on need-buff edges and inside long fights (`fighter.go:754-762`), skipped in a crowd, retried when a hit cut the cast (`fighter.go:716-726`).
- "Is it up" is a state read for five skills only, in a name-keyed map (`necro.go:42-43`); the others run on `BuffRecastSec` or the skills.txt duration (`necro.go:473-507`). State numbers for Holy Shield, the war cries, Cyclone Armor, Fade, Burst of Speed: unknown [S hammerdin H6, barbarian B3, druid D4].
- Curses are explicitly excluded from the buff pass by a name map (`fighter.go:695`, `necro.go:44-45`), because their table kind is `buff`.
- Party range, two-step buffs (Battle Command before Battle Orders) and weapon-swap prebuffs: none [S barbarian B3, B6].

### 2.6 Summons upkeep (K7)

| who | code | proof |
|---|---|---|
| Necromancer | census by monster class ids in a switch (`necro.go:59-69`, `:72-97`); plan golem, skeletons, mages, Revive by skill name (`necro.go:182-240`); walk to a corpse out of fights (`necro.go:245-285`); cap learned when mana was paid and nothing grew (`necro.go:309-314`) | [V live 2026-10-29: skeletons, mages, Clay Golem; S docs/summon.md]. Fire Golem name mismatch `Fire Golem` against the table's `FireGolem` is open [S necromancer-support S5] |
| Druid | census by class id constants, several U (`druid.go:16-24`); plan list with skill names and caps in Go (`druid.go:105-111`) | never cast live [V comment `druid.go:10-14`] |
| Assassin | shadow by two skill names (`assassin.go:88-110`) | never live |
| Amazon | Valkyrie is in the profile's `buffs` group, so the buff clock recasts it; no census | [V profiles.json amazon row; effect U: a recast may replace a healthy Valkyrie] |
| Sorceress Hydra | recognised as friendly, no cast code | [S capabilities.json reason] |

All three side loops are entered through a class switch (`druid.go:230-256`) and share one memory struct named after the Necromancer (`necro.go:133-138`). Minion ownership is inferred from monster class and AI because the owner words read 0 for Necromancer minions (`friendly.go:83-87`) [V live per that comment].

### 2.7 Curses and other marks (K9)

- Necromancer only: target = boss or elite, else the densest pack member within 20 (`necro.go:323-348`); curse chosen by an if-chain over three names (`necro.go:353-372`); "already cursed" by a state when the state number is proven readable, else a 30 s cast memory with a 5 subtile cover radius (`necro.go:376-402`).
- State numbers: Amplify Damage 9 [V live per `necro.go:28`], Weaken 19 [S], none for the other eight curses.
- Fidelity: the curse put rule differs on d2d (gap A2) [S review 02].

### 2.8 Ground-targeted area spells (K3)

- No driver. A `ground` skill in a band is cast at the target's tile (`fight.go:646`, `:799`). No lead for the spell's delay, no cluster choice, no effect radius in the data, no check that anything was hit.
- Line of sight IS checked on the normal path (`fight.go:726`) and is NOT on the watchdog path (`castfirst.go:157-179`).
- Cooldown: generic from skills.txt `delay`, on the bot's clock (`skills.go:118-120`, `fighter.go:862-868`). On d2d the game's own delay was not enforced at the last check (gap A5) [S review 02], so the two can disagree.
- Never live for any skill [S capabilities.json: Blizzard, Meteor, Firestorm rows "U live"].

### 2.9 Traps (K10) and corpse targets (K8)

- Traps have the best placement code in the package: pack centroid clamped to range and a minimum distance (`trap.go:165-177`), a census with a pending-cast memory, a cap of five, renewal. All keyed by a skill-name ladder (`trap.go:32`) and one name test (`trap.go:196`). Never live; d2d Assassin is untested against game.exe (gap A7) [S].
- Corpses: living-or-dead mode test and filters (`necro.go:102-130`); Corpse Explosion picks the corpse with the most hostiles within 6 (`necro_corpse.go:47-95`); Revive takes the highest-level corpse. One shared "tried" set arbitrates between raise, revive and explode [S necromancer-support S6].
- On the product backend a corpse cast carries NO unit: the adapter attaches a unit id only for a living monster near the point (`d2dcombat/input.go:68-75`, `internal/d2dctl/combat_verbs.go:19-31`) [V]. Whether d2d resolves the corpse from the point is U.

### 2.10 Movement skills (K12, K16)

- Teleport: used by navigation only, by hotkey (`nav/teleport.go:11-25`, `nav/telecap.go:3-10`), the key supplied by the Fighter by skill name (`telekey.go:18`). Hops are chosen from the known grid by distance and verified by position afterwards [V comment; V live: docs/act2.md, S]. Combat never teleports (no reposition, no escape).
- Charge has a melee row with range 20, Whirlwind a melee row with `hold`, Leap Attack a melee row; none is in a band, none has code [V skills.json, profiles.json]. Leap has no row.

### 2.11 Hold, channel, charge-up, shapeshift

- `hold` reaches the game only as a left-button held click through `HoldClick` (`fighter.go:785-791`). A swinger backend gets one `Swing` (`fighter.go:783-784`); a right-hand skill gets one `CastAt` (`fighter.go:792-796`). So `hold` on Strafe, Fend, Arctic Blast and Whirlwind (right-hand rows) does nothing, and on d2d Zeal's chain depends on what one `cmd skill` does [U].
- The held click blocks the fight loop for the hold time (`fighter.go:808-813`, 600 ms default): no survival check during it.
- Charge-up: `KindCharged` exists (`skills.go:35`, `classdata.go:164`), no row uses it, no code reads it. The martial-arts rows are plain melee.
- Shapeshift: no form skill has a row; Maul, Fury, Rabies, Fire Claws are plain melee rows; nothing reads the form state [V skills.json].

## 3. Inventory of class and skill literals in engine code

Non-test Go, found with `grep -n 'd2\.(Sorceress|Paladin|...)'` and a skill-name grep over `internal/combat`, `internal/bot`, `internal/nav`, `internal/survival` [V]. This is the list the design must drive to zero; it is not added to the audit file (AGENTS.md).

| kind of literal | places |
|---|---|
| class switch for behaviour | `combat/necro.go:157`, `combat/druid.go:94-95`, `:230-256`, `combat/gearcheck.go:95`, `combat/rules.go:228-253`, `bot/plan.go:96`, `bot/geargate.go:50`, `bot/task.go:357-359`, `bot/routeguard.go:24`, `:108`, `:324-326` |
| Sorceress as the silent default | `combat/classdata.go:373`, `combat/skills.go:127-135`, `:17-22` (build constants and a respec level) |
| skill names in control flow | `combat/fight.go:816`, `:828`; `combat/affix.go:142`; `combat/castfirst.go:104`; `combat/skills.go:112`; `combat/necro.go:47`, `:216`, `:232`, `:310`, `:363`; `combat/necro_corpse.go:76`; `combat/assassin.go:70`, `:98`; `combat/trap.go:32`, `:196`; `combat/druid.go:107-109`, `:177-190`; `combat/telekey.go:18` |
| skill names as data in Go | `combat/castgap.go:17-24`; `combat/necro.go:39-45`; `combat/rules.go:229-253` |
| monster class ids and state ids | `combat/necro.go:31-35`, `:59-69`; `combat/druid.go:16-24` |
| knobs named after a skill | `env/tuning_combat.go:53-54`, `:58`, `:86-90` |
| per-class number tables | `combat/potioneffect.go:23-25` (a game fact per class: belongs in a class traits table, not this design) |

## 4. What the planner and the capability table do today

- The capability table has rows for four classes; Druid rows are on a branch, Amazon and Assassin have none [V `data/capabilities/capabilities.json` at this commit]. A skill without a row is derived as experimental from `power` `skillKinds` (`internal/agency/caps.go:129-147`).
- Roles and deliveries are closed sets checked at load (`agency/caps.go:101-103`): roles damage, aura, buff, summon, curse, movement, passive, utility; deliveries melee, projectile, nova, ground, trap, pet, self.
- The planner derives main, secondary, aura, buffs, summons and movement and keeps the combat profile's pick as the prior (`agency/derive.go:109-131`, `:175-182`). It runs in shadow and writes one log line (`agency/planner.go:66-89`). Nothing in `internal/combat` imports `internal/agency` or reads the table [V grep].
- The survival model resolves the hero's archetype from the combat profile bands (`internal/power/archetype.go:87-101`) and a `Contact` flag that mirrors the profile's melee block (`archetype.go:20`).

## 5. Dead ends of this review

- I looked for an existing "ground-target" branch on main to cite; there is none. The shared section is in `docs/process/druid-support.md` on branch `agent/druid1` (not merged when this was written); I read it there and cite it as S.
- I did not find a test that fails when the three skill tables disagree; `TestPlannerAgreementFixture` compares picks, not classifications [S docs/agency.md section 4].
