## Why

P23 (waypoint) fails in every soak (295 "nav waypoint failed" lines, 0 successes) and the waypoint is the fallback for most failed town portals. The bot and the SME assume a 10 s refusal after a level change and a row model, both UNVERIFIED. This traces game.exe 1.14d's waypoint panel and the C->S 0x49 path, compares d2d and the bot, and classifies the failure lines. no observable behaviour change: research, docs and emulator / log scripts only; no Go code, no game instance.

## What Changes

- `docs/research/waypoint-trace.md`: game.exe's panel and 0x49 chain with addresses, d2d with file:line, the bot's model, the failure-line table by cause, the drill D-wp expectation, findings W1-W4 for the fix queue.
- `tools/emu/wp_refuse.py`: oracle running game.exe's dispatcher, 0x49 handler and hostile setter natively against a copy of d2d's refusal test; `tools/emu/wp_soaklines.py`: read-only soak-log classifier.
- `docs/research/pr-wp-{proof,draft,evidence-pack}.md`: proven d2d deviation (the 10 s window counts from hostility in game.exe, from every arrival in d2d); local PR candidate branch on the worker host, nothing pushed.
- Index lines: docs/decomp-queue.md item 3, docs/research/README.md, docs/INDEX.md.

owns: docs/research/waypoint-trace.md, docs/research/pr-wp-*.md, tools/emu/wp_*.py, openspec/changes/decompwp/**, docs/decomp-queue.md (item 3 row and result line), docs/research/README.md (index line), docs/INDEX.md (decomp lane line)
exclusive: none (offline; no game instance, no d2d slot, no Go changes)
gates: none (research)
