# decomp41: merc absorb, crush follow-ups D5 / D6, the held candidates (2026-11-18)

> status: current, 2026-11-18 (agent decomp41, Sonnet; docs, `tools/emu` oracles and branches in the private fork, nothing posted, no live game). Subject: the next upstream candidates after #606 (gear absorb), #609 (crush) and #615 (merc regen), all merged. The independent Opus verifier (singly) comes before anything is posted. Bodies: `docs/upstream/bodies/merc-absorb-body.md`, `crush-d5-body.md`.

Tags: **V run** (game.exe's own code, or d2d's real code, run and compared), **V read** (disassembly or source read), **D** (data), **S** (an earlier doc), **I** (inferred), **U** (unverified).

## 0. Answers in short

| Candidate | Status | Branch (private fork remote) | Proof |
|---|---|---|---|
| 1. Merc gear absorb and heal-back | built | `fid/merc-absorb-r1`: oracle `4909c7e2`, claim `cc0e92c3` | section 1 |
| 2a. Crush D5: a plain unique divides by 4, not 8 | built | `fid/crush-d5-r1`: oracle `f3e764f3`, claim `b0e6297a` | section 2 |
| 2b. Crush D6: the get-hit size test does not count the crush | scoped, not built | none | section 3 |
| 3a. death-corpse-exp | stays held, re-checked | `fid/death-corpse-exp-r1` (unchanged) | section 4 |
| 3b. Player magic resistance (decomp37 T3) | scoped, not built | none | section 5 |
| 3c. Next best by player impact | list | none | section 6 |

Upstream at this run: `psi29a/d2decomped` main `9b64c11a` (V read, `git fetch upstream`). Our merged PRs include #587, #599, #600, #603 to #606, #609 and #615 (`gh pr list --repo psi29a/d2decomped --author shatteredeagle`, V read: all MERGED; none open). The crush base is therefore stable (the brief's condition for D5 and D6 is met). Both branches are single new branches on `9b64c11a`, two commits each (oracle first), author and committer the public noreply identity, `Agent: Claude Code` the only trailer after a blank line.

## 1. Merc gear absorb and heal-back (`fid/merc-absorb-r1`)

### Questions

