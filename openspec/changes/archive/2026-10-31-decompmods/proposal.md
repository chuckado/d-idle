## Why

The fidelity risk map lists "unique monster mods untraced" and "no spawn-rate measurement for uniques or champions" (docs/research/d2d-fidelity-riskmap.md, mon_specials), and DECOMPMON covered only Lightning / Fire / Cold Enchanted among the specials (docs/research/monster-attack-trace.md). The early Sorceress farms Act 1 Normal packs and bosses, so the champion / unique / minion modifiers decide kill time, XP and deaths. This card (decomp-queue item 13) traces how game.exe 1.14d applies the MonUMod modifiers at spawn and in the fight, compares d2d's implementation with a differential, and lists the bot's assumptions. No observable behaviour change: research, docs and emulator scripts only; no Go code, no game instance, no d2d slot.

## What Changes

- `docs/research/monster-mods-trace.md`: rules with addresses and tags, d2d's implementation table per mod, the deviation table, the bot's model, the drill recipe.
- `tools/emu/umod_spawn.py` (real room population, stat writes), `tools/emu/umod_stats.py` + `tools/emu/umod_stats_harness.cpp` (real MonUMod inits vs d2d's real `rules::boss_stats`, 157,618 cases).
- `docs/decomp-queue.md` item 13 row and result line; `docs/research/README.md`, `docs/INDEX.md` lines; a `docs/research/pr-bundle-queue.md` row (candidate not built); a `docs/BACKLOG.md` line.

owns: docs/research/monster-mods-trace.md, tools/emu/umod_*, openspec/changes/decompmods/**
exclusive: none (offline; no game instance, no d2d slot, no Go changes)
gates: none (research)
