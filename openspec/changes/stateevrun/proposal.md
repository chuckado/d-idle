# stateevrun: SKSTATEEV second pass with runs

Board item SKSTATEEV (#21). The first pass (`docs/research/sk-state-events-trace.md`) was read-only: every claim V read. This pass runs game.exe's real functions (the swing gate FUN_0057dbf0, the swing resolver FUN_0057d4f0 and the damage applier FUN_0057c6c0 with its resistance pass FUN_0057c1e0) under the emulator on the worker host, with counted cases and seeds, and says which first-pass claims hold, hold with limits or are wrong.

Deliverable: `docs/research/stateevrun-trace.md`, `tools/emu/ser_events.py` and `tools/emu/ser_run.sh`. Docs and emulator scripts only: no Go, no live game.

No spec delta (`skip_specs: true`). No overlap with another card.
