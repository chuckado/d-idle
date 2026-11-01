## Why

The SKRESIST A/B (docs/research/model-ab-skresist.md) left two claims open: (1) Fire / Lightning Mastery as resist pierce (stats 333 / 334), claimed by the Sonnet arm and not by the Opus reference; (2) the whole-point physical roll on the missile path (Sonnet D3), unverified by a second harness.

## What Changes

- docs/research/skabrecheck.md: per claim a verdict (CONFIRMED / REFUTED / UNCOVERED) with real-build counts from game.exe's own code run under the emulator against d2d's rules.
- tools/emu/skab_*: table scan, physical-roll oracle and the d2d-side test source.

## Impact

Docs and offline tools only. No game, no Go, no spec / ADR / glossary change.
