## Why

The bot hires, revives and equips mercenaries (internal/merc, docs/merc.md) and leans on them as a second damage dealer and a tank (merc-policy spec, board items SKDEFENDERS and the SKMISS family). The rules behind those choices were S or U: hire price and offer level from upstream's notes, merc stats by level never compared, the merc's own life rules unread.

## What Changes

Docs and emulator scripts only (no Go, no live game, no d2d slot): `docs/research/mercstats-trace.md` decomposes game.exe 1.14d vs d2d vs our bot for the Act 1 Rogue archers and the Act 2 Desert Mercenaries: stats by level, hire offer and price, resurrection price, levelling, the AI (think and attack, skill roll), gear, potion rule, resistances as a defender (SKDEFENDERS in part). Differentials against d2d's real C++ rules under `tools/emu/mst_*` (levelup 2,940, hire 108,000, revive 256, think 100,000, attack 100,000, experience 100,000 cases, all 0 differ), deviations D1-D4, bot findings B1-B6, a drill recipe. One ledger row (`monster.ai Hireable`) and the regenerated rule-coverage table.

## Impact

No spec delta (`skip_specs: true`). Shared-file lines (BACKLOG, INDEX, CHANGELOG) are in merge-lines.txt for the merger. Overlap (declared): docs/coverage/ledger.tsv and rule-coverage.md are edited for one row only (the brief asks for it); pets-merc-trace.md is cited, not edited.
