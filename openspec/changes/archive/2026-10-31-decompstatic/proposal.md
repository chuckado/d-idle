## Why

Decomp-queue item 6 (DECOMPSKILL follow-on, skill-damage-trace.md "Not covered"): Static Field's floor, the life - 1 cap and negative resistance were read but never run, and the cold length / chill application on the hit was never traced. The bot's Sorceress opens elites with Static Field and uses Frost Nova and Ice Bolt early, so d2d must match game.exe there. Research, docs and emulator scripts only; no Go code, no game instance.

## What Changes

- `docs/research/static-field-chill-trace.md`: Q1-Q3 with addresses and tags, the deviation tables, bot findings, the DRILL RECIPE.
- `tools/emu/static_field.py`, `tools/emu/cold_length.py`: differential oracles on game.exe's own code.
- `docs/research/pr-static-proof.md`, `pr-static-draft.md`, `pr-static-evidence-pack.md`: local PR candidate (fix/static-field), not posted.
- `docs/decomp-queue.md` row 6 and result line; `docs/research/README.md` and `docs/INDEX.md` lines.

owns: docs/research/static-field-chill-trace.md, docs/research/pr-static-*.md, tools/emu/static_field.py, tools/emu/cold_length.py, openspec/changes/decompstatic/**
exclusive: none (offline; no game instance, no d2d slot, no Go changes)
gates: none (research)
