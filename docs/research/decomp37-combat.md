# decomp37 (lane C): gear absorb, Open Wounds and the player's magic resist: game.exe 1.14d vs d2d vs the bot

> status: current (2026-11-17, card decomp37, TASK-607, agent d2-decomp on Sonnet). Offline: no game, no d2d slot, no Go changed, nothing posted upstream. The game.exe side ran under the emulator and the d2d side as fork builds, both on the builder container (nothing on the desktop worker, nothing heavy on the head box). Follows [decomp35-combat.md](decomp35-combat.md) (open items D5, D6, D8) and [dc36-lane-b.md](dc36-lane-b.md); consumer: [../upstream/pr-queue.md](../upstream/pr-queue.md) (the verifier queue). Scratch tooling here: `tools/emu/dc37_absorb.py`, `dc37_ow.py` (the fork branches carry the same files under upstream-style names `gear_absorb.py`, `open_wounds.py` with their harness sources).

Tags: **V run** = game.exe's own code under the emulator (`docs/emu-image.md`), compared byte for byte with a d2d build of the real code; **V read** = disassembly read (addresses given); **D** = 1.14d tables; **S** = an earlier doc of ours; **I** inferred; **U** unverified.

## Answer in eight lines

1. **Picked (scope: elemental pipeline, to-hit, skill damage; impact x tractability):** the survey below found most of the scope already traced with 0 differences (resistance, pierce, immunity, to-hit, block, avoid, auras, skill element damage). The three open, tractable items are: **(T1) a player's gear absorb** (stats 142..149), **(T2) Open Wounds** (stat 135, the open item of decomp35 D8), **(T3) the player's magic resistance** against a magic blow.
2. **T1 is a missing mechanic.** FUN_0057bf80 ends with FUN_0057bf10 on every fire, lightning, cold and magic field: after the flat magic reduction and the resistance it takes the absorb percent (capped at 40) and then the flat absorb. d2d reads none of the eight stats, so absorb gear does nothing. V run: 60,000 cases (seed 1) 39,678 differ on the parent, 0 on the claim; 20,000 cases (seed 7) 13,189 differ, 0 after; five mutants of the claim are caught by the oracle and abort `test_combat`. Branch `fid/gear-absorb-r2`.
3. **T2 found the Open Wounds handler** (decomp35 could not): the item event table at `0x7325b0` entry 15, `FUN_005bfe60`, the neighbour of Crushing Blow's `FUN_005bffc0`. d2d's table of levels is right (the 1.10 bracket 9 / 18 / 27 / 36 / 45 plus 40 equals game.exe at all 99 levels), but d2d (a) never halves it on a champion, a unique or a superunique (game.exe halves it, mask 12 of the monster's data word `+0x16`) and (b) cuts it to whole points a second (up to a quarter of the wound at level 1). V run: 21,782 cases (seed 1) 16,463 differ on the parent, 0 on the claim; 2,000 cases (seed 7) 1,748 differ, 0 after. Branch `fid/open-wounds-r2`.
4. **T3, the player's magic resistance (stat 37, with the difficulty penalty and the cap) is applied by game.exe to a magic field of a monster blow; d2d has no input for it** (`Fighter::res` has four entries, `monster_blow` skips type 4). V run: 1,550 of 3,000 random magic cases differ. Reach: Act 5 succubi El rows, so no early effect; recorded, not built.
5. **Bot finding:** the resistance model caps at 75 before it adds the difficulty penalty (`internal/combat/rules.go:311`, `internal/readiness/model.go:357`); game.exe adds the penalty first and caps after (V read `0x57be3d..0x57bf02`). A raw 90 in Nightmare is 50 in game.exe and 35 in the bot.
6. **Not traced again (already clean, so skipped):** resistance and pierce (sk-resist-trace), immunity and the one-fifth rule (curse-on-immune), hit chance and the 5/95 clamp (monsterhit-trace Q2), block, avoid, dodge (monsterhit-trace Q4), Holy Fire / Shock / Freeze pulse and filter (palaura-trace), skill element damage (decomp29-skilldamage: 24,750 cases, 0 differ). Evidence list in section 1.
7. **PR state:** two local branches on the private fork remote, none posted. Both ready for the independent verifier. Neither is a large effect for the bot's current builds (see "Impact" in each body); both are real, V run, and small.
8. **Dead ends and open items** (decomp35 D5, D6) are in section 8.

## 1. Survey: what the scope's other topics already have (so they were not redone)

| Scope item | Existing evidence | Status |
|---|---|---|
| Resistance application: cap, negative resist, immunity (100), pierce | [sk-resist-trace.md](sk-resist-trace.md) (147,628 and 101,626 cases), [verify-skresist.md](verify-skresist.md); merged #135, #160 | clean (V run, monster defender) |
| Resistance on a player: difficulty penalty 0 / -40 / -100, cap 75 + max stat | [playerdef-trace.md](playerdef-trace.md) section 7b, [model-ab-skresist.md](model-ab-skresist.md) Q10; FUN_0057be00 re-read here | clean for the four elements; magic: T3 |
| Elemental mastery order | decomp35 Q9 (FUN_0057a8e0, 3,947 cases): D6 small | open, not built (section 8) |
| Absorb (flat and percent) | not traced before; `FUN_0057c1e0` row table at `0x732980` carries the absorb stats | **T1** |
| Damage to monster vs to player for fire / cold / lightning / poison, duration | [monsterhit-trace.md](monsterhit-trace.md), [sk-resist-trace.md](sk-resist-trace.md), [freeze-rules-trace.md](freeze-rules-trace.md); poison replacement P08 | clean except T1 / T3 |
| To-hit vs defence, order, 5/95, level difference, ranged vs melee | [monsterhit-trace.md](monsterhit-trace.md) Q2 / Q3 (100,000 cases 0 differ), [playerdef-trace.md](playerdef-trace.md) 7b (102,604 block, defence, AR cases) | clean |
| Block / avoid / dodge / evade and their order | [monsterhit-trace.md](monsterhit-trace.md) Q4, D3 (order, rates equal) | clean for rates |
| Skill damage: Holy Bolt, Zeal, Sacrifice, Vengeance, Blessed Hammer, auras; Frost Nova, Static Field, Blizzard, Glacial Spike | [decomp29-skilldamage.md](decomp29-skilldamage.md) (24,750 element cases 0 differ), [palaura-trace.md](palaura-trace.md), [pal-skill-matrix.md](pal-skill-matrix.md) rows; merged #146, #195, #158, #163 | clean; the open deviations (Zeal gap, sequence -30, Smite) have rows P17..P19 |
| `internal/power/skilldamage.json` estimates | element min / max per level equal game.exe's (decomp29 section 3) | **no invalidation from this card**: T1 / T2 do not touch skill damage; the consumer of absorb is the player as a defender |

## 2. T1: gear absorb (stats 142..149), V read then V run

### Rule (game.exe)

| Q | Answer | Tag |
|---|---|---|
| Q1 Which table applies it? | `FUN_0057c1e0` loops the 12 rows at `0x732980` (stride `0x2c`) and calls `FUN_0057bf80` on each (`0x57c378..0x57c3a3`). Row fields: record offset, resist stat, max-resist stat, pierce stat, **absorb percent stat (+0x10), absorb flat stat (+0x14)**, the flat-reduction index. Fire: offset `0x10`, absorb stats `0x8e` / `0x8f`; lightning `0x1c`, `0x90` / `0x91`; magic `0x20`, `0x92` / `0x93`; cold `0x24`, `0x94` / `0x95`. Physical, poison, chill, freeze rows have `-1`. | V read (the table dumped from the image) |
| Q2 Where is the absorb? | `FUN_0057bf80` order (V read `0x57bf80..0x57c057`): field at or under 0: stored 0, stop; else the flat reduction (stat 34 for physical, stat 35 for the elements, `0x57bfd0`), the resistance (`FUN_0057be00`, `0x57bfdc..0x57bff9`), then **`FUN_0057bf10`** (`0x57c02b`) unless the record ignores resistances (flag in the third argument, `0x57c01e`). | V read |
| Q3 What does `FUN_0057bf10` do? | percent: `p = stat(+0x10)`; `p >= 40 -> 40`, `p <= 0 -> skip`; `a = MulDiv(field, p, 100)` (FUN_00483360), the record's `+0x48` (absorbed) grows by `a`, `field -= a`. Flat: `f = stat(+0x14) << 8`; `f <= 0 -> skip`; `a = min(f, field)` (signed, so a negative field takes itself), `+0x48 += a`, `field -= a`. (`0x57bf14..0x57bf71`) | V read, V run |
| Q4 A field the reduction left under 0? | The absorb still runs: the percent moves it toward 0 (`MulDiv` of a negative), the flat clears it to 0. Example (V run): a magic field of 2 points, flat reduction 3, absorb 36 percent: 512 - 768 = -256, then -256 - MulDiv(-256, 36, 100) = -164; with a flat absorb of 5: 0. | V run |
| Q5 Order against the resistance? | After it: fire 100 points, 50 percent resist, 20 flat: 12,800 - 5,120 = 7,680 (V run); flat before the resistance would be 10,240. | V run |

### What d2d does (V read, upstream `ed007569`)

`rules::monster_blow` took `roll x 256 - mdr x 256`, then `resisted_fine` for fire / lightning / cold, and stopped (`combat.hpp` element loop). `make_fighter` reads no stat 142..149 (grep: no hit), `StruckBlow::Part` has no absorb, so the armors' recompute (`elemental_fine`, Cyclone Armor and Energy Shield) never did either. Any item or charm that carries an absorb property (item_absorbfire_percent and the seven like it) did nothing against a monster's fire, cold, lightning or magic blow.

### The oracle (V run)

`tools/emu/dc37_absorb.py` (fork copy `gear_absorb.py`) runs `FUN_0057c1e0` natively on a player defender in a classic Normal game (no penalty, the cap is 75 + the max stat) with one element field in the record, the defender's resist, flat magic reduction and absorb stats set through the stat-getter hook of upstream's `monster_damage.py` rig; it prints the field and the record total `+0x4c` in 256ths. d2d side: `tools/emu/gear_absorb_harness.cpp` runs the real `monster_blow` (an `exact` record: one El row of min == max, no physical).

| Run | Cases | parent | claim |
|---|---|---|---|
| seed 7 (`gen 20000 7`: type x field x res -100..75 x absorb 0..60 x flat x mdr) | 20,000 | **13,189 differ** (of 16,607 with an absorb; the 3,393 without are equal) | **0** |
| seed 1 (`gen 60000 1`) | 60,000 | **39,678 differ** (50,062 with an absorb; 9,938 without equal) | **0** |
| mutants of the claim, seed 7: no cap at 40 / flat before percent / flat not capped to the field / negative fields skipped / absorb before the resistance | 20,000 each | n/a | 2,271 / 1,556 / 7,983 / 4,883 / 1,826 differ (each is caught) |
| `test_combat` with the same five mutants | pins (section 5) | n/a | each aborts on its own assert |

### Size (I, arithmetic)

A player with 20 percent cold absorb takes 20 percent less of every cold blow in game.exe and none less in d2d; the flat absorbs are 1 to 15 points a hit on rarer pieces. Both are small for the bot today (its plans do not seek absorb, bot finding B3), but it changes the Hell and Nightmare intake of any hero that wears them, and the parent never reduces a negative-field case.

## 3. T2: Open Wounds (stat 135), V read then V run

### Rule (game.exe)

| Q | Answer | Tag |
|---|---|---|
| Q6 Where is the consumer? | ItemStatCost row 135: `itemevent1 domeleedamage` func 15, `itemevent2 domissiledamage` func 15 (Crushing Blow, row 136, is func 16). The handler table is at `0x7325b0` (the dword `0x5bffc0` of Crushing Blow sits at `0x7325f0`, index 16), so entry 15 is **`FUN_005bfe60`** (`0x7325ec`). decomp35 looked in the wrong functions (the apply and the leech); the handler is not reached from them directly. | V read (itemstatcost dump, the table and the pointers read from the image) |
| Q7 What does it do? | Chance `stat 135` of the attacker (`FUN_00625500`, `0x5bfe89`), a draw `rand % 100 < chance` on the attacker's seed (the LCG step of the crush handler, `0x5bfe9e..0x5bfec9`). Then `L = max(1, stat 12)`, `FUN_005bfda0(L, {9, 18, 27, 36, 45})`: `(L-1) x 9` to level 15, then `14 x 9 + (L - 15) x 18` to 30, `+15 x 18` ... `(L - 60) x 45` above 60 (breakpoints 15 / 30 / 45 / 60), **plus 40** (`0x5bff20`). A player as the struck unit divides by 4 (and by 2 more for a missile, event 6); a monster whose data word `+0x16` has bit 4 or 8 (FUN_005a0180 mask 12, `0x5bff3c..0x5bff56`: champion, unique, superunique; not a minion) halves it (toward zero). The negated value goes to `FUN_0056e970` as a timed stat change: stat **0x4a** (life regeneration, 256ths a frame), state **0x3e** (States.txt row 62 `openwounds`), **0xc8 = 200 frames** (`0x5bff8a..0x5bffa2`). | V read; V run (every number below) |
| Q8 Who loses the life? | The monster's per-frame regeneration function `FUN_005a6920` adds stat `0x4a` (256ths) to stat 6 every frame (V read `0x5a6926..0x5a69f9`; not run), with state 0x34 holding a gain back at full life. d2d's `Fight::monster_dots` is the same arithmetic in milliseconds (`regen / (256 x 40)` a ms). | V read; the consumer is not run (I that the two agree) |

### What d2d does (V read)

`Fighter::open_wounds` (stat 135) rolls `rng(100)` in `player_blow`; `Fight::land` set `bleed_rate = open_wounds_per_sec(clvl) / 1000` a ms for 8,000 ms, where `open_wounds_per_sec` is the bracket table "1.10's table" in whole points a second (`value x 25 / 256`, truncated). The bracket is game.exe's (V run, all 99 levels), the truncation and the missing halving are not.

### The oracle (V run)

`tools/emu/dc37_ow.py` (fork copy `open_wounds.py`) calls the handler `FUN_005bfe60` natively with the chance fixed at 100 (the draw is the crush handler's LCG), a level, a player or a monster whose data word is set, and event 5 or 6, and reads the record handed to `FUN_0056e970`. The compared quantity is the life the wound takes over its whole length in 256ths (`value x frames` on game.exe; `open_wounds_per_sec x 8 x 256` on the parent; `open_wounds_fine x kOpenWoundsFrames` on the claim).

| Run | Cases | parent | claim |
|---|---|---|---|
| seed 7 (`gen 200 7`: levels 1..99 x data words 0, 2, 4, 8, 0xa, 0xc, 0x10, 0x12 x events 5, 6, plus 200 random) | 2,000 (244 player rows skipped: d2d has no monster that wounds a player) | **1,748 differ** (plain 859 of 867, special 889 of 889) | **0** |
| seed 1 (`gen 20000 1`) | 21,782 (5,256 player rows skipped) | **16,463 differ** (plain 7,032 of 7,095, special 9,431 of 9,431) | **0** |
| mutants of the claim: no halving / slope 36 above level 60 / 100 frames / base 31 | pins | n/a | each aborts `test_combat` (section 5) |

The data-word bits: bit 2 and bit 0x10 do not halve; bit 4 and bit 8 do (V run, 16 cells at level 0). Which spawner sets which bit is V read in the spawners (`0x5a083e` champion `|= 4`, `FUN_005a0320` mask 8 on every unique, champion and superunique placed, `0x5a4599` / `0x5a4a99` `|= 2` the explicit uniques, `0x5a0cca` / `0x5a47df` `|= 0x10` minions) and the fork's `docs/research/re/skills.md` (S): d2d's `Boss` kinds champion, unique, superunique map to "special" exactly as its stun code already does (I for the mapping, S for the stun code).

### Size (I, arithmetic)

Life taken from a plain monster over the 8 seconds, game.exe against the parent: level 1 31.3 against 24 (+30 percent), level 5 59.4 against 56, level 12 (Paladin early game) 108.6 against 104, level 30 340.6 against 336 (the truncation costs 0 to 1 point a second). The halving matters more: a champion or unique takes half in game.exe, the whole in d2d: a hero with Open Wounds on a weapon (the Paladin's mace, Blade of Ali Baba and the like) kills uniques in d2d faster than in game.exe.

## 4. T3: the player's magic resistance (stat 37), V run, not built

game.exe: row 4 of the table (magic: record `+0x20`, resist stat `0x25`, max stat `0x26`) goes through `FUN_0057be00` like the elements: the player's difficulty penalty (`0x57be5b..0x57be72`: -20 / -50 classic, the difficulty row's value in expansion), then the cap 75 + stat `0x26` (`0x57be8f..0x57bec8`). `Fighter::res` has fire, lightning, cold, poison; `monster_blow` takes the magic field through the flat reduction only (`type != 4`), and the char panel's resist sums have no magic. V run: `dc37_absorb.py game` on 3,000 random magic fields with a resist from -100 to 75 against d2d's `monster_blow` with no input for it: 1,550 differ (all with a non-zero resist). Reach: monsters with a magic El row are the Act 5 succubi and the like (monsterhit-trace D2: 21 rows, none in Acts 1 and 2); Hell doubles the field at the penalty's -100. No branch: the fix is `Fighter::magic_res` plus the char panel's penalty and cap for stat 37, a larger change than a late-act rule is worth today.

## 5. Branches (private fork remote only; nothing posted)

The `-r1` branches of the same names were pushed first (same code; the ruff and doc-claims findings of `presubmit.py --quick` on the oracle commit were fixed afterwards, so the fixed history is the new branch: no force-push). Use `-r2`.

| branch | commits | tree | proof | verifier |
|---|---|---|---|---|
| `fid/gear-absorb-r2` | oracle `00e7b018`, claim `82eced5f` | upstream `ed007569` | section 2: 20,000 and 60,000 cases, parent 13,189 and 39,678 differ, claim 0; 5 mutants caught by the compare and by `test_combat`; `test_combat` ok, `test_game` rc 0 (Release, GCC 14, builder, `-j2`) | not run; body [decomp37-gear-absorb-body.md](../upstream/bodies/decomp37-gear-absorb-body.md) |
| `fid/open-wounds-r2` | oracle `857ece45`, claim `6d4b506d` | upstream `ed007569` | section 3: 2,000 and 21,782 cases, parent 1,748 and 16,463 differ, claim 0; mutants; `test_combat` | not run; body [decomp37-open-wounds-body.md](../upstream/bodies/decomp37-open-wounds-body.md) |

`test_combat` pins (game.exe's values from the oracle): absorb `(fire 100, res 25) 19,200`, with 20 percent `15,360`, `(cold 50, 60 percent, flat 10) 5,120`, `(lightning 10, res 50, flat 100) 0`, `(fire 100, res 50, flat 20) 7,680` (the order), `(magic 2, mdr 3, 36 percent) -164`, flat on it `0`, and the stat to element mapping of `make_fighter`; Open Wounds `open_wounds_fine(1 / 15 / 16 / 30 / 61 / 99) = 40 / 166 / 184 / 436 / 1,426 / 3,136`, half `20 / 218 / 713 / 1,568` at 1 / 30 / 61 / 99. The pins use their own `Rng` or none, so the RNG-luck check of the decomp35 verifier does not apply.

Hygiene (V run, local): commits authored and committed with the project's public noreply identity (not the scratch one), `Agent: Claude Code` the only trailer after a blank line; oracle first (builds alone: the absorb harness skips the absorb columns when the field is absent, the Open Wounds harness prints the parent's rule), claim second; the compares exit 1 on an empty file, unequal lengths, a malformed line or ids out of step. `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN and `presubmit.py --quick --base upstream/main` 0 FAIL on both `-r2` heads (WARNs left: build and gcc skipped by `--quick`; for Open Wounds also siblings, answered in its body). `test_combat` ok and `test_game` rc 0 on both claims (Release, GCC 14, builder, `-j2`, built on the `-r1` commits whose code the `-r2` heads equal). **Not run: the full `ctest` (a 713-target build at `-j2` was started and stopped at 38 steps: hours on the shared builder), clang-tidy, any Clang or Windows build.** The verifier runs them on the worker host.

## 6. Deviation list (d2d against game.exe)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D1 | Gear absorb (stats 142..149) on a player's fire, lightning, cold, magic field, after the resistance: percent (cap 40) then flat. d2d: none | V run + V read | yes, built: `gear_absorb.py`, `fid/gear-absorb-r2` |
| D2 | Open Wounds halved on a champion, unique, superunique. d2d: not halved | V run (game), V read (d2d), I (kind to data word) | yes, built: `open_wounds.py`, `fid/open-wounds-r2` |
| D3 | Open Wounds in 256ths a frame for 200 frames; d2d whole points a second for 8 s (up to a quarter lost at level 1) | V run (value), V read (consumer) | same oracle |
| D4 | A player struck by a monster with Open Wounds: `/ 4`, missile `/ 2` more. d2d has no such monster | V run (game side only) | no d2d side (no monster carries the property in d2d); buildable if MonProp openwounds is ever read |
| D5 | A player's magic resistance (stat 37, penalty, cap) on a magic field. d2d: none | V run + V read | yes, oracle ready (`dc37_absorb.py`, the 3,000 magic cases); fix not built |
| D6 | Absorb on a monster, merc or pet defender and on the burn / blast paths | U | not traced: the same function serves them, d2d's merc / pet fighters carry no gear stats; the burn and blast paths take no gear absorb in d2d (`ai.hpp:1096`, `objects.cpp:134`) |
| D7 | The order of the events (crush, Open Wounds) on one blow: game.exe walks the item's property list; d2d draws crush then bleed | U | would need the item property order of FUN_00626920 run with two items |
| D8 | A second Open Wounds on a wounded monster: replace (refresh) or add | U | `FUN_0056e970` was stubbed, its replace rule not read past `0x56ea53` |

## 7. Bot findings (B1..), described, NOT changed

| B | Finding | file:line | Tag |
|---|---|---|---|
| B1 | The player's resist is capped at 75 BEFORE the difficulty penalty is added: `max(min(v, 75) + p, -100)`. game.exe adds the penalty (0 / -40 / -100, or -20 / -50 classic) first and caps after (`0x57be3d..0x57bf02`; [model-ab-skresist.md](model-ab-skresist.md) Q10, V run): a raw 90 is 50 in Nightmare in game.exe and 35 in the bot; a raw 120 is 20 in Hell and -25 in the bot. The cap is `75 + stat 0x26`, not a constant. The readiness gate and `ResistDanger` read pessimistic whenever the raw sum is over 75 | `internal/combat/rules.go:311`, `internal/readiness/model.go:357`, test `internal/readiness/readiness_test.go:217` | V read + S (model-ab-skresist) |
| B2 | The bot's resist model has no magic resist (stat 37 is read into `Resists` at `internal/d2/player.go:184` but the readiness elements are fire, cold, lightning, poison): a Hell Act 5 magic hit is not rated | `internal/readiness/areatable.go:24`, `internal/readiness/model.go:344` | I |
| B3 | No gear scoring or plan term for absorb stats, Open Wounds or leech (grep of `internal/` for the stat names: none outside generic stat tables), so absorb gear is never preferred, and the survival model has no absorb term | `internal/` (grep) | I |
| B4 | `internal/power/skilldamage.json` is not invalidated: its element ranges equal game.exe's (decomp29), and neither T1 nor T2 changes a skill's damage. The one effect on estimates is on the DEFENDER side (absorb): none of the power estimates models the player's absorb | `internal/power/skilldamage.json` | I |

## 8. Dead ends, not built, and open items

- **D5 of decomp35 (a monster's MonProp Crushing Blow on the player, the merc, the pets)**: the order problem remains. d2d aggregates a frame's crushes after all its hits (`crushed` list, `tick.cpp:428, 474, 647, 773`) while game.exe raises the strike event BEFORE each blow's damage (V run, decomp35 and the verifier's pair oracle). A faithful fix needs the per-blow order in `Foe`; the oracle is `crush.py` of `fid/crush-r2` (V run: the handler is attacker-agnostic), which is not yet merged: build after the leech and crush PRs merge. Not built.
- **D6 of decomp35 (gear element roll in 256ths)**: still 2,223 of 3,947 differ, mean +1 percent; the fix touches 15 users of `Fighter::elem` and the element sum, the chill mark and the poison split. Not built (size against effect).
- **Evade / avoid / dodge (stats 338..340)** were read again against [monsterhit-trace.md](monsterhit-trace.md) Q4 and D3 and left: Amazon passives, no build of ours uses them.
- **First attempt at the Open Wounds consumer** looked for stat 135 readers by `push 0x87` immediates (a scan of the code for every use of the number) and found none: the stat is read through the event packet (`[ebp+0x14] >> 16`) by the generic handler, not by a literal. The table pointer scan (`0x5bffc0` at `0x7325f0`) is what worked.
- **The absorb stats are read by a literal scan** (`push 0x8e..0x95` before a call) found nothing for the same reason: they sit in the row table at `0x732980` and are read through the row. Recorded so nobody repeats the scan.
- **Tooling note:** a `pkill` was typed once inside the builder container to stop my own `ninja` (it was ineffective, no `pkill` binary there); the build was then stopped with `kill` on the three pids of my own job. Nothing of another agent was touched.

## 9. Drill recipe (fork verbs of [../forkdbg.md](../forkdbg.md) and [../forkcheats.md](../forkcheats.md); nothing was run, no game)

Throwaway character, hidden slot, lab binary of [../d2d-fork.md](../d2d-fork.md), setup writes declared before the run start, ledger entry per write; a character with setup-written stats shows the mechanism only (provenance rule).
- **R1 absorb (D1):** `debug stat 144 25` (25 percent lightning absorb; confirm with `dump stats`), `debug spawn <a lightning-hitting monster> 1 near` (a Rakanishu-style unique with a lightning El row, or any `debug mon set <id> dmg <lo> <hi>` plus `res`), `debug hit` is not enough (it is not a monster blow): step until the monster lands a hit and read the life lost: expected the lightning part x 0.75 less your lightning resist; the parent takes it whole. Repeat with `debug stat 145 20` (flat 20) on a 12-point lightning hit: expected 0 from the lightning part.
- **R2 absorb cap:** `debug stat 144 60`: the same hit takes 40 percent, not 60.
- **R3 Open Wounds (D2, D3):** `debug stat 135 100`, a Normal monster with `debug mon set <id> life 100000`, attack once; read the life every `step 5` for 200 frames: the loss per frame is `(bracket + 40) / 256` (level 1: 0.156 a frame, 31 in all); the same on `debug spawn <monster> 1 near champion` (`champion` argument): half. On the parent the champion loses the full amount and the plain one 24 over 8 s.
- **R4 magic resist (D5):** `debug stat 37 30` and a magic-damage monster (Act 5, `debug spawn succubus1`): expected 30 percent less in game.exe, none in d2d.
- Each U (D6, D7, D8) needs more than the verbs: D8 `FUN_0056e970` run natively with a state already on the unit.

## 10. Matrix-cell updates (proposed text; the owners apply them)

- **[fidelity-matrix.md](fidelity-matrix.md) ST03 (resists):** append "gear absorb (stats 142..149): DEV V run, `fid/gear-absorb-r2` (60,000 cases, 39,678 differ on the parent); the player's magic resist on a magic blow: DEV V run, not built (decomp37-combat.md section 4)."
- **IT / SK melee riders** (decomp35 text): replace "open wounds U" by "Open Wounds: handler `FUN_005bfe60` (event func 15), bracket table equal, not halved on a champion / unique and cut to whole points a second in d2d: DEV V run, `fid/open-wounds-r2`".
- **[pal-skill-matrix.md](pal-skill-matrix.md) / [sorc-skill-matrix.md](sorc-skill-matrix.md) "Not covered":** "absorb and Open Wounds traced in decomp37-combat.md".
- **Pipeline** (docs/upstream/pipeline-2026-11-14.md, section 1b): rows P23 (gear absorb, fork `fid/gear-absorb-r2`, stage S2, verifier next) and P24 (Open Wounds, `fid/open-wounds-r2`, stage S2); a row for T3 at stage S0 (oracle ready, fix not built).
- **monsterhit-trace.md D2** (the magic element of a monster swing, U): "V run in decomp37: 1,550 of 3,000 magic cases differ (the player's stat 37, penalty and cap)".

## 11. Reproduce (builder container, emulator image at `/data/prfix6-emu`)

```
python3 tools/emu/dc37_absorb.py gen 20000 7 > c.tsv                  # cases only, any box
<emu env> python3 dc37_absorb.py game c.tsv > g.tsv                    # game.exe, in the container
g++ -O1 -std=c++23 -I<d2d>/components/rules -I<d2d>/components -I<d2d>/components/d2s -I<d2d>/components/txt \
    -o gear_absorb_harness gear_absorb_harness.cpp && ./gear_absorb_harness < c.tsv > d.tsv
python3 dc37_absorb.py compare c.tsv g.tsv d.tsv                       # parent: 13,189 differ; claim: 0, exit 0
```
Open Wounds the same with `dc37_ow.py` and `open_wounds_harness.cpp` (`gen 20000 1`: 21,782 cases, parent 16,463, claim 0). The harness sources are `tools/emu/dc37_absorb_harness.cpp` and `dc37_ow_harness.cpp` here (the claim versions; the scripts import upstream's `monster_damage.py` rig, so they run from upstream's `tools/emu` on the builder).

## Shared-file lines for the merger

Generated from the task card (TASK-607); the pipeline rows and matrix cells of section 10 are for their owners.
