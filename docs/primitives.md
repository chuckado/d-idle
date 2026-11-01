# Primitives: the solid list of basic game actions and how to prove each live (agent PRIMLIST, 2026-10-31)

> status: current. Offline write-up (no game, no d2d slot). Feeds the live drill suite in `openspec/changes/m1-liveproof/` and its spec delta (capability `primitives`). Review of this list: `docs/primitives-review.md` (agent PRIMREVIEW).

Why: the offline tests use fakes that encode our assumptions, so they pass while the live game differs. A **primitive** is an atomic, verifiable game action or state handling. Each row says what *correct* is, as an observable end state, so a scripted drill on a hidden d2d with a throwaway character can pass or fail it. Policy work (trips, kit, gold floor) sits on these.

Tags. Criticality: **S** death or lethal state, **A** lost items or gold, **B** time drain, **C** stuck state. High = S or A. Evidence: **L** proven live (doc cited), **L~** proven in part, **O** offline fakes only, **U** none. Code paths are under `internal/`. Live counts = lines in the three long navsoak logs on the worker host (navsoak, navsoak2, navsoak3; navsoak3 may still be growing; read-only streaming greps, 2026-10-31); `-` = no log line exists for it (itself a gap). SME = the d2-sme consult of this job (veteran memory, not sourced; marked where unsure).

## The list (36)

