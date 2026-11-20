# verify43: independent verification of the monster block gate PR (P07, `fid/monster-block-r1`)

> status: current, 2026-11-20 (agent verify43, Opus, the independent verifier; docs only, nothing posted, no live game). Verifies [decomp43-monster-block-trace.md](decomp43-monster-block-trace.md) section 1 and the body `docs/upstream/bodies/monster-block-body.md`. Branch in the private fork: oracle `e5425263`, claim `e0236ed1`, README row `91e3d4c5`, on upstream `8ae5cfcb` (the private remote holds the same tip, V run `git ls-remote`).

Tags: **V run** (run here, builder long lane or locally), **V read** (disassembly or source read here), **S** (taken from decomp43 or another doc), **I** (inferred), **U** (unverified).

## Verdict: GO WITH FIXES

The rule is right on every input the oracle covers, the compare refuses bad input, and all 17 non-equivalent mutants abort the pins (15 of them also differ on the compare; the two cap mutants are caught by the cap pins only); two mutants are equivalent on the shipped data. Before posting: rebase (one trivial conflict), fix the test's skip code, correct the body (commit count, mutant paragraph). Details in section 6.

## 1. The claim against game.exe (re-derived, V read)

Disassembled here from a 1.14d `Game.exe` with capstone (`FUN_00622720` `0x622720..0x622822`, `FUN_006225f0` `0x6225f0..0x622696`):

- Unit type 1 (monster): the MonStats record (`[0x744304]+0xa78`, stride `0x1a8`, class bounded by `+0xa80`), byte `+0xf` `and [0x6ce270]`; non-zero goes straight to `FUN_00625480(unit, 0x14, 0)` (stat 20). Else `FUN_006225f0(unit)` (ecx); 0 returns 0. Then `cmp ebx, 0x4b; jge` returns 75. The player-only part (`[ebp+0xc]`, dex and clvl) is not on the monster path. A class out of range skips gate A and goes to gate B. Matches decomp43 Q1. Negative values pass through uncapped on both sides (d2d `std::min(.., 75)`; `combat.hpp` tests `block > 0`).
- `FUN_006225f0`: class `- 0xf3`, if at most `0x74`, an index byte table at `0x6226a4` into three targets at `0x622698`. **Decoded here (decomp43 left it U):** classes 243 (diablo), 310 (doomknight1) and 333 (diabloclone) return 1 (gate B always open), 359 (act3hire) returns 0, the other 113 classes take the normal path. The first three are NoShldBlock rows, so gate A already opens them; act3hire is neither NoShldBlock nor carrying a shield. So the jump table changes no outcome, and the oracle ran all of them natively (game.exe 25/40/50, 24, 50, 0 by difficulty, V run from the stored `game.tsv`). The `hireling range` wording in the trace is wrong: it is a four-class special case.
- Normal path: monster data `[unit+0x14]+0xb` (byte) as idx, `FUN_00664860(unit, 7, idx)`; the code is refused when it is in the list at `0x744580` (count at `0x744584`); else `FUN_00633640(code, &tmp)` and the record's word `+0x11e` must be 2. Matches decomp43 Q3.

**The oracle (`monster_block.py game`) is the right instrument (V read):** it spawns each row with the real spawner, writes the variant into `[unit+0x14]+0xb` (the byte `FUN_006225f0` reads) and calls `FUN_00622720(unit, 0)` natively; it prints the chance, which is the player-visible quantity (`combat.hpp:956` rolls it against `rng(100)`). The d2d side (`test_monster_block --monster-block`) builds the monster with the real `make_monster`, sets `npc.comp[7]` from `parts[7][variant]` as `spawn_monsters` does (`ai.cpp:215`, V read), and prints `Monster::target(...).block`.

## 2. Counted differential (V run, builder long lane, job 20261120-132016-a098)

