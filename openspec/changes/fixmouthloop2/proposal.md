## Why

DRILLS20 proved fixmouthloop did not work live: the restart storm is count-for-count the DRILLS19 one (129 step-aways per DW.m3 run, 173 `walk did not finish in 2m0s`, 0 back-off lines). The loop is inside one `Navigator.walk`: the nested step-away walk bumps `walkSeq`, fails with `ErrMouthGate`, and the outer loop re-arms the step-away (`mouthSeq != walkSeq`) every ~7 ms until the 2 minute timeout, so `Fighter.move` never sees `ErrMouthGate`. Details: docs/fixmouthloop2.md.

## What Changes

- `internal/nav/walk.go`: after the nested step-away, `mouthSeq = n.walkSeq` (one attempt per walk) and a nested `ErrMouthGate` is returned from the outer walk.
- Tests: held step-away sim (old code 682213 held clicks until the test deadline, new code 1), working step-away, wanted exit.

## Impact

Movement guard only; the gate itself is unchanged (0 bounces must stay). Needs a live DW.m3 x3 and DW.m4 rerun (recipe in the doc).

Overlap (declared): internal/nav/walk.go (two small edits in the step-away block) is main-owned by default and not listed in owns; cards fixmouthloop and drills20 are unmerged and open on the combat Fighter back-off and the drill harness, which this card relies on and does not edit.

## Shared-file lines for the merger
- docs/BACKLOG.md: FIXMOUTH6 (DRILLS20 loop inside one walk): fixed offline in fixmouthloop2 (walk.go); live DW.m3/DW.m4 proof open.
- docs/INDEX.md: docs/fixmouthloop2.md.