| id | primitive (group) | precondition -> correct end state | top failure modes (SME) | crit | code | ev | live counts | DRILL (setup -> pass) |
|---|---|---|---|---|---|---|---|---|
| P1 | Belt fill from pack (inv/belt) | town or field, potions in pack, free belt slots -> each potion in a compatible column, count in pack drops by the same | shift-click ignored; column mixes types; fill goes to wrong slot | A | town/beltfill.go FillBelt, bot/resource.go beltFillStep, d2dctl click modifiers | L~ (V `belt put` lab, d2d-fork.md; shift path on items U, beltfill.md); drill PASS 2026-10-31 (D1.P1 belt fill, D1.P1s shift-click; docs/probes/drills/2026-10-31-d1rerun.txt) | 359 fills, 26 moved 0 (all `failed:1`) | cheat gold, buy 6 HP + 4 MP into pack, belt empty, run FillBelt -> belt grid HP/MP per plan, pack potions = 0, 3 runs, 0 on cursor |
| P2 | Belt layout by belt type (inv/belt) | belt equipped -> rows and columns read correctly (1-4 rows x 4 cols) | row table guessed; slot index vs column mixed up | S | town/belt/plan.json, resource/potion.go PlanFor, d2dsrc item dump | U (probes 6: row table U); drill PASS 2026-10-31 (D1.P2, 5 belt types) | - | equip each belt type via debug equip, read grid -> rows match Belts data, 4 types |
| P3 | Drink potion by belt column (consumable) | life below threshold, column has potion -> life up within 2 s, that column's count -1, front potion consumed | key bound to empty column; panel open eats the key; wrong potion type first | S | d2dctl Drink (cmd belt), d2dcombat input.go, backend-contract "belt keys" | L~ (spike, d2d-probes 6: drinks verified; heal rates not run); drill PASS 2026-10-31 (D1.P3; merc drink BLOCKED) | 4391 skips "no healing potion in belt", 487 "no rule asks", 344 crowd-retreat first | belt 4 cols, set life low via damage/cheat heal off, drink each column -> life delta > 0, count -1, repeat per column and merc |
| P4 | Pack fit and make room (inv) | item w x h, pack partly full -> placed or MakeRoom sells junk until it fits | fragmentation by 1x1; large items refused; cell math off | A | town/room.go MakeRoom, loot/pickup.go, mem inventory grid | O | 360 "pack full, item left", 131 "no room" | fill pack with 1x1 charms + 2x4 free columns, pick 2x3 armour -> placed or sell visit then placed, no loop |
| P5 | Stash items and gold (stash) | stash open -> item or gold moved, shows in stash grid/gold; withdraw returns it | click before panel ready; deposits all gold; stash full; cap | A | town/stash.go, stashfull.go, d2dctl OpenStash MoveGold | L~ (deposit geometry V d2d-town.md gap 3; cap 2.5M U) | 58 "partial gold deposit needs a controller that can type" | gold 1500 -> deposit 1000, withdraw 300, move 1 item in and out; stash gold + purse sums constant |
| P6 | Cube transmute (inv) | cube in pack, recipe items -> output item, inputs gone, cube kept | wrong inputs; cube lost; recipe order | B | town/cube.go, cubejobs.go | O | - | 3 chipped gems of one type -> 1 flawed gem, cube still held |
| P7 | Pickup potion with belt priority (pickup) | potion on ground, belt slot free -> potion goes to belt (or pack then belted), never skipped while belt is empty | "already in pack" skip with empty belt; pickup loop; leaves it | S | loot/pickup.go rules (keep: potion already in the pack), d2dloot | L~ (pickup works; belt priority U); drill PASS 2026-10-31 (D1.P7, D1.P7a; picked 2) | 939 "potion already in the pack" (471 whole-line) while belt empty | drop 2 HP potions at feet with belt empty and pack holding 3 -> belt gets them (directly or after FillBelt) within 10 s |
| P8 | Pickup gold, full pack (pickup) | gold pile, pack full -> purse up by pile, no slot needed | full-pack check blocks gold; small pile ignored; carry cap | A | loot gold path, d2dloot | O | 92 "small gold pile" skipped | fill pack, drop 300 gold -> purse +300; at purse cap pile stays (SME cap 10k x level, U) |
| P9 | Pickup item by rule (pickup) | item in range, rule keep -> in pack with right size; rule drop -> left alone, no retry | pack-full spam; clicking vanished items; loop on item that does not fit | A | loot/loot.go pickup.go rules.go | L~ (d2d-loot.md wired) | 1000 "plain or inferior" filtered, 18483 "recovery: loot +0 items" | drop 4 mixed items (keep, junk, no-fit, gem) -> exactly the keepers picked, each logged once |
| P10 | Cast TP and enter portal to town (consumable/travel) | scroll or tome in pack -> portal object, area id = town, cursor empty | cast with panel open; tome at 0; walking in before portal exists | S | town/trip.go ToTown, d2dctl CastTP EnterPortal, d2dtown/adapter.go Interact | L~ (cast V; walk-in = click sprite V d2d-town.md gap 5; adapter route UNVERIFIED); drills 2026-10-31: verb PASS (D2.a/a2/b/c/e/h), walk-in FAIL (D2.t, D2.g), cut cast BLOCKED (D2.d); docs/probes/drills/2026-10-31-d2.txt | 2683 "portal attempt failed", 710 "no portal opened", 376 trips abandoned | in Blood Moor, scroll + tome variants, 10 casts -> area id 1 (town) each, success >= 9/10 |
| P11 | Return through portal to field (travel) | town, open portal -> back in the same field area, portal still there | takes wrong portal; "did not lead to the level" | S | town/trip.go, d2dctl EnterPortal | L~; drill BLOCKED 2026-10-31 (D2.f: no working portal walk-in in the lab) | 176 "portal did not lead", 207 "no portal in the unit table", 178 ok | after P10, return -> area id = source, position near portal, 10/10 |
| P12 | Buy TP scroll/tome and fill tome (consumable) | Akara/Charsi window, gold -> scrolls in tome up to 20, gold delta = price | scrolls loose beside a tome; tome unbought; price assumed | A | town/shop.go buyFirstTP buyScrolls | L~ (scroll V; tome U, price 250 vs 450 knob) | - | gold 600, buy tome + fill -> tome qty = 20 or gold spent, delta logged |
| P13 | Identify item (consumable) | ID scroll/tome + unid item -> item identified, scroll -1, cursor empty | scroll left on cursor; IDs junk | B | town/identify.go | O | - | unid ring, 1 ID scroll -> identified props read, scroll count -1 |
| P14 | Approach NPC and open window (NPC) | town, NPC index -> window open (vendor tab or dialog list), no stale panel | stops at object on the way; click before ready; panel already open | A | town/npc.go openNPC, d2dctl ApproachNPC OpenNPCMenu PickMenuLine, d2dtown | L~ (Akara, Charsi, Gheed, Kashya V; Warriv U) | 2233 opens ok, 200 "panel open before interact" | from stash corner open each of 7 Act 1 NPCs -> window flag true in <= 15 s, close leaves UI clean |
| P15 | Vendor buy and sell (NPC) | window open -> bought item in pack, gold delta = price; sold item gone, delta = price | sell price assumed (Cost/4 vs 50 %); wrong tab; sells quest or worn item | A | town/shop.go, d2dctl BuyItem SellItem | L~ (scrolls/potions 50 % V d2d-fidelity.md:196; items U) | - | sell unid and known ring, buy potion, scroll -> gold deltas recorded, nothing equipped sold |
| P16 | Repair (NPC) | worn item below max, gold -> durability max, gold down by price | not confirmed (button geometry); repairs at poor gold; reruns every visit | A | town/chores.go repair, d2dctl RepairAll Durability | L~ (3 of 4 ok; failures seen) | 36 "repair not confirmed" | cheat durability to 30 %, gold 200 -> durability 100 %, price logged; gold 0 -> one log, no retry |
| P17 | Heal at NPC (NPC) | hurt, healer in town -> life = max after dialog, no potion used | skips and drinks; wrong dialog line | S | town/chores.go heal, merc.go | L~ | 208 "town heal unavailable" | set life 30 % in town -> full after NPC visit, potions unchanged |
| P18 | Merc hire, revive, feed (NPC) | gold, Kashya or Act 2/5 hirer -> merc alive, owned; dead merc revived; potions fed | blind hire; no revive gold; never feeds | A | town/merc.go, merc/policy.go, d2dmerc, task/claim.go | L~ (free Rogue V probes 1; hire/revive dialog U) | - | claim reward, kill-less revive test via cheat death of merc if possible -> merc alive flag |
| P19 | Gamble (NPC) | gold above floor -> never when poor; buy one slot item, gold delta | spends potion money | B | none (class listed only) | U | - | not built: drill is "never opens gamble tab with gold < floor" (log grep) |
| P20 | Quest dialog and reward (NPC) | quest NPC -> dialog closed, reward taken once, quest flag set | re-clicks finished NPC; dialog left open; misses reward | A | quest/, questmem/, task/ | L~ (Kashya, Akara menus V probes 5) | - | Akara skill-point reward at clvl 8+ -> unspent skill points +1, flag set, dialog closed |
| P21 | Travel NPC / act change (travel) | quest done -> new act area id, town of that act | wrong menu row; Act 4 pillar portal; Tyrael | B | town/acts.go, d2dtown | U (Warriv U) | - | Warriv with Andariel flag via cheat quest -> area = Lut Gholein |
| P22 | Walk to a point (move) | reachable target -> within 1.5 cells before timeout | clicks on walls; no progress check; run off | C | nav/, d2dinput controller, d2dctl CmdMove | L (soaks run) | 9778 "combat move failed" | 10 targets across Blood Moor -> arrive within 1.5 cells, 9/10 (run toggle after a fresh game: [fixrun.md](fixrun.md), live proof pending; a goal snapped onto the player is already-there, not a failed walk: [fixemptyplan.md](fixemptyplan.md), live proof pending) |
| P23 | Waypoint: activate and choose row (travel) | adjacent to waypoint -> panel, act tab and row, new area id = target | wrong tab; not activated; click mid-animation; 10 s refusal after level change | S | bot/waypoints.go, d2dctl OpenWaypoint ChooseWaypoint | L~ (panel opens V; row choice U) | 220 "nav waypoint failed", 127+71 "waypoint unavailable" in no-portal trips | cheat-activate 3 waypoints, go to each -> area id correct, 3/3 |

