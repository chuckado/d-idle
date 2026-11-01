# Paladin probe (second-class probe, offline part)

> status: current. Card `palprobe` (user go 2026-11-01). Offline only: nothing here ran on the game or on d2d. Builds on `docs/research/paladin-act1-recon.md` and `docs/research/build-flexibility-audit.md` (rows cited as "row n").

Tags: **V** verified by a test in this change (name given) or by reading the file (named); **I** inferred; **U** unverified or a model input. Test runs: `go test -p 1 -count=1 -v -run PalProbe ./internal/leveler ./internal/combat ./internal/resource` on the herdr box; the town and install-backed tests (`internal/town`, `TestPalProbeRulesMatchInstall`) skip there and ran on the worker host (`labrun orion -- go test -p 4 -count=1 -run PalProbe ./internal/town ./internal/leveler`).

## 1. Part A: the seams (Sorceress behaviour unchanged)

| audit row | change | evidence |
|---|---|---|
| 1, 2 | `combat.Build` is now a string (the profile id inside a class, "" = class default). `LookupProfile(class, id)` returns an error for an unknown class or id; `ProfileFor` is the panicking wrapper for tests. `SelectProfileChecked` is the strict selector; `SelectProfile` logs the error once and falls back to the class default, never silently to the Sorceress for a known class | `TestLookupProfileByID`, `TestLookupProfileUnknownIsAnError`, `TestSelectProfileCheckedErrorsAndFallbackLogs` (V) |
| 3 | `planForClass` reads the profile id from the leveler table's `combat` key (data); only the legacy `fallback` table keeps its name rule; the plan carries `Err` and `applyClass` logs it loudly instead of fighting on a guess | `TestPlanForClassProfileFromTable` (V) |
| 6 | `prefsByClass` has an explicit row for all seven classes; `LookupPrefs` errors for an unknown class; `PrefsFor` logs once and scores as damage-only, not as a caster | `TestLookupPrefsPerClass` (V) |
| 29 | `panel.classRow` returns an error for an unknown class, `NewLayout` passes it up (it already returned an error, so no caller changed) | `TestClassRowUnknownIsAnError` (V) |
| 5 | checked, nothing to do: `cmd/d2bot/telewire.go:16` sets `nv.TeleportKeyFn = cb.TeleportKey` in production (V file read) | |

Existing tests: all pass untouched except two edits, both forced by the change: `pure_test.go` format verb `%d` to `%v` (the build is a string), and `leveler/class_test.go` pinned the silent fallback ("unknown class must keep the old behaviour" for a Necromancer); it now asserts the Necromancer's explicit caster row. Compatibility choices to know: a single-profile class (Paladin, Amazon, ...) still accepts the ids "", "blizzard" and "lightning" because `planForClass` has always handed every class `BuildBlizzard` and `castgap_test` runs an Amazon with `BuildLightning`; any other id on such a class is an error. The Amazon, Barbarian and Assassin rows of `prefsByClass` are new judgement values (U): weapon classes score damage only (they previously got the Sorceress staff/wand/orb bonus by accident).

## 2. Part B: probe table

