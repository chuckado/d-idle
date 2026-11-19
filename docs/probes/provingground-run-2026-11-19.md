# Proving Ground: fast tier on the builder, and what the expansion makes runnable (2026-11-19)

> status: current (card pgexpand). Tags: V verified (run named), S secondary, U unverified. Builder long lane job 20261119-135007-52f0, lab d2d slots, bot of branch agent/pgexpand, Paladin seed 11, one run each, `tools/provingground/run.sh -tier fast -parallel 2`; raw `results.json` in `docs/probes/provingground-run-2026-11-19/fast-tier-results.json`. Every setup is CHEATED or SETUP (see each scenario's `provenance`): mechanics evidence, never balance evidence. A RED here is a finding about the bot or the scenario, nothing was fixed in this card.

## Fast tier, first run (V)

| scenario | verdict | wall | what the state shows |
|---|---|---|---|
| belt.refill-after-fight | GREEN | 149 s | belt topped up (min 3 hp-class column potions; U that 3 is the policy floor) |
| cain.read | GREEN | 113 s | `bkd` in the pack, no `bks`. CAUTION: the driver tapped nothing (empty taps log); the deciphering happened at a normal Akara visit, so the board hand-in path was NOT exercised (the assertion is satisfiable without the board) |
| cain.scroll | RED | 373 s | no `bks`/`bkd`; same signature as pgcain's `cain.scroll` from the camp: the chunk starts, town trips and fights keep the hero off the tree. Trail 1,2,3,4,10,1,10,5,6,5,10,4,1,4,1,4,1,4,3,9,1 |
| flail.areachanges | RED | 387 s | 26 area changes (limit 12). The driver tapped `start scroll`. Trail ends 1,9,13,9,1,2,1: the hero returns to the camp (area 1) six times and never completes the tree step. A town-start scenario with no travel to speak of (town.resupply) already shows 14 changes, so 12 is tight for a 40 minute window that includes the walk to the Dark Wood; U whether the limit or the stall is the finding. Both are worth a look; the stall is the same one as cain.scroll |
| heal.town | RED | 18 s | life 56 percent at the end. The run reached its 4 virtual minutes (`virtual run length reached`), the hero was healed (`status` life 22 then 136 after the chores) and then walked out and fought, so the end state is life after a fight, not after the heal. Scenario artifact (U): needs an observable taken right after the heal, or a goal that keeps the hero in town |
| town.resupply | GREEN | 94 s | hp1 in the belt, `tbk` bought |
| town.stash-gold-withdraw | GREEN | 140 s | stash 128000 to 127250 (750 gold taken), hp1 in the belt, `tbk` bought |

Counts: GREEN 4, RED 3, BLOCKED 0, PENDING 0 (fast tier, 7 runs). `writes after start: 0` on every town-start run (V, the `no_writes` assertion passed where present).

## All 19 scenarios by status (what runs now)

| status | scenarios |
|---|---|
| GREEN (run above) | belt.refill-after-fight, cain.read, town.resupply, town.stash-gold-withdraw |
| RED (run above) | cain.scroll, flail.areachanges, heal.town |
| RED or GREEN by pgcain (`docs/probes/provingground-cain-2026-11-19.md`) | cain.scroll.field GREEN with cainfix2 and `bks` or `bkd`; cain.scroll.sandbox RED (same stall as the lab build) |
| RUNNABLE, not run yet (full tier or nightly, field starts need the attach path) | cain.stones, cain.cage, cain.report, cain.chain, chicken.recovery, loot.rare-ring-classified |
| BLOCKED by design | cain.stones.shipped: a field start needs the lab d2d (verbs on) handed to the bot and the shipped build refuses the warp; stone lit state is a lab fact |
| PENDING | town.repair-trip, merc.hire-act1, difficulty.nightmare-start (reasons in the files) |

Counts over 19: GREEN 4 (plus 1 by pgcain), RED 3 (plus 1 by pgcain), runnable-not-run 6, BLOCKED 1, PENDING 3.

## Decisions (the observables)

- Area changes: the number of `zone entry` lines after the first in the bot's own log (`state.area_changes`, `area_trail`). S: the bot's report of what the game told it. The d2d holds one control client and has no level counter.
- Cairn stones: the observable is `debug portals` listing a portal in level 4 (`portal_open`), V L1 on the lab build; it stays listed from Tristram. The per-stone `NU` to `ON` column is readable only while the hero is in Stony Field and is the bot's own check. Shipped build: BLOCKED (reason above); U that it lists portals.
- Board-driven scenarios use the `board-taps` driver; a scenario that must exercise the hand-in should assert something only the board path produces (open item for `cain.read`).

## Determinism (V, job 20261119-141800-a155 plus the fast-tier run)

`town.resupply` Paladin seed 11, three runs in total: GREEN, GREEN, GREEN, same verdict every time (`d2scenario determinism` exit 0, hard rule met). State digests: `40ec13d9` (fast-tier run), `4e9f6d77` (pair run 1), `40ec13d9` (pair run 2). The digest differs in one run (gold 700 against 712, stash 395 against 250; area changes 14 in all three): the documented wall-clock decisions of the bot loop, a warning not a failure (`docs/d2d-determinism.md`). The second scenario of the pair was not run (one scenario only, U for the rest).

## Coverage (`d2scenario coverage`)

91 behaviours: COVERED 22 (was 10), PENDING-ONLY 5, UNCOVERED 64. Newly covered: hook.flail-hold, hook.heal-in-town, hook.low-life-gate, hook.board, hook.task, hook.claim (by the chain, not yet run), resource.chicken, loot.pickup-classify, travel.portal-return, combat.fight, hook.belt-fill and town.resupply with more scenarios, and the quest chunks `stones`, `free_cain`, `report` (all five Search for Cain chunks now have a scenario). Still uncovered: every quest outside Search for Cain (42 chunks), the vault, gear, claim-trip, farm, recovery and potion-drink hooks.
