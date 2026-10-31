## Why

Decomp lane item 1 (docs/decomp-queue.md; user: "start decomp lane", 2026-10-31). Bot need: P7 (docs/primitives.md) asks whether a ground potion goes to the belt or the pack; the SME says the belt, our pickit assumes the pack (939 "potion already in the pack" skips); the belt model had two bugs (docs/primitives-review.md F7, F8). Trace the belt and pickup rules in game.exe 1.14d offline (the unicorn oracle), compare with d2d and with our bot's model, prove any d2d deviation with a differential, record bot-side findings for the primitives fix queue.

## What Changes

- docs/research/belt-pickup-trace.md: Q1 pickup destination, Q2 belt model, Q3 shift/ctrl clicks with addresses; d2d's version; our bot's model; findings B1-B5 and deviations D1-D4; the answer drill D1 (P7) should expect
- docs/decomp-queue.md (new): the decomp lane's ranked queue and rules
- tools/emu/belt_pickup.py: the oracle (game.exe FUN_0062bad0 / FUN_00628ba0 / FUN_0063c600, no hooks)
- docs/research/pr-belt-{proof,draft,evidence-pack}.md: the PR candidate for D1 (isc/tsc into a same-scroll belt column), local branch `trace/belt` in the PR checkout dir on the worker host (2 commits, authored by the user's GitHub identity); nothing pushed or posted

## Impact

Docs and tooling only (skip_specs). No bot code changed: B1-B5 go to the primitives fix queue. The upstream PR needs the user's go.
