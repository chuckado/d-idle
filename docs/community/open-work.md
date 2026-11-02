# Open to contributors

> status: current, APPROVED by the maintainer 2026-11-01 (card communitywork). The source list for `tools/contrib/board.py`, which renders [work-board.md](work-board.md) and the GitHub issue bodies. How to take an item: [../contributing/agent-work.md](../contributing/agent-work.md).

Only items that need nothing but the game binary of your own legitimate Diablo II LoD 1.14d install and public sources (this repo, upstream `psi29a/d2decomped`) are listed. Live drills, soaks and anything on the maintainer's private lab are NOT open (see the guide). The maintainer keeps this file in step with `docs/decomp-queue.md` and `docs/BACKLOG.md`; where they disagree those two win.

Format (read by the board script): one `## ID: title` section per item, then `- key: value` lines. Keys: `size` (S, M, L), `status` (`ready` = open for a claim now, `proposed` = open but the maintainer may still reshape it, so ask in the issue first; `claimed` and `done` are shown but get no issue; this is the contributor status, not the queue row's), `kind` (decomp, verify), `source` (`queue:N` = row N of docs/decomp-queue.md, `backlog:NAME` = a line in docs/BACKLOG.md), `question`, `start`, `deliverable`, `needs`. Function addresses are tags only (`FUN_xxxxxxxx` as in the repo's trace docs); never paste code.

## PALSKILL: Smite, Zeal and Sacrifice dispatch in game.exe
- size: M
- status: proposed
- kind: decomp
- source: backlog:SKMELEE+
- question: How does game.exe dispatch Smite, Zeal and Sacrifice (hit count, AR and damage bonus, mana and life cost, knockback and stun, what a held button repeats), and which of these does d2d's `skill_built` branch implement versus falling back to a plain attack?
- start: docs/research/paladin-act1-recon.md section 6 rows 1 and 5; the skill bind trace docs/research/skill-bind-trace.md for the packet side (FUN_0054be70, FUN_00647960); upstream d2decomped fight code (public) for what d2d does.
- deliverable: docs/research/pal-skill-trace.md; a static result (which skills fall back in d2d, read from upstream source, tagged S) is a valid first PR on its own.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn if you run code.

## PALAURA: how an aura applies in game.exe
- size: M
- status: proposed
- kind: decomp
- source: backlog:PALAURA
- question: Prior work (do not redo): docs/research/pal-skill-trace.md section 3 already has the pulse period (50 frames), the ranges (16 + 2 per level for Might, Prayer, Defiance and Concentration; 6 + 1 per level for Holy Fire) and every calc slot (21,760 cases, 0 differ against d2d). Open: (1) the aura target filter (aurafilter 0x12003 for the friendly auras, 0xa783 for Holy Fire) and the shape of the area search FUN_0056b7e0 (who is passed: self, merc, pets, party, enemies; the exact radius test); (2) the damage an aura does to enemies per pulse (Holy Fire, Shock, Freeze: roll, resist, which units), read and then run; (3) Prayer's per-pulse mana and state 85 (`nomanaregen`): does the regeneration code really stop while it is set. Also useful: the D5 Might mana floor and aura reach to merc and party on d2d (the trace's D1 to D7).
- start: docs/research/pal-skill-trace.md sections 3, 6 (D5, D6) and 8 (drills A2, A3); FUN_005cf010 (friendly aura do), FUN_005cf3a0 / FUN_005d0920 (Holy Fire, Shock, Freeze do), FUN_0056b7e0 (area apply), FUN_005cedc0 (callback), FUN_00644b10 and FUN_0056c110 (mana); the emulator style of tools/emu/pal_aura.py; docs/research/paladin-act1-recon.md section 6 rows 3 and 11; docs/research/d2d-fidelity-riskmap.md.
- deliverable: docs/research/pal-aura-trace.md covering only the open parts above, each claim tagged V (run in the emulator), V read, S or U; for each part end with the exact probe the maintainer should run on d2d (the live half is maintainer-only) and the number you expect. Done when: the filter bits are decoded and the area search shape is stated with a V run on a fake player, merc and monster set; the enemy hit of one aura is run against the trace's element tables; Prayer's mana and state 85 effect are read to the regeneration code; the trace's D5 and D6 rows can be updated from it.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.

## PALDUR: durability loss per hit and per block
- size: S
- status: proposed
- kind: decomp
- source: backlog:PALDUR
- question: In game.exe 1.14d, what is the chance and the slot choice for durability loss when the player hits and when the player is hit or blocks, and does d2d implement the same rule (it has `debug wear` verbs but the loss rule is unknown)?
- start: docs/research/price-trace.md (the repair gate and price are done: FUN_0062e660, FUN_00629930, use them as the "durability" anchor); docs/research/paladin-act1-recon.md section 6 row 7.
- deliverable: docs/research/durability-trace.md, small; rule with tags, d2d comparison from upstream source (S) or a differential (V) if you build one.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.

