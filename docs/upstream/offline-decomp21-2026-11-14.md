# Offline upstream work, 2026-11-14 (card decomp21, TASK-440, Sonnet)

> status: current. Two pipeline items taken to "ready for the verifier" while the worker host was off limits (no emulator, no game). Tags: V run (run and seen), S (an earlier doc of ours), U (unverified). Nothing pushed to GitHub or posted; the fork branches are on the private fork remote.

## Items

| Id | Stage reached | d2d side (V run) | game side | Branch |
|---|---|---|---|---|
| P05 load-gold clamp | S2 done, draft + body, for the verifier | 363 of 539 differ before, 0 after; `test_game` rc 0 | S (decomp5's run of `FUN_0056b180`, 539 cases); one fresh `fl_gold.py game` run open | `fid/load-gold-clamp-r2` `3e679d8f` |
| P03 hit-recovery modes | S2 done, draft + body, for the verifier | block: 17,195 of 34,314 cells differ before, 0 after; restart `R 10 15` -> `R 10 10`; `test_game` and `test_combat` rc 0 | S (HITREC trace: `hr_gh.py` 14,448 cells, `hr_modeswitch.py`); one fresh `block_len.py game` run open | `fid/hitrec-modes-r1` `2cad31d5` |
| P02 warp walk-stall | not started: blocked | none | needs the emulator (FUN_00548a50 walk to a blocked goal) and the headless harness on the worker host | none |

P02 was skipped because both its steps (S-1 re-test, game.exe approach-walk trace and oracle) need the emulator or the worker host's harness. P04 (interrupt) is behind the U of the used-skill lifetime (a live drill), P06 and P07 (monster missile and monster block gates) are the next ones that can be built the same offline way once their game column exists.

## What the finding changes

- Swing-start (P01) is in upstream main (`43eac438`, `379a2916`, `b857effa`, 2026-11-14, V read `git log upstream/main`); P03 is therefore cut from `upstream/main` `f5e11693`, not stacked, and the pipeline row P01 can be closed once the maintainer's merge is confirmed against our branch (U, not checked here).
- The oracle shape: where the game column was already V run by a decomp card (the closed form `hr_gh.py` proved on 14,448 cells; the loader on 539), the d2d column is a fresh V run of d2d's real code (`Fight::set_pmode`, `Fight::hurt_side`, `World::enter`) against that rule. The verifier should make the game column a fresh V run (`fl_gold.py game`, `block_len.py game`; the latter was written in this card from `hr_gh.py`'s run code and has never run: its emulator part is U until then).

## Offline build recipe (not on the worker host)

The private fork was cloned to a work directory on the head box from the private remote and fetched from upstream. The NAS builder container (`/data/deps` staged, game data under `/data/gamedata`) builds a revision streamed with `git archive`: configure with the `build-fork.sh` flags (tests ON), `ninja test_hit_recovery test_game test_combat`, then run the tests inside the container with `D2_MPQ_DIR=/data/gamedata`. One build is about 3 minutes with the cache; the builder's labrun route hung for this card, the ssh path of `build-fork.sh` did not. Counted runs: `test_game --load-gold cases.tsv` against the rule written out; `test_hit_recovery --block-length | grep '^D ' > d2d.tsv; block_len.py closed d2d.tsv`; `test_hit_recovery --restart`.

## Drill recipe (nothing run)

None needed for these two items: both are decided by oracle, not by a live probe. The reachability of P05 (an over-cap save) is U and not claimed; P03's player-visible effect (a Paladin's block recovers 6 ticks, not 3) can be seen with the lab verbs: `debug goto 2`, a Fallen with a high to-hit, `step 1` per tick, read the player mode in `state`.

## Matrix and queue lines

- `docs/upstream/pipeline-2026-11-14.md` rows P03 and P05 and `docs/upstream/pr-queue.md` rows 3a and 3b are updated in this branch.
- docs/decomp-queue.md: one line proposed for the end of the file: `decomp21: P05 load-gold clamp and P03 hit-recovery modes at "ready for the verifier" (fork branches fid/load-gold-clamp-r2, fid/hitrec-modes-r1); P02 blocked on the emulator.`
