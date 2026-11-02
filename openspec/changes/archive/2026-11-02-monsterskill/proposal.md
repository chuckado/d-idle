## Why

Decomposition item: how game.exe 1.14d decides which skill a monster uses and when, for the Act 1 and 2 casters and shooters (Fallen Shaman family, Skeleton and Corrupt Archers, Quill Rats, Sand Raiders, Scarabs, Act 1 uniques), compared with d2d and the bot's threat model. The earlier monster AI, missile and mana traces leave the skill-slot layer, the usability gate, monster mana and delay, and the Corrupt Archer think untraced.

## What Changes

- `docs/research/monsterskill-trace.md`: Q1..Q9 with addresses and tags, deviation list D1..D3, bot findings B1..B5, drill recipe, matrix proposals.
- `tools/emu/msk_*`: table readers and three oracles (slot levels on spawned monsters, monster mana check and payment, the Corrupt Archer think replayed on d2d's `rules::mon_think`).
- One ledger row (`monster.ai CorruptArcher TESTED-0-DIFF`, 3,000 cases) and the regenerated `docs/coverage/rule-coverage.md`.
- Shared-file lines in `merge-lines.txt`. No Go code, no live game, no d2d slot, no PR candidate.

## Overlap (declared)

The Corrupt Archer replay reuses `tools/emu/mai2_ais.py`, `mai2_patch.py` and `mai_ais.py` (read, imported, not edited). The ledger and `rule-coverage.md` were also owned by the monsterai2 card (archived).