| # | probe | verdict | evidence | tag |
|---|---|---|---|---|
| a1 | both plan tables run clvl 2 to 30 through the planner with the game's skill rules | WORKING | `TestPalProbeTablesSimulate`: no skill before its prerequisites or required level (legacy 29 points, v2 30 points, 0 violations); fixture equals the install's skills.txt (`TestPalProbeRulesMatchInstall`, worker host) | V |
| a2 | Zeal reaches its cap | WORKING for v2, BROKEN-ish for legacy | v2: Zeal 4 by clvl 15 (first point 12) `TestPalProbeZealCap`; legacy: Zeal 1 until clvl 30 (F4) | V |
| a3 | Sacrifice before Zeal, Zeal not before clvl 12 | WORKING | `TestPalProbeSacrificePrecedesZeal` | V |
| a4 | stat points follow the plan | WORKING | `TestPalProbeStatPointsFollowPlan`: every point spent, Str +40 Dex +23 Vit +82 on both tables (Str 65, Dex 43, Vit 107 at clvl 30) | V |
| a5 | legacy table spends the Den point | BROKEN (minor) | legacy has 29 skill rows, one point idle at clvl 30 with the Den point (F6) | V |
| b1 | ladder along both tables, clvl 1 to 30 | WORKING | `TestPalProbeLadderAlongBothTables`: Smite/Attack to clvl 11, Zeal from the first Zeal point at 12 or 13, Blessed Hammer from 24, always ends in the free Attack, never Sacrifice | V |
| b2 | Blessed Hammer band start | BROKEN (design) | tables train it at 20 to 22, band starts at 24 (F2, audit row 10) | V |
| b3 | held button for Zeal only | WORKING offline | `TestPalProbeHoldClickIsZealOnly` (Smite/Attack clicks, 0 holds), `TestPaladinFightsWithLeftHand` | V; on d2d UNVERIFIED (PALHOLD) |
| b4 | ladder flips when mana drains | WORKING | `TestPalProbeLadderFlipsWhenManaDrains` (Zeal held, ends on Attack, 6 kills) | V |
| b5 | Zeal without a held button (the d2d input) | WORKING offline | `TestPalProbeZealWithoutHoldClickStillFights`: one click per tick, fight ends; kill speed on d2d is the open question | V / U |
| b6 | aura choice and KeepAura | WORKING | `TestPalProbeAuraSelection` (Might, Holy Fire, fire-immune switches to Might, no aura known gives none), existing `TestKeepAura*`, `TestWantAuraPrayerWhileResting` | V |
| b7 | auras do anything on d2d | UNVERIFIED | the bot presses the key; the effect is d2d's (PALAURA) | U |
| b8 | melee back-off | WORKING | `TestPalProbeMeleeGuardThresholds`: surrounded and healthy stays, surrounded below 65 percent or one attacker at 40 percent walks away, 3 around and hurt stays | V |
| c1 | WantsShield, no two-hander over a shield | WORKING | `TestPalProbeWantsShield`, `TestPalProbeTwoHanderNeverOverShield`, `TestPalProbeRebuyNeverBuysTwoHander` (worker host, sim stock) | V |
| c2 | weapon choice below clvl 40 | WORKING | `TestPalProbeWeaponChoice`: sword over wand, one-hander over a two-hander on score | V (weights U, judgement) |
| c3 | repair triggers for a melee weapon | WORKING | `TestPalProbeRepairTriggersForMeleeWeapon` (28 percent weapon, repaired); the policy is class-agnostic | V on the sim; durability loss per hit UNVERIFIED (PALDUR) |
| d | merc auras for a non-caster | WORKING | `TestPalProbeMercAuras`: Normal Defiance then Prayer, Nightmare Might first; Might and Blessed Aim refused in Normal | V |
| e1 | belt plan | WORKING | `TestPalProbeBeltPlan`: three healing columns and one mana column at clvl 1 to 30 (Sorceress has two mana columns) | V (split U) |
| e2 | Hammerdin belt in Nightmare and Hell | BROKEN (late) | `paladin_hammerdin_endgame90` is not a mana-heavy build (F5); skipped test | V |
| f1 | melee model, Paladin vs Sorceress staff rung | WORKING (model) | `TestPalProbeMeleeModelVersusStaffRung`: clvl 8 sword 5.94 dmg/swing vs staff 3.13; Paladin clvl 8 melee-only 11 kills in 300 s vs staff-only Sorceress 15 (the Paladin run is held back by the caster retreat rule, below) | I, inputs U |
| f2 | early-game survival clvl 1 to 8 | WORKING in the model, thin at clvl 1 and 2 | `TestPalProbeEarlyLifeCurve`: closed-form life lost per pack against the 40 percent retreat line: pack of 1 costs 9 percent at clvl 1; a pack of 4 costs 157 percent at clvl 1, 80 at clvl 2, 54 at clvl 3, 42 at clvl 4, 26 at clvl 8. The replay had 0 deaths at clvl 1, 4, 8 | inputs U (foe level 3 gives 42 percent to-hit at clvl 1; level 1 would give about 58) |
| g | Zealot-lite table as data | NOT DONE, ready card | `leveler.buildFiles` is a Go map (S4 seam absent); no data file added, no glue written | V file read |

Counts: WORKING 20, BROKEN 5 (a2 legacy, a5, b2, e2, plus F3 below), UNVERIFIED 2 probes (b7 and the d2d side of b3, b5) plus the live-only list in section 5.

Model note for f1: the resource policy's melee fallback only stands against packs of 3 or fewer, a rung written for a Sorceress out of mana (the replay shows 14 to 25 retreats for the Paladin). The production Paladin does not use it: the melee ladder always ends in the free Attack, so `AttackChoice` is never asked (V `TestPaladinFallsBackToAttackWithoutMana`; I for the code path in `combat/resource.go:289`).

## 3. Findings (ready fix cards)

