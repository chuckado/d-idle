## Why

docs/research/original-game-lockstep.md (probe P0) asks whether a proxy-DLL import-table time hook can scale the frame gate of a Win32 program under Wine, before any game work. No game is needed to answer it.

## What Changes

- tools/lockstep/: small C probes (stand-in hook DLL, module with its own clock imports, probe exe), build.sh and run.sh for the worker host.
- docs/research/lockstep-p0.md: measured numbers, recipe, verdict, effect on the tier A estimate.

## Impact

Docs and a small tool. No spec, policy, ADR or glossary change. Nothing touches the game, the game prefixes or the d2d-lab slots.
