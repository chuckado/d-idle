## Why

SKMELEE (docs/research/sk-melee-trace.md) proved its d2d deviations for the player's plain Attack only against a port of d2d's rule, and left D7 (does the weapon's enhanced damage also reach the player's stats 17 / 18) and the shield carrier open. A real d2d build differential is needed before any deviation can become a PR candidate.

## What Changes

- A d2d test mode `tools/emu/sk_melee2_test.cpp` (upstream path tests/test_melee.cpp, scratch checkout only) that runs the REAL `rules::make_fighter` / `rules::player_blow` on the case grid of `tools/emu/sk_melee.py`, and an oracle driver `tools/emu/sk_melee2.py` (gen / game / cmp) that runs game.exe's FUN_0057d9b0, FUN_0057b660 and FUN_0057b420 on the same cases (seeds 7 and 1, 100,000 cases a part).
- D7 and the shield question read in game.exe's statlist code and in d2d.
- Trace doc `docs/research/sk-melee2-trace.md`: per-part differences on the real build, which of D1..D6 are real, bot-finding updates, PR decision, drill recipe, shared-file lines for the merger.
- No Go code, no game, no slot; nothing pushed or posted.
