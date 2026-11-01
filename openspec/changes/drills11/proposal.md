## Why

Two merged fixes were proven offline only: fixemptyplan (ea5e2ae, a back-off goal that snaps onto the player's cell) and fixportalmelee (8ac245e, no town portal cast with 3 or more hostiles adjacent). DRILLS11 proves them live on the hidden slot, builds the pre-fix binary as a control, and measures what the fixportalmelee rule rests on (does a hit cut a town portal cast, how long a cast takes).

## What Changes

- Harness only (cmd/d2drill, internal/drill): group `P` with D-emptyplan, D-portalmelee, D-portalmelee.ctl, D-castcut.
- Results in docs/probes/drills (drills11 files, README section DRILLS11), primitives rows, `live:` lines on fixemptyplan and fixportalmelee.
- No product code changes: failures are findings.
