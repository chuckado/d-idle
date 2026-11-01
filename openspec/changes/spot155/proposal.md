## Why

CodeRabbit's second nitpick on upstream PR #155 (and the independent verifier's limit 1, docs/research/pr-static-proof.md): `test_static_field --static-field` restated the call site's arithmetic instead of calling `Fight::spot`, so a wrong input in the production `kStaticField` block (pierce index, resistance index, floor lookup) would not have changed its diff. User go 2026-11-01 ("drive to spot in 155").

## What Changes

- Upstream, one fixup commit on the PR branch (tests/game/test_static_field.cpp and tools/emu/static_field.py only): the reviewer mode builds a World and casts Static Field through the real `Fight::spot` on two monsters (the case's, and the previous case's max life, life and resistance), prints both lives after the cast, and keeps the inner-function result as a third column; the oracle prints the same three columns from game.exe's own code. Mode name and reviewer command unchanged.
- This repo: dated `posted:` lines in docs/research/pr-static-evidence-pack.md and pr-static-proof.md, the verifier-limit note, a BACKLOG line. No product code, no gate changed.
