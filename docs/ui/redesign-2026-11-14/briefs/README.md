# Worker briefs for the watch UI redesign

> status: revised 2026-11-14 for the second attempt (wave 2 on hold until the user accepts it); first written after the user's answers in [../H-open-questions.md](../H-open-questions.md)), except S8 and L1 which are on HOLD. One file per card. Each is self-contained: goal, owned paths, steps, acceptance, the card command, the gates. The freeze on new /watch panels is lifted (answer 9). `<worker>` in a command is the worker host's labrun name.

Three families: **S** the watch screen (S1 to S8), **C** the character screen (C1 to C3) with **L1** the loot rules tab, **B** bot and fork cards (B1 to B9).

## Order and parallelism after the second attempt (2026-11-14)

Wave 1 is merged or running (S1, C1, B1, B4, B7, B8). The user rejected the look of the first result; the S and C briefs from S2 on were rewritten for [../J-second-attempt.md](../J-second-attempt.md) and **wave 2 is on hold until the user accepts the second attempt**. The bot cards B2 and B3 (and B1, B4 if still open) do not depend on it and can proceed.

| Wave | Send | Why then |
|---|---|---|
| 2a | **S1b** (the cockpit template and the fit check) | everything visual after it is gated by the fit check |
| review | Fable reviews S1b before it merges | it touches the shell again; old layouts must not move |
| 2b | **S2, S4, S6** (after S1b), **C2** (after S1b and C1) | independent directories |
| 3 | **S3** (after S2), **S5** (after S4, B4, B7), **C3** (after C2; B8 for instant states), **B5, B6** (after S2), **B9** (after B8 and C1) | they read what wave 2b built |
| 4 | **S7** (after S3, S5, S6): the cockpit becomes the default | every slot is filled by then |
| hold | **L1** (needs the loot-rules engine), **S8** (needs the user's word) | |

## Merge order

S1b -> S2 -> S4 -> S6 -> C2 -> B5 -> B6 -> S3 -> S5 -> C3 -> B9 -> S7 -> (L1) -> (S8). Forced: S1b before every S and C slice (slots, fit check); S2 before S3, B5, B6; S4 before S5; C2 before C3; B8 before C3's instant states and B9. Everyone merges `origin/main` before queueing.

## What Fable reviews

| At | Looks at |
|---|---|
| S1b, before merge | the shell diff; every older layout renders the same cells; the slot rules; the fit check really fails on overflow |
| S2 + S3 merged | renders of the cockpit at the four sizes with the band and the strip filled: fit, alignment, wording of the sentence for every Why code |
| S4 + S5 merged | Orders and the act map in the rail at 1366x768 and 1440x900: rows readable, the menu sheet, two clicks to an order, Loop and queue states, the dispatcher shares code |
| C2 + C3 merged | the character canvas at the four sizes against the accepted renders: one unit, no dead zone, the tooltip, the 21 validity cases, refusals never look like success |
| S7, before the default flips | the whole default page, dark and light, with empty and full data; a stored layout untouched |
| B8, B9, L1, S8, before merge | as before |

## Files

| Brief | Agent / branch | State |
|---|---|---|
| [S1](S1-foundations.md) | `uis1` | wave 1 (sent) |
| [S1b](S1b-cockpit-template.md) | `uis1b` | new; first of wave 2 |
| [S2](S2-glance-statusline.md) | `uis2` | rewritten |
| [S3](S3-strip-and-needs.md) | `uis3` | rewritten (was S3-needs-health-progress) |
| [S4](S4-orders-readonly.md) | `uis4` | rewritten |
| [S5](S5-orders-write-zonemap.md) | `uis5` | rewritten |
| [S6](S6-areamap-drawers-phone.md) | `uis6` | rewritten (was S6-log-map-phone) |
| [S7](S7-default-polish-a11y.md) | `uis7` | rewritten (was S7-polish-a11y) |
| [S8](S8-retire-old-tiles.md) | `uis8` | hold, unchanged |
| [C1](C1-usable-requirements.md) | `charreq` | wave 1 (sent), unchanged |
| [C2](C2-character-canvas.md) | `charui` | rewritten |
| [C3](C3-tooltip-menu-gestures.md) | `chargest` | rewritten (was C3-gestures-instant-ui) |
| [L1](L1-loot-rules-ui.md) | `lootui` | hold; the rules list now sits under the stat sheets (J.4) |
| [B1](B1-pause.md) to [B9](B9-merc-item-verbs.md) | as named in each | unchanged; B2, B3 can proceed |
