## Why

Decomp queue row 22 (SKGROUND, sorc-skill-matrix rank 9): the Blizzard endgame table and any class that uses ground and area skills need the game.exe 1.14d rules for Blizzard, Meteor, Frozen Orb, Thunder Storm, Hydra and Fire Wall (area, ticks, damage per tick, skill delay, duration, mana, how the cast reaches the ground, the hit roll), compared with d2d and with the bot.

## What Changes

- docs/research/sk-ground-trace.md: questions Q1.. with address tags, d2d deviations D1.., bot findings B1.., drill recipe, matrix cell updates, shared-file lines for the merger.
- tools/emu/sk_ground*.py (oracle scripts, own prefix) and, when a real-build comparison is made, the d2d-side probe tools/emu/sk_ground_test.cpp.

## Impact

Docs and offline tooling only. No Go code, no live game, no d2d slot. Nothing pushed or posted upstream.
