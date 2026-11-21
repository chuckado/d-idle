# Proving Ground: scenario catalogue and coverage map (2026-11-21)

> status: snapshot for the design review [proving-ground-gate.md](proving-ground-gate.md) (card fablepg, TASK-712). It is a hand-made table and will go stale: the review proposes a generated page from the results ledger (section 5 there) to replace it. Do not copy verdicts from here; follow the source column.

Tags: V = verified in this review (`go run ./cmd/d2scenario list`, `coverage`, `provenance` and the scenario files at 5b12ce5e7), S = taken from a run record of another card (the doc is named), U = no record found in the docs read for this review. No scenario was run for this table.

## 1. How to read the columns

- **setup**: the tag proposed in the review (EARNED, SETUP, CHEATED, IMPOSSIBLE). SETUP is today's `SETUP-only`. Nothing is EARNED today.
- **tier now / proposed**: `fast`, `full`, `nightly` in the file today; T2 smoke, T3 full, T4 e2e as proposed. "T2 after card 4" means the scenario fits the smoke budget but uses wall-clock taps today.
- **stability**: every scenario is `candidate` until it has 5 clean runs with equal verdict and assertion digest (review section 2.7). The column says what evidence exists today.
- **last verdict**: the most recent result on record, with class and seed. All records are from one host (the builder) unless stated.

## 2. Catalogue (38 scenarios: 32 active, 6 pending)

### 2.1 Town, resources, points (no field travel)

| scenario | driver | classes x seeds | virtual min | setup | tier now / proposed | last verdict (source) | determinism evidence |
|---|---|---|---|---|---|---|---|
| heal.town | none | Pal, Sor x 11, 3 | 1 | SETUP | fast / T2 | GREEN Pal 11, Sor 3; GREEN Pal 11 on the second worker (S run record 2026-11-19, `docs/process/second-worker.md`) | 3 runs, 2 hosts, verdict only |
| town.resupply | none | Pal, Sor x 11, 3 | 15 | SETUP | fast / T2 | GREEN Pal 11 three times (S run record 2026-11-19) | verdict equal 3 of 3; state digest differs in 1 of 3 (gold 700 against 712) |
| town.stash-gold-withdraw | none | Pal, Sor x 11, 3 | 20 | SETUP | fast / T2 | GREEN Pal 11 (S same record) | one run |
| belt.refill-after-fight | none | Pal, Sor x 11, 3 | 15 | SETUP | fast / T2 | GREEN Pal 11 (S same record) | one run |
| points.bank-by-default | none | Pal, Sor x 11, 3 | 10 | SETUP | fast / T2 | U | none |
| points.autopilot-spends-by-verb | none | Pal, Sor x 11, 3 | 10 | SETUP | fast / T2 | U | none |
| town.tp-scroll-hoard | none | Pal, Sor x 11, 3 | 15 | CHEATED (given items) | fast / T2 | U | none |
| agency.plan-matches-profile | none | Pal, Sor x 11, 3 | 5 | SETUP | full / T3 (asserts on log lines only) | GREEN Sor 11 once; Paladin not run (S `docs/process/proving-ground.md` section 8) | one run |
| points.sorc-fire-lands-at-30 | none | Sor x 11 | 10 | SETUP | full / T3 | GREEN once (S same) | one run |
| points.necro-summoner-lands-at-30 | none | Nec x 11 | 10 | SETUP | full / T3 | U | none |
| points.paladin-hammerdin-lands-at-30 | none | Pal x 11 | 10 | SETUP | full / T3 | U | none |
| points.barbarian-frenzy-lands-at-30 | none | Bar x 11 | 10 | SETUP | full / T3 | U | none |

### 2.2 Search for Cain

