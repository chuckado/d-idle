# PR draft: Blizzard's and Eruption's shards fall -(R-1)..(R-2) of the centre (SKGROUND2, D1)

> status: READY, held for the user's go (nothing pushed, opened or posted). Local candidate: branch `fix/blizzard-shard-offsets-v2` on the worker host in the scratch checkout `~/d2d-lab/pr/blizzardfix`, upstream main 7d5b3a5c + two commits, **46c55621** (tooling) and **3d6f3fe7** (claim). The first candidate (9a1a90eb, 681d52a9; branch `fix/blizzard-shard-offsets`) is superseded by it: same code change, plus the verifier's three required changes. Independent verdict (own harness, Opus): READY WITH CHANGES, [pr-blizzard-verifier.md](pr-blizzard-verifier.md) (branch agent/verifyblizzard); the changes are applied here (card blizzardfix). Evidence: [sk-ground2-trace.md](sk-ground2-trace.md) section 5; earlier trace: [sk-ground-trace.md](sk-ground-trace.md).

Title (claim commit subject): `Blizzard and Eruption shards fall -(R-1)..(R-2) of the centre, not -(R-1)..(R-1) (FUN_005a9820)`

## What changes

game.exe's Blizzard centre (missile do 10, FUN_005aea60; Eruption's do 25, FUN_005af880, calls the same FUN_005a9820) drops one shard every calc2 frames at `rand(2r) - r` subtiles from its own subtile on each axis (x drawn first), with `r = calc1 - 1`. At Blizzard's calc1 of 7 that is -6..5: twelve values, so the pattern is a 12 x 12 square sitting half a subtile toward -x / -y of the cast point.

d2d drew `rng(2 radius - 1) - (radius - 1)`: -6..6, thirteen values. The extra +6 row and column are cells game.exe never uses; 1 - (12/13)^2 = 14.8 % of d2d's shards landed there.

The changed line serves both do 10 (Blizzard) and do 25 (Eruption), so Eruption is fixed with it; both skills have calc1 = 7 at every level, so r = 6 for both.

The fix is `reach = radius - 1; rng(2 reach) - reach` (and a comment, a header comment line and one line of docs/research/re/skills.md). Radius 1 draws 0 (`rng(0)` is 0), as game.exe's `rand(0)`.

## Proof

`tests/game/test_sk_ground.cpp` (`--blizz SEED N`) casts Blizzard through the real `Fight::fire` on an open level and steps the centre 100 frames with the real `Fight::missile_tick`; it prints each shard's subtile offset from the centre. `tools/emu/sk_ground_blizz.py game` runs game.exe's FUN_005a9820 natively under the emulator on the same kind of cases (seed, age, frequency, radius); `d2d` compares the two (in our repo that mode is tools/emu/sk_gnd2_blizz_d2d.py): how many of d2d's shards have an offset game.exe never makes at radius 7.

| build | casts | shards | shards outside game.exe's -6..5 | game.exe's offsets | d2d's offsets |
|---|---|---|---|---|---|
| upstream main + the test, seed 1 | 4,000 | 100,000 | 14,926 (14.93 %) | 12 values, -6..5 | 13 values, -6..6 |
| upstream main + the test, seed 7 | 4,000 | 100,000 | 14,957 (14.96 %) | same | same |
| this branch, seed 1 | 4,000 | 100,000 | 0 | same | 12 values, -6..5, none missing |
| this branch, seed 7 | 4,000 | 100,000 | 0 | same | same |

The table was re-run on the new head 3d6f3fe7 (V): the same 14,926 and 14,957 before, 0 and 0 after.

Every cast makes 25 shards on 25 distinct frames in both builds (game side: FUN_005a9820 on 64,200 cases, 0 differ from the reading, SKGROUND). 

