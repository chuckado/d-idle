# decomp35: a player's melee riders (leech, Crushing Blow, gear elements) against game.exe 1.14d

> status: research, 2026-11-16, agent decomp35 (Sonnet), lane A of the decomposition wave, offline (the emulator and the fork builds on the NAS builder container; no game, no live d2d, no Go change, nothing posted upstream). Fork branches `fid/leech-r1` (head `99679ead`) and `fid/crush-r1` (head `de243e2b`, stacked on the first) are pushed to the private fork remote only. Follows [sk-melee-trace.md](sk-melee-trace.md) and [sk-melee2-trace.md](sk-melee2-trace.md) ("Not covered: leech, crushing blow, open wounds", still open before this card), [monsterhit-trace.md](monsterhit-trace.md) Q9 and [damage-intake-2026-11-15.md](damage-intake-2026-11-15.md).

Tags: **V run** = game.exe's own code run under the emulator (the image of [docs/emu-image.md](../emu-image.md), scripts `tools/emu/d35_*.py`) and compared with d2d's real `rules::player_blow` built from upstream main `b6e6940d`; **V read** = disassembly read (addresses given); **D** = 1.14d tables; **S** = an earlier doc of ours; **I** = inferred; **U** = unverified. Every number called V run below comes from game.exe's code, byte for byte against the rule or against d2d's compiled function.

## Topic choice (why these three)

The brief asked for the top three un-traced topics of the combat and skill list by (effect on the bot's survival and XP fidelity) x (tractability). The finished traces cover the to-hit chain, the damage roll, resist, block, the get-hit modes, Smite, Zeal, the sequences, Static Field, Frost Nova (D1 fixed by #158, the aim table is upstream's `kNovaRing`) and the player's defence. What the audit of [sk-melee-trace.md](sk-melee-trace.md) / [sk-melee2-trace.md](sk-melee2-trace.md) and the 1d row of [pipeline-2026-11-14.md](../upstream/pipeline-2026-11-14.md) left open was everything a melee hit does AFTER the physical roll: life and mana leech, Crushing Blow, Open Wounds, the gear's elemental rolls, the crit and deadly strike rolls. The first two were picked because a Paladin or any melee hero with the commonest gear affixes uses them on every swing, the third (the gear elements) because the oracle was one call away.

| Topic | Result | Branch / body |
|---|---|---|
| T1 Life and mana leech of a player's blow | three real deviations, 0 differ after the claim | `fid/leech-r1`, [decomp35-leech-body.md](../upstream/bodies/decomp35-leech-body.md) |
| T2 Crushing Blow | one real deviation (the order and the 256ths), 0 differ after the claim | `fid/crush-r1` (stacked), [decomp35-crush-body.md](../upstream/bodies/decomp35-crush-body.md) |
| T3 The gear's elemental damage roll | real but small (V run); fix not built (reason below) | none |