P23 verdict (fixwp, 2026-10-31): live PASS on d2d for act 1 rows 1 and 8 and act 2 rows 6 and 8, the inactive case and the 10 s second use (docs/fixwp.md, drills DW.*); cross-act rows PASS (DRILLS4, `debug act N`: DW.a3, DW.a3b, DW.a6 to DW.a11, acts 1 to 5 both ways, each twice identical, 2026-10-31-dw3.txt; tab rects measured, act 4 at x 304 and act 5 at 368; BLOCKED only: `knownWaypointPos` lacks towns 75, 103, 109, so a never-visited level fails `none on level N`). The 295 `unit not found` failures were W1 (class -1), fixed in d2dsrc.
| P24 | Level transition (move) | exit/stairs/door -> new area id, position on the arrival side | tile offset; re-entering same area; link wait | C | nav/, d2dworld, d2dexit | L (docs/d2d-exit.md Blood Moor in 32 s) | 5746 "nav stuck" | take 5 exits in Act 1 outdoors -> area id changes, no ErrNoExit |
| P25 | Stuck recovery (move) | no progress N s -> different click, then waypoint or TP, position moves | endless retry of the same click; marks good edge bad | C | nav/ failsafe ladder | L~ (corpus WedgePocket offline) | 5746 stuck, 69 "failsafe ladder", 113934 "target not reachable" | pin bot in a pocket (portal wedge) -> out within 60 s |
| P26 | Skill bind and cast (combat) | skills allocated -> BindSkills count, cast at cell/unit changes mana and spawns effect | popup clicks; wrong active skill; no mana | S | d2dcombat/rig.go BindSkills, d2dctl CastUnit CastCell SetSkill | O (BindSkills UNVERIFIED, backend-contract) | - | clvl 6 Sorc with 2 attack skills -> bind 2, cast each at a target, mana delta > 0, hit logged |
| P27 | Target selection (combat) | mobs present -> nearest reachable threat, champions avoided at low level | attacks across walls; loops on unreachable | S | combat/, d2dcombat meter | L~ (soaks), O (corpus) | 342 flee aborts | 3 mobs one behind wall -> never targets the walled one |
| P28 | Life/mana monitor and thresholds (combat) | every tick state read -> chicken or drink at knob thresholds, no chug loop | stale read; hysteresis; threshold off the damage curve | S | resource/potion.go, bot/ chicken, d2dcombat Meter | L~ (Meter totals UNVERIFIED; offline in fixmeter: `Attach` builds the Meter, [fixmeter.md](fixmeter.md)); [fixportalmelee.md](fixportalmelee.md): no cast under melee | 5340 "chicken", 4805 "xp at risk" | scripted damage -> drink at thresholds, damage totals match life drop |
| P29 | Retreat and escape use (combat) | swarm or boss -> breaks line, heals, or TPs; never runs into dead end | runs through pack; dead end; stuck path | S | combat/ retreat, bot/ | L~ | 510 "resting before pack", 50 "no way to heal" | pack of 6 at clvl 5 -> retreat then TP, alive, kit used logged |
| P30 | Stat points (build) | level up -> points spent per plan, 0 unspent | leaves points hours; wrong stat | B | leveler/ | L~ | 58 "equip retry" (equip not stat) | cheat level 12 -> unspent 0, stats match plan |
| P31 | Skill points (build) | level up -> skill levels per build order, 0 unspent | wrong order; tab click | B | leveler/ | L~ | - | cheat level 12 -> skill levels match, 0 unspent |
| P32 | Equip and swap gear (build) | item in pack, reqs met -> in slot, old item in pack, cursor empty | old item on cursor; reqs ignored | A | leveler/equip.go, gear/ | L~ | 58 "equip retry" | give better armour via vendor -> worn, old one in pack, cursor empty |
| P33 | New game, save and exit, rejoin (life) | ->  character in town, save written, no panel | kills process mid-save; wrong difficulty | C | d2dinput/lifecycle.go | L (5091 Save and Exit re-enter ok) | 5095 "channel not answering" after exit (expected) | 10 leave/enter cycles -> in town, items equal |
| P34 | Death and corpse recovery (life) | dead -> town naked, walk to corpse, pick, re-equip | re-die; abandon gear; no hotkey rebind | S | bot/death.go | L (probes 3) | 4805 "xp at risk" | kill with god off -> gear back, corpse unit gone, binds restored |
| P35 | UI hygiene (UI) | every tick: no panel, popup, dialog, empty cursor | cursor item after failed place; stale panel; dialog open | S | town/panel, mem UI flags, d2dctl UI | L~ (UI flags verb) | 200 "panel open before interact" | run before and after every drill -> UI clean flag true, cursor empty |
| P36 | Trip trigger worth it (town policy) | belt/scroll/repair/sell needs -> a trip only if plan >= floor or belt empty; trip fills the belt | many tiny trips; same empty belt after trip | B | bot/goal.go, town/chores.go, env tuning_town | O (A/B on d2d: goldfloor) | 2234 trips, 269 "trip done", 376 abandoned | 8 seeds soak: trips < 30 gold <= 35 %, belt-empty minutes < 10 % (gate) |

