## Why

The user asked (2026-11-01) to tune the worker host: builds and checks were suspected of starving the game, and the capacity check (idle+nice >= 25 percent) ignores I/O wait while the old capacity script hung for callers with a short timeout.

## What Changes

- Measurement of where the load comes from (private ops doc `orion-tuning.md`), user-level tools only: a build wrapper with nice and a CPU quota, a user-level ccache and a cached fork build script, a bounded capacity check with one verdict line (all in the private ops repo).
- Public: three lines in docs/process/worker-capacity.md pointing to the new check and saying that sustained I/O wait counts only when a local disk is really busy.

## Impact

Docs and card only in this repo. No spec, ADR or glossary change. Nothing on the worker host was killed, reniced or changed system-wide.
