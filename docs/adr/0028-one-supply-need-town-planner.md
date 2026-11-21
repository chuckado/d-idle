---
status: accepted
date: 2026-11-21
decision-makers: user (popups, Fable review 2026-11-21, town planner)
---
# 0028: One supply need (town planner)

Source: user popups, 2026-11-21, on [town-planner.md](../design/fable-review-2026-11-21/town-planner.md) section 9 (Q1 to Q13), with [current state](../design/fable-review-2026-11-21/town-planner-a-current.md) and [model](../design/fable-review-2026-11-21/town-planner-b-model.md). Builds on [ADR 0020](0020-proving-ground-and-stat-skill-agency.md) (arbiter intent, Proving Ground scenarios).

## Context

Several rules (guards, shopping, recovery, the purse, chunk gold) each ask in their own way whether the hero has enough to leave town, and disagree (16 rule conflicts listed in the review).

## Decision

One supply need. Every rule that asks whether the hero has enough potions, scrolls, ammunition, repairs or gold for where it goes next reads the supply need of that excursion and the one supply count. Town visits execute a town plan solved from the two. A guard may block only on a line the plan cannot close, and names it. Numbers live in profile data with the build's needs in the build profile; no class or skill is named in the planner. (Proposed text of the review, section 9 Q12, verbatim.)

Answers to the review's questions:
- Q1: a guard counts belt PLUS pack, for healing AND mana.
- Q2: keep today's minimums (healing 4, mana 2 or 3, portal 2) as profile defaults.
- Q3: keep the carry line `300 + 50 x clvl`.
- Q4: recovery reads WEALTH (purse plus stash).
- Q5: use the review's short-gold priority table (decision D3, potions before the first scroll; `gold_carry` at position 6).
- Q6: Vault items are never an automatic gold source.
- Q7: chunks declare supplies (a supply profile name and an optional gold number, replacing `noGold`).
- Q8: below-floor potions count only while gold is short, are never bought, and are sold first.
- Q9: survival-model sizing is shadow only until two green runs per class.
- Q10: staged knob `town.plan`: 0 off, 1 shadow, 2 guards and shop, 3 purse, 4 sell and pickup, 5 route.
- Q11: `docs/contracts/town-basics.md` "Bank before the shop" is reworded to "the purse is settled once per visit to planned spend + chunk gold + carry line", in the card that changes the code.
- Q12: this ADR and the glossary terms excursion, supply need, supply count, town plan, verdict.
- Q13: the SME agent reviews the profile numbers once (one card).
- Delivery: 11 cards; stages 1 and 2 first (counter and shadow, then guards and shop); pause for the user's look before purse, sell and pickup, and route.

## Consequences

- Guards, shopping and recovery read one counter; a refusal names its line.
- Chunk data changes shape (`supply`, `gold`); the quest board restock rule is in [ADR 0025](0025-quest-board-lifecycle-stored-runs.md) item 9.
- Measured runs record the `town.plan` stage in the run record.