## Live drill verdicts (m1-liveproof, 2026-10-31, hidden d2d lab build with the click modifier, throwaway Sorceress, each drill twice)
| prim | drill | verdict | evidence |
|---|---|---|---|
| P1 | D1.P1 | PASS | 4 hp1 + 4 mp1 in the pack, Sash: fill moved 8, hp in hp columns, mp in mp columns, pack empty (fix-belt F1/F3/F6/F7/F8) |
| P1 | D1.P1s | PASS | `click x y shift` on a pack hp1 belts exactly one per click (box 0, then box 4 above it); the bot does not use this verb on d2d (it uses `cmd belt put`), the drill drives it |
| P2 | D1.P2 | PASS | reader boxes = engine boxes for 5 belt types (8/8/12/12/16), overflow stays in the pack |
| P3 | D1.P3 | PASS | each drink removes one potion of the asked kind from the right column; empty kind returns the error; D1.P3m (merc) BLOCKED: no merc cheat (`debug merc hire\|kill\|revive` exists in the lab binary as of 2026-10-31, docs/forkcheats.md: unblocked, rerun pending) |
| P7 | D1.P7a, D1.P7 | PASS | engine auto-belts a ground potion; pickit takes both and reports picked 2 (arrival by held count works) |
| P10 | fixrecast | OFFLINE | armors (buff upkeep) are recast by Skills.txt duration minus a margin and the Frozen Armor state, not a flat 150 s; live recipe in [fixrecast.md](fixrecast.md) |
| P10 | fixportalmelee | OFFLINE | a cast under 3+ adjacent attackers is cut (soak 2 death): Decide/AbortFight choose potion or chicken instead; live recipe in [fixportalmelee.md](fixportalmelee.md) |
| P10 | D2.a | PASS | portal verb from the field: 1 right click, 32 virtual ticks from the call (click at +22, the fixed panel sleeps), one portal object, one source spent |
| P10 | D2.a2 | PASS | raw cast: portal object 8 ticks after the right click, 3 casts, identical (trace: about 7) |
| P10 | D2.b | PASS | empty tome, no scroll: `ErrNoPortal`, 0 clicks, 0 keys, 0 ticks (fix-portal F4) |
| P10 | D2.c | PASS | empty tome + loose scroll: the scroll is cast, the tome untouched (a scroll picked up while a tome is in the pack goes INTO the tome: engine rule, V) |
| P10 | D2.e | PASS | busy player (mode SC): the verb clicks at tick +37 vs +22 idle, one click, portal opens (fix-portal2 idle wait) |
| P10 | D2.h | PASS | one-pair rule: second cast 12 cells away closes the first, one portal object |
| P10 | D2.t | PASS (was FAIL) | bot ToTown with the portal open next to the player arrives in town in 5 ticks, no scroll spent (enterportal: `cmd interact -2000-k`, docs/d2d-portal-entry.md); the first run's FAIL was W1 (docs/probes/drills/README.md) |
| P10 | D2.g | PASS (was FAIL) | control: d2dctl `EnterPortal` now sends `cmd interact -2000-k` (enterportal); the sprite click it used walked in the lockstep lab |
| P10 | D2.d, D2.d3 | PASS | `debug hit 5` in the same tick as the verb's click: the click is dropped, source unchanged, the retry opens the portal (2 clicks, 1 scroll left); 3 ticks into the cast: the cast is cut with the scroll spent (`the cast was cut (source spent)`), the retry takes the next scroll (2 clicks, 0 left); each twice, identical (2026-10-31-d2d.txt, DRILLS3) |
| P11 | D2.f | PASS (was BLOCKED) | round trip with the working walk-in, arrival 6 subtiles from the origin (enterportal, docs/d2d-portal-entry.md) |
DRILLS3 rows (2026-10-31, lab binary with the test cheats, throwaway Sorceress, each drill twice, identical; files `2026-10-31-<group>.txt/.json` in docs/probes/drills, audit in hover-audit.md):
| prim | drill | verdict | evidence |
|---|---|---|---|
| P12 | D3.tp | PASS | scrolls chore at gold 1000 with no tome: tome and 5 scrolls bought (the tome ends with 5 charges, not 20: the bot's fill target), gold 1000 -> 200 = the report's 800 |
| P14 | D3.n | PASS | Akara, Charsi, Gheed, Kashya, Warriv each open a dialog from the start point (22 to 258 ticks), close leaves the UI clean; Cain and the stash/waypoint rows are not in this drill (stash: D3.s) |
| P15 | D3.b, D3.b2 | PASS | shop chore with an empty belt: 2 hp1 + 2 mp1 + 1 scroll for 280 gold = the report, belt 0 -> 4; sell an identify scroll to Akara: +40 gold, scroll gone (a potion sells for 15 = half of 30) |
| P16 | D3.r, D3.r0, D3.r1, D3.rt | PASS (price table); poor purse FAIL pre-fix-bank | `debug durability all 1`: repair at 5 percent restores 20/20 for 19 gold; at gold 0 one failed attempt (`repair not confirmed`, 101 ticks: needRepair does not check gold, chores.go:178) then the latch skips. DRILLS4: price table D3.rt, 13 bases x 2 losses, proportional to durability lost, e.g. Full Plate 69 lost 5814 gold, Short Sword 23 lost 8, Staff 19 lost 19 (docs/probes/drills/README.md, 2026-10-31-repair.txt); gold 5 on the pre-fix main still opens the smith once (D3.r1) |
| P17 | D3.h | PASS | life 12/40 -> 40/40 at Akara, pack and gold unchanged |
| P5 | D3.s, D3.s2 | FAIL | stash opens and an item goes in and out, but BOTH gold deposits fail: the whole purse (`gold deposit of 1500 not confirmed`) and a partial one (`partial gold deposit needs a controller that can type`); the control `d2dctl MoveGold` deposits exactly. Cause: d2dtown/adapter.go:253 `hoverFirst` hides `TypeChar`, and the d2d prompt starts empty so even a full deposit needs digits (town/stash.go:196-215) |
| P13 | D3.i | BLOCKED | `debug unid` unidentifies plain pack items; the judge selects only magic or better: needs an unidentified magic item cheat (`debug give <code> pack q4 unid` exists in the lab binary as of 2026-10-31, docs/forkcheats.md: unblocked, rerun pending) |
| U1 | D3.g0 | PASS | starting gold and stash 0/0 on a fresh character (classes in 2026-10-31-d3g0*.txt); unidentified sell prices not measurable (no cheat; `debug give <code> pack q4 unid` exists as of 2026-10-31) |
| P9 | D4.a | PASS | cap (equip rule), chipped amethyst (sell rule) picked, ring (plain) and the potion (already in the pack) left; report = end state, a rerun picks nothing |
| P4 | D4.b, D4.c | PASS | pack full of rings: a gem is left once per call (no loop); MakeRoom at Gheed sold 12 rings (10000 gold) and freed a 2x2 hole |
| P8 | D4.g, D4.cap | PASS | gold pile into a full pack taken; the purse caps at 10000 at clvl 1 (a pile at 9990 leaves the pile): the SME's 10000 x clvl is V for level 1 |
| P6 | D4.x | PASS | cube: 3 chipped amethysts -> 1 flawed, cube kept |
| P26 | D5.a | PASS | `debug points` learns 2 skills, BindSkills binds 2, a Fire Bolt cast costs 3 mana (hit effect not checked: no target; `debug spawn` exists in the lab binary as of 2026-10-31, docs/forkcheats.md: the hit effect is unblocked, rerun pending) |
| P28 | D5.b | PASS | combat.Fighter.Safety at 100, 60 percent none; 50, 35, 22 percent drink-life; no mana drink without a fight; mana in a fight and Meter totals BLOCKED (no monster; `debug spawn` exists as of 2026-10-31: unblocked, rerun pending) |
| P27, P29 | D5.c, D5.d | BLOCKED | need a monster (`debug spawn <MonStats id> [n]` exists in the lab binary as of 2026-10-31, docs/forkcheats.md: unblocked, rerun pending) |
| P22 | D6.a, D6.b | PASS | 10 of 10 navigator walks arrive within 1.5 cells in the camp and in Blood Moor (0 to 1.0 cells) |
| P24 | D6.c | PASS | 5 of 5 transitions Blood Moor, Cold Plains, camp by ToArea |
| P25 | D6.d | BLOCKED | no way to pin the character in a pocket (`debug pin on [r]\|off` exists in the lab binary as of 2026-10-31, docs/forkcheats.md: unblocked, rerun pending) |
| P30, P31 | D7.a, D7.b | PASS | level 12 by cheat with 55 stat and 11 skill points: the leveler spends all (66 actions), 0 unspent |
| P32 | D7.c | PASS | a cap in the pack is worn by the leveler, cursor empty |
| P18 | D8.a | PASS | merc hired and killed by cheat: the bot's Reviver through Kashya's menu brings it back; no gold was taken (U whether free revive is a rule at merc level 1) |
| P3 (merc) | D8.m | BLOCKED | a hired merc stands at full life and nothing hurts it: missing `debug merc hurt` (exists in the lab binary as of 2026-10-31, docs/forkcheats.md: unblocked, rerun pending) |
| P18 (claim), P20, P21, P19 | D8.c, D8.q, D8.w, D8.g | BLOCKED | `debug quest` sets bit 0 only (Blood Raven stays at 1, Den at 1, Andariel 0); the act hop is package main; no gamble code (`debug quest list\|set\|bit\|join` exists in the lab binary as of 2026-10-31, docs/forkcheats.md: the record states are unblocked, rerun pending; the act hop and gamble stay bot gaps) |
| P34 | D9.a | PASS | `debug die` in Blood Moor, the bot's Respawn lands in town with nothing worn, back in the field `cmd interact -3000` returns the staff and the corpse is gone |
| P33 | D9.b | PASS | 10 leave and enter cycles by the bot's lifecycle, in town with a clean UI each time (about 3 s each) |
| P36 | D9.c | BLOCKED | a soak, not a drill |
| P36/Gate 1 fault | D10 | FAIL on the restarts counter only (restart 1.5 s, loop resumed, saved XP not lower, no leftovers, P35 clean; counter +0, finding F-D10-1) | d2d killed with SIGKILL under the managed bot (docs/probes/drills/README.md DRILLS4, 2026-10-31-d10.txt) |
| audit | H.pause, H.obj, H.ground, H.monster | FAIL, FAIL, PASS, BLOCKED | hover blind-spot audit: field objects and the merc claim and revive path are at risk (hover-audit.md); H.monster needs a monster: `debug spawn` exists as of 2026-10-31 (docs/forkcheats.md), rerun pending |

DRILLS5 rows (2026-10-31, newest cheats build, throwaway Sorceress, each drill twice; files `2026-10-31-drills5-*.txt/.json`, README section DRILLS5). These supersede the BLOCKED and FAIL rows above for the same drill ids:
| prim | drill | verdict | evidence |
|---|---|---|---|
| P5 | D3.s, D3.s2 | PASS (was FAIL) | full deposit 1500, withdraw 300: purse 300 stash 1200; partial 1000 of 1500: purse 500 stash 1000 (fix-bank) |
| P16 | D3.r, D3.r0, D3.r1 | PASS | estimate 19 = paid 19; gold 0 and 5: `repair skipped`, 0 ticks, no smith opened (fix-bank, fix-price) |
| P15 | D3.b2 | PASS | sell price 40 = the price formula 40 |
| P13 | D3.i, D3.u | PASS (was BLOCKED) | `give rin pack q4 unid`: identified by the chore (scroll -1); unidentified magic Cap sells for 32 = cost/2 of the base |
| P1, P7 | D1.P1b, D1.P7b | PASS | 6 healing + 4 mana in a sash: 4 + 4 belted in the planned columns, no `marked full`; healing in 3 columns and one on the ground: column 3 free, pickit leaves it (fix-belt2, fix-pickup) |
| P23 | DW.m1, DW.m2, DW.a12, DW.a9 to a11 alone | PASS | mouth trigger: 1 and 0 bounces, town in 497 and 554 ticks; waypoint touched in passing handed over; towns 75, 103, 109 unvisited: no `none on level N` (fix-nav) |
| P36/Gate 1 | D10 | PASS (was FAIL on restarts) | restarts delta exactly 1 in both runs, back in game within 8 s, xpRisk after the kill 108 = unsaved XP only (fix-restart) |
| P27 | D5.dodge | PASS | Fallen Shamans: 8 dodges, 1.00 cast commands per cast over 17 casts; other runs 4 to 17 dodges (fix-missile, fix-missile2) |
| P27, P29 | D5.p27, D5.p29 | PASS (recorded; was BLOCKED) | pack with a champion at clvl 1: kills and ignores recorded; life 9 of 40: Safety decides chicken (not run), with potions drink-rejuv |
| P26, P28 | D5.p28 | PASS (was BLOCKED) | rich character, 20 zombies: 2 mana drinks, 12 kills; the poor control drinks none (poor rule, finding F3); Meter totals still BLOCKED (`Attach` builds no Meter) |
| P3 (merc) | D8.m | PASS (was BLOCKED) | `debug merc hurt 20`, the merc belt verb: life 20 to 32, one potion |
| P18 | D8.c | PASS (was BLOCKED) | `debug quest set 2 0x0002`, the bot's ClaimRogue under the game-time Pause: Rogue granted (one run per character) |
| P20, P21 | D8.q, D8.w | PARTIAL PASS (was BLOCKED) | quest records by `debug quest set`: Akara +1 skill point, Warriv `go east` to area 40; the bot's own talk nodes are package main (UNVERIFIED as the bot's path) |
| P25 | D6.d | FAIL (was BLOCKED) | pinned: 6 ladder rungs in 66 s, 0.4 cells moved, `ErrStuck`; Escape (town portal) only on the exit path (finding F2) |
| audit | H.obj, H.obj.chest, H.obj.shrine, H.obj.cage | PASS, PASS, PASS, PARTIAL | the bot's Interact operates the waypoint, a chest (class 139) and a shrine (class 81); the cage (class 26) is listed and accepted, effect needs the Cain quest state |
| audit | H.pause | FAIL (finding F1) | game-time deadline works; the approach to Kashya from the camp arrival cell stalls and takes 16 to 38 s against 15 to 20 s timeouts |

