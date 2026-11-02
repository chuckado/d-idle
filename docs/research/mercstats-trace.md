# Hirelings (Act 1 Rogue archers, Act 2 Desert Mercenaries): game.exe 1.14d vs d2d vs our bot (agent MERCSTATS, 2026-11-02)

> status: research, current (card mercstats; board item SKDEFENDERS in part, see Q7). Offline only: no game, no d2d slot; nothing pushed, posted or filed. Tags: V run = game.exe's own code run under the emulator on the worker host against d2d's real C++ rules or an oracle (counts below); V read = read in the disassembly of the read-only copy (address and instructions named); D = from the patched hireling.txt / Skills.txt / ItemTypes.txt (read with `mpq-table.sh`); S = an existing upstream or earlier proof (named); I = inferred; U = unverified. d2d rules: the scratch upstream checkout the harness links (main 7d5b3a5c plus shard commits); its fight code read in that checkout and in the older reading checkout. Earlier work this builds on: `pets-merc-trace.md` (merc target, leashes, 0 of 100,000), `sk-resist-trace.md` Q4 / Q10 (the defender resistance function), `monster-stats-trace.md` Q9 (life regeneration of monsters). Walk speed of the merc: `walk-speed-trace.md` (not redone).

## Answer in eight lines

1. **Stats by level are exact in d2d.** game.exe's level-up `FUN_00572840` sets Str, Dex, life, defence, damage, attack rating and the four resistances from the hireling.txt row of the merc's level; d2d's `merc_stats` and `merc_skill_levels` equal it in all 2,940 (Id x level) cases (30 Ids x levels 1..98). (V run)
2. **The hire price is `Gold x (15 d + 100) / 100`** (d = offered level - the row's base level, never under Gold), the offered level `clvl - 5 + rand(5)` (at least 2); d2d equals game.exe in 108,000 offers (3,000 seeds x 3 acts x 3 difficulties x 4 character levels), and so does our bot's formula. The **resurrection price is `(level^2 / 2) x 15`, at most 50,000** (reached at merc level 82): 256 levels equal in d2d, bot formula equal. (V run)
3. **A merc levels with its owner's kills**: it gets 2 x the kill's experience at its own level (x 86 / 256 unless the merc struck the blow), only while its level is under its owner's, up to level 98; d2d's `merc_kill_exp` + `merc_stats` equal game.exe's `FUN_0057e860` in 100,000 cases (the 86/256 scaling is V read only). (V run, V read)
4. **AI: d2d's think and attack code equal game.exe in 100,000 cases each** (follow at 16 / 24 off, teleport over 100, the foe under 25, backing off for the Rogue, the skill roll). Neither think reads life: a merc **never flees on low life**. (V run, V read)
5. **Skills are gated by Skills.txt `reqlevel`**: a Cold Rogue has no Cold Arrow below level 6, a Defiance Mercenary no Defiance below level 6 (Blessed Aim 12, Holy Freeze 18). Offers at character level 8 / 9 are levels 3..7 / 4..8, so many are without the skill that names them. (V run)
6. **Two real gaps in d2d, both effect-U:** the merc's own life regeneration (stat 0x4a, set at every level-up) and the Act 2 auras (they count as running but do nothing). No PR candidate.
7. **The merc accepts Rejuvenation Potions** (both classes pass the healing-potion test of `FUN_0054d230`), our bot's comment says it does not. Mana potions are refused. (V run)
8. **Bot findings B1-B6** below (offer level ignored, Rejuvenation, aura order without a number, ...).

## Q1 stats by level (V read FUN_00572840, V run 2,940 cases)

`FUN_00572840(ECX game, EDX owner, [8] merc, [c] level)`, row = `FUN_006562f0(expansion, Id, level)` (0x6562f0..0x656389: the LoD rows of the Id in file order, the first always, a later one while its Level is not over the merc's; an Id of 0x100 or more gives none). With d = level - row.Level (negative below the first row):

| stat | rule (0x572904..0x572af7) | address |
|---|---|---|
| 0 Str, 2 Dex | max(10, base + (per-level x d) / 8, to zero) | 0x572961, 0x572985 |
| 7 max life, 6 life | max(40, HP + HP/Lvl x d), x 256 (0x2800 = 40 x 256) | 0x5729ac..0x5729dc |
| 31 defence | max(0, Def + Def/Lvl x d) | 0x5729e1 |
| 0x17 / 0x18 damage | max(0, Dmg-Min + Dmg/Lvl x d / 8), max(1, Dmg-Max + Dmg/Lvl x d / 8) (to zero) | 0x572a05..0x572a51 |
| 19 attack rating | max(0, AR + AR/Lvl x d) | 0x572a56 |
| 0x27 / 0x29 / 0x2b / 0x2d resists | all four: max(0, Resist + Resist/Lvl x d / 4) (to zero); difficulty penalty and cap come later, in the hit (Q7) | 0x572a77..0x572abd |
| 0x4a (74) | stat 7 (256ths) / 2000 | 0x572ac2..0x572af2 |
| skills 1..6 | for each skill that exists and has `reqlevel` <= level (Skills.txt `+0x174`, `FUN_0045c4b0`): Level + (LvlPerLvl x d >> 5) (floor), 1..32; a skill id under 1 or a Mode over 15 ends the list | 0x572af7..0x572b88 |

Run: `tools/emu/mst_oracle.py levelup` calls the real function on a hooked unit (stat setters recorded, the row and the skill table game.exe's own) for every LoD Id x level 1..98; `tools/emu/mst_d2d_test levelup` prints d2d's `rules::merc_stats(exp of that level)` and `rules::merc_skill_levels` (names through d2d's Skills table, `reqlevel` from it). **2,940 cases, 0 differ** (str, dex, life, def, damage, AR, resist, six skill levels, level from exp). Sample (game.exe = d2d; Rogue Id 0, Desert Id 6):

| merc | level | life | str / dex | defence | damage | AR | base resist | regen stat | skill levels |
|---|---|---|---|---|---|---|---|---|---|
| Rogue Fire | 5 | 63 | 37 / 49 | 31 | 1-3 | 34 | 4 | 8 | Inner Sight 1, Fire Arrow 1 |
| Rogue Fire | 15 | 153 | 50 / 69 | 111 | 4-6 | 154 | 24 | 19 | 4, 4 |
| Rogue Fire | 30 | 288 | 68 / 99 | 231 | 7-9 | 334 | 54 | 36 | 9, 9 |
| Rogue Fire | 50 | 594 | 94 / 139 | 489 | 16-18 | 742 | 90 | 76 | 16, 16 |
| Desert | 9 | 120 | 57 / 40 | 45 | 7-14 | 20 | 18 | 15 | Jab 3, Prayer 3 |
| Desert | 15 | 210 | 67 / 49 | 111 | 10-17 | 92 | 30 | 26 | 4, 4 |
| Desert | 30 | 435 | 93 / 71 | 276 | 17-24 | 272 | 60 | 55 | 9, 7 |
| Desert | 43 | 630 | 117 / 91 | 419 | 24-31 | 428 | 86 | 80 | 14, 11 |
| Desert | 75 | 1,430 | 173 / 139 | 1,027 | 48-55 | 1,196 | 142 | 183 | 24, 18 |

(life, Str, Dex, defence, AR, resist, regen: before gear. The Rogue's damage is added to a weapon only when it is two-handed, i.e. the bow: `hireling.md`, S; the merc's AR in a fight is `stat 19 + dex x 5` scaled, the defence `stat 31 + dex / 4`, S.) Resist grows 2 a level for the Rogue (Resist 0, Resist/Lvl 8), the Desert Mercenary starts at 18 (2 a level); the cap 75 + gear max (at most 95) and the difficulty penalty (Nightmare -40, Hell -100) apply on every hit.

## Q2 hire offer, cost, resurrection (V run 108,000 + 256)

- The offer (`FUN_006637f0`, 0x6637f0..0x663a06): `FUN_00656580` lists the rows of the act, difficulty and version that share the first row's Level; one is picked by `rand(n)` on the entry's seed {seed, 666}; level `max(2, clvl - 5 + rand(5))` (0x663895..0x6638c0, `x % 5` by the magic multiply); d = level - row.Level; life `max(40, HP + HP/Lvl x d)`, Str / Dex `max(10, base + (per x d >> 3))` (floor, unlike the level-up's truncation: a level under the row's Level, only the first row, rounds one point differently: kept as game.exe does), cost `(15 d + 100) x Gold / 100` (truncated) at least Gold, experience `(level + 1) x Exp/Lvl x level^2` (0 if negative), defence, damage like the level-up. (V read, V run)
- Run: `mst_oracle.py hire 1001-4000` (the same `hire.dump` as upstream's `hire.py`, my seeds) against `mst_d2d_test hire` (d2d's real `rules::merc_offer`): **108,000 offers (3,000 seeds x 3 acts x 3 difficulties x 4 character levels 1 / 9 / 30 / 75), 0 differ**. Seeds 1..20 (720 more) the same.
- What it gives (Normal, from the run; Gold: Rogue Fire 100, Rogue Cold 150, Desert 350, all rows in `hireling.txt`):

| clvl | Rogue offered levels | Rogue cost | Rogue life | Desert levels | Desert cost | Desert life |
|---|---|---|---|---|---|---|
| 9 | 4..8 | 115..262 | 54..90 | 4..8 | 350 (flat: under the row's Level 9) | 45..105 |
| 30 | 25..29 | 430..735 | 243..279 | 25..29 | 1,190..1,400 | 360..420 |
| 75 | 70..74 | 1,105..1,747 | 648..684 | 70..74 | 3,552..3,762 | 1,035..1,095 |

- The merc is created from the entry's seed and re-rolled with `FUN_006637f0` at the claim (0x573343..0x573376; the owner's level at that moment sets the level), its experience set to the offer's and `FUN_00572840` run for that level (0x5733d6..0x5733f3). A merc loaded from a save (0x56aa88..0x56abc2): level = the highest whose `(L + 1) x Exp/Lvl x L^2` is at most the saved experience, one `FUN_00572840` call (V read; the bound of the loop is I, 98 by the XP function below).
- Resurrection `FUN_006637b0` (0x6637b0..0x6637de): `(L x L / 2) x 15`, unsigned minimum 0xc350 = 50,000, L = stat 0xc (V read). Run: `mst_oracle.py revive` (hooked stat) vs `rules::merc_resurrect_cost`, levels 0..255: **256 equal**. Examples: level 9 600, 15 1,680, 20 3,000, 30 6,750, 50 18,750, 81 49,200, 82 and up 50,000. Pay: carried gold, then stash (S, `npc-menu.md`); who revives: Kashya, Greiz, Asheara, Tyrael, Qual-Kehk (S).

## Q3 levelling with the player (V read FUN_0057e990 / FUN_0057e860, V run 100,000)

- The split is in `FUN_0057e990` (owner's kill share): the merc's share is `FUN_0057e480(victim experience)` with the merc as the receiver at ITS level (the player's own experience function: level gap, table, bonus), then `x 0x56 / 256` (to zero) unless the merc struck the blow (0x57ea28..0x57ea39) (V read). `FUN_0057e860` (0x57e860..0x57e983) then: amount over 0; the merc's level under its owner's (stat 0xc of the owner, `cmp esi, eax; jge`); level under `FUN_00611830(0) - 1` = 98 (the table at 0x96c8a8 holds 99, V run); experience += **2 x amount** (`lea ecx, [esi + ebx * 2]` at 0x57e8ce); the level loop (`FUN_00663790(L + 1, Exp/Lvl)` at most the new experience, up to 98) and `FUN_00572840` for the new level (V read).
- Run: `mst_xp.py` hands the real `FUN_0057e860` random cases (merc Id, level, experience inside the level, award with the edge values, merc struck or not, owner level around the merc's) and prints the new experience and level; `mst_d2d_test xp` replays them through d2d's `rules::merc_kill_exp` + `merc_stats`. **100,000 cases (seeds 51..54 x 25,000), 0 differ.** The 86/256 step is done in Python (V read, not run), `FUN_0057e480` is upstream's F10 oracle, not run here.
- Thresholds (Exp/Lvl: Rogue Fire 100, Rogue Cold 105, Desert 110): experience for level L is `(L + 1) x Exp/Lvl x L^2`: Rogue Fire level 10 110,000; 20 840,000; 30 2,790,000; Desert level 20 924,000; 43 8,949,160 (the `next` field of the run holds the step to the next level).
- d2d: the newer checkout wires it (`fight/deaths.cpp:256`); the older reading checkout (`components/game/fight.cpp`, no `merc_kill_exp`) never writes `merc_exp` after a hire. Say which build you test.

## Q4 the AI (think `FUN_005e52d0`, attack `FUN_005e5050`, picks `FUN_005e4d30`; V run 100,000 + 100,000)

- Think (rules from `pets-merc-trace.md` Q2 and upstream `merc-ai.md`, S): in a walk or run nothing; gap to the owner (`FUN_005dc380`: per axis the gap less the merc's size, (short + 2 x long) / 2): over 100 teleport into the owner's room, over 24 run at pace 60, over 16 follow (owner walking, or running at pace 60), then not standing -> stand 5; collision bit 0x40 wander 5 (6 in 128, 12 for the bow Rogue); out of town a foe under 25 -> attack think; another spawn area -> follow; within 1 of the owner step to a spot 4 about him; else 5 in 100 a spot 16 about him; else stand 5.
- **Run (new, the C++ that d2d really runs):** `mst_think.py` (upstream `merc.py`'s setups and hooks) writes 25,000 cases per seed with game.exe's effects (every move tried with its path ctx type / pace / steps, stand, teleport, attack hand-off, the seed after); `mst_d2d_test think` runs d2d's real `rules::hireable_think` on the same lines. **100,000 cases (seeds 31..34), 0 differ** (seed 31: about 6 % teleports, 7 % attack hand-offs, 19 % stands of 5 frames, 39 % first move). Upstream's own check was game.exe against a Python copy; this is the C++.
- Attack think (V run): `mst_attack.py` (upstream `merc_attack.py` cases: any Id, level 1..98, its row, skills at random levels, auras running, gap 0..20, melee or not, the AI control's +0x14) vs `mst_d2d_test attack` (real `rules::merc_attack` with the pick). **100,000 cases (seeds 41..44), 0 differ** (skill, aura then stand 10, swing, stand, moves, +0x14 and the seed after).
- The roll (from the equal rules): chance to roll at all `min(growth + 40 + 2 x level, 95)` % per think for the Rogue (growth +10 per miss, reset on a roll), 98 % for the Desert Mercenary; then `rand(total + 1)` over DefaultChance (75 Rogue, 30 Desert) plus each skill's `Chance + ChancePerLvl x (level - row Level) / 4`. Exact shares (`mst_picks.py`; D + run rules):

| merc, level | plain shot / swing | Inner Sight or aura | the arrow or Jab | thinks that roll |
|---|---|---|---|---|
| Rogue Fire 5 | 66.4 % | IS 9.7 % | Fire Arrow 23.9 % | 56 % |
| Rogue Fire / Cold 15 | 59.5 % | 8.7 % | 31.7 % | 73 % |
| Rogue Fire / Cold 30 | 52.1 % | 7.6 % | 40.3 % | 95 % |
| Rogue Fire / Cold 36 and up | 49.0 % | 7.2 % | 43.8 % | 95 % |
| Rogue Cold 5 | 87.2 % | IS 12.8 % | no Cold Arrow (reqlevel 6) | 56 % |
| Desert (Prayer) 9 / 15 / 30 | 27.0 / 25.6 / 22.7 % | aura 9.0 / 8.5 / 7.6 % (once, then left out) | Jab 64.0 / 65.8 / 69.7 % | 98 % |
| Desert (aura running) 15 | 28.0 % | - | Jab 72.0 % | 98 % |

  The Rogue's "plain shot" is MonStats `Skill1` in `Sk1mode` (S, `merc-ai.md`); the Rogue stands 10 frames (0.4 s) on a think that does not roll. Averages of damage are not here (missiles and the hit pipeline: `sk-missile-trace.md`, `monster-attack-trace.md`).
- **No flee, no life test (V read):** `FUN_005e52d0` calls no stat reader at all (its callee list, 0x5e52d0..), `FUN_005e5050` one (`FUN_00625480`, stat 0xc at 0x5e508e: level), `FUN_005e4d30` one (stat 0xc at 0x5e4d75). A hurt merc keeps fighting; only the leash moves it.
- Target (`FUN_005ddc30` mode 6), leash order, the 55-subtile cut of monsters: `pets-merc-trace.md` (0 of 100,000, S).
- Teleport: `FUN_005e3930` mode 3 puts the merc in a spot of the owner's room (`FUN_0054dc40`, up to 20 tries), standing 5 (S, `merc-ai.md`; the hand-off is V run in the 100,000).

## Q5 gear and slots (S upstream V read; the filter V run)

- Wearing (`FUN_0054d230` / `FUN_0048b290`, S `hireling.md`): identified, not broken, helm or body armour; per class: Rogue 271 a bow, Desert 338 a spear or polearm, Iron Wolf 359 a shield or one-handed sword, Barbarians 560 / 561 an axe / sword (+ primal helm); the merc's own Str, Dex and level against the item's; slots head, torso, left (shield), right (weapon). d2d: `rules::merc_slot` (the newer one has `meets_stats`).
- Stats from gear join the base as for any unit (S `hireling.md`; d2d `merc_fighter` / `merc_gear`, read): damage from a two-hander adds the base (0x17 / 0x18), a one-hander uses none of it; AR, defence, resists with penalty and cap, life `base + items' maxhp + hp per level (216) / 8 x level + maxhp% on the base`; block none (only an Iron Wolf holds a shield). Not run here (no oracle for the merged stat list).

## Q6 potions (V read + V run)

`FUN_0054d230` (0x54d230) accepts a hand-over when `FUN_00629bb0(item, type)` holds for 0x4c, 0x51 or 0x50; the function reads game.exe's item type matrix (the closure of ItemTypes Equiv1 / Equiv2), so a child type passes. **Run (`mst_potion.py`): over all 659 item classes**, type 0x4c (healing) passes 11 classes, including the two Rejuvenation Potion classes (515, 516, which are exactly the classes of type 0x4e); 0x50 (antidote) 514; 0x51 (thawing) 517; every mana potion class (510, 512, 592-596, 608, 609) and the stamina potion (513) pass none of the three. So **Rejuvenation Potions are accepted, Mana and Stamina are not**; the merc then drinks through the item's own use function on its unit (`FUN_005bf240`, V read; the heal amount on a merc is not read: U). d2d accepts the same set (`type_is(.., "hpot")` with equivalence, `feed_merc_item`). The drinking is a one-shot action; there is no merc belt.

## Q7 resistances of a hireling as a defender (SKDEFENDERS, partly)

- Which function says "plain monster": `FUN_0063ee90(unit)` (0x63ee90): a monster (type 1) of class 0x10f -> 1, 0x152 -> 2, 0x167 -> 3, 0x230 / 0x231 -> 4, anything else 0 (V read). The hit's context bit +0x14 is "defender is a monster and this is 0" (`sk-resist-trace.md` Q4, V read + V run for the plain case); so a hireling is NOT plain.
- For a not-plain defender `FUN_0057be00` (0x57be00..0x57bf05) adds the DifficultyLevels ResistPenalty (expansion; 0 / -20 / -50 in classic by the game byte at +0x6d), caps the total at 75 + the max stat (+0x4b = 75, at most 0x5f = 95) and floors at -100 (V read; `sk-resist-trace.md` Q10). d2d's merc fighter does the same (`clamp(resist + gear + penalty, -100, min(75 + max, 95))`, read in `fight.cpp` / `fight/tick.cpp`, I for equality: no run of the penalty path for a merc).
- The base: all four resistances = the hireling.txt `Resist + Resist/Lvl x d / 4` (V run, Q1), NOT the MonStats columns (the level-up writes stats 0x27..0x2d itself; the Normal column of MonStats is forced for the other stats by `FUN_00573cb0`, S `merc-ai.md`).
- Not traced: absorb, the poison per-tick path, curses and auras on a merc (no one-fifth rule: `FUN_0063ee90` is the test there too, I), the get-hit gate (S, `pr-get-hit-rule-proof.md`). SKDEFENDERS stays open for those; recipe T6 below.

## Deviations d2d vs game.exe

| id | rule on both sides | tag | differential |
|---|---|---|---|
| D1 | game.exe: every merc has life regeneration stat 0x4a = (max life in 256ths) / 2000 (Rogue level 15: 19 256ths, level 50: 76; Desert level 20: 36, level 75: 183), set at every level-up and creation; the adder is the monster regeneration event (`FUN_00580610` / `FUN_00580810`, `monster-stats-trace.md` Q9: added to stat 6 each frame once hurt, I). That is life / 2000 per 40 ms frame, 1.25 % of max life a second if it runs every frame (Rogue level 15 about 1.9 life/s, Desert level 20 about 3.6; U). d2d: a merc's life only rises by potions, revive and gear changes (`fight.cpp merc_drinks`, `fight/setup.cpp`, `fight/merc.cpp`: no regen) | value V run (2,940 cases of the stat), d2d absence V read, trigger and rate I, effect U | could be built: create a merc with `FUN_00572840` (done in `mst_oracle.py levelup`), hurt it with `FUN_0057c8d0`, run the event `FUN_00580810` N frames, read stat 6; oracle d2d: constant life. No d2d rule exists to compare, so it is a feature gap (not a PR on the existing harness) |
| D2 | game.exe: the Act 2 merc starts its aura (`FUN_005701b0`, a roll of 5..9 %, once, Q4) and the party in range gets the state's stats (Prayer, Defiance, Blessed Aim at Normal and Hell; Thorns, Holy Freeze, Might at Nightmare). d2d: the aura index is only kept so the pick skips it (`fight/merc.cpp:379, :400`, and for Diablo's score `tick.cpp:222`); nothing is applied to the owner or the merc | V read d2d, the game side start V run in the attack differential, the effect on the owner U (radius and `aurastat` calcs not run) | could be built: start the aura through `FUN_005701b0` on a merc, read the owner's stat list (defence, damage, resist) through `FUN_00625480`; oracle: d2d's player stats (unchanged) |
| D3 | the merc's inputs to `merc_attack` (the "in melee" flag, the distance of skill 0x29, aip1 per difficulty) come from d2d's own geometry (`unit_distance <= melee_rng + 1`, `target_gap + 1`), game.exe's are `FUN_00622c40` / `FUN_006416d0` | I (the rules above are V run, the mapping is not) | could be built: run `FUN_00622c40` natively on random unit pairs, compare with d2d's expression; only the Act 3 / 5 mercs read skill 0x29 |
| D4 | the older reading checkout has no merc experience: `merc_exp` is never written after a hire (the merc stays at its hire level); the newer one has it and equals game.exe (0 of 100,000) | V read (old), V run (new) | closed in the newer checkout |

Equal (0 differences): stats by level, skill levels, hire offers, hire cost, resurrection price, experience award and levelling, the think, the attack think and the skill pick, the potion type set.

## Bot findings (not changed)

| id | where | finding | tag |
|---|---|---|---|
| B1 | `internal/merc/policy.go:113` (`offerOK`), `types.go:106` (`HireOffer.Level` is never read), `AuraPrefs` `policy.go:83` | the offer's level is ignored, but the skill that names the offer is inactive under its `reqlevel`: Cold Arrow 6 (a Cold Rogue offered at levels 3..5 has Inner Sight and a plain shot only), Defiance and Thorns 6, Blessed Aim 12, Holy Freeze 18 (Fire Arrow, Prayer, Might 1). At clvl 8 the offers are levels 3..7 (3 of 5 Cold Rogues have no Cold Arrow, 3 of 5 Defiance Mercenaries no Defiance); at clvl 9 levels 4..8 (2 of 5); clvl 10 one of 5; from clvl 11 (levels 6..10) none | V run (levelup oracle: Cold Rogue skill levels 0 below level 6, Def Id 7 0 below 6), D |
| B2 | `internal/merc/gear.go:166..168` (`ShouldGiveHealing` comment and rule) | "Rejuvenation is NOT given" rests on a doc that lists three potion types; the game's test is by type equivalence and accepts both Rejuvenation Potion classes (Q6). Whether to spend them is a policy call; the effect size on a merc is U | V run (acceptance), U (effect) |
| B3 | `internal/merc/formulas.go:11..23`, `internal/resource/merc.go:9..14` | `HireCost` equals game.exe in 108,000 offers and `ReviveCost` / `MercReviveCost` in all levels 1..255 (formulas written out in `mst_botcost.py`; the Go is read, not run). The comments there still say S / U: they are V run now | V run (formula), V read (Go) |
| B4 | `internal/merc/policy.go:83..99` (AuraPrefs: Defiance, then Prayer; Holy Freeze / Might at Nightmare) | the order is a judgment with no rate behind it; the aura starts on 5..9 % of the merc's rolls and is left out of the roll once running (so the share of attacks goes to Jab), and d2d applies none of it (D2): soaks on d2d say nothing about the aura. Prayer heals the party: for a caster on a thin life bar it may beat Defiance (U, needs the aura calc) | V run (rolls), U (value) |
| B5 | `internal/merc/gear.go:168` (`PotionLifePct` knob, the potion rule) and docs/merc.md: the potion is the only life control | confirmed: no flee, no life test in the merc's AI (V read); a hurt merc regenerates 1.25 % of max life a second in the game if D1 runs each frame (U), which d2d does not show | V read, U |
| B6 | `internal/merc/policy.go:110, :115..118` (`coldArrow`): only Cold Rogues are accepted | Fire Arrow works from level 1 and has the same roll shares as Cold Arrow from level 6 (table in Q4), so a Fire Rogue is the better early hire when the offer's level is under 6; Cold's reason (fire-immune monsters, chill) is the bot's, not a game rule | V run, I |

The merc leash findings of `pets-merc-trace.md` (B1-B5 there) stand; the merc's level now has a rule (Q3), so a merc stuck at the hire level (the older d2d) understates a merc that levelled by the player's kills.

## PR candidates

None. All rule differentials show 0 mismatches ("before" is 0); D1 and D2 are missing features with no d2d rule to patch, effect U. Nothing was built in the fork checkout; nothing pushed.

## Dead ends

- An exact run of `FUN_0057e480` (the experience function) and of the potion drink amount on a merc was not built: both need a populated unit; the first is covered by upstream's F10 oracle for players, the second is U.
- `FUN_0057be00` on a hireling (ctx +0x14 = 0) is V read only: the existing resist oracle copies d2d's line (not the real merc fighter, which is a `Fight` method), so a run would be game.exe against a transcription: left to SKDEFENDERS.
- The level a merc has on load is V read (loop 0x56aa88..0x56abc2), its loop bound I.
- The worker-host runner refuses a script whose text holds the word "kill" even inside a name; the experience scripts are therefore `mst_xp*` and the d2d mode `xp`.

## Drill recipe (hidden d2d instance, throwaway character, ledgered cheats only; fork verbs by name: `merc`, `spawn`, `hit`, `give`, `dump merc`, `step`)

| id | what | confirms |
|---|---|---|
| T1 | `merc hire 0` (Rogue Fire), `merc hire 1` (Cold), `merc hire 6` (Desert): `dump merc` level, life, str, dex | Q1 table rows (level 1 at hire with experience 0: the fork's verb uses experience 0; to read level 15 use `give` of experience if the fork has it, else the real offers) |
| T2 | the hire window at clvl 8, 9, 11 in the original: read level, life, def, price of each row | Q2 table; B1 (how many rows lack the arrow or aura) |
| T3 | `merc hurt` to 50 % life, then `step` 250 and `dump merc` on d2d (expect no change) and on the original (expect a rise of about 1.25 % max life a second if D1 runs per frame) | D1 trigger and rate |
| T4 | Desert merc (Def or Comb): stand with the merc at full life, `spawn` a monster, watch the owner's defence / life regen when its aura starts (the first 10 to 60 rolls) | D2 effect |
| T5 | `give` a Rejuvenation Potion to the merc (belt shift-key) at 30 % life; read the merc's life and the potion count | Q6 effect, B2 |
| T6 | (SKDEFENDERS) `mon set` a Fire Bolt caster, hit the merc at Normal / Nightmare / Hell, read the merc's life drop vs its base resist (Q1) | Q7 penalty and cap on a hireling |
Pass rule per step: two identical runs; the character ends in town, no panel open.

## Matrix-cell updates (proposed text)

- `docs/research/sorc-skill-matrix.md` rows that mention hirelings (line "(a) hit effect ... boss / champion / hireling exceptions", row 2 / SKARMOR (1)): no change from this card; the hireling "exception" there is `FUN_0063ee90`'s five classes (Q7).
- `docs/community/open-work.md` SKDEFENDERS: "partly answered in docs/research/mercstats-trace.md: which function (V read), the base resists by level (V run 2,940), the penalty and cap (V read); open: absorb, poison per tick, curses and auras on a merc, a run of FUN_0057be00 with a hireling context".
- `docs/research/pets-merc-trace.md` D3 (merc auras count as running but do nothing): confirmed and numbered D2 here (V read d2d, effect U).

## Reproduce (the worker host)

```sh
tools/emu/run-on-worker.sh --ttl 1500 mst_build.sh > build.txt            # d2d side, links the scratch upstream build
tools/emu/run-on-worker.sh mst_oracle.py hire 1001-1500 ...              # 3 chunks (1001-1500, 1501-2750, 2751-4000), game side
tools/emu/run-on-worker.sh mst_run.sh hire 1001-1500 ...                 # d2d side; python3 -I tools/emu/mst_cmp.py game.txt d2d.txt
tools/emu/run-on-worker.sh mst_oracle.py revive; tools/emu/run-on-worker.sh mst_run.sh revive
tools/emu/run-on-worker.sh --ttl 900 mst_oracle.py levelup;  tools/emu/run-on-worker.sh mst_run.sh levelup   # mst_cmp.py --skip regen --skip next --skip life6 --skip lv
tools/emu/run-on-worker.sh --ttl 3000 mst_think_all.sh 25000 31 32 33 34
tools/emu/run-on-worker.sh --ttl 3000 mst_attack_all.sh 25000 41 42 43 44
tools/emu/run-on-worker.sh --ttl 3000 mst_xp_all.sh 25000 51 52 53 54
tools/emu/run-on-worker.sh mst_potion.py                                  # the potion filter over 659 item classes
python3 -I tools/emu/mst_botcost.py hireling.txt game_hire_*.txt game_revive.txt   # bot formulas: 108,000 offers, 256 levels, 0 differ
python3 -I tools/emu/mst_picks.py hireling.txt                            # the roll shares
```

Limits: the oracles hook the unit plumbing (stat getters and setters, the path search, the effect hand-offs) as upstream's `merc.py` does; the think's moves are "found a path or not" bits, not a real path search; d2d's inputs to `merc_attack` are mapped as in D3. Cases count 108,000 + 256 + 2,940 + 3 x 100,000 = 411,196 runs; the checksum-free text outputs live on the worker host's scratch only.

## Shared-file lines for the merger

In `openspec/changes/mercstats/merge-lines.txt` (BACKLOG, INDEX, CHANGELOG).
