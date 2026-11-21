# Town and chores planner: one supply need, one town plan

> status: current (Fable design review 2026-11-21, docs only; nothing here is built or decided). Tree read: `origin/main` at `5b12ce5e7`.
> Tags: V = read in this tree at the cited file:line (code reading, not a run) unless a run id is named; S = secondary; U = unverified judgment. Design statements ("the plan does X") are proposals, not facts.
> Appendices: [A: the chores today and the rule conflicts](town-planner-a-current.md), [B: data model, scenarios, migration](town-planner-b-model.md). Earlier review: [05-architecture](../fable-review-2026-11-19/05-architecture-and-overhaul-options.md) section 3.1 (arbiter, batching), ADR 0020.

## 1. Finding in one paragraph

The town has seventeen chore steps in a fixed order, each with its own need rule, and about thirty numbers that each mean "how many do I need" (appendix A3). The rules that BLOCK the hero (route guard, boss gates, the quest chores step) count healing potions in belt and pack and want 4; the rule that BUYS (the shop) fills the belt plan, whose healing capacity on a 4-slot belt is 3 or 2; the rule that SELLS trims the pack to its own keep line; the rule that BANKS knows no purchase but the shop's. So a guard can ask for something no chore will provide, and the step ends as `needs_attention` with gold in hand (appendix A4, I1 and I2; V runs 20261120-200952-46d2 and the Countess finding 4). Sixteen such pairs are listed in A4. The cause is one missing object: nothing states what the NEXT excursion needs, so nothing can check the hero against it, buy toward it, keep it from the sell chore, or hold its gold back from the bank.

The proposal is that object: a **supply need** per excursion, one **count** of what the hero carries, and a **town plan** computed from the two. The guard, the quest preconditions, the shop, the sell keep, the pickup rule, the bank and the recovery start all read the same need. A guard may only block on a line the plan cannot close, and then it says which one and what closes it.

## 2. Vocabulary (proposed; not in GLOSSARY.md, see question 12)

| term | meaning |
|---|---|
| excursion | the next thing the hero leaves town for: a quest chunk, a farm zone, a boss, a recovery farm, a player errand |
| supply need | per item line, a `min` (below it the hero does not leave) and a `target` (what a visit fills to when gold allows), each with the rule that produced it |
| supply count | what the hero carries and owns: belt, pack, stash, purse, by the one counter |
| town plan | the lines (have, min, target, buy, cost, status), the budget, the ordered stops and the verdict |
| verdict | `ready`, `ready-lean` (min met, target not), `short-gold` (a min line is unaffordable), `blocked` (a min line cannot be bought here at any price) |

## 3. The supply need

### 3.1 Lines

`healing`, `mana`, `rejuvenation` (a cap, not a need), `antidote`, `thawing`, `portal` (charges, loose scrolls and tome together), `identify`, `keys`, `ammo`, `repair` (lowest durability percent at departure), `merc` (alive, or the revive price), `gold_carry` (gold the excursion itself needs in the purse), `free_cells` (pack room for a quest item and loot).

### 3.2 How a need is derived (layers; the larger min and the larger target win)

