## Why

User decision 2026-11-02: d2d stays the long-term primary target, but the bot resumes on the original game (memory backend under Wine on the worker host) as the ground-truth validation backend. Before any live work, a read-only audit says what the memory backend can do today, what the gates assume, and what to build first.

## What Changes

- docs/research/backend-parity.md: build and test state on this checkout and the worker host, a 26-row feature parity matrix (memory backend vs d2d), the gate tooling gaps with a cost estimate, a 10-minute smoke-run runbook (not run) and an ordered parity backlog.

## Impact

Docs only. No spec, policy, ADR or glossary change. Nothing run live; the game host was busy.

## Overlap (declared)

The deferred cards gameexe-retire-soak and orig-decision-probes point the same way (Game.exe-only criteria, decision probes on the original game). This audit feeds them and replaces neither.