| Check | Result |
|---|---|
| The claim tree on the builder | the six touched files hash-identical to the fork's `fid/monster-block-r1` |
| Pins (`test_monster_block`, no argument) | `ok`, rc 0 |
| Parent's stored d2d output (`d2d-parent2.tsv`) against `game.tsv` | 228 of 2,469 differ (195 + 33), rc 1 |
| Parent, first run (`d2d-parent.tsv`) | 240 (the extra 12 "both block, chances differ" are the `by_id` artefact decomp43 names) |
| Claim, a fresh d2d run here | 0 of 2,469, rc 0; byte-identical to decomp43's `d2d-claim4.tsv`; no `-1` line |
| `NoShldBlock` values (m14) | `''` 716, `1` 18: m14 is equivalent; `primeevil` disagrees on 6 rows |

The game side rerun, the `MB_SPOT=5` room control, `codes` and `cap`: section 3.

### 2.1 Exit codes of `compare` (V run)

On an 8-case slice: a good triple 0; empty game, empty d2d, empty cases, short d2d, a malformed line, a wrong id, a flipped chance, all `-1` on the d2d side, all `-2` on the game side (100 percent skipped), a doubled d2d file, a 3-column line, a case line with difficulty 3, a missing file, `game` on an empty file: **all 1**. (b) holds.

## 3. Game side rerun and room control

V run, builder long-lane job 20261120-132422-41a1 (`v43_game.sh`; my first job could not run these: the stored environment pointed at a build directory whose `mpq-cat` is gone, this one takes the newest built `mpq-cat`):