- **Q1. Does game.exe's absorb and its heal-back test the struck unit's type?** No. V read: `FUN_0057c6c0` (`0x57c6c0..0x57c844`) tests the unit with `FUN_00554200` and `FUN_00620bb0`, and for a type-1 unit one MonStats flag of its class row (`FUN_00457490(class, 0xf)`, `0x57c76a..0x57c777`: a zero ends the call); nothing names a player. `FUN_0057a980` (`0x57a980..`): amount over 0, the unit not dead (`FUN_005541b0`) and not in state `0x5c` (`FUN_00639df0(unit, 0x5c)`), then `life = min(life + amount, max life)` (`FUN_00625d10`, stat 7). V run (`merc_absorb.py game` against `gear_absorb.py game` on the same 300 cases, the hireling's real MonStats rows loaded with `drlg.boot`, the MonStats readers `FUN_00457490`, `FUN_0063e9f0`, `FUN_0063ee90`, `FUN_0063e990`, `FUN_0063e940` running for real): the merc's life after every blow equals the player's in 300 of 300.
- **Q2. Does a hireling's equipment put its absorb stats on the hireling?** I. d2d's `merc_gear` sums every property of the worn items into the stats `merc_fighter` reads (resistances, defense, life, to-block already come that way, V read `merc.cpp`, `fight/merc.cpp:35..`). game.exe has one stat-list path for a unit with an inventory; the equip path itself was not run, so the proof starts from the stat list. U for the equip path (drill recipe R1).
- **Q3. Does d2d's merc take fractional life, and can it be healed?** V read `ai.hpp:506..512`: `Monster::life_fine()` is `hit_points x 256 - life_frac`, `set_life_fine` its inverse; `Fight::hurt_side` already carries the merc's 256ths (`life_frac`, the dead-under-one-point rule at `tick.cpp`, `hurt`'s `0057c86c` rule). The heal needed only `merc_max_life() << 8` as the cap. V run: the claim equals game.exe's merc life in 18,000 of 18,000 cases (section 1.2).
- **Q4. Which d2d code zeroed it?** V read: `fight/merc.cpp` set `fighter.absorb_pct = {}; fighter.absorb_flat = {};` with a comment that the absorb without its heal-back would be half the rule; `test_gear_absorb.cpp` asserted it (the `Players only` block, removed in the claim).

### 1.1 Differential

`merc_absorb.py gen N SEED` writes cases (life in 256ths, max life, magic reduction, resistances -100..75, four absorb percents (cap 40 applies), four flats, the hireling class 271 / 338 / 359 / 560 / 561, 1 to 3 blows of 1 to 3 fields). `game` runs `FUN_0057c6c0` per blow on a type-1 unit; `test_merc_absorb --life` runs d2d's real `merc_foe` (the fighter from worn items), `monster_blow`, `Foe::take` and `Fight::hurt_side` on a merc whose torso piece carries the case's stats; `compare` exits 1 on any difference, an empty, short or malformed file (V run, edge cases below).

| | seed 7 | seed 5001 | seed 9091 |
|---|---|---|---|
| parent (oracle commit `4909c7e2`), 6,000 cases | 3,075 differ | 3,056 | 3,057 |
| claim `cc0e92c3` | 0 | 0 | 0 |

Edge exits (V run): an empty game file, an empty d2d file, an empty case file, a short d2d file, a malformed d2d line, a missing file, no argument, `game` on an empty case file: every one exits 1; a d2d file equal to the game file exits 0.

### 1.2 Mutants of the claim (V run; each one text change, built on the builder, `test_merc_absorb` and the compare on seed 7)

MERC_MUTANTS

### 1.3 Build and tests

MERC_BUILD

### 1.4 Deviations (merc, d2d against game.exe)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D1 | game.exe: a hireling's worn absorb lowers an elemental or magic field and heals it back. d2d (parent): zeroed. Fixed by the candidate | V run | yes: `merc_absorb.py` + `test_merc_absorb` (built) |
| D2 | Pets and summons: carry no gear in d2d, so no stat 142 to 149; a pet's absorb from an aura or a curse stat is not read by `Fight`'s pet fighter (U whether any aura or state gives a unit absorb stats) | U | none; run `merc_absorb.py game` with a pet's class and stats if a state with absorb stats exists |
| D3 | The burn and blast paths (a field that bypasses `monster_blow`, `ai.hpp` `burn`, `objects.cpp`) take no absorb in d2d for the player either (#606 limit); the merc is the same | U | none |

### 1.5 Bot findings (described, not changed)

| B | Finding | Where | Tag |
|---|---|---|---|
| B1 | `merc.Score` values defense, life, damage, sockets and Hell resists; no absorb (stats 142 to 149). `merc.Item` has no absorb field. With the candidate a merc's absorb gear heals it in d2d; the bot makes no merc equip decision today (no caller), so nothing changes until the doll lands (docs/research/merc-gear.md) | `internal/merc/gear.go:11-20`, `internal/merc/types.go:60-75` | V read |
| B2 | The comment on `d2.Absorb` says "d2d reads none" and the player's stat list may not expose 142..149 (U). #606 made d2d read them for the player; the candidate for the merc | `internal/d2/player.go:47-50` | V read |
| B3 | The readiness model counts absorb for the player only | `internal/readiness/model.go:42-45`, `internal/bot/readinessfacts.go:32` | V read |

### 1.6 Drill recipe (fork verbs, lab binary; run later, not now)

R1 (equip path, answers Q2): `debug merc hire 12` (a Desert Guard), then put an item with an absorb property on the merc (verb for an item with chosen properties: U; `debug give <code>` gives a base item, `docs/forkcheats.md`; if no verb sets properties, load a `.d2s` with a merc item that carries cold absorb as a SETUP, logged in `docs/probes/test-characters.md`), `dump merc` for the life, `debug spawn` a cold-damage monster (a Frost-ball caster or `debug mon set <id> dmg <n> <n>` on a melee Fallen with a cold property), `debug merc hurt <n>` for the same blow without a monster, and read `dump merc` life before and after. Expect: merc life after a cold blow of 100 at 90 of 100 with 20 percent absorb = 20; with no absorb the merc dies. R2: the same blow with the stat list read back (`dump merc` shows life only: a stat dump of the merc is U).

### 1.7 Matrix-cell updates (proposed text)

- `docs/research/decomp37-combat.md` section 6, D6 row: "Absorb on a monster, merc or pet defender and on the burn / blast paths | **merc: V run (decomp41), built `fid/merc-absorb-r1`**; monster and pet defenders carry no worn gear (U for states); burn and blast U".
- `docs/research/re/combat.md` (upstream, edited in the claim): the merc sentence replaced.
- `docs/upstream/pr-queue.md` (shared; the merger): add a row "merc absorb, `fid/merc-absorb-r1`, ready for the verifier".

## 2. Crush D5: the divisor of a plain unique (`fid/crush-d5-r1`)

### Questions

- **Q5. Which monsters does game.exe give the divisor 8?** V read `FUN_005bffc0` `0x5c0031..0x5c00a8`: 4; a hireling 10 (`FUN_0063ee90`: classes 0x10f, 0x152, 0x167, 0x230, 0x231); else 8 when `FUN_0063e9f0(0, unit)` (a MonStats boss row, row byte `+0xc` and mask `[0x6ce280]`) or `FUN_005a0180(unit, 2)` (data word `+0x16`, bit 2); plus `MulDiv(divisor, FUN_005738f0(max(stat 0x64, 1)), 100)` (a table at `0x6e1590`, then `(n - 2) x 50`).
- **Q6. Who writes bit 2 of the data word `+0x16`?** V read, a scan of every `or`, `and`, `mov` of a word or byte at `[reg + 0x16]` in `.text` (46 hits, of which the monster ones are `0x5a01c9`, `0x5a0375`, `0x5a083e`, `0x5a0b76`, `0x5a0cca`, `0x5a10c5`, `0x5a1258`, `0x5a44a3`, `0x5a4599`, `0x5a4905`, `0x5a4967`, `0x5a4a99`): bit 2 is written only by `or word ptr [eax + 0x16], 2` at `0x5a4599` and `0x5a4a99`, the two superunique branches (the second sits beside the SuperUniques spawned-bit table at `+0x1d30`, `0x5a4a76..0x5a4a89`). A unique (`FUN_005a4940`: `or ..., cx` with `cx` = the unit type 1 at `0x5a4967`, then `FUN_005a0320`'s bit 8 at `0x5a0375`) carries 9; a champion (`FUN_005a48c0`: `or ..., 5` at `0x5a4905`, then bit 8) 0xd. `FUN_005a01b0`, the setter with a mask, has no caller in `.text` (a scan of every `e8` call, V read). Upstream's `drops.md` had this as "possible disagreement, not established: the writers were not found"; settled.
- **Q7. What does the whole divisor look like, run?** V run (`crush_div.py game`): `FUN_005bffc0` natively on a type-1 unit of every MonStats row (734) made plain, champion, unique, superunique, or minion, at 1 to 8 players (29,360 cases, every table loaded, `FUN_0063ee90`, `FUN_0063e9f0`, `FUN_005a0180`, `FUN_005738f0` real). Totals of game.exe's divisor by kind at one player: plain, champion, unique, minion 4 on 704 rows, 8 on 25 (the MonStats boss rows), 10 on 5 (the hirelings); a superunique 8 on 729 rows, 10 on 5.

### 2.1 Differential

D5_RESULTS

### 2.2 Mutants of the claim

D5_MUTANTS

### 2.3 Deviations (crush on a monster defender)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D5 | game.exe: 8 for a MonStats boss row or a superunique; a unique and a champion 4. d2d (parent): 8 for a unique too. Fixed by the candidate | V run (game), V read (writers) | yes: `crush_div.py` + `test_crush_divisor` (built) |
| D6 | see section 3 | V read both | buildable, not built |
| D7 | The monster-defender MonProp crush on the player, merc and pets is applied after the hits and reads the life they left (`ai.hpp:683`, `tick.cpp:634/762`); game.exe reads it before the hits (the same handler, event 5) | V read d2d, game side S (`crush.py order`) | partly: `crush_div.py game` gives the handler's crush; the order needs a stateful merc / pet run |

### 2.4 Bot findings

| B | Finding | Where | Tag |
|---|---|---|---|
| B4 | `party/scaling.go` says "Crushing Blow is divided by 0.5 + 0.5 n": equal to game.exe's divisor share for 2 and 3 players (+50 and +100 percent, V run `crush_div.py`, the table at `0x6e1590`); the sentence does not say that bosses divide by 8 and that a plain unique, like a champion, divides by 4 | `internal/party/scaling.go:29` | V read |

## 3. Crush D6: the get-hit size test counts the crush in d2d (scoped, not built)

- **Rule on game.exe's side.** V read (S: `verifier-crush3-2026-11-17.md`): `FUN_0057cb00` (`0x57cb22..0x57cb35`) tests the record's total `+0x4c`, which the crush never enters: the handler writes the defender's stat 6 directly and the record's physical is cut to the life the crush left (`FUN_0057c6c0` `0x57c7f8..0x57c826`). `hit_rule.py` (`FUN_0057cee0` -> `FUN_0057cb00`) is the existing game-side oracle for that decision (V run upstream, 100,000 cases).
- **Rule on d2d's side.** V read `fight/strike.cpp:548..549`: `taken = soak(target, blow.phys_fine, (blow.damage << 8) + blow.frac)` and `hurt(..., taken >> 8, ..., taken & 255, blow.hit_class, ...)`; `hurt` (`ai.cpp:2926..2943`) uses `blow_fine` (the whole damage, which includes `player_blow`'s crush) as the size of the blow for `rules::get_hit`. So a Crushing Blow proc on a monster raises its get-hit chance by the crush share: a monster staggers more often under Crushing Blow than in game.exe.
- **Provable?** Yes in principle, not cheaply: a differential needs `hit_rule.py` cases built from player blows (life, crush, physical), the d2d side through `Fight::land` -> `hurt` with the monster's seed, and the claim a `crush_fine` carried in `Blow` and subtracted for the size test only (a signature change on `hurt` or on `Blow`, two call sites in `strike.cpp`, one in `pets.cpp`). The seeds must agree on how many rolls a blow consumes (the existing get-hit oracle does cover that). Estimated one full agent, M; the effect is the share of Crushing Blow procs that stagger a monster the game would not (a visible but small effect on kill speed). Not built here: the budget went to D5, the merc, and the held list; no PR candidate.

## 4. death-corpse-exp (`fid/death-corpse-exp-r1`): why it is held (re-check)

- **What it is.** The experience the corpse is told it lost is one low when the share reaches the level's start: `FUN_005359f0` (`0x535a63..0x535a6f`) tells the corpse `experience - start` and sets the experience to `start + 1`; d2d told `experience - (start + 1)`. V run (decomp36, two seeds of 120,000 cases): 31,792 and 31,750 differ on the parent, 0 with the claim.
- **Why it was held as trivial.** The corpse returns 75 percent of what it is told, so the visible effect is at most one point of experience (and 0.75 of a point back), and only for a Nightmare or Hell death within 5 or 10 percent of a level's span after the level's start (Normal loses nothing); the grid over-samples that band. A real run of play loses 0.75 of an experience point once in a long time.
- **The second half (D2 of decomp36).** `FUN_0057f700` (`0x57f81c..0x57f86c`) computes `(told / 100) x 75` above 1,048,576 and `told x 75 / 100` below; d2d always `told x 75 / 100`. V read only. It matters from clvl 60 in Hell (a share above 1,048,576 points) and the game would give back less than d2d there; no oracle (the function is a whole-corpse routine).
- **Re-check today.** The branch sits on `b6e6940d` and was never verified; the rule is in `death_penalty` unchanged on `9b64c11a` (V read, `git grep` of the clamp). Recommendation unchanged: hold it; post it only with a PR that also settles the 75 percent step (needs an oracle that runs `FUN_0057f700` with the corpse inventory hooked), or when a Nightmare / Hell play log shows the case. It is not a candidate for the next batch.

## 5. Player magic resistance (decomp37 T3): scope

- **The rule.** V run (decomp37, `dc37_absorb.py`, 3,000 random magic fields: 1,550 differ, every one with a resist other than 0): a monster's magic field goes through the player's magic resistance (stat 37: row 4 of the table at `0x732980`, `FUN_0057be00`: the difficulty penalty -20 / -50 classic or the expansion's row, the cap 75 + stat 38). d2d: `Fighter::res` has four entries (`combat.hpp:78`), `monster_blow` skips type 4 (`combat.hpp:1139,1142`: `type != 4 ? defender.res[type] : 0`), the character panel sums no magic resistance.
- **What is needed.** (1) `Fighter::res` to five entries (magic at index 4, which `monster_blow` already indexes by `type`), `make_fighter` taking it. (2) The three callers that build the player's resistances: `Fight::update_fighters` (player), `merc_fighter` (`fight/merc.cpp`, with its penalty and cap), the pets' fighters. (3) `monster_blow` using `defender.res[4]` and its cap/penalty exactly as the elements (the penalty row must be the expansion's for magic too). (4) The oracle: add a magic-resistance column to `gear_absorb.py` (it hard-sets 0 at `(4, 0, p_magic, f_magic)`), plus d2d's cases. (5) `test_combat` pins.
- **Reach.** Few monsters deal magic damage (the 21 MonStats rows with a magic El row are Act 5 and later, `monsterhit-trace` D2); few items carry stat 37 (a handful of uniques and charms). Value: low for a normal Acts 1 to 3 playthrough, moderate for a Hell Act 5 melee build. Provable with the existing oracle. Effort M (about 400 lines). Not built (the brief asked for scope only).

## 6. The three most player-visible remaining combat or skill deviations (scan)

Source: the open ranks of `docs/upstream/pipeline-2026-11-14.md` section 1b (P06 and P07 were re-read against upstream: P06 monster-missile block is posted as #553), `docs/upstream/pr-queue.md`, `docs/probes/fid/fid-monster-block.md`, `covstates2-trace`. State of each, read today, not re-run:

1. **P08 timed-state put rule (curses, Taunt, Inner Sight, Attract; poison replacement; the stun draw).** V run (covstates2): 54,685 of 100,000 sequences differ. Player-visible: how long and how often a curse (Amplify Damage, Decrepify, Lower Resist), Taunt or Battle Cry holds, and a stronger one replacing a weaker. Reach: every curser, Barbarian shouts, Paladin auras on monsters. A missing-rule change (the put rule is a different function than d2d's per-state timers): needs a new differential on the fix, one rule at a time. This is the highest impact, not provable in one afternoon.
2. **P07 monster block gate.** V run game side: game.exe lets a monster block only when its row has the bit-4 flag (Mephisto, Diablo, Baal's crab and clone) or it is drawn with a shield component (skeletons always; Fallen and corrupt rogues on some spawns); d2d lets every MonStats2 `mBL = 1` row block with its ToBlock (a Fallen blocks about 9 percent of landed blows in d2d, about 1.5 percent in game.exe at Normal). Player-visible in Act 1 Normal for every melee build (a blocked blow wastes a swing, the block animation plays). Blocked by D3 of that trace: the shield is a per-spawn choice (`FUN_00664860`), d2d needs a variant pick at spawn first. Not a one-rule fix.
3. **P10 / P11 friendly auras on the merc and pets, pierce rows.** The merc's friendly auras (Might, Defiance on the merc: stat V run, effect U) and the arrow, javelin and Blessed Hammer pierce rows (S, no oracle) are the next ones for Amazon and Paladin builds; both need an oracle before a claim.

Result: no further candidate is provable by the existing differentials in this run. The next one to build is P08 (rank 1) with a stateful oracle, or T3 (section 5) as the provable but low-reach filler.

## 7. Bot findings summary

B1, B2, B3 (section 1.5), B4 (section 2.4). None changed.

## 8. Drill recipes for the other Us

- R3 (D5 in play): `debug spawn zombie1 1`, make it a unique (`debug mon set`? there is no verb that sets the boss kind: U; a fixed-seed lockstep spawn that comes up a rare monster in Act 1, then `dump monsters`), `debug mon set <id> life <n>`, a Crushing Blow weapon (U: `debug give` with the property), one melee swing, read the monster's life: a unique takes `life / 4`, a boss `life / 8`.
- R4 (D6): the same with the monster's mode read after the swing (`monsters`: GH or not) over several seeds, comparing a Crushing Blow weapon and the same weapon without it.

## 9. Not done

No live game, no lab drill, no Windows or macOS build, the equip path of a merc item (Q2) not run, the dispatch of Crushing Blow's event for a missile not run; D6 and T3 scoped only; the verifier has not read any of this.

## 10. Shared-file lines for the merger

See `backlog/merge-lines/agent-decomp41.txt`.
