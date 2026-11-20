# decomp43: the monster block gate (P07), the crush follow-ups, the next deviations (2026-11-19)

> status: current, 2026-11-19 (agent decomp43, Sonnet; docs, `tools/emu` scripts with prefix `mb_` and a branch in the private fork; nothing posted, no live game). Follows [decomp42-crush-d6-trace.md](decomp42-crush-d6-trace.md) (its scan item 1) and [../probes/fid/fid-monster-block.md](../probes/fid/fid-monster-block.md) (decomp11: the game side and the data side; the shield variant was the blocker). The independent Opus verifier (singly, scheduled by the coordinator) comes before anything is posted. Body: `docs/upstream/bodies/monster-block-body.md`.

Tags: **V run** (game.exe's own code run and compared), **V read** (disassembly or source read), **D** (data), **S** (an earlier doc or an upstream proof), **I** (inferred), **U** (unverified).

## 0. Answers in short

| Candidate | Status | Branch (private fork) | Proof |
|---|---|---|---|
| 1. P07 the monster block gate | **built**: the per-spawn shield IS derivable (section 1) | `fid/monster-block-r1` (oracle `e5425263`, claim `e0236ed1`) on upstream main `8ae5cfcb` | 228 of 2,469 cases differ on the parent, 0 on the claim; mutants in 1.5 |
| 2. The crush rule for merc / pet / missile blows | **missile: already in the verified harness** (r3 of #637 carries the `missile` and `fire` columns); **merc and pet: not provable cheaply**, negligible reach (section 2) | none | |
| 3. P09 raised-monster mode | not built (negligible, as decomp42) | none | |
| 4. The next three player-visible deviations | scan, section 3 | none | |

Upstream at this run: `psi29a/d2decomped` main `8ae5cfcb` (V read, `git fetch`; #632 and #633 are in; about 1,500 commits since 2026-11-12). The branch is two commits on it; author and committer the public noreply identity; `Agent: Claude Code` the only trailer after a blank line (V run, `git interpret-trailers`); `pr_guard.py upstream-gh/main HEAD`: 0 FAIL, 0 WARN (V run).

## 1. P07: a monster blocks only with NoShldBlock or a shield on its look

### Questions

- **Q1. What are the two gates of a monster's block chance?** V read, V run. `FUN_00622720` (`0x622720..0x622823`), unit type 1: gate A is the byte `+0xf` of the MonStats record (table `[0x744304]+0xa78`, stride `0x1a8`) `&` the dword at `0x6ce270` (4 when read in a booted game: the byte is a static constant); non-zero opens it (`0x62276e`). Gate B is `FUN_006225f0(unit)` (`0x622778`). Open: `FUN_00625480(unit, 0x14, 0)`, stat 20; `cmp ebx, 0x4b` caps at 75 (`0x62280e..0x622823`); closed: 0. `mBL` is read by neither.
- **Q2. What is gate A in the tables?** V run. `mb_gate.py`: bit 26 of the dword at `+0xc` of the compiled MonStats row is the gate, and it equals the MonStats.txt column **`NoShldBlock`** on all 734 rows (the one 0/1 column that agrees; `primeevil` disagrees on 6). The 18 rows at the mask: the Unravelers 1 to 5, Andariel, Duriel, Mephisto, Diablo, Doom Knight 1, Diablo's clone, the Baal crab and clone, the Ubers (Mephisto, Diablo, Andariel, Duriel, Baal). Andariel and Duriel have ToBlock 0 at Normal, so the gate opens on a chance of 0.
- **Q3. What is gate B?** V read, V run. `FUN_006225f0` (`0x6225f0..0x622696`): for a class in the hireling range (`class - 0xf3` under `0x75`) a jump table (not read: U, no Normal row there); else idx = monster data `+0xb` (byte), `FUN_00664860(unit, 7, idx)` (layer 7); the returned component code is refused when it is in the one-entry list at `0x744580` (value `tch`, the torch); else `FUN_00633640(code)` finds the item record (hash `0x96bcc4`, stride `0x1a8`, table `0x96ca5c`) and its word `+0x11e` must be 2 (`0x62267e`). `mb_explore.py` read the record words of all 115 component codes: 2 for `buc`, `lrg`, `kit`, `sml`; 21 for the torch; 25 to 44 for weapons.
- **Q4. What is word `+0x11e` of an item record, and is the code test derivable from the tables?** V run, D. It equals ItemTypes' row of the item's `type` (row 2 is `shie`, row 51 `shld`; weapons 25 to 44). `monster_block.py codes` runs it over **all 659 item codes** of armor, weapons and misc: 24 shields in game.exe and **0 codes where `type == shie` disagrees**. d2d's rule is that equality (`GameData::shield_code`, over `item_pieces`).
- **Q5. What is monster data `+0xb`, and what does `FUN_00664860` return?** V read. `FUN_00664860(unit, layer, idx)` (`0x664860..0x66492d`): the unit must be a monster (`[unit] == 1`), its MonStats2 row (`FUN_00451fe0(row +0x18)`, stride `0x134`); the count byte `[row + layer + 0x15]` must be above idx; the variant byte `[row + layer*12 + idx + 0x26]` goes through `FUN_006117d0` (the component-code table at `0x96bcac`) to the code. `+0xb` = monster data `+4` + 7: the look's byte for layer 7 (I: the look bytes sit at `+4..+0x13` in layer order; `FUN_005739d0` writes them). In an Act 2 level (`FUN_006427f0` of the level id is 1) a monster whose MonStats BaseId (`+2`) is 0 takes the code from the table at `0x7489a8` (one pointer per layer) and one with BaseId 0xaa from `0x748a18`: layer 7 of the first lists `'', buc, lrg, kit, sml`, the second has no layer 7 (V run, `mb_act2.py`): the same four shields as MonStats2's `SHv` for the skeleton rows (BaseId 0: skeleton1 to 5, 648, 649, 725), so the layer-7 outcome does not change in Act 2 (I; a unit in an Act 2 level could not be run, the level bring-up faults in the emulator: U).
- **Q6. Is the per-spawn shield derivable in d2d?** **Yes.** `spawn_monsters` already rolls each spawn's look from the unit seed (`monster_look`, `ai.cpp:210..215`: the region's component set or a roll per layer) and puts layer 7's pick in `monster.npc.comp[7]` (`type_info.parts[7][look[7]]`, the MonStats2 `SHv` entry), the same index game.exe keeps in `+0xb`. So d2d needs no new variant pick: the rule reads `comp[7]`. That d2d's pick equals game.exe's per spawn is the population port's claim (S: commit `43a2bcc5`, "bit-exact" per level; not re-run here: U for this link). decomp11's 24-spawn sample (Fallen 1: the shield in 4 of 24 spawns) is the game's region-limited look sets, not a uniform draw over the four entries.
- **Q7. What does the machine do for every row, difficulty and entry (V run)?** `mb_block.py game` (shipped as `monster_block.py`): boots a game of each difficulty, spawns each MonStats row with `FUN_00555230`, writes the entry into `+0xb` and calls `FUN_00622720` natively: 2,469 cases (all 734 rows x 3 difficulties x their list entries; 93 cases are 31 placeholder rows the spawner refuses: Wolf Riders, Suckernests, Dark Guards, Blood Mages, Chaos Hordes, Evil Holes, Familiar, Iceglobe and others, skipped and counted; no shield list, no NoShldBlock there: D). 300 cases block in game.exe. The cap: `monster_block.py cap` sets Mephisto's stat 20 to 0, 1, 74, 75, 76, 90, 200 with `FUN_00627260`: 0, 1, 74, 75, 75, 75, 75 (V run; no MonStats row has a ToBlock above 75, so the cases never reach it).

### 1.1 What the parent does (V run)

d2d (upstream `8ae5cfcb`, `ai.hpp` `Monster::target`): `type_info.can_block ? to_block : 0`, `can_block` = MonStats2 `mBL`. `test_monster_block --monster-block` prints it through the real `make_monster` and `target()`.

| | cases | differ | of which |
|---|---|---|---|
| parent (oracle commit `e5425263`) | 2,469 | **228**, exit 1 | 195 d2d blocks where game.exe does not; 33 game.exe blocks where d2d does not |
| claim (`e0236ed1`) | 2,469 | **0**, exit 0 | |

By name (Normal): d2d blocked and game.exe does not, 57 rows: Fallen 1 to 8 (variants `nil` and the torch; ToBlock 9 to 49), Corrupt Rogues 1 to 6 (12 to 40, no shield), Skeletons 1 to 8 (10 to 30, the `nil` entry), the Necro-Skeleton, **all the Skeleton Mage rows** (cold, fire, lightning, poison, 9 to 17: their list is empty, game.exe never blocks), the Shadow Warrior and Master (50), the death-knight figures, the eagle; game.exe blocked and d2d did not at Normal: Mephisto 20 (Nightmare 40, Hell 50), the Baal crab and clone 40, the Ubers; from Nightmare Andariel (20), Duriel (25), the Unravelers (15). The compare exits 1 on a mismatch (V run: the parent's exit 1, the claim's exit 0) and, by the code (V read, not run: `mb_edge.sh` is written but the builder was out of memory, section 4), on an empty, short or malformed file, a case or id mismatch, and when more than 5 percent of the cases cannot be spawned.

### 1.2 The claim (`e0236ed1`), 2 commits on `8ae5cfcb`

`MonType::no_shield_block` (MonStats `NoShldBlock`), `GameData::shield_code(code)` (`item_pieces[code].type == "shie"`), `Monster::block_chance(game_data)`: 0 unless `no_shield_block` or `shield_code(npc.comp[7])`, then `min(ToBlock of the difficulty, 75)`; `Monster::target` uses it. `docs/research/re/combat.md` records the rule. `test_monster_block` pins 32 chances and the cap (7 values) from game.exe's own run.

### 1.3 Mutants of the claim (V run; builder, long lane, never pushed)

Each mutant is one text change on a copy of the claim (`mb_mut.py`: rebuild `test_monster_block` in a copy of the tree, run its pins and the compare of 2,469 cases).

| Mutant | pins | compare (of 2,469) | status |
|---|---|---|---|
| m1 the parent: `can_block` decides | abort | 228 differ | V run |
| m2 gate A (NoShldBlock) dropped | abort | 45 | V run |
| m3 gate B (the shield) dropped | abort | 255 | V run |
| m4 no gate at all | abort | 1,560 | V run |
| m5 the cap of 75 removed | expected abort (the cap pins), compare 0 (no row has ToBlock above 75) | | **not run** |
| m6 the cap 74 | expected abort | expected 0 | **not run** |
| m7 the difficulty ignored | | | **not run** |
| m8 layer 6 for layer 7 | | | **not run** |
| m9 `mBL` ANDed into the rule | | | **not run** |
| m10 shield by the `component` column (the torch has 7) | | | **not run** |
| m11 any item code counts as a shield | | | **not run** |
| m12 gate A read from the `boss` column | | | **not run** (needs the library rebuilt: `mb_mut.py`) |
| m13 `shld` for `shie` | | | **not run** |
| m14 `NoShldBlock` any non-empty | equivalent (the column holds `1` or nothing) | | not run |
| m15 gate A only with `mBL` | | | **not run** (`mb_mut2.py`) |

**4 of the 15 mutants were run (all four abort the pins and differ on the compare).** The other eleven are scripted: `mb_mut2.py` compiles only `test_monster_block`'s own translation unit against the mutated header and links it with the built libraries (seconds, one compiler process), `mb_mut.py` rebuilds the library (minutes). They did not run because my first `mb_mut.py` job was cancelled while the builder's memory guard had stopped it: **six `cc1plus` processes of job 20261119-221208-2e4b (pids 2221328, 2221425, 2221453, 2221529, 2221501, 2221685; state T, about 6 GB) were left on the builder**, the container had under 1 GB of headroom, and every later long-lane job of mine waited on the guard. Reaping them (the coordinator or the user) unblocks `mb_mut2.py`. The verifier's mutants are the check that counts. Pins that should catch them: the cap pins (m5, m6), `skeleton1` difficulty 1 and 2 and Mephisto at three difficulties (m7), the torch variant and `skeleton1` variant 0 (m8, m10, m11), Mephisto's `mBL` 0 (m9, m15).

### 1.4 Build and tests (V run)

Fresh build directories per claim (a header edit rebuilt a stale object once: the first claim build kept the parent's `MonType` layout and read `NoShldBlock` as 1 for the skeletons; the numbers above are from fresh builds). Parent oracle binary: rc 0; `test_monster_block` (pins) on the claim: `ok` (V run); compare on the claim 0 of 2,469 (V run). The claim's compare was run once, on the one set of cases: the control (the same cases with the unit spawned in another room, `MB_SPOT=5`) is written but not run (section 4). `pr_guard.py`: 0 FAIL 0 WARN. **The full ctest of the tree was NOT run** (builder queue). `Target::block` is read in `combat.hpp:956` (`player_blow`) and `:1113` (`monster_blow`) and copied in `fight/tick.cpp:429`; the only producer is `Monster::target` (V read, `git grep '\.block'`). `presubmit.py` was not run.

### 1.5 Deviations (d2d against game.exe)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D1 | game.exe: gate A (NoShldBlock) or a layer-7 shield code (type `shie`), chance ToBlock (cap 75). d2d (parent): `mBL` then ToBlock. Fixed by the candidate | V run | yes: `monster_block.py` + `test_monster_block` (built, fork branch) |
| D2 | The shield entry per spawn: game.exe `+0xb` from `FUN_005739d0`'s look; d2d `monster_look` into `comp[7]` | S (population port), I | an oracle would run `FUN_005739d0` natively on a unit with a set seed and the region's component sets (`FUN_005bdb20`) and compare the layer-7 byte with `monster_look`; not built |
| D3 | A monster made outside `spawn_monsters` (a summon, a clone, a minion's pick in `make_monster`, `ai.cpp:2753`): d2d draws `comp[layer]` from the game rng, uniform over its list; game.exe's own look pick per path not read | U | read the callers of `FUN_005739d0` for a raised, cloned or summoned unit; then the same oracle |
| D4 | An Act 2 level's BaseId 0 and 0xaa override tables | I (layer 7 equal by the tables) | bring up an Act 2 level in the emulator (the bring-up faults at `0x67c581`) |
| D5 | The monster-vs-monster block (a merc or pet hitting a monster) | S | same `Target`, so it follows D1 |

### 1.6 Bot findings (described, not changed)

| B | Finding | Where | Tag |
|---|---|---|---|
| B1 | The bot loads the monster's `ToBlock` per difficulty and nothing reads it; no kill-time or aggro model counts a block. A Paladin kill-time model must use the two gates (not `mBL`) | `internal/data/monsters.go:65` and `:137` | V read |
| B2 | A Paladin's Smite and Zeal against Mephisto (Act 3 Normal): 20 percent of landed blows block in game.exe, which d2d ignored until the claim; boss-fight durations measured on d2d before the claim are short by about that share | none (follows B1) | I |
| B3 | Fallens and Corrupt Rogues blocked 9 to 40 percent in d2d: any bot calibration of Act 1 kill rates against d2d is pessimistic for those rows | measured runs (none located) | I |

### 1.7 PR candidate

**Yes, `fid/monster-block-r1`** (local; private fork remote only; not posted; awaiting the independent verifier). Mismatches before: 228 of 2,469; after: 0. Standard (docs/upstream/verifier-*): the oracle commit first and it builds alone; game.exe's functions run natively; `compare` exits 1 on the parent and on an empty, short or malformed file; the claim's test pins game.exe's values; no force-push; the body is `docs/upstream/bodies/monster-block-body.md`. For the verifier to grind: the look link (D2), the Act 2 tables (D4), the placeholder rows, the `by_id` index (the test finds the row by `types[row].id`, because `monsters.by_id` lands one row early after a repeated Id: a first oracle run via `by_id` showed 12 false differences on `skeleton8`).

### 1.8 Drill recipe (lab binary, fork verbs; run later, not now)

K1 (D1, Fallen): `debug spawn fallen1 40`, then `debug mon set <id> life 5000 maxlife 5000` on each; hit each with the same plain melee swing from a throwaway character with a high attack rating (always hits) and count blows with no damage and the block mode; expect about 9 percent on the Fallen that carry a shield (the `monsters` listing does not show a component: U whether a verb shows `comp[7]`) and none on the others in the claim; 9 percent on every Fallen in the parent. K2 (D1, Mephisto): `debug spawn mephisto` at Normal, same swing: expect 20 percent in the claim, none in the parent. K3: a skeleton pack: the `nil` entry never blocks, the others 10 percent (skeleton1) in the claim. K4 (D2): spawn 100 `skeleton1` in a region and count the share that block: game.exe's share is the population's, compare with a game.exe level run (needs the original game).

### 1.9 Matrix-cell updates (proposed text)

- `docs/probes/fid/fid-monster-block.md` D1 / D3 / D4 and its MA13 row: "**V run (decomp43)**: the shield is the layer-7 look entry; gate A is NoShldBlock; built `fid/monster-block-r1`".
- `docs/research/fidelity-matrix.md` MA10 (F14 missile block gate): unchanged by this; MA13 as above.
- `docs/upstream/pr-queue.md` (shared; the merger): a row "Monster block gate, `fid/monster-block-r1`, ready for the verifier".

## 2. The crush rule for merc, pet and missile blows

- **Missile:** already done. The harness of #637 (r3) takes the `missile` and `fire` columns (V read of `crush_hit.py`'s header in the fork branch `fid/crush-d6-r3`: FUN_0057c6c0's second argument 1, events 6 and 2, divisor doubled), so decomp42's D3 and D4 are closed there; nothing to build.
- **Merc and pet as the struck unit or the striker:** not built. d2d applies a monster's crush on a merc after the hits (`fight/tick.cpp:698`, `:837`), game.exe before them (decomp41 D7); the get-hit size of a merc or pet needs the same chain with the unit type 0 or a pet's record (`FUN_0057cee0` hooks differ). Reach: a merc with a Crushing Blow item or a monster with a crush mod against the merc: none in Normal Act 1-3 for a Paladin or Sorceress hero (I). Not provable without a harness for the merc's get-hit path: effort M for no early reach.
- **P09 raised-monster mode:** unchanged from decomp42 (negligible; 24 ticks).

## 3. Scan: the three most player-visible remaining deviations (Normal Act 1-3, idle Paladin and Sorceress)

Method: our docs only (`docs/probes/fid/*`, `docs/research/fidelity-matrix.md`, `mon-wake-trace.md`, `sk-teleport-trace.md`, `docs/research/re/todo.md`), checked against upstream main `8ae5cfcb` with `git grep` where a row could have been fixed since. No live game, no survival-log analysis in this run (the survival A/B anomalies named in decomp42 are bot-side: S). **Not an exhaustive scan: I.**

1. **Which monsters think (the active rooms): the biggest open structural item.** `mon-wake-trace.md` section 4: game.exe runs the units of the near list's rooms (`FUN_0066c370`), d2d every monster within 30 cells (150 subtiles); a monster 31 to 35 subtiles away acts in d2d and not in game.exe; the first think after a spawn is also unread. Tag S (V read of d2d), game side U. Provable by an emulator run of the room list against `monsters_awake()`; effort M. Visible in every fight (who joins, when).
2. **Andariel and the other Act 1 bosses' skill chains.** `fid-boss-skills.md`: the death handler and level / life / defence / experience are 0 differ, but the skill and AI chains (Andariel's poison cloud and spray, Blood Raven's raise, the Countess, Corpsefire) have no oracle; upstream ports Andariel's missiles (`ai.cpp:490..548`, `:1659`, `:1818`) and `tools/emu` has Countess scripts (`ai_countess_*`) but none for Andariel, Blood Raven or Corpsefire (V read of the directory). Tag U. Visible: the Act 1 quest boss every character fights. Effort L (one chain per boss).
3. **Monster crowd collision and stuck cases (MA12).** Velocity and pathers are 0 differ; crowd collision and the exit trigger at a cave mouth are not compared (`fid-path-stuck.md`). Tag U. Visible: a Paladin walled in by a pack. Effort M; needs a unit-collision oracle.

Closed since the earlier lists (V read upstream `8ae5cfcb`): the Teleport mana on a refused hop (SKTELE D4: `fight/strike.cpp:1115`, a refused hop is free), the minions' mod inits (`ai.cpp:2857`), Zeal's strike gap (`276f85c0`), curse slot, hit recovery, merc life.

Nothing in this section is a PR candidate: each needs a new differential first.

## 4. Limits and dead ends

- Act 2 levels cannot be brought up in the emulator (`mb_act2.py` 41, 40, 42, ...: the bring-up faults in `drlg`; a later pass hung the lane and was cancelled): the Act 2 override of Q5 is read, not run.
- The first d2d-side oracle used `monsters.by_id` and reported 12 false differences on `skeleton8` (the map is off by one after a repeated Id: `bugs.md`): fixed in the test before the oracle commit.
- A first claim build kept stale objects (montypes.hpp changed; `gamedata_load.cpp` was not rebuilt) and printed the parent's numbers: a fresh build directory fixed it; every number in section 1 is from fresh builds.
- The builder's memory (see 1.3) blocked: eleven mutants, `mb_edge.sh` (the exit codes of `compare` on bad input), the room control `MB_SPOT=5`, and the first-time `ctest` of the fight tests. Each is a command in this branch (`tools/emu/mb_mut2.py`, `mb_edge.sh`).
- A Windows or macOS build was not run; `ctest` of the whole tree was not run.

## 5. Shared-file lines for the merger

None written: BACKLOG, CHANGELOG and INDEX lines are generated at merge from the task card (TASK-683: title, `result:`, ticked boxes, `owns:`).
