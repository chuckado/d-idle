## Why

Gate 1 rerun 2 (docs/soaks/gate1r2-20261101.md, Review 2 section 6): at clvl 11 and up 366 of 925 virtual minutes were in the Blood Moor and Cold Plains because the kit gate (internal/bot/kitgate.go, `kitSafeTier` 2 and 3) allows no other field while the kit is missing. At clvl 14 those fields pay 0.5 and 7 XP per kill in the run (real per-kill pay at clvl 15: 0.4 and 0.5, docs/research/monster-stats-trace.md section 4) and 0.39 gold per kill (treasure-trace.md section 4). The gate was written for clvl 1 to 7.

## What Changes

A policy question for the user and the d2-sme agent first (no code until answered): should the kit gate widen its field list with the clvl (Stony Field from clvl 8, Dark Wood, Black Marsh from 10), or stay and let the poverty farm run only for the gold? The answer goes in docs/kitgate.md and, if code, in a follow-up.

## Impact

No spec delta. Overlap (declared): none; depends on tpglitch for the share of gate activations that are real.
