## Why

Primitive P26 (skill bind and cast) is UNVERIFIED: the bot's `BindSkills` on d2d and its hotkey popup on Game.exe rest on assumptions about how game.exe binds a skill to a hand and resolves a cast's target (docs/primitives.md P26, decomp-queue item 4). This card traces the select packet (C->S 0x3c), the cast packets and their client gates in game.exe 1.14d, compares d2d (upstream main) and our bot, and proves the d2d select deviation with an oracle. No observable behaviour change: research, docs and an emulator script only; no Go code, no game instance.

## What Changes

- `docs/research/skill-bind-trace.md`: Q1-Q5 with addresses and tags, deviations D1-D4, bot findings B1-B5, the answer the drills should expect and a drill recipe for P26.
- `tools/emu/skill_bind.py`: differential of game.exe's own 0x3c handler vs d2d's World select (before) and the proposed rule (after).
- `docs/decomp-queue.md` item 4 row and result line; the `docs/research/README.md` and `docs/INDEX.md` lines; `docs/OWNERS` mirror lines.

owns: docs/research/skill-bind-trace.md, tools/emu/skill_bind*, openspec/changes/decompskillbind/**
exclusive: none (offline; no game instance, no d2d slot, no Go changes)
gates: none (research)
