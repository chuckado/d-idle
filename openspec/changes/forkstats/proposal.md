## Why

Pending live questions need player damage and skill-bonus stats that the fork's `dump stats` did not show: SKMELEE2 D7/D8 (do the weapon's and other gear's enhanced damage reach stats 17/18, do stats 21/22 hold the weapon damage range), fixskilllvl's unverified skill-bonus stat ids, and the bot's melee model (SKMELEEWIRE assumes stats 21/22).

## What Changes

- Private fork (not in this repo, GPL): branch `lab2-stats` from `lab2-freezeresume`: `dump stats` gains `gear` (per stat and param sums of the worn gear, charms, socketed items and set bonuses for stats 16..25, 83, 107, 127, 188) and `weapons` (worn weapons' base damage). Read only.
- This repo: docs/forkdbg.md section "dump stats gear and weapons (lab2-stats)" and this card.

## Impact

Docs only here. Soaks and drills that need these stats use the named lab2-stats binary.
