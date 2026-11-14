# Draft (local, NOT posted): "game: a saved purse or stash over its cap is 0 on entering a game"

> status: r2, **ready for the independent verifier** (card decomp21, TASK-440, 2026-11-14); not posted, needs the verifier and then the user's go. Branch `fid/load-gold-clamp-r2` on the private fork remote, head `3e679d8f`, two commits on upstream `f5e11693` (oracle `1a57eb4c`, claim `3e679d8f`; the numeric noreply identity, last line `Agent: Claude Code`). Replaces `fid/load-gold-clamp-r1` (one commit on `b6b59ae9`, 143 commits stale, split now as the standard asks). Earlier draft and proof: [../../research/pr-load-gold-draft.md](../../research/pr-load-gold-draft.md), [../../probes/fid/fid-inv-limits.md](../../probes/fid/fid-inv-limits.md). Body: [../bodies/load-gold-clamp-body.md](../bodies/load-gold-clamp-body.md).

PR title: `game: a saved purse or stash over its cap is 0 on entering a game`

## What changes

`World::take_character` keeps the saved purse (`kGold`) and stash gold (`kGoldBank`) as they are. game.exe's loader `FUN_0056b180` restores them through `FUN_0056af80` (`0x56afb4..0x56b00f`) and `FUN_0053ff00`, which sets a stat to 0 when it would pass its cap: the purse's `clvl x 10000` (`FUN_00622e70`), the stash's 2,500,000 (`FUN_00623460`). The claim commit zeroes both in `take_character` under the same two tests and adds a `test_game` block (at the cap kept, one over zeroed).

Two commits: the oracle (`tools/emu/fl_gold.py`, `test_game --load-gold`; it builds alone and prints what the parent does), then the claim (`world/setup.cpp`, the unit block in `test_game`).

## Claims

| Claim | Evidence | Re-check |
|---|---|---|
| game.exe zeroes a loaded purse over `clvl x 10000` and a stash over 2,500,000; equal values are kept | game side: V run of the real `FUN_0056b180` on 539 cases (decomp5, card TASK-203; **not re-run in this card: the emulator image is on the worker host and the host was off limits**, so the game side here is S) | `uv run python fl_gold.py game cases.tsv` |
| d2d's `World::enter` kept an over-cap purse or stash | V run (card decomp21): `test_game --load-gold` on the parent (oracle commit `1a57eb4c`, upstream `f5e11693`, built in the NAS builder with the game data) against the rule above: **363 of 539 cases differ** (the same count as the earlier run on `b6b59ae9`) | build the parent, `test_game --load-gold cases.tsv` |
| after the claim d2d matches | V run: the same 539 cases on `3e679d8f`: **0 differ**; full `test_game` rc 0, prints `OK: a save's gold over its cap is 0 on entering a game` | `test_game` |
| reachability | U: needs an edited or imported save; not claimed in the PR | none |

Cases: `tools/emu/fl_gold.py gen` (11 levels x 7 purse values x 7 stash values: 0, 1, cap-1, cap, cap+1, 2 cap, 2^25-1). The d2d column of this card's run was compared with the rule written out (`gold > level x 10000 -> 0`, `bank > 2,500,000 -> 0`), not with a fresh emulator run: that last step is the one open item.

## Limits and open items

- **Blocked on the worker host's emulator:** one `fl_gold.py game` run on `f5e11693`'s image to make the game column a fresh V run (the file is the one decomp5 ran; only the d2d side was rebuilt).
- The maintainer may not want a loader that zeroes a player's gold: say in the PR that this is game.exe's rule (an over-cap save cannot come from play), ask in review.
- Checks still to run on the fork checkout before posting: `python3 tools/presubmit.py --quick`, `python3 tools/pr_guard.py`, ctest equal to the parent (the four known data-set failures).