## Top live failure messages (normalised; navsoak, navsoak2, navsoak3)
`area gate refused the walk` 171k; `leveling target not reachable` 114k; `recovery: loot +0 items` 18.5k (every recovery line); `combat move failed` 9.8k; `nav stuck` 5.7k; `chicken` 5.3k; `no healing potion in the belt` (drink skipped) 4.4k; `town portal attempt failed` 2.7k; `potion already in the pack` 939; `no portal opened` 710; `town trip abandoned` 376; `pack full, item left on the ground` 360; `action failed` 798 (127+ "to town: no waypoint on level"); `repair not confirmed` 36. Area gate and wedge lines are nav-soak artefacts (banned areas), not primitives.

## SME alignment
Agreement: the 36 ranks match; belt hotkey drinks the front potion of a column (our spec); tome holds 20; stash 6x8, pack 10x4, cube 3x4; first drills should be belt, potion pickup, gold, trips.
Disagreement or correction (SME vs us or code): (1) SME says the game auto-belts a ground potion; our `potion already in the pack` keep-rule assumes the opposite order and skips them: UNSURE on both sides, P7 decides. (2) SME gives the gold carry limit 10,000 x clvl (UNSURE) and stash cap 2.5M; our code has only the stash cap (UNVERIFIED) and no purse cap: P8 and P5. (3) SME thinks A5 mercs may need Nightmare (UNSURE); we plan Normal: P18 checks. (4) SME did not list the interaction-order rule (cursor item blocks right-click on potions), P35 covers it. (5) SME rank puts merc and escape first; the logs put belt and portal first. UNSURE per SME: shift/ctrl-click on pack, stash, vendor in 1.14d; vendor refresh; repair/sell formulas; belt rows beyond 3; waypoint 10 s refusal. A second SME pass over this draft was not run (time box); PRIMREVIEW does the review.