## SKGROUND: ground skills (Blizzard, Meteor, Hydra, Fire Wall, Thunder Storm, Frozen Orb)
- size: L
- status: proposed
- kind: decomp
- source: queue:22
- question: How many sub-missiles do these skills create, at what ticks and radius, and what is the skill delay (status 8, which d2d does not model)? Pick ONE skill (say Blizzard) and trace it fully; the others are separate PRs.
- start: docs/research/sorc-skill-matrix.md section 5 row 9; docs/research/sk-missile-trace.md and sk-nova-trace.md for the sub-missile timeline style (FUN_0056d400 is the Frost Nova ring spawn).
- deliverable: docs/research/sk-ground-<skill>-trace.md; advanced item, claim it only if you have done a decomp before.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.

## VERIFY-SKMANA: independent re-check of the mana trace
- size: S
- status: ready
- kind: verify
- source: queue:19
- question: Re-derive the mana cost check and payment from game.exe yourself and with your own harness; do NOT start from our script. Our claims: the check is stat 8 >= (mana + lvlmana x (lvl - 1)) << manashift with no minmana floor (FUN_0056c160); the payment is max(formula, minmana x 256) (FUN_0056bfe0), nothing is taken when short and the do ignores it. Say which hold, which hold with limits, which fail.
- start: docs/research/sk-mana-trace.md (the claims, read AFTER you formed your own view of the two functions); tools/emu/sk_mana.py (only to compare numbers at the end).
- deliverable: docs/research/verify-skmana.md: per claim HOLDS / HOLDS WITH LIMITS / FAILS, your own case count and seed, the command to reproduce, differences from the author's numbers.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.

## VERIFY-WAYPOINT: independent re-check of the waypoint refusal claim
- size: S
- status: ready
- kind: verify
- source: queue:3
- question: Our claim: game.exe refuses a waypoint use only within 10 s of the player's last hostile declaration (pcdata offset 0x160, the only writer FUN_0055b720), never in single player, while d2d counts the window from every arrival. Verify the writers and readers of that field yourself and the claimed 59,309 of 132,964 differing cases for the d2d rule.
- start: docs/research/waypoint-trace.md (after your own look); upstream d2decomped waypoint code (public).
- deliverable: docs/research/verify-waypoint.md in the same HOLDS / LIMITS / FAILS format, with the reproduction command.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.

## VERIFY-FREEZE-KINDS: independent re-check of which monsters game.exe can freeze
- size: S
- status: ready
- kind: verify
- source: queue:6
- question: Our claim: game.exe freezes only plain monsters and the minions of a unique; bosses, champions and uniques are chilled at most (the test is the MonStats boss bit, monster data flag 8, or a hireling class; the flag-8 makers are the champion, unique and boss pack functions). Re-derive the test and its flag writers yourself, then check the claim over a population of spawned monsters with your own harness. Say which parts hold, which hold with limits, which fail.
- start: your own look at FUN_0057c6c0 and FUN_0057b230 first; then docs/research/freeze-rules-trace.md (the claim C3 and its flag table, read AFTER you formed your own view); docs/research/pr-freezeresume-verifier.md for the report style.
- deliverable: docs/research/verify-freeze-kinds.md: per claim HOLDS / HOLDS WITH LIMITS / FAILS, your own case count and seed, the command to reproduce, differences from the author's numbers; an oracle script under tools/emu/ if you ran game code.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout, Python 3 with unicorn.

## VERIFY-TELEPORT-ROOM: independent re-check of the Teleport room rule
- size: S
- status: ready
- kind: verify
- source: backlog:SKTELE
- question: Our claim: game.exe refuses a Teleport hop whose target lies outside the caster's room and the rooms in that room's near list (FUN_00554ea0 with FUN_00463740 over FUN_00619790), so the range is a room rule and not a distance; a per-level Teleport value of 0 refuses and 2 adds a line test (FUN_006229f0). Re-derive the destination rule yourself on a grid of caster positions and targets, class-generic, and compare with the claim.
- start: docs/research/sk-teleport-trace.md (read AFTER your own look at FUN_005ca360 and FUN_00554ea0); tools/emu/sk_tele_spot.py only to compare numbers at the end.
- deliverable: docs/research/verify-teleport-room.md: per claim HOLDS / HOLDS WITH LIMITS / FAILS (accepted and refused counts, grid, seed), the reproduction command, differences from the author's numbers.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout, Python 3 with unicorn.

