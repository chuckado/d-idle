# Pets, the merc and monster targeting: game.exe 1.14d vs d2d vs our bot (agent DECOMPPETS, 2026-10-31)

> status: research, current (card decomppets, decomp-queue item 8; scope beyond PR #134's get-hit rule, docs/research/pr-get-hit-rule-proof.md). Offline only: no game instance, no d2d slot; nothing pushed, posted or filed. d2d read at upstream main 72132b99 (own read worktree on the worker host). Walk / run speed of mercs and pets (merc.cpp ~286, pets.cpp ~217 / 270 still at MonStats Run) is NOT redone here: docs/research/walk-speed-trace.md and the WALKFIX card. Tags: V = run in game.exe's own code under unicorn (by us, this card) or read in the disassembly of the read-only copy (md5 2b3600c1...); V-d2d = read or run in d2d's source; S = upstream d2decomped docs (`docs/research/re/monster-ai.md`, `merc-ai.md`, `pet-ai.md`) or ours; D = game data; I = inference; U = unverified.

## Answer in four lines

1. **A monster goes for whoever is strictly nearest, the player first.** No threat, damage or "aggro" score for an evil monster: the player, then its pets (the merc is one) newest first, each taken only when its AI distance is under the best so far; ties go to the player. The player at 55 subtiles or more (or in town, or another act) drops the player AND its pets out of the search. A dead player is skipped, its pets are still tried. (V)
2. **The merc leashes before it fights.** Every think checks the owner first: over 100 off a teleport into the owner's room, over 24 a run back, over 16 a follow when the owner walks or runs; only then a foe under 25 off (nearest threat >= 2 in sight under 49) is attacked. The Rogue (a shooter) backs off 4 from a foe under 4 off half the time; no life test anywhere, so a merc never retreats on low life. It does not react to the player's attack; it acts at its next think, at most 10 frames (0.4 s) after a stand, at once when a walk ends. (V for the functions the oracles run, S for the rest of the reading)
3. **d2d is ported and close.** Our differential against d2d's own C++ rules: the merc's target (mode 6) 0 of 100,000 differ; the monster's search 1,254 of 100,000 differ, every one a tie between pets: game.exe tries the newest pet first, d2d the merc first and then the oldest. No effect for the Sorceress (no summons); a PR candidate is not built (the after side is a reordered harness, not a patched d2d build; see D1).
4. **The bot's merc model over-credits the Rogue as a tank.** The route guard and the Andariel floor drop by 2 with a living merc, and the immune workaround says "leave the kill to the merc", but a Rogue draws a monster only when nearer to it than the bot and never stays behind once the bot is 24 subtiles away. Findings B1-B5; the drills below measure it.

## Q1 monsters, pets and the target (game.exe)

| Address | Rule | Tag |
|---|---|---|
| `005dd7f0` (driver `005b1740` step 2) | evil monster: skill-set target first (Attract / Confuse only), then lists 0..7 (one per player: the player at the head, its pets after it), list 8 (good monsters), list 9 (neutral) | V (search.py 100,000 native, ok; S monster-ai.md) |
| `005b1880` / `005b1900` | a player heads a free list; each pet (summons `00573270` / `005c4b00` and the merc) is inserted right AFTER the head (0x5b1966..0x5b1979), so the list is player, newest pet, ..., oldest pet | V (disassembly read; run natively by pets_merc.py) |
| `005a6520` -> `005b1a90` | a dying monster leaves its list and loses `+0xc4` bit 2 (targetable): a dead merc or pet is never a candidate | V (read: 0x5a6582..0x5a658f); called from the death paths 0x5a7019 / 0x5a73a6 (I) |
| `005dd7f0` lists 0..7 | the best starts at MonStats aidist (blank: 35); the player is skipped (with its pets) in another act, with no room, in a town room or at AI distance >= 0x37; a dead player counts 0x7fffffff (its pets still tried); each candidate is taken when its distance `(min + 2 max) / 2` is under the best and sight passes or isn't needed; pets have no act / room / town / 0x37 / death test | V (pets_merc.py S: 100,000 native) |
| sight | needed only in a preset room whose LvlPrest isn't Outdoors, and only until the monster's first target (AI flag 8, never cleared) or with flag 0x40; outdoors every monster "sees" the pets and the player | S monster-ai.md, V (search.py) |
| threat | MonStats +0x4e (a player 14) only splits primary (>= 2) from secondary for good / neutral searchers (mode 5) and the merc's search (mode 6); evil monsters ignore it | V (search.py, pets_merc.py M) |
| the attack's target | set when the think starts the mode (`005a7c20`, path +0x58); a hit goes at that unit, not at whoever is nearest at the swing; each think searches again | S monster-ai.md |
| pets' foes (NecroPet `005e4cf0`) | a pet searches as a good monster (mode 5): any enemy, i.e. any hostile monster near it, not "the owner's target"; kept if within 6, else the first under 0x24 with a clear line (`005dc640`); follow-or-fight `005e45d0`: over 0x50 from the owner with a foe teleports back | V (necropet.py 100,000 native, ok); S pet-ai.md |
| pets' leash | gap to owner over 50: teleport into the owner's room; 29 or more: walk at Run x 100 / Velocity - 100 pace; the other pets (Raven, wolves, bear, totems, vines, Shadows) per pet-ai.md | V (pet_ais.py 13 AIs x 8,000 native, ok) |
| get-hit / knockback | #134's get-hit rule covers the merc (its own seed) and monsters; pets hurt by monsters fall back to class 0 (divisor 16) in d2d; the merc is knocked back as a monster (size class) | S pr-get-hit-rule-proof.md limits; V-d2d tick.cpp:211 |

## Q2 the merc (MonAI 61 Hireable, think `005e52d0`, attack `005e5050`, skills `005e4d30`, target `005ddc30`)

| Rule | Tag |
|---|---|
| in a walk or run it does nothing; distance to the owner `005dc380` (gap less its size): over 100 teleport (`005e3930` mode 3), over 24 run (mode 1, pace 60), over 16 with the owner walking / running follow (mode 0); THEN out of town a foe under 25 off goes to the attack think | V (merc.py 100,000 native, ok); S merc-ai.md |
| its foe (`005ddc30` mode 6): the near rooms' units, an enemy under 0x31 in sight (always tested): nearest threat >= 2 primary, else secondary; a secondary under 6 with no path to the primary takes mode 7's nearest other primary within 0x13 | V (pets_merc.py M, search.py mode 6) |
| attack think: skill chance 98 for the melee mercs, else 40 + 2 x level + 10 per miss, at most 95; aip1 0 (the shooters: Rogue) under 4 off backs off 4 half the time, else the skill pick; aip1 != 0 (Act 2) walks to melee | V (merc_attack.py 100,000 native, ok); S merc-ai.md |
| skill pick: hireling.txt DefaultChance then Chance + ChancePerLvl per skill; the Rogue's default is MonStats Skill1 in Sk1mode (its plain shot) | S merc-ai.md (values not re-read: D left open) |
| no life test in the think or the attack think: the merc never retreats on low life; it only follows the owner | V (read of the ported thinks: merc.py / merc_attack.py cover every branch) , I for "never" outside them |
| reaction: the think runs at the think event: a stand waits n frames (5 or 10 here, 40 ms each), a walk's end thinks at once (`005a8030`); there is no hook from the player's attack | S monster-ai.md driver; V-d2d merc.cpp:290, :353..356 |
| death: the merc leaves the player's list (`005a6520`), monsters switch to the player; revive at the hireling NPC (bot D8.a PASS) | V (read), S |

## Q3 d2d's version and the differential

d2d ports all of it (V-d2d): ai.cpp:508..570 `search_target` -> `rules::search_pick` / `search_near`; fight/tick.cpp:263..273 builds the foes per player as player, merc, then `actor->pets` (push_back, oldest first, summons.cpp:258); fight/merc.cpp:294..330 the merc's mode 6 (`search_sight`, `sight_choice`), :331.. `hireable_think`, `merc_attack`; fight/pets.cpp the pets' thinks. Upstream checks game.exe against Python line-for-line copies; `tests/monsters/test_monster_ai.cpp` replays 16..40 of them through the C++.

Our oracle `tools/emu/pets_merc.py` (in upstream's tools/emu) runs game.exe's `005dd7f0` (S: one player, the merc, 0..3 summons; the list built by game.exe's own `005b1880` / `005b1900`) and `005ddc30` (M: a merc and 0..8 monsters in its room and a town room) natively; `tools/emu/pets_merc_harness.cpp` runs d2d's real `rules::search_pick` / `search_sight` / `sight_choice` from components/rules/monsters.hpp on the same cases with ai.cpp's and merc.cpp's input mapping. Seeds 21..24, 25,000 each; distances biased to the edges (0x37, aidist +-1, 6, 0x13, 0x31, ties).

| Kind | cases | found | d2d before | after (newest pet first) | what differs |
|---|---|---|---|---|---|
| S monster's search (player, merc, pets) | 100,000 | 52,033 | **1,254** | 0 | all 1,254 equal-distance ties between pets; 780 of them d2d picks the merc where game.exe picks a newer summon; 0 involve the player |
| M merc's target | 100,000 | 56,664 | 0 | 0 | none: clean |

Upstream's own oracles re-run at scale by us (game.exe vs their copies): search.py 100,000 ok, merc.py 100,000 ok, merc_attack.py 100,000 ok, necropet.py 100,000 ok, pet_ais.py all 13 AIs x 8,000 ok.

```sh
# the worker host, in the upstream checkout's tools/emu (game.exe read-only copy as $D2_MPQ_DIR/bin/game.exe)
python pets_merc.py 25000 21 c21.txt > g21.txt     # and seeds 22..24; ~2 min each
g++ -O2 -std=c++23 -I<d2d>/components $(find <d2d>/components -name '*.hpp' -printf '-I%h\n' | sort -u) pets_merc_harness.cpp -o pm_h
cat c2?.txt > c.txt; cat g2?.txt > g.txt
./pm_h before < c.txt | diff g.txt - | grep -c '^>'   # 1254
./pm_h after < c.txt | diff g.txt -                    # nothing
```
Checksums (md5 prefix): cases b4a2be90, game 168c5cf7, before 33fcbe0f, after 168c5cf7.

| id | deviation (d2d vs game.exe) | where | evidence | tag |
|---|---|---|---|---|
| D1 | the player's pets are tried merc first then oldest first; game.exe newest first (the merc is re-inserted when hired, loaded or revived, so after a revive it is the newest) | fight/tick.cpp:263..273, summons.cpp:258 | 1,254 of 100,000 -> 0 (rule level) | V |
| D2 | mercs and pets move RN at MonStats Run | merc.cpp:286, pets.cpp:217 / 270 | walk-speed-trace.md; WALKFIX card | V (not redone) |
| D3 | the merc's auras count as running but do nothing | upstream merc-ai.md "Not built yet" | S | S |
| D4 | a Hydra / trap-like pet (no Velocity, no Run, ranged) is never a monster's foe in d2d | pets.cpp:83..84 | game.exe side not traced | U |
| D5 | a merc's chase ends when within `reach` (hypot) of the foe; game.exe walks the path to its end | merc.cpp:277..283 | read only | I |

PR candidate: NOT built. The proof above is rule level (the "after" reorders the harness's input, not a patched `Fight::world`); a faithful patch needs a per-player insertion order (the merc re-inserted on hire / load / revive), a `test_game` mode to run it, and the shadowing / clang-tidy sweeps. Low value for our bot (no summons) and small: worth a bundle slot only with a Necromancer or Druid plan.

## Q4 our bot's model

| id | assumption or code | rule says | tag | finding |
|---|---|---|---|---|
| B1 | route guard and Andariel floor 2 lower with a living merc (merc-policy spec, ADR 0011); SME: "casters want a tank" | a monster goes for the merc only when the merc is strictly nearer than the bot (ties: bot); a Rogue shooter keeps 16-24 subtiles of the bot and backs off from a foe under 4 off half the time: it rarely holds a monster; the Act 2 melee merc walks into melee and does | V rules, effect I | the -2 is a judgment on damage share, not aggro; the act1run soak's merc-alive deaths decide (soak gate), do not assume the Rogue tanks |
| B2 | immune workaround `Merc: true` ("leave the kill to the merc / pets", internal/combat/rules.go:203, :276); the target is then ignored (fight.go:539..541) | the merc attacks only after its leash checks: the bot moving away more than 16 (walking / running) or 24 subtiles pulls it off the kill | V rules, I effect | the workaround only logs; if it is ever wired, the bot must stand still within 16 subtiles of the merc while it shoots |
| B3 | SME rule 9: wait until the merc is next to the player after TP / WP (10 s) | the merc teleports into the owner's room when over 100 off and runs when over 24; area changes (FUN_0061b130) make it follow at pace 60 | S/V | the wait can end when the merc is within 16 subtiles; no need for 10 s |
| B4 | merc potions at 40 % life (merc.potionLifePct) | the merc never retreats on low life; the bot's potion is its only life control apart from walking away (the leash) | V (ported thinks) | keep; consider walking off 25+ subtiles as a merc "retreat" lever (U value) |
| B5 | bot safety counts monsters near the player (crowd retreat) | a monster on the merc 55+ subtiles from the bot drops the merc and wanders; one within 55 switches to the bot the moment the bot is nearer | V rule, I | conservative; nothing to change; drills should log the target switch |

Not found in the bot: any model of which unit a monster targets, of the merc's reaction time, or of the 55-subtile cut (grep internal/combat, internal/merc, internal/d2dcombat).

## The answer the drills should expect

- Monster, bot and merc in range: the monster attacks whichever is strictly nearer by `(min + 2 max) / 2` subtiles; equal: the bot. Bot 55+ subtiles from that monster: the monster ignores the merc too.
- Merc with a foe under 25 subtiles: the first Rogue shot within 0.4 s of game time (10 frames) if it was standing, sooner at a walk's end; never before its leash checks.
- Bot walks 17+ subtiles away (walking or running): the merc stops shooting and follows; 25+: it runs; 101+: it appears in the bot's room.
- Merc dies: monsters on it switch to the bot at their next think.

## DRILL RECIPE (hidden d2d instance, throwaway character, ledgered cheats only)

Fork debug commands by name only: `spawn`, `merc`, `hit`, `give`.
1. T1 aggro split: `merc hire` (Rogue), stand still in an outdoor level; `spawn` one zombie at 10 subtiles from the bot and 5 from the merc; log each think's target (merc life drops, bot life flat). Repeat with the zombie equidistant (bot targeted) and with the bot 56+ subtiles away (zombie idles or wanders, merc untouched until it shoots). Observables: who loses life first, the zombie's facing.
2. T2 merc reaction: `merc` life full, bot standing; `spawn` a fallen at 20 subtiles of the merc; count ticks from spawn to the first arrow (expect <= 10 frames after the merc's current stand; log the merc mode per tick).
3. T3 leash: during T2 walk the bot 20, then 30, then 110 subtiles away; observe merc mode WL at 17+, RN at 25+, a room teleport at 101+; it stops shooting while following.
4. T4 back-off: `spawn` a melee monster next to the Rogue (under 4 subtiles): about half its thinks step back 4 subtiles.
5. T5 death switch: `merc hurt` / `hit` the merc to death with a monster on it; the monster's next target is the bot; revive via Kashya (D8.a path).
6. T6 (Necromancer only, D1): two skeletons and the merc equidistant from a monster: game.exe picks the newest skeleton; d2d today picks the merc.
Pass rule per step: two identical runs; leave the character alive in town, no panels open.