## 10 highest-risk primitives (criticality x evidence gap x counts)
1. P7 pickup potion with belt priority (S, L~, 939 skips) 2. P1 belt fill (A, L~, 26 zero moves) 3. P3 drink by column (S, L~, 4391 no-potion skips) 4. P10 TP and enter portal (S, L~, 2683+710) 5. P2 belt layout by type (S, U) 6. P23 waypoint row choice (S, L~, 220+198) 7. P26 skill bind and cast (S, O, UNVERIFIED) 8. P28 life/mana monitor (S, L~, Meter totals U) 9. P35 UI hygiene and cursor (S, L~, 200) 10. P16 repair (A, L~, 36 fails). Next: P8, P4, P36.

## Proposed drill order (m1-liveproof)
D0 harness + P35 gate, then D1 belt (P1 P2 P3 P7), D2 portals and town entry (P10 P11 P12 P14 P23), D3 vendor and gold (P15 P16 P17 P13 P5 P8), D4 pack (P4 P9 P6), D5 combat (P26 P27 P28 P29), D6 movement (P22 P24 P25), D7 build (P30 P31 P32), D8 NPC (P18 P20 P21 P19), D9 lifecycle (P33 P34 P36 re-gate).

DRILLS5 follow-up (fixmisc, docs/fixmisc.md): P28 poor-character mana rule is intended (pinned by a test); P27 first Clear after a position jump now syncs the navigator map (`Navigator.Sync`), live proof pending.
fixstall (2026-10-31, offline, not live): H.pause F1 (NPC approach stall: early stall detection, distance-sized game-time budget) and D6.d F2 (a plain GoTo stalled in a pocket escapes by town portal, P25) fixed in code; proof = rerun H.pause and D6.d; docs/fixstall.md.

