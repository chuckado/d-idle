## Why

Decomp lane item 9 (docs/decomp-queue.md; user: "start decomp lane", 2026-10-31). DECOMPBELT found three client-side gaps in d2d by reading only (D2 pickup with the inventory open, D3 shift-click to the belt, D4 ctrl-click sell). Decide for each whether it is a proven d2d deviation from game.exe 1.14d and, where it is a world-level rule, build a local upstream PR candidate with a differential proof.

## What Changes

- docs/research/click-handlers-trace.md: Q1 shift-click, Q2 ctrl-click, Q3 inventory-open pickup with addresses; d2d upstream's handling; verdict per gap; bot relevance
- tools/emu/click_auto_belt.py: the oracle (game.exe FUN_0062bad0 / FUN_00628250 / FUN_0063c600 in FUN_0054d520's order, no hooks); upstream copy `tools/emu/auto_belt.py` on the PR branch
- docs/research/pr-click-{proof,draft,evidence-pack}.md: the PR candidate for D3 (cmd::AutoBelt, C->S 0x63), local branch `trace/click` in the PR checkout dir on the worker host, authored by the user's GitHub identity; nothing pushed or posted
- docs/decomp-queue.md item 9, docs/research/pr-bundle-queue.md row, docs/research/README.md and docs/INDEX.md index lines

## Impact

Docs and tooling only (skip_specs). No bot code changed. The upstream PR needs the user's go.
