# PR draft: shift-click a beltable inventory item into the belt, as game.exe's 0x63

> status: DRAFT (agent DECOMPCLICK, 2026-10-31). Nothing is pushed, posted or filed. Branch `trace/click` in the click worktree of the PR checkout on the worker host: `c4038d15` (claim), `ff0d8b2f` (tooling) on upstream/main 0dc4c62a. Proof: [pr-click-proof.md](pr-click-proof.md); pack: [pr-click-evidence-pack.md](pr-click-evidence-pack.md). Bundle-sized: 13 files, 2 commits.

````
## What changes
One claim (c4038d15) plus one tooling commit (ff0d8b2f).

**Shift + left click on a beltable inventory item puts it in the belt.** In game.exe the inventory grid's click handler `FUN_0048ffe0` (0x49029d) sends C->S 0x63 when shift is held, nothing is in hand, the item's own ItemTypes row is Beltable (`FUN_0062bad0`) and it is on inventory page 0 (`FUN_00628250`); otherwise the click picks the item up as usual. The server's `FUN_0054d520` checks the same two things, asks `FUN_0063c600` for the box (the same function the pickup's belt step uses) and moves the item there; with no box it does nothing (the item stays in the pack, nothing goes into the hand). d2d had no command for it (net-packets.md listed 0x63 auto-belt as "none"), so shift was ignored and the item went into the hand.

Now:
- `rules::belt_box` is the pickup's belt step moved into its own function (`rules::pick_up` calls it; no pickup changes: the #147 oracle still matches), `rules::beltable` is `FUN_0062bad0`, `rules::auto_belt` is `FUN_0054d520`'s page test and move.
- `cmd::AutoBelt{item}` (opcode 0x63); the World applies it with nothing in hand.
- The client sends it for a shift + left click on a Beltable item in the inventory grid (`cursor.cpp`, shift from `SDL_GetModState`); a joined d2d sends game.exe's 0x63 with the host's item id.

Not changed: pickups; the cursor -> belt drop (0x23); belt hotkeys; ctrl-click selling and the inventory-open pickup (client behaviour, not in this PR).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0048ffe0` 0x4901f2 / 0x49029d | ctrl tested first (sell with a store); then flags & 4 (shift), `FUN_0062bad0`, `FUN_00628250` == 0 -> 0x63; else the item to the hand (0x19) |
| `FUN_0054d520` | size 5; the player's stored item (`FUN_00549150`); hand empty (`FUN_0063c1e0`); `FUN_0062bad0`; page 0 (`FUN_00628250`); `FUN_0063c600(inv, item, &box)`, 0 or box < 0 -> return; `FUN_0063ad90` out, `FUN_0063c4f0` into box |
| `FUN_0063c600` | 1x1; the first column whose bottom matches (same code, or the hp / mp / rv groups of `FUN_00628a40`), its lowest free box below the belt's boxes; else, autobelt, the first empty bottom box |

## Proof
`tools/emu/auto_belt.py` runs game.exe's `FUN_0062bad0`, `FUN_00628250` and `FUN_0063c600` in `FUN_0054d520`'s order (no hooks); `test_game --auto-belt` runs `rules::auto_belt` on the same cases; `diff`. 390,000 cases (every belt x bottom row over {-, hp1, mp1, rvs, isc} x {bottom only, full} x 29 items, plus 100,000 random stacked belts, a tenth on the stash / cube page, seed 1): game.exe moves 129,283; before this PR d2d moved none (129,283 mismatches), after 0. Independent re-derivation: VERIFIER_LINE

```sh
cd tools/emu && uv run python auto_belt.py gen 100000 1 > cases.tsv && uv run python auto_belt.py game cases.tsv > g.txt
../../build/tests/test_game --auto-belt cases.tsv | grep -E '^[0-9]+ ' | diff g.txt -        # empty
uv run python belt_pickup.py gen 100000 1 > p.tsv && uv run python belt_pickup.py game p.tsv > pg.txt
../../build/tests/test_game --belt-pickup p.tsv | grep -E '^[0-9]+ ' | diff pg.txt -         # empty: pickups unchanged
```

## Limits (ponytails)
- `FUN_00535610`'s check while trading or talking is not modelled.
- The click sound (`FUN_004b9a00` 0xeb) is not played.
- Presubmit "net roles": the rule lives in the host's World; d2d's own client sends `cmd::AutoBelt`, a joined d2d sends game.exe's 0x63.

## Tests
`test_rules` (six auto_belt cases), `test_protocol` (round trip), `test_combat`, `test_game` green; `pr_guard.py` 0 FAIL 0 WARN.

Agent: Claude Code
````
