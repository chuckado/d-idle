# Watch UI redesign, design phase (2026-11-14)

> status: ACCEPTED by the user on 2026-11-14 (twelve answers, H-open-questions.md) and extended the same day with the character screen, instant item actions and the loot rules UX (TASK-419, branch `agent/uidesign2`). Ready to build from the briefs in [briefs/](briefs/README.md). Nothing is built yet; no file under `internal/watch/` changed. Read from `origin/main` at 883770a9 and from pictures of the fixture viewer; no live game, no sandbox. Tags in the detail pages: V verified (how is said there), S secondary, U unverified.

> **Second attempt (2026-11-14).** The user saw the first wave live and rejected the look: the watch page does not fit one screen and has dead space; the inventory is a masonry of cards. The answer is [J-second-attempt.md](J-second-attempt.md): a watch **cockpit** that fills the window (picture sized by the window's height, one rail for Needs you, Orders and the map, one strip for Progress, Journey and Health) and a **character canvas** on one cell size (hero and mercenary in one frame, pack over stash, a real tooltip). Research and the iteration renders: [research/](research/README.md). Review page with before and after (private link): https://claude.ai/artifact/VmFHcDo5Ekq6ErJxhabYx3 ; the two mockups full page: https://claude.ai/artifact/T3BvUP65Rup5fTzpgtTan1 (watch) and https://claude.ai/artifact/STgQna6SVGRrjRYCbGFxfZ (character). Wave 2 of the build is on hold until the user accepts it; revised cards in [briefs/README.md](briefs/README.md). Where this page's "One page" section below shows the Standard layout and the three-zone character screen, J supersedes it.

Clickable mockup (private link, fake data, placeholder shapes): https://claude.ai/artifact/6PFYh9zdx1A8LUA15FYuq8 . The same file is [mockups/watch-orders.html](mockups/watch-orders.html); pictures: [1440](mockups/shots/mockup-1440.png), [900](mockups/shots/mockup-900.png), [390](mockups/shots/mockup-390.png).

Character screen mockup (private link): https://claude.ai/artifact/2svSGZsWxCPXD4qdjmbj2z . File [mockups/character-screen.html](mockups/character-screen.html); pictures: [1440](mockups/shots/character-1440.png), [900](mockups/shots/character-900.png), [390](mockups/shots/character-390.png).

## One page

### The ten biggest findings

1. The same fact is on screen three or four times: life and mana in three tiles, the current order in three, order lines in two logs.
2. Nothing says what needs the player: a broken weapon, a risky area, two alerts and an open question are in four tiles with four looks.
3. Commands are scattered over six places and two pages, with five separate queues and no shared status.
4. The first screen spends its top on the onboarding card and three text tiles; the game picture is cut by the fold at 1440 and absent at 900 and 390.
5. The minimap tile has thirteen debug buttons and three lines of legend above the map.
6. There are two maps that do not know each other, and the one beside the picture cannot take a click.
7. State words disagree, because three tiles each compose their own sentence from different endpoints.
8. Empty states fill whole tiles ("Nothing queued...", "No rule learned yet", "xp/h measuring").
9. "Am I progressing" has no answer in one place; kills and gold per hour are nowhere.
10. Pause, revive mercenary and sell now do not exist as commands, although the bot has most of the machinery.

### The proposed default view (Standard, 1440 px)

```
 d.idle  * FARMING Cold Plains, for level 10, to unlock Blood Raven     life == 86  mana === 105   (2 need you)  Inventory  Skills  ...
+--------------------------------------+--------------------+--------------------+
| game picture                         | minimap  [Layers]  | NEEDS YOU          |
|                                      |                    |  Weapon broken   [Repair now]
|                                      +--------------------+  Go for Blood Raven? [Go now][Not yet]
|                                      | act map            +--------------------+
|                                      |  click a zone      | ORDERS             |
|                                      |  pins, route,      |  [Go to][Town][Repair][Pause][More]
+--------------------------------------+  level bands       |  > Farm Cold Plains  12 min left
| PROGRESS  lvl 7->10 62% 25 min       |                    |  1 Identify all      next
|  xp/h 12.4k ~  chapter 3 of 9  gold/h|                    |  ~ then the plan ...
+--------------------------------------+                    +--------------------+
| journey strip (as built)             |                    | HEALTH o o o o o   |
|                                      |                    | LOG                |
+--------------------------------------+--------------------+--------------------+
```

Three seconds: the sentence (what and why), Needs you, the picture, Orders, Health, Progress. About 16 numbers on the first screen instead of more than 60.

### The Orders model in six lines

1. A player command is `{intent, args}`: go, fight here, town, repair, farm gold, identify, stash, take, gold, equip, trash, gear set, journey go/skip/hold; later pause, sell now, revive mercenary.
2. One Orders panel shows every open command from all five queues with one state set: queued, waiting, active, done, failed, cancelled; plus blocked rows ("Journey blocked: reason").
3. Row order is the bot's real priority ladder: safety banner, player commands, then the plan row. No drag to reorder.
4. **Primary surface: a hybrid (confirmed by the user).** The Orders panel is primary; the act map is its place picker and draws the queue back as pins, route and failed marks. Departures from "map for everything", flagged for overruling: repair, errands, pause, mercenary and gear sets are not spatial and live in the Orders verb row (and on the town node as the same buttons); the local minimap is for watching and takes no commands.
5. Confirm only what cannot be undone or costs currency (trash, charter, revoke all, a too-hard zone inline); everything else gets a five-second Undo.
6. API: new read `GET /api/commands` (`watch.commands/1`, a join in the Hub, no bot change) and a write dispatcher `POST /api/commands` over the existing routes; nothing existing changes shape.

### Presets

Simple (picture + Needs you, Orders, Health, Progress, Journey), **Standard** (the proposed default, above), Stream (big picture, minimap, progress, orders, log; no buttons), Debug (everything that exists today, nothing hidden). Classic is today's layout and stays loadable. Old saved layouts are untouched: no id is removed, new panels are opt-in, and a preset hides old tiles with the editor's own `hidden` flag.

### Decided by the user (2026-11-14)

Hybrid control surface, the minimap takes no commands; a new hero starts on Simple and **Standard is the default with S1** (made safe by per-cell fallbacks to the old tiles); **the queue stays an unlock** with a progression: single orders, then the Loop order (go to an area for X minutes, home, repair, repeat), then the queue; pause walks to town and stops; Sell now and Revive mercenary become commands; old tiles are retired last, on the user's word; the freeze on new panels is lifted; borderless tiles; red-only notifications, off by default; times in the viewer's zone; stash and take stay at the chest.

### The character screen in five lines (I-character-screen.md)

1. One frame, three touching zones: who wears (hero doll, gear set chips, mercenary strip), the pack (grid, belt, gold, inspector), the other container (Stash, Cube, Loot rules tabs). On a phone the pack stays and the upper part switches.
2. Requirements are first-class: every item shows level, strength, dexterity and class needed against the wearer's numbers, for the hero and for the mercenary; unusable items are dimmed with a reason badge; slot highlights come from the server's own rule.
3. One tooltip and one inspector; comparison as facts, never a verdict.
4. Instant actions: equip, take off, identify, pack moves and drop run at once through the item verbs; stash, take and gold are instant at the chest and otherwise queued for the next town visit with a badge; a refusal never looks like success.
5. The mercenary is shown with his own level, life and gear now; giving him items needs fork work (card B9).

### Cards

Watch screen S1 to S8, character screen C1 to C3, loot rules L1, bot and fork B1 to B9: the table is in [G-build-plan.md](G-build-plan.md), the briefs and the wave order in [briefs/README.md](briefs/README.md). First wave: S1, C1, B7, B8, B1, B4.

### What the user still decides

Questions 14 to 25 in [H-open-questions.md](H-open-questions.md) (instant actions in a fight, for friends and in measured runs; the loot rules unlock and place; facts-only comparison; the strict unlock order of Loop and queue; the Loop trigger; replacing the inventory page outright).

## Details

| Page | Content |
|---|---|
| [A-audit.md](A-audit.md) | every tile with a verdict, the noise ranking, what is missing, the command table; pictures in [audit/shots/](audit/shots/) |
| [B-principles-ia.md](B-principles-ia.md) | eight rules, the three-second hierarchy, Needs you rows, Health lights, drill-down map, density tiers, units |
| [C-orders.md](C-orders.md) | the primary-surface decision with the four options, the two maps, intents with V/U backend support, states, rules, API draft, keys, phone |
| [D-layouts.md](D-layouts.md) | presets, positions, the migration table, rollout; drafts in [layouts/](layouts/) (`python3 -I docs/ui/redesign-2026-11-14/layouts/gen.py` rewrites and checks them) |
| [E-visual-system.md](E-visual-system.md) | tokens, components, accessibility, motion |
| [I-character-screen.md](I-character-screen.md) | the character screen: audit of today's page, patterns from other games, the canvas, requirements, comparison, gestures, instant actions, the mercenary, the validity test table |
| [loot-rules-ux.md](loot-rules-ux.md) | the loot rules editor and what it needs from the engine |
| [G-build-plan.md](G-build-plan.md) | what the answers changed, the card table, what not to build |
| [briefs/](briefs/README.md) | one ready-to-send brief per card, the wave and merge order, the review points |
| [H-open-questions.md](H-open-questions.md) | the twelve answered questions, and questions 13 to 25 of the second round |

## What was and was not checked

- V: panel manifests, layout files, `shell.json`, routes and order code were read; the fixture viewer was rendered at 1440, 900 and 390 px with headless Chrome on the worker host; the layout drafts pass the generator's checks; the mockup was rendered at the three widths in the dark theme and looked at once.
- U: nothing ran against a live bot; the mockup's light theme was seen once before a fix and not again; the drafts were not loaded in the real editor (they name panels that do not exist yet); whether the shell scheduler makes one or four `/api/hero` requests; every row marked U in C.3.
- The design adds no polling: `/api/glance` and `/api/commands` replace readers of `/api/uptime`, `/api/why`, `/api/tally`, `/api/alerts`, `/api/campaign` (chip), `/api/orders` and `/api/repair` in the new presets.
