# Sorceress skill matrix: inventory, coverage, gap decomps, live skill-smoke drill, per-act gate

> status: research, 2026-11-01, agent SKILLINV (card skillinv; user go 2026-11-01 "verify the sorc skills"). Steps 1 and 2 are DONE here (inventory and coverage matrix); steps 3 and 4 are PROPOSALS only (nothing run, no game, no d2d slot, no product code, nothing posted or pushed). Tags: **V** verified (an oracle, a disassembly read in an earlier trace of ours, or a live drill we ran; the cite says which), **S** read in source or docs (our code, d2d's code, a doc), **I** inference (computed or concluded by this agent from V/S inputs, not run through an oracle), **U** unknown. A cell may carry several tags, one per statement. Skills.txt values are from the PATCHED 1.14d tables extracted for the chill trace on the worker host (D = data read, counted as S unless a trace verified the use). d2d file:line cites are the lab fork checkout on the worker host at the time of this read (upstream plus our local patch set); lines drift.

> status: cells (d) live and (b) d2d status of sections 3.3, 3.4, 3.7, 3.8, 3.9, 3.10, 3.12 and 3.13 (Cold Mastery) were updated 2026-11-01 by card capdocs from the later traces (sk-armor, sk-nova, sk-teleport, sk-lightning, sk-passives, sk-resist, sk-melee, sk-melee2) and the live drills DRILLS24 to 29. Each cell says which; the older text beside it is the state when the matrix was written. The counts in section 4 were NOT re-run.

## Answer

- **Skills the plan uses on the default build, cumulative by act: A1 5, A2 6, A3 8, A4-A5 and NM 9** (Fire Bolt, Charged Bolt, Frozen Armor, Frost Nova, Telekinesis as a prerequisite point only; Act 2 adds Teleport; clvl 26+ adds Lightning Mastery and Chain Lightning; clvl 40 adds Lightning); plus Static Field on characters that still carry the legacy table and the plain staff swing, which make it 7 / 8 / 10 / 11. The Blizzard endgame table (opt-in, never a default) adds Ice Bolt, Ice Blast, Glacial Spike, Blizzard, Cold Mastery, Warmth. Per-act counts are in section 1.
- **Coverage is thin.** Of the skills the plan casts today only the damage formula (V, `skill-damage-trace.md`), cast ticks (V, `cast-trace-gameexe.md`, PR #128 merged upstream), Static Field and chill (V, `static-field-chill-trace.md`) and bind (V, `skill-bind-trace.md`) were traced against game.exe. Mana cost is formula-only (S), the missile pipeline (speed, range, Charged Bolt bolts, Chain Lightning hops, hit tests), the armors (Frozen Armor, Energy Shield), Teleport, Frost Nova's ring, ground/area skills and the passives (masteries, Warmth) are UNTRACED. Live proof exists for exactly one skill with an oracle (Fire Bolt mana and hp, drill D5.a).
- **Two findings before any trace (both bot-side, both S):** (F1, resolved 2026-11-01 by fixstatic: dropped, see docs/fixstatic.md) a FRESH Sorceress on the default build never learns Static Field (BUILDV2 moved it to clvl 86), but the Andariel card and `staticOpener` assume it; the bot's Static Field code is dead on v2 characters until the table or the card changes. (F2) `BuffRecastSec` 150 s is longer than Frozen Armor's duration at levels 1-2 (3,000 + 300 x (lvl-1) frames = 120 s, 132 s), so the armor is down for about 20 % of every cycle (I, from Skills.txt `auralencalc` and the knob).
- Ranked gap decomps (section 5): missile pipeline, armors and states, Teleport, Lightning / Chain Lightning, Frost Nova ring, mana and cast pipeline, then the staff swing, resist / immunity, ground and area skills, passives.

## 0. Sources read

Plan and class model: `internal/plandata/build-fallback-lightning-v2.json` (BUILDV2, fresh characters), `build-fallback-lightning.json` (legacy), `build-primary-blizzard.json`, `build-sorceress_blizzard_endgame90.json`; `internal/leveler/build.go` (table selection: `freshBuild`, `endgameBuild`), `cmd/d2bot/class.go` (default combat build for a Sorceress = lightning, leveler `fallback`); `internal/combat/classdata/profiles.json` and `skills.json` (attack order, buffs, F-key layout, ranges), `internal/combat/skills.go`, `fight.go`, `castfirst.go`, `fighter.go` (Buff), `internal/nav/teleport.go`, `internal/env/tuning_combat.go`; `docs/plan/plan-v2.md`, `docs/plan/STATUS.md`, `docs/earlygame.md`, `docs/resource.md`, `docs/act2.md`; traces `skill-damage-trace.md`, `skill-bind-trace.md`, `static-field-chill-trace.md`, `freeze-rules-trace.md`, `cast-trace-gameexe.md`, `walk-speed-trace.md`, `monster-attack-trace.md`, `monster-mods-trace.md`, `pets-merc-trace.md`; drills in `docs/primitives.md`, `docs/forkcheats.md`, `docs/fixmeter.md`, `docs/probes/drills/`. d2d: the lab fork checkout (`components/game/fight.cpp`, `fight.hpp`, `components/rules/skills.hpp`), read for the dispatch on Skills.txt `srvstfunc` / `srvdofunc`.

How d2d implements skills (S, one paragraph, it explains most "partial" verdicts below): d2d is table-driven. A skill is a row of Skills.txt; its behaviour comes from the start function (`srvstfunc`), the do function (`srvdofunc`) and the missile columns. Melee-like skills run through `skill_built` (`fight.hpp:62`, everything else "swings a plain attack for now" with a `not implemented: skill` log line, `fight.cpp:669`). Missile and spot skills run through `cast_missile` (`fight.cpp:~2570`), `skill_missile` (`~2554`, a fixed list of do functions and hit functions), `fire` (`~2605..2740`, one branch per do function: scattered bolts 17, nova 22, chain lightning 26, missile at target 28, missile wall 24, breath 19), `spot` (`1112..1215`: Static Field 20, Teleport 27, Telekinesis 21) and self casts through `self_cast` (`fight.hpp:83`, armors 18, Energy Shield / Blaze 23, Enchant 25, Thunder Storm 29). So "implemented in d2d" below means "has its own branch with source comments citing the game.exe function"; it never means "differential-proven" unless a trace says so.

## 1. STEP 1: inventory of every Sorceress skill and what the plan uses

Skills.txt rows with `charclass` sor (30 rows, ids 36-65). `srv st/do` = server start / do function; mana = `mana, lvlmana, minmana, manashift`; cost at level L = `max(minmana, (mana + lvlmana x (L-1)) x 2^manashift / 256)` (S, FUN_0056c160 not re-read). reqlvl = Skills.txt `reqlevel`.

Plan columns. Which table: **FRESH** = `fallback-lightning-v2` (default for a fresh Sorceress, combat profile `sorceress/lightning`); **LEG** = `fallback-lightning` (existing characters; Static Field at clvl 6, Telekinesis at 10, Frozen Armor at 22); **END** = `sorceress_blizzard_endgame90` (opt-in `--build endgame`, combat profile `sorceress/blizzard`, never a default). Act bands are the plan's own where it states them (Andariel gate clvl 15 by knob `boss.andarielMinLevel`, Act 2 entered about clvl 15-18, 24 h run: "Act 2 entered and a Normal Act 2 chain", Acts 3-5 / NM / Hell "not built until a run gets there", plan-v2 sections on stages) and my reading elsewhere [I]: A1 clvl 1-16, A2 16-26, A3 26-32, A4 32-36, A5 36-40, NM from 40, Hell from 60. Levels are points from the table cumulated at the band's end (+ item levels, not planned).

| Id | Skill | reqlvl | prereq | srv st / do | anim / seqtrans | mana at L=1 (+per lvl) | Plan use (FRESH table) | Levels reached (FRESH) | END / LEG note |
|---|---|---|---|---|---|---|---|---|---|
| 36 | Fire Bolt | 1 | none | - / 0 (plain missile `firebolt`) | SC / SC | 2.5 (0) | **A1 yes** (start skill, fallback attack, F7) | 1 (never raised) | END: unused |
| 37 | Warmth | 1 | none | passive `warmth` | SC / SC | 0 | not used | 0 | END 1 pt at clvl 15 |
| 38 | Charged Bolt | 1 | none | - / 17 | SC / SC | 3.0 (+0.5) | **A1-A3 yes: THE main attack to clvl 25**, then third in the order | A1: 1-13, A2: 13-19, to 20 by clvl 25 | LEG same; END unused |
| 39 | Ice Bolt | 1 | none | - / 0 (`icebolt`) | SC / SC | 3.0 (0) | in the attack list (third) but never learned on FRESH | 0 | END: main attack clvl 1-7, fallback after |
| 40 | Frozen Armor | 1 | none | - / 18 | SC / SC | 7.0 (0), cast in town allowed | **A1 yes** (Den point at clvl 5, F3, buffed on a timer) | A1-A2: 1; 2 at clvl 22 | END 1 pt |
| 41 | Inferno | 6 | none | 11 / 19 | SQ / SQ | 36 (+1) shift 2 | not used | 0 | |
| 42 | Static Field | 6 | none | - / 20 | SC / SC | 9.0 (0) | **FRESH: not learned until clvl 86** (F1); LEG: A1 yes (elite / boss opener, 4 casts) | FRESH 0; LEG 1 | END 1 pt (cold-immune fallback) |
| 43 | Telekinesis | 6 | none | 12 / 21 | SC / SC | 7.0 (0), in town ok | **prerequisite point only** (A1 clvl 11); never cast (no F-key) | 1 | |
| 44 | Frost Nova | 6 | none | - / 22 (`frostnova`) | SC / SC | 9.0 (+1) | **A1 yes** (crowd of >= 3 within 4 subtiles, mana >= 25, F8) | 1 | END 1 pt |
| 45 | Ice Blast | 6 | Ice Bolt | - / 0 (`iceblast`) | SC / SC | 6.0 (+0.5) | not used | 0 | END: main clvl 8-17 |
| 46 | Blaze | 12 | Inferno | - / 23 | SC / SC | 22 (+1) | not used | 0 | |
| 47 | Fire Ball | 12 | Fire Bolt | - / 0 (`fireball`) | SC / SC | 10 (+1) | not used | 0 | |
| 48 | Nova | 12 | Static Field | - / 22 | SC / SC | 15 (+1) | not used | 0 | |
| 49 | Lightning | 12 | Charged Bolt | - / 0 (`lightningbolt`) | **SQ** / SC | 8.0 (+0.5) | **A5+ in the lightning band** (clvl 40: 1 pt, 20 by 90) | 0 until clvl 40 | |
| 50 | Shiver Armor | 12 | Ice Blast, Frozen Armor | - / 18 | SC / SC | 11 (0) | in the armor group (second choice), never learned | 0 | |
| 51 | Fire Wall | 18 | Blaze | - / 24 | SC / SC | 22 (+1), delay 35 | not used | 0 | |
| 52 | Enchant | 18 | Warmth, Fire Ball | - / 25 | SC / SC | 25 (+1) | not used | 0 | |
| 53 | Chain Lightning | 18 | Lightning | - / 26 | **SQ** / SC | 9.0 (+1) | **A3+ (first point clvl 27; 5 by clvl 35, 9 by 40, 20 by 75)**, first in the clvl >= 26 order | 1-9 until clvl 40 | |
| 54 | Teleport | 18 | Telekinesis | - / 27 | SC / SC | 24 (-1) | **A2+ yes** (clvl 18, nav hops, F2) | 1 | END 1 pt at clvl 18-20 |
| 55 | Glacial Spike | 18 | Ice Blast | - / 0 (`glacialspike`) | SC / SC | 10 (+0.5) | not used | 0 | END: main clvl 18-23 |
| 56 | Meteor | 24 | Fire Ball, Fire Wall | - / 28 | SC / SC | 34 (+1), delay 30 | not used | 0 | |
| 57 | Thunder Storm | 24 | Nova, Chain Lightning | 13 / 29 | SC / SC | 19 (0) | not used | 0 | |
| 58 | Energy Shield | 24 | Teleport, Chain Lightning | - / 23 | SC / SC | 5 (0), in town ok | in the lightning buff order (second group) **but the table never buys it** | 0 | |
| 59 | Blizzard | 24 | Frost Nova, Glacial Spike | - / 28 | SC / SC | 23 (+1), **delay 45** | not used on FRESH | 0 | END: main from clvl 24 (20 by 45) |
| 60 | Chilling Armor | 24 | Shiver Armor | - / 18 | SC / SC | 17 (0) | armor group (third), never learned | 0 | |
| 61 | Fire Mastery | 30 | none | passive | SC / SC | 0 | not used | 0 | |
| 62 | Hydra | 30 | Enchant | 14 / 144 | SC / SC | 40 (+1), delay 40 | not used | 0 | |
| 63 | Lightning Mastery | 30 | none | passive `lightningmastery` | SC / SC | 0 | **A3+** (table rows from clvl 26; the planner's rule skips points while clvl < reqlvl 30, see note) | 4 at clvl 30, 5 at 35, 10 at 50, 20 at 60 | |
| 64 | Frozen Orb | 30 | Blizzard | - / 0 (`frozenorb`) | SC / SC | 50 (+1), delay 25 | not used | 0 | |
| 65 | Cold Mastery | 30 | none | passive `coldmastery` | SC / SC | 0 | not used on FRESH | 0 | END: from clvl 50 (5), 20 by 75 |

Note on Lightning Mastery: the FRESH rows put the first Lightning Mastery point at clvl 26 while Skills.txt `reqlevel` is 30 (D). The leveler's `rule` (`leveler.go:272`) refuses a spend while `ReqLevel > level`, so those points wait and the effective first point is clvl 30 (S; the later spend of the saved points was not read, U). Not a verification target here.

Plan use summary (distinct skills the CODE can cast or learn on the default build; "casts" = has a fight / nav / buff call path):

| Act (band) | Skills with a call path (casts or learned as a prerequisite point) | Count | Skill levels in the band (FRESH table) |
|---|---|---|---|
| A1 Normal (clvl 1-16) | Fire Bolt, Charged Bolt, Frozen Armor, Frost Nova, Telekinesis (point only); Static Field on LEG characters; plain Attack (staff swing as the mana-out fallback) | 5 (7 with Static Field on LEG and the staff swing) | Fire Bolt 1, Charged Bolt 1-13, Frozen Armor 1, Frost Nova 1, Telekinesis 1 (from clvl 11) |
| A2 Normal (clvl 16-26) | + Teleport (nav hops) | 6 (8 with both) | Charged Bolt 13-20, Teleport 1 (clvl 18), Frozen Armor 2 (clvl 22) |
| A3 Normal (clvl 26-32) | + Lightning Mastery (passive), Chain Lightning, attack order switches to Chain Lightning, Lightning, Charged Bolt, Fire Bolt | 8 (10 with both) | Chain Lightning 1-5, Lightning Mastery 4-5 |
| A4 / A5 Normal (32-40) and NM from 40 | + Lightning (clvl 40) | 9 (11 with both) | Chain Lightning 5-9, Lightning 0-1, Lightning Mastery 5 |
| NM / Hell (40+, "not built") | the FRESH table's later rows only deepen the same skills (Lightning 6-20 by 90, Lightning Mastery 10-20, Chain Lightning to 20, Static Field 5 only at the very end) | 9 (11 with both) | as the table |
| END table (opt-in; replaces the above) | Ice Bolt (clvl 2), Frozen Armor, Telekinesis, Static Field, Frost Nova, Warmth (clvl 10-15), Ice Blast (8+), Teleport (18), Glacial Spike (18+), Blizzard (24+), Cold Mastery (30+, 5 by 50) | A1 4 (Ice Bolt, Frozen Armor, Static Field, Frost Nova), A2 8, A3+ 11 | Ice Bolt 6 then 15 at clvl 90, Ice Blast 12 then 20, Glacial Spike 6 then 20, Blizzard 20 by 45, Cold Mastery 5-20 |

Counts are "distinct plan skills with a path in that band", cumulative. The FRESH Act 1 set the matrix must cover is therefore Fire Bolt, Charged Bolt, Frozen Armor, Frost Nova (+ Telekinesis as a point, + Static Field if the plan keeps it, + Attack).

The skills no band uses (inventory only, no matrix row): Warmth (END only), Inferno, Blaze, Fire Ball, Nova, Shiver Armor, Fire Wall, Enchant, Meteor, Thunder Storm, Chilling Armor, Fire Mastery, Hydra, Frozen Orb. If the plan or the user adds one, the gap decomp for its family (section 5) applies first.

## 2. Common pipeline facts every skill row refers to

| Id | Fact | Source and tag |
|---|---|---|
| K1 | Cast time of a plain `SC` cast: `ceil(256 f / s) - 1` ticks; Sorceress 13 / 12 / 11 / 10 / 9 / 8 / 7 at FCR 0 / 9 / 20 / 37 / 63 / 105 / 200; rate capped at 175 (`100 + 120 FCR/(120+FCR)`); a `SQ` skill whose seqtrans is SC (Lightning, Chain Lightning) takes 19 steps at 0 FCR | V `cast-trace-gameexe.md` "Claims" (oracle `cast_rates.py`, 2,807 + 1,203 lines, 0 differ) |
| K2 | d2d measured 14 ticks for every spell before PR #128 (+8 %; Lightning / Chain Lightning 14 instead of 19 = -26 %) | V `d2d-fidelity.md` 7a (8 casts x 8 spells, lab); fixed by upstream PR #128 (merged, `research/README.md`); whether the lab binary carries it: U |
| K3 | Damage formula: brackets from `EMinLev1..5`, `(EMin + brackets) << HitShift` in 256ths, synergy `MulDiv(min, calc, 100)` with the min gate, mastery once at creation, element byte table | V `skill-damage-trace.md` (116,000 cases, d2d differs only in the min gate for "1-x" lightning skills and MulDiv above 0x100000, fixed in a candidate PR) |
| K4 | Missile damage applied to a monster: resistance, pierce, 256ths of life, death (PR #135 family); d2d truncated spell damage to whole points (-17 % at skill 1) | V `pr-missile-damage-fine-proof.md`; PR branch built, not pushed (STATUS) |
| K5 | Mana cost formula `(mana + lvlmana x (L-1)) << manashift / 256`, floor `minmana` | S Skills.txt D + d2d `skills.hpp:516` `mana_cost`; FUN_0056c160 not re-read: U; Fire Bolt value V by drill D5.a |
| K6 | Bind and cast packets: server stores a select only for a row with an entry of that id at level > 0; right-hand in-town refusal; cast packets carry no skill id | V `skill-bind-trace.md`; d2d D1 (stores any id) 81,756 of 127,032 differ, 0 under the rule; PROOF incomplete, no PR |
| K7 | Skill `delay` (status 8) is not modelled by d2d (Blizzard 45, Meteor 30, Hydra 40, Fire Wall 35, Frozen Orb 25 frames) | S `skill-bind-trace.md` D4 (d2d gates 6 / 8 / restrict / minmana not modelled); no early skill is affected |
| K8 | Cast gap in the bot is 650 ms (above the 13-tick cast, test `TestCastGapCoversEarlyCast`); mana is the bot's own read | S `tuning_combat.go:77`, `tuning_combat_test.go` |

## 3. STEP 2: coverage matrix (used skills and the next band)

Cell format: statement [tag]. "Live" is the strongest drill or run, with the doc. Expected values are computed from Skills.txt and the traced brackets (tag I: computed here, not run through an oracle; no synergies, no masteries, no item levels). Damage is in whole points per hit (256ths / 256).

### 3.1 Fire Bolt (36) used A1 (start skill, fallback attack)

| Aspect | Entry |
|---|---|
| (a) damage | `EMin 6 EMax 12 HitShift 7`, brackets Lev1 3/3 to 8, Lev2 4/6 to 16: L1 3-6, L4 7.5-10.5, L8 13.5-16.5, L12 21.5-28.5, L20 45.5-60.5 [I]; synergy Fire Ball + Meteor x par8, Fire Mastery [S]; formula equal in d2d [V K3]; whole-point truncation in d2d [V K4] |
| (a) mana | 2.5 at every level (lvlmana 0) [I]; 35 -> 32 on a cast [V D5.a, display floors 2.5] |
| (a) cast time | SC, 13 ticks at 0 FCR [V K1] |
| (a) behaviour, range, targeting | plain missile `firebolt`, one per cast, flies to the click point; range in the bot 20 (Kolbot, UNVERIFIED in 1.14d) [S]; player missile to-hit / `UseAttackRate` not looked at [U] |
| (a) resist, immunity | fire resistance and pierce through the #135 path [V K4]; fire immune monsters take 0 [S]; base = MonStats ResFi of the difficulty, no penalty / cap for monsters, immune at total >= 100, pierce only under 100, floor -100, curses on an immune / 5 (V read + V run, sk-resist-trace.md; d2d D1) |
| (b) d2d status | implemented: generic plain-missile path, `srvmissile`, hit function table (`fight.cpp:~2554`) [S]; known deviations: cast ticks (K2, fixed upstream) [V], whole-point damage (K4) [V]; missile speed / range per level not traced [U] |
| (c) bot use | third in the attack order below clvl 26, first only when Charged Bolt is unaffordable (`PickAttack` falls to the cheaper spell); bound F7 (lightning layout); hotkey is the right hand; assumes 650 ms gap [S `skills.go`, `classdata/profiles.json`]; ranges UNVERIFIED [U] |
| (d) live | D5.a PASS (mana 35 -> 32, zombie hp 11 -> 8; `primitives.md` P26, BACKLOG DRILLS10) [V]; many Game.exe soak casts, no oracle [S] |

### 3.2 Charged Bolt (38) used A1-A3 (the main attack to clvl 25)

| Aspect | Entry |
|---|---|
| (a) damage | `EMin 4 EMax 8`, Lev1 1/1, Lev2 1/1: per bolt L1 2-4, L4 3.5-5.5, L8 5.5-7.5, L12 7.5-9.5, L20 13.5-15.5 [I]; synergy Lightning x par8 [S]; formula equal [V K3]; truncation [V K4] |
| (a) bolt count | `calc1 = min(24, ln12)` = 3 + (L-1): L1 3, L4 6, L8 10, L12 14, L20 22 [I from Skills.txt]; tooltip differential 22,800 lines equal [V `upstream-gaps/item1-charged-bolt-hit.md`, via `skill-damage-trace.md`] |
| (a) mana | 3.0 + 0.5 (L-1): L1 3.0, L4 4.5, L8 6.5, L12 8.5, L20 12.5 [I; GAPCASTS read of the csv vs the formula S] |
| (a) cast time | SC, 13 ticks [V K1] (the maintainer's "S1" note was wrong, Skills.txt anim SC [V `mpq-recheck-pr-claims.md`]) |
| (a) behaviour | `srvdofunc 17` (FUN_005c9300): the bolts leave one cast and wander (FUN_005c9290) in a fan; bolt flight path, speed, spread angle, lifetime and who each bolt can hit (one monster takes many bolts?) not traced [U]; range in the bot 10 (Kolbot) [S, UNVERIFIED] |
| (a) resist, immunity | lightning resist and pierce [V K4]; lightning immune takes 0 [S]; as fire with ResLi; Lower Resist 1 leaves a lightning-110 monster at 104 (immune) in game.exe, 79 on d2d (D1, V run 147,628 cases, sk-resist-trace.md) |
| (b) d2d status | implemented (`kScatteredBolts`, `fight.cpp:2631`: `send(...).bolt = k`, "each wandering (FUN_005c9290)") [S]; the wander path is d2d's reading of FUN_005c9290, never differentialled [U]; casts per kill measured on d2d only [V-d2d `d2d-casts-real.md`], never on game.exe [U]; the earlier "gap in casts per kill" was a probe artifact [V `gap-casts-per-kill.md`] |
| (c) bot use | attack order band clvl < 26: Charged Bolt, Fire Bolt, Ice Bolt; F4; `castGapMs` 650; per-kill casts and the low-mana ladder (Charged Bolt too dear -> Fire Bolt -> staff) [S `resource.md` low-mana ladder]; assumes bolts hit what is in the aimed direction [I] |
| (d) live | `drills10` log: Charged Bolt decisions on Zombies, fight counters kills 13 / dealt 121 / taken 106 match the independent sampler (fixmeter) [V `fixmeter.md`, BACKLOG DRILLS10] but no per-cast oracle; D5.p28 20 zombies, 12 kills [V `primitives.md`]; per-cast mana / hp / missile-count drill: none [U] |

### 3.3 Frozen Armor (40) used A1+ (Den point at clvl 5, rebuffed on a timer)

| Aspect | Entry |
|---|---|
| (a) rules from Skills.txt | `srvdofunc 18` (FUN_005c9480); armor +`ln12` % = 30 + 5 (L-1) (`skill_armor_percent`); duration `auralencalc` = 3000 + 300 (L-1) frames (+250 per Shiver / Chilling Armor point) = 120 s at L1, 132 s at L2; striker freeze `calc1` = 30 + 3 (L-1) frames (x synergy) [I from D] |
| (a) mana | 7.0 flat, castable in town (InTown 1) [I; in-town cast V by the bot's town buff in Game.exe live runs S] |
| (a) cast time | SC 13 ticks [V K1] |
| (a) hit effect | event 2 `damagedinmelee` (FUN_005cac50): d2d cites it; whether the striker is frozen or only chilled, the boss / champion / hireling exceptions of FUN_0057b230 (`freeze-rules-trace.md` C3 / C5 apply to FREEZE sources; Frozen Armor's route was listed as "next candidate" there) and cold immunity / cold resist on it: UNTRACED [U] |
| (b) d2d status | implemented: armor %, state timer, and in `buff_events` (`fight.cpp:~1374`) `stun_until` and `chill_until` for `calc1` frames on any melee striker that did damage [S]; this bypasses the freeze rules (plain monsters frozen only at coldeffect < 0, specials chilled) that C3 / C5 prove for Ice Bolt-style hits, so a deviation is likely [I]; no ponytail note on it |
| (c) bot use | `Fighter.Buff` casts the first known armor of the group at the F3 key when `buffDue`: not cast within `BuffRecastSec` 150 s (the state is not readable: UNVERIFIED offset) [S `fighter.go:582`, `tuning_combat.go:79`]; F2: the armor is down about 20 % of every cycle at L1-2 (120 / 132 s vs 150 s) [I]; also cast in town [S] |
| (d) live | Update 2026-11-01: recast timer live (DRILLS13, 16: casts at 114 s, armor never absent) and the cut-cast retry live (DRILLS29, D29-C: cut at 114.0 s, retried 1.6 s later, absent 0 ms over 544 samples) [V live]; a cut at normal monster damage [U]. Game rule (sk-armor-trace): the striker effect is the freeze rule, a recast restarts the full duration, one armor of the group at a time [V run]. d2d deviates (stuns and chills every striker, armors stack); the fix is in the private fork and proposed upstream as PR #159 (open) [S]. The bot-side recast gap inside one long clear (B1 of the trace) is covered by the fixrecast and armorcut cards, whose live proof is above [S] |

### 3.4 Frost Nova (44) used A1 (crowd rule)

| Aspect | Entry |
|---|---|
| (a) damage | `EMin 4 EMax 8`, L1 2-4, L4 8-11.5, L8 16-21.5, L12 28-35.5, L20 56-67.5 [I]; synergy via Ice Bolt line [S]; formula [V K3] |
| (a) mana | 9 + (L-1): L1 9, L4 12, L8 16, L12 20 [I] |
| (a) behaviour | `srvdofunc 22` (FUN_005c9b50): a nova of the `frostnova` missile, ring range = missile Range + calc1; each monster once; cold hit: chill or freeze through FUN_0057b230 [V rules `freeze-rules-trace.md` for the freeze / chill decision on a cold hit]; ring size and spacing in game.exe not traced [U] |
| (a) cast time | SC 13 ticks [V K1] |
| (b) d2d status | implemented: 64 missiles on a circle, a shared "struck" set so a monster is hit once, range `missile range + calc1` (`fight.cpp:2633`) [S]; chill / freeze application equals game.exe only on the local branch fix/freeze-rules (#155's head), not in the lab default [V `freeze-rules-trace.md`, PR candidate, not posted]; the 64-ray ring vs game.exe's ring is U |
| (c) bot use | `novaTick` (`castfirst.go:95`): at least 3 non-immune hostiles within 4 subtiles, mana >= `NovaManaFloor` 25, gap 2.5 s; F8; assumes the nova hits everything inside 4 subtiles and chills it (no chill model, finding B4 of `static-field-chill-trace.md`) [S] |
| (d) live | none [U]. Update 2026-11-01: game rule read and run (64 missiles, reach about 11 to 12 subtiles, a monster hit once) [V run, sk-nova-trace]; d2d hit 6.2 times per monster per cast on the same flights [V run, model I]; the hit-once fix is merged upstream as PR #158 [S]; the bot's 4 subtile crowd rule is still narrower than the real ring [S] |

### 3.5 Static Field (42) used on LEG characters (A1 elite / boss opener); dead on FRESH characters (F1)

| Aspect | Entry |
|---|---|
| (a) rules | life = stat 6 >> 8 whole points; none at or below `StaticFieldMin` of max life (Normal 0, NM 33, Hell 50 on the expansion); damage `MulDiv(life, 25, 100)`, at most life - 1; a negative lightning resist is cancelled; then 256ths with pierce; radius `ln12` = 5 + (L-1) subtiles; mana 9 flat [V `static-field-chill-trace.md` Q1, read + run]; cast SC 13 ticks [V K1] |
| (b) d2d status | implemented but wrong in three ways: no difficulty floor, no "at most life - 1", negative resistance amplifies; plus ceiling / whole-point damage instead of 256ths (`fight.cpp:1112`) [V 115,103 of 309,664 cases differ, 0 after local candidate fix/static-field, PR #155 open] ; the cold chill part C1-C5 [V, `freeze-rules-trace.md`] |
| (c) bot use | `staticOpener` (`fight.go:726`): elite or boss, `tr.statics < StaticCasts` 4, floor by difficulty 25 / 33 / 50 %, range `staticRange(level) = (level+4) x 2/3`, key F1 or right hand; the unit of that range vs `Target.Dist` is open (B1 of the trace) [S]; the Andariel card (`m3-andariel`) says "Static Field until the boss cap" but the FRESH table has Static Field at clvl 86: the call never fires on a fresh character [S, F1] |
| (d) live | none; bound and rotated in a cheated Game.exe session (`combat.md:97`, no damage oracle) [S] |

### 3.6 Telekinesis (43) used as a prerequisite point only (A1 clvl 11)

| Aspect | Entry |
|---|---|
| (a) rules | prerequisite of Teleport; mana 7, damage 1-2 + 1 per level lightning, knockback `srvdofunc 21` [I from D] |
| (b) d2d status | implemented (`fight.cpp:1188`: damage and a knockback) [S] |
| (c) bot use | learned (point), never cast (Layout comment: "Telekinesis is dropped", no F-key) [S] |
| (d) live | none; nothing to verify beyond the legal point spend [U] |

### 3.7 Teleport (54) used A2+ (nav hops, F2)

| Aspect | Entry |
|---|---|
| (a) mana | 24 - (L-1): L1 24, L4 21, L8 17, L12 13, L20 5 [I from D]; bot assumes 24 as an upper bound [S `teleport.go:20`] |
| (a) cast time | SC 13 ticks [V K1]; no skill delay (delay empty) [S] |
| (a) rules in game.exe | works only where Levels.txt `Teleport` allows (0 none, 2 clear line only, Duriel's lair 2) [V Levels.txt for the lair, `act2.md`]; refused in town (Game.exe live: hops "just print and the character stays", `act2.md` X3) [V]; landing spot nearest free (FUN_00554ea0) [S d2d comment]; range limit, click-to-destination rule, line-of-sight, `Range none` handling and the client-side range clamp: not traced [U] |
| (b) d2d status | implemented in `spot()` (`fight.cpp:1134`): in town no, Levels.txt Teleport 0 / 2 + clear line (sampled every fifth of a cell), blocked cell no, the player is set at the exact clicked spot at any distance [S]; ponytail: the line test sampling and the fallback spots [S]; no distance clamp, no mana clamp on arrival [S]; never compared with game.exe for the player [U] |
| (c) bot use | `nav` hops with `CastAt(F2, cell)`, `defaultTeleportRange` 18 subtiles per hop (UNVERIFIED), waits up to 40 s for 24 mana, counts a hop successful only if the position moved, three bad hops end it [S `nav/teleport.go`]; the d2d backend's `cmd skill` for a ground target is an open question (`d2d-backend.md:416`) [U] |
| (d) live | Game.exe: chain tool `act2live` hops, exits, contact crossing (nav3), town refusal (X3) [V `act2.md`]; d2d: Update 2026-11-01: the hop wiring, break-even (18 ticks a hop), path-aware cell, stall escape and pockets are proven live on the lab build (DRILLS24, 25, 27; Stony Tomb 211 vs walking 254, Halls of the Dead 1 533 vs 578) [V live]. The game's own rules (room-based range, landing snap, refused cast is free, Duriel's Lair allows a clear line) were read and run in sk-teleport-trace [V read, V run]; d2d deviates (any distance, mana charged for a refusal) [S]. The hops are wired from the bound skill by the teletune cards (stated by the maintainers of those cards; the older text in (c) above is the pre-wiring state) [S] |

### 3.8 Lightning Mastery (63) used A3+ (passive)

| Aspect | Entry |
|---|---|
| (a) rules | stat `passive_ltng_mastery` = `ln12` = par1 + par2 (L-1) percent added to lightning damage once at missile creation (stat 0x14a on the unit; not on the missile) [V K3 mastery mapping, `skill-damage-trace.md`]; reqlvl 30 [S] |
| (b) d2d status | `passive_stats` + `elem_damage(.., mastery = true)` applied once [S, V mapping by K3 cases with masteries 0..600] |
| (c) bot use | points per table only; no read of the stat [S] |
| (d) live | none [U]. Update 2026-11-01: magnitude 50 + 12 (L-1) percent read once at missile creation [V read, sk-passives-trace]; 1.14d has synergies (base points of other skills) and the table plan was replayed against them [D, I]; not live [U] |

### 3.9 Chain Lightning (53) used A3+ (first in the clvl >= 26 order)

| Aspect | Entry |
|---|---|
| (a) damage | `EMin 1 EMax 40`, min stays 1 whatever the synergies [V `skill-damage-trace.md` min gate]; L1 1-40, L4 1-73, L8 1-117, L12 1-169, L20 1-281 [I]; synergy Charged Bolt + Lightning + Nova x par8 [S] |
| (a) mana | 9 + (L-1): L1 9, L8 16, L20 28 [I] |
| (a) cast time | SQ with seqtrans SC: 19 steps at 0 FCR, 11 at FCR 194 [V K1] |
| (a) hops | `calc1 = ln34 / 5` (hops per cast) [S Skills.txt]; chain target choice, range per hop, repeat hits, hit function: not traced [U] |
| (b) d2d status | implemented as `send(...).hops = calc1` (`fight.cpp:2677`) [S]; cast time 14 vs 19 before the fix/lightning-sequence candidate [V `d2d-fidelity.md` 7a, `cast-trace-gameexe.md`]; the min gate candidate [V]; hop selection never compared [U] |
| (c) bot use | first of the clvl >= 26 order, key F5 / F6 (lightning layout); same gap, cost and low-mana ladder as Charged Bolt [S] |
| (d) live | none [U]. Update 2026-11-01: hops are at most calc1 per cast, next target within 20 subtiles, never a monster in `justhit` [V read, sk-lightning-trace]; d2d re-hits the struck monster on about 73 percent of hops [I, model]; to-hit never misses (equal) [V read]; the hit-once fix is merged upstream as #158, its coverage of the hop path is U; the bot's 650 ms cast gap is shorter than the 760 ms cast (B1) [S]. Not live |

### 3.10 Lightning (49) used A5+ (clvl 40 and on)

| Aspect | Entry |
|---|---|
| (a) damage | `EMin 1 EMax 40`; L1 1-40, L8 1-96, L20 1-272 [I]; min gate [V]; mana 8 + 0.5 (L-1) [I]; SQ 19 steps [V K1] |
| (a) behaviour | `lightningbolt` missile that pierces along its line (pierce, length, hit function) [U] |
| (b) d2d status | plain missile path [S]; cast time / min gate as 3.9 [V]; pierce rule [U] |
| (c) bot use | second in the clvl >= 26 order; table gives it from clvl 40 [S] |
| (d) live | none [U]. Update 2026-11-01: Lightning pierces one unit a frame and skips only the last one hit (D2, V read); about 10 percent more hits per cast in game.exe in packs [I, model]; not live [U] |

### 3.11 Energy Shield (58) in the buff order, never learned by the table

| Aspect | Entry |
|---|---|
| (a) rules | `min(edmn, 95)` % of damage absorbed, mana per point from `calc2 = par5 - Telekinesis blvl` (32 - TK), state length `ln12` = 3600 + 1500 (L-1) frames, mana 5 [S Skills.txt] |
| (b) d2d status | implemented in the hurt path (`fight.cpp:1358`): part = damage x calc1 %, per-16 mana ratio from calc2, uses `>> 8` mana [S]; no game.exe trace [U] |
| (c) bot use | listed in `buffs` (second group); `earlygame.md` row 3: the table never buys it, so the path never fires [S] |
| (d) live | none [U] |

### 3.12 Plain Attack, the staff swing (skill 0) used as the mana-out fallback (A1)

| Aspect | Entry |
|---|---|
| (a) rules | do 1 FUN_0056f070: to-hit FUN_0057ec10 -> FUN_0057d9b0 (AR FUN_00622560 + MulDiv by mastery, 119, stat 325, 179), damage FUN_0057b7d0 -> FUN_0057b420 (stats 21 / 22 << 8 + 111, p = 25 + Str / Dex bonus + mastery + ED, + stats 17 / 18, roll in 256ths), Blunt +50 % vs undead (FUN_0057b660) [V run, sk-melee-trace.md]; a staff is Blunt by ItemTypes equivalence [D; the equivalence walk I]. Speed / IAS not traced [U]. |
| (b) d2d status | same to-hit and whole-point range (100,000 cases, sk-melee-trace.md); floors the physical to whole points (D1, about -0.5 a hit), no undead / demon term (D2), barehanded rule differs (D4) [V run against a port]. |
| (c) bot use | the last rung of the low-mana ladder (life >= `meleeMinLife`), early clvl 1-3 "Fire Bolt on the right hand and a staff swing as the mana-out fallback" [S `resource.md`] |
| (d) live | Game.exe early soaks (melee hits counted, `resource.md` table) [S]; d2d: none with an oracle [U]. Update 2026-11-01: d2d's real build differs from the game (D1 floors damage, D2 no Blunt vs undead, D4 barehanded, D5, D6, D8 off-weapon enhanced damage lost; see the trace) [V run on the real build, sk-melee2-trace]; the bot's resource simulator now rolls to-hit and damage by the game's formula and is fed from the live snapshot (skmeleebot.md, skmeleewire.md) [offline-verified, not live] |

### 3.13 Next band (END table, opt-in; "likely next" for the Blizzard line)

| Skill | (a) game.exe rule | (b) d2d status | (c) bot use | (d) live |
|---|---|---|---|---|
| Ice Bolt (39) | `EMin 6 EMax 10`, L1 3-5, L4 6-9.5, L8 10-15.5, L12 18-25.5, L20 38-49.5, mana 3 [I]; cold length `ELen 150` frames resisted like damage, chill or freeze by FUN_0057b230 [V `static-field-chill-trace.md`, `freeze-rules-trace.md`] | plain missile [S]; chill C1 patched in local fix/static-field-chill, C3 / C4 / C5 in fix/freeze-rules (local, not posted) [V] | END main attack clvl 1-7, fallback; also third in the standard order [S] | none [U] |
| Ice Blast (45) | `EMin 16 EMax 24`, L1 8-12, L4 29-34.5, L8 57-64.5, L12 113-122.5 (no synergies) [I]; mana 6 + 0.5 (L-1) [I]; freeze length / freeze-vs-chill through the cold hit [I]; splash [U] | plain missile [S]; same chill / freeze status [V] | END main clvl 8-17 [S] | none [U] |
| Glacial Spike (55) | `EMin 32 EMax 48`, L1 16-24, L8 65-76.5, L20 225-242.5 [I]; mana 10 + 0.5 (L-1) [I]; splash radius and freeze on the splash [U] | plain missile with a hit function in the d2d list [S]; splash [U]; the freeze trace's drill recipe names it as the live test [S] | END main clvl 18-23 [S] | none [U] |
| Blizzard (59) | `EMin 45 EMax 75`, L1 45-75, L4 90-123, L8 150-187, L12 270-311, L20 570-619 (+ synergy Ice Bolt + Ice Blast + Glacial Spike) [I]; mana 23 + (L-1) [I]; **delay 45 frames** [S]; meteor count `par1`, radius, shard sub-missiles and tick timing [U] | `kMissileAtTarget` (`fight.cpp:2679`: one `launch` at the cast point; sub-missiles / timers in missile_tick) [S]; delay not modelled (K7) [S]; area / tick rules [U] | END main from clvl 24; ground-targeted cast with the Fighter's cast delay awareness ("Glacial Spike / Ice Blast during the delay") [S `skills.go`: `SkillSpec.Delay`] | none [U] |
| Cold Mastery (65) | `passive_cold_pierce` = `ln12` (reduces the target's cold resistance, pierce not damage; unlike Lightning / Fire Mastery) [S Skills.txt; the element byte table maps cold to stat 0x14b, so the "damage" reading of K3 is for the mastery %, the pierce effect is U] | `passive_stats` [S]; the pierce use [V read + V run, sk-resist-trace.md Q4 / Q8: stat 335 only while the monster's cold total is under 100, never breaks an immunity, floor -100; bot ignores it, B2] | END from clvl 30 [S] | none [U] |
| Warmth (37) | `manarecoverybonus` = `ln12` = 30 + 12 (L-1) percent [S Skills.txt]; applied to mana regen [U] | `passive_stats` [S]; regen formula [U] | END 1 point; no read [S] | none [U] |

## 4. Coverage summary

Counts are over the tagged cells of section 3 (statements, one tag each; a cell with several tags counts each). Generated at the end of the run by a tag count of this section; re-run the one-liner in the checklist after editing:

**172 tagged statements in section 3: V 42 (24 %), S 62 (36 %), I 26 (15 %), U 42 (24 %).** About a quarter of what the plan relies on is verified, a quarter is unknown, the rest is read or inferred. Skills with no V on their own behaviour (only common-row V): Frozen Armor, Frost Nova (V only on the freeze rule), Telekinesis, Teleport (V on game.exe live use, not on the rule), Lightning, Energy Shield, the staff swing, Ice Blast, Glacial Spike, Blizzard, Cold Mastery, Warmth.

Reading: V comes from five sources only (K1 cast ticks, K3 damage formula, K4 missile damage on the monster, K6 bind, Static Field / chill / freeze) plus the single Fire Bolt drill; everything about missiles in flight, armors, Teleport, Frost Nova's ring, the ground skills and the passives is S / I / U.

## 5. STEP 3 (PROPOSAL, not run): ranked gap decomps

Rank = (how soon the act is reached: A1 5, A2 4, A3 3, A4-A5 2, END 1) x (gameplay impact: XP/h, deaths, stalls, 1-5) x (likelihood d2d deviates, from what is read: 1-5). Scores are judgements [I]. Every item follows `docs/decomp-queue.md` rules (offline, Opus per lane rule 3, one agent, upstream read worktree on the worker host, GitHub GET-only). "Opus needed" = the trace is a multi-function game.exe read (do a differential in a unicorn oracle) where Sonnet will probably loop; "Opus per lane" = Sonnet could do the structural part, the lane rule says Opus.

| Rank | Id | Skills | Exact question | Why it matters (act, metric) | Oracle approach | Size | Opus |
|---|---|---|---|---|---|---|---|
| 1 (score 100) | SKMISS | Fire Bolt, Charged Bolt, Ice Bolt, Ice Blast, Glacial Spike, Lightning, Chain Lightning (the common missile pipeline: `fire` / `missile_tick` / `land`) | For each missile the player's skill makes (Missiles.txt row of the skill): how many leave a cast, from where, with what velocity per level (`vel + vel_lev x lvl`), range (`range + lev_range x lvl`), what collides (monsters, walls, other missiles), who it can hit, whether a player spell missile rolls to-hit (`UseAttackRate`), pierce / splash on hit, and Charged Bolt's wander (FUN_005c9290) and Chain Lightning's hop choice: does d2d's `fire` + `missile_tick` + `land` produce game.exe's hits per cast? | A1: the whole Act 1-2 plan is Charged Bolt (clvl 1-25) and Fire Bolt; casts per kill and mana per kill decide XP/h and potion burn; d2d's missile list is a reimplementation (`fight.cpp` ponytails 1147, 2646, 2855) never differentialled | structural: a unicorn oracle of the missile creation + per-frame advance (FUN_0059f900 / FUN_0064b860 path already read in the damage trace) vs d2d's missile list for one cast in an empty room; pure functions for velocity, range, bolt count (PR #135 style 100k+ cases); a fan geometry diff for Charged Bolt; hop choice as structural | L | yes |
| 2 (60) | SKARMOR | Frozen Armor (also Shiver, Chilling, Energy Shield: do 18 / 23 / state timers) | (1) Frozen Armor's event 2 (FUN_005cac50): freeze or chill, length (30 + 3 (L-1) frames x synergy), the boss / champion / minion / hireling exceptions, cold immunity and resistance on it; does d2d's `stun_until + chill_until` equal it? (2) state length 3000 + 300 (L-1) frames and the armor percent; (3) Energy Shield absorb and mana ratio; (4) what a re-cast does (refresh vs reject) so the bot's recast timer (finding F2) can be set right | A1 from clvl 5 on every fight; the freeze on a striker is the only crowd control the early Sorceress has and F2 says the armor is off 20 % of the time | pure function differential on the hit effect (reuse the freeze oracle of `freeze-rules-trace.md`, FUN_0057b230 / FUN_0057af80) + structural state timer; bot-side fix (recast timer from the state) as a finding | M | per lane rule (the freeze oracle exists, an Opus agent reuses it) |
| 3 (48) | SKTELE | Teleport | mana by level, the `delay` empty, cast ticks; the destination rule (click point vs clamp to range, snap to the nearest free spot FUN_00554ea0, line-of-sight on Levels.txt Teleport 2, what is refused: town, blocked, other act); the `Range none` handling; does d2d's `spot()` equal it (it places the player at the exact point at any distance)? | A2 (clvl 18): every Act 2 chain leg uses hops (`act2.md`); a d2d that teleports further or through more than game.exe fakes nav speed and stalls | pure function differential for the destination rule (range, snap, refusal) against FUN_0054dc40 / the do 27 handler + Levels.txt cases; mana is K5 | M | yes (disasm of do 27 and the client picker) |
| 4 (48) | SKLIGHT | Lightning, Chain Lightning, Lightning Mastery | the 19-step sequence on d2d (candidate fix/lightning-sequence exists), hop count `ln34/5`, per-hop range and target choice, pierce of `lightningbolt`, the mastery's use as stat 0x14a at creation; the min-1 gate candidate | A3 (clvl 27+ in the FRESH table): the main attack of the lightning plan for clvl 26 to 90; wrong speed or hops moves XP/h in every later act | pure functions for hops / mastery; sequence lines of `cast_rates.py` (exists); the hop / pierce path as structural on the missile oracle of SKMISS | L | yes |
| 5 (45) | SKNOVA | Frost Nova (and Nova, Static Field radius) | ring: how many missiles, at which angles, range (`missile range + calc1`), hit once per monster, damage and cold effect (chill / freeze) per the freeze rules; Static Field's radius unit (`ln12` subtiles) against the bot's `(lvl+4) x 2/3` | A1 from clvl 6: the crowd rule casts it at 3+ hostiles; a d2d ring of 64 missiles may hit or miss differently than game.exe's | structural ring diff + the existing freeze oracle; radius as pure function | M | per lane rule |
| 6 (40) | SKMANA | all used skills | FUN_0056c160 (cost: minmana floor, rounding, `usemanaondo` / `startmana`), when it is paid (cast start vs do), refusal at insufficient mana (status), the 256ths mana pool and regen per tick (mana recovery stat, Warmth), and K1 / K2 on the lab binary (does it carry PR #128?) | every cast of every act; mana is what stops the early Sorceress (`d2d-pace.md`: mana gone after 10-15 casts); the formula is S only | pure function differential of FUN_0056c160 over all 30 Skills rows x levels 1..60 (like `skill_damage.py`); regen as pure function; K2 as one drill | S | no (Sonnet with the oracle skeleton; Opus if the pay timing needs disasm) |
| 7 (30) | SKMELEE | plain Attack with a staff | the player's physical hit (FUN_00647bc0, weapon share, to-hit, speed / IAS, block) vs d2d's `player_blow` / `attack_line` | A1 clvl 1-3 fallback and the low-mana ladder; small but the first levels run on it | pure function differential of the player's blow, 100k cases | M | per lane rule; traced (SKMELEE, 2026-11-01): see sk-melee-trace.md; open Q3 / drill R1 |
| 8 (30) | SKRESIST (done: sk-resist-trace.md, D1 proven on the real build, no PR built yet) | all damaging skills | the monster side of K4 across difficulty: resist rows (Normal / NM / Hell penalties), 100 % immunity and what an immune hit still does (flinch, chill), pierce of masteries (Cold Mastery lowers resist), the bot's `Target.Immune` threshold | A1 (Static Field vs cold immune), NM / Hell (immunities everywhere) | pure function differential on `target_of(...).res` vs the game's resist build for the monster (the function is part of the question) | M | per lane rule |
| 9 (24) | SKGROUND | Blizzard, Meteor, Hydra, Fire Wall, Thunder Storm, Frozen Orb (END table and future) | area, number of sub-missiles, tick timing, `delay` (status 8 not modelled in d2d: K7), radius, damage per tick, Hydra's pet shot rule | only after the END table is chosen (opt-in; Blizzard from clvl 24) | structural: sub-missile timelines (like the missile oracle), K7 as a gate rule | L | yes |
| 10 (12) | SKPASS | Cold / Lightning / Fire Mastery, Warmth, Telekinesis, +skills items | blvl vs lvl (FUN_006442a0 flag), mastery stat stacking, Warmth regen, Telekinesis damage / knockback | A3+ (masteries), END (Warmth) | pure function differential; small | S | no |

Grouping used: the missile skills share one pipeline (SKMISS) so Fire Bolt, Charged Bolt, Ice Bolt, Ice Blast, Glacial Spike, Lightning and Chain Lightning are one decomp; the armor and aura-state skills (SKARMOR) share do 18 / 23 and the state timers; Teleport stands alone; Blizzard, Meteor, Hydra, Fire Wall, Thunder Storm and Frozen Orb share the sub-missile and delay machinery (SKGROUND).

Top six to launch first: SKMISS, SKARMOR, SKTELE, SKLIGHT, SKNOVA, SKMANA. (SKMANA is the cheapest and unlocks the "mana" row of every gate; if the user wants one quick win first, launch it with SKMISS.)

Bot-side findings that need no decomp (card candidates for the fix queue): F1 (Static Field vs the FRESH table and the Andariel card: either put one Static Field point in the table at clvl 6 or drop the opener from the card; the SME text in BUILDV2 called it "a dead point" for Normal), F2 (`BuffRecastSec` 150 s vs 120 s armor; set it from the state or to about 100 s), and B1 of the Static Field trace (range unit).

## 6. STEP 4 (PROPOSAL, not run): the live skill-smoke drill ("SK" group of the `d2drill` harness)

Goal: one short, ledgered drill per used skill that learns N points, casts at a spawned monster or the ground, reads the observables, and compares each with the expected value computed by our oracle (not a number copied from a doc). Runs on one hidden slot of the lab d2d (lockstep, `--no-save`, throwaway Sorceress), group id `SK`, drills `SK.<skill>`; the stage-1 harness is `d2drill` and its `-probe` mode (as `docs/forkcheats.md`). Pass rule for every drill: all observables inside the oracle's interval for at least 95 % of N casts (N given below) AND the recorded mismatches are listed; a skill is "live PASS" only with the drill id and the commit of the lab binary in the matrix.

Setup per drill (existing verbs; every call ledgered with a reason): `debug stat 12 <clvl>` (level), `debug points <skill id> <n>` (learn n points; ids: Fire Bolt 36, Charged Bolt 38, Ice Bolt 39, Frozen Armor 40, Static Field 42, Telekinesis 43, Frost Nova 44, Ice Blast 45, Lightning 49, Chain Lightning 53, Teleport 54, Glacial Spike 55, Energy Shield 58, Blizzard 59, Lightning Mastery 63, Cold Mastery 65, Warmth 37), `debug stat 6` / `debug stat 8` (life and mana, 24.8 fixed point), `debug goto <level>` (Blood Moor 2; Duriel's lair for Teleport 2), `debug spawn zombie1 <n> [near]` (known hp 11, known resists from the table), `step N` (lockstep), `cmd skill <id> X Y [unit]` and `cmd select`, `dump` (player mode, life, mana, rskill, stats), `monsters` (hp, mode, states, resists), `debug missiles` (missiles in flight: unit, missile id, position, speed, ttl, owner, name).

| Skill | Observable | Expected value (oracle) | Notes |
|---|---|---|---|
| Fire Bolt | mana delta in 256ths; hp delta of an unresisting spawn; missiles in flight; ticks of mode SC | cost 2.5 x 256 = 640 exactly; hp loss inside [3, 6] at L1 (more levels: the interval from `tools/emu/skill_damage.py` incl. synergy); exactly one missile; 13 ticks (14 on a binary without PR #128) | the D5.a drill extended to all four; N = 200 casts for the damage mean |
| Charged Bolt | mana delta; number of missiles right after the cast; per-hit hp loss; kills per cast on a pack of 5 | cost `3 + 0.5 (L-1)`; count = `min(24, L + 2)`; per hit inside the per-bolt interval; pack kills compared with the missile oracle of SKMISS (rule: equal in distribution) | L in {1, 4, 8, 13, 20}, N = 100 each |
| Ice Bolt | mana, hp loss, monster state chill / freeze and its length | cost 3; interval; chill length `ELen 150` resisted, plain monster frozen only at coldeffect < 0 (`freeze-rules-trace.md` recipe); boss or champion chilled not frozen | only END table; needs `monsters` states (exists as `monster.states`) |
| Frozen Armor | `dump` player states: `frozenarmor` present and its remaining frames; defense change; a striking zombie's state | state length 3000 + 300 (L-1) frames; defense x (1 + (30 + 5 (L-1)) %); striker frozen / chilled per SKARMOR's result | needs the missing verb `debug states` (below) for exact frames; a second cast resets or rejects (record) |
| Frost Nova | mana; hp loss of N spawns around; their states; missiles at the cast | cost `9 + (L-1)`; each monster hit once; hp loss in the interval; chill / freeze per the freeze rules; missile count equal to game.exe's ring (from SKNOVA) | spawn 8 zombies on the ring at distances 1, 3, 4, 6 subtiles |
| Static Field | hp of spawns; hp fraction removed; radius | `MulDiv(life, 25, 100)` whole points, at most life - 1, none under the floor; radius `5 + (L-1)` subtiles (hit at radius - 1, miss at radius + 1) | in Normal `debug difficulty` to NM to test the 33 % floor; negative resist needs `debug mon set` (below) |
| Telekinesis | mana; hp loss and the knockback of a spawn | cost 7; damage 1 + (L-1) to 2 + (L-1); knockback true | low value; run once at L1 |
| Teleport | mana; new position vs the click; refusals | cost `24 - (L-1)`; position = click snapped to the nearest free spot within the range rule of SKTELE; refused in town (mana unchanged, no move); refused beyond the line rule in Duriel's lair; ticks SC | clicks at 6, 12, 18, 25, 40 subtiles; needs `debug tp <x> <y>` (below) to place the start |
| Lightning, Chain Lightning | mana; ticks of the mode; hp loss; hops | cost per L; **19 ticks** at FCR 0; hops `ln34/5` on a line of zombies; each hit inside the interval incl. the min gate (min 1) | FCR via `debug stat 105 <v>` (verify the setter accepts the stat; else a verb) |
| Energy Shield | mana and life across a `debug hit <dmg>` | absorbed = `min(calc1, 95) %` of the damage, mana lost = absorbed x per16 / 16 | `debug hit` exists (real hurt path) |
| Lightning / Cold Mastery, Warmth | stat 0x14a / 0x14b (or the pierce stat), mana regen over 500 ticks | stat = `ln12`; regen = base x (1 + `ln12` %) | needs `dump stats` for a given id (a read verb) or the existing `debug stat` read mode |
| Bind (all) | `rskill` after `cmd select` of a 0-point skill; right-hand in-town refusal | game.exe rule: select refused, hand unchanged (3) (D1 of `skill-bind-trace.md`); d2d stores it today | the one drill that records a known deviation as expected FAIL until the PR |

Cheats the existing verbs can drive: points, level, life, mana, spawn (with champion / unique mods), goto, `debug hit`, `debug missiles`, `step`, `cmd skill / select`, `dump`, `monsters`.

Cheats that are missing (names only; pre-approved by the user if documented and vetted; each needs a docs/forkcheats.md entry, a negative case and one ledgered live call; never in an upstream PR):
- `debug states [player|<monster id>]`: one row per active state of a unit (state id, name, remaining frames, level, source skill). Needed for Frozen Armor, Energy Shield, Thunder Storm, chill / freeze, curses.
- `debug mon set <monster id> res <fire|cold|ltng|pois> <v>` and `debug mon set <monster id> hp <v>`: set a spawned monster's resists / life to test immunity, negative resist, the Static Field floor.
- `debug tp <x> <y>`: place the player at a free cell (no skill), to test Teleport at fixed distances.
- `debug cast <skill id> <x> <y>`: cast a skill without the bind or the mana check, to isolate the cast pipeline from the bind rule (record both).
- `debug mana full` and `debug life full`: convenience (exists as `debug stat 8` / `6` today, the loop would call it hundreds of times).
- `debug stat <id> <v>` for stat 105 (faster cast rate), 99 (FHR): verify first with one probe call that the setter accepts any stat id; if not, the verb is `debug fcr <v>`.

Observables per drill are all read after `step 2` (the client copy follows the world one tick later), as in the existing drills.

### Per-act gate proposal (matrix-green before the bot proceeds past the act)

"Matrix-green" = the game.exe rule is V, d2d matches (no open deviation in the lab binary, or the deviation is fixed in the lab patch set), and the live SK drill PASSED. The gate is on what the bot casts in that act, not on the whole class.

| Act | Minimal skill set that must be green before progression past this act | Currently |
|---|---|---|
| A1 (clvl 1-16, through Andariel) | Fire Bolt, Charged Bolt (levels 1, 4, 8, 13), Frozen Armor 1 (SKARMOR), Frost Nova 1 (SKNOVA), the common rows K1 / K4 / K5 / K6 (cast ticks on the lab binary, damage on the monster, mana, bind), the staff swing (SKMELEE) for clvl 1-3, and Static Field only if F1 is resolved in favour of keeping it | Fire Bolt: V rule + drill D5.a PASS, d2d truncation fixed in a candidate only (not green); the rest NOT green |
| A2 (clvl 16-26, through Duriel) | A1 set at Charged Bolt 14-20 (bolt count to 22, brackets above 16) + Teleport (SKTELE: mana, refusal, destination, Levels.txt) + Frozen Armor 2 and the recast timer (F2) | NOT green |
| A3 (clvl 26-32) | A2 set + Lightning Mastery, Chain Lightning, Lightning when learned (SKLIGHT, 19 ticks, hops) + SKRESIST for lightning immunes | NOT green |
| A4 / A5 (clvl 32-40) | A3 set + Lightning at its levels + Energy Shield only if the table or the plan starts buying it | NOT green |
| NM / Hell | everything above + SKRESIST across difficulty (traced: sk-resist-trace.md; open: D1 fix, melee D3) (immunities, Static Field floors 33 / 50 %), SKGROUND only if the END (Blizzard) table is the build | NOT green |

Proposed rule for the coordinator: the A1 gate is the blocker before the d2d container's 24 h soak counts as a Sorceress fidelity proof (the soak is XP/h through Charged Bolt, and none of its skill rules is live-proven today); the A2 gate is the blocker before the bot's Act 2 progression; later gates are advisory until a run reaches them.

## 7. Checklist and how to re-count

Re-count the matrix tags: `python3 -I -c "import re,sys;t=open('docs/research/sorc-skill-matrix.md').read().split('## 3. STEP 2')[1].split('## 4.')[0];print({k:len(re.findall(r'\['+k+r'[^\]]*\]',t)) for k in 'VSIU'})"` (tags are counted by the first letter inside a bracket; a bracket like `[V K1]` counts as V).

Open questions this note could not settle (all U above): whether the lab binary carries PR #128; the player missile to-hit rule; Teleport's range and click rule; Frozen Armor's freeze versus chill; Frost Nova's ring size; Chain Lightning's hop rule; FUN_0056c160 and the pay timing; the player's melee blow.
