# FIXMOUTHLOOP: no step-away restart loop after a mouth-gate hold

## Restart path (V, read from the code and the DRILLS19 numbers)
`Fighter.engage` (approach branch, no shot) -> `Fighter.move` -> `Navigator.GoTo` starts a NEW walk (new `walkSeq`) -> the walk's start sees the character inside the mouth radius and runs the step-away -> the gate holds (rewrite cap 3, monsters block the away route, `nav stalled by monsters`) -> the step-away's error is dropped and the main click is gated too -> `ErrMouthGate` -> `move` logs "move failed" at info and sleeps one poll -> next tick the same again. `tr.approaches` ends one target after `MaxApproaches`, so a pack of targets keeps the loop alive: 117 restarts per Den cycle (V DRILLS19, docs/probes/drills/README.md on agent/drills19).

## Design
`Fighter.move` keeps a `mouthHoldState` (position, time, consecutive holds, skipped walks), only for `errors.Is(err, nav.ErrMouthGate)`.
- Debug line "combat move held by the mouth gate, backing off" on the first hold of an episode and on reaching the cap, with the skipped-walk counter.
- Back-off `mouthHoldBackoff` 3 s between restarts for the first `mouthHoldCap` 3 holds at one spot, then `mouthHoldReset` 30 s with no restart. A skipped walk returns after a poll; the fight casts what is in range, and evade's pin detection (position unchanged twice) turns kiting off for `pinHold`, so the fighter holds position and fights what is adjacent.
- The state clears when the character moves more than `pinRadius` (2) from the held spot, after the window, or after a successful walk.
- Derivation: the Den cycle is the 120 s phase limit with 117 restarts, about one per second. At 3 s x 3 plus 30 s holds a 120 s cycle allows about 3 + 4 = 7 walks ("a handful"); per 60 s at most 3 + 1 = 4 (U: constants chosen from that rate, not tuned live).
- Unchanged (V by the existing tests): fixmouth2's evade cap, fixmouth4's gate; other errors (`ErrNoPath`, `ErrStuck`) are not touched, so the wanted-exit and ordinary failed-walk behaviour is as before. Which areas are visited and when the bot leaves is not changed.

## Tests
`go test -p 1 -count=1 -run MouthHold -v ./internal/combat`: ten unreachable Fallen and a mover that always returns `ErrMouthGate`, 60 s: 120 walks without the back-off (check by disabling it), 3 with it.

## Live recipe (not run)
Same as docs/fixmouth4.md, DW.m3 x3 (`D2D_AGENT=<agent> D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh <agent> --ttl 900 -- tools/d2dslot.sh run -n 0 -t 600 -- <built>/d2drill -group F -only DW.m3 -label fixmouthloop-m3N -v`, d2drill built from this branch on the worker host). PASS: `nav: stepping away from a cave mouth` count per run falls from 129 (DRILLS19) to about 20 or fewer, `combat move held by the mouth gate` lines appear, no `walk did not finish in 2m0s` flood, zero mouth bounces still. On FAIL attach the filtered log.
