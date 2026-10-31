# PR draft: a picked-up scroll goes into a belt column of the same scroll

> status: DRAFT (agent DECOMPBELT, 2026-10-31). Nothing is pushed, posted or filed. Branch `trace/belt` in `~/d2d-lab/pr/belt` on the worker host (commits 6c2c7347, 41bd5a88 on upstream/main f8aab0e8). Proof: docs/research/pr-belt-proof.md. The body below follows the earlier PRs' template.

````
## What changes
One claim, one commit (plus one tooling commit), on `rules::pick_up`, the ground pickup's auto-place:

A Scroll of Identify or Town Portal that finds no tome with room now goes into a belt column whose bottom item is the same scroll, at its lowest free box, as game.exe does. d2d sent them to the inventory. game.exe's FUN_00628ba0 calls FUN_0063c560 (the "not isc / tsc" bottom test) but only refuses when it returns -1, and FUN_0063c560 returns 0 or 1, so the test never refuses; FUN_0063c600 then chooses the box with no scroll test.

Not changed: potions (0 of the 386,889 non-scroll lines moved), the tome step before the belt, the inventory step after it. The docs/research/re/combat.md step 5 note is corrected.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00563560` 0x563703..0x56372f | the belt step: `FUN_0062bad0` (Beltable), `FUN_00628ba0`, `FUN_0063c790` -> `FUN_0063c600`; any 0: the inventory (`FUN_005600a0`) |
| `FUN_00628ba0` 0x628bea..0x628c10 | autobelt (Items +0x131) -> 1; else `FUN_0063c560`, `cmp eax, -1`, `jne` -> 1 |
| `FUN_0063c560` | returns 0 (0x63c56b, 0x63c5e1, 0x63c5f5) or 1 (0x63c5ea), never -1 |
| `FUN_0063c600` | 1x1; columns 0..3 whose bottom matches (`FUN_00628a40`) give their lowest free box below the belt's boxes; else, autobelt, the first free bottom box; no scroll test |

## How to re-check

```sh
python3 tools/pr_guard.py
cmake --build build && ctest --test-dir build --output-on-failure      # test_game needs D2_MPQ_DIR and D2_PATCH_INSTALLER
# the oracle: game.exe's own FUN_0062bad0 / FUN_00628ba0 / FUN_0063c600, no hooks, on the loaded tables;
# 290,000 grid cases (8 belts x bottom rows x full or not x 29 picked codes) + 100,000 random belts,
# against d2d's rules::pick_up on the same belts:
cd tools/emu && uv run python belt_pickup.py gen 100000 1 > cases.tsv && uv run python belt_pickup.py game cases.tsv > g.txt
../../build/tests/test_game --belt-pickup cases.tsv | grep -E '^[0-9]+ ' | diff g.txt -    # empty (3,111 lines differed before: isc 2,859, tsc 252)
```

Limits: the cases carry no tome and no inventory items (the tome and inventory steps are unchanged and outside the claim); the auto-equip step before the belt (FUN_0055d710) is still not done (existing ponytail). The oracle's units are {type, class} only: every function on the path reads nothing else.

Presubmit: ruff absent on the machine that ran it (the Python is a new oracle in tools/emu); the build and ctest were run by hand (above).

## Screenshots
Not applicable.
````
