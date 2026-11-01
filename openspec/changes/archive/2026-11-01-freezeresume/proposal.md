## Why

The independent verifier of the Frozen Armor candidate (VERIFYARMORHIT, docs/research/pr-armorhit-verifier.md on its branch) found that game.exe delays a monster's next think after every freeze to the freeze end + MonStats aidel (the state-1 callback 0x57b170) and drops its thinks while frozen (FUN_005a7f80), while d2d lets it act at the end. That affects every d2d freeze, not only Frozen Armor, and no differential or candidate existed.

## What Changes

- On the worker host (not in this repo): local upstream candidate `fix/freeze-resume` (base upstream main 72132b99): the game-side oracle `tools/emu/freeze_resume.py` (FUN_0057b230 / FUN_005a7f80 / 0x57b170 natively), the d2d test mode `test_freeze_resume --freeze-resume` (the real `Fight::strike` freeze and `monster_update`), and one claim commit (the next think at the freeze end + aidel).
- In this repo: docs/research/pr-freezeresume-proof.md (trace, differential before / after per class and aidel, limits, bot implication, fork proposal), pr-freezeresume-draft.md (PR body, limits first), copies of the tooling (tools/emu/freeze_resume.py, tools/emu/freeze_resume_mode.inc), this card.
- Nothing pushed to the public repo or posted upstream; the private fork is not touched (proposal only).

## Impact

Docs only here. Cold-skill results measured on d2d are pessimistic by this effect (monsters act about aidel frames sooner after each freeze), against the optimistic freeze-rule, Frozen Armor and over-hit deviations.
