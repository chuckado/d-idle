# Mercenary AI and the survival data it gives the bot (agent decomp30, 2026-11-15)

> status: research, current (card decomp30, TASK-506; consumers: docs/survival/design-2026-11-14.md section 5.5, docs/mercrevive.md, internal/merc policy, hire policy). Offline, docs plus one display guard. **No new emulator run in this card**: the emulator image and the read-only game.exe copy live on the heavy worker, which was down, and waking it needs the user's go; the NAS builder has neither. So every V below is a V run or V read of an EARLIER card, cited by name (tag S for this card); what this card adds is new reading of those results for the bot (tag I or D) and the open questions with a drill recipe. Nothing pushed, posted or filed.
>
> Tags: V run (game.exe's own code under the emulator against an oracle, an earlier card), V read (disassembly), D (derived from tables and formulas of those cards), S (earlier proof, named), I (inferred), U (unverified).

## Answer in eight lines

1. **The merc never retreats and never drinks.** Its think and attack functions read no life stat; a potion reaches it only when the player hands one over (`FUN_0054d230`: healing and rejuvenation classes accepted, mana refused). The bot's belt hand-over at `merc.potionLifePct` is the only life control there is. (V run + V read, S: mercstats Q4/Q6, merc2 Q3)
2. **The leash decides how well he fights.** Owner over 100 subtiles: teleport into the owner's room; over 24: run; over 16 with the owner moving: follow; only after those checks a foe under 25 subtiles gets attacked. A hero that keeps moving is a hero whose merc does not shoot. (V run, 100,000 cases, 0 differ from d2d; S: pets-merc Q2)
3. **Monsters go for whoever is strictly nearest, the player first on a tie** (merc newest-pet rule aside); the player 55+ subtiles away drops the player and all pets from the search. There is no aggro score that keeps monsters off the merc. A Rogue (shooter) stands 16..24 off the hero and backs off from a foe under 4 off half of the thinks, so monsters reach the hero first unless the merc is nearer. (V run, S: pets-merc Q1)
4. **Revive is a price, not a rule gate**: `(level^2 / 2) x 15`, at most 50,000, carried gold then stash, no portal, no potion. (V run, S: mercstats Q2, merc2 Q2)
5. **Replacing a dead merc is 3 to 15 times cheaper than reviving him** from merc level 9 up (table Q6, D from the two V run formulas). The revive buys back only the merc's level and the gear he wears (what dismissal does to the gear is U, drill R3).
6. **The display "life 4200%" is a unit mix-up in the dump, not a game fact.** The bot computes `current * 100 / max` from a d2d dump pair; a percent over 100 needs a current value in a larger unit than the maximum. Root cause U (the fork's dump source is not in this checkout); the guard clamps the percent (B1, fixed here with a test).
7. **d2d deviations** (all already known): pets tried merc-first not newest-first (D1), the merc's own life regeneration stat missing (D2), Act 2 auras do nothing (D3). None is new, none is built as a PR here.
8. **The survival model's merc term** should be per hireling type and level from the table in Q6, not one constant (B2).

## Q1 The AI loop (all S, V run in the cited cards)

| Question | Answer | Address | Tag |
|---|---|---|---|
| Think entry | MonAI 61 Hireable think; in a walk or run it does nothing; the think runs at the think event (a stand waits 5 or 10 frames of 40 ms, a walk's end thinks at once) | `005e52d0`, `005e5050` (attack), `005a8030` | V run (merc.py 100,000), V read |
| Leash | gap to the owner (distance less size): over 100 teleport (mode 3 at `005e3930`), over 24 run (mode 1, pace 60), over 16 with the owner walking or running follow (mode 0) | `005dc380` | V run |
| Fight | only then, out of town, a foe under 25 subtiles goes to the attack think | `005e5050` | V run |
| Target | nearest threat (MonStats +0x4e, 2 or more) in sight under 0x31; else secondary; a secondary under 6 with no path to the primary takes the nearest other primary within 0x13 | `005ddc30` mode 6 | V run (pets_merc.py M 100,000, 0 differ from d2d) |
| Attack skill | melee mercs roll 98; others 40 + 2 x level + 10 per miss, at most 95; the shooter (aip1 0, the Rogue) under 4 off backs off 4 half the time; aip1 != 0 (Act 2) walks to melee | `005e5050` | V run (merc_attack.py 100,000) |
| Low life | no life test in either function | all of the above | V read, I for "nowhere else" |
| Own potion use | none; hand-over only | `0054d230` | V read + V run |
| Aura / skill cadence | Act 2 auras count as running but do nothing in d2d; game.exe cadence not traced here | merc-ai.md | S, U |
| Reaction | no hook from the player's attack; at most 10 frames after a stand | driver | S |

## Q2 Area change, town, portal, waypoint

- Rule of the leash applies in every area: he teleports into the owner's room at over 100 subtiles (V run). After an area change the merc is placed by the area-change code (`FUN_0061b130`, S pets-merc B3) and follows at pace 60; so after a town portal or waypoint he arrives at once or within a few seconds, not 10 s. The SME rule 9 wait (10 s) is longer than the rule needs (I). Exact arrival delay and position: U, drill R2.
- In town the merc does not attack (the foe branch is out of town only) (V run).
- A dead merc is not in the owner's pet list (`005a6520`), stays dead across areas and game exit until paid for (S merc2 Q2).

## Q3 Damage taken: who the monsters pick

| Rule | Tag |
|---|---|
| An evil monster searches the player then its pets newest first, taking a candidate only when strictly nearer by `(min + 2 max) / 2` subtiles than the best so far; ties go to the player; the player at 55+ subtiles, in town or another act drops itself and its pets out | V run (search.py 100,000, pets_merc.py 100,000) |
| Good and neutral monsters (mode 5) search any enemy; threat only splits primary from secondary | V run |
| Nothing keeps monsters off the merc, and nothing draws them to him except being nearer | V run, I for "nothing else" |
| Merc stats: life, defence, damage, AR, resists are the hireling.txt row at the merc's level (2,940 Id x level cases, 0 differ from d2d); resistances take the difficulty penalty on every hit (Nightmare -40, Hell -100) | V run (S mercstats Q1, Q7) |
| Merc's own life regeneration: stat 0x4a set at every level-up to stat 7 (256ths) / 2000 | V read; effect U (d2d D2) |

## Q4 The display "life 4200%" (B1)

The merc block of GET /api/inventory carries `lifePct` and the page prints `'life ' + lifePct + '%'` (`internal/watch/web/inventory-merc.js`, tile and header, before this card unclamped; the ten-pip bar clamped, so the bar looked right and the text did not). `lifePct` is `Merc.LifePct`, which comes from `MonsterInfo.LifePct`: on the d2d backend `internal/d2dsrc/player.go` computes `c * 100 / mx` from the dump's merc `life` pair; on the memory backend `internal/d2/units.go` computes stat 6 x 100 / stat 7. Both stats are in 256ths, so the memory path cannot mix units. The d2d fixtures (`dump_head_late.json`: `"life":[63,80]`) are whole points. 4200 percent means the current value was about 42 times the maximum, e.g. the pair mixing a larger unit (hit points scaled, or life plus gear) with the unscaled row maximum of the hireling table. **Root cause U**: the dump emitter is in the private fork's devctl dump code, not in this checkout, so I could not read it. Recipe R1 reads both. The percent in the bot is never over 100 by the game's own rules (cur <= max, V read of the stat clamp is I), so a value over 100 is always a source error. Guard (fixed, with test): clamp to 0..100 in `d2.FindMerc` and in the page. A clamp hides the symptom but keeps every consumer correct (`resource.MercLifePct`, `merc.potionLifePct`: a too-high percent would have made the bot never hand a potion over, so the merc potion rule may have been silently off whenever the mix-up was present, U).

## Q5 What the bot should use (data)

Hire choice, replace-or-revive and potions, from the V run formulas (offer `Gold x (15 d + 100) / 100`, d = offered level - row Level; revive `(L^2/2) x 15`, cap 50,000). Row Levels derived from the cited sample costs (I): Rogue 3, Desert Guard 9. Gold column: Rogue Fire 100, Rogue Cold 150, Desert 350, Iron Wolf Act 3 Normal 1,000 to 1,500 (S merc2 Q1).

## Q6 Revive against replace (D, Normal Act 1 and 2, equal level)

| merc level | revive | hire Rogue Fire | hire Rogue Cold | hire Desert Guard | revive / Cold |
|---|---|---|---|---|---|
| 5 | 180 | 130 | 195 | 350 | 0.9 |
| 9 | 600 | 190 | 285 | 350 | 2.1 |
| 12 | 1,080 | 235 | 352 | 507 | 3.1 |
| 15 | 1,680 | 280 | 420 | 665 | 4.0 |
| 20 | 3,000 | 355 | 532 | 927 | 5.6 |
| 30 | 6,750 | 505 | 757 | 1,452 | 8.9 |
| 40 | 12,000 | 655 | 982 | 1,977 | 12.2 |
| 50 | 18,750 | 805 | 1,207 | 2,502 | 15.5 |

(Nightmare and Hell hires cost 6,000 to 14,000 flat for a Rogue, S merc2 Q1, so replace is a Normal idea.) Life per level (hireling samples, S mercstats Q1): Rogue about 9 a level to 30 then about 15; Desert Guard about 15 a level to 43, about 25 after: a Desert Guard has roughly 1.6x a Rogue's life at the same level and costs 3 to 4x as much. Per gold, **Rogue Fire is the survivability winner under level 30; the Desert Guard wins only where the damage taken per second makes life the limit** (I: no ledger data, U).

Rules the data suggests for design section 5.5 (all I, to be tuned on ledger data):
1. Below merc level 9 revive and replace cost about the same: replace only if the offered level is higher.
2. From merc level 9 up prefer replace when the merc carries no gear worth keeping (gear score from `internal/merc/gear.go`) and the hero's clvl is within 5 of the merc's level (the offer lands at clvl-5..clvl-1, so level is not lost).
3. Revive only when `price < replace price + value of worn gear + level gap x per-level value`; with the merc's median life of 25 min (sandbox replay) a 3,000 gold revive at level 20 is about 120 gold per minute of expected life, above the gold a clvl 20 farm makes: **a dead merc at level 15 and up is not worth reviving unless he wears gear** (I, needs the ledger).
4. The Rogue backs away, the Desert Guard walks to melee (V run): the Guard takes more damage but also holds monsters; do not count a Rogue as a tank (S pets-merc B1).
5. Potions to give: any healing or rejuvenation potion; rejuvenation is accepted and is the better use (full heal for one hand-over, V run). Mana potions are refused. The hand-over at 40 % (`merc.potionLifePct`) is correct: nothing else raises his life in a fight.

## Deviations (d2d vs game.exe)

| id | rule | tag | differential |
|---|---|---|---|
| D1 | pets tried merc first then oldest; game.exe newest first (the merc is re-inserted on hire, load and revive) | V run (1,254 of 100,000 -> 0 at rule level) | exists (`pets_merc.py` + harness, S pets-merc); no Sorceress effect |
| D2 | merc's own life regeneration (stat 0x4a, set at every level-up) not modelled | V read game side; effect U | could be built: run the monster regen function of game.exe over N frames for a merc unit with stat 0x4a set, oracle life per 25 frames against d2d's `hit_points` |
| D3 | Act 2 merc auras count as running but do nothing | S | none; build: run `FUN_0056b7e0` with the merc as caster (palaura-trace) |
| D4 | merc dump `life` pair units (the display bug) | U | read the dump in the fork; a fixture test (R1) |

No PR candidate: D1 is rule level only (no patched d2d build), D2 and D3 are effect-U.

## Bot findings

| id | where | finding |
|---|---|---|
| B1 | `internal/d2/merc.go` FindMerc, `internal/watch/web/inventory-merc.js` | an unclamped percent could print over 100; fixed with `TestFindMerc` case "percent over 100 is clamped"; root cause U (D4) |
| B2 | `internal/merc/policy.go`, design 5.5 | the merc term is one number for any hireling; use the Q6 table (class, level, row) |
| B3 | revive gate (`merc.ReviveNeed`) | the bar is price x mult plus reserve; add "replace instead" when revive / hire price over 3 and no gear worth keeping (rule 2) |
| B4 | SME rule 9 wait | 10 s wait after a portal or waypoint longer than the leash needs (I) |
| B5 | `merc.potionLifePct` | works only if the percent is right: B1's mix-up could have kept it from ever firing (U) |

## Drill recipes (lab fork verbs, docs/forkdbg.md; not run here)

- R1 the dump: `debug merc hire`, `dump` after a hit on the merc (`debug merc hurt` does not exist, S d1.go missingMerc; use `spawn` a zombie next to the merc), compare the dump's merc `life` pair against `debug monsters` hit points and the max in the hireling row. Expect both in whole points.
- R2 portal arrival: merc at 5 and 30 subtiles, cast a town portal and enter it, count ticks until the merc is within 16 subtiles in the new area.
- R3 dismissal: hire a second merc while the first wears an item; see where the item goes (`dump inventory`).
- R4 regen: reduce the merc's life, let 60 s of game time pass in town, compare against stat 0x4a / 2000 (D2).

## Matrix cell proposals

- merc survival model: hire and revive costs V run, behaviour V run, replace-vs-revive D (Q6), regen U, dump life unit U.

## Not done

No emulator run (the heavy worker down, user's go needed to wake it); no check of the dump source; the table in Q6 is arithmetic on already run formulas.
