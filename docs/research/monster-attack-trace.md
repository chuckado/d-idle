# Monster melee and missile attacks on the player: game.exe 1.14d vs d2d (DECOMPMON, decomp-queue item 7)

> status: current (2026-10-31, second pass on upstream main 0dc4c62a; the 2026-10-30 melee-only pass is docs/archive/monster-attack-trace-2026-10-30.md, kept for its spawn-stat, Lightning Enchanted and hp/s tables). Tags: V = read in game.exe (`disasm.py`) or run in game.exe's own code under unicorn by this agent; V-d2d = d2d code read at file:line; D = 1.14d patched tables (`mpq-cat`, `Id<TAB>hcIdx` checked); S = upstream doc (`docs/research/re/*.md`); I = inference; U = unverified. Classic, Normal, one player. Offline only: no game, no slot, nothing pushed or posted. Oracles: `tools/emu/monster_hit.py`, `tools/emu/mon_missile.py` (run from upstream's `tools/emu`).

## Verdict

1. **Melee: clean.** To-hit, the physical roll and the get-hit tiers run in game.exe equal d2d on current main: 100,000 / 100,000 / 100,672 cases, 0 differ. The get-hit deviation of the first pass (18,502 of 100,000) is fixed upstream (`rules::get_hit`, combat.hpp:203).
2. **Missiles: same formula, two narrow gate deviations, none in the early game.** (a) d2d lets a missile with no physical part be blocked; game.exe passes can-block = physical != 0 (10,503 of 100,000 block cases). (b) d2d rolls to-hit for every MissA shot; game.exe only when the row's ToHit is set (94 of 142 MissA rows, 0 in early Act 1). One inference: game.exe reads the shooter's to-hit at impact (Quill Rat: A2TH while still in A2), d2d always A1TH; if anything d2d's rats hit less.
3. **Monster rules explain about none of the -48 % XP/h.** Every deviation found makes d2d monsters hit less or equally. d2d's fights are not longer (4.6 s of fight per kill vs 5.5 s on Game.exe). The d2d bot takes 1.6x more damage per fight second through exposure (3.6 approaches per kill vs 0.85, no missile dodge on d2d), which costs 2 to 4 s per kill in exits and potion trips: 20 to 45 % of the gap, all bot-side. The other 55 to 80 % is walking and stalls outside fights (walk 14.5 s of 18.6 s per kill).

## Q1 melee (V unless marked)

| Step | game.exe | d2d (main 0dc4c62a) | Verdict |
|---|---|---|---|
| Reach | `0057ec10` -> `00622c40`: 3 for a monster + extra reach | `unit_distance(..) > melee_rng + 1` (ai.cpp:1863 and siblings) | same rule (S, upstream AI oracle) |
| Swing timing | A1/A2 for its AnimData frames, the hit on the action frame; then NU and the think `aidel` frames on (`005a7c20`, `005a73e0`); aidel 15 for every Act 1 type (D) | `timing.action_ms()` (ai.cpp:1990); `next_act = now + aidel x 40 ms` (ai.cpp:1966) | same (S; DECOMPWAKE ran the scheduler: 55,268 / 0) |
| Mode stats | `005a4f50` at every mode set (`005a7d39`): A2 -> A2 damage / A2TH, SC / S1 -> S1, else A1 | A1 or A2 / S1 stats by mode (ai.cpp:2642-2645) | same |
| Running target | a player in RN is hit without a roll (`0057ec71`) | `!(running && !missile)` (combat.hpp:582) | same |
| To-hit | `0057d9b0`: AR x 100 / (AR + DEF) x 2 alvl / (alvl + dlvl), 5..95; DEF + stat 33 (melee) / 32 (missile) | `hit_chance` (combat.hpp:28) | 100,000 cases, 0 differ |
| Block | `0057dfb0(.., missile 0, can-block 1)` (`0057ece1`): chance `00622720`; a third for a player in RN or TW (town walk); roll on the defender's seed | a third while running (combat.hpp:584) | 0 differ outside town (`mon_missile.py block`) |
| Damage | `0057b420`: min<<8 + rand((max - min)<<8), top exclusive | `rng.range(a1_min, a1_max)` whole points, inclusive | 100,000 cases equal the 256ths reading; d2d mean +1/512 hp, wider spread |
| Order | to-hit -> block / dodge -> damage `0057c6c0` -> takes-damage `0057cee0` -> get-hit `0057cb00` | same (combat.hpp:578-615, fight/tick.cpp:495) | same |
| Get-hit | tiers from 1/16 of max life by hit class | `rules::get_hit` | 100,672 cases, 0 differ |

