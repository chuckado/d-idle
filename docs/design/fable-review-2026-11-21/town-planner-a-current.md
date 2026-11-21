# Town planner, appendix A: the chores today and where their rules disagree

> status: current (design review 2026-11-21, docs only). Main doc: [town-planner.md](town-planner.md). Tree read: `origin/main` at `5b12ce5e7`.
> Tags: V = read in this tree at the cited file:line (code reading, not a run) unless a run id is named; S = secondary (a probe doc's own claim); U = unverified, a judgment or an inference.

## A1. The fixed chore order

One function runs one fixed list. `Town.Chores` (V `internal/town/chores.go:336-503`) builds the step table at `:415-455` and runs it top to bottom; each step is `enabled` (from the `ChorePlan` flags), then `need(snapshot)`, then `run`. A failed step does not stop the later ones (`:405-410`). One exception to the order exists: a repair the player requested moves up behind `sell` (`:456-473`).

| # | step | enabled by | need rule (V file:line) | what it reads |
|---|---|---|---|---|
| 1 | idtome | `Scrolls` and `town.identifyTome` (0) | `needIDTome` | pack |
| 2 | identify | `Identify` | `identifyNeeded` `internal/town/basicsfix.go:53-65` | unidentified items the pickit wants read, and a scroll or tome |
| 3 | bulkid | `Identify` and a bulk identifier | `needBulk` | what the scroll chore left |
| 4 | equip | always | `needEquipFound` | empty slots |
| 5 | sell | `Sell` and `town.sellJunk`, or any player sell mark | `needSell` `chores.go:505-508` -> `sellableIf` `internal/town/shop.go:642-709` | pickit verdicts, marks, pack fullness |
| 6 | ammo | `Ammo` | `ammoTodo` | launcher, stacks |
| 7 | beltbuy | `Potions` | `needBelt` `internal/town/beltbuy.go:20-31` (buys a belt ITEM from `town.beltBuyLevel` 10, and only when the bot decides gear, which is not the default; it is not a potion step) | worn belt, gold |
| 8 | repair | `Repair` | `needRepair` `chores.go:204-224` | durability, estimate, repair gate, purse and stash |
| 9 | belt | `Potions` | `NeedsBeltFill` `internal/town/beltfill.go:93-99` | pack potions the belt plan has room for |
| 10 | bank | `Gold` | `needBank` `basicsfix.go:98-100` | purse above `goldKeep` (`internal/town/stash.go:124-131`) |
| 11 | withdraw | `Gold` and (`Potions` or `Scrolls`) | `needWithdraw` `basicsfix.go:154-157` or `needShopWithdraw` `:212-228` | purse against the shop list |
| 12 | shop | `Potions`, `Scrolls` or antidotes | `needPotions` `chores.go:168-187`, `needScrolls` `:189-202`, `needAntidotes` | belt plan shortfall minus pack, scroll targets |
| 13 | beltafter | `Potions` | `NeedsBeltFill` | as 9 |
| 14 | heal | `Heal` | `needHeal` `chores.go:144-153` | life under 70 percent or mana under 30 |
| 15 | cube | `Cube` | `needCube` | cube jobs |
| 16 | stash | `Stash` or `Gold` | `needStash` `chores.go:510-531` | purse above `goldKeep`, items the pickit wants stashed |
| 17 | tidy | `Stash` or `Sell` | `needTidy` | tidy policy table |

Notes, all V:
- The bot-side adapter forces `Potions`, `Scrolls`, `Gold` and `Ammo` on for every visit (`cmd/d2bot/adapters.go:170-171`), so the plan flags only choose heal, identify, sell, repair and stash. The "plan" is five booleans; it carries no quantities and no reason beyond a text.
- The step named `beltbuy` buys a belt item. The log line "beltbuy (not needed)" therefore says nothing about potions; the potion steps are `belt`, `shop` and `beltafter`.
- NPC grouping is implicit: `vendorStep` (`chores.go:330`) keeps a trade window open across consecutive steps of the same NPC. The walk order between NPCs is the list order: repair vendor (sell, ammo, beltbuy, repair), stash (bank), stash again (withdraw), potion vendor (shop, heal), stash a third time (stash, tidy). The act table lists who does what (`internal/town/acts.go:41-52`); positions come from presets at run time (`acts.go:9`).

## A2. Who decides that a visit is due (three deciders)

| decider | where | healing rule | counts |
|---|---|---|---|
| `resource.TownTripTrigger` | `internal/resource/town.go:50-118` | belt healing `<= resource.tripHPMin` (1) (`:67`) | belt only |
| `Runner.choresDue` | `internal/bot/goal.go:377-` | healing `< bot.resupplyPotions` (2), doubled in town (`:431-434`) | belt healing and rejuvenation (`beltAndInv`, `goal.go:220-244`) |
| `Town.NeedsTrip` | `internal/town/chores.go:38-128` | belt plus pack `< town.tripMinHPPotions` (2) (`:78`); a column under `town.tripBeltColumnPct` (50) with an empty pack (`:103`) | belt and pack, all tiers |

`choresDue` calls the other two and adds its own rule, so one tick can hold three different opinions on "low on healing potions".

## A3. Every number that means "how many do I need"

Healing potions (V at each line; the counts column says what is counted):

| rule | value | counts | where |
|---|---|---|---|
| quest chores step `Need` | `minPotions` 4 | belt + pack, healing + rejuvenation | `internal/task/steps.go:104`, `:122`; filled at `internal/bot/task.go:347-349` |
| Monastery and tower guard | `routeHealing` 4 | same | `internal/task/route.go:169`, `:213-217` |
| boss gates (Tools, Tower, Andariel) | 4 | same | `internal/task/gate.go:177-185`, `:213-225` |
| Blood Raven gate | `brPotions` 4 | same | `gate.go:28`, `:49-62` |
| kit gate | `resource.kitMinHP` 4 | same | `internal/bot/kitgate.go:41` |
| poverty kit | `resource.hpFirst` 3 | belt healing + rejuvenation | `internal/resource/poverty.go:21-27` |
| recovery floor | `resource.recoveryPotionFloor` 2 | belt healing + rejuvenation, adapted by `KitHealing` | `internal/resource/recovery.go:39`, `internal/bot/broke.go:67-72` |
| Den entry and stay | 2 and 1 | belt | `internal/bot/goal.go:250-251` |
| readiness row `potions` | warn 2, fail 0 (Normal) | `Hero.Potions` | `internal/readiness/checks.json:64-69` |
| shop buy target | belt plan capacity x `town.beltFillPct` (100) minus the pack | belt plan columns | `chores.go:174-187` |
| pickup buffer | `loot.potionBuffer` 4 in the pack beyond the belt | pack, all tiers | `internal/loot/loot.go:575-580`, `internal/env/tuning_loot.go:49` |
| sell keep | `town.spareKeep` 4 per item CODE, only when the pack is at least 80 percent full | pack, per code | `shop.go:684`, `:704`, `:760-781` |
| arbiter shadow fact | belt healing | belt | `internal/bot/shadow.go:91` |

Mana potions: guard and gates 2 or 3 (`route.go:171`, `gate.go:182-184`), kit gate `resource.kitMinMP` 3 (`kitgate.go:42-44`), all counted in the BELT ONLY (`task.go:350`, `routeguard.go:107`, `kitgate.go:41`); pickup buffer `loot.potionBufferMana` 2 in the pack.

Portal scrolls: shop target `town.tpScrollTarget` 2 without a tome, a literal 5 with one (`shop.go:529-530`), `town.tpReserve` 2 when poor (`shop.go:540`), guard `routePortals` 2 charges (`route.go:170`), Blood Raven 2 (`gate.go:28`), empty-source restock 2 (`chores.go:857`), pickup stops at `town.looseScrollKeep` 2 beside a tome or at `tpScrollTarget` without one (`loot.go:606-612`), sell sells loose scrolls above `looseScrollKeep` beside a tome (`shop.go:715-727`).

Gold lines:

| line | value | reads | where |
|---|---|---|---|
| purse keep (bank) | `300 + 50 x clvl`, at least the kit (190) | purse | `stash.go:124-131`; contract `docs/contracts/death-rule.md:70` |
| minimal kit | `kitTPGold` 100 + `hpFirst` 3 x `kitPotionGold` 30 = 190 | | `poverty.go:6`, `basicsfix.go:88-91` |
| recovery start / target | 500 / 900 | PURSE only | `recovery.go:62`, `:81`; `broke.go:77` |
| trip needs gold | `resource.tripMinGold` 60 | purse | `chores.go:40` |
| restock free above | `resource.tripGoldFloor` 300 | purse | `shop.go:534-543` |
| optional buys reserve | `town.minGoldReserve` 100; merc revive reserve | purse | `shop.go:569`, `:617`, `:932`; `shop.go:1073` |
| Blood Raven gold | 250 | purse (`World.Gold`) | `gate.go:28`, `:65-70` |
| antidotes count as kit from | 150 | purse | `gate.go:85` |
| chunk gold | a boolean `noGold` | | `internal/questboard/def.go:91-93` |
| shop withdraw target | the shop list, at least the kit | purse vs stash | `basicsfix.go:178-207` |

## A4. The rule pairs that disagree

Each row is a pair (or triple) of rules that cannot both be satisfied, with the evidence. "Structural" means it follows from the code alone.

| id | rule A | rule B | conflict | evidence |
|---|---|---|---|---|
| I1 | chores step `Need`: healing (belt + pack) `< 4` (`steps.go:104`) | shop buys the BELT PLAN shortfall only (`chores.go:174-187`) | Structural. The class plans are `[hp hp hp mp]` (mana need 1) and `[hp hp mp mp]` (mana need 2) below clvl 31 (`internal/resource/potion.go:150-165`). On a 4-slot belt the healing capacity is 3 or 2 (and the bot buys a bigger belt only when it decides gear, A1 row 7). With no spare healing potion in the pack the step runs, buys to 3, is "still needed" and ends as `needs_attention` (`internal/task/executor.go:395`, `:485`). | V structure. V run 20261120-200952-46d2 and -192148-b1a8: "3 healing potions < 4", "healing potions 3/4 with a 3+1 belt" (`docs/probes/provingground-act1more-2026-11-20.md` finding 2). |
| I2 | guard and gates want 4 healing (`route.go:169`, `gate.go:182-184`) | the same shop target | Same mechanism at the guard: `gateReady` inserts a chores step (`gate.go:249`), which is I1. The hero then idles in town with gold. | V countess finding 4 (700 gold, 3 healing, 25 virtual minutes idle, `docs/probes/provingground-countess-2026-11-20.md`). |
| I3 | sell: a full pack keeps `spareKeep` 4 per code; below-floor potions are plain junk | guard counts every healing potion in the pack | The visit that should prepare the hero lowers the count the guard reads. | V countess finding 1 ("28 sold", pack healing 16 to 4) and 2 (Minor potions sold first). |
| I4 | step 5 `sell` runs before step 9 `belt` | belt fill would have belted some of those potions | Order: potions leave the pack before the fill looks at them. | V order `chores.go:429`, `:437`. U how often it costs a potion in play. |
| I5 | pickup: keep up to 4 healing in the pack beyond the belt (`loot.go:575-580`) | shop: never buys a pack spare; sell: the keep line applies only at 80 percent pack | Three targets for one stock: the looter builds a buffer the shop will not rebuild and the sell chore may remove. | V code. |
| I6 | pickup refuses potions under the tier floor (`loot.go:554-558`) | guard, shop subtraction and trip rules count them (`belt.PackCount`, `internal/town/belt/belt.go:253-264`); belt fill belts them | A below-floor potion is worthless to one rule and a full kit item to four. | V code. |
| I7 | kit gate: mana `>= 3` in the BELT (`kitgate.go:41-44`); Andariel gate 3 (`gate.go:184`) | mana-need-2 plan has 2 mana columns: capacity 2 on a 4-slot belt | Structural: unsatisfiable until a belt with more rows is worn; the kit trip repeats until `resource.kitTripCap` (2) and is then skipped. | V structure. V countess finding 5 ("2 mana potions < 3"). |
| I8 | healing is counted belt + pack (`task.go:349`) | mana is counted belt only (`task.go:350`) | Same guard, two counting rules. | V code. |
| I9 | bank: purse down to `300 + 50 x clvl` (350 to 450 at clvl 1 to 3) | recovery starts while the PURSE is under 500 and the missing kit costs more than the purse (`recovery.go:62-71`, `broke.go:77`) | The bank leaves 350 to 450 gold at clvl 1 to 3 and the shop spends from that; as soon as the missing kit (up to 250 gold, `internal/resource/poverty.go:103-109`) costs more than the purse, recovery starts and holds the quest, whatever the stash holds. | V code. V Den finding 1 (`docs/probes/provingground-act1next-2026-11-20.md:36`: setup gold banked, chunk held 45 virtual minutes by the purse recovery). U: the exact purse sequence of that run was not re-derived here. |
| I10 | bank before the shop (contract `docs/contracts/town-basics.md:43-44`) | quest chunk needs a purse (Blood Raven 250, antidotes from 150, the chunk's own resupply) | The bank does not know what the next excursion will spend; the only link is the boolean `noGold`. | V act1more finding 1 (6900 of 8000 banked, chunk parked 28 minutes). U: the second run with the purse kept still parked on I1, so how much of finding 1 is the bank and how much is I1 is not separated. |
| I11 | withdraw fires for the SHOP list (`basicsfix.go:212-228`) | nothing withdraws for a gate's gold line, a chunk, the merc or ammo; repair withdraws for itself only when requested (`chores.go:742-755`) | Four separate "who may take gold out" rules. | V code. |
| I12 | shop holds 2 scrolls without a tome, 5 with one | guard wants 2 charges; pickup and sell cap loose scrolls at 2 | Fixed by a cap knob (#1013), but the three numbers still live in three places and the 5 is a literal. | V code; S the hoard (22 loose scrolls) from the scenario note in `tools/provingground/scenarios/tphoard.json`. |
| I13 | `BackToField` no longer casts in town (`internal/town/trip.go:398-399`) | `ToTown` checks "in town" once, before its three tries (`trip.go:58-80`) | A try that arrived without being seen can cast again in town. | V code; U whether this is the path behind "town portal did not appear" in town. |
| I14 | three visit deciders (A2) with thresholds 1, 2 and 2 on three different counts | | The trip can be raised by one and judged futile by another. | V code. |
| I15 | `w.Caster` and `w.BRHigh` are class lists in Go (`task.go:357`, `:359`; `routeguard.go:108`) | the mana need is already a trait (`resource.ManaNeed`, itself a class switch, `internal/resource/types.go:172-180`) | Two class tables decide who needs mana; the build is not asked. Against the build-flexibility rule. | V code. |
| I16 | potion kind by the bot catalog (`internal/bot/resource.go:46-64`) | potion kind by the data tables (`belt.PotionKind`) | Two classifiers feed the two sides of every rule above. | V two exist; U whether they ever disagree. |

## A5. What the list says

1. No rule knows the next excursion. Every number in A3 is a constant or a clvl formula. The guard is the only place where "where am I going" meets "what do I carry", and it can only say no.
2. "Need" (the guard, the chores step) and "target" (the shop, the pickup, the sell keep) are written separately, in different units (belt + pack against belt plan), so a need can be unsatisfiable by the step that is supposed to meet it (I1, I2, I7).
3. Gold is judged at the purse by some rules and moved by others that do not know those rules exist (I9 to I11).
4. A blocked precondition has no path to its own fix: a kit shortfall that the chores cannot close becomes `needs_attention`, not a restock, a withdrawal or a gold farm.

## A6. Dead ends and limits of this reading

- No run was made for this review. Run ids are quoted from the probe docs; the setups there are CHEATED (level, gold and potions by SETUP verbs), so they are evidence of mechanics only, never of balance.
- U: whether finding 2 of act1more is I1 alone. The setup gave 8 Minor healing potions in the pack at clvl 22, which the sell chore treats as junk (I3, I6); after that sale the belt capacity of 3 explains "3 < 4" exactly, but the log was not re-read here.
- Not read: `internal/town/rebuy.go` (poverty rebuy), `ammo.go`, `merc.go`, `internal/invdesk` (player item verbs), `internal/goldfarm`. They are named in the design only as consumers.
