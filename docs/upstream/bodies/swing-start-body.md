## What changes

Two rules of one function (`Fight::set_pmode`, an attack mode), both the clock FUN_005539b0 files a swing's events on:

1. **Start frame (Amazon and Sorceress).** FUN_00623b10 is 0 for a monster and for a player mode other than A1 and A2 (S3 and S4 only for a Barbarian or Assassin: FUN_006235a0 is a class test). For an A1 or A2 swing it looks up the weapon in hand (FUN_00623990: slot 4, else slot 5, a weapon that is not broken (item flag 0x100) and whose requirements are met (0x4000): FUN_0063c9b0 -> FUN_0062a4e0), its class key (FUN_00629fe0: the weapons.txt `wclass` through the map at 0x7446a0) and reads the dword at `0x6e8e60 + 4 (row x 7 + player class)`. Only an Amazon and a Sorceress are not 0: frame 1 barehanded, frame 2 with a 1hs, 1ht, stf, 2hs or 2ht weapon, 0 with a bow, crossbow or claws. d2d played the whole animation, so these two classes' melee swings were 1 to 3 ticks too long (a Sorceress staff swing 18 ticks against 16, about 12 percent of her melee damage per second).
2. **The strike tick (every class).** FUN_005539b0 walks the animation from the start frame at the unit's rate and files the action event on the tick its AnimData flag frame is reached: `max(1, ceil(256 (flag - start) / rate))`, the end event at `max(1, ceil(256 (frames - start) / rate))`. d2d scaled the action frame by the rounded swing length instead, so its strike was one tick late in 1,068 of 8,120 cells of the five classes without a start frame (the same cause as the one-tick residual left in Zeal's chained strikes after `276f85c0`).

`rules::attack_start_frame` holds the table (rows by weapon class, 7 player classes, as `swing_start.py table` prints it from the image); `rules::frame_tick`, `attack_ticks` and `attack_action_tick` the clock; `ItemBase::wclass` the weapon's class (weapons.txt); `Fighter::attack_start` the start frame of the usable weapon in hand (`player_fighter`'s `weapon`: slot 4, else slot 5, broken and requirement-short ones skipped) `Fight::action_at()` the strike time the three strike sites read (an attack mode's `action_ms`, any other mode's old formula).