| id | finding | severity | ready card line |
|---|---|---|---|
| F1 | none beyond the seams: the Paladin path found no crash and no wrong choice in the offline ladder | | |
| F2 | profile band for Blessed Hammer starts at clvl 24, tables train it at 20 (legacy) or 22 (v2) | low (BH is not an Act 1 skill) | PALBAND: derive the band minimum from the table's `main_attack` (audit S1 phase 1) or lower `minLevel` to 18 with `requires` doing the gating; test `TestPalProbeHammerBandStartsAtTraining` flips to passing |
| F3 | `paladin-normal30` never trains Prayer, the profile's resting aura rung is dead on it | low | PALTABLE: add Prayer 1 to the legacy table or retire it for fresh and owned-Might characters; test `TestPalProbeLegacyTableHasNoPrayer` |
| F4 | `paladin-normal30` trains Zeal to 1 only | medium (Zeal at level 1 hits few targets; every legacy Paladin swings it from clvl 12) | PALTABLE (same card): Zeal to 4 by clvl 15 as v2 does; test `TestPalProbeLegacyZealStaysAtOne` |
| F5 | `paladin_hammerdin_endgame90` missing from `manaHeavyBuilds` | low until Nightmare | PALMANA: add it to the Go set and `potions-belt.json`, or do S4 `mana_heavy` metadata; test `TestPalProbeHammerdinManaHeavy` |
| F6 | legacy table has 29 skill rows, the Den point idles | trivial | covered by PALTABLE |
| F7 | resource simulator `world.input` hard-codes a Fire Bolt slot and a Staff melee skill, so a class scenario is a disguise (audit row 33) | model quality | SIMCLASS: let `Scenario` carry its skill list; then the Paladin Zeal rung can be a real skill |

Ready card, not started: **PALZEAL (the Zealot-lite table)**: add `internal/plandata/build-paladin-zealot-lite.json` from recon section 3.2 (Prayer 3, Smite at clvl 4, Zeal 4 by 15, Vigor line), register it. It needs one line in `buildFiles` (Go) today; with the S4 manifest (audit phase 2) it needs none. Acceptance: `TestPalProbeTablesSimulate` and `TestPalProbeZealCap` run on it with 0 illegal points; Smite and Prayer prerequisites hold; the SME confirms Smite at clvl 4 (recon open point).

## 4. What in this probe was Sorceress-fitted

Bespoke Sorceress fit that the Paladin ran into (fixed by part A): the two-value build enum and the silent Sorceress fallback in `ProfileFor` (rows 1, 2); the table-name rule in `planForClass` (row 3); gear preferences that gave every other class the staff bonus (row 6); the inventory row fallback (row 29). General, not Sorceress-fitted (worked first time): the profile data and melee loop, the plan tables and planner, the left-hand binder, the belt plan, the merc policy, repair, WantsShield. Still Sorceress-fitted and found on the way, not touched: the resource simulator's Fire Bolt and Staff slots (F7, row 33), the melee-fallback pack limit of the resource policy (a caster rung), the profile band minimum (F2), the mana-heavy build list (F5). The Zealot-lite table cannot be a data-only change yet because `buildFiles` is a Go map.

## 5. Live probes still owed (docs/BACKLOG.md)

| card | what it answers |
|---|---|
| PALHOLD | does d2d hold or repeat a left attack; does Zeal, Smite or Sacrifice fall back to a plain attack ("not implemented: skill"); needed before any Paladin number on d2d means anything |
| PALHAND | do `SetLeft`, `KeepAura`, `ActiveAura` and the left-skill read work through the d2d source and backend |
| PALAURA | do auras apply in d2d (stat 350, hp loss on a fixed target) and are merc auras built |
| PALDUR | durability loss per hit and per block in the original vs d2d, then a 200-swing drill and a smith repair |

Not tested offline and still open: the fidelity of the player's own blow on d2d (SKMELEE), melee animation and IAS (recon section 6).

## 6. Shared-file lines for the merger

Not edited here.
- **docs/BACKLOG.md**: `PALBAND/PALTABLE/PALMANA/SIMCLASS/PALZEAL (from palprobe): ready fix cards in docs/classes/paladin-probe.md section 3`.
- **docs/INDEX.md**: `- docs/classes/paladin-probe.md: second-class probe, Paladin offline results (see the table), findings F1 to F7, Sorceress-fitted pieces, live probes owed (PALPROBE)`.
- **docs/CHANGELOG.md**: `palprobe: profile id string (LookupProfile errors for an unknown class or id), explicit per-class gear preferences, classRow error for an unknown class; Paladin probe tests and docs/classes/paladin-probe.md`.
