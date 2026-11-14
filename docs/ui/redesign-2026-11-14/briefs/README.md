# Worker briefs for the watch UI redesign

> status: ready to send (2026-11-14, after the user's answers in [../H-open-questions.md](../H-open-questions.md)), except S8 and L1 which are on HOLD. One file per card. Each is self-contained: goal, owned paths, steps, acceptance, the card command, the gates. The freeze on new /watch panels is lifted (answer 9). `<worker>` in a command is the worker host's labrun name.

Three families: **S** the watch screen (S1 to S8), **C** the character screen (C1 to C3) with **L1** the loot rules tab, **B** bot and fork cards (B1 to B9).

## Order and parallelism (at most six agents at once)

| Wave | Send | Why then |
|---|---|---|
| 1 | **S1**, **C1**, **B7**, **B8**, **B1**, **B4** | S1 is the only slice that touches the shell and flips the default; C1 and the bot cards own other files. B7 and B8 are the long ones, so they start first |
| review | Fable reviews S1 before it merges | the one place a mistake breaks a stored layout |
| 2 | **S2, S4, S6** (after S1), **C2** (after S1 and C1), **B2, B3** | independent directories |
| 3 | **S3** (after S2), **S5** (after S4), **C3** (after C2 and B8), **B5, B6** (after S2), **B9** (after B8 and C1) | they read what wave 2 built |
| 4 | **S7** (after S3, S5, S6) | look and access over the watch screen |
| hold | **L1** (needs the loot-rules engine), **S8** (needs the user's word) | |

## Merge order

S1 -> C1 -> B4 -> B1 -> B2 -> B3 -> B8 -> B7 -> S2 -> S4 -> S6 -> C2 -> B5 -> B6 -> S3 -> S5 -> C3 -> B9 -> S7 -> (L1) -> (S8).

Where the order is forced: C1 before B8 and C2 (`desk.go` view fields; `usable`); B8 before C3 and B9; B4 before S5 (`http_answer.go`); S2 before B5 and B6 (`glance.go` setters); S4 before S5 (`commands.go`); B7 before S5's Loop goldens are final; B7 is the only card that edits `data/unlocks/ladder.json` until L1. Everyone merges `origin/main` into their branch before queueing.

## What Fable reviews

| At | Looks at |
|---|---|
| S1, before merge | the shell diff; that a stored layout and Classic render the same cells; the fallback rule; the default flip; the lint rule |
| C1 merged | one validity function (no second rule); the reason vocabulary; the U list of requirement modifiers |
| S2 + S3 merged | the first screen at 1440, 900, 390 against B.2; the sentence for every Why code; number count |
| S4 merged | row order equals the priority ladder; state mapping; no mutex held across desks |
| B8, before merge | the fast lane against the safety ladder and the channel ownership; what stays an errand; named runs |
| C2 + C3 merged | the character screen against the mockup and I: no dead zones, requirement display on the 21 cases, no rule in the page, refusals never look like success |
| S5 + B7 merged | the Orders flows against the mockup and C.4 to C.8: two clicks to an order, Undo, confirmations, locked and unlocked states of Loop and queue |
| S7, before the user looks | the whole watch page on Standard, Simple and a stored layout; light theme; keyboard |
| B9, L1, S8, before merge | fork commit and tests; the editor against loot-rules-ux.md; every compat fixture |

## Files

| Brief | Agent / branch | Starts after | Can run next to |
|---|---|---|---|
| [S1](S1-foundations.md) | `uis1` / `agent/uis1` | nothing | B1 B2 B3 B4 B7 B8 C1 |
| [S2](S2-glance-statusline.md) | `uis2` / `agent/uis2` | S1 | S4 S6 C1 |
| [S3](S3-needs-health-progress.md) | `uis3` / `agent/uis3` | S2 | S4 S5 S6 |
| [S4](S4-orders-readonly.md) | `uis4` / `agent/uis4` | S1 | S2 S3 S6 |
| [S5](S5-orders-write-zonemap.md) | `uis5` / `agent/uis5` | S4 (B4 and B7 before its goldens are final) | S3 S6 C2 |
| [S6](S6-log-map-phone.md) | `uis6` / `agent/uis6` | S1 | S2 to S5 |
| [S7](S7-polish-a11y.md) | `uis7` / `agent/uis7` | S3 S5 S6 | C3 L1 |
| [S8](S8-retire-old-tiles.md) | `uis8` / `agent/uis8` | S7 and the user's go (HOLD) | - |
| [C1](C1-usable-requirements.md) | `charreq` / `agent/charreq` | nothing | everything |
| [C2](C2-character-canvas.md) | `charui` / `agent/charui` | S1 and C1 | S2 to S7 |
| [C3](C3-gestures-instant-ui.md) | `chargest` / `agent/chargest` | C2 (B8 for the instant states) | S slices |
| [L1](L1-loot-rules-ui.md) | `lootui` / `agent/lootui` | C2 and the loot-rules engine (HOLD) | S slices |
| [B1](B1-pause.md) | `botpause` / `agent/botpause` | nothing | everything |
| [B2](B2-sell-now.md) | `botsell` / `agent/botsell` | nothing | everything |
| [B3](B3-revive-merc.md) | `botmerc` / `agent/botmerc` | nothing | everything |
| [B4](B4-until-level.md) | `botuntil` / `agent/botuntil` | nothing | everything |
| [B5](B5-merc-supplies-source.md) | `botlights` / `agent/botlights` | S2 | everything |
| [B6](B6-journey-blocked.md) | `botblocked` / `agent/botblocked` | S2 (second half S4) | everything |
| [B7](B7-loop-controller.md) | `botloop` / `agent/botloop` | nothing | everything (sole owner of the unlock ladder file in its wave) |
| [B8](B8-instant-actions.md) | `botinstant` / `agent/botinstant` | nothing | everything but C1's edit of `desk.go` (merge origin/main first) |
| [B9](B9-merc-item-verbs.md) | `forkmercverbs` / `agent/forkmercverbs` | B8 and C1 | S slices |
