## Why

Three of seven API-credit decomposition jobs wrote desk-only traces: their `labrun <host> -- bash tools/emu/x.sh` calls were denied by the runner's allow-list (the successful jobs used the `tools/agentrun.sh NAME --ttl N -- labrun <host> -- bash ...` spelling, which no brief told the others about).

## What Changes

`tools/emu/run-on-worker.sh` (one audited entry point: only a bare script name from tools/emu, argument and script-text checks, agentrun ttl plus remote timeout, per-agent scratch), `tools/emu/run-py.sh`, `tools/emu/mpq-table.sh`, a smoke script, an offline test with a stub labrun, and `docs/api-agent-cookbook.md`. The runner side (private ops repo) allows exactly the wrapper and adds a runtime preamble.

## Impact

No spec delta. Overlap (declared): none.