| Check | Result |
|---|---|
| `gen` | byte-identical to the stored `cases.tsv` |
| `codes` | 659 item codes, 24 shields in game.exe, 0 where the `type` column disagrees, rc 0 |
| `cap` (Mephisto's stat 20 set by `FUN_00627260`) | 0, 1, 74, 75, 75, 75, 75 for 0, 1, 74, 75, 76, 90, 200 |
| `game` rerun | byte-identical to the stored `game.tsv` (deterministic); claim 0 differ rc 0, parent 228 rc 1 |
| `MB_SPOT=5` room control (decomp43 wrote it, never ran it) | claim 0 of 2,469 differ, rc 0 (93 skipped as before) |

The result does not depend on the spawn room.

## 4. Mutants (V run)

decomp43 ran m1 to m4 (S). Job 20261119-225204-80bf (decomp43's queue, it did finish: V read of its log) ran m5 to m9 and m15; **m10, m11 and m13 failed to compile** in that job (a shadow `gamedata.hpp` on the include path is included twice next to the real one), m12 and m14 were not in it. I ran those as the same rule written into the one gate line of `ai.hpp` (the header the runner can shadow), plus a control and four extra mutants (`v43_mut.py`, job 20261120-132016-a098):

| Mutant | Pins | Compare (of 2,469) | Source |
|---|---|---|---|
| m0 control: the claim unchanged through the same runner | **ok (rc 0)** | 0, rc 0 | verify43 |
| m1 parent rule (`mBL`) | abort | 228 | decomp43 (S) |
| m2 gate A dropped | abort | 45 | decomp43 (S) |
| m3 gate B dropped | abort | 255 | decomp43 (S) |
| m4 no gate | abort | 1,560 | decomp43 (S) |
| m5 cap removed | abort (rc 134) | 0 (no row above 75: only the cap pins catch it) | job 80bf |
| m6 cap 74 | abort | 0 (same) | job 80bf |
| m7 difficulty ignored | abort | 30 | job 80bf |
| m8 layer 6 for 7 | abort | 255 | job 80bf |
| m9 `mBL` ANDed | abort | 33 | job 80bf |
| m10 shield by the `component` column | abort | 24 | verify43 |
| m11 any item code is a shield | abort | 24 | verify43 |
| m12 gate A from the `boss` column | abort | 49 | verify43 |
| m13 `shld` for `shie` | abort | 255 | verify43 |
| m14 `NoShldBlock` non-empty | equivalent (the column holds `1` or nothing, V run) | | verify43 |
| m15 gate A only with `mBL` | abort | 33 | job 80bf |
| m16 gate A from `primeevil` | abort | 13 | verify43 |
| m17 gate B by code name (not empty, `nil`, `tch`) | abort | 66 | verify43 |
| m18 gate B also by `type2 == shie` | **survives**: pins ok, compare 0 | | verify43 |
| m19 the torch counts as a shield | abort | 24 | verify43 |

m18 is equivalent on the shipped tables (no layer-7 code has `type2` `shie` without `type` `shie`, I from the 0 differ over all 2,469 cases and `codes`' 0 disagreements); a `type2` reading is not what game.exe does (word `+0x11e` is the primary type), so nothing to add. **14 of 14 non-equivalent mutants of the decomp43 list (m14 is equivalent) and 3 of 3 non-equivalent extra ones (m16, m17, m19; m18 is equivalent) abort the pins.** m5 and m6 rely on the cap pins alone (the compare cannot see them): keep the cap block in the test.

## 5. Scope and commit hygiene (V run unless tagged)

- Three commits on `8ae5cfcb`, each author and committer the public noreply identity; each message ends with a blank line and `Agent: Claude Code` as the only trailer (`git interpret-trailers --parse`). The private remote's tip equals the local tip (no force-push visible there; U for the reflog of earlier pushes).
- `pr_guard.py 8ae5cfcb fid/monster-block-r1` (the branch's own copy): 0 FAIL, 0 WARN.
- One topic: the block gate, its oracle and its README/combat.md lines. 8 files, +384 −3.
- The oracle commit carries its own test file and does not need the claim (decomp43 built and ran it on the parent: S, consistent with the stored `d2d-parent2.tsv`).
- **Upstream has moved:** `upstream-gh/main` is 121 commits past `8ae5cfcb` (#634 and #637 are merged). `git merge-tree` reports **one conflict in `components/game/ai.hpp`**: upstream `13c0b939` added `gear_res` to `Monster::target`'s resistance loop on the very line block the claim edits. Trivial to resolve; the oracle and pins must be rerun on the rebased tree.

## 6. Findings

| # | Rank | Finding | Evidence | Recommendation |
|---|---|---|---|---|
| F1 | high | The branch does not apply to current upstream: conflict in `ai.hpp` (`gear_res`, upstream `13c0b939`) | V run `git merge-tree` | Rebase on current upstream main, keep `res[i] += boss_res[i] + gear_res[i]` and call `block_chance`; rebuild fresh, rerun pins and the compare (expect 0) and the parent compare (expect 228) |
| F2 | high (CodeRabbit will flag it) | `test_monster_block` without arguments returns **0** when the MPQs or the 1.14d tables are missing (`return oracle ? D2D_TEST_SKIPPED : 0`); upstream's rule (and CodeRabbit's learned note from #634, where we fixed the same thing in `9daf732b`) is `D2D_TEST_SKIPPED` in every mode | V read of the test and the #634 thread | Return `D2D_TEST_SKIPPED` on both guards in both modes |
| F3 | medium | Body says "Two commits: the oracle (..., a README row), then the claim"; the branch has three and the README row plus `MB_SPOT` are in the third | V run `git log` | Either fold `91e3d4c5` into the oracle commit (private branch, before any push upstream) or say three commits |
| F4 | medium | Body "Not yet run" mutant paragraph is stale | this doc, section 4 | Replace with: 14 of 14 non-equivalent mutants (cap, difficulty, layer, `mBL`, component column, any code, `boss` column, `shld`) abort the pins; m5/m6 only through the cap pins; m14 equivalent |
| F5 | medium | The trace calls the jump table in `FUN_006225f0` a "hireling range ... not read: U" | V read, section 1 | Correct the trace and body Evidence row: four classes (243, 310, 333 forced open; 359 forced closed), no outcome changes |
| F6 | low | The test is not in `d2d_oracle_tests` / `oracle_test.py` RECIPES, so ctest never runs the differential | V read `tests/CMakeLists.txt` | Optional (the recent oracles of #632 to #637 are not registered either); mention in the body only if asked |
| F7 | low | `compare` accepts Python int spellings like `1_0` or `+5` | V read `load()` (`int(v)`) | Optional: match `-?[0-9]+` as in #634's `crush_div.py` fix, so CodeRabbit does not raise it |

## 7. Risks: what else the change touches (V read unless tagged)

- **Producers and readers of `Target::block`:** only `Monster::target` produces it; readers are `player_blow` (`combat.hpp:956`), `monster_blow` (`:1113`) and the Confuse/Attract path that turns a monster into a foe (`fight/tick.cpp:427..429`). In game.exe all of these ask `FUN_00622720` of the struck monster, so the claim applies uniformly (S/I: `FUN_0057dfb0` is the shared caller, decomp43).
- **Mercenaries:** d2d computes a merc's block from its gear (`fight/merc.cpp:109..112`), not through `Monster::target`: unchanged. game.exe gives the merc rows 0 (V run, stored `game.tsv`).
- **Player pets** (`Fight::pet_foe`, `fight/pets.cpp:81..93`) get no block at all in d2d: unchanged by the claim. game.exe's `FUN_00622720` makes no pet distinction, so a Raise Skeleton pet whose look holds a shield should block there (`necroskeleton` variant 1: 3 percent at Normal, pinned). **New deviation D6 (I):** d2d pets never block. Not in this PR; a follow-up oracle can reuse `monster_block.py`. Note `fight/summons.cpp:173` gives a pet `parts[layer].front()`, entry 0 (`nil` for the necroskeleton): even with a fix it would never block unless the pet's look pick is ported.
- **Monsters made outside `spawn_monsters`** (minions, `make_monster` at `ai.cpp:2753`: a uniform pick over the list): the block now follows that pick. Parent behaviour for these was the `mBL` rule (wrong for every no-shield entry), so the claim is no worse; whether game.exe's pick matches is decomp43's D3 (U).
- **MonEquip gear:** the oracle overwrites `+0xb` after the spawn, so a look change that an equipped MonEquip shield might cause in game.exe is not covered; d2d's `wear_gear` never touches `comp[7]` (V read: the only `comp[` writers are `ai.cpp:215`, `ai.cpp:2753`, `summons.cpp:173`, `gamedata.cpp:192/693`). U: whether any MonEquip row equips a shield. Same for stat 20 from gear or boss mods: both sides read `ToBlock` from the table, the parent did too (no regression; unique and champion units were not in the oracle).
- **Act 2 override table** and the **per-spawn look** (decomp43 D2, D4): unchanged, correctly declared as limits in the body.

## 8. Open upstream PRs #634 and #637 (read-only, 2026-11-20)

Both **merged** by the maintainer (#634 at 11:11 UTC, #637 at 11:19 UTC); all CI checks SUCCESS (Flatpak and Publish skipped, CodeRabbit neutral). CodeRabbit's last comments were answered in-thread (#634: fixes `9daf732b`, `90b76e9a`, `47f20829`; #637: `d54deb57`); no open CodeRabbit request remains, **no human comment** on either PR besides the merge. Nothing to follow up there.

## 9. Recommended actions (in order)

1. Rebase `fid/monster-block-r1` on current upstream main as `fid/monster-block-r2` (F1), apply F2 (and F7 if cheap), fold or describe the README commit (F3).
2. Fresh build; rerun pins, claim compare (0), parent compare (228), `pr_guard.py`.
3. Update the body: commit count, the mutant paragraph (F4), the jump-table row (F5), and add "a player pet's block (d2d gives none) is not changed here" to Limits.
4. Card a follow-up for D6 (pet block) and the MonEquip question; both are research, not this PR.
