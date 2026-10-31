## Why

P10 (town portal) fails massively in the soaks and PRIMREVIEW (F4, F5) found the bot clicks an empty Tome of Town Portal. The bot fix (FIXPORTAL) needs the game truth: what an empty tome does, what a portal read requires and does in game.exe 1.14d, whether d2d matches, and which share of the failure lines is an empty tome, a timing issue or something else. no observable behaviour change: research, docs and emulator / log scripts only; no Go code, no game instance.

## What Changes

- `docs/research/town-portal-trace.md`: game.exe's 0x20 use-item chain with addresses, d2d's with file:line, the bot's model, the failure-line table by cause, the drill D2 expectation, findings for FIXPORTAL.
- `tools/emu/tome_use.py`: oracle running game.exe's dispatcher and use-item chain natively against a copy of d2d's `read_portal`; `tools/emu/tome_soaklines.py`: read-only soak-log classifier.
- `docs/research/pr-tome-{proof,draft,evidence-pack}.md`: a proven d2d deviation (inventory TP casts SC, drops the use while busy); PR candidate NOT ready (no patch yet).

owns: docs/research/town-portal-trace.md, docs/research/pr-tome-*.md, tools/emu/tome_*.py, openspec/changes/decomptome/**
exclusive: none (offline; no game instance, no d2d slot, no Go changes)
gates: none (research)
