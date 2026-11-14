# Brief: P03b "a hit never restarts a get-hit" needs a game-side oracle (BLOCKED, for the worker-host operator)

> status: written 2026-11-14 (card prfix2). Nothing here was run: it needs the emulator image on the worker host (`docs/emu-image.md`), and the verifier's rule (upstream AGENTS.md section 5: "a claim whose oracle isn't committed can't be re-checked") forbids posting the claim as "proven" with only a cited table. The block-length half of P03 is a separate PR and does not wait for this.

Branch (private fork remote): `fid/hitrec-norestart-r1`, stacked on `fid/hitrec-modes-r2` (the block PR). Commits on top of it: `d8d35ca9` (the d2d probes `test_hit_recovery --restart` and `--hit-modes`, no game.exe column yet) and `9169146a` (the claim: one more condition in `Fight::hurt_side`, `tick.cpp`). Tags: V run = built and run here; S = another card's run; U = unverified.

## What is already known

- d2d side (V run, NAS builder, card prfix2): `--restart` prints `R 10 15` on the parent and `R 10 10` with the claim; `--hit-modes` (one big hit in each mode 0..18 with a hold running) differs between the two builds in exactly one cell: GH (4) restarted 1 on the parent, 0 with the claim; BL (9) is already refused on both.
- game.exe side (S, HITREC trace Q4.1, `git show agent/decomp20:tools/emu/hr_modeswitch.py`): `FUN_0057edd0` (may the unit enter mode N from its current mode) refuses GH (4) and BL (9) from GH, BL, DT, DD. That script is not in the fork branch, prints a matrix for people (not a diff), and its run output was never committed. The server's get-hit path reaches `FUN_0057edd0` through `FUN_005809d0` (call at 0x57d383), V read only.

## The run the worker-host operator must do (not run yet)

Everything on the worker host, the emulator image, one core, `nice -n 10`, started through `tools/agentrun.sh`. No game, no Battle.net. Idle check first: the standing idle check of the rules.

1. Check the image: `source ~/d2d-lab/emu/env.sh && ls $GAMEEXE` (stop if it is missing; do not rebuild).
2. Check out `fid/hitrec-norestart-r1` of the private fork (a new worktree) and build the d2d side with the builder recipe of `docs/upstream/offline-decomp21-2026-11-14.md`: `ninja test_hit_recovery`.
3. Write `tools/emu/mode_switch.py` in that worktree as a third-from-top commit (oracle before the claim; reorder with the claim last): port of `hr_modeswitch.py`, same hooks (`FUN_005415a0` returns the case's queued end frame, unit +0x60 = 0), with two subcommands:
   - `game`: for every current mode 0..18 (DT NU WL RN GH TN TW A1 A2 BL SC TH KK S1 S2 S3 S4 DD KB), class 3 (Paladin; also print the class-dependent cells), new mode GH (4) and BL (9), with and without an end event queued, call `FUN_0057edd0` (`e.call(0x57edd0, ecx=game, regs={"eax": unit, "ebx": new})`) and print `M cur new queued allowed` per line. It must exit 1 when it printed no line.
   - `compare game.tsv d2d.tsv`: d2d's file is `test_hit_recovery --hit-modes | grep '^M '` (`M mode after restarted`: standing in `mode` with a hold running, one big hit; `after` is the mode right after, `restarted` is 1 when the hold's end moved). d2d "entered GH" is `restarted == 1`, or `after == 4 and mode != 4`; game.exe "allowed" is the `queued = 1` answer for new mode 4. A cell is same when the two agree. Print `N cells, M differ`; exit 1 on M > 0 or N == 0. Cells the d2d rule leaves to other claims (the swing modes A1 KK S1 S3 S4, which d2d never interrupts: P04; DT and DD, a dead unit) are listed and excluded by name, never silently.
4. Run (expected in the brackets):
   ```sh
   set -o pipefail
   source ~/d2d-lab/emu/env.sh
   ./test_hit_recovery --hit-modes | grep '^M ' > d2d.tsv
   nice -n 10 $PY tools/emu/mode_switch.py game > game.tsv     # a few seconds
   $PY tools/emu/mode_switch.py compare game.tsv d2d.tsv        # [the parent build prints the GH-from-GH cell as differing: M 4 -> 4 restarts; the claim build: 0 differ outside the named exclusions]
   ./test_hit_recovery --restart                                # [R 10 15 on the parent, R 10 10 with the claim]
   ```
   Run `compare` against the build of `d8d35ca9` (parent of the claim) and against `9169146a` (the claim): the first must show the GH-from-GH cell (restarted 1), the second none (BL is refused on both builds already).
5. Add the row to `tools/emu/README.md`, fold the commit into the branch (oracle commit before the claim), push to the fork remote the same day.
6. Report: the two `compare` outputs, the commit hashes, and whether `FUN_0057edd0`'s +0x60 branch (a unit with a non-null +0x60 object may enter only mode 0x11) matters for a player (U in the trace).

## What stays open after the run

- Whether a refused big hit still plays the struck cue (flag 0x8000): U (the verifier's note). The claim removes the cue d2d played through `set_pmode(GH)` on the second hit; either keep the claim's wording "no new get-hit" and say the cue is untraced, or trace `FUN_005809d0`'s refusal path.
- The boundary tick (a second hit exactly at the first hold's end) is V read only (event order); a test pins it.
- The verifier's pass 2 is needed for this branch once the oracle exists: the claim is new material to it.
