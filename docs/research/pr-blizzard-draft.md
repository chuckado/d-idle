# PR draft: Blizzard's shards fall -(R-1)..(R-2) of the centre (SKGROUND2, D1)

> status: DRAFT, local candidate only: branch `fix/blizzard-shard-offsets` on the worker host in the scratch checkout of upstream main 7d5b3a5c (two commits, 9a1a90eb tooling and 681d52a9 claim). Nothing pushed or posted. An independent verifier (d2-reviewer, own harness) is the next step; the user's go comes after it. Evidence: [sk-ground2-trace.md](sk-ground2-trace.md) section 5; earlier trace: [sk-ground-trace.md](sk-ground-trace.md) (D1, game side).

Title: `Blizzard's shards fall -(R-1)..(R-2) of the centre, not -(R-1)..(R-1) (FUN_005a9820)`

## What changes

game.exe's Blizzard centre (missile do 10, FUN_005aea60; Eruption's do 25 calls the same FUN_005a9820) drops one shard every calc2 frames at `rand(2r) - r` subtiles from its own subtile on each axis (x drawn first), with `r = calc1 - 1`. At Blizzard's calc1 of 7 that is -6..5: twelve values, so the pattern is a 12 x 12 square sitting half a subtile toward -x / -y of the cast point.

d2d drew `rng(2 radius - 1) - (radius - 1)`: -6..6, thirteen values. The extra +6 row and column are cells game.exe never uses; 1 - (12/13)^2 = 14.8 % of d2d's shards landed there.

The fix is `reach = radius - 1; rng(2 reach) - reach` (and a comment, a header comment line and one line of docs/research/re/skills.md). Radius 1 draws 0 (`rng(0)` is 0), as game.exe's `rand(0)`.

## Proof

`tests/game/test_sk_ground.cpp` (`--blizz SEED N`) casts Blizzard through the real `Fight::fire` on an open level and steps the centre 100 frames with the real `Fight::missile_tick`; it prints each shard's subtile offset from the centre. `tools/emu/sk_ground_blizz.py game` runs game.exe's FUN_005a9820 natively under the emulator on the same kind of cases (seed, age, frequency, radius); `d2d` compares the two: how many of d2d's shards have an offset game.exe never makes at radius 7.

| build | casts | shards | shards outside game.exe's -6..5 | game.exe's offsets | d2d's offsets |
|---|---|---|---|---|---|
| upstream main + the test, seed 1 | 4,000 | 100,000 | 14,926 (14.93 %) | 12 values, -6..5 | 13 values, -6..6 |
| upstream main + the test, seed 7 | 4,000 | 100,000 | 14,957 (14.96 %) | same | same |
| this branch, seed 1 | 4,000 | 100,000 | 0 | same | 12 values, -6..5, none missing |
| this branch, seed 7 | 4,000 | 100,000 | 0 | same | same |

Every cast makes 25 shards on 25 distinct frames in both builds (game side: FUN_005a9820 on 64,200 cases, 0 differ from the reading, SKGROUND). pr_guard 0 FAIL / 0 WARN on the range and on each commit; test_game and test_rules pass (rc 0) on the branch.

## Limits

- The draw itself is not compared value for value: d2d's generator is not game.exe's seed rule ({cx + age, 666}); the claim is the SET of offsets (which cells can get a shard), not which cell a given cast picks.
- Not claimed: no shard on a blocked spot (FUN_0064cb30 mask 5), the shard's hit area, the skill delay. The collision check is hooked free in the game-side run.
- The centre's float position within its subtile (d2d) against game.exe's whole subtiles is not modelled; the test casts at subtile centres.
- Not run: det state hashes, an MSVC build, a live game.

## How to review

1. `components/game/fight/missiles.cpp`: the one changed expression in `missile_tick` (do 10 / 25).
2. Disassemble FUN_005a9820 (0x5a9820..0x5a9921) with `tools/emu/disasm.py`: `r = calc1 - 1`, two calls of FUN_0045c3e0 with `2r`, then `- r`.
3. Run the differential: `python sk_ground_blizz.py gen 40000 7 > c`, `game c > g`, `test_sk_ground --blizz 7 4000 | grep '^S' > d`, `python sk_ground_blizz.py d2d c g d`.

## Files

`components/game/fight/missiles.cpp`, `components/game/fight.hpp` (comment), `docs/research/re/skills.md` (one line); tooling `tests/game/test_sk_ground.cpp`, `tools/emu/sk_ground_blizz.py`.
