## What changes

A knocking hit (Smite, a monster's knockback) starts nothing while the player is in a get-hit, a block or dead. `FUN_00580a70`, the server's knockback switch (flag 8 of the hit chain in `FUN_0057cee0`, `push 0x13` at `0x57d350`), asks `FUN_0057edd0` first (`0x580adc`, new mode 0x13, above 0x11, so the current-mode table at `0x57ee08`); that gate refuses mode 0x13 from DT, GH, BL and DD, and the function then returns without setting the mode. The flag chain returns after the knock call (`0x57d361`), so the same hit does not try a get-hit either. d2d slid a player standing in a get-hit and restarted her hold. One condition in `Fight::hurt_side` fixes it: a get-hit that is not itself a knockback slide takes no knockback.

A refused knock still does its damage, in both programs: callers of `FUN_0057cee0` run the damage function `FUN_0057c6c0` just before it (`0x57d5c7`/`0x57d658` and the other call sites), and the claim changes only the knock block's condition. Only the mode change is refused.

Two commits, oracle first: `tools/emu/mode_switch.py` also asks `FUN_0057edd0` for new mode 19 (`compare-knock`), and `test_hit_recovery --knock-modes` prints what `hurt_side` does with a 1-point knocking hit in each mode; then the claim (one condition in `components/game/fight/tick.cpp`, an assert in `test_hit_recovery`, the finding in `docs/research/re/combat.md`, which replaces #548's "not changed: a knocking hit inside a get-hit" limit). Built on upstream main `1023c4ac`.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0057cee0` `0x57d332`, `0x57d350`, `0x57d356`, `0x57d361` | flag 8 (knock): `push 0x13`, `call FUN_00580a70`, then return; a knocking hit never also tries the get-hit |
| `FUN_00580a70` `0x580adc` | `call FUN_0057edd0; test eax, eax; je 0x580b63`: refused means nothing set, the mode function is not called |
| `FUN_0057edd0` `0x57ee08`, tables `0x57eea8` (index), `0x57ee8c` (targets) | new mode 0x13: current DT, GH, BL, DD -> `0x57edf0` (refused); NU WL RN TN TW S2 S4 allowed; A1 A2 SC TH allowed while the end event is queued; KK the queued test; S1 refused for class 0, S3 for class 5; a current mode above 0x12 (a knockback in flight) allowed (`0x57ee0b`) |
| `FUN_0057c6c0` before `FUN_0057cee0` at every call site | life is taken first, the mode answer comes after |

## Counted differentials (V run: game.exe's own function under the emulator, d2d built and run, same day)

`mode_switch.py game` runs `FUN_0057edd0` for 20 current modes x new modes 4, 9 and 19 x queued 0/1 x classes 0..6. `test_hit_recovery --knock-modes` gives d2d's side: one 1-point knocking hit on a standing player in each mode, `K mode after knock restarted`. `compare-knock` compares them.

| Build | `compare-knock` (cells) | differ | `compare` (#548 cells) | `mode_switch.py selftest` |
|---|---|---|---|---|
| oracle `1d5c6846` (parent behaviour) | 13 | 1 (GH: d2d knocks, game.exe refuses) | 13, 0 differ | 12 of 12 |
| claim `0b0c5346` | 13 | 0 | 13, 0 differ | 12 of 12 |
| the claim's tests on the oracle's `tick.cpp` | 13 | 1 (GH) | | `test_hit_recovery` aborts at the new assert (`test_hit_recovery.cpp:273`) |

Excluded by name: the dead modes (DT, DD) and the swing modes (see the limits). ctest on the claim and on upstream main `1023c4ac`: 321 tests each, the same 2 fail on the data set used (`test_outdoor`, `smoke_d2d`), every other test passes or skips, per-test results identical. `test_hit_recovery`, `test_game`, `test_combat`, `test_rules`, `test_swing_start`: rc 0. clang-tidy-19 on the two changed C++ files: 0 findings; clang-format-19: 0 warnings. `pr_guard.py` 0 FAIL, 0 WARN; `presubmit.py --quick` 0 FAIL, 2 WARN (the build and the gcc check are skipped by `--quick`: no `compile_commands.json`).

## How to re-check

```sh
set -o pipefail
cmake --build build --target test_hit_recovery test_game test_combat
cd tools/emu && uv sync
export D2_MPQ_DIR=<the game data folder, with bin/game.exe>   # or GAMEEXE=<path to game.exe>; EMU=<dir of emu.py> only when run from elsewhere
uv run python mode_switch.py game > game.tsv
../../build/tests/test_hit_recovery --hit-modes | grep '^M ' > d2d.tsv
uv run python mode_switch.py compare game.tsv d2d.tsv            # #548: 13 cells, 0 differ
../../build/tests/test_hit_recovery --knock-modes | grep '^K ' > knock.tsv
uv run python mode_switch.py compare-knock game.tsv knock.tsv    # claim: 13 cells, 0 differ; oracle commit alone: 1 differ (GH)
uv run python mode_switch.py selftest
cd ../.. && python3 tools/pr_guard.py upstream/main && python3 tools/presubmit.py --quick
```

## Honest limits

- What game.exe refuses here is what `FUN_0057edd0` says: DT, GH, BL and DD refuse mode 0x13. This PR does not claim game.exe allows the other modes. `FUN_00580a70` has a second gate, `FUN_0057eec0` (`0x580af6`, the used skill and state 0x36, as in #557), which the oracle does not run: a knock during a cast or swing whose skill's `interrupt` is blank, or during a Leap in flight, is refused in game.exe and d2d still knocks. Not a regression; a follow-up with #557's gate.
- The swing modes (A1, KK, S1, S3, S4) are left out on purpose: there the knock is decided by both gates, and the second one is #557's ("a big hit interrupts a swing or cast only if the skill allows it").
- Not modelled: the `unit + 0x60` branch (a player holding a cursor item is refused every new mode but 0 and 0x11, mode 0x13 included: read in the disassembly); it only refuses more and cannot change the GH cell. The harness sets `+0x60` to 0.
- A knocking hit with no room, and a hit that both knocks and is big in one call: neither touches the claim (from GH the knock is refused before room is asked, and the get-hit is refused by the #548 rule). The one-call case is right by reading but the unit assert uses a 1-point hit.
- The dead cells (DT, DD) are excluded and not traced; d2d's knock block has no dead test, reachable only if a hit lands on a dead player (not traced).
- Textual overlap: this PR's `tests/game/test_hit_recovery.cpp` edit sits next to #557's addition to the same file (both add a block after `--hit-modes`); no semantic overlap, #557 touches the big-hit line and this one the knock line. Whichever merges second may need a rebase of that file.

## Checklist

- [x] A test in `tests/` (`tests/game/test_hit_recovery.cpp`: the `--knock-modes` grid and the GH assert) and an emulator oracle (`tools/emu/mode_switch.py compare-knock`).
- [x] The finding is in `docs/research/re/combat.md`.
- [x] `python3 tools/pr_guard.py upstream/main` passes.

Agent: Claude Code
