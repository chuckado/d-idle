## Why

Decomp row 19 (SKMANA, from the Sorceress skill matrix, docs/research/sorc-skill-matrix.md section 5): the mana cost was formula-only (S, FUN_0056c160 not re-read), SKMISS found that game.exe pays a missile skill's mana at the action frame while d2d and the bot book it at the command, and the bot's regen and potion model is UNVERIFIED. Mana is what stops the early Sorceress, so every cast of every act depends on it.

## What Changes

- Trace in game.exe 1.14d (disasm + emulator) the player's mana check (FUN_0056c160, client FUN_00647960 / FUN_00647540 / FUN_006440f0), the payment (FUN_0056bfe0) and when the skill start / do run it (FUN_0056f640, FUN_0056f7f0, the mode handlers), mana regen (FUN_005806f0), the potion class bonus (FUN_0062a620), the skill delay (FUN_0056ef90), with V/D/S/I/U tags.
- A pure-function differential of the check and the payment over every Skills.txt row x levels 1..60 x the mana thresholds (game.exe natively vs d2d's real `rules::mana_cost` on the loaded patched tables).
- In this repo: `docs/research/sk-mana-trace.md` (deviations D1..., bot findings B1..., drill recipe for the FORKDBG verbs, matrix cell updates), `tools/emu/sk_mana.py` and `tools/emu/sk_mana_test.cpp` (copies of the oracle and the d2d test mode), decomp-queue row 19, BACKLOG and INDEX lines.
- A local upstream candidate only if a d2d deviation is proven with the harness and worth a PR (nothing pushed or posted).

Overlap (declared, as fixcrowd and fixemptyplan do): docs/decomp-queue.md, docs/BACKLOG.md and docs/INDEX.md are owned by the open card skmiss (decomp-queue row 14 and its BACKLOG and INDEX lines). This card edits them for its row 19 and its own lines only and does not list them in owns.
