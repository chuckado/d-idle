## Why

Decomp queue row 15 (SKARMOR, sorc-skill-matrix rank 2): the early Sorceress casts Frozen Armor from clvl 5 on every fight and relies on its freeze of melee strikers as her only crowd control, and the bot recasts it on a timer (FIXRECAST). What Frozen, Shiver and Chilling Armor and Energy Shield do in game.exe 1.14d (state, defence, durations, the striker effect and its exceptions, replacement on recast, absorb and mana), how d2d implements them and where it deviates, and what the bot assumes were untraced.

## What Changes

- Trace in game.exe (disasm + emulator) do 18 / do 23, the state creation and group replacement, the aura events 1 / 2 / 3 / 24, with V/D/S/I/U tags.
- Game-side oracle `tools/emu/sk_armor.py`: Frozen Armor's event 2 (FUN_005cac50) run natively, checked against the freeze oracle (freeze_rules.py) and against a model of d2d's `buff_events`.
- In this repo: `docs/research/sk-armor-trace.md` (deviations, bot findings, drill recipe, matrix-cell updates, shared-file lines for the merger), `tools/emu/sk_armor.py`.
- No Go code, no game, no d2d slot. A PR candidate only if the existing differential harness proves a d2d deviation.

## Impact

Docs and one emulator script. Shared files (decomp-queue, BACKLOG, INDEX, CHANGELOG, research README) are not edited: the lines are in the trace's last section.
