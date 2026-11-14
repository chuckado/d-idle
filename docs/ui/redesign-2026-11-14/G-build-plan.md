# G. Build plan

> status: accepted and ready to build (2026-11-14): the user answered the open questions and lifted the freeze on new /watch panels. The work is cut into cards with one ready-to-send brief each in [briefs/](briefs/README.md); that page has the wave order, the merge order and the review points. This page is the overview and says what changed after the answers. Estimates are agent runs of the usual size and are guesses.

## What the answers changed

| Answer | Change to the plan |
|---|---|
| 4: Standard is the default with S1 | S1 now carries the default flip, the menu and the one-time notice (they were S7). Standard is progressive: every new tile's cell names the old tile that stands in until the new one ships (layout entry key `fallback`, D.1 and D.4), so the page is never missing a function. Today's default stays loadable as Classic |
| 3: a new hero starts on Simple | in S1; Simple keeps the act map as a tile until S5 can open it from "Go to" |
| 5: the queue stays an unlock, with a progression | new card B7 (the Loop order, its controller and ladder rows; C.8). S5 shows Loop and Add to queue by unlock state. `mode: queue` uses the order queue that exists, not `replace: false` |
| 6, 7: pause walks home; Sell now and Revive mercenary | cards B1, B2, B3 as drafted, now with briefs |
| 8: old tiles retired later | new last slice S8, on hold until the user says the new layouts have proven themselves |
| 9: the freeze is lifted | wave 1 can be sent now |
| 10, 11, 12 | borderless look in S3 to S7; red-only notifications, off by default, in S3; viewer-local times in S6 |
| Added scope: character screen | cards C1 (requirements and usability in the view), C2 (the canvas), C3 (gestures and instant feedback); design in I-character-screen.md |
| Added scope: instant actions | card B8 (fast lane over the item verbs); stash and take stay at the chest, queued for town (user decision; the unlock is TASK-420, not built here) |
| Added scope: mercenary gear | card B9 (fork verbs and bot wiring); until then the screen shows him read-only with requirement checks |
| Added scope: loot filter | UX in loot-rules-ux.md; card L1 on hold until the engine (another agent's design) exists |

## Cards

| # | Card | Does | Est. | Risk | Needs |
|---|---|---|---|---|---|
| S1 | Shell foundations and the default flip | `optIn` panels, layout keys `status` and `fallback`, presets Classic, Simple, Standard, Stream, Debug, Standard as default, Simple for a new hero, the menu, tokens | 2 runs | Medium: touches the shell and the default; reviewed before merge | - |
| S2 | `/api/glance` and the status sentence | one cached read for sentence, needs, health, progress; panel `statusline` | 1 | Low | S1 |
| S3 | Needs you, Health, Progress | four panels on `/api/glance`; red-only notifications | 1 to 2 | Low | S2 |
| S4 | Orders, read-only | `/api/commands` join; panel `orders` with Cancel | 2 | Medium: the join across desks | S1 |
| S5 | Orders write path and the act map | dispatcher `POST /api/commands`; verb row; `zone-map` with the zone menu; Loop and queue by unlock state; Undo; keys | 2 to 3 | Medium to high: a second door to the same actions | S4, B4, B7 |
| S6 | Log, compact minimap, phone | panel `log`; a compact mode of `map`; narrow widths | 1 to 2 | Medium: `map` is large | S1 |
| S7 | Polish and access | shared look, contrast, focus, touch, reduced motion, light theme, the `/` palette | 1 | Low | S3, S5, S6 |
| S8 | Retire the old tiles (HOLD) | stubs, `mergedInto`, aliases per D.3; remove dead drawing code | 2 | High: the one slice that removes things | S7 and the user's go |
| C1 | Requirements and usability in the view | `req` and `usable` for hero and mercenary from the equip path's own rule; the 21-case table | 1 to 2 | Medium: requirement modifiers are U | - |
| C2 | The character screen canvas | one frame, three zones, one inspector, requirement display, mercenary strip, stash tabs | 2 to 3 | Medium: a full rework of the inventory page layout | S1, C1 |
| C3 | Gestures and instant feedback | drag preview, double-click, ctrl-click, pending, refused, queued for town | 1 to 2 | Medium | C2, B8 |
| L1 | Loot rules tab (HOLD) | routes over the engine; the editor, presets, preview, import and export, the unlock | 2 to 3 | Medium: depends on another design | C2, the engine |
| B1 | Pause and resume | walk to town, then stop | 1 | Medium: next to safety | - |
| B2 | Sell now | an errand that sells what the player marked | 1 | Low | - |
| B3 | Revive mercenary | a player request with a button state like repair | 1 | Low | - |
| B4 | Until-level for zone orders | `until: level` on `/api/orders` | 1 | Low | - |
| B5 | Mercenary and supplies for Health | Hub sources for two lights | 1 | Low | S2 |
| B6 | Journey blocked on the watch side | the bot's reason as a Needs you row and a blocked row | 1 | Low | S2, S4 |
| B7 | The Loop order and its unlock | loop controller, `/api/orderloop`, ladder rows, events | 2 | Medium: a ladder data change | - |
| B8 | Instant item actions | the fast lane over the item verbs; `instant` in the view; `wait: town` | 2 | High: runs beside the bot on the control channel; reviewed before merge | - |
| B9 | Mercenary item verbs | fork verbs wrapping the mercenary item command; Go client and executor | 2 to 3 | High: fork work, never run live | B8, C1 |

Rules every card keeps are in its brief: the compatibility checklist, both-class tests, no new polling, goldens, pictures looked at, the gates, merge `origin/main` before queueing.

## Not recommended to build

1. **Click on the minimap to walk somewhere.** No order kind for a point, it fights the bot's pathing and safety, and the tile redraws twice a second.
2. **Drag to reorder the queue.** The bot runs commands by its priority ladder; a reorder the bot does not follow would be a lie.
3. **A free-text command box as the main input.** A palette as an accelerator (S7 or later) is fine; as the primary surface it loses the map and is worst on a phone.
4. **Retiring the old tiles now.** Hide them in the new presets; keep them mounted for old layouts. Retire only on the user's word, by the table in D.3.
5. **A settings panel for knobs on /watch.** Knobs are for runs and tuning, not for daily commanding; the Policy page is the player's rule surface.
6. **Acts 2 to 5 map layouts inside this redesign.** It is data work with its own verification (area ids, edges, order names); until then "Go to" falls back to the list.
7. **More numbers on the first screen** (defence, kills per hour, per-skill damage). Each fails the "would you act on it" test today.
8. **Any gear suggestion in Needs you** ("better armor in stash"). The bot never makes gear decisions (`docs/gear-policy.md`); alerts state facts about durability and resistances only.