Independent counts (the verifier's own harness and probe, written from the disassembly before reading ours; sections 1 and 2 of pr-blizzard-verifier.md; re-run on the new head here, same numbers): real `Fight::fire` and `Fight::missile_tick` in two Release builds, 5,000 casts (`gen-d2d 5000 777`).

| build | skill | casts | shards | off game.exe's -6..5 grid | offsets seen |
|---|---|---|---|---|---|
| upstream main | Blizzard | 3,997 | 99,925 | 14,804 (14.81 %) | -6..6 |
| upstream main | Eruption | 1,003 | 13,039 | 1,872 (14.36 %) | -6..6 |
| this branch | Blizzard | 3,997 | 99,925 | 0 | -6..5, none missing |
| this branch | Eruption | 1,003 | 13,039 | 0 | -6..5 |

Base and head agree on the shards per cast, the frame of every shard and the world rng state after every case (rng(12) and rng(13) both take one step), so no other roll shifts. The verifier's game side: FUN_005a9820 under the emulator against its own model, 0 of 120,000 cases differ (radius -3..100, power-of-two 2r, negative and wrapped ages, blocked spots, masks 5 and 0x45).

Checks on the final head (V, run on the worker host): pr_guard 0 FAIL / 0 WARN on the range and on each commit; `tools/presubmit.py --quick` 0 FAIL (the one WARN is the skipped build); ruff 0.8.6 with upstream's rules plus B905 (`--target-version py312`) clean; Release build, `test_game` and `test_rules` rc 0; gcc -Wshadow -Werror clean on missiles.cpp and test_sk_ground.cpp; clang-tidy (the repo's identifier-length and include-cleaner) clean on both touched C++ files; the candidate fast-forwards onto upstream main 7d5b3a5c, clean against #160 and with the same #159 conflict (ai.hpp) that main has.

## Not in this PR

- D2, the shard's own seed ({cx + age, 666} in game.exe, the shared world rng in d2d): determinism only. The claim is the SET of offsets, which does not depend on the seed. Taking it would take two world-rng steps out per shard and shift every later roll, so it is a separate, larger change.
- D3, a shard on a blocked spot: game.exe makes none (FUN_0064cb30, mask 5 for Blizzard, 0x45 for Eruption); d2d launches it. Correction of the earlier impact line ("they can still hit a monster next to the wall"): by the verifier's reading of ai.hpp:695 (S), `missiles_update` removes a standing missile on a `level->blocked(x, y, 0x04)` cell on its first update, before any hit test, with a visual-only explosion if the row has one; so for mask-0x04 walls it cannot hit. What remains is the mask difference (5 / 0x45 against 0x04), the one-frame visual and Eruption's 0x40 bit. Separate card.

## Limits

- The draw itself is not compared value for value: d2d's generator is not game.exe's seed rule ({cx + age, 666}); the claim is the SET of offsets (which cells can get a shard), not which cell a given cast picks.
- Not claimed: no shard on a blocked spot (see Not in this PR), the shard's hit area, the skill delay. The collision check is hooked free in the author's game-side run (the verifier's run models it).
- U: game.exe's shard count per Eruption cast (13 or 14 depends on whether the centre's do runs at age 0 and at its last frame; d2d gives 13). Not this claim.
- The centre's float position within its subtile (d2d) against game.exe's whole subtiles is not modelled; the test casts at subtile centres.
- Not run: det state hashes, an MSVC build, a live game.

## How to review

1. `components/game/fight/missiles.cpp`: the one changed expression in `missile_tick` (do 10 / 25).
2. Disassemble FUN_005a9820 (0x5a9820..0x5a9921) with `tools/emu/disasm.py`: `r = calc1 - 1`, two calls of FUN_0045c3e0 with `2r`, then `- r`.
3. Run the differential: `python sk_ground_blizz.py gen 40000 7 > c`, `game c > g`, `test_sk_ground --blizz 7 4000 | grep '^S' > d`, `python sk_ground_blizz.py d2d c g d`.

## Files

`components/game/fight/missiles.cpp`, `components/game/fight.hpp` (comment, re-wrapped), `docs/research/re/skills.md` (one line); tooling `tests/game/test_sk_ground.cpp`, `tools/emu/sk_ground_blizz.py` (locals renamed `cell_x` / `cell_y`).

## Shared-file lines for the merger

- docs/BACKLOG.md: BLIZZARDFIX: Blizzard / Eruption shard-offset PR candidate READY, held for the user's go (docs/research/pr-blizzard-draft.md; local branch fix/blizzard-shard-offsets-v2, commits 46c55621 + 3d6f3fe7); verifier's three changes applied (short names, Eruption in the commit, ruff); D2 / D3 stay out.
- CHANGELOG.md: docs: Blizzard shard-offset PR candidate brought to READY (names, Eruption, ruff), draft updated with the verifier's counts.
- docs/INDEX.md, docs/decomp-queue.md, docs/research/README.md: no change.