1. **Build base.** The belt plan of the build (`resource.PlanFor`, `internal/resource/potion.go:124-165`, and the player's belt spec) gives the belt capacity per kind; the pack buffers (`loot.potionBuffer` 4, `loot.potionBufferMana` 2) give the pack part. Base target = belt capacity + pack buffer; base min = today's trip lines.
2. **Excursion profile.** A data row per excursion class (`data/supply/profiles.json`, appendix B2). Today's Go constants become rows with the same numbers: `monastery_road` (healing 4, portal 2, mana 2 when the build needs mana), `forgotten_tower` (same), `blood_raven` (healing 4, portal 2, mana 2, gold_carry 250), `andariel` (healing 4, mana 3, antidote 2), `field` (the base), `recovery_farm` (healing `recoveryPotionFloor`, portal 1). A quest chunk names its profile; a farm zone takes one from its danger tier.
3. **Model estimate (later, shadow first).** `survival.Eco.PotionsPerKill` (`internal/survival/economy.go:16`, `:35`) times the kills expected before the next visit raises the healing and mana TARGET (never the min) for a zone the model has data for. U: the model's potion term is not calibrated for this use.
4. **Feasibility.** A min the hero cannot carry (min above belt capacity plus pack buffer for that kind) is a data error: the loader refuses the row with the class and belt named, and at run time the plan lowers the line to the capacity and logs `need clamped` once. No guard ever waits on an impossible line (this removes I7).

Mana lines come from the build's mana need, never from a class list in Go (removes I15; the data lives in the build profile, appendix B3).

### 3.3 One counter

`supply.Count(snapshot)` is the only code that turns a snapshot into numbers. It replaces `beltAndInv` (`internal/bot/goal.go:220`), the belt loop of `resInput` (`internal/bot/resource.go:41-64`), `kitRead` (`internal/bot/kitgate.go:35-63`), `healingCarried` (`internal/town/basicsfix.go:145-148`) and the direct `belt.PackCount` calls. Rules (proposed defaults, questions 1 and 8):

- Healing and mana count **belt plus pack**, both kinds the same way (removes I8). The belt-fill invariant (section 7, T2) is what makes a pack potion as good as a belt potion.
- A potion under the tier floor (`internal/loot/potionfloor.json`) is counted in a separate field `below_floor`. It counts toward a min only while the verdict would otherwise be `short-gold` (a poor hero uses what it has); it never counts toward a target, is never bought, and is the first thing sold. One rule for pickup, guard, shop and sell (removes I6).
- Portal = charges in the tome plus loose scrolls; identify the same.
- Gold is reported as purse, stash and wealth (the sum). Which of the three a rule reads is stated per rule in section 5, not chosen at each call site.

## 4. The town plan

### 4.1 Solve

`supply.Solve(need, count, prices, policy) -> Plan` is a pure function (no input, no clock), so it runs in the guard, in the viewer and in a unit test with the same result.

1. For each line: `buy_to_min = max(min - have, 0)`, `buy_to_target = max(target - have, 0)`, cost from the vendor prices when the shop is open, else from the kit prices (`resource.kitPotionGold` 30, `kitTPGold` 100) as today (`internal/town/basicsfix.go:178-199`).
2. Budget = purse + stash gold + the estimated sale of what the sell policy will sell, minus the reserve lines (merc revive when the knob is on, `gold_carry`).
3. Spend in priority order (4.2) until the budget is out. A line that got less than its min makes the verdict `short-gold`; a line that got its min but not its target makes it `ready-lean`.
4. Emit actions and order them into stops (4.3).

### 4.2 Priority when gold is short (what is dropped first is the bottom of the list)

| prio | line | today's rule it keeps |
|---|---|---|
| 0 | free actions: heal, belt fill from the pack, equip, identify with a held tool, sell, stash items | |
| 1 | healing to min | potions before the first scroll, user decision D3 (`town.potionsFirst`, contract `docs/contracts/town-basics.md:29`) |
| 2 | one portal charge | `tpAllowed` (`basicsfix.go:123-133`) |
| 3 | a repair the player requested | moves ahead of supplies today (`internal/town/chores.go:456-473`) |
| 4 | the build's weapon supply (ammo for a launcher build) | "the bow is the only attack" (`chores.go:430`); a flag in the build profile, not a class check |
| 5 | mana to min, antidote and thawing to min | `resource.ManaBuy` (`internal/resource/poverty.go:158-166`) |
| 6 | `gold_carry` of the excursion | new (I10) |
| 7 | portal to min charges, repair below the repair line | |
| 8 | merc revive | `town.mercReserve` |
| 9 | targets above min: full belt, pack buffer, tome fill | |
| 10 | tome purchase, identify tome and scrolls, keys | `town.tpTomeMinGold`, `town.minGoldReserve` |
| 11 | belt item, gear shop, gamble | `town.beltBuyLevel`, the gear budget |

The order is data (`policy.priority` in the profile file) with this list as the default; the trace names every dropped line with the gold it was short (section 8).

### 4.3 One visit, cheapest order

Actions are grouped into stops: inventory (no walk), Cain, each vendor, the stash, the merc NPC. Precedence constraints:

- belt fill and equip BEFORE sell (so the sell never sees a potion the belt wants: removes I4);
- identify before sell; sell before buy at the same vendor;
- the stash stop is ONE stop that settles the purse (5.1): it comes before the first paid stop only when the purse cannot pay that stop, else last;
- heal is free at the healer when her window opens (`chores.go:449`), so it rides the potion stop.

Within the constraints the order is the cheapest walk by nav distance between the stop positions read from the presets at run time (at most six stops: brute force). In Act 1 the stops are the repair vendor (sell, ammo, belt item, repair), the potion vendor (heal, potions, scrolls, tome, antidotes), the stash, Cain, the merc captain and the gamble vendor; the table is `town.Acts` (`internal/town/acts.go:41-52`) for every act, so nothing is named in Go. U: the saving against today's fixed order is not measured; today the stash can be visited three times in one visit (appendix A1).

Each action is verified as today (gold change, item moved) and reported; a failed action keeps its line open and the verdict honest.

## 5. Policies that fall out of the plan

### 5.1 Bank: settle the purse once

`purse_target = planned spend of this visit + gold_carry + carry line`, where the carry line is today's `300 + 50 x clvl` (`internal/town/stash.go:124-131`). At the stash stop: purse above the target deposits the surplus; purse below it withdraws the gap (never more than the stash holds). `bank`, `withdraw`, the repair's own withdrawal and the shop withdrawal become this one action (removes I9, I10, I11). The stash holds everything else. Rules that ask "can the hero afford X" read WEALTH; only the death-loss question reads the purse (U: the gold lost on death is taken from carried gold first, which is why the purse stays small; contract `docs/contracts/death-rule.md:70`).

### 5.2 Sell: never what the plan needs

The sell chore keeps, per line, `max(target, min)` usable items, whatever the pack fullness; it sells below-floor potions first, loose scrolls above the cap when a tome is held (today's #1013 rule), and the surplus above target only under pack pressure. `town.spareKeep` becomes the default pack buffer of lines that have no other target (removes I3, I5).

### 5.3 Pickup: wanted while under target

`loot.Needs.Want("potion" | "scroll" | "key")` answers `have < target` from the same plan line (removes I5, I12). The belt-landing rule for ground potions stays as it is (`internal/loot/loot.go:565-572`).

### 5.4 Guards and quest preconditions: read the verdict

`RouteKitWhy`, `TowerKitWhy`, `BossGate.kitShort`, `brShort` and `StepChores.Need` (`internal/task/steps.go:104`) become one call: `plan := Solve(needOf(excursion), count)`. Outcomes:

- `ready` or `ready-lean`: go.
- a min line open and the hero is NOT in town: a town trip with this plan (today's inserted chores step).
- a min line open, in town, the plan can close it: run the plan. A step that "ran but is still needed" can then only be a real failed purchase, and its text names the line.
- `short-gold`: not `needs_attention`. The step hands over to the gold sources in order: sell, stash, the vault sale prompt when the knob is on (5.6), then recovery or the player's gold farm, with "needs N more gold for <line>" as the reason.
- `blocked`: `needs_attention` with the line named (vendor has no stock, act has no such vendor).

The level and merc parts of the guards are not supplies and stay where they are.

### 5.5 Recovery and poverty: same lines

Recovery starts when the verdict for the `recovery_farm` profile is `short-gold` on WEALTH (not on a purse constant of 500), and ends when the min lines of the excursion the hero was held from are affordable. `resource.recoveryTargetGold` stays as the farm margin. `PovertyCheck`, `KitShortfall`, `SavingFor` and `goldfarm.Want` become views of the plan ("the first line the gold does not cover").

### 5.6 The Vault and the gold sources

With `loot.vault` on, items are recorded, not carried, so the sale estimate of the pack is smaller and the vault is a gold source the plan can name but never use by itself: when the verdict is `short-gold` and `loot.vaultSellPrompt` is on, the Needs-you row says "short N gold for <line>: sell these vault items?" (the existing prompt, `internal/env/tuning_loot.go:40-41`). Measured runs force the vault off (`tuning_loot.go:36`), so the plan there has pack sale, purse and stash only.

### 5.7 Trips from the field

One trigger replaces the three deciders (appendix A2): a trip is due when a line is under its `trip_below` (defaults: today's `town.tripMinHPPotions` 2, `town.tripMinMPPotions` 1, column fill 50 percent, pack 80 percent, durability 20 percent), the plan for the CURRENT excursion would change something (the futile memory, `chores.go:54-63`, becomes "the plan buys nothing new"), and a way home exists. A portal is cast only outside a town: the check sits in the one cast function (removes I13).

## 6. The plan as an arbiter intent

The arbiter already has the slot: when an upkeep intent for town wins, every `TownVisitor` is asked what it wants in the same visit (`internal/arbiter/arbiter.go:69-73`, `:254-262`). Proposed:

- Source `town.plan` proposes ONE upkeep intent (`NeedsArea: town`) when 5.7 says a trip is due; urgency from the worst line (`have / min`), `EstCost` from the stop route, `Steps` = the plan's actions, `Reason` = the trace line. It replaces the sources `trip` and `resupply` (`internal/arbiter/sources.go:66-98`), whose "withdraw_gold, buy_potions" steps are a two-line version of this plan.
- `WhileInTown` contributions from other sources (claim, hand-in, vault keep, player item actions, gear) are STOPS added to the same route, not separate visits.
- The plan is solved twice: on arrival (what to do) and at the town exit (the departure check, which is the guard of 5.4). The second solve is the only place a leave can be refused, and it logs both reasons like every arbiter refusal.
- `Report(intent, false)` only for `blocked` or a failed action. `short-gold` is not a failure of the town intent: it raises the gold objective (recovery or the gold farm) with the amount, so the back-off does not hide a restock that became affordable.

Until the arbiter drives (it is in shadow, `docs/arbiter.md`), the same `Plan` is called from `choresDue` and the adapter; the plan does not depend on the arbiter.

## 7. Invariants (each has a scenario in appendix B4)

- T1 A guard or chores step never waits on a line the plan can close with the gold the hero owns.
- T2 After a visit the belt holds every usable potion the plan has a column for; a pack potion of a planned kind exists only when its belt columns are full.
- T3 A visit never lowers a line below `max(min, min(target, what it was))`: the sell chore does not sell what the plan needs.
- T4 The purse after the stash stop is at least planned spend + gold_carry; wealth changes only by sales and purchases.
- T5 Loose scrolls beside a tome never exceed the cap; no portal cast is attempted in a town.
- T6 With gold short, lines are bought strictly in priority order, and the trace names each dropped line.
- T7 One visit: each stop is visited once, except a second pass forced by a failed action.
- T8 The same snapshot and need give the same plan (pure function; determinism pair).
- T9 No class or skill literal in `internal/supply`; every rule passes for the default Sorceress and one other build.

## 8. What the player sees

A card on the watch page, "Town plan", fed by `GET /api/townplan` (schema `watch.townplan/1`, additive):

- header: the excursion ("Next: Tools of the Trade, Barracks") and the verdict chip (Ready, Ready (lean), Short of gold, Blocked);
- the shopping list: one row per line with have / min / target, the buy count, the cost and a status word (ok, buy, short, dropped), and the rule that set the min ("Monastery road profile", "belt plan", "survival model") on hover;
- the budget row: purse, stash, expected sales, reserved (merc, chunk), left;
- the stops in order, ticking off during a visit;
- when a step waits: one sentence, the same text as the guard: "Waiting: healing potions 3 of 4. Buying 1 at Akara (30 gold)." or "Waiting: 310 gold short for a Tome of Town Portal (dropped, not needed for this step)" or "Short of gold: needs 120 more for healing potions. Selling 6 items, then farming gold in Cold Plains."

The minute line and the log get one line per solve:

`town plan: for=tools_of_the_trade/malus verdict=ready-lean buy=hp2x1(30),tsc x1(100) sell=6(~140) stash=withdraw 0 deposit 5200 dropped=tbk(short 310) stops=charsi,stash,akara`

The player can change the plan only through what already exists: the belt spec, the knobs, the repair and farm-gold buttons, sell and keep marks. A mark always wins over the plan (keep is never sold, `internal/town/shop.go:655-663`).

## 9. Open questions (the user decides; recommendations are mine)

1. **Does a guard count belt plus pack, or the belt only?** Recommendation: belt plus pack for healing AND mana, with invariant T2. Belt-only makes four of today's lines unsatisfiable on a 4-slot belt.
2. **Keep today's min numbers (healing 4, mana 2 or 3, portal 2) as the profile defaults?** Recommendation: yes, unchanged, in the first cards; revisit with the survival model, not by argument. The numbers are U judgments in the code comments too (`internal/task/gate.go:172-176`).
3. **Purse carry line.** Recommendation: keep `300 + 50 x clvl`, and make recovery and the gates read wealth (question 4), so the line no longer has to be above 500.
4. **Recovery reads wealth, not the purse.** Recommendation: yes. A hero with 6900 gold in the stash is not broke. Cost: a stash the bot cannot read counts as 0 (it is logged today, `basicsfix.go:223-225`).
5. **Order under short gold.** Recommendation: the table in 4.2, which keeps decision D3 (potions before the first scroll). The one new position is `gold_carry` at 6.
6. **Vault items as a gold source.** Recommendation: never automatic; only through the existing prompt, and never in measured runs.
7. **Does a quest chunk declare its supplies?** Recommendation: a `supply` profile name and an optional `gold` number per chunk in `data/questboard`, replacing the boolean `noGold` (`gold: 0` means the same). Tools of the Trade then states its real need. U: the Malus hand-in and the imbue cost no gold in the game (from memory, not checked against a source here); the chunk's gold need is the resupply it will buy on the way.
8. **Below-floor potions.** Recommendation: counted toward a min only while gold is short, never bought, sold first.
9. **Survival-model sizing of targets.** Recommendation: shadow only (log "model would carry N") until the Proving Ground has two green runs per class on the plan; then a knob.
10. **One staged knob `town.plan`** (0 off, 1 shadow, 2 guards and shop, 3 purse, 4 sell and pickup, 5 route), like `loot.rulesEngine` (`tuning_loot.go:43`). Recommendation: yes; measured runs record the stage in the run record. The default moves one stage per green card.
11. **Contract text.** `docs/contracts/town-basics.md` "Bank before the shop" (`:43-44`) and the death-rule carry line would be reworded to "the purse is settled once per visit to planned spend + chunk gold + carry line". Recommendation: change it in the card that changes the code (B5 card 5); it is sourced from the basics audit, not from a user decision, but the user should see the new sentence.
12. **Glossary and ADR.** Proposed glossary terms: section 2. Proposed ADR text (not written; for the user's session): "One supply need. Every rule that asks whether the hero has enough potions, scrolls, ammunition, repairs or gold for where it goes next reads the supply need of that excursion and the one supply count. Town visits execute a town plan solved from the two. A guard may block only on a line the plan cannot close, and names it. Numbers live in profile data with the build's needs in the build profile; no class or skill is named in the planner."
13. **Who owns the profile numbers?** Recommendation: the SME reviews the profile file once (one card), since every number in it is a U judgment today.

## 10. What this review did not do

No run, no Go change. The cited runs used CHEATED setups (mechanics only). The saving of the batched route, the model-sized targets and the gold-on-death rule are U. Appendix A6 lists the files not read.