Scope kept narrow: the start frame applies to a plain swing and to a skill whose `anim` is A1 (Power Strike, Charged Strike, Lightning Strike, **Fend**, the bow skills; `Fight::set_pmode` is also reached for a throw, TH, which plays the A1 record in d2d: it gets the strike rule only, start 0, as the game's TH mode has). Jab and Impale (`anim` SQ, `start_sequence`) are unchanged. Kicks (KK) and the S modes get the strike rule only.

Two commits: the oracle (`tools/emu/swing_start.py`, `tests/game/test_swing_start.cpp --swing-start`, README row; it prints what `set_pmode` does on the parent), then the claim (`rules/combat.hpp`, `rules.hpp`, `gamedata_load.cpp`, `fight.hpp`, `fight/setup.cpp`, `fight/strike.cpp`, the test's rule checks and 32 frozen cells, `docs/research/re/combat.md` "A player's attack swing").

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00623b10` `0x623b10..0x623bfa` | the start frame: not a player 0; a player: the byte table at `0x623c10` (modes 7 to 16) sends A1 and A2 to the weapon path `0x623b88`, S3 and S4 to it only when `FUN_006235a0` (a class test: a Barbarian or Assassin) is true (`0x623bbc`), every other mode to 0 (`0x623bfa`) |
| `FUN_00629fe0` and the map at `0x7446a0` | the item record's class key (+0xc0) to a row: bow 1, 1hs 2, 1ht 3, stf 4, 2hs 5, 2ht 6, xbw 7, ht1 12; no weapon or a key the map lacks: row 0 (`0x62a040`) |
| table `0x6e8e60` (dumped from the booted image, `swing_start.py table`) | 7 dwords a row, d2s class order: row 0 `1 1 0 0 0 0 0`; rows 2 to 6 `2 2 0 0 0 0 0`; rows 1, 7 (bow, crossbow) and 12 (claws) all 0 |
| `FUN_00623990` -> `FUN_0063c9b0` -> `FUN_0062a4e0` `0x63c9fb..0x63ca8f`, `0x62a4f8..0x62a516` | the held weapon: slot 4 if type weap and usable, else slot 5 under the same test, else none; unusable = flag 0x100 or 0x4000 (V read) |
| `FUN_005539b0` `0x5539b0..0x553afc` | `start = FUN_00623b10 << 8`, rate unit `+0x4c`, end `+0x48`; files an event for each flagged AnimData frame the walk passes (flags 1, 2, 4 take one branch of the table at `0x553afc`, 3 another) and the type-1 end event when the position reaches the end (filed at `frame + max(n, 1) + 1`: a floor of 2 ticks, never reached by a player record) |

## Counted differentials (V run, tree = upstream `f9b8cfed` plus the branch)

Oracle: `swing_start.py compare` runs `FUN_00623b10`, `FUN_00629fe0` and `FUN_005539b0` natively with the real AnimData flag bytes for each cell `test_swing_start --swing-start` prints: 7 classes x main-hand weapons (none, sword, dagger, great sword, spear, staff, bow, crossbow; claws for the Assassin only) x the modes d2d plays (A1, S1, S3, S4, KK) x IAS {0, 10, 20, 30, 40, 60, 80, 100, 150, 200} x weapon speed {-30 .. +30}, with a hero whose stats meet any requirement: 11,270 cells that have an AnimData record and a flag. The rate fed to the game side is d2d's own `attack_speed` (the rate itself: `player_attack_rate.py`). The oracle also checks the closed form against the game's recorded events: 0 differ on the strike and on the end.

| 11,270 cells | differ before (parent) | differ after |
|---|---|---|
| Amazon / Sorceress A1 with a start frame (840): swing end | **815** (d2d 1 to 3 ticks longer) | **0** |
| the same 840: strike tick | **834** (+1 to +4) | **0** |
| Amazon / Sorceress, bow, crossbow, kicks (the strike rule only) | 108 strike | **0** |
| Barbarian, Necromancer, Paladin, Druid, Assassin (8,120): swing end | 0 | 0 |
| the same 8,120: strike tick | **1,068** (always +1 tick) | **0** |
| all | **2,011** cells (815 ends, 2,010 strikes) | **0** |

An independent grid through d2d's REAL attack path (`start_swing` -> `set_pmode` -> `player_modes` ticked every 40 ms -> `strike`; a Fallen beside the hero in the Blood Moor; 7 classes x gear sets (including a broken or requirement-short item in hand, dual wielding with a broken right hand, a weapon in the left hand only) x IAS x each weapon's own speed modifier x every class skill whose `anim` is A1, A2, S1 to S4, KK or TH plus the plain attack; game side a separate oracle script with its own parser and the held-weapon rule applied to the flags d2d actually holds), 78,172 cells printed, 3,520 without an animation record skipped: of the 51,056 reachable cells (wearable gear, a usable skill, a single swing) the parent differs on 18,064 and this branch on **0** (the first revision of the branch differed on 5,940, all Amazon or Sorceress cells with a broken or requirement-short item in hand: the defect the held-weapon rule fixes). Class-restricted gear (7,280 cells), throw skills' plain-attack fallback (480) and skills unreachable in play (10,560) also show 0. The chained skills (Zeal, Fury, Dragon Talon: 5,276 cells) are not comparable in that probe (it times the whole chain against one swing; 4,976 differ on both trees); Zeal's per-hit gaps are covered by `test_zeal_chain` and the local harness mentioned below.

Per cell before (one line each): Sorceress, staff, IAS 0: swing 18 and strike 11, game.exe 16 and 9; barehanded 16 and 9 against 15 and 8; sword 20 and 12 against 18 and 10. Amazon sword 16 and 10 against 14 and 8; spear 18 and 11 against 16 and 9; bow 14 and 6 on both.

Build and tests (V run, GCC 16.2.1, Release, `D2D_WERROR=OFF`): `presubmit.py --quick --base upstream/main` 0 FAIL (7 WARN, answered below), `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN, ctest 296 of 300 pass, the 4 failures the same as on a pure upstream build on my box (`test_mpq`, `smoke_d2d`, `quests_act1`, `quests_act4`: no video MPQs here). Each commit builds alone: Linux GCC only, no Windows, macOS or Clang build run (U).

Unit test (`test_swing_start`, no argument): the table and the clock without data (`attack_start_frame` for every class x weapon class, a key the map lacks, `frame_tick`), and with the data 34 cells whose answers the oracle printed: 32 of the grid (Sorceress and Amazon with each weapon class, and the Paladin, Barbarian, Necromancer, Druid and Assassin cells where the strike was a tick late) and two unusable weapons (a broken sword on a Sorceress, a great sword she lacks the strength for: the swing starts as with nothing in hand, on that weapon's record). The parent has no `attack_start_frame`, so the test is built against the claim only; on the parent the same cells give the "d2d before" numbers in the comments of `kFrozen` (15 of the 32 were looked up in the parent's grid: they differ), and the oracle's comparison is the check on the parent.

Zeal side effect (V run, a local harness against game.exe's FUN_00553b10, not part of this PR): parent 50 of 150 cells (200 of 600 gaps) one tick late, claim **0 of 150 (0 of 600)**: the one-tick residual left in Zeal's chained strikes is this rule.

## Collateral (what else changes)

- Same function, same clock: Amazon A1 skills with a melee weapon in hand (Power Strike, Charged Strike, Lightning Strike and **Fend**, whose `anim` is A1: its swing takes the start frame too, 16 to 14 ticks at 0 IAS with a sword); with a bow or crossbow only the strike rule applies (Magic Arrow, Fire, Cold, Ice, Exploding, Immolation, Freezing, Guided Arrow, Multiple Shot, Strafe). Sorceress: only the plain attack (her skills are SC). Jab and Impale (SQ) keep their length.
- A throw (TH: Poison, Plague and Lightning Javelin, Lightning Bolt, Lightning Fury) now strikes on `attack_action_tick` with start 0 instead of the old formula: one tick earlier in 56 of 360 cells of an independent probe, the swing length unchanged. game.exe plays the TH record there, which d2d does not (it plays A1), so this is neither closer nor further from game.exe in a way the oracle can show.
- A hit-recovery or a cast does not change (SC, GH, BL take the unchanged branches). Whirlwind's gap reads `attack_ticks` of the Barbarian's A1: start 0, same value.
- Not changed, U or out of scope: A2 (d2d never plays it), the S3 and S4 start of a Barbarian or Assassin (table entries are 0), what the client draws for the skipped start frames (d2d still draws the whole animation over the shorter swing; not checked here), the idle tick between held swings. The end event's floor in game.exe is 2 ticks, d2d's is 1: they differ only when one tick covers the whole rest of the animation (rate >= 256 (frames - start)), which no player record reaches (attack rates stay under 450, frames - start is at least 6).

## Limits

- No kill-rate or damage-per-second effect is claimed. The claim is the swing length and the strike tick against game.exe's. (Ten-minute bot runs on both builds moved by less than their run-to-run noise.)
- What the client draws for the skipped start frames is not checked (U): d2d still draws the whole animation over the shorter swing.

- The rate itself is not re-run here (`player_attack_rate.py` covers it). The game's choice of animation record for a weapon is U (d2d's COF name comes from the worn weapon).
- A weapon whose class key is not in the 8-entry map takes row 0 (the nothing-in-hand row) as in game.exe (V run, native sweep: key `zzz` gives 1 for an Amazon or Sorceress); no vanilla weapon has one (every weapons.txt `wclass` value is in the map).
- The held weapon is the usable one of slot 4, else slot 5 (V read); d2d's `weapon` in `player_fighter` is that choice. Class-restricted gear is not worn in d2d (the 630 cells above).

## Answers to the presubmit WARNs

- Unit kinds (player, not merc, pet, monster, co-op partner): the rule is the player's `Fight::set_pmode`; a co-op partner is another `actor` of the same code and gets it. Monsters, mercenaries and pets time their swings through `npc_timing` (`ai.cpp`, `fight/merc.cpp`, `fight/pets.cpp`); FUN_00623b10 returns 0 for a unit that is not a player, so their start frame is 0 on both sides and their own strike rule is a separate question (not checked here).
- Hit paths (melee, skill, not missile): an arrow or bolt skill is an A1 swing of the same class: the strike rule applies to it (three strike sites read `action_at()`), the start frame is 0 for a bow or crossbow; the missile's own flight is not touched.
- error-path: `swing_start.py` checks the AnimData blob's length before each bucket and record (`SystemExit`).
- doc-claims (combat.md "every / never"): reworded; the remaining words are about player records, checked in the grid.
- ruff is not on the box: ruff run from a venv on `swing_start.py` passed.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_swing_start
cd tools/emu && uv sync
set -o pipefail
../../build/tests/test_swing_start --swing-start | grep '^D ' > /tmp/swing.tsv
uv run python swing_start.py compare /tmp/swing.tsv      # last line: 11270 cells  same (parent: 2011 differ)
uv run python swing_start.py table
../../build/tests/test_swing_start
```

---

Agent: Claude Code