DRILLS6 rows (2026-10-31, lab build cheats3, throwaway Sorceress, each drill twice; files `2026-10-31-drills6*.txt/.json`, README section DRILLS6). These supersede the DRILLS5 rows of the same drill ids:
| prim | drill | verdict | evidence |
|---|---|---|---|
| AUDIT | H.pause | PARTIAL (finding F7) | bot's own Pause: 1 of 2 tries opens Kashya in 136 ticks, the other fails at 718 ticks (28.4 s budget, stalled near 25.9,20.5); identical in both invocations (lockstep is deterministic) |
| P25 | D6.d | PARTIAL (finding F8) | the escape fires (`nav: stalled in a pocket, escaping by town portal`) and run 1 ends in ErrLevelChanged, town in 493 ticks; run 2 (another pin position) opens the portal but cannot walk to it: stuck 62 s |
| P27 | D-sync | PASS | first `Fighter.Clear` after `debug goto`, no warm-up walk: 3 of 3 zombies killed, 0 unreachable lines, 9 s |
| P22 | D-run | PASS | run toggle: 0.556 to 0.563 subtiles per tick (running) in a fresh game, in the same game again and after a restart (runKeys 1 then 2, runResets 1 to 3), stamina drains; the baseline without the run key walks at 0.375 (ratio 0.67), no drain |
| P10 | D2.* | PASS alone, order-dependent in the full group | D2.a, t, a2, b, c, g, d, d3, e, h, f PASS when run as subsets; the full group run has D2.a2 (stale portal), D2.d3 (DIFF) and then D2.e, h, f (setup: the `debug hit` left the character dead) FAIL |
| P26 | D-skill | NOT RUN | docs/research/skill-bind-trace.md is not on main and its branch worktree is gone |

DRILLS7 rows (2026-10-31, lab build cheats3, code at ea07758 with fixstall2; files `2026-10-31-drills7*.txt`, README section DRILLS7). These supersede the DRILLS6 rows of the same drill ids:
| prim | drill | verdict | evidence |
|---|---|---|---|
| AUDIT | H.pause | FAIL (finding F10) | fixstall2: 3 invocations, 6 bot tries, 3 open Kashya (136 ticks), 3 fail at 718 ticks at (25.9,20.5), identical to DRILLS6: `planAround` has no route from the stall cell (clearance 1), clearance 0 would |
| P25 | D6.d | FAIL (finding F11) | 3 invocations identical to DRILLS6 (run 1 escapes in 493 ticks, run 2 stuck 62 s): the walking portal entry is not wired (`hoverBelt` has no `InteractWalks`) |
| P27 | D-sync.tp | PASS | first Clear after a town portal arrival, no warm-up walk: 3 of 3, 0 unreachable lines |
| P27 | D-sync.wp | FAIL 1 of 2 (finding F12) | first Clear after a waypoint arrival: run 1 3 of 3, run 2 0 of 3 (ignored: 12 approaches without closing in), no unreachable lines |
| P10 | D2.* | PASS x2 in the full group | 11 of 11 in one process, twice (harness reset: ensureAlive, longer stale-portal walk-away, D2.h signature) |
| P26 | D-skill probe | PARTIAL | `debug skill right 54` (0 points) `err not usable`, `right 36` ok; the `cmd select` effect (D1) not observed without a step |

