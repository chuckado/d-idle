# decomp38 (lane D): shrine row roll and Mana Burn on a monster's hit: game.exe 1.14d vs d2d vs the bot

> status: current (2026-11-17, card decomp38, agent d2-decomp on Sonnet). Offline: no game, no d2d slot, nothing posted, no Go changed. The game.exe side ran under the emulator and the d2d side ran as fork builds, both on the builder container (nothing on the desktop worker, nothing heavy on the head box). Tags: **V run** (game.exe's own code under the emulator, compared with a d2d build of the real code; counts of cases), **V read** (disassembly read, `FUN_` addresses of 1.14d `Game.exe`), **D** (tables as game.exe loaded them), **S** (an earlier proof of ours or upstream's, named), **I** inferred, **U** unverified. Scratch tooling here: `tools/emu/dc38_shrine.py`, `tools/emu/dc38_burn.py` (the upstream-style copies live in the fork branches below).

## Answer in eight lines

1. **Scope check.** Lane D asked for item drops, shrines/chests/objects and monster special abilities. Item drops, gold, quality, affixes and chests were already run against game.exe for Act 1 and 2 ([treasure-trace.md](treasure-trace.md): 100,000 rolls, 1,012 of 1,012 class tables, 20,000 chests, 39,975 of 39,980 items equal; the five misses are the known magic elixirs): not repeated. Shrine effect arithmetic is covered upstream (`objects_mode_calc.py`, 1,600 cases) and the monster-mod spawn stats too (`umod_stats.py`). Picked: the shrine row roll (the one shrine function with no oracle) and what a Mana Burn monster's hit does (the runtime half the spawn trace never ran).
2. **Shrine row roll, no deviation (V run).** `FUN_0054f9d0` (with `FUN_0054f770` and the class lists `FUN_00546c60` builds) against `rules::roll_shrine`: 200,000 cases (two seeds, Parm0 0/1/2/3, every Levels.txt id 0..136, the object seed), the row and the seed after: **0 differ**. Branch `fid/shrine-roll-oracle-r1` (oracle only, nothing to claim).
3. **Mana Burn takes 256 times what d2d takes (V run, D1).** `FUN_005a1f90` stores the burn as `x << 8` in stats 62 / 63 and the damage build rolls them with another `<< 8` (`FUN_0057a8e0` in `FUN_0057b7d0` at `0x57ba6c..0x57baac`), so a hit takes `x_lo << 16` up to `x_hi << 16` in 256ths: **x_lo * 256 up to x_hi * 256 mana points**. d2d takes x_lo to x_hi points. A Normal Act 1 unique (level 9: x = 4..7) takes 1,024 to 1,792 mana in game.exe and 4 to 7 in d2d. 1,750 of 2,000 and 1,800 of 2,000 cases (two seed sets) differ on the parent, **0 after** the fix, and the rule alone equals game.exe byte for byte on 4,000 of 4,000. Branch `fid/mana-burn-amount-r1` (two commits: oracle, claim), body draft [mana-burn-amount-body.md](../upstream/bodies/mana-burn-amount-body.md).
4. **A Mana Burn monster heals itself (V run, D2).** `FUN_0057c420` for a plain monster attacker caps each leech field by what the defender has and heals the attacker by the sum, up to its max life: a mana burn hit heals `min(burn, the player's mana)` life points. d2d has no such heal (V read). Game side V run on the chain; no d2d differential yet.
5. **Two smaller rules (D3, D4):** the player's mana after the burn is 0 when it would be under 1 point (`FUN_0057aa60` `0x57aa76..0x57aa7e`), and the roll is in 256ths, not whole points. Both vanish into D1 except at the very end of a mana pool.
6. **The bot already treats Mana Burn as a mana wipe, but only in Hell** (B1, B2): the tracker and the potion rule are gated on `Diff != Normal`. The rule above says a Normal unique does it too (a 4..7 x 256 burn on an 80-mana Sorceress is the whole pool).
7. **Not done:** the heal fix (needs a d2d monster-life differential), the other abilities of the brief (Cursed, Spectral Hit, Aura Enchanted, Teleportation, Stone Skin, Berserk): one-line status in section 6, none traced here.
8. Shared-file lines and dead ends at the end.

## 1. Method

- Emulator of `docs/emu-image.md` on the builder (`skill_tips.boot()` loads the patched tables; real functions are called with fake units; hooks only where named). d2d side: a new test file per topic in the fork (`tests/game/test_shrine_roll.cpp`, `tests/game/test_mana_burn.cpp`), built on the builder (`-j3`).
- Branch shape (pipeline S2): commit 1 oracle (script + test mode + README row; builds alone, prints what the parent does), commit 2 claim. Identity: the user's noreply identity as author and committer, `Agent: Claude Code` the only trailer after a blank line (`git interpret-trailers --parse`). Compare scripts exit 1 on a mismatch, an empty file, a short file or a malformed line.
- Both branches are pushed to the private fork remote (the same day), not to the GitHub fork copy, nothing posted.