Who rolls: all of it is the server's (game.exe hosts the server in single player); the client plays the GH / BL modes it is sent (S).

## Q2 missiles (V unless marked)

| Step | game.exe | d2d | Verdict |
|---|---|---|---|
| When a shooter fires | its MonAI think (QuillRat, CorruptArcher, SkeletonBow, FallenShaman) picks A1 / A2 or the skill; the missile at the action frame or sequence event | MissA1 / MissA2 at the action frame (ai.cpp:1997-2020); ShamanFire by MonSeq do 85 (ai.cpp:2037) | same (S: upstream's think oracle `mon_ais.py`, monster-ai.md) |
| Flight | velocity ((Vel + VelLev x lvl / 8)<<8) x 75 / 100 (`0059fa30`); collision `00641cb0` by Size | `missile_speed`, `missile_struck` | same (S, missiles.md) |
| Damage at launch | `0059f900` -> `0064b860` -> `0064ac60` writes the missile's stats 21 / 22 and elements from the owner | `src_min x SrcDamage / 128 + MinDamage` (ai.cpp:2015-2017) | `0064b860` read only (U) |
| Hit entry | `005adf10`: `0057d9b0(owner, target, bonus = missile stat 19, missile 1)` ONLY when the row's ToHit (+0x18b) is set (`005ae04c`); no running skip; the bonus is set only by creation flag 0x1000 (`0059fe77`) | MissA: `monster_blow(.., missile.src, true, rng, true)` always rolls, ignores `info->to_hit` (ai.hpp:793); skill missiles: `missile_blow` never rolls (ai.hpp:657) | deviation (b): 94 MissA rows (bighead, goospit, mummy1 ..); 4 skill rows with ToHit that d2d never rolls (blade creeper, rogue1, bladefragment1); early Act 1 none (spike, cr_arrow, skbowarrow have ToHit 1; shafire none: both always hit) |
| AR in that roll | the owner's stat 19 at impact (its current mode's overlay) | A1TH; S1TH for an S1 shot (ai.cpp:2018) | I: differs for an A2 shooter still in A2 at impact (Quill Rat 101 vs 134 %); rate U; about +2 points of a 14 % chance at clvl 3 |
| Block | `005ad730` -> `0057dfb0(.., missile 1, can-block = record +8 != 0)` (`005ad7f9`) | always blockable | deviation (a): 100,000 cases, 0 differ from the reading, 10,503 from d2d; 181 monster-skill missile rows have no physical part; early: Fallen Shaman fire, only with a shield |
| Dodge | `0057dd60`: Evade in WL / RN, else Avoid (missile) or Dodge | `moving ? evade : missile ? avoid : dodge` | same (none on a Sorceress) |
| Damage roll | `005a89a0`: min + rand(max - min) in 256ths, top exclusive | whole points, inclusive | 100,000 cases equal the reading; d2d mean +1/512 hp on whole ends |

Early Act 1 shooters (D): quillrat1 MissA2 spike1 (ToHit 1, A2 15-25 %, poison El A2); cr_archer1 MissA1 cr_arrow1 (ToHit 1, A1 60-90 %, A1TH 256); sk_archer1 skbowarrow1 (ToHit 1); fallenshaman1 ShamanFire -> shafire1 (no ToHit, fire 1-4).

## Deviation table

| Term | Cases | Differ | Early-game effect |
|---|---|---|---|
| melee to-hit / damage / get-hit | 100,000 / 100,000 / 100,672 | 0 / 0 / 0 | none |
| missile to-hit formula (missile flag in the `hit` set) | about 33,000 | 0 | none |
| missile block, can-block gate (a) | 100,000 | 10,503 (+653 town walk: untargeted in town) | Shaman fire with a shield: d2d blocks some, game.exe none |
| MissA ToHit gate (b) | 142 MissA rows, exhaustive | 94 rows | none in the early areas |
| A2 shooter AR at impact | 9 MissA2 rows | I, rate U | Quill Rats: d2d hits a little less |
| missile damage roll | 100,000 | 0 vs the reading | tail only |

