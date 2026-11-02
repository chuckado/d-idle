## Why

docs/research/monster-ai-trace.md (first pass) left open: six AIs with no `mon_think` port, the Fallen flee and the Fallen Shaman corpse reach (D9, D11), the `keep` y field (D7), the Fallen command word (D8), the run base speed (D10), and four unexplained mismatch families.

## What Changes

- docs/research/monsterai2-trace.md: the second pass, with tags upgraded to V run where the emulator ran them and the limits stated.
- tools/emu/mai2_*: harness additions that import the first pass's `mai_ais.py` and upstream's `mon_ais.py` without editing either.
- docs/coverage/ledger.tsv and rule-coverage.md: only rows proved, with case counts.

## Impact

Docs and tools only. No Go, spec, ADR or glossary change. No game, no slot. Nothing pushed or posted.