## VERIFY-AIDEL: independent re-check of the monster think period
- size: S
- status: ready
- kind: verify
- source: backlog:FREEZERESUME
- question: Our claim: a monster thinks again every `aidel` frames (MonStats value, 15 for the Normal rows read, difficulty-dependent in the later rows) and resumes thinking at the freeze end; d2d's `set_off` uses `aidel`, or 15 when the value is 0. Check the aidel column per difficulty and the think scheduling (FUN_005a73e0, FUN_005a7c20) with your own harness, and the 55,268 equal cases claimed for stand and aidel.
- start: docs/research/mon-wake-trace.md (the aidel rows, read AFTER your own look at the two functions); docs/research/pr-freezeresume-proof.md for the freeze-end claim.
- deliverable: docs/research/verify-aidel.md: HOLDS / HOLDS WITH LIMITS / FAILS per claim, your case count and seed, the command, differences from the author's numbers.
- needs: your own game.exe 1.14d install (the monster table read only, never shared), upstream d2decomped checkout, Python 3 with unicorn.

## VERIFY-SKMELEE2: independent re-check of the melee deviations on the real build
- size: M
- status: ready
- kind: verify
- source: backlog:SKMELEE2
- question: Our trace says five deviations between game.exe and d2d's player blow are real on the build: D1 (the physical is kept in 256ths), D2 (Blunt weapon plus stat 122 against undead), D4 (barehanded Strength term), D5 (attack rating percent floor) and D6 (the -90 floor applied once). Re-run the comparison with your own harness: game.exe's blow functions against a build of upstream d2decomped, same cases, and say which hold, which hold with limits, which fail.
- start: docs/research/sk-melee2-trace.md (the deviation table, read AFTER you formed your own view of FUN_0057b420 and FUN_0057b7d0); docs/research/sk-melee-trace.md for the formula; tools/emu/sk_melee2.py only to compare numbers at the end.
- deliverable: docs/research/verify-skmelee2.md: per deviation HOLDS / HOLDS WITH LIMITS / FAILS, your own case counts and seeds, the commands to reproduce, the differences from the author's numbers.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout and its build tools, Python 3 with unicorn.

## VERIFY-SKRESIST: independent re-check of the resistance rules Q1 to Q10
- size: M
- status: ready
- kind: verify
- source: queue:21
- question: Re-check our monster resistance rules with your own harness, not ours: spawn base per difficulty, immune at a total of 100 or more, the damage formula from a resistance, the one-fifth rule for curses and auras on an immune, pierce, and the penalty and cap that apply only to non-plain defenders. Say per rule which hold, which hold with limits, which fail.
- start: docs/research/sk-resist-trace.md (section 1 table Q1 to Q10, read AFTER your own look at FUN_00573cb0, FUN_0057be00, FUN_0057bf80 and FUN_005c3540); docs/research/pr-curseimmune-verifier.md for the earlier verifier's style; tools/emu/sk_resist.py only to compare numbers at the end.
- deliverable: docs/research/verify-skresist.md: per rule HOLDS / HOLDS WITH LIMITS / FAILS, your own case count and seeds, the reproduction commands, differences from the author's numbers.
- needs: your own game.exe 1.14d install (the data tables read only, never shared), upstream d2decomped checkout, Python 3 with unicorn.

## SKGEARAR: attack rating terms from gear (stats 115, 116, 120, 123, 124, 179)
- size: S
- status: ready
- kind: decomp
- source: backlog:SKMELEE2
- question: game.exe adds attack rating terms for the stats 115, 116, 120, 123, 124 and 179 (FUN_0057d8a0 and FUN_0057b7d0); d2d has no input for them. What does each term do (percent or flat, which target kinds, order against the other terms), and how many gear-built characters would a to-hit roll differ for? Run FUN_0057d8a0 natively with its boss tests stubbed, as the existing melee harness does for the other parts.
- start: docs/research/sk-melee2-trace.md (row D3 of the deviation table); docs/research/sk-melee-trace.md Q2; tools/emu/sk_melee2.py for the harness pattern.
- deliverable: docs/research/sk-gear-ar-trace.md in the trace style (rules with address tags, V read / V run / S / U, deviations, bot findings); an oracle script under tools/emu/ if you ran game code; 100,000 cases or exhaustive boundaries for any number you call V.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout, Python 3 with unicorn.

## SKSHIELD: does a shield's damage reach the player's stats
- size: S
- status: ready
- kind: decomp
- source: backlog:SKMELEE2
- question: Does a carried shield add its base damage, or any enhanced damage, to the player's damage stats 21, 22, 17 and 18 in game.exe (the carrier flag 0x40000000 on a child list, FUN_00624fe0 and FUN_00627910), or only a weapon in the active hand? Does a barehanded player with a shield in the off slot get the shield's stats? d2d ignores all of it.
- start: docs/research/sk-melee2-trace.md Q3 (the shield question, marked I there); docs/research/sk-melee-trace.md Q3; docs/research/paladin-act1-recon.md section 6 row 4 and the slot function FUN_0063bef0.
- deliverable: docs/research/sk-shield-damage-trace.md; a static answer from the equip path (tagged V read) is a valid first PR, a run of the stat-list merge natively (tagged V run) is better.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout, Python 3 with unicorn.

