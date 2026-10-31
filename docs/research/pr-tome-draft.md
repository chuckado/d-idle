# PR draft: a Town Portal read from the pack opens at once, as game.exe does

> status: DRAFT, complete (agent TOMEPR, 2026-10-31; first draft DECOMPTOME). Nothing is pushed, posted or filed. Branch `fix/tome-use` in the tome worktree of the PR checkout on the worker host: d72b97a7 (tooling), 112f5de0 (claim A), c56670b9 (claim B) on upstream/main 27113a31. Proof: [pr-tome-proof.md](pr-tome-proof.md); pack: [pr-tome-evidence-pack.md](pr-tome-evidence-pack.md). The body follows the earlier PRs' template.

````
## What changes
Two claims, one commit each (plus one tooling commit), on `World::read_portal`, the server side of a right-click on a Scroll of Town Portal or a Tome of Town Portal in the pack (C->S 0x20, `cmd::UseItem`).

**A. An item in hand refuses the read (112f5de0).** game.exe's use-item function `FUN_0055e170` returns before anything else when the player has an item on the cursor (`FUN_0063c1e0`). d2d's own client never sends the use with an item in hand (apps/d2d/town.cpp), but the World had no such test, so another client could read a tome with an item held. Now `read_portal` refuses it.

**B. The portal opens on the use's tick, in any live mode (c56670b9).** game.exe runs the inventory read straight through books.txt pSpell 2 (`FUN_005bf240` -> `FUN_005be290`): no SC, no mode set; the player's mode is read only by the dispatcher's dead test (`FUN_0057eec0`: mode 0 / 0x11, or state 0x36). `FUN_005be290` closes the player's old pair (`FUN_00535430`), then makes the new one (`FUN_0056d130`); `FUN_0055e170` spends the charge (stat 0x46 -1) or removes the scroll only when that succeeded. d2d started an SC cast instead (`Fight::cast_scroll`): the portal came at the SC's action frame (9 ticks for the test's barbarian), the use was dropped while the player was in get-hit, attacking, casting or blocking (`pmode >= 0`), the charge went at the click (a hit cutting the SC, or no free spot, lost it) and with no free spot the old pair stayed. Now `read_portal` closes the old pair, calls `open_portal_at` (which now returns whether it opened) and spends the charge or the scroll only then. `Fight::cast_scroll` and `Hero::item_cast` had no other user and are removed (the ItemCastSound line keeps its charged-skill case).

Not changed: the skill bar's Town Portal (skill 220, srvdofunc 113, `FUN_005bf3d0`: the SC and `portal_due`); the other `open_portal_at` callers (quest portals; they ignore the new bool); the empty tome (a no-op before and after); the portal's spot, its town twin and `use_portal`.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0054d750` | C->S dispatcher: 0x20 passes unless `FUN_0057eec0(player, 1, 0, 0)` is 0 |
| `FUN_0057eec0` | with these args: 0 for state 0x36 or mode 0 / 0x11; no other mode is read |
| `FUN_0055e170` | no cursor item (`FUN_0063c1e0`), item mode 0, a tome's stat 0x46 > 0 (0x55e225..0x55e236), then `FUN_005bf240`; on success stat 0x46 -1 (`FUN_006272b0`, 0x55e2a2) or the scroll removed (`FUN_0055e000`) |
| `FUN_005bf240` | books.txt pSpell (tsc / tbk: 2) -> the table at 0x741790, entry 2 server = `FUN_005be290` |
| `FUN_005be290` | town or level 0x88 -> refuse; `FUN_00535430` (old pair closed), then `FUN_0056d130` (the portal; 0 when no spot); no mode set |
| `FUN_005bf3d0` | the SC path (srvdofunc 113) is the skill bar's; it re-enters `FUN_0055e170` at the action frame |

## How to re-check

```sh
python3 tools/pr_guard.py
cmake --build build && ctest --test-dir build --output-on-failure      # test_game needs D2_MPQ_DIR and D2_PATCH_INSTALLER
# the oracle: game.exe's FUN_0054d750 chain in the emulator (hooks listed in the script), 8,000 grid cases
# (tbk / tsc x charges -2..20, 255, 2^20 x town x item in hand x modes 0..19 x free spot) + 100,000 random,
# against d2d's World::apply(cmd::UseItem) on the same cases:
cd tools/emu && uv run python tome_use.py gen 100000 7 > cases.tsv && uv run python tome_use.py game cases.tsv > g.txt
../../build/tests/test_game --tome-use cases.tsv | grep -E '^[0-9]+ ' | diff g.txt -    # empty
# before (d72b97a7): 29,551 of 108,000 lines differ; after claim A (112f5de0): 23,127; after claim B (c56670b9): 0
```

Limits: state 0x36 and level 0x88 also refuse the read in game.exe; d2d's player has no state 0x36 and Uber Tristram's refusal is not compared (ponytail in `read_portal`). The oracle hooks the item's requirements (`FUN_00628c20`: met), its type and stat 0x46, the town test and `FUN_0056d130`'s spot search (whether a spot exists is a case input; where the portal stands is not compared: that is d2d's existing `portal_spot`). The item is always in the pack (item mode 0): a scroll in the belt is not a case. The client side (what game.exe's client sends on the right-click) is not part of the claim. Sounds are not compared.

Presubmit: ruff absent on the machine that ran it; build and ctest by hand (above). Siblings: only players read tomes (no merc, pet or monster path); local co-op partners use the same `read_portal` with their own pair. Hit paths: none touched (a get-hit only mattered because it cut the SC).

## Screenshots
Not applicable.

Agent: Claude Code
````
