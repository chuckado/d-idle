# Survival: review and design (2026-11-14)

> status: proposal, docs only (TASK-430). Nothing is implemented. Read this page; the detail is in [review-2026-11-14.md](review-2026-11-14.md) (what exists and why it fails) and [design-2026-11-14.md](design-2026-11-14.md) (what to build). Tags as in those docs (V, S, I, U).

## Why we keep going in circles (5 lines)

1. The bot never estimates how strong the hero is or how dangerous a zone is. 54 mechanisms and about 150 knobs compare proxies (character level, life percent, gold lines) with hand-set numbers.
2. Nine separate notions of "may the hero be here" disagree: the Countess is safe at clvl 10, 11, 13, 14 or 15 depending on the rule; the Smith at 8 or 16.
3. The proxies fail where it hurts: a clvl 12 Paladin died in Black Marsh at a "level gap" of minus 6. Real cause: packs of 14 to 25 at 9 s per kill, 88 potions per 100 kills (Cold Plains: 1.5 s, 0 potions).
4. The rules are open loop: after the first bad visit the hero went back three times in four minutes (1 death, 3 merc deaths, the purse emptied). The only memory is a time ban, which parked another hero in town for 86 percent of 15 hours with 0 deaths.
5. So every incident gets a new rule or a moved number, tested on the run that caused it, and the next run breaks where the proxy is wrong the other way.

## The new model (8 lines)

1. **Danger profile** per zone, generated from the game tables: packs (typical, bad, elite), each monster's damage per second, life, reach and hazards, bosses, XP and gold per kill. Known once; hand overrides only with evidence.
2. **Power estimate** of the hero, continuously, in the same units: life and potion pool, sustain, chance to be hit, block, resists, kill speed, merc, escape; each input flagged measured, derived or default, so uncertainty is a visible discount.
3. **Margin** = life the hero can spend / life the pack takes while it is being killed. It grows with levels, gear, a full belt and a living merc; it falls with the square of the pack size and with slow kills.
4. **Risk budget**: one dial (Cautious, Balanced, Bold) sets the margin needed to enter a zone and a lower one to stay (hysteresis).
5. **Zone choice**: best XP per minute (plus gold when the purse is short) among zones whose bad-pack margin clears the dial; travel and town time only amortise. It ratchets up as power grows; no entry or exit levels anywhere.
6. **Ledger** (feedback): what each zone really cost (potions, life dips, chickens, merc deaths) corrects the model per zone and per hero; a zone that costs more than it pays is left and not re-entered until power changed. No bans.
7. **Archetypes**: one engine, five kinds of character over two axes (how it is exposed: contact, range, screened by pets; how it deals damage: sustained, mana-gated, pets). Bruiser (Paladin), ranged physical, caster (Sorceress), summoner, hybrid. Parameters are derived from class and skill tables; a build maps to the nearest archetype by rule; per-class overrides need evidence, expire, and are capped at five.
8. **Unchanged**: the reflex layer (drink, chicken, crowd rule), the death rule, "warn, never wall". Quests stay goals; a boss gets a margin instead of a level floor. Every decision logs one "why" line and feeds "Why here" and "Why not X" on the watch screen.

## Slices

| Slice | What | Changes behaviour |
|---|---|---|
| S-A | danger profile generator, overrides, tests | no |
| S-B | archetypes, power estimate, margin; shadow log next to every planner line; replay tool; validated against the swap7, hour 4 and hero14 logs | no |
| S-C | ledger and controller in shadow, "why" line, API fields for the UI | no |
| S-D | zone choice switches behind `bot.zoneModel` (default off); A/B on the sandbox | behind a knob |
| S-E | retire old rules one group at a time with the evidence: recovery, poverty and kit gate; plan level bands; stale-pack and death bans; boss and route floors; trip triggers | yes |
| S-F | optional: the same margin for per-pack engage and retreat | behind a knob |

Friends' 0.1.x builds are not affected: S-A to S-C only add files and log lines; S-D ships default-off in a later release.

## Open questions (recommendation in brackets; full list in the design, section 12)

1. Dial names and default Balanced? [yes]
2. Show a single margin number on the watch screen? [yes, always with its weakest term]
3. May the margin replace the ADR 0011 level floors for bosses once it matched the logs, keeping "warn, never wall"? [yes, boss by boss]
4. May the bot leave the merc dead when the zone it would farm kills him, even with gold? [yes, with the reason shown]
5. May the controller start a fresh game when the best safe zone is depleted (15 s cost)? [yes]
6. Five archetypes over two axes: is that the "in between" you meant? [yes]
7. Override budget of five per class, with evidence and expiry? [yes]
8. Second live drill class after the Paladin? [Sorceress]
9. Fold the unwired readiness model (TASK-217, 218) into S-A and S-B and close TASK-219 to 223 as superseded? [yes]
10. S-D default-off in the next minor release, default flips only after the A/B and your go? [yes]
