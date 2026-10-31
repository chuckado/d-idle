# fixrun: the run toggle after a fresh game (change fixrun)

> status: offline fix, live proof pending (card stays open at merge). Findings: [research/walk-speed-trace.md](research/walk-speed-trace.md) B1..B3. Run model: [d2d-run.md](d2d-run.md).

## B1: belief reset

The trace said `ResetRun` had no caller (a grep for the call `ResetRun(` misses it: it is passed as a method value). Read of the code (V, read): `d2denv.Build` sets `Proc.OnInGame = ctl.ResetRun`, and `Proc.WaitInGame` calls it when `state` reports `screen=ingame`. Every managed game start goes through it: `Cycler.EnterGame` (new game, restart, chicken re-entry via `LeaveGame` with Reenter). So the "never reset" premise of B1 does not hold for the d2d backend as it is on main; the walking share must come from elsewhere (routes, stalls, the policy's Off threshold) until the live proof says otherwise.

Gaps closed: `d2dinput.Process.WaitInGame` (the lifecycle helper) now resets too. Not reset, on purpose: attach mode (`EnterGame` returns early when d2d is already in game): the controller is new and believes walking; if the attached d2d had the toggle on, the first move would switch it off (UNVERIFIED, nothing in the bot can read the truth).

Real toggle semantic: the key `r` TOGGLES (V live, d2d-run.md: `key r` made stamina drain); there is no set-state verb. The true state is not readable: `pmode` is -1 both ways (V live) and `state` has no run field. `state` does carry `walking=`; its meaning is UNVERIFIED, so it is only logged raw (`walkingRaw`). No keypress self-heal: pressing on a wrong guess while really running turns the run off, and stamina cannot separate walking from running in town (no drain there). The divergence is made observable instead.

Tests: `TestResetRunPressesRunAgainOnce` (belief reset, next move presses `r` exactly once, later moves none), `d2denv.TestFreshGameResetsRunBelief` (supervisor path through `Proc.WaitInGame` and Build wiring).

## B2: logging

The bot's `minute` line gains `runBelieved`, `stam`, `stamMax`, `runKeys`, `runResets`, `walkingRaw` (from `Controller.LogKV`, through an optional interface so other backends add nothing). Run share = share of minutes with `runBelieved=true` (a belief, not the truth). `d2dsoak` report and its contract are untouched: no metric added yet, the lines are there to add one.

## B3: the 400 ms stall wait

Left unchanged. It runs on the controller clock (`LateClock` under lockstep, virtual time; `TestMoveToStallWaitIsClockTime`), costs nothing on a moving character and only 400 ms per real stall (a stalled leg pays at most about 1.2 s before ErrMoveFailed). A shorter window would re-issue `cmd move` and kick while d2d is still pathing, and nothing offline can show old and new agree on correctness: the fake does not model pathing latency (UNVERIFIED either way). Revisit only with a live trace of the delay between `cmd move` and the first subtile of progress.

## Live proof recipe (DRILLS lane)

1. Start a d2d game (managed), note `minute` lines: `runBelieved`, `stam`.
2. Move a fixed leg outside town (about 40 subtiles), record ticks per subtile (lockstep tick count between the two positions). Expect about 0.5625 subtiles/tick running, 0.375 walking.
3. Restart the game (`LeaveGame` or chicken), repeat the same leg. Compare ticks per subtile: the same as step 2 means the reset works; 1.5x slower means d2d starts with the toggle in a state the bot does not expect.
4. Also read stamina during the leg: drain while moving outside town means running.
5. Log `walkingRaw` in both states to learn the field's meaning.
