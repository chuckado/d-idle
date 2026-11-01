# FIXMOUTHLOOP2: the step-away loop inside one walk

> status: fix in review, live proof pending (recipe at the end). Follows docs/fixmouthloop.md, whose Fighter back-off
> never ran live (DRILLS20).

## The loop (V from the DRILLS20 m31 log, lines 1437-1470, and the code)

Chain: `nav.walk < GoTo < Fighter.move < Fighter.engage`. One outer `Navigator.walk` (internal/nav/walk.go):

1. `walk` start: `n.walkSeq++` (line 74), say to s; `n.gateRewrites = 0`; local `mouthSeq := 0`.
2. Loop top, after observing: `!n.inMouth && mouthSeq != n.walkSeq` is true (0 != s). `mouthSeq = s` (line 125).
3. `n.mouthStep(o)` says the character is inside the arrival or stay radius of an unwanted exit: it returns an away point.
4. The nested `n.walk(...)` (line 130, `oneClick`, 10 s) starts: it bumps `n.walkSeq` to s+1 and resets `gateRewrites`.
5. Its first click passes the mouth gate (mouthgate.go). The gate rewrites up to `gateMaxRewrites` clicks, then holds:
   the nested walk fails with `ErrMouthGate`.
6. Back in the outer walk: the error is neither `ErrLevelChanged` nor a context error, so `continue`.
7. Loop top again: `mouthSeq` (s) != `n.walkSeq` (s+1): the trigger is armed again, step 3 fires, and so on. About 7 ms
   per round (observe plus a held click, no sleep) until the outer `deadline` (2 min): `walk did not finish in 2m0s`.
8. `Fighter.move` only sees that timeout (`ErrTimeout`, not `ErrMouthGate`), so `isMouthHold` is false, no hold is
   noted, and the next engage tick starts a fresh walk: the restart storm of DRILLS19 and DRILLS20.

Variables: `walkSeq` counts walks of the Navigator (also used by stuckmem.go to tell walks apart); `mouthSeq` is the
walk number for which the outer walk last ran its one step-away check (the intent: once per walk); `inMouth` stops the
nested walk from recursing.

## Fix

In the outer walk after the nested step-away returns:

- `mouthSeq = n.walkSeq`: the nested walk's own bump no longer re-arms the trigger. One step-away attempt per outer
  walk, whatever it returns (no restore of `walkSeq`, so walk numbers stay unique for stuckmem).
- `errors.Is(err, ErrMouthGate)`: return it. The caller then sees the hold. Other nested failures (no path, stuck,
  its 10 s timeout) still fall through to the outer walk once, as before.

## Invariants

- a. A step-away that fails with `ErrMouthGate` makes the outer walk return `ErrMouthGate` at once; the step-away runs
  at most once per outer walk.
- b. `Fighter.move` now gets `ErrMouthGate`: `noteMouthHold`, then `mouthHoldActive` skips restarts from that spot for
  `mouthHoldBackoff` (3 s) for 3 holds, then `mouthHoldReset` (30 s). Unchanged constants.
- c. A step-away that works is untouched (it ran once before too: the second trigger found the character clear). The
  gate (mouthgate.go) is not weakened: no click is added, none is allowed that was held.
- d. Wanted exits stay unguarded (`mouthStep` and `gatedExits` skip `Dest == wantExit`); policy PAUSED.

## Tests

internal/nav/fixmouthloop2_test.go: a sim whose clicks always fail with `ErrMouthGate` from inside the bound. Old code:
V (run on the unfixed walk.go): 682213 held clicks until the 20 s test deadline; new code: 1 click and `ErrMouthGate`. A working step-away still
ends the walk normally; a wanted exit gets no step-away. internal/combat: an end-to-end hold through a real
Navigator is not possible without the game; the Fighter side is fixmouthloop_test.go (heldMover).

## Live recipe (not run: no game here)

As DRILLS20, one instance, DW.m3 x3 and DW.m4 once (docs/probes/drills/README.md, DRILLS20 section). PASS needs all of:
1. 0 bounces and 0 level-change errors (gate preserved, DRILLS19 PASS).
2. Step-aways per DW.m3 run: about 20 or fewer (was 129).
3. `walk did not finish in 2m0s`: toward 0 (was 173).
4. The `combat move held by the mouth gate, backing off` debug line appears.
5. The Den cycle no longer burns the whole phase limit: report the cycle time (was the 120 s limit).
