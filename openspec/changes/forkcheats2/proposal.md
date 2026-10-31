## Why

The drills still report BLOCKED rows that only need test hooks the lab d2d binary lacks: a monster to fight (D5.c, D5.d, H.monster, the fight halves of D5.a and D5.b), a way to hurt the merc (D8.m, D1.P3m), quest records the bot needs in a reward-pending state (D8.c, D8.q, D8.w), an unidentified magic item (D3.i, unidentified sell prices) and a pocket to pin the character in (D6.d). The user's standing approval of 2026-10-31 allows adding cheats or functionality to the private fork when documented and vetted, and keeps them out of upstream pull requests. no observable behaviour change: the bot, the cheat ledger code and the specs are untouched; this card records the new lab verbs and which drills they unblock (docs only).

## What Changes

- The lab d2d binary gets `debug spawn`, `debug merc hurt`, `debug quest list|set|bit|join`, `debug give <code> <quality> [unid]` and `give pack|belt q<n> [unid]`, `debug pin on|off|status` and `debug help` (the patch lives in the private fork, never in this repo).
- docs/cheats.md: the table of d2d debug verbs gains the new rows (arguments, what changes, negative cases).
- docs/forkcheats.md: live results of the second set; docs/probes/drills/README.md and docs/primitives.md: missing-cheats table and BLOCKED notes updated (not rerun here).
