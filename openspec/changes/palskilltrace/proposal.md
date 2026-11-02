## Why

Work item PALSKILL (Issue #3): trace the melee dispatch of Smite, Zeal, and Sacrifice in game.exe 1.14d vs d2d's skill_built.
no observable behaviour change: docs and offline formula tooling only, no code under internal/** or cmd/**.

## What Changes

- Dedicated trace in `docs/research/pal-smite-zeal-sacrifice-trace.md`
- Cross-reference link in `docs/research/pal-skill-trace.md`
- Formula verification script in `tools/emu/verify_palskills.py` (seed 0x50414c53, 120,000 cases, 0 mismatches)
- No Go code modified
