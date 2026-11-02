## Why

The independent re-verifier (docs/research/pr-missilerecord-reverify.md, branch agent/reverifymissile) found `local/missilediff-record-up` (ec09c73f) and `local/missilepoison-skillrows` (70fa479d) READY WITH CHANGES: the owner-poison add in `monster_shot_stats` did not follow game.exe (FUN_0064a930 only when SrcDamage != 0, scaled by SrcDamage / 128, lengths summed); the skill PR had to name (or fold in) the trap sibling; the record PR text overclaimed its "independent counts". This card does those changes on the private fork, re-runs the re-verifier's own differentials on the new heads, and updates the two PR drafts.

## What Changes

- Private fork (worker host, private remote only): `local/missilediff-record-up` gets the owner-poison commit (1824870b); `local/missilepoison-skillrows` gets the trap fix through a shared helper (71f98145). Both tests extended.
- docs/research/pr-missilerecord-draft.md (copied from agent/missilefix, edited) and docs/research/pr-missilepoison-draft.md (new): PR text held for the user's go.

## Impact

Docs only in this repo. No Go, no live game, no spec, ADR or glossary change. Overlap (declared): card missilefix (in progress, not yet merged) owns docs/research/pr-missilerecord-draft.md; this card edits the same file on top of its text, so merge missilefix first (or take this card's version of the file on conflict).