| scenario | driver | start | classes x seeds | virtual min | setup | tier now / proposed | last verdict (source) | determinism evidence |
|---|---|---|---|---|---|---|---|---|
| cain.scroll | none (chunk accepted at setup) | town | Pal, Sor x 11, 3, 5, 7 | 30 | CHEATED (waypoints, level) | fast / T3 | Sor 3 GREEN (S cainfix3); Pal 11 RED before cainfix3, not on record after (U) | none |
| cain.scroll.field | none | field | Pal, Sor x 5 seeds | 15 | CHEATED (position) | full / T3 | Pal 11 GREEN; Sor 3 RED for a setup reason, no weapon (S pgcain) | none |
| cain.scroll.sandbox | none | town | Pal, Sor x 11, 3 | 30 | CHEATED | full / T3, needs the shipped build | RED Pal 11 (S pgcain); not on record after cainfix3 (U) | none |
| cain.read | taps | town | Pal, Sor x 11, 3 | 10 | CHEATED (given item) | fast / T2 after card 4 | GREEN, but the taps log was empty: the board hand-in path was not exercised (S run record) | one run |
| cain.stones | taps | field | Pal, Sor x 11, 3 | 25 | CHEATED (position, given item) | full / T3 | GREEN Pal 11, Sor 3 (S cainfix3) | the only identical full-digest pair on record, on a RED run before the fix (S pgrun2) |
| cain.stones.shipped | taps | field | Pal, Sor x 11, 3 | 25 | CHEATED | full / none: BLOCKED by design, keep as a marker | BLOCKED | n/a |
| cain.cage | none | field | Pal, Sor x 11, 3 | 25 | CHEATED (position, quest word, given item) | full / T3 | GREEN Pal 11, Sor 3 (S cainfix3) | none |
| cain.report | taps | town | Pal, Sor x 11, 3 | 10 | CHEATED (quest word) | full / T2 after card 4 | GREEN Pal 11, 31 s wall (S pgrun2) | one run |
| cain.chain | taps | town | Pal, Sor x 11 | 80 | CHEATED and IMPOSSIBLE (clvl 32, 5 x 31 points, free merc) | nightly / T4 | GREEN: Pal 11 four times, Pal 5 once, Sor 3 three times (S cainchain, pgfix4) | verdict equal in every repeat; no digest pair was ever compared (two result files lost); the bot build changed between two of the runs |
| flail.areachanges | taps | town | Pal, Sor x 11, 3 | 40 | CHEATED | fast / T3 | RED, 26 changes against 12; limit or stall undecided; not re-run (S run record) | none |
| board.two-accepted-chunks | none | town | Pal, Sor x 11, 3 | 30 | CHEATED | fast / T3 | U (a pair was running on the long lane while this was written; V builder status) | none on record |
| board.stop-parked-chunk | API call at a wall second | town | Pal, Sor x 11, 3 | 30 | CHEATED | fast / T3 until card 4 | U (same) | none on record |

### 2.3 Act 1 quests outside Cain

| scenario | driver | classes x seeds | virtual min | setup | tier now / proposed | last verdict (source) | determinism evidence |
|---|---|---|---|---|---|---|---|
| act1.den.clear | taps | Pal, Sor x 11, 3 | 45 | CHEATED (level 10, 8000 gold) | full / T3 | GREEN Pal 11 after the purse was raised (S act1next) | none |
| act1.burial.blood_raven | taps | Pal, Sor x 11, 3 | 60 | CHEATED | full / T3 | GREEN Pal 11 twice; the Kashya hand-in was not tapped (S act1more) | verdict equal, state digest differs |
| act1.forgotten.countess | taps | Pal x 11, 3 | 60 | CHEATED | full / T3 | RED: boss not found in the arena (S countess) | earlier runs of one seed gave different outcomes (S countess, last section) |
| act1.forgotten.countess.sorc | taps | Sor x 11, 3 | 60 | CHEATED; the point-jump leaves 35 mana | full / T3, blocked on the hero library | BLOCKED by the mana guard (S countess) | none |
| act1.tools.malus | taps | Pal, Sor x 11, 3 | 60 | CHEATED | full / T3 | BLOCKED: route guard on belt potions (S act1more) | none |
| act1.andariel.slaughter | taps | Pal, Sor x 11, 3 | 90 | CHEATED | full / T4 | RED: Start never offered (S act1more) | none |

### 2.4 Field behaviour

| scenario | driver | classes x seeds | virtual min | setup | tier now / proposed | last verdict (source) | determinism evidence |
|---|---|---|---|---|---|---|---|
| chicken.recovery | none, `farm:` goal | Pal, Sor x 11, 3 | 15 | CHEATED (position) | full / T3 | GREEN Pal 11, Sor 3 (S pgfix4) | none; the farm timer reads the wall clock (review 2.3, B2) |
| loot.rare-ring-classified | none, `farm:` goal | Pal, Sor x 11, 3 | 10 | CHEATED (given item, position) | full / T3 | GREEN Pal 11, Sor 3 (S pgfix4); an earlier GREEN proved nothing (ring given before the warp) | none |

### 2.5 Pending skeletons (never run)

`town.repair-trip`, `merc.hire-act1`, `difficulty.nightmare-start`, `input.skill-select-verb`, `points.api-skill-legal-and-illegal`, `points.sheet-matches-dump` (V `d2scenario list`; reasons in `docs/process/proving-ground.md` section 8).

