# Open to contributors

> status: current, 2026-11-01 (card communitywork). The source list for `tools/contrib/board.py`, which renders [work-board.md](work-board.md) and the GitHub issue bodies. How to take an item: [../contributing/agent-work.md](../contributing/agent-work.md).

Only items that need nothing but the game binary of your own legitimate Diablo II LoD 1.14d install and public sources (this repo, upstream `psi29a/d2decomped`) are listed. Live drills, soaks and anything on the maintainer's private lab are NOT open (see the guide). The maintainer keeps this file in step with `docs/decomp-queue.md` and `docs/BACKLOG.md`; where they disagree those two win.

Format (read by the board script): one `## ID: title` section per item, then `- key: value` lines. Keys: `size` (S, M, L), `status` (`ready` = open for a claim now, `proposed` = open but the maintainer may still reshape it, so ask in the issue first; `claimed` and `done` are shown but get no issue; this is the contributor status, not the queue row's), `kind` (decomp, verify), `source` (`queue:N` = row N of docs/decomp-queue.md, `backlog:NAME` = a line in docs/BACKLOG.md), `question`, `start`, `deliverable`, `needs`. Function addresses are tags only (`FUN_xxxxxxxx` as in the repo's trace docs); never paste code.

## SKMELEE: the player's own blow (staff or weapon swing)
- size: M
- status: done
- kind: decomp
- source: queue:20
- question: In game.exe 1.14d, how does the player's plain Attack roll to-hit and physical damage (weapon share, Strength bonus, shield carrier), and does d2d's `player_blow` / `attack_line` give the same numbers? The first levels of every melee class and the Sorceress staff rung run on this.
- start: FUN_00647bc0 (the player's blow, S: docs/research/paladin-act1-recon.md section 6 row 4); docs/research/sorc-skill-matrix.md section 5 row 7; the monster-side twin docs/research/monster-attack-trace.md for the style; tools/emu/hit_rule.py as the oracle template.
- deliverable: docs/research/sk-melee-trace.md in the trace style (rules with address tags, V read / V run / D / S / I / U, deviations D1.., bot findings B1..), plus an oracle script under tools/emu/ if you ran game.exe code; 100,000 cases or exhaustive boundaries for any number you call V.
- needs: your own game.exe 1.14d, Python 3 with unicorn, a checkout of upstream d2decomped for its emulator harness.

## SKRESIST: monster resistance and immunity by difficulty
- size: M
- status: ready
- kind: decomp
- source: queue:21
- question: How does game.exe build a monster's resistances per difficulty (Normal, Nightmare, Hell penalties), when is an element immune, what does an immune hit still do (flinch, chill), and how does Cold Mastery style pierce interact? Compare with d2d's resist build and with the bot's immune threshold.
- start: docs/research/sorc-skill-matrix.md section 5 row 8; docs/research/static-field-chill-trace.md and docs/research/freeze-rules-trace.md (resist handling already traced for chill and freeze, reuse their function tags); docs/data.md for how monster rows are read.
- deliverable: docs/research/sk-resist-trace.md (tags, deviations, bot findings, what the drills should expect) and any oracle script under tools/emu/.
- needs: your own game.exe 1.14d and its MPQ data directory (read only), Python 3 with unicorn, upstream d2decomped checkout.

## PALSKILL: Smite, Zeal and Sacrifice dispatch in game.exe
- size: M
- status: proposed
- kind: decomp
- source: backlog:SKMELEE+
- question: How does game.exe dispatch Smite, Zeal and Sacrifice (hit count, AR and damage bonus, mana and life cost, knockback and stun, what a held button repeats), and which of these does d2d's `skill_built` branch implement versus falling back to a plain attack?
- start: docs/research/paladin-act1-recon.md section 6 rows 1 and 5; the skill bind trace docs/research/skill-bind-trace.md for the packet side (FUN_0054be70, FUN_00647960); upstream d2decomped fight code (public) for what d2d does.
- deliverable: docs/research/pal-skill-trace.md; a static result (which skills fall back in d2d, read from upstream source, tagged S) is a valid first PR on its own.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn if you run code.

## PALAURA: how an aura applies in game.exe
- size: M
- status: proposed
- kind: decomp
- source: backlog:PALAURA
- question: Which function applies an aura state and its stats (Might, Holy Fire, Thorns, Prayer, Defiance, Resist X), how does the magnitude scale per level, how often does it refresh, and does it reach a mercenary? Does d2d skip aura stats ("aurastats and passives skipped" in the old risk map, U)?
- start: docs/research/paladin-act1-recon.md section 6 rows 3 and 11; docs/research/sk-armor-trace.md (state timers and the buff event, same family of functions: FUN_005cac50 is the armor event); docs/research/d2d-fidelity-riskmap.md.
- deliverable: docs/research/pal-aura-trace.md; the live half (d2d probes) is maintainer-only, so end with the exact probe the maintainer should run and the number you expect.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.

## PALDUR: durability loss per hit and per block
- size: S
- status: proposed
- kind: decomp
- source: backlog:PALDUR
- question: In game.exe 1.14d, what is the chance and the slot choice for durability loss when the player hits and when the player is hit or blocks, and does d2d implement the same rule (it has `debug wear` verbs but the loss rule is unknown)?
- start: docs/research/price-trace.md (the repair gate and price are done: FUN_0062e660, FUN_00629930, use them as the "durability" anchor); docs/research/paladin-act1-recon.md section 6 row 7.
- deliverable: docs/research/durability-trace.md, small; rule with tags, d2d comparison from upstream source (S) or a differential (V) if you build one.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.

## SKGROUND: ground skills (Blizzard, Meteor, Hydra, Fire Wall, Thunder Storm, Frozen Orb)
- size: L
- status: proposed
- kind: decomp
- source: queue:22
- question: How many sub-missiles do these skills create, at what ticks and radius, and what is the skill delay (status 8, which d2d does not model)? Pick ONE skill (say Blizzard) and trace it fully; the others are separate PRs.
- start: docs/research/sorc-skill-matrix.md section 5 row 9; docs/research/sk-missile-trace.md and sk-nova-trace.md for the sub-missile timeline style (FUN_0056d400 is the Frost Nova ring spawn).
- deliverable: docs/research/sk-ground-<skill>-trace.md; advanced item, claim it only if you have done a decomp before.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.

## VERIFY-SKMANA: independent re-check of the mana trace
- size: S
- status: ready
- kind: verify
- source: queue:19
- question: Re-derive the mana cost check and payment from game.exe yourself and with your own harness; do NOT start from our script. Our claims: the check is stat 8 >= (mana + lvlmana x (lvl - 1)) << manashift with no minmana floor (FUN_0056c160); the payment is max(formula, minmana x 256) (FUN_0056bfe0), nothing is taken when short and the do ignores it. Say which hold, which hold with limits, which fail.
- start: docs/research/sk-mana-trace.md (the claims, read AFTER you formed your own view of the two functions); tools/emu/sk_mana.py (only to compare numbers at the end).
- deliverable: docs/research/verify-skmana.md: per claim HOLDS / HOLDS WITH LIMITS / FAILS, your own case count and seed, the command to reproduce, differences from the author's numbers.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.

## VERIFY-WAYPOINT: independent re-check of the waypoint refusal claim
- size: S
- status: ready
- kind: verify
- source: queue:3
- question: Our claim: game.exe refuses a waypoint use only within 10 s of the player's last hostile declaration (pcdata offset 0x160, the only writer FUN_0055b720), never in single player, while d2d counts the window from every arrival. Verify the writers and readers of that field yourself and the claimed 59,309 of 132,964 differing cases for the d2d rule.
- start: docs/research/waypoint-trace.md (after your own look); upstream d2decomped waypoint code (public).
- deliverable: docs/research/verify-waypoint.md in the same HOLDS / LIMITS / FAILS format, with the reproduction command.
- needs: your own game.exe 1.14d, upstream d2decomped checkout, Python 3 with unicorn.
