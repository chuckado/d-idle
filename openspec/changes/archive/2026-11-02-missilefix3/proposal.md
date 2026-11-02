## Why

The third independent check of the two private-fork PR candidates (docs/research/pr-missilerecord-reverify.md, "Third check") returned READY WITH CHANGES for both: one optional code gate (SrcDamage -1 on the owner poison) and text overclaims in the PR drafts.

## What Changes

- docs/research/pr-missilerecord-reverify.md: updated with the verifier's third-check section.
- docs/research/pr-missilerecord-draft.md and pr-missilepoison-draft.md: the verifier's overclaim list applied (1,776, the trap row primepoisoncloud x 23 / 28 / 34, the test covers the helper, SrcDamage -1 stated exactly, the "Third check owed" section replaced by a pointer).
- Fork (private remote only): `local/missilediff-record-up` gets one commit (owner poison reads SrcDamage as the unsigned byte, so -1 is 255, game.exe's plain shot path); `local/missilepoison-skillrows` gets one text-only commit (right trap row in the test comment, comment placement). Nothing is posted upstream.
