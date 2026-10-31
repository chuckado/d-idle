## Why

The live drills D1 and D2 and the waypoint drills report BLOCKED rows that only need test cheats the lab binary does not have yet (act unlock, a hit on the player, merc hire/kill/revive, an item straight into the pack or belt, an arbitrary durability, an empty belt). The user gave the go on 2026-10-31 to add whatever is needed for testing. no observable behaviour change: the bot, the cheat ledger code and the specs are untouched; this card records the new lab verbs and which drills they unblock (docs only).

## What Changes

- The lab d2d binary gets the test verbs `debug act`, `hit`, `merc hire|kill|revive`, `give <code> pack|belt`, `durability`, `belt clear`, `unequip` (the patch lives in the private fork, never in this repo).
- docs/cheats.md: a section for the d2d debug verbs (kind `d2d-debug`): argument form, what each changes, throwaway offline characters only.
- docs/probes/drills/README.md and docs/primitives.md: the missing-cheats table and the BLOCKED drills updated (not rerun here).