DRILLS8 rows (2026-10-31, lab build cheats3, code at 1628eeb with fixstall3; files `2026-10-31-drills8*.txt`, README section DRILLS8). These supersede the DRILLS7 rows of the same drill ids:
| prim | drill | verdict | evidence |
|---|---|---|---|
| AUDIT | H.pause | PARTIAL (finding F13) | 3 invocations: first-run bot tries 3 of 6 open (144 ticks), 3 fail at 796 ticks with the player 3 cells from Kashya (the clearance-0 route frees the stall, but late); second runs 6 of 6 open |
| P25 | D6.d | PASS x3 | `ErrLevelChanged` after the pocket escape, 225 ticks (9 s), in town, identical repeats (DRILLS7 run 2: stuck 62 s) |
| P27 | D-sync.wp | PASS x3 (after the harness mana fix) | before: 1 of 2 (repeat 2 started with 1 of 35 mana, melee fallback, no obstacle in the grid); after: 3 of 3 killed in every repeat |
| P14, P15, P16, P5, P13, P18, P3 | D3, D8 NPC smoke through `d2dtown.Input` | PASS | D3 12 of 12, D8 a, m, q, w PASS (D8.c is once per character) |
Note (fixtalk, 2026-11-01): the bot's talk nodes (P20 Akara reward, P21 Warriv act hop) now live in internal/talk (moved out of cmd/d2bot, behaviour identical); D8.q and D8.w call them and are built, awaiting live proof ([fixtalk.md](fixtalk.md)).

Note (fixf17, 2026-11-01): DRILLS9 F17 (DW.m1 walk timeout, F group run 1) is a harness Pacer lag, not a nav stall (the lab's direct `Client.Step` is invisible to the Pacer); fixed in cmd/d2drill, live proof pending ([fixf17.md](fixf17.md)).

Note (fixstall4, 2026-10-31): DRILLS8 F13 (H.pause, NPC approach) cause found offline: the grid route's 12-step waypoints cut the camp block corner (each leg is a straight `cmd move`), and stalls spent the budget; fixed with line-of-sight legs, routing at every stall and a route-sized budget, live proof pending ([fixstall4.md](fixstall4.md)).

DRILLS9 rows (2026-11-01, lab build cheats3, code at 912086a with fixstall4; files `2026-11-01-drills9*.txt`, README section DRILLS9). These supersede the DRILLS8 rows of the same drill ids:
| prim | drill | verdict | evidence |
|---|---|---|---|
| AUDIT | H.pause | PASS (F13 closed) | 3 invocations x 2 repeats: the bot's Pause opens Kashya in 6 of 6 first-run and 6 of 6 second-run tries, 248 ticks each (DRILLS8: 3 of 6 failing at 796) |
| P36/Gate 1 | D10 | PASS x2 | two full runs (kill -9 of d2d under the managed bot): restarts delta 1, loop resumed, saved XP not lower, no leftovers, P35 clean, both inside the TTL |
| P1 | D1.P1s | PASS in the group x3 | harness: `stagePack` picks up loose items first (F14) |
| AUDIT | H.obj.chest, H.obj.shrine | PASS in group F x4 | harness: tries further objects of the kind, clears leftover monsters first (F14) |
| P18 | D8.c | PASS run 1, BLOCKED run 2 | runs first; BLOCKED when a merc is already owned (the reward is once per character) |
| P23 | DW.m1 | PASS alone, BLOCKED run 1 in the full F group | finding F17 (README DRILLS9) |

DRILLS10 rows (2026-11-01, lab build cheats3, code at a7f5bc6; files `2026-11-01-drills10*.txt`, README section DRILLS10):
| prim | drill | verdict | evidence |
|---|---|---|---|
| P23 | DW.m1 (F17) | PASS in 4 of 4 full F groups, run 1 and 2 | fixf17 live: 483 ticks, 1 bounce; no `the Pacer lagged` line (threshold 1 s) |
| P20 | D8.q | PASS | fixtalk live: the bot's `talk.Adapter` Akara reward, skill points 0 to 1, record 0x0001 |
| P21 | D8.w | PASS | fixtalk live: Warriv `go east`, area 40 |
| P28 | D-meter | PASS | fixmeter live: kills 13 = 13, dealt 121 = 121 (spawn hp sum 222), taken 106 = 106, hits and misses `n/a`; totals are cumulative across `Enable` (a reader needs deltas) |
| P26 | probe | RECORDED | Fire Bolt mana 35 to 32 and zombie hp 11 to 8 (GH); `cmd skill` in town and a 0-point skill do not spend mana; `cmd select 54 0` leaves rskill 0 (UNVERIFIED) |
