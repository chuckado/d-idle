## Why

The incoming-damage side decides deaths. docs/research/monster-attack-trace.md proved the to-hit, physical roll and get-hit of a monster melee swing equal to game.exe, but not the defender's side after the hit (damage reduction, the magic element, the El1..3 side parts, block / dodge order, leech or burn on the row), nor what the bot's survival model assumes about it.

## What Changes

- docs/research/monsterhit-trace.md: the trace (Q1..Q9), deviations D1..D5, bot findings B1..B5, no PR candidate, dead ends, drill recipe, matrix proposals. This pass ran without game.exe or the emulator, so nothing is V run; every claim is tagged V read, S, I or U.
- tools/emu/mh_rows.py: table-only helper (MonStats / MonLvl rows to hit chance and mean swing, lists El rows), with a synthetic selftest.

## Impact

Docs and a table helper only. No spec, Go, ADR or glossary change; no ledger row (no cases were run).
