## What changes

When a death's experience share reaches the level's start, game.exe (`FUN_005359f0`, `0x535a63..0x535a6f`) sets the experience to `start + 1` and tells the corpse `experience - start`: one more than the points that went. `World::death_penalty` told it `experience - (start + 1)`. It now tells what game.exe does.

Two commits: the oracle (`tools/emu/death_penalty.py`, `tests/game/test_death_penalty.cpp`; the whole penalty, not only this value), then the claim. The gold half, the stash, the pile and the goldlost stat were compared as well and already equal.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00580ec0` `0x580f51` | a death calls `FUN_00535ab0(game, player, killer)` |
| `FUN_005359f0` `0x535a3b..0x535a95` | loss = `DifficultyLevels +4 x (high - low) / 100`; if `experience - loss` is not above `low`: told `experience - low`, experience := `low + 1`; else told `loss` |
| `FUN_005357d0` | gold: `min(clvl, 20)` percent of purse + stash, floor `clvl x 500` in single player (`game +0x6a == 3`), the rest falls as one pile (`FUN_00535510`), stat 175 |

The game side is game.exe's own `FUN_00535ab0` under the emulator, 120,000 random cases per seed (classes 0..6, difficulties, levels 1..99, experience at the edges of the loss band, purse and stash around the toll and the floor).

## Claims

| Commit | Claim or change | Evidence |
|---|---|---|
| `787628ff` | the oracle and `test_death_penalty --death-penalty CASES` | seed 1: 31,792 of 120,000 differ, seed 7: 31,750, all in the experience told to the corpse |
| `c7746e98` | the corpse is told `experience - start` when the share reaches the level's start | 0 of 120,000 on both seeds |

## How to re-check

```sh
set -o pipefail
cd tools/emu
uv run python death_penalty.py gen 120000 1 > cases.tsv
uv run python death_penalty.py game cases.tsv > game.tsv
../../build/tests/test_death_penalty --death-penalty cases.tsv | grep -E '^c[0-9]+( -?[0-9]+){6}$' > d2d.tsv
diff game.tsv d2d.tsv      # no output with the claim
cd ../..
python3 tools/pr_guard.py
```

## Limits

The effect is at most one point of the corpse's experience (the corpse returns 75 percent of it) and only when a Nightmare or Hell death happens within 5 or 10 percent of a level's span after its start; the grid over-samples that band. Separately, `FUN_0057f700` (`0x57f81c..0x57f86c`) takes `(told / 100) x 75` above 1,048,576 points where d2d takes `told x 75 / 100`; this is read, not run (clvl 60 and up), and not in this change. Game types other than single player (the stash overflow, a killing player) are not modelled by d2d and not touched.

## Checklist

- [ ] `python3 tools/presubmit.py --quick` 0 FAIL, `python3 tools/pr_guard.py` 0 / 0, ctest as the parent (not run yet; verifier)
- [ ] each commit builds alone (both built on the builder: yes)