## SKCURSEVAL: the curse value calculations on game.exe against d2d
- size: S
- status: ready
- kind: decomp
- source: backlog:SKRESIST
- question: Curse values come from skill calcs (Lower Resist `dm56`, Conviction `ln34` capped at 150, Amplify Damage and others). Does d2d's shared calc evaluator give the same value as game.exe's for every level from 1 to 99, including the synergy terms and the cap? Compare the two evaluators on the same calcs and levels.
- start: docs/research/sk-resist-trace.md section 10 (not traced: the curse value calcs) and Q7; docs/research/sk-passives-trace.md for synergies; upstream d2decomped `eval_calc` (public).
- deliverable: docs/research/sk-curse-calc-trace.md with the table of calcs, levels and cases compared (counts and seed), plus an oracle script under tools/emu/ if you ran game code.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout, Python 3 with unicorn.

## SKDEFENDERS: hirelings and players as defenders against resistances
- size: M
- status: proposed
- kind: decomp
- source: backlog:SKRESIST
- question: A hireling or a player that is hit gets the resistance penalty and the cap, and a curse or aura on it applies in full (no one-fifth rule). Trace the full rule for these defenders: which function decides "plain monster", the Normal column read for hirelings, absorb, and the poison per-tick path. Does d2d match, and does the bot's reading of a mercenary's resistances need a change?
- start: docs/research/sk-resist-trace.md section 4 limits (hireling defenders and players as defenders; FUN_0063ee90 is the plain-monster test); docs/research/pets-merc-trace.md; docs/research/d2d-fidelity-riskmap.md.
- deliverable: docs/research/sk-defender-resist-trace.md; oracle script under tools/emu/ if you ran game code; 100,000 cases for any number you call V.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout, Python 3 with unicorn.

## SKIAS: swing speed, increased attack speed and the weapon speed modifier
- size: M
- status: proposed
- kind: decomp
- source: backlog:SKMELEE+
- question: How many frames does one melee swing take in game.exe for a given weapon speed modifier, increased attack speed gear, a skill speed bonus and a class animation table, and what does d2d's `attack_speed` and `attack_ticks` compute for the same inputs? Cover the breakpoints for any class (class-generic: pick the animation tables from your own data).
- start: docs/research/sk-melee-trace.md (Not covered: swing speed); docs/research/paladin-act1-recon.md section 6 row 4 (the recon measured only the swing length); docs/research/walk-speed-trace.md for the velocity and frame-rate style.
- deliverable: docs/research/sk-attack-speed-trace.md; an oracle script under tools/emu/ if you ran game code; exhaustive boundaries or 100,000 cases for any number you call V.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout, Python 3 with unicorn.

## SKZEROHIT: what a zero-damage hit still triggers
- size: M
- status: proposed
- kind: decomp
- source: backlog:SKRESIST
- question: A hit that deals no damage (an immune target, a full block, a rounded-down roll) may still wake or alert the monster, fire on-hit effects of enchanted monsters and apply knockback. Which of these does game.exe do, in which function, with which condition, and does d2d match?
- start: docs/research/sk-resist-trace.md section 10 (first bullet) and Q9; FUN_0057cb00 (the get-hit gate, S: docs/research/pr-get-hit-rule-proof.md); docs/research/mon-wake-trace.md for the wake rules.
- deliverable: docs/research/sk-zero-hit-trace.md; a static answer (tagged V read) is a valid first PR, a native run (V run) is better.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout, Python 3 with unicorn.

## SKSTATEEV: the state events fired on a hit (FUN_005d3ac0)
- size: L
- status: proposed
- kind: decomp
- source: backlog:SKARMOR
- question: FUN_005d3ac0 walks a unit's states and fires the events their stat lists name (attacked in melee, absorb damage, on-hit effects). Which events fire on a hit and on a miss, in which order against resistances and other absorbs, and where are events 3 and 11 fired? This settles Shiver Armor on a miss and the place of Energy Shield against resistances; advanced item, claim it only if you have done a decomp before.
- start: docs/research/sk-armor-trace.md section 11 (Not traced) and Q3, Q5; docs/research/sk-melee-trace.md Q1 (the pre-events row); docs/research/sk-melee2-trace.md (Not covered).
- deliverable: docs/research/sk-state-events-trace.md in the trace style, plus an oracle script under tools/emu/ if you ran game code.
- needs: your own game.exe 1.14d install, upstream d2decomped checkout, Python 3 with unicorn.
