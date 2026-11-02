## Why

The monster missile record candidate (local/missilediff-record) needed an independent verification before any upstream PR, with a harness written from scratch by someone other than the author.

## What Changes

- docs/research/pr-missilerecord-verifier.md: the independent verifier report. Verdict NOT READY: the record claim (D1 brackets, D2 HitShift, D3 row element) holds with 0 differences on head, but the row poison reaches monster_blow up to 62 times too strong. Required changes are listed.
