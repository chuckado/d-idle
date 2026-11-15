# Pause and resume as player commands

> status: current (uib1, 2026-11-14). Not verified live.

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
