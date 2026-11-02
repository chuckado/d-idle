# Monster melee against the player: the damage that decides deaths (card monsterhit, decomp-queue item 7 follow-up)

> status: current (2026-11-02). Offline, docs and one table helper only: no game, no d2d slot, nothing pushed or posted. Tags: V read = read in the game.exe 1.14d disassembly or the d2d source at file:line; V run = observed running (emulator, lab binary, live); D = derived from the 1.14d tables; S = shown by an existing proof (the doc is named; its V run counts are quoted, not repeated as mine); I = inferred; U = unverified.
>
> **Upgrade pass (2026-11-02, second run): V run for the to-hit, damage-roll, get-hit and block rows.** The first pass was desk-only. This pass ran the existing oracles `tools/emu/monster_hit.py` and `tools/emu/mon_missile.py` through `tools/emu/run-on-worker.sh` against game.exe 1.14d's own code under unicorn (section 0 has the counts). Rows that still say S rest on the doc named, and rows that say I or U were not run: the player-defender damage path (D1, D2, D5) needs a harness not built in this pass (see Dead ends). The d2d side of the monster_hit / mon_missile diffs is the ports inside those scripts (written against d2d at their card date); the d2d source re-read here is `components/rules/combat.hpp` of the lab checkout `src` (main moved on since the first pass: `monster_blow` now has `dr_pct`, `dr_flat`, `weapon_block`, line 403 onward), V read.

## 0. Run results of the upgrade pass (V run, worker host)

| Command (all through `tools/emu/run-on-worker.sh`) | Result |
|---|---|
| `monster_hit.py --exe /tmp/casttrace/game.exe 100000` | `hit` (`FUN_0057d9b0`, to-hit roll, monster attacker, melee): **100,000 cases, 0 differ** from the port. `dmg` (`FUN_0057b420`, monster physical roll in 256ths): **100,000 cases, 0 differ**. `gh` (`FUN_0057cb00`, get-hit decision): **672 boundary + 100,000 random cases, 0 differ** from the tiered port and 0 from d2d's `rules::get_hit`. A 20,000-case run first gave the same 0 / 0 / 0 |
| `mon_missile.py --exe /tmp/casttrace/game.exe 100000` | `block` (`FUN_0057dfb0`): **100,000 cases, 0 differ from the reading** (a third in RN / TW; none when can-block is 0); **11,156 differ from the d2d port**: can-block 0 melee 5,250, can-block 0 missile 5,253, mode 6 melee 324, mode 6 missile 329. Melee always passes can-block 1, so the 5,250 melee line is a synthetic case; the real gaps are the missile gate (carried verdict, F14) and the mode 6 (town-walk) third. `roll` (missile physical, `FUN_005a89a0`): **100,000 cases, 0 differ** from the 256ths exclusive reading; d2d whole-point inclusive minus game mean -0.308 hp on random 256ths ends |
| `mpq-table.sh monstats` and `monlvl` (patched stack, first line `Id hcIdx`, 735 rows; MonLvl 111 rows) then `mh_rows.py --el` | 176 monster / difficulty rows with an El side part (D): cold 37, fire 20, lightning 46, poison 36, magic 21 (only the succubi and succubus witches, level 32 to 42, Act 5), mana 19, stamina 9, life 1, stun 4, random 6 |

Counting: the harness states cases and mismatches; the ports are the d2d rule at the card date, so "0 differ" is game.exe against that port, not against a rebuilt d2d binary.

## Verdict

