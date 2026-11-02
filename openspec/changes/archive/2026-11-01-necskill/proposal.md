## Why

Coverage gap rank 7 (docs/coverage/rule-coverage.md): the Necromancer rows (skills.nec, 30 of 30) had no game.exe comparison, and the class plan (docs/classes/necromancer.md) needs Raise Skeleton, Corpse Explosion, Teeth, Bone Spear and the curses.

## What Changes

- docs/research/nec-skill-trace.md: trace of the Necromancer skills (game.exe vs d2d vs bot), deviations D1..D7, bot findings B1..B4, no new PR candidate, drill recipe, matrix cell text.
- tools/emu/nec_*: calc differential (27,800 cases, 0 differ) and magic missile hit differential (36,600 cases, 10,509 differ before the SKMISS fix, 0 after).
- docs/coverage/ledger.tsv, gaps.tsv, rule-coverage.md (regenerated): skills.nec 30 rows and missiles 114, 192, 193.

## Impact

Docs and emulator tooling only. No Go, spec, policy, ADR or glossary change. Glossary candidates for the coordinator: "corpse (Corpse Explosion base life)", "pettype / petmax".