## 3. Counts (V unless tagged)

| measure | value |
|---|---|
| scenarios | 38: 32 active, 6 pending |
| by setup tag (active) | EARNED 0, SETUP 11, CHEATED 21 (one of them also IMPOSSIBLE) |
| driven by wall-clock taps or API calls | 13 of 32 active (12 `board-taps`, 1 API) |
| field starts (no `no_writes` evidence: the lab d2d keeps verbs on) | 6 active (`cain.scroll.field`, `cain.stones`, `cain.stones.shipped`, `cain.cage`, `chicken.recovery`, `loot.rare-ring-classified`) |
| with at least one GREEN on record | 17 (S, section 2) |
| with a repeat of one class and seed on record | 6: `heal.town`, `town.resupply`, `cain.stones`, `cain.chain`, `act1.burial.blood_raven`, `act1.forgotten.countess` (the last with different outcomes) |
| meeting the proposed `stable` bar (5 clean runs, equal assertion digest) | 0: the assertion digest does not exist yet |
| behaviours in the catalogue | 97: covered 32, pending-only 7, uncovered 58 |

## 4. Coverage map

### 4.1 By quest (V `d2scenario coverage`; "green" is S from section 2)

| act | chunks on the board | with a scenario | with a GREEN on record | gap |
|---|---|---|---|---|
| 1 | 11 (Den; Blood Raven; Malus; Cain scroll, read, stones, free Cain, report; Countess; Andariel; sail east) | 10 | 7 (Den, Blood Raven, the five Cain chunks) | Malus and Countess blocked, Andariel red, `act1_done.sail_east` has no scenario |
| 2 | 9 | 0 | 0 | everything |
| 3 | 11 | 0 | 0 | everything |
| 4 | 6 | 0 | 0 | everything |
| 5 | 10 | 0 | 0 | everything |

The 37 uncovered quest chunks are 64 percent of the uncovered behaviours. ADR 0020 holds new acts until the Proving Ground exists, so this is the expected state, not a defect; the rule to keep is that an act opens with its chunk scenarios, not after.

### 4.2 By class

| class | scenarios that list it | runs on record | note |
|---|---|---|---|
| Paladin | 28 of 32 active | most | seed 11 almost exclusively |
| Sorceress | 28 of 32 active | clearly fewer than the Paladin's | several REDs were setup (no weapon, no points, 35 mana) |
| Necromancer | 1 (points) | U | |
| Barbarian | 1 (points) | U | |
| Druid | 0 on `main` (a points scenario was waiting on the long lane; V builder status) | none | |
| Amazon, Assassin | 0 | none | |

The same class with a second seed is on record for one scenario (`cain.chain`, Paladin 5). The matrix exists in the files (two seeds for most) and is not being run: "lack of builder time" is the stated reason in two run records (S act1more, pgcain).

### 4.3 By system (behaviour groups of `behaviours.json`)

| system | covered | not covered |
|---|---|---|
| loop hooks | points, belt fill, flail hold, town trip, heal in town, low-life gate, claim, board, task | prepare, player pause, player action, vault keep, idle rest, gear, save point, claim trip, farm (9); repair trip and difficulty are pending-only |
| town chores | stash gold, resupply, scroll hoard | stash items, identify at Cain, vault (3); repair pending-only |
| resources | belt policy, chicken | potion drinking |
| recovery | none | purse recovery, gold farm, corpse recovery (3): the recovery gate caused REDs in four cards and has no scenario of its own |
| loot | pickup and classify | gold pickup |
| travel | waypoint, portal return, area transition | travel hook, sweep |
| combat | fight (only as a side effect of quest scenarios) | no scenario asserts on a fight itself |
| merc | none | revive; hire pending-only |
| points and agency | bank, verb, plan | API pending-only |
| input | none | skill select pending-only |
| shrine | none | use |

### 4.4 What the map says to do first

1. The recovery gate (purse and kit) needs its own scenario before any more quest scenarios are written: it is the most frequent cause of a RED that was not about the quest.
2. Run the second seed and the Sorceress for the scenarios that are already GREEN for Paladin 11 before adding scenarios; one class and one seed is a demo, not coverage.
3. Combat has no assertion of its own. A scripted-fight scenario per supported skill (the L3 matrix of the first review) is the missing tier for new classes.
4. The seven T2 candidates have at most three runs each. They need the 5-run series before anything gates on them.