## Q1 to Q12 (the questions and their answers)

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q1 | Where does a player's leech come from, and what does game.exe do with it? | FUN_0057b7d0 copies the gear's life steal (stat 60) and mana steal (stat 62) into the damage record (`+0x38`, `+0x3c`) at `0x57bb00` (full value for a player attacker; a monster attacker's is scaled by the record's SrcDam share). FUN_0057c6c0 then calls FUN_0057c420 (`0x57c826`) with the record, the game and the defender. | V read |
| Q2 | Which damage is leeched? | The record's PHYSICAL (`+8`, after the resistance pass), cut to the defender's current life first: `0x57c7f8..0x57c820` loads `+8`, reads stat 6 (life, 256ths) and stores the smaller one back. Elements, crushing and poison are not leeched. | V read; V run (the clamp is inside the oracle, 3,133 of 20,351 cases differ without it) |
| Q3 | What is the arithmetic? | `FUN_0057c420`: stops when both steals are 0; a monster defender takes its MonStats `Drain` of the game's difficulty (`byte [row + 0xa0 + difficulty]`, `0x57c466..0x57c474`; 0 or less: no leech), any other defender 100. `steal << 6` (`0x57c47d`), and for a PLAYER attacker (`[attacker] == 0`, `0x57c494`) divided by the difficulty's `LifeStealDivisor` (row `+0x28` of FUN_00611d30) or `ManaStealDivisor` (`+0x2c`) (whole division). Then `MulDiv(physical, that, 100)` (FUN_00483360), `MulDiv(.., Drain, 100)` unless Drain is 100, then `/ 64` toward zero. The result is in 256ths of life (FUN_0057a980) or mana (FUN_0057aa00). | V read, and V run: the rule re-derived from the code equals game.exe in 20,351 of 20,351 and 60,351 of 60,351 cases (`leech.py selfcheck`) |
| Q4 | What are the divisors? | DifficultyLevels `LifeStealDivisor` / `ManaStealDivisor` are 1 / 2 / 3 (Normal, Nightmare, Hell). A player's leech is halved in Nightmare and thirded in Hell. | D (the table, read by the oracle at run time) + V run |
| Q5 | What does d2d do? | `rules::player_blow` computed `phys (whole points) x steal x Drain / 10000` in whole points (`blow.life`), no cap at the target's life, no divisor, and `Fight::land` added `blow.life << 8` (`strike.cpp:501`). | V read |
| Q6 | Where does a Crushing Blow run, and on which life? | FUN_005bffc0 is a strike-event handler (a draw `% 100` under the item's chance, stat 136; event 6, a missile's, doubles the divisor: `0x5c00a2`). FUN_0057c6c0 raises the attacker's event 5 (melee) or 6 (missile) and the defender's 1 or 2 at `0x57c7c7..0x57c7f3`, BEFORE it clamps the physical and takes the damage at `LAB_0057c852` (`0x57c852`). The handler reads stat 6 as it is then: the life BEFORE the hit. | V read; V run: with FUN_0057c6c0 natively and the dispatcher FUN_005c0c30 replaced by a recorder, events 5 and 1 saw the life before the damage and the end life was `life - damage` in 2,360 of 2,360 cases |
| Q7 | What does the handler take? | `life / divisor` (256ths, whole division) with the divisor 4, 8 for a boss or unique (`FUN_0063e9f0`, `FUN_005a0180`), 10 for a hireling (`FUN_0063ee90`), x 2 for event 6, plus the player-count share (FUN_005738f0); less `MulDiv(that, min(res, 100), 100)` when the target's stat 0x24 is above 0; it sets stat 6 (0 when it would not be above 0, and sets the kill bit `rec+4 \| 2`). | V read + V run (FUN_005bffc0 natively, 2,360 cases) |
| Q8 | What does d2d do? | `crushing_blow = max(hit_points - phys - elem, 0) / (divisor x (missile ? 2 : 1))`, whole points, "the life left after the hit" (`combat.hpp` comment of the parent). | V read |
| Q9 | Where is the gear's elemental damage rolled, and how? | FUN_0057b7d0 `0x57b932..0x57ba4c` calls FUN_0057a8e0 for fire (stats 0x30 / 0x31, mastery 0x149), lightning (0x32 / 0x33, 0x14a), cold (0x36 / 0x37, 0x14b) and magic (0x34 / 0x35, 0x165): min and max are `stat << 8`, each scaled by `MulDiv(x, mastery, 100)` in 256ths, the roll is `min + rand(max - min)` on the unit's seed (top exclusive, 256ths), multiplied by the SrcDam share. | V read; V run (FUN_0057a8e0 natively, 3,947 cases) |
| Q10 | What does d2d do for the gear elements? | `Fighter::elem` holds whole points scaled by `value x (100 + mastery) / 100` (truncated), the roll is `rng.range(lo, hi)` in whole points (top inclusive), resisted in whole points. | V read |
| Q11 | The crit, deadly and mastery-crit rolls | FUN_0057b7d0 `0x57b8b0..0x57b924`: only when the third argument is 0; the mastery crit (FUN_00645830 on the weapon, only with a weapon in hand) rolls first, then stat 0x151 (337), then stat 0x8d (141); each only when its chance is above 0, each a separate `rand % 100`; any hit doubles the record's physical and sets `0x2000`. d2d's `roll(mastery) \|\| roll(critical) \|\| roll(deadly)` draws in the same order and under the same conditions. | V read both sides; no run: the rolls are one LCG draw each, an order or condition error would show as a changed damage stream (not built; I that it matches) |
| Q12 | Where does Open Wounds live? | Not in FUN_0057b7d0, FUN_0057c6c0, FUN_0057c420 or the neighbours FUN_005bf960 (knockback), FUN_005bffc0 (crush), FUN_005c0130 / FUN_005c01d0 / FUN_005c0260 (life, mana, life-on-kill style events). The consumer of stat 135 was not found in the functions read; d2d's table (`open_wounds_per_sec`, "1.10's table") is unchecked against 1.14d. | U (dead end: recorded, section "Dead ends") |

