## Why

The walk-speed trace (docs/research/walk-speed-trace.md, findings B1 to B3) says the bot most likely walks for most of a session: d2d starts every game with the run toggle off and the controller's belief could survive a fresh game. Walking is 68 percent of each kill and running is 1.5x faster.

## What Changes

- B1: verified that every managed game start already clears the belief (d2denv Proc.WaitInGame calls OnInGame = ResetRun; the seam covers restart, chicken re-entry and new game); the d2dinput.Process path now resets too. Tests pin both paths.
- B2: the minute log line carries runBelieved, stam, stamMax, runKeys, runResets and walkingRaw (RunInfo, LogKV) so a soak can show the run share. d2dsoak is not changed.
- B3: the 400 ms stall wait is left as is (documented, test pins that it is clock time).
- docs/fixrun.md holds VERIFIED LIVE vs UNVERIFIED and the live proof recipe.
