# Town planner, appendix B: data model, scenarios, migration

> status: current (design review 2026-11-21, docs only; proposals, nothing built). Main doc: [town-planner.md](town-planner.md); current state: [appendix A](town-planner-a-current.md).
> Tags: V = read in this tree at the cited file:line; U = unverified judgment. Type and file names below are proposals.

## B1. Package and types

A new pure package `internal/supply` (no input, no clock, no snapshot type beyond the counter's argument). Sketch, not final code:

```go
type Line string // "healing", "mana", "rejuvenation", "antidote", "thawing", "portal", "identify", "keys", "ammo", "repair", "merc", "gold_carry", "free_cells"

type Qty struct {
    Min, Target, TripBelow int
    MinWhy, TargetWhy      string // the rule that set it: "profile monastery_road", "belt plan", "model"
}

type Need struct {
    Excursion string        // "quest:tools_of_the_trade/malus", "farm:cold_plains", "recovery"
    Profile   string        // row of profiles.json
    Lines     map[Line]Qty
}

type Count struct {
    Belt, Pack, BelowFloor map[Line]int
    Capacity               map[Line]int // belt columns x rows of the plan, per kind
    PortalCharges, LooseScrolls int
    HasTome                bool
    Purse, Stash           int
    StashKnown             bool
    LowestDurabilityPct, RepairCost int
    MercDead               bool
    MercReviveCost         int
    SaleEstimate, Sellable int
    FreeCells              int
}

type PlanLine struct {
    Line                   Line
    Have, Min, Target, Buy int
    Cost                   int
    Status                 string // ok, buy, short, dropped, clamped
    Why                    string
}

type Plan struct {
    Need    Need
    Lines   []PlanLine
    Budget  Budget   // purse, stash, sale, reserved, spend, left
    Purse   int      // the purse target of the stash stop
    Stops   []Stop   // ordered; each with its actions
    Verdict Verdict  // ready, ready-lean, short-gold, blocked
    Short   int      // gold missing for the first open min line
}

func NeedFor(excursion Excursion, build BuildSupply, tune Inputs) (Need, error)
func CountOf(s *d2.Snapshot, floors Floors) Count
func Solve(n Need, c Count, prices Prices, pol Policy) Plan
```

Errors, never fallbacks: an unknown profile, an unknown line, a build without a supply block or a min above the carry capacity returns an error that names the profile, the build and the line.

## B2. Profiles as data

`data/supply/profiles.json` (embedded, validated at load like `internal/loot/potionfloor.json`). Rows keyed by profile name, parameters per difficulty; the first version copies today's constants exactly:

| profile | healing min | mana min | portal min | other | today's source (V) |
|---|---|---|---|---|---|
| `field` | base | base | 1 | | trip knobs, appendix A3 |
| `monastery_road` | 4 | 2 when the build's mana need is 2 | 2 | | `internal/task/route.go:169-171`, `:213-226` |
| `forgotten_tower` | 4 | `TowerKitMana` or 2 | 2 | | `route.go:128-146` |
| `blood_raven` | 4 | 2 | 2 | gold_carry 250 | `internal/task/gate.go:28`, `:49-70` |
| `tools_of_the_trade` | 4 | 2 | 1 | | `gate.go:182` |
| `countess` | 4 | 2 | 1 | | `gate.go:183` |
| `andariel` | 4 | 3 | 1 | antidote 2 while wealth is at least 150 | `gate.go:184`, `:85`, `:221` |
| `den` | 2 to enter, 1 to stay | | | | `internal/bot/goal.go:250-251` |
| `recovery_farm` | `resource.recoveryPotionFloor` (2) | `recoveryManaFloor` (2) | 1 | | `internal/resource/recovery.go:39` |

A quest chunk names a profile in `data/questboard/act*.json` (`"supply": "monastery_road"`, optional `"gold": N`); `noGold: true` reads as `gold: 0` until the files are migrated. A farm zone takes its profile from its danger tier in the plan table (default `field`).

## B3. Knobs become plan inputs (today's values stay the defaults)

No knob changes its default in the migration; each becomes a field the planner reads, so `docs/knobs.md` stays the owner of the numbers.

| knob (value) | plan input |
|---|---|
| `town.beltFillPct` (100), belt spec, `resource.PlanFor` | `Capacity`, base healing and mana target |
| `loot.potionBuffer` (4), `loot.potionBufferMana` (2) | pack part of the base target |
| `town.spareKeep` (4) | default pack buffer of a line with no other target |
| `loot.rejuvSpare` (3) | rejuvenation cap |
| `town.tripMinHPPotions` (2), `town.tripMinMPPotions` (1), `town.tripBeltColumnPct` (50), `town.tripPackUsedPct` (80), `town.tripRepairPct` (20), `resource.tripHPMin` (1), `bot.resupplyPotions` (2) | `TripBelow` per line; the three healing knobs collapse to one line (the largest wins until one is retired, B5 card 10) |
| `town.tpScrollTarget` (2), `town.looseScrollKeep` (2), `town.tpReserve` (2), the literal 5 | portal min, loose cap, poor min, tome fill target |
| `town.idScrollTarget` (2), `town.identifyTome` (0), `town.idScrollNeed` (1) | identify line |
| `town.keyTarget` (0) | keys line |
| `town.repairBelowPct` (30), repair gate modes | repair line |
| `town.stashGoldBase` (300), `town.stashGoldPerLvl` (50) | carry line of the purse target |
| `town.minGoldReserve` (100), `town.mercReserve` (1), `bot.goldFloor` | reserves in the budget |
| `resource.kitTPGold` (100), `resource.kitPotionGold` (30) | fallback prices when no vendor stock is read |
| `resource.hpFirst` (3), `resource.mpFirst` (2), `resource.kitMinHP` (4), `resource.kitMinMP` (3) | min of the `field` profile below and from `resource.kitGateLevel` (8) |
| `resource.recoveryKitGold` (500), `recoveryTargetGold` (900) | recovery start margin on wealth, farm margin |
| `town.potionsFirst` (1), `town.tpScrollPotions` (2), `town.tpMinGold`, `town.tpFirstLevel` | `Policy.priority` rows 1 and 2 |
| `town.bankFirst` (1), `town.withdrawShop` (1) | retired by the purse settle (stage 3); kept as aliases that select stage behaviour until then |

Per build, in the build profile (the combat profile or plan table; the build-flexibility rule): `supply.manaNeed` (0, 1, 2: today a class switch at `internal/resource/types.go:172-180`), `supply.beltColumns` by level band (today `potion.go:124-165`), `supply.weaponSupply` (`ammo` for a launcher build, else none), optional per-line overrides. `World.Caster` and `World.BRHigh` (`internal/bot/task.go:357`, `:359`) are filled from the build, not from class lists. Unknown build or missing block: an error naming it, never the Sorceress values.

## B4. Proving Ground scenarios (written first, RED on today's main where marked)

Format `didle.scenario/1` (`tools/provingground/scenarios/`); assertion kinds that exist today (V `internal/scenario/state.go:81-100`): `item_in_pack` (counts pack, belt, stash and body), `gold_between`, `stash_between`, `wealth_between`, `quest_bits`, `log_line`, `area_changes_le`, `no_writes`. One format addition is needed: `belt_at_least {kind, min}` (the belt alone; `item_in_pack` cannot tell belt from pack). Every scenario runs the matrix Paladin and Sorceress, and carries `no_writes`. Setups are SETUP writes (CHEATED where the level or gold is not reachable by play): mechanics evidence only.

| id | invariant | setup | assertion sketch | today |
|---|---|---|---|---|
| `town.plan.guard-blocked-restock` | T1 | clvl 22, 1220 gold, 4-slot belt with 3 healing and 1 mana potion of the floor tier, empty pack, waypoint to the Outer Cloister; tap the Malus chunk | `log_line "^quest board: starting chunk" min 1 to_min 10`; `log_line "ran but is still needed" max 0`; `log_line "needs_attention" max 0`; `item_in_pack hp2 min 4` | RED (I1; run 20261120-200952-46d2) |
| `town.plan.broke-after-bank` | T4 | clvl 3, 8000 gold in the purse, empty belt, no scroll; goal: the Den chunk | `log_line "^recovery: start" max 0`; `log_line "^quest board: starting chunk" min 1 to_min 10`; `wealth_between min 7000`; `stash_between min 1` | RED expected (I9; Den finding 1 had the same shape; U whether this exact setup reproduces it) |
| `town.plan.malus-errand-gold` | T4 | clvl 22, 8000 gold, default `town.bankFirst`, a full kit; tap the Malus chunk | `log_line "^town plan: .*verdict=ready" min 1`; chunk starts by minute 10; `gold_between min <carry line + plan gold>`; `stash_between min 1`; full tier: `quest_bits record 3 set 1` | RED (I10; run 20261120-184924-bb3e) |
| `town.plan.belt-refill-from-pack` | T2, T3 | clvl 16, 0 gold, belt with 1 healing potion, 8 floor-tier healing and 4 mana potions in the pack, pack 85 percent full of junk | `belt_at_least healing <capacity>`; `belt_at_least mana <capacity>`; `item_in_pack hp2 min <capacity + buffer>`; `log_line "^town plan: .*sell=" min 1` | U (expected RED on the keep count, I3 and I4) |
| `town.plan.sell-keeps-need` | T3 | the Countess setup before its workaround: clvl 24, pack full of 16 healing potions, default `town.spareKeep` | after the first visit `item_in_pack hp2 min <target>`; `log_line "tower guard \\(.*healing" max 0` | RED (I3; Countess finding 1) |
| `town.plan.scroll-hoard` | T5 | the existing `town.tp-scroll-hoard` setup plus four scrolls on the ground by a full tome | `item_in_pack tsc max 2`; `item_in_pack tbk`; `log_line "town portal did not appear" max 0` | green in town since #1013; the ground half U |
| `town.plan.no-cast-in-town` | T5 | arrive in town through a portal whose entry is not confirmed at the first try (U: needs a way to force it; else a unit test of the cast function) | `log_line "town portal did not appear" max 0`; `log_line "portal cast refused: in town" max 0` | U (I13) |
| `town.plan.mana-kit-satisfiable` | T1 | clvl 10 mana-need-2 build, 2000 gold, no belt item | `log_line "kitgate-skip" max 0`; `item_in_pack mp1 min 3`; no chunk hold on mana after minute 5 | RED for the Sorceress (I7) |
| `town.plan.gold-short-priorities` | T6 | clvl 6, 130 gold, empty belt, no scroll, nothing to sell | `item_in_pack hp1 min 3`; `item_absent tsc`; `log_line "^town plan: .*verdict=short-gold.*dropped=.*portal" min 1`; `gold_between max 40` | partly green today (potions first is built) |
| `town.plan.short-gold-handover` | T1 | clvl 12, 20 gold in purse and stash, empty belt, 6 junk items; a chunk tapped | `log_line "needs_attention" max 0`; `log_line "^town plan: .*sell=6" min 1`; then `log_line "^recovery: start|farm gold" min 1` with the amount in the line | RED (Countess finding 4) |
| `town.plan.one-visit` | T7 | clvl 20, damaged gear, junk, low belt, gold above the keep line | `log_line "^town plan: .*stops=" min 1`; `log_line "^town plan: .*stops=[a-z,]*stash,[a-z,]*stash" max 0` (the stash is one stop); `log_line "chore budget exhausted" max 0` | U (today up to three stash stops) |
| `town.plan.determinism` | T8 | any of the above, same seed twice | the `town plan:` lines of both runs are equal (the determinism pair of `internal/scenario/determinism.go`) | U |

Offline twins: each row also gets a `go test` case on `supply.Solve` with the same numbers (pure function, runs on a small box), so a card can go green offline before the builder is free; the Proving Ground run stays the acceptance.

## B5. Migration: eleven small cards, scenario first

Every behaviour card flips ONE stage of the knob `town.plan` (question 10) and names the scenario that must go from RED to green. Rollback of any card is the previous stage value; stage 0 is today's code path, kept until card 11.

| # | card | change | scenario gate | risk | rollback |
|---|---|---|---|---|---|
| 1 | `tp-scenarios` | the B4 scenarios and `belt_at_least`; run once, record RED and green in a probe doc | none (it is the gate) | scenario setups that cannot reach the state; low | delete the file |
| 2 | `tp-count` | `supply.CountOf`; the five old counters call it; a test proves equal numbers on the recorded fixtures and lists where they differed (I6, I8, I16 become visible) | all existing town scenarios unchanged | a counter that differed on purpose; medium | revert one commit; no knob |
| 3 | `tp-need-shadow` | profiles file, `NeedFor`, `Solve`; stage 1 logs `town plan (shadow)` beside every chores call and every guard verdict, with `agree=true/false` | shadow line present; no behaviour change | none on behaviour | stage 0 |
| 4 | `tp-guards` | stage 2: guards, boss gates and `StepChores.Need` read the verdict; the shop buys to the plan (pack spares to the min) | `guard-blocked-restock`, `mana-kit-satisfiable` | a guard that opens too early; the min numbers are unchanged, so only impossible lines open | stage 1 |
| 5 | `tp-purse` | stage 3: one stash stop settles the purse; recovery, poverty and the gold gates read wealth; contract lines reworded (question 11) | `broke-after-bank`, `malus-errand-gold`, existing `town.stash-gold-withdraw` | more gold carried when a plan is large, so more lost on a death; bounded by the carry line plus the plan | stage 2 |
| 6 | `tp-sell-pickup` | stage 4: sell keep and pickup want read the line target; belt fill before sell; scroll numbers from the portal line | `belt-refill-from-pack`, `sell-keeps-need`, `scroll-hoard` | a pack that fills with supplies; the pack-pressure rule stays | stage 3 |
| 7 | `tp-short-gold` | priority table as data; `short-gold` hands over to sell, stash, vault prompt, recovery; no `needs_attention` for gold | `gold-short-priorities`, `short-gold-handover` | a recovery loop when income is zero; the stall rule stays (`recovery.go:77`) | stage 4 with the old handover |
| 8 | `tp-route` | stage 5: stops ordered by nav cost under the precedence rules; the fixed list retired behind the knob | `one-visit`, `determinism` | an NPC walk that fails more often in a new order; the chore budget stays | stage 4 |
| 9 | `tp-viewer` | `GET /api/townplan`, the Town plan card, the minute line, the waiting sentence | UI test on recorded plans; `log_line "^town plan:"` | none on play | hide the card |
| 10 | `tp-arbiter` | `town.plan` as arbiter source and visitor; the three trip deciders collapse to one; needs the arbiter out of shadow for the town family (roadmap R4) | arbiter replay fixtures; `area_changes_le` on the ping-pong fixture | depends on R4; do not start before | the arbiter knob |
| 11 | `tp-cleanup` | delete stage 0 code and the retired knobs (`town.bankFirst`, `town.withdrawShop`, duplicate trip knobs), update `docs/town.md` and the contracts | the whole B4 set green twice per class | removing the escape hatch; only after two clean nightly runs | revert |

Optional later card: `tp-model` (survival-model sizing of targets, shadow first, question 9).

Sizing (U): cards 2 to 8 are each one package family and under about 400 lines with tests; card 4 is the first that changes play and is the one that unblocks Tools of the Trade and the Countess scenario.

Dependencies: 1 before all; 2 before 3; 3 before 4 to 8; 9 after 3; 10 after 8 and after R4; 11 last. Cards 4, 5 and 6 touch `internal/town` together and must not run in parallel.

## B5a. Status (updated 2026-11-22)

| stage | card | state | evidence |
|---|---|---|---|
| 1 shadow | tp-need-shadow | merged | [stage 1-2 probe](../../probes/town-plan-stage2-2026-11-21.md) |
| 2 guards and shop | tp-guards | merged, default stays 0 | same |
| 3 purse | tp-purse (TASK-757) | on branch `agent/townstage3`, default stays 0 | [stage 3 probe](../../probes/town-plan-stage3-2026-11-22.md) |
| 4 sell and pickup | tp-sell-pickup (TASK-758) | not started | |
| 5 route | tp-route (TASK-761) | not started | |

## B6. Tests required by the build-flexibility rule

Every `Solve` and `NeedFor` test runs for the default Sorceress and one other build (a mana-need-1 and a mana-need-0 build cover the three column layouts). `internal/supply` imports no class constant; a lint test greps the package for class names and skill ids.
