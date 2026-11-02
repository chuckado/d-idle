## Why

The baseline audit (docs/research/baseline-audit.md, card baselineaudit) found the original-game baselines too thin to gate on: the late file held 3 runs (one cut to its first 64 minutes, 0 deaths in 5 h), the logged G2 (8 h, 4 deaths) and Fresha 4 h (3 deaths) runs were missing, and the Gate 1 XP clause compared a clvl 1-14 run with a clvl 1-7 mean. Fix items 0 and 5 of its list, with no new run: the raw logs are still on the worker host.

## What Changes

- tools/d2dsoak/baselines.json and baselines-early.json: add a14-full (whole 2 h), g2-8h and the Fresha clvl 1-11 chain, keep every old row, add provenance (run_id, bot, backend, date, profile, weak), per-run clvl band tables, and `exclude_from_mean` for rows that overlap or are superseded.
- cmd/d2dsoak: new `bands` command (XP/h per clvl band with the band verdict), `baseline` takes `LOG+LOG` chains, provenance and `-bands`; internal/d2dsoakstat: band maths and `Baselines.Counted`, tested.
- Delta spec specs/soak-gates/spec.md: Gate 1 XP per clvl band, band baseline sample and tolerance (ADDED), Gate 1 deaths and progression (MODIFIED).
- docs/research/baseline-audit.md: appended "After" section with the recomputed counts, spread, per-band baseline and the Gate 1 reruns rescored per band.

Overlap (declared): m2-latebase owns tools/d2dsoak/baselines-late.json (a different file; its run rows would join `bands -baselines` the same way, nothing to change); soak-minutes-virtual owns internal/d2dsoakstat/minute.go and invariants (read only here: `Minute` is used as is).
