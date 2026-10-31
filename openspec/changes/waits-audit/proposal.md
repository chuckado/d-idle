## Why

FIXOBJ found that `d2dctl.Pause` was a wall sleep: under the fork's `--lockstep` the game only moves when stepped, so a wall wait for the game never saw a tick. The same class of bug can hide anywhere code waits on the wall clock for the game to change. docs/probes/waits-audit.md lists every wait and its class.

## What Changes

- docs/probes/waits-audit.md: the audit table (OK, BROKEN, RISK) of the bot, the d2d wiring and the d2d backends.
- internal/d2dctl: `Now`, the clock deadlines of the verbs read (wall by default); `ApproachNPC` uses it, so its timeout is game time under lockstep (before: a wall deadline that ran about 11x too long in game time).
- internal/d2denv: `NewLateClock` and `LateClock.Install` put the clock's `Pause` and `Now` into d2dctl; `Backend.Start` uses it.
- cmd/d2dshop: installs the same clock after dialing (before: raw wall pauses, so the lab tool froze under lockstep).
- Delta: backend-contract ADDED requirements. d2d only; Game.exe and the shared interfaces are untouched.