## 2. Q1: which Shrines.txt row a shrine object is (shrine-trace D4)

### Rule (V read)

`FUN_0054f9d0` (ECX = the init record {game, unit, room, seed pointer, objects row}; Parm0 = objects row `+0x178`): the level id is `FUN_0061a1b0(room)`. Parm0 0 (`0x54fa64..0x54fae2`): up to 8 tries of `1 + rand(rowcount - 1)` (`FUN_00641500` the row count, `FUN_006414b0` the row, LevelMin at `+0xb4`), stopping when the level id is not below LevelMin. Parm0 1 -> class 2, 2 -> class 3, otherwise one draw `% 10 == 0` -> class 1 else class 4 (`0x54fa01..0x54fa58`), then `FUN_0054f770(game, level id, ..)` with the same 8-try loop over the class's row list (`[block + 8 + class*4]` the count, `[block + 0x28 + class*4]` the list; the class byte is Shrines row `+0xb2`); a class whose count is 0 is a fatal check (`0x54f83b`). Afterwards `5 -> 3`, `4 -> 2`, `0x10 -> 0x12` (`0x54fae4..0x54fb06`), the row byte goes to the unit's data `+4` and `FUN_00621bb0` applies it. The object seed is the first eight bytes of the class block (`FUN_00546fa0` returns it; the init record's seed pointer is the same one).

### What ran (V run)

`tools/emu/dc38_shrine.py game` (upstream-style copy `shrine_roll.py` in the fork branch): `FUN_00546c60` builds the real class lists (hooked: the pool allocator `FUN_0040b430`, `FUN_00650e30/40`, the memset `FUN_00681ef0`), `FUN_0054f9d0` runs for real, only `FUN_0061a1b0` (the level id of the case) and `FUN_00621bb0` are hooked. 100,000 cases per seed (seeds 1 and 7): `parm0 level seed_lo seed_hi -> row seed_lo seed_hi`. The d2d side is `rules::roll_shrine` over the loaded Shrines.txt (`test_shrine_roll --shrine-roll`): **0 of 100,000 and 0 of 100,000 differ**, the seed after included. The grid is not a trivial one: rows 1..22 all occur (Parm0 0: 809 to 1,652 each; Parm0 1 always 2, Parm0 2 always 3, Parm0 3 the boost and magic rows). Shrine-trace D4 is closed.

### The reset timer (shrine-trace D2) (V read, S)

`FUN_00583c70` (the operate): `[unit data + 0xc]` set from the player's object id, mode 1, the message string `3683 + row`, the effect through the table at `0x6e1850` by Code, then the reset: byte `[Shrines row + 0x10]` times `0x4b0` (1,200) plus the current frame plus 1 as a timer event of kind 5 (`0x583de1..0x583e0b`). d2d's `shrine_reset_frames` is `Reset x 1200 + 1` (S, `world/tick.cpp:459`): the same. The timer-5 handler that puts the mode back to 0 was not read: U.

## 3. Q2: what a Mana Burn monster's hit does (the runtime half of MonUMod 25)

### The chain (V read, then V run)

| Step | Address | Rule | Tag |
|---|---|---|---|
| spawn | `FUN_005a1f90` `0x5a2088..0x5a20c8` | `x = MonLvl DM x MonUMod constant% / 100`, stored with the setter as `x << 8` in stats `0x3e` (min) and `0x3f` (max) (unique rows `0x1c`/`0x1f`, minions `0x10`/`0x13` plus the difficulty) | V run (`umod_stats.py`, S upstream; here stat 62 = 1,024 for DM 7, level 9) |
| stat read | `FUN_00625480` | stats 62 / 63 have ValShift 0: the getter returns the stored `x << 8` (checked: setter 1024 -> getter 1024 on a fake stat list, stat 6 likewise in 256ths) | V run |
| record | `FUN_0057b7d0` `0x57ba64..0x57baac` | attacker type 1, not a hireling (`FUN_0063ee90` 0), record flag bit 8 clear: `rec +0x3c += FUN_0057a8e0(stat 0x3e, stat 0x3f)` scaled by the SrcDam byte / 128 (0x80: x1) | V run |
| roll | `FUN_0057a8e0` / `FUN_0057a880` | `(min << 8) + rand((max << 8) - (min << 8))`, the unit's seed, bound exclusive; on the stored values that is `(x_lo << 16) + draw` | V run (model: 1,500 of 1,500 rolls equal, the seed after one earlier draw) |
| apply | `FUN_0057c6c0` `0x57c879..0x57c87e` -> `FUN_0057aa60` | the defender's mana (stat 8, 256ths) minus rec `+0x3c`; a result under `0x100` is set to 0 | V run (4 chain rows) |
| leech step | `FUN_0057c420` | for a plain monster attacker: life min(`+0x38`, damage), mana min(`+0x3c`, the defender's mana), stamina min(`+0x40`, the defender's stamina); the sum (times Drain / 100, 100 vs a player) heals the attacker up to its max life (`0x57c60a..0x57c68a`, the adder `FUN_006272b0`) | V run (40 isolated rows, 4 chain rows) |

Two worked rows (chain, V run): a Normal level 9 unique (stat 62 / 63 = 1,024 / 1,792) vs a player with 2,000 mana: 976 mana left; the same hit on 80 mana: 0 mana, and the monster gains 80 life points.

### What ran (V run) and the d2d result

`tools/emu/dc38_burn.py chain` (fork copy `mana_burn.py game`): the real gate `FUN_0057dbf0` and resolver `FUN_0057d4f0` (the chain of `vse_run.py` X3, record `+0 = 0`) on a monster attacker (type 1) and a player with 60,000 mana, stats 62 / 63 = `lo << 8` / `hi << 8`, the stat code a dict (identity, checked), events and the reach hooked; the attacker's unit seed per case. Output: the mana lost in 256ths. d2d side: `test_mana_burn --mana-burn`, the real `monster_update` on a zombie standing at a foe with `Monster::mana_lo / mana_hi` set (what `make_boss` computes: whole points).

| set | rows | parent d2d outside game.exe's range | claim | rule alone, byte for byte |
|---|---|---|---|---|
| seeds 1..250 x 8 ranges | 2,000 | 1,750 | 0 | 2,000 of 2,000 |
| seeds 1001..1200 x 10 ranges (reseed) | 2,000 | 1,800 | 0 | 2,000 of 2,000 |

Range = `x_lo << 16` up to (not incl.) `x_hi << 16`; the 250 / 200 rows with x = 0 agree on both sides. The game's own rows equal the model on all 4,000. Mutant (`<< 15`): 1,750 of 2,000 differ, exit 1 (caught). Game means per range (points): (4, 7) 1,403.8, (25, 40) 8,386, (100, 150) 31,964; d2d: 5.5, 32.5, 125.5.

## 4. Deviation list

| D | Rule, game.exe | Rule, d2d | Tag | Differential |
|---|---|---|---|---|
| D1 | Mana Burn hit: `x_lo * 256 .. x_hi * 256` mana points | `x_lo .. x_hi` points (`ai.cpp` `foe.mana_burn += rng.range(..)`, `fight/tick.cpp` `<< 8`) | game V run, d2d V run (4,000 rows), PR candidate | yes, built: `mana_burn.py` + `test_mana_burn`, branch `fid/mana-burn-amount-r1` |
| D2 | the plain monster attacker heals by `min(burn, defender's mana)` (plus life / stamina leech) | no heal | game V run (chain), d2d V read | could be built: `monster_update` on a Mana Burn zombie vs a player with known mana, read `Monster::hit_points` before and after; oracle `dc38_burn.py chain` extended with the attacker's life (the adder `FUN_006272b0` hooked as a dict add; without it nothing shows) |
| D3 | mana after the burn under 1 point is set to 0 (`0x57aa76`) | `max(mana - burn, 0)` keeps up to 255 fine units | game V run (isolated), d2d V read | folded into D1's test with a player at a small mana |
| D4 | the roll is in 256ths on the unit's seed, SrcDam scales it | whole points on the monster's rng, no SrcDam | I (after the claim: the amount is in 256ths, the SrcDam byte is the default 0x80 for a melee hit) | the claim covers the 256ths; SrcDam only for skills with a non-default byte, not run |

## 5. Bot findings (described, NOT changed)

| B | Finding | File:line |
|---|---|---|
| B1 | the Mana Burn rule (a mana potion against a Mana Burn unit, flee when 20 percent of the pool is lost in 3 s) is gated on a non-Normal difficulty; a Normal unique with the mod wipes a small pool in one hit too (D1) | `internal/resource/hell.go:82` (`in.Diff != d2.Normal`), constants `:14-18` |
| B2 | the target rank and the "avoid" list treat Mana Burn as a Hell danger below clvl 75 only | `docs/combat.md:40-41` (the rank and danger text; the code reads the same table) |
| B3 | nothing models the heal (D2): a Mana Burn unique that hits a caster with a full pool gains the pool as life, so a kill estimate from its life alone is too short | no field for it in the combat plan (`internal/combat/fight.go`, the plan rules in `internal/combat/rules_test.go:16-54` name KillFirst only), `docs/combat.md:40` |

## 6. The rest of the brief's list (status, nothing run)

| Ability | Status here |
|---|---|
| Fire / Lightning / Cold Enchanted | spawn values run (`umod_stats.py`), Cold length PR #445 merged; pulse behaviour (what the enchanted element does per hit) not run |
| Mana Burn | this doc |
| Cursed (Amplify) | d2d `cursed_hit` (`ai.cpp:2759`) mirrors `FUN_005a2530` by its comment (seed draw `& 3`, level `mlvl / 5 + 1`, radius from `aurarange`); no emulator run: S |
| Teleportation, Stone Skin, Spectral Hit, Aura Enchanted, Berserk | spawn stats run upstream; runtime hooks not run |
| Extra Fast / Strong pacing | S (monster-mods-trace, `monsterpace` merged) |

## 7. PR candidate

`fid/mana-burn-amount-r1` (head `d7a97516`, two commits on upstream `ed007569`: oracle `a6093f9d`, claim `d7a97516`): D1 proven with the existing harness style (parent 1,750 / 1,800 of 2,000 differ, claim 0, exact 2,000 of 2,000 twice, one mutant caught). Draft body: [../upstream/bodies/mana-burn-amount-body.md](../upstream/bodies/mana-burn-amount-body.md). Verifier not run (own harness, d2-reviewer singly, before any posting); the claim was built and run at the equal scratch commit `1b43c1e7` (same code; only the README row context differs). Not run: ctest of the whole suite (U: same-failures check). `fid/shrine-roll-oracle-r1` (`29e31f6d`) is an oracle-only branch (no claim): useful as a regression test upstream if the maintainer wants it; not a PR on its own.

## 8. Drill recipe (the fork verbs, `docs/forkdbg.md`, `docs/forkcheats.md`; nothing run here)

| U | Recipe |
|---|---|
| D1 in the real game | `debug spawn zombie1 1 unique 25` (or the lab's mod-spawn form), `debug mana full` on a Sorceress, let it land one hit, `debug stat 8` before and after: game.exe should leave 0 (a Normal level 9 unique takes about 1,000 to 1,800 mana); d2d before the fix loses 4 to 7 |
| D2 | the same spawn with `debug mon set <id> life <n> maxlife <n+500>`: read the monster's life after one hit that burns mana (it should rise by the player's mana, capped by max life) |
| D3 | `debug stat 8 <256 x 1 + 100>` (1.4 points) and one hit: mana 0 |
| shrine reset (D2 of the shrine trace) | `debug operate <i> 1` (Reset 2), `step 2400`, `debug objects`: back to NU; row 15 never |

## 9. Matrix cell updates (proposed text; the owners apply them)

- `docs/research/fidelity-matrix.md`, a row "monster Mana Burn (hit)": game.exe `x << 16` fine on the monster's seed, d2d whole points (D1), no monster heal (D2); status "PR candidate, verifier pending"; bot B1.
- `docs/research/shrine-trace.md` section 6: D4 "closed, V run 200,000 cases, 0 differ (dc38-lane-d.md)"; D2 "V read FUN_00583c70, equals `Reset x 1200 + 1`; timer-5 handler not read".
- `docs/upstream/pipeline-2026-11-14.md` row P14 (shrines): "D4 closed; D5 / D6 covered upstream by `objects_mode_calc.py`".

## Dead ends and why

- Item drops, chests, gold, quality: already V run (treasure-trace.md); a rerun would repeat the 100,000 rolls.
- Shrine operate reach (`OperateRange`, shrine-trace D1): the distance check is not in `FUN_00584420` (it only tests a reach flag and the cursor item); the packet handler that checks the range was not found in the time box. Open.
- The first chain run showed no monster heal: the adder `FUN_006272b0` was not hooked and went to a missing stat list. With the adder as a dict add the heal appears (section 3). Note for any future run of this chain.
- Reading the umod trace as "stat 62 is in whole points" would have made the claim wrong by 256; the stat's ValShift 0 and the identity getter settle it.

## Shared-file lines for the merger

- INDEX: `docs/research/dc38-lane-d.md` | shrine row roll (0 differ) and Mana Burn on a monster hit (x256, heal) | decomp38
- BACKLOG: decomp38 lane D: shrine row roll equals game.exe (200,000 cases); Mana Burn hit takes 256 times d2d's amount (fid/mana-burn-amount-r1, verifier pending), monster heals by the burn (U, no d2d differential).
- CHANGELOG: research: decomp38 (shrine row roll, Mana Burn).