## Q4 where the gap is (V, read-only logs: Game.exe = the Fresha baseline window, 114 min, 742 kills; d2d = m2early 8 seeds x 180 virtual min, 4,599 kills)

| Per kill | Game.exe | d2d |
|---|---|---|
| seconds per kill (minute lines) | 9.2 | 18.8 |
| fight seconds (`combat fight` lines) | 5.5 (79 % of the time) | 4.6 (26 % of the time) |
| damage taken (hp) / per fight second | 0.87 / 0.16 | 1.19 / 0.26 |
| potions / missile dodges / approaches | 0.035 / 0.24 / 0.85 | 0.065 / 0.00 / 3.6 |
| d2d `killtime` split | | walk 14.5 + cast 0.8 + wait 3.4 of 18.6 s |
| resupplying minutes | 24 % | 33 % (218 of 480 "Returning to the field") |
| game exits | 0 in the window | 66 followed by a kill, 203 s each = 15.5 % of the time; 19 "chicken first" with an empty belt = 4.9 % |

Attribution of the +9.6 s per kill: damage-driven time on d2d (exits 0.9 to 2.9 s, potion and emergency trips with their walks back about 2 s) less Game.exe's (about 0.65 s) = 2.2 to 4.2 s, 20 to 45 %. That is exposure, not rules. The rest: walking between packs, stalls, no-portal walks (976 "no town portal" events vs 21) and density (per-kill-time-split.md).

## Q5 bot findings (fix queue)

- B1 d2d gives the bot no missiles: `internal/d2dsrc` never fills `Snapshot.Missiles` (items.go, player.go build the snapshot), so the dodge (`internal/combat/fight.go:476`) and `sidestepMissiles` (`internal/combat/dodge.go:88`) never fire on d2d (0 vs 0.24 dodges per kill). Needs the d2d dump to carry missiles (U whether it can).
- B2 Running into packs: the d2d run policy runs whenever stamina is at least 25 % (`internal/d2dinput/run.go:25`); a running player is hit by every swing and blocks a third (V `0057ec71`, `0057dfb0`; same in d2d). With 3.6 approaches per kill this is the likeliest source of d2d's extra damage. Walk within about 8 subtiles of hostile melee, run for travel.
- B3 `CastGapMs` 450 (`internal/env/tuning_combat.go:73`) is under the 560 ms Sorceress cast: 5.4 cast commands per kill on d2d vs 1.1 (dropped mid-cast commands, gap-casts-per-kill.md).
- B4 Chicken with an empty belt (19 exits, 225 s each): potion poverty with `TownPotionMin` 2 (`internal/env/tuning_combat.go:69`), economy, not monsters.
- The fight telemetry `damage` is life lost between samples (`internal/combat/dodge.go:131`): a heal inside one sample hides a drop, alike on both backends.

## PR candidate state

(a) and (b) are proven at the function and table level. Fix shape (two claims, two commits): (b) in the MissA hit (ai.hpp:793) `src.to_hit = 1 << 20` when `!missile.info->to_hit`; (a) no block rolls in `monster_blow` for a missile whose physical roll is 0. Not built in this time box (no branch, no after-run, no verifier); neither moves the early-game gate. Row in pr-bundle-queue.md.

## Reproduce (the worker host; upstream worktree on main 0dc4c62a; tables from `mpq-cat` into a scratch dir)

```sh
cd <upstream>/tools/emu && cp <this repo>/tools/emu/monster_hit.py <this repo>/tools/emu/mon_missile.py .
D2_MPQ_DIR=<mpq dir> <venv>/bin/python monster_hit.py --exe <game.exe> 100000   # hit 0, dmg 0, gh 0 / 0 differ
D2_MPQ_DIR=<mpq dir> <venv>/bin/python mon_missile.py --exe <game.exe> --tables <tables dir> 100000
#  block: 0 differ from the reading; 11156 from d2d (canblock 0: 10503, mode 6 town walk: 653); roll: 0 differ
#  tables: 94 of 142 MissA rows without ToHit, 9 MissA2 with A2TH != A1TH, 181 unblockable skill-missile rows
```

Open: `0064b860` not run; the A2-at-impact rate; whether the d2d dump can carry missiles (B1); nothing checked live.