1. **The to-hit roll, the damage roll and the get-hit decision of a monster melee swing are already proven equal to game.exe** (S: monster-attack-trace.md Q1, 100,000 / 100,000 / 100,672 cases, 0 differ; stat rows 242,220 cases and 85,213 real Normal units, monster-stats-trace.md). Nothing in this card reopens them.
2. **What was not covered before, and is covered here by reading only:** the defender's side after the hit (physical damage reduction, magic damage reduction, the magic element), the El1..3 side parts of a swing, block/avoid/dodge order, leech and burn on the row, and what the bot's survival model assumes. Findings: three d2d deviations that stay I or U (D1 to D3, none matters for an Act 1 / 2 Normal Sorceress without damage-reduction gear), one carried Nightmare / Hell deviation (D4), five bot findings (B1 to B5).
3. **PR candidate: none.** No new deviation is proven by a run (the one run deviation, mode 6 block third in the d2d port, is town-only and already carried), and the proven ones (monster-attack-trace.md, monster-stats-trace.md) already have their verdicts.

## Q1 to Q8 (what happens, in order)

| Q | Question | game.exe 1.14d | d2d (combat.hpp, main `173b79cc`) | Tag |
|---|---|---|---|---|
| Q1 | Which attack rating and level does the swing use? | `FUN_005a4f50` at every mode set copies the mode's stats (A1 or A2; SC / S1 for S1): to-hit = MonLvl TH(diff) x A1TH / 100 (a NoRatio row: the column), level = unit stat 12 (the area level in Nightmare / Hell for a ratio row: `FUN_00573cb0` 0x573dc1..0x573e4d) | `monster_stats` monsters.hpp:378-398: `pct(level_row.to_hit[diff], a1_th)`, level from `MonType::level[diff]` | S (monster-stats-trace Q1; 242,220 cases equal in the row function; NM / Hell level is D4) |
| Q2 | To-hit chance | `0057d9b0`: AR x 100 / (AR + DEF) x 2 x alvl / (alvl + dlvl), clamp 5..95; DEF = defense + stat 33 (melee) / 32 (missile); a defender with negative DEF or AR moves the amount to the other side | `rules::hit_chance` combat.hpp:28-34 (same lines, `int64`) | V run (section 0: `hit`, 100,000 cases, 0 differ; S: monster-attack-trace Q1 the same) |
| Q3 | Who skips the roll | a player in RN is hit without the roll (`0057ec71`) | `!(running && !missile)` combat.hpp:609 | S (same table) |
| Q4 | Block, avoid, dodge | block `0057dfb0` (chance `00622720`), a third in RN or town walk; evade in WL / RN, else dodge (swing) / avoid (missile) `0057dd60`; for a missile the order is evade / avoid / dodge first, then block (`005ad730`, V read in monster-attack-trace Q2) | block first (`defender.block`, a third while running), weapon block standing, then evade / dodge / avoid, combat.hpp:611-614; shield block = (shield block + class + stat) x (dex - 15) / (clvl x 2), at most 75 (comment combat.hpp:87-88, code `make_fighter`) | block V run (section 0: 100,000 cases, 0 differ from the reading, a third in RN / TW; d2d port differs 11,156, of which mode 6 = 653); evade / dodge S; the melee ORDER on the game side is U (D3) |
| Q5 | The physical roll | `0057b420`: min<<8 + rand((max - min)<<8), top exclusive, in 256ths from stats 21 / 22 | `rng.range(a1_min, a1_max)` whole points, inclusive (combat.hpp:616) | V run (section 0: `dmg` 100,000 cases, 0 differ from min << 8 + rand((max - min) << 8); S: d2d mean +1/512 hp, wider spread) |
| Q6 | Reduction of that roll | `0057c6c0` -> `FUN_0057c1e0` applies the defender's resistances and damage reduction to the record in 256ths (V read: pr-armorhit-proof.md "+8 is the PHYSICAL damage after reduction, in 256ths"; sk-armor-trace.md); the resist function `FUN_0057be00` / `FUN_0057bf80` is V run on a monster defender (skabrecheck.md, sk-resist-trace.md: 147,628 and 101,626 cases) | `phys = max(phys x (100 - dr_pct) / 100 - dr_flat, 0)` in whole points (combat.hpp:617); `dr_pct` is stat 36 at most 50, `dr_flat` stat 34, `mdr` stat 35 (comment combat.hpp:94-95); then each El row (Q7) | game side read for a MONSTER defender only; the player-defender path is I (D1) |
| Q7 | Side parts El1..3 | `FUN_005a4f50` writes each element row of the mode as its own stat (poison x 10, duration stat), rolled at the hit by `FUN_0057b7d0` / the record builder; the element goes through `FUN_0057be00` with the defender's resist (fire, lightning, cold, poison and, for magic, stat 0x25), cold adds a chill length, poison a per-tick damage (S: nec-skill-trace.md for the magic row on a monster defender; monster-mods-trace.md for the boss rows) | one roll per matching row with its chance (0 = 100): `element.mode == "A1"/"A2"`, `rng.range(min, max)` whole points, fire / lightning / cold `resisted() - mdr`, poison `resisted(roll x 10)` over `dur x 2` ticks, magic `roll - mdr` with NO resist (combat.hpp:619-632) | V read d2d; game side per player defender is I (D2, D5) |
| Q8 | Hit recovery after the hit | `0057cee0` after a landed hit: `0057cb00` decides by hit class and 1/16 of max life; the length is `speed x (50 + FHR) / 100` (movement-and-recovery-trace.md) | `rules::get_hit` combat.hpp:231 (tiers) and `get_hit_ticks` combat.hpp:305 (50 + FHR), both on this checkout | V run (section 0: `gh` 672 + 100,000 cases, 0 differ from the tiered port and from d2d's get_hit) for the decision; the length S; closed: the two earlier deviations are fixed upstream (monster-attack-trace.md verdict 1; the comment at combat.hpp:288 gives the Sorceress ticks 16 at 0 FHR, the older doc says 15, same rule) |
| Q9 | Life leech or mana burn on the row | no leech column in MonStats; drain reaches a monster only through a MonUMod (25 mana burn: stats 0x3e / 0x3f = L-DM x rows, `5a1f90`) or a MonProp item property (U: monster-stats-trace "MonProp effects" not traced beyond speed, rate, crush, thorns) | `if (taken.hit && monster.mana_hi > 0) foe.mana_burn += rng.range(...)` ai.cpp:2746; the MonUMod init equal for bosses | S (monster-mods-trace.md, 157,618 cases); MonProp leech U |

Act 1 and 2 melee rows (D, from the archived first pass docs/archive/monster-attack-trace-2026-10-30.md section 3, quoted): Fallen A1 1 to 2 at AR 8 (Blood Moor, level 1) = 14 percent against DEF 20 at clvl 3; Carver 2 to 4 at AR 23 to 31; Rakanishu as a Lightning Enchanted unique 2 to 5 + 3 to 5 lightning. **Act 2 melee rows and the elemental melee rows of Acts 1 and 2 were not tabulated by any trace (U): no MonStats dump was available on this box.** `tools/emu/mh_rows.py` lists them from a MonStats / MonLvl dump (section 8).

## Worked rates (I: arithmetic on the S / D rows above, per monster in reach, swing period 1.0 s for the Act 1 types)

| Case (clvl 3 Sorceress, DEF 20, Normal) | Standing | Running (no roll) | Source of the swing |
|---|---|---|---|
| Fallen, Blood Moor | 14 percent x 1.5 = 0.21 hp/s | 1.5 hp/s (7.1x) | D: 1..2, AR 8 |
| Carver, Stony Field (level 4, AR 23) | 61 percent x 3 = 1.8 hp/s | 3.0 hp/s (1.6x) | D: 2..4 |

The ratio between standing and running is the lever the bot controls (B2); the absolute numbers move with DEF and clvl.

## Deviation list (d2d vs game.exe)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D1 | Physical reduction of a monster swing on a PLAYER: game.exe reduces the 256ths record in `FUN_0057c1e0` (percent and flat, order U); d2d floors `phys x (100 - dr) / 100` to whole points, then subtracts the flat part (combat.hpp:617). Effect: d2d floors away on average about half a hit point per hit when `dr_pct` is non-zero, zero effect at `dr_pct` 0 (every Normal Sorceress before gear that carries it); the order percent-then-flat is V read for d2d, U for game.exe | V read both sides; no run; I for the size | **Buildable.** Game: `skab_phys.py`'s runner with a player unit as the defender (stats 0x22 flat, 0x24 percent, 0x23 magic reduction) and the attacker record from `0057b420` (min / max 1..40); oracle: life lost in 256ths after `0057c6c0`. d2d: `monster_blow` with `Fighter::dr_pct / dr_flat`, life x 256. Grid: dr_pct 0..50 x dr_flat 0..20 x roll bounds, 100,000 random; expected: every case with a fractional product differs (d2d lower by 0..255/256 hp) |
| D2 | The MAGIC element of a monster swing: game.exe resists it by the defender's stat 0x25 through the same `FUN_0057be00` row used for monsters (S: nec-skill-trace.md, V run on monster defenders), d2d does not (`case 4: max(roll - mdr, 0)`, `Fighter::res` has four entries, fire / lightning / cold / poison; combat.hpp:625). Matters only for an El row of type magic: **D (section 0, 21 rows): only the succubi and succubus witches, level 32 to 42, Act 5; no Act 1 / 2 row has one, so D2 is cosmetic for this question.** Other El types that Act 1 / 2 rows carry and d2d's `type < 0` test skips or routes to `default` (mana, stamina, life, stun, random: 39 rows) are not traced on the game side: U (their mapping to d2d's type index was not read this pass) | V read d2d, S game side for monsters, I for the player | Buildable with D1's harness (stat 0x25 on the player, an El row type magic); first run `mh_rows.py --el` to see whether any Act 1 / 2 row exists; if none, D2 is cosmetic for this question |
| D3 | Order of block versus avoid / dodge on a melee swing: d2d rolls block first, then evade / dodge (combat.hpp:611-614); game.exe's missile path evades / avoids first, then blocks (V read, S); the melee order is U. Both are independent draws on the defender's seed, so the probability of a landed hit is `(1 - block)(1 - dodge)` either way; only the split blocked / dodged and the number of draws differ (the draws matter for lockstep seed streams, not for rates) | I (the independence is U: one shared draw would change the rate) | Cheap: run `0057dfb0` and `0057dd60` natively on a defender seed 100,000 times with both chances set (extend `mon_missile.py block`), compare the joint distribution with d2d's two-draw rule. Not worth a PR |
| D4 | Nightmare / Hell level of a ratio-row monster: area level in game.exe, Level(N) / Level(H) in d2d. It moves the level term of the hit chance, the AR and the damage (all read at the level) | S: monster-stats-trace.md D1, V run 33,693 of 82,171 NM units and 36,322 of 79,592 Hell units differ in level | yes, `ms_pop.sh`; not Normal |
| D5 | The El row roll: d2d rolls `rng.range(min, max)` in whole points inclusive; game.exe's min / max for an El row are written by `FUN_005a4f50` into stats, whether in points or 256ths is U (the MonUMod rows are `<< 8`, S) | U | Buildable: run `FUN_005a4f50` for a monster with an El row, then the record builder, 100,000 cases; as D1 for the d2d side |

Carried, already proven (not new here): melee damage roll whole-inclusive versus 256ths-exclusive (monster-attack-trace.md: mean +1/512 hp, 0 differ from the reading, 100,000); missile block gate and MissA to-hit gate (10,503 of 100,000 block cases; 94 of 142 MissA rows, none in early Act 1).

## Bot findings (described, not changed)

| B | Finding | file:line | Tag |
|---|---|---|---|
| B1 | The survival simulator charges every monster in contact a flat `MonsterDps 1.6` (times `Exposure`: 0.35 Sorceress kiting, 1 Paladin) whatever the monster, its AR or the player's defense. The traced standing rates for a clvl 3 Sorceress are 0.2 (Fallen) to 1.8 (Carver) hp/s, a Blood Moor Fallen has AR 8 to 12, so against any real defense it lands far less often than the constant implies. The constant is U in the code (nearest default, "BloodMoorSorc"); it is not derived from the rows | internal/resource/sim.go:123, :140, :218 | I |
| B2 | No standing / running distinction: a running player is hit by every swing (V read, S) so the per-monster rate rises 1.6x to 7x against the table above, while the sim has one rate. The d2d run policy that runs near packs (monster-attack-trace.md B2, `internal/d2dinput/run.go:25`) is the real driver of damage taken on d2d, and no sim shows it | internal/resource/sim.go:218 | I |
| B3 | Drink / chicken thresholds are percentages of max life and carry no swing size: at clvl 3 life about 45, the drink line (55 percent) leaves about 25 hp and the chicken line (25 percent) about 11 hp, i.e. 1 to 3 Carver swings or one Rakanishu swing plus bolts; a Carver pack of six on a runner deals about 18 hp/s (archived first pass, section 3, I), so the gap from the drink line to the chicken line is under a second of pack damage. Whether the samples are fast enough is U (poll interval not read here). Early low-life characters cannot rely on the thresholds alone, the crowd rule is the cover | internal/env/tuning_combat.go:85 (DrinkLifePct 55, ChickenPct 25), internal/combat/safety.go:154-216 | I |
| B4 | `Blockable = ToHit || MaxDamage > 0` for a missile row; game.exe lets a missile be blocked only when its physical part is non-zero (S: 181 monster-skill missile rows have none, early Fallen Shaman fire is one). Matters only for a character with a shield (Paladin builds); a Sorceress has block 0 | internal/combat/threat.go:28, :35 | I |
| B5 | Nothing in the bot's incoming-damage model reads an El side part (poison per tick, chill length, magic or elemental damage after resist) or an A2 / S1 swing: the sim has one flat rate and the thresholds are life percent. A poison hit costs life over ticks that a drink threshold only sees later. Not measured here | internal/resource/sim.go, internal/combat/safety.go | I |

Not a finding (checked): `crowdAdjacent` 8 subtiles counts a monster as in reach well beyond the AI's melee reach, but it was measured live (docs/fixcrowd.md: monsters in A1 / A2 at 1 to 9 subtiles centre to centre), so it is not flagged.

## PR candidates

**None.** A d2d PR needs a deviation proven by the existing differential (mismatches before > 0, after the fix 0). D1, D2, D5 are buildable but not built (no run was possible), D3 is a rate-neutral split, D4 is Nightmare / Hell and needs area plumbing (monster-stats-trace.md). No branch was made in the fork checkout.

## Dead ends

- First pass: running the oracle was not possible (worker host unreachable). Second pass: ran it (section 0). The disassembly was not re-read; only the existing oracles ran, so the V run claims are exactly those of `monster_hit.py` and `mon_missile.py`.
- The MonStats dump is now done (`mpq-table.sh`, section 0); `mh_rows.py --el` on it found no magic El row outside Act 5. Left undone: D1 / D5 (a player defender through `FUN_0057c1e0` / `0057c6c0`, which needs a new harness, `skab_phys.py` retargeted at a player unit with stats 0x22 / 0x23 / 0x24) and D3 (block versus dodge order on melee): not built in this pass, they stay I / U. A Sonnet-sized job; not a failed decomposition.
- The player-defender damage path (`FUN_0057c1e0` for a player, stats 0x22 / 0x23 / 0x24): the monster-defender half is V run (SKAB), the player half is only read in other traces. Stat ids here come from the d2d comment (combat.hpp:94-95), not from a fresh read of ItemStatCost: I.

## Matrix and ledger proposals (applied by the merger, not by this card)

- docs/coverage/ledger.tsv: row F02 (defence and block, monster side) added by this card (`TESTED-DEVIATION`, 100000 cases, section 0), `rule-coverage.md` regenerated. F13 and F14 unchanged. When D1 is run, add `formulas F19 TESTED-DEVIATION docs/research/monsterhit-trace.md <cases> <date> monster swing vs player: DR, MDR, magic element, El rows` with the real case count; `python3 tools/coverage/coverage.py --write` then regenerates rule-coverage.md.
- docs/research/baseline-audit.md, row on monster rules: add "defender-side reduction and El side parts of a monster swing: d2d reads (whole points) and differs from game.exe's 256ths path by about half a hit point per hit only with damage-reduction gear; not run (docs/research/monsterhit-trace.md)".
- docs/research/monster-attack-trace.md Q1 table, row Damage: no change; the note "Order" row may add "block before dodge is d2d's order, melee order on game.exe U (monsterhit D3)".

## 8. Drill recipe (what to run later; nothing here was run)

The worker host first (`labrun <host> -- top -bn2 -d1`, idle + nice at least 25 percent), every process through `tools/agentrun.sh <agent> --ttl SECS --`.

1. **Rows (D, no game):** dump MonStats.txt and MonLvl.txt with `mpq-cat` into a scratch dir, then `python3 tools/emu/mh_rows.py MonStats.txt MonLvl.txt --el` (rows with an El side part) and `--only fallen1,carver1,... --def 20 --clvl 3` (hit chance and mean swing). Decides whether D2 and D5 touch any Act 1 / 2 row.
2. **D1 / D2 / D5 differential (V run):** copy `tools/emu/skab_phys.py` to `mh_dr.py` in the upstream emulator directory, change the defender to a player unit with stats 0x22 / 0x23 / 0x24 / 0x25 set from the case, the record from `0057b420` as `monster_hit.py dmg` does, and read the life lost after `0057c6c0`; d2d side `monster_blow` through a test like `skab_phys_test.cpp`. Report cases and mismatches before; a fix (floor to 256ths in `monster_blow`) would have to show 0 after, and then the PR candidate needs the independent verifier.
3. **Live confirmation of D1 (U, lab binary, docs/forkdbg.md, docs/forkcheats.md):** a fresh throwaway Sorceress `--no-save`, `debug goto 2`, `debug spawn fallen1 1`, give damage-reduction gear with `debug give <code> pack` + equip (or a `debug stat 36 20` / `34 2`; the stat verb is in forkcheats.md), `debug mon set <id> dmg 3 3` so the roll is fixed, a hit on the player with 25 percent damage reduction: 3 x 0.75 = 2.25. Read life before and after with `dump stats` over 8 landed hits: d2d should lose 16 (2 per hit), the 256ths rule of game.exe predicts 18. This confirms only the d2d side (U until the emulator run of step 2 gives the game side); there is no live game.exe run.
4. **Rate check of B1 (U, lab):** `debug spawn carver1 6` around a standing Sorceress at clvl 3 with DEF known (`debug states` prints `def=`), `monsters` and life samples every 25 frames for 60 s; compare the mean loss per carver-second with 1.8 hp/s and with the sim's 1.6 x Exposure.

## Reproduce

```sh
python3 -I tools/emu/mh_rows.py --selftest      # table helper, synthetic data only: "selftest ok"
tools/emu/run-on-worker.sh --ttl 900 monster_hit.py --exe /tmp/casttrace/game.exe 100000
tools/emu/run-on-worker.sh --ttl 900 mon_missile.py --exe /tmp/casttrace/game.exe 100000
tools/emu/run-on-worker.sh --ttl 300 mpq-table.sh monstats > monstats.txt   # drop the first "labrun:" line, same for monlvl
python3 -I tools/emu/mh_rows.py monstats.txt monlvl.txt --el --def 20 --clvl 3
```

## Shared-file lines for the merger

See `openspec/changes/monsterhit/merge-lines.txt` (BACKLOG, INDEX, CHANGELOG).

Glossary: no new term proposed; "damage reduction (percent / flat)" and "El side part" are used as in docs/data.md.
