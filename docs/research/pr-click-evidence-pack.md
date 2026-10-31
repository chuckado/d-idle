# Evidence pack: trace/click (for the user's go / no-go)

> status: current (agent DECOMPCLICK, on Opus, 2026-10-31). One page. Detail: [pr-click-proof.md](pr-click-proof.md); body: [pr-click-draft.md](pr-click-draft.md); trace: [click-handlers-trace.md](click-handlers-trace.md); standard: docs/process/upstream-pr-standard.md. posted: nothing.

**Claim.** A shift + left click on a Beltable item stored in the inventory, nothing in hand, moves it to the belt box a pickup would choose; no box: nothing moves (game.exe client `FUN_0048ffe0` 0x49029d -> C->S 0x63 -> server `FUN_0054d520` -> `FUN_0063c600`). d2d had no such command (shift ignored, item into the hand).

**Proof summary.** game.exe's `FUN_0062bad0`, `FUN_00628250`, `FUN_0063c600` run natively in `FUN_0054d520`'s order in upstream's emulator (`tools/emu/auto_belt.py`, no hooks) against d2d's `rules::auto_belt` (`test_game --auto-belt`), same cases, `diff`. Every link in the chain V (disasm.py read; the rule functions run).

**Numbers.**

| Scenario | Cases | game.exe moves | before (upstream/main 0dc4c62a) | after (c4038d15) |
|---|---|---|---|---|
| grid (8 belts x 625 bottom rows x 2 fills x 29 items, page 0) + 100,000 random (10 % stash / cube), seed 1 | 390,000 | 129,283 | 129,283 | 0 (md5 equal) |
| pickups after moving the belt step into `belt_box` (#147's `belt_pickup.py`, seed 1) | 390,000 | | | 0 |

Moves per item (game.exe): hp1-5 / mp1-5 / rvs / rvl ~9,700 of ~13,100 each (autobelt), hpo 8,147, isc 2,805 (an isc column existed), other Beltable codes ~200 each (only onto their own column), non-Beltable 0, stash / cube page 0.

Their checks on ff0d8b2f: `test_rules`, `test_protocol`, `test_combat` ok, `test_game` rc 0 (18 OK), `pr_guard.py upstream/main` 0 FAIL 0 WARN, `presubmit.py --quick` 0 FAIL 2 WARN (net roles, build skipped: answered in the body). Two commits authored by the user, body `Agent: Claude Code`, no trailer. 13 files (+259 -23).

**Commands (reviewer re-runs on their copy).**

```sh
python3 tools/pr_guard.py && cmake --build build && ctest --test-dir build --output-on-failure
cd tools/emu && uv run python auto_belt.py gen 100000 1 > cases.tsv && uv run python auto_belt.py game cases.tsv > g.txt
../../build/tests/test_game --auto-belt cases.tsv | grep -E '^[0-9]+ ' | diff g.txt -     # empty
```

**What the bot gains.** Little today: on d2d the bot fills the belt with the lab's `belt put` verb or a drag, on game.exe it already shift-clicks (`internal/town/beltfill.go`). With this merged, one input path (shift-click) works on both backends and the lab-only verb can go. Mostly a fidelity / live-play parity fix for human players.

**Residual risks.** `FUN_00535610` (trade / NPC check) not modelled; the click sound not played; the client input path is read, not run (a reviewer clicks it by hand); the oracle composes the three rule calls rather than running `FUN_0054d520` whole (its unit lookups, owner test and cursor test are d2d's `held` / own-items lookup).

**Verifier.** Independent Opus verifier: holds with limits, no counterexample. 150,000 adversarial cases (seed 4242) 0 mismatches, seed 7 390,000 cases 0, pickups seed 9 390,000 cases 0; `FUN_00535610` re-derived (trade / NPC states, not modelled); joined-game optimistic apply noted. Report verbatim in pr-click-proof.md.

**PROOF:** complete (limits stated)

**Recommendation.** Low priority. Bundle it into the next open bundle with room (2 commits, 13 files); not a reason to open a bundle by itself. D2 (inventory-open pickup) and D4 (ctrl-click sell) stay notes: client behaviour, no rules oracle.

**Push and PR commands (not run; need the user's go):** `git -C <click worktree> push fork trace/click` then add it to the open bundle per docs/research/pr-bundle-queue.md.
