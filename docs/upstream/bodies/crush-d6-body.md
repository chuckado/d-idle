## What changes

A Crushing Blow takes a share of the monster's life directly (`FUN_005bffc0` writes the unit's stat 6) and is not part of the damage record's total (`+0x4c`), which `FUN_0057cb00` tests to decide whether a blow is big enough for a get-hit. d2d added the crush to the blow's damage and `hurt` used the sum as the blow's size for `rules::get_hit`, so a Crushing Blow proc staggered a monster that the same blow without the crush would not.

`Blow::crush_fine` carries the crush's 256ths; `hurt` takes it off the size for the get-hit only (the monster's life still loses it); `Fight::land` calls the new `blow_hurt` (the soak and the `hurt` call it made inline) so the test drives the same line. `test_crush_get_hit` pins nine game.exe decisions; `docs/research/re/combat.md` says where the crush stands in the total.

Two commits: the oracle (`tools/emu/crush_hit.py`, `tests/game/test_crush_get_hit.cpp --crush-hit`, a README row), then the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0057cb00` `0x57cb22..0x57cb35` | the size test reads the record total `+0x4c` (under `0x100`: no get-hit) |
| `FUN_0057c1e0` `0x57c3a5..0x57c3bd` | `+0x4c` = `+8 + 0x10 + 0x1c + 0x20 + 0x24 + 0x28` (plus `+0x38` for a flagged unit): physical, the five elements; no crush |
| `FUN_0057c6c0` `0x57c793..0x57c7a0`, `0x57c7b7..0x57c7f3`, `0x57c7f8..0x57c826`, `0x57c852` | a missile flag (second argument) makes it call `FUN_0057c1e0` itself; the strike events (5 then 1 melee, 6 then 2 missile) run `FUN_005bffc0`, which writes stat 6; the record's physical is cut to the life the crush left; `+0x4c` is read at `0x57c852` and taken off stat 6 |
| `FUN_005bffc0` | the crush handler writes stat 6 and the kill bit; it never writes the record's `+0x4c` |

## Counted differentials (V run)

`crush_hit.py game` runs, per case, `FUN_005bffc0` natively (Crushing Blow 100 percent, a melee strike) for the life the crush leaves, `FUN_0057c1e0` natively for the total, `FUN_0057c6c0` natively (the dispatcher sets stat 6 to the crush's life at event 5, as `crush.py pair`) and `FUN_0057cee0` natively on the same record (the get-hit and the seed's rolls, as `hit_rule.py`). It prints whether the monster went into GH and its seed. `test_crush_get_hit --crush-hit` prints the same from d2d's real `player_blow`, `soak` and `hurt`. A case is a monster that survives the blow: max life, life, physical, hit class (0 to 13 and random up to `0x5f`), the crush divisor (4 or 8) and the monster's seed.

| | seed 7 | seed 5001 | seed 9091 |
|---|---|---|---|
| cases | 100,800 | 100,800 | 100,800 |
| parent | 72,108 differ | 72,163 | 72,111 |
| claim | 0 | 0 | 0 |
| control (no crush, both trees) | 0 | 0 | 0 |

Mutants of the claim (one text change each, `test_crush_get_hit` rebuilt): the parent, `crush_fine` unset, `blow_hurt` passing 0, the crush taken off twice, the crush halved abort the pins and differ on the compare (72,108, 72,108, 72,108, 23,035, 50,307 of 100,800); the crush cut to whole points differs on 9,300 (and the ninth pin is its pin); a floor at 0 on the size is an equivalent mutant (the size is never negative); a missile blow carrying no crush and `marked` dropped are not caught: the cases are melee and carry no element. Both test binaries print ok; the 40-test fight subset of ctest passed (5 skipped), the whole-tree ctest has not run.

## In play

A Paladin with a Crushing Blow weapon hits plain monsters of 12 to 1,024 points of life for 2 to 20 points. On a grid of 4,890 such blows that crush: game.exe puts the monster into hit recovery on 1,921 (39 percent); the parent on 4,513 (92 percent), because the crush alone, a quarter of the life, is more than the size game.exe needs (a sixteenth for class 0). The claim: 1,921, equal blow by blow (seed included).

## Limits

- The melee path only: the missile record and event 6 were not run (the divisor doubling is pinned upstream).
- The record's elements are 0 in the cases, so a crush together with an element's hit class is not run.
- `Fight::land` calling `blow_hurt` is read, not run: `test_crush_get_hit` drives `blow_hurt`, and putting the old inline `soak` and `hurt` back in `land` would not fail it.
- Which function fills the melee record's total before `FUN_0057c6c0` (the second argument is 0, so it does not) was not read; the harness runs `FUN_0057c1e0` itself first.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cd build && ninja test_crush_get_hit && cd ../tools/emu && uv sync
set -o pipefail
uv run python crush_hit.py gen 100000 7 > /tmp/c.tsv
uv run python crush_hit.py game /tmp/c.tsv > /tmp/g.tsv
../../build/tests/test_crush_get_hit --crush-hit /tmp/c.tsv | grep -E '^[0-9]+ [01] [0-9]+ [0-9]+$' > /tmp/d.tsv
uv run python crush_hit.py compare /tmp/c.tsv /tmp/g.tsv /tmp/d.tsv   # parent: differs, exit 1; claim: 0, exit 0
../../build/tests/test_crush_get_hit
```

---

Agent: Claude Code
