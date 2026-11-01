## Why

The boss-stat candidate (upstream fix/boss-stats, head c048a356) got the verifier's verdict "claims hold, not ready": the draft needed four edits and the det state hashes were never run.

## What Changes

- docs/research/pr-mods-draft-v2.md: the draft with the verifier's four edits (champion life range 1.1-2.1x, traced follow-on callers, the area-level limit, the siblings answer, the residuals).
- docs/research/pr-mods-evidence-addendum.md: the verifier verdict pointer, the corrected impact numbers and the det state hash table (patched d2d vs unpatched base, one hidden slot).
- Old draft gets a status banner; queue gets one new line.

## Impact

Docs only, plus builds and runs on the worker host (scratch dir `~/d2d-lab/pr/modsprep`). No game.exe, nothing opened upstream.
