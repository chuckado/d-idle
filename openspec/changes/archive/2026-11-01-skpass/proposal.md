## Why

Decomp row 23 (SKPASS, sorc-skill-matrix rank 10): the user asked (2026-11-01) how the Sorceress passives work and whether 1.14d has synergies. Lightning Mastery (FRESH from clvl 26, reqlevel 30), Cold Mastery and Warmth (END table), Telekinesis (prerequisite point) and item +skills (base level vs level) are untraced; the matrix tags their effects S or U. No observable behaviour change: docs and emulator scripts only.

## What Changes

- Trace in game.exe 1.14d how passive skills put their stats on the unit (masteries, Warmth), when the stats are read (missile creation, regen, resistance), how +skills split the base level (`blvl`) from the level with items (`lvl`), and Telekinesis's do.
- List the synergies of 1.14d from the patched Skills.txt (EDmgSymPerCalc, DmgSymPerCalc, ELenSymPerCalc, aura / duration calcs) for the plan's skills and compute the bonus per plan band (tag I).
- Compare d2d and the bot: deviations D1.., bot findings B1.. (not changed), drill recipe, matrix cell updates, shared-file lines for the merger.
- In this repo: `docs/research/sk-passives-trace.md`, `tools/emu/sk_pass_*` scripts. No PR candidate unless the existing differential proves a d2d deviation.
