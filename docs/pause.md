# Pause and resume as player commands

> status: current (uib1 2026-11-14; frozen Pause added by card pause, 2026-11-15). Not verified live.

## Behaviour
- `POST /api/pause {"on": true}` asks for a pause. In town, with no town trip running, the runner pauses at once. In the field the hero walks home at the next safe moment (`Runner.safeMoment`, the errand trip's rule: no hostile within the clear radius) with `Town.ToTown`, then the runner enters its paused state. The state is never `paused` while the hero is outside town. A failed walk retries after 30 s.
- Safety (chicken, potions, town portal) is never touched: it ranks above every player command and keeps running while the hero goes home. Item actions the player queued stay queued and run after resume.
- `POST /api/pause {"on": false}` resumes; `goals.Manager.Resume` (Pact, auto-resume) clears it too.
- The request is stored in `<state dir>/playerpause.json`; after a restart the hero goes home and stops again.
- A measured run (`D2BOT_RUN`) never installs the desk: the route answers 503 `off`.

## API (additive)
`GET /api/pause` (viewer) and `POST /api/pause` (answer listener, same gates as the repair POSTs; `on` is required, unknown fields refused):
`{"schema":"watch.pause/1","state":"running|going_home|paused","since":"<RFC3339, absent when running>","reason":"..."}`.
Goldens: `internal/watch/testdata/pause/view-*.json`.

## Not done
The `player_pause` Why code (needs entries in `internal/watch/whyline.json` and `whychange.json`, not owned here). Proposed: say "You asked me to pause, so I am going home first." / change "Resume and I go back to the plan."

## Live checklist (U)
1. Hero in a field, fight on: POST on; the hero finishes the pack, walks home, state `paused`; the game window shows him in town.
2. POST off; he returns to the plan.
3. Pause in town: stops at once. Restart the bot while paused: still paused after the restart.
4. While going home, drop life below the chicken line: the potion or chicken still fires.

## The frozen Pause (card pause, user request 2026-11-15: "a Pause button that freezes the game")

### What Pause means now
`POST /api/pause {"on":true}` freezes the GAME when the backend can (`can_freeze` in the view); the older walk-home pause below remains
for backends that cannot freeze (Game.exe under Wine) and for `{"on":true,"freeze":false}`. `{"on":true,"freeze":true}` is refused with
`error` in the view where the game cannot be frozen. Pause is allowed mid-fight (user's choice); the view then carries `monsters_near`
and the banner says so as plain info.

### Mechanism, and the ones rejected (tags: V verified, S read in code or docs, U unverified live)
- **Chosen: the fork's runtime lockstep switch.** `lockstep on` with no `step` behind it stops the game clock while the main loop keeps
  running; `lockstep off` resumes in the same process (S, docs/d2d-fork.md section 3: "runtime switches", used by the lab for menus; V there
  that `lockstep on` then `step`/`click` works after a load). Why it is the most robust: the control socket keeps answering (`ping`, `clock`,
  `screenshot`), so no channel monitor sees a dead d2d; the video keeps showing the frozen frame; the world (monsters, timers, the hero) stops
  because world time only advances on stepped frames (S, docs/d2d-fork.md); nothing is saved or exited. No new fork verb or branch is needed.
  The bot checks the answer of `clock` after each switch (`internal/d2denv/freeze.go`) and reports a failed switch in the view `error`.
- Capability probe: `clock` answers with `lockstep=0` (free-running d2d) and `lockstep off` is accepted (a no-op there). A lockstep run
  (the lab: the bot steps the game itself), a build without the verb, or a Game.exe backend: no `can_freeze`, no button.
- **Rejected: SIGSTOP/SIGCONT of the d2d process.** The socket goes silent, so `d2dHealth` would read a dead channel and restart a healthy game
  (pingGrace 45 s), the video freezes on its last frame with no way to keep serving it, and the wall clock jumps on SIGCONT (the fork snaps
  deltas over 1000 ms, S docs/d2d-fork.md). It would also need a supervisor in the friends app that does not exist as a seam.
- **Rejected: game speed 0 (`timescale 0`).** Not documented, U whether 0 divides or freezes; it also leaves the frame clock running.
- **Never: save and exit** as a pause.

### Safety wiring (what stands still while frozen)
`Runner.step` returns at once while frozen (`internal/bot/freeze.go`): no poll, no safety, no chicken, no goal step, no stuck check, no zone or
farm clock. The running action is cancelled with `errPause` at the press (the benign cause: no failure count). Per item:
| Item | How |
|---|---|
| campaign stall ladder (30+30 min), stuck watchdog, drought, thought fade, zone park, self-pause auto-resume | all read `e.Clock`; `d2denv.LateClock.SetFrozen` stands the wall clock still, and the thaw also re-arms `campaignAt`, `rearmAt`, `lastProgress`, the stall stage (tests: a clock that does not stand still is safe too) |
| loop watchdog (`runner.LastStep`, ten minutes) | the loop keeps running and beating (it sleeps one tick per step) |
| "d2d channel not answering" | pings still work; the Heartbeat probe (counters) is switched to the lockstep branch while frozen (`runnerRef.frozen` in `d2dHealth`), so a stopped game is not a hang |
| Reset button | `PlayerReset` and `BeginHardReset` call `ForceThaw` first, the state clears even if the thaw call fails |
| Resume from the goal manager | `Runner.Resume` thaws too |
| orders | the commands are accepted and queue; the parked loop runs them after Resume |
| auto-resume | `bot.pauseCapMin` (60, 0 = off), wall minutes; a Needs you info row "Paused for N min" stays 30 minutes (`PlayerPauseView.ended`) |
| ledger | one log line `paused: the game is frozen` and one `resumed` (with `pausedMin`); the loop prints nothing per minute while frozen. Lab soak tools that count minutes from the log read the pair as a gap |
| measured run | `D2BOT_RUN` set: no desk, no freezer: the route answers 503 `off` and no button is drawn. A soak that wants a pause needs a separate decision (then record it as an annotated gap) |
| friends app | it launches the same bot; the button shows only if the shipped d2d answers `clock` free-running and accepts `lockstep off`. U for the friends build: check with the checklist below; if the verb is gated, no button appears, never a dead one |
Not frozen on purpose: the lab sentinel's idle reaper (`internal/sentinel`) watches memory and locks of lab instances and does not apply to a player's hero.

### API additions (all optional fields, older goldens stay valid)
- `GET /api/pause`: `frozen`, `can_freeze`, `cap_min`, `ends_at`, `monsters_near`, `ended {minutes, auto, at}`, `error`. While frozen `state` is `paused`.
- `POST /api/pause {"on":bool,"freeze":bool?}`; `freeze` absent = freeze when able.
- `GET /api/glance`: `paused` (bool) and `pausedSince` (RFC 3339), the band's word `Paused` with "paused since HH:MM".
- Goldens: `internal/watch/testdata/pause/view-canfreeze.json`, `view-frozen.json`.

### UI
`internal/watch/web/pauseui.js` and `pause.css`: a Pause/Resume item at the top of the "..." menu (a page without that menu: the links row),
44 px; a bottom banner "Paused since HH:MM (resumes by itself after N min). Monsters are near." with a 44 px Resume button; `<html data-paused>`
dims the game view with a pause glyph (the frame keeps the frozen picture); Space toggles it when no input, button, link, menu or dialog has the
focus. The Orders panel's existing Pause/Resume verb posts the same route, so it freezes too. The control polls `GET /api/pause` every 3 s and is
absent (never dead) when the route answers 503 or `can_freeze` is false.
Fit fixtures: `tools/uishots.sh --fit` and `--pics` pages `default-frozen`, `default-frozen-light`, `default-pause-menu`, `frozen`, `frozen-light`,
`frozen-menu` (query `pause=frozen|can`, served by `tools/uishots/fixtures_commands.go`): U, not run (no browser on the box that wrote this).

### Tests (offline)
`go test -count=1 -p 1 -run 'Freeze|PlayerPause' ./internal/bot`, `./internal/d2denv`, `-run 'Pause' ./internal/watch ./cmd/d2bot`: freeze and thaw, the
loop touching nothing, stall/stuck on a standing-still and a running clock, the cap, failures, Reset and Resume thaw, two classes (Sorceress, Paladin).

### LIVE CHECKLIST (U, not run; needs a hero on a realtime d2d with the fork build)
1. `GET /api/pause`: `can_freeze:true`; the "..." menu shows Pause.
2. Mid-fight press Pause: within a second the picture stops (dimmed, glyph), the banner says Paused since HH:MM and "Monsters are near", `d2dprobe clock` shows `lockstep=1` and `game_ms` flat for 60 s while `ping` answers, monsters do not move in two screenshots 10 s apart.
3. Leave it 5 minutes: the bot log has one `paused` line and no `stall`, `stuck`, `heartbeat flat`, `stopped answering` or `safety` lines; the video still serves the frozen frame.
4. Resume (button, then again with Space): `lockstep=0`, game_ms advances, the hero continues the same fight without a chicken or a stall warning; Needs you shows "Paused for 5 min".
5. Set `bot.pauseCapMin` to 2: pause; after 2 wall minutes it resumes by itself and the row says the cap did it.
6. Pause, then Reset: the game thaws first and the reset runs.
7. Pause, queue an order: it is accepted and starts only after Resume.
8. Restart the bot while frozen (managed d2d): the new process starts a fresh d2d, not frozen. Attach mode (bot restarts, d2d lives on): U, the new bot reads `lockstep=1` and treats it as a lab run: run `lockstep off` by hand.
9. Friends build: repeat 1; if no button appears check `d2dprobe raw 'lockstep off'` for `unknown verb`.

### Fix 2026-11-16: "could not do that" after a successful freeze (card pausefix)
Root cause (V by test `TestCommandsPauseReplyCarriesOK` failing before the fix): the Orders Pause verb posts to `/api/commands`, which relays `POST /api/pause`; that reply was the bare view without the `ok` flag every command route sends, and the Orders panel treats a reply without `ok` as a refusal and showed its generic line. Now every `POST /api/pause` reply carries `ok` (false with `code: pause_refused` and the reason in `error` when the desk refused), the Orders panel accepts a `watch.pause/1` view, and `pauseui.js` judges an odd answer by reading `GET /api/pause` back. A double press and Resume are idempotent (tested).
