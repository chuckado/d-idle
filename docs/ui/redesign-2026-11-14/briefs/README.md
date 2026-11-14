# Worker briefs for the watch UI redesign

> status: ready to send (2026-11-14, after the user's answers in [../H-open-questions.md](../H-open-questions.md)). One file per slice. Each is self-contained: goal, owned paths, steps, acceptance, the card command, the gates. The freeze on new /watch panels is lifted (answer 9). `<worker>` in a command is the worker host's labrun name.

## Order and parallelism (at most six agents at once)

| Wave | Send | Why then |
|---|---|---|
| 1 | **S1**, and next to it **B1, B2, B3, B4, B7** | S1 is the only slice that touches the shell and flips the default; the bot cards own other files |
| review | Fable reviews S1 before it merges | the one place a mistake breaks a stored layout |
| 2 | **S2, S4, S6** (after S1 merged) | independent directories: `glance.go` + `statusline`, `commands.go` + `orders`, `log` + `map` |
| 3 | **S3** (after S2), **S5** (after S4), **B5, B6** (after S2; B6's second half after S4) | they read the endpoints of wave 2 |
| 4 | **S7** (after S3, S5, S6) | look and access over everything |
| hold | **S8** | only on the user's word |

## Merge order

S1 -> B4 -> B2 -> B3 -> B1 -> B7 -> S2 -> S4 -> S6 -> B5 -> B6 -> S3 -> S5 -> S7 -> (S8). Bot cards may merge whenever green; the order only matters where two touch one file: B4 before S5 (`http_answer.go`), S2 before B5 and B6 (`glance.go` setters), S4 before S5 (`commands.go`), B7 before S5's Loop goldens are final. Everyone merges `origin/main` into their branch before queueing.

## What Fable reviews

| At | Looks at |
|---|---|
| S1, before merge | the shell diff; that a stored layout and Classic render the same cells; the fallback rule; the default flip; the lint rule |
| S2 + S3 merged | the first screen at 1440, 900, 390 against B.2; the sentence for every Why code; number count |
| S4 merged | row order equals the priority ladder; state mapping; no mutex held across desks |
| S5 + B7 merged | the flows against the mockup and C.4 to C.8: two clicks to an order, Undo, confirmations, locked and unlocked states of Loop and queue, the dispatcher shares code |
| S7, before the user looks | the whole page on Standard, Simple and a stored layout; light theme; keyboard |
| S8, before merge | every compat fixture; what a Classic user sees change |

## Files

| Brief | Agent / branch | Starts after | Can run next to |
|---|---|---|---|
| [S1](S1-foundations.md) | `uis1` / `agent/uis1` | nothing (first slice) | the bot cards B1, B2, B3, B4, B7 |
| [S2](S2-glance-statusline.md) | `uis2` / `agent/uis2` | S1 merged | S4, S6, the bot cards |
| [S3](S3-needs-health-progress.md) | `uis3` / `agent/uis3` | S2 merged (its goldens are your fixtures) | S4, S5, S6 |
| [S4](S4-orders-readonly.md) | `uis4` / `agent/uis4` | S1 merged | S2, S3, S6, the bot cards |
| [S5](S5-orders-write-zonemap.md) | `uis5` / `agent/uis5` | S4 merged | S3, S6, S7 preparation; B7 may land before or after (the Loop control appears only when `/api/orderloop` says open) |
| [S6](S6-log-map-phone.md) | `uis6` / `agent/uis6` | S1 merged | S2, S3, S4, S5 |
| [S7](S7-polish-a11y.md) | `uis7` / `agent/uis7` | S3, S5 and S6 merged | B cards |
| [S8](S8-retire-old-tiles.md) | `uis8` / `agent/uis8` | S7 merged AND the user's go | nothing that edits `watch.html`, `app.js` or `shell.json` |
| [B1](B1-pause.md) | `botpause` / `agent/botpause` | nothing | everything |
| [B2](B2-sell-now.md) | `botsell` / `agent/botsell` | nothing | everything |
| [B3](B3-revive-merc.md) | `botmerc` / `agent/botmerc` | nothing | everything (B5 reads the same merc state: agree on the file split with its owner through the coordinator) |
| [B4](B4-until-level.md) | `botuntil` / `agent/botuntil` | nothing | everything except S5's edit of `http_answer.go` (two lines there; merge origin/main before queueing) |
| [B5](B5-merc-supplies-source.md) | `botlights` / `agent/botlights` | S2 merged (the setters exist) | S3 to S6, other bot cards |
| [B6](B6-journey-blocked.md) | `botblocked` / `agent/botblocked` | S2 merged; S4 for the second half | other bot cards |
| [B7](B7-loop-controller.md) | `botloop` / `agent/botloop` | nothing | S1 to S4, B1 to B4 (you are the only owner of `data/unlocks/ladder.json` in this wave) |