## What was run (the differentials)

Oracles (also on the fork branches as `tools/emu/leech.py`, `crush.py` with their harnesses; here as `tools/emu/d35_leech.py`, `d35_crush.py`, `d35_elem.py` and the harness sources `d35_*_harness.cpp`):

| Run | Cases | game.exe vs d2d parent | vs d2d after the claim |
|---|---|---|---|
| Leech, seed 7 (`leech.py gen 20000 7`): 3 difficulties x 13 steals x 8 physical sizes, overkill rows, Drain rows, 20,000 random | 20,351 | **17,097 differ** | **0** |
| Leech, seed 1 (`gen 60000 1`) | 60,351 | **50,626 differ** | **0** |
| Leech, selfcheck (the rule read from the code against game.exe), both seeds | 80,702 | not d2d: 0 differ | not d2d: 0 |
| Leech mutants of the claim (seed 7): no divisor / no clamp to life / whole points / no Drain | 20,351 each | n/a | 11,620 / 3,133 / 17,068 / 8,376 differ (each mutant is caught) |
| Crush, seed 7 (`crush.py gen 2000 7`): 6 lives x 4 hits x 5 resists x 3 divisors, 2,000 random | 2,360 | **2,223 differ** | **0** |
| Crush order (FUN_0057c6c0 natively, the recorder) | 2,360 | the life at events 5 and 1 is the life before the damage in 2,360 of 2,360 | n/a |
| Crush mutants of the claim: no resist share / whole points of life / the life after the hit | 2,360 each | n/a | 1,258 / 1,905 / 2,241 differ (each mutant is caught) |
| Gear elements, seed 7 (`d35_elem.py gen 3000 7`): FUN_0057a8e0 vs `make_fighter` + `player_blow` | 3,947 compared (a to-hit miss on d2d's first draw is dropped) | whole points differ in **2,223** (56 percent), mean 16.50 (game) against 16.31 (d2d) | not built |

Domain of the leech and crush oracles: a player attacker, a monster defender (class 60), flags apply + hit, no hireling or boss on the defender side except the divisor rows of the crush grid, one player in the game (FUN_005738f0 = 0). The record's physical is the input; the resistance pass (FUN_0057c1e0) is left out on purpose (it is traced: [skabrecheck.md](skabrecheck.md), [sk-resist-trace.md](sk-resist-trace.md)). Tests of the branches (NAS builder, Release, GCC 14): `test_combat` ok and `test_game` rc 0 on `fid/leech-r1`; the full ctest rows for both branches are in the "Verification state" section below.

### Size, in play (I: arithmetic on the V run rule)

hp per hit healed by a melee hero (target not near death, Drain 100): game.exe against the parent d2d.

| steal | hit | Normal game / d2d | Nightmare game / d2d | Hell game / d2d |
|---|---|---|---|---|
| 5 % | 10 hp | 0.50 / 0.00 | 0.25 / 0.00 | 0.16 / 0.00 |
| 5 % | 40 hp | 2.00 / 2.00 | 1.00 / 2.00 | 0.66 / 2.00 |
| 8 % | 20 hp | 1.60 / 1.00 | 0.80 / 1.00 | 0.53 / 1.00 |
| 10 % | 10 hp | 1.00 / 1.00 | 0.50 / 1.00 | 0.33 / 1.00 |
| 15 % | 5 hp | 0.75 / 0.00 | 0.38 / 0.00 | 0.25 / 0.00 |
| 20 % | 40 hp | 8.00 / 8.00 | 4.00 / 8.00 | 2.66 / 8.00 |

So d2d under-heals small hits in Normal (the whole-point floor) and over-heals by 2x in Nightmare and 3x in Hell. The Nightmare and Hell factor is the larger effect for any hero whose gear carries steal, the floor matters most for the early game (a 6 to 8 percent leech on a 5 to 10 hp blow is 0 on d2d). Overkill: a 20 percent steal on a 40 hp blow that kills a 6 hp monster heals 1.2 hp in game.exe and 8 on d2d.

Crushing Blow example: a 100 hp monster hit for 30 with Crushing Blow: game.exe takes 25 (100 / 4) and then the 30; d2d took (100 - 30) / 4 = 17. The two agree only when the hit is 0. For a monster at full life the missing crush is `hit / 4` hp per proc (10 to 25 percent of a hit's damage at a 25 to 33 percent proc rate).

## Deviation list (d2d against game.exe)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D1 | The leech is in 256ths in game.exe (FUN_0057c420, `/ 64` of 256ths) and the life stat takes it as is; d2d floored it to whole points (`blow.life`, `phys x steal x Drain / 10000`). | V run (17,097 of 20,351, 50,626 of 60,351 differ before; 0 after) | yes, built: `leech.py`, `fid/leech-r1` |
| D2 | A player's steal is divided by DifficultyLevels `LifeStealDivisor` / `ManaStealDivisor` (1 / 2 / 3): Nightmare halves it, Hell thirds it. d2d had no divisor, so its heroes heal 2x and 3x too much in Nightmare and Hell. | V run (a mutant without the divisor differs in 11,620 cases) | yes, same oracle |
| D3 | The physical is cut to the target's current life before the leech (`0x57c7f8..0x57c820`): a killing blow leeches only the life that was there. d2d leeched the whole blow. | V run (3,133 cases differ without it) | yes, same oracle |
| D4 | Crushing Blow reads the life BEFORE the hit (strike event 5 / 6 runs before LAB_0057c852 takes the damage) and works in 256ths; d2d took the life left after the hit, in whole points. | V run (2,223 of 2,360 differ before, 0 after; order 2,360 of 2,360) | yes, built: `crush.py`, `fid/crush-r1` |
| D5 | The monsters' MonProp Crushing Blow on the player, the merc and the pets takes `crushing_blow(life - damage, ..)` after the hit (`tick.cpp:634`, the comment of `monprop_hit` `ai.hpp:~668` says "after the hit"). The same FUN_005bffc0 and the same FUN_0057c6c0 order apply to any attacker, so the game takes the life before the hit there as well. | V read (same function, same call site); the d2d side V read; not run on a player defender | buildable: `crush.py game` is attacker-agnostic; the d2d side needs a `Fight` scenario (a MonProp crush monster against the player at a known life) in `test_game`; not built (reach: Nightmare and Hell monsters with a crush MonProp) |
| D6 | The gear's elemental damage: game.exe scales min and max by the mastery in 256ths and rolls `min + rand(max - min)` in 256ths (top exclusive); d2d scales in truncated whole points and rolls whole points with the top inclusive. | V run (whole points differ in 2,223 of 3,947; mean +0.19 hp of 16.5) | yes, `d35_elem.py` + `d35_elem_harness.cpp`; fix not built: `Fighter::elem` has 15 users and the element sums, the chill mark and the poison split would all move to 256ths; the mean error is about 1 percent of an element hit, so it stays a documented deviation |
| D7 | Deadly Strike per level: d2d adds `stat 250 x clvl / 8` once over all gear; whether game.exe truncates per item is unread | U | would need FUN_00626920's op machinery run with two items |
| D8 | Open Wounds: d2d uses "1.10's table" (`open_wounds_per_sec`); the 1.14d consumer of stat 135 was not found | U | see Dead ends |

Not deviations (checked): the order and conditions of the crit, deadly and mastery-crit draws (Q11, V read, I); the Drain column per difficulty (`gamedata_load.cpp:237` reads `Drain`, `Drain(N)`, `Drain(H)`; a Drain of 0 takes nothing on both sides; V run in the leech grid); the crush divisor rows 4 / 8 / 10 and the missile doubling (`Monster::crush_divisor` equals FUN_005bffc0's branches; the grid fed all three).

## Bot findings (described, not changed)

| B | Finding | file:line | Tag |
|---|---|---|---|
| B1 | The survival and kill-time model says "crit, deadly strike and the 256ths life tail are not modelled" and has no leech or crush term, so a hero with life steal is rated as if it only healed by potion. With D1 to D3 fixed in d2d the sustain depends on the difficulty (halved, thirded), which the model would have to know. | `internal/resource/melee.go:106` | I |
| B2 | The party-scaling note for Crushing Blow ("divided by 0.5+0.5*n") is a bare constant text; game.exe's divisor is `base + MulDiv(base, FUN_005738f0(players), 100)` (V read, `0x5c008c..0x5c009f`), not checked against the bot's number. | `internal/party/scaling.go:29` | U |
| B3 | Gear scoring and the build plan have no steal or crush term (no hit on `lifesteal`, `leech`, `crushing` in `internal/`), so a leech item is never preferred over an equal one without it. | `internal/` (grep: only the two lines above mention these) | I |

## PR candidates

Two, both proven with the existing harness style (oracle first, claim second), local branches in the private fork checkout, pushed to the private remote only, nothing posted.

| Branch | Head | What | V run before / after |
|---|---|---|---|
| `fid/leech-r1` | `99679ead` (oracle `5970e2fb`, claim `99679ead`) on upstream `b6e6940d` | A player's leech in 256ths, under the target's life, over the difficulty's divisor (D1, D2, D3). | 17,097 of 20,351 and 50,626 of 60,351 differ, 0 after; 4 mutants caught |
| `fid/crush-r1` | `de243e2b` (oracle `2b0f3b80`, claim `de243e2b`), STACKED on `fid/leech-r1` (it reuses the `Target::life_fine` the first adds) | A player's Crushing Blow takes the life before the hit, in 256ths (D4). | 2,223 of 2,360 differ, 0 after; 3 mutants caught; order 2,360 of 2,360 |

Bodies: [decomp35-leech-body.md](../upstream/bodies/decomp35-leech-body.md), [decomp35-crush-body.md](../upstream/bodies/decomp35-crush-body.md). Both are ready for the independent Opus verifier (not run: no verifier ran on these, and every number is from this agent's own harness). Posting needs the verifier and the user's go; the crush branch posts after the leech PR merges.

## Verification state

- `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN and `presubmit.py --quick` 0 FAIL on both heads; the WARNs left are answered in the bodies (unit kinds, hit paths, no full build in `--quick`).
- Full ctest of the stacked tip (`fid/crush-r1`): see the line the merger finds in the card result (the build was queued behind another agent's gate on the shared builder).

## Dead ends and what was NOT done

- Open Wounds (D8, Q12): the consumer of stat 135 is not in the blow builder, the apply function, the leech or the three event handlers next to FUN_005bffc0. It is probably a state or a later event; finding it needs the item-property to event table (not read). Not claimed.
- The monster-attacker leech/crush (FUN_0057b7d0 scales a monster's steal by the SrcDam share, `0x57ba64..0x57baf6`): not run; d2d credits leech only to the player (`by_player`), the merc and pets do not leech in d2d (V read, `deaths.cpp:412` does not use `blow.life`).
- A first mistake worth recording: the first run of the crush order used FUN_0057c6c0's second argument 1 (the missile path) and saw no event; with 0 (melee) events 5 and 1 appear. The rule is the same function, the arguments differ.
- Large hits: the 32-bit product of `MulDiv` wraps for a physical over about 2,600 hp at a 50 percent steal (negative leech in the oracle's far corner); the claim reproduces it through `mul_div`, which already ports FUN_00483360.

## Drill recipe for the fork verbs (docs/forkdbg.md; nothing run)

R1 (leech, Normal): throwaway Paladin, `debug stat 60 10` (10 percent life steal; confirm with `dump stats`), `debug mon set <id> life 100000` on a dummy, hit it with Attack 40 times (`step`), read life before and after: expected gain per landed hit = `phys x 0.1` in 256ths (a 20 hp blow heals about 2.0 hp). Under the parent d2d a 5 hp blow heals 0. Then `debug difficulty` to Nightmare and Hell and repeat: the gain must halve and third on a build with this branch, stay the same on the parent.
R2 (overkill): `debug mon set <id> life 3`, one hit of 20 hp with 20 percent steal: expected heal about 0.6 hp (3 x 0.2), not 4.
R3 (crush): `debug stat 136 100` (Crushing Blow 100 percent), a dummy with `life 400`, one hit of 50 physical with no resistance: expected life after 400 - 100 - 50 = 250 on game.exe (crush from the life before the hit), 400 - 87 - 50 = 263 on the parent.
R4 (monster crush, D5): a Nightmare monster with a crush MonProp against a hero at 400 life, `debug hit` is not enough (it needs the monster's swing): `debug mon set` the damage, step; expected crush = life before / divisor.
Each R must run on a throwaway character (provenance rule), with the write set recorded in the ledger.

## Matrix-cell updates (proposed text)

- [pal-skill-matrix.md](pal-skill-matrix.md) / [sorc-skill-matrix.md](sorc-skill-matrix.md) section "Not covered": replace "leech, crushing blow" by "leech and Crushing Blow traced in decomp35-combat.md (V run, branches `fid/leech-r1`, `fid/crush-r1`; open wounds U)".
- [fidelity-matrix.md](fidelity-matrix.md): add rows "IT/SK melee riders: player's life and mana leech (256ths, difficulty divisor 1 / 2 / 3, cut to target life): DEV V run, `fid/leech-r1`" and "player's Crushing Blow (life before the hit, 256ths): DEV V run, `fid/crush-r1`".
- [monsterhit-trace.md](monsterhit-trace.md) Q9: "player leech traced in decomp35 (FUN_0057c420); monster attackers' steal scaling FUN_0057b7d0 `0x57ba64` not run".
- [damage-intake-2026-11-15.md](damage-intake-2026-11-15.md): no change; the intake path of this card is the player's output, not the player's intake. The D5 row above is the one that touches intake (a monster's crush).

## Reproduce (builder container, emulator image at `/data/prfix6-emu`)

```
python3 tools/emu/d35_leech.py gen 20000 7 > c.tsv                 # on any box, the cases only
<emu env> python3 d35_leech.py game c.tsv > g.tsv                   # game.exe, in the container
g++ -O1 -std=c++23 -I<d2d>/components/rules -I<d2d>/components -I<d2d>/components/d2s -I<d2d>/components/txt \
    -o leech_harness d35_leech_harness.cpp && ./leech_harness < c.tsv > d.tsv
python3 d35_leech.py cmp c.tsv g.tsv d.tsv                          # parent: 17,097 differ; claim: 0
```
`crush` and `elem` the same with `d35_crush.py` (`game`, `order`) and `d35_elem.py`. The fork branches carry the scripts under upstream's names (`tools/emu/leech.py`, `crush.py`).

## Shared-file lines for the merger

See `backlog/merge-lines/agent-decomp35.txt`.
