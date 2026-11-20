# C. Orders: one control surface

> status: design accepted by the user's answers of 2026-11-14 (H-open-questions.md), not built. Section 8 (progression) was added after the answers and replaces the earlier "queue free for everyone" assumption. Tags per line: V = the backend exists and was read in the code named, U = unverified or needs bot work. The glossary already has "Player command" (a request the player queues on the watch page; runs right behind safety). "Orders" is the name of the panel; a row in it is a player command. No glossary edit is proposed.

## 1. The decision: which surface is primary

The user's position (2026-11-14, as relayed): the map is the main command input only if the design concludes it should be. It was judged against what the player actually does: watches daily and gives a handful of commands (go to a zone, repair now or farm until repair, sell, stash and identify errands, revive the mercenary, pause, apply a gear set).

| Option | What it is | For | Against | Cost |
|---|---|---|---|---|
| 1. Map-first | Every command starts with a click on the act map; non-spatial commands hang off the town node or the hero marker | The picture of the world is the menu; fun; level bands and waypoints are visible while choosing | One of the eight daily commands is spatial. Repair, identify, pause and gear sets have no place on a map, so they would be hidden behind "click the town". Only Act 1 has nodes (`internal/watch/mapdata.go`: acts 2 to 5 are empty, V), so in Act 2 the primary input is blank. A map needs about 350 by 640 px to be legible; a phone cannot keep that on screen with the game picture. The queue still needs a list somewhere | Map data for four acts before it works at all; a second list anyway |
| 2. Command bar first | A text or button bar ("Go to...", "Repair", "Pause") with a palette | Fast for a keyboard user; tiny; works for every command, spatial or not | Choosing a zone from a text list loses what the map gives for free (where it is, how far, how dangerous, what is next). Weak for a friend's first run | Cheap, but throws away the best part of today's UI |
| 3. Hybrid (picked) | One Orders panel owns the queue, the state and the verbs. The act map is its place picker: clicking a zone is the fastest way to say "where", and the map draws the queue back as pins and a route | Every command has one home and one status list. The map does what a map is good at. Works in acts with no map data (the picker falls back to the list view that exists today). On a phone the panel is a row of four buttons and the map is one tap away | Two tiles show the same orders; they must read one source (section 6 gives them one endpoint) | One new read endpoint, one new panel, a restyle of the existing map card |
| 4. Contextual buttons on tiles only | What exists: Repair on the gear tile, intents on the Journey strip, errands on the inventory page | No new surface | It is the problem the user described: six places, five queues, no shared status | Nothing, and nothing improves |

**Picked: 3, the hybrid, with the contextual buttons of option 4 kept as shortcuts into the same queue** (the gear alert still has "Repair now" on it; pressing it adds a row to Orders). Reason in one line: the queue is the thing that must be unified, a queue is a list, and only one verb needs a map.

The user confirmed the hybrid (answer 1) and that the minimap never takes commands (answer 2).

## 2. The two maps

| | Act map (`orders-map` today, `zone-map` in the new presets) | Local minimap (`map`) |
|---|---|---|
| Data | `/api/orders/map` (`watch.ordersmap/1`): nodes, edges, `status {here, order, hard, cleared}`, `range`, `why`, `waypoint` (V) | `/api/senses`: fog, trail, units, coverage (V) |
| Job | **Command input for place**, and the picture of the queue | **Watching**: where the hero is, what is around, what it explored |
| Sits | Next to the Orders panel | Next to the game picture |
| Click | A zone opens the zone menu (section 4) | Nothing new. It keeps zoom and the layers popover |
| Shows of the queue | Numbered pins on queued zones, a gold ring on the current target, the route from the hero through town to the target, a red mark with the reason on a zone whose order failed | The current zone's explored percent (`/api/hero.explored`, V) and, when the bot reports one, the walk target (U: no target field in `/api/senses` today) |

Why the minimap is not a command input: there is no order kind for a coordinate (V: `orders.go` has `go`, `fight`, `home`), "walk to this spot" would fight the bot's own pathing and safety rules, and the tile refreshes twice a second, which makes a click target move. The one local command that makes sense, "Fight here", is a button in the Orders verb row (the API kind `fight` exists, V; no button today).

The two relate by one mark: the node with `status.here` is the zone the minimap shows. In the Standard preset they sit side by side; in Simple the minimap is the only map on the first screen and the act map opens from the "Go to" button.

What the act map gains (all additive fields on `/api/orders/map`, schema stays 1 with new optional fields, or `watch.ordersmap/2` if a field changes meaning):

| Mark on a node | Source | Status |
|---|---|---|
| Level band as colour (below, in, above the hero's level) | `range` + `hero.clvl` | V today (drawn as text only) |
| Too hard, with the guard's reason | `status.hard`, `hardWhy` | V |
| Waypoint found | `/api/atlas` | V as a separate card; U merged into the node |
| Journey chapter here, and its band | `/api/campaign` `chapters[].area`, `band`, `state` | V as data; join by area name, U |
| Explored percent | `/api/hero.explored` | V for the current zone only; per-zone history U |
| Locked (act not reached, area not orderable) | `orderable` | V |

## 3. The model

A **player command** is `{intent, args}`. The Orders panel shows every open one, in the order the bot will run them, with one state vocabulary.

| Intent | Args | Spatial | Backend today | Needs |
|---|---|---|---|---|
| `go` | `area`, `until: {minutes N} | forever | {level N}` | yes | V `POST /api/orders {kind:"go"}`; minutes and forever V; until level exists in `unlockorderqueue.go` (`OrderQueue.Add ... untilLevel`) but `postOrder` refuses it ("until-level comes later") | wire until-level (small) |
| `fight_here` | `until` | here | V `{kind:"fight"}`; no button | UI only |
| `town` | - | town | V `{kind:"home"}` (an idle goal; U that it always walks home) | live check |
| `repair` | `mode: now | farm` (the server picks from live gold, as today) | no | V `POST /api/repair/request`, `/cancel`, `GET /api/repair` schema 2 `trip` | none |
| `farm_gold` | `on` | no | V `POST /api/farmgold {on}` | none |
| `identify` | `all` or item | no | V `POST /api/inventory/identify`, `buyid`, `cain` | none |
| `stash`, `take`, `gold` | item or amount | no | V `POST /api/inventory/{stash,take,gold}` | none |
| `equip`, `unequip`, `trash`, `mark` | item, slot | no | V invdesk queue | none |
| `gear_set` | set id | no | V `POST /api/inventory/sets/apply`, `/cancel` | none |
| `journey` | `chapter`, `do: go_now | skip | hold | resume` | chapter area | V `POST /api/campaign/intent` | none |
| `sell_now` | - | no | U. Marked items sell on the next town visit; nothing forces the visit. The errand trip (`errandTripStep`) is the mechanism | a small invdesk errand kind |
| `revive_merc` | - | no | U. The merc policy revives in town when gold allows; no player request | a request flag that forces the trip |
| `pause`, `resume` | - | no | Half: `bot.Runner.Pause()` and `safeToPause` exist (V `internal/bot/runner.go`, `loop.go`); no watch route | one desk and two routes; rule for pausing in the field |
| `skill_points` | - | no | none (`docs/skilltrees.md`: no write path) | out of scope |

Never in this model: any gear decision by the bot (`docs/gear-policy.md`). `equip`, `gear_set` and `trash` are the player's own clicks and only that.

### States (one vocabulary for every row)

`queued` (accepted, not started) -> `waiting` (wants to start; the reason says what it waits for: "a safe moment", "the town trip to finish", "20 min cooldown") -> `active` (with a phase: `going`, `in town`, `repairing`, `returning`, `farming 140 of 300`) -> `done` | `failed` (reason, Retry where the backend allows) | `cancelled` (by the player or by the bot, with the reason).

A second kind of row is not a command: **`blocked`** rows state why the plan cannot go on (`journey_blocked` with `goals.Status.JourneyBlocked`, V in the bot; the charter needed for a zone; a pact that forbids). They carry the buttons that unblock them and also appear in `needs-you`.

Mapping from today's words (no backend rename): orders `queued|active|done|cancelled|failed` (V); invdesk `queued|doing|done|failed|cancelled` (`doing` shows as `active`); repair `trip.state` `queued|waiting|going|repairing|returning|farming|done|cancelled|failed` (the middle five show as `waiting` or `active` with the word as the phase).

### Order of rows = the bot's real order

The panel does not invent a priority. It prints the ladder of `docs/inventory-actions.md`:

1. **Safety** (not a row you can cancel): when the bot is fleeing, healing or recovering a corpse, a one-line banner sits on top of the list: "Safety first: retreating to town. Your orders wait."
2. **Player commands**, in the bot's own order: item actions (identify first, then equip and trash), errands (stash, take, gold, Cain, scrolls), repair, then the zone order. Each row carries a "next" mark when it is the one that runs at the next safe moment.
3. **The plan** (one grey row, always last, not cancellable): what the hero does when the list is empty, from `/api/campaign` `decision.why` ("Farming Cold Plains for level 10 to unlock Blood Raven") or the Why line.

Because the order is the bot's, **drag to reorder is not offered**. Zone orders are the only kind that can wait behind one another, and that ability is earned (section 8): at the start an order replaces the running one (`replace: true`, V); later the Loop and then the queue open. The queue is the sequencer that already exists (`internal/watch/unlockorderqueue.go`, `GET/POST /api/orderqueue`, up to three waiting orders, one ladder row per slot, V), not `replace: false`.

## 4. Giving a command

### From the act map (place)

Click or tap a zone. A small menu opens at the node (a bottom sheet on a phone). It replaces today's centred order sheet; the content is the same data.

```
Cold Plains                      clvl 6-12 . waypoint
"open ground, waypoint"          you are level 6: in range
-------------------------------------------------------
[ Farm here ]  30 min . 1 h . until level 8 . until I cancel
[ Go now ]     [ Loop... ]     [ Add to queue ]      <- Loop and queue only once unlocked (section 8)
Journey: Blood Raven is next door (too early, floor 9)   [ Go now anyway ]
Route: Rogue Encampment > Blood Moor > Cold Plains
```

| Node | Menu |
|---|---|
| A field or dungeon, orderable | Farm here (duration chips; default is the last one used), Go now; Loop and Add to queue once unlocked, a one-line teaser before that; the journey line when a chapter lives there |
| Too hard (`status.hard`) | the same, with the guard's reason inline in amber and the button reading "Ask anyway" (the guard keeps the last word, as today) |
| The zone the hero is in | Stay and fight (`fight_here`), Explore the rest (`go` here until explored, U), Back to town |
| Town | Rest in town, Repair, Identify all, Sell now, Revive mercenary: the town errands, because this is where they happen. They are the same buttons as the verb row, not a second implementation |
| Not orderable | name, "not orderable yet", and why if known. No toast |
| Charter needed | the price inline: "Above your range. Charter: 12 tally" and a Pay and go button. No `window.confirm` |

### From the verb row (everything else)

The Orders panel header carries at most five buttons; which five depends on state, so a button is never shown disabled without a reason next to it:

`Go to...` (opens the act map picker, or the list in an act without map data) . `Town` . `Repair` (reads "Farm until repair" when gold is short, exactly as `trip.button.label` says today) . `Pause` (reads `Resume` when paused) . `More` (Identify all, Sell now, Stash gold, Revive mercenary, Apply gear set >, Farm gold).

### From an item tooltip

The inventory tooltip's actions (equip, trash, identify, stash, take, mark) post to the same routes as today. What changes: the row they create is in the Orders list on `/watch` too, with its state, and the inventory page's own queue list becomes the same component reading the same endpoint. Nothing moves off the inventory page.

### From an alert

A `needs-you` row carries the one command that answers it ("Weapon broken, about 180 gold" -> `Repair now`). Pressing it adds the Orders row and the alert collapses to "Repair queued".

## 5. Rules

- **Priority.** Player commands run right behind safety. The panel says so in words when it matters ("waiting for a safe moment: 2 monsters near"), using the existing Why codes (`repair_wait`, `player_items`, `repair_farm`, `journey_blocked`).
- **Confirm only what cannot be taken back or costs currency:** trash an item; pay a charter; revoke all rules; "Go now" into a zone the guard calls too hard (inline, one extra click, not a modal). Nothing else asks.
- **Undo.** Every accepted command shows a toast for five seconds with Undo, which is Cancel on the queued row (cancel from `queued` exists for orders, invdesk and repair, V). After that the row's own Cancel works while the backend allows it: `queued` and `waiting` always; `active` for zone orders, repair (until the smith) and farm legs; never for an item action in `doing` (V `TestDoingCannotBeCancelledAndRestartFailsIt`).
- **Failure.** A failed row stays for 20 minutes with its reason in plain words and, where allowed, Retry (repair has a 20 min cooldown, V: the button then reads "Retry in 14 min"). It also raises one `needs-you` row.
- **Empty.** "No orders. The hero follows the plan: <decision.why>." Under it up to three suggestion chips built from state: the next journey chapter ("Go now: Blood Raven"), "Repair (1 broken)", "Identify 3 items". No chip when nothing applies.
- **Disconnected.** The verb row disables as a whole with one line ("No connection to the hero. Orders are kept and shown when it is back."); the reconnect banners of `docs/onboarding.md` stay as they are.
- **ETA.** Shown only where a real number exists: minutes left of a timed zone order (`goal.leftMin`, V), the journey level meter (`meter.etaSeconds`, V), the repair farm leg ("140 of 300 gold", V). A repair trip shows its phase, not a guessed time.

## 6. API (additive)

Nothing existing changes shape. `GET /api/orders`, `POST /api/orders`, `/api/repair/*`, `/api/inventory/*`, `/api/farmgold`, `/api/campaign/intent` stay and keep their goldens.

**New read endpoint `GET /api/commands`** (`schema: "watch.commands/1"`): the Hub joins the five sources it already holds (orders, the repair desk view, the invdesk queue, the gold-farm desk, the campaign view) into one list. No bot change (V that all five are bound in `internal/watch`; U the cost of the join, which must stay a cached read like `/api/hero`).

```json
{
  "schema": "watch.commands/1",
  "seq": 41,
  "safety": { "active": false, "text": "" },
  "paused": false,
  "rows": [
    { "id": "inv:44", "intent": "identify", "args": {"all": true}, "text": "Identify all (3 items)",
      "source": "inventory", "state": "waiting", "phase": "", "next": true,
      "why": {"code": "player_items", "text": "waiting for a safe moment"},
      "progress": null, "etaSeconds": null, "can": {"cancel": true, "retry": false}, "givenAt": 1731600000 },
    { "id": "repair:1", "intent": "repair", "args": {"mode": "farm"}, "text": "Farm until repair",
      "source": "gear", "state": "active", "phase": "farming",
      "why": {"code": "repair_farm", "text": "farming for repair: 140 of 300"},
      "progress": {"pct": 0.47, "label": "140 of 300 gold"}, "etaSeconds": null,
      "can": {"cancel": true, "retry": false}, "givenAt": 1731599000 },
    { "id": "order:12", "intent": "go", "args": {"area": "Cold Plains", "until": {"minutes": 30}},
      "text": "Farm Cold Plains for 30 min", "source": "map", "state": "queued",
      "can": {"cancel": true, "retry": false}, "givenAt": 1731598000 }
  ],
  "blocked": [
    { "id": "journey:blood_raven", "code": "journey_blocked", "text": "Journey blocked: Blood Raven needs level 9 (you are 7)",
      "actions": ["go_now", "hold"] }
  ],
  "plan": { "text": "Farming Cold Plains for level 10 to unlock Blood Raven", "etaSeconds": 1500 },
  "verbs": [
    { "intent": "repair", "label": "Farm until repair", "enabled": true, "reason": "need 160 more gold" },
    { "intent": "pause", "label": "Pause", "enabled": true, "reason": "" }
  ],
  "limits": { "zoneOrders": 5, "itemActions": 8 }
}
```

Times in the reply are epoch seconds; the page prints them in the viewer's local time (the project reports in EST; the page must not print UTC).

**New write endpoint `POST /api/commands`** on the answer listener, same gates as today (origin list, token, JSON, small body, unknown fields refused): `{"intent": "go", "args": {...}, "mode": "now" | "queue"}` (`queue` goes to the existing order queue and answers its own `locked` code while the slot rows are closed; intent `loop` goes to the loop controller). In v1 it is a dispatcher: it validates and calls the same Hub and desk methods the existing routes call, so there is one behaviour and two doors. `POST /api/commands/cancel {"id": "repair:1"}` does the same for cancel. Error codes are the existing ones (`charter-needed`, `active`, `cooldown`, `undamaged`, `pact`, `full`).

Goldens to add (fixtures are only added, never edited): `testdata/commands/view-{empty,three-one-failed,blocked,paused,safety}-{sorceress,paladin}.json`, a compat fixture `testdata/compat/commands-v1.json` once shipped, and the node DOM goldens of the `orders` panel for the same five states.

## 7. Keyboard and phone

| Key | Does |
|---|---|
| `g` | Go to... (opens the act map picker, focus on the hero's zone; arrows walk the graph, Enter opens the zone menu) |
| `t` | Town |
| `r` | Repair (whatever the button reads) |
| `p` | Pause or resume |
| `i` | Identify all |
| `o` | Focus the Orders list; up and down select a row; `Delete` cancels it; `Enter` opens its detail |
| `z` | Undo while the toast is up |
| `/` | Command palette: type "cold" or "rep" (the option 2 idea, kept as an accelerator, last slice) |
| `?` | The key list |
| `Esc` | Close the menu, sheet or palette |

Keys are off while a text field has focus and in overlay mode. Every key has a visible button; none is the only way.

Phone (390 px): the status line and the game picture stay on top; under the picture one row of four 44 px buttons (`Go to`, `Town`, `Repair`, `More`); under it the Orders list, one line per row (text, state chip, a cancel target at the end). `Go to` opens the act map as a full-height sheet; a zone tap swaps the sheet content to the zone menu. The bottom tab rail stays (Watch, Orders, Hero, Log).

## 8. Progression: single orders, then the Loop, then the queue (user answer 5)

The user's words: "early game you can set up rough automation by go to area for X number of minutes, come back, repair, repeat; but early you have to tell it to go to a zone, watch for it to need repairs, and THEN unlock the ability to queue these actions."

### Tiers

| Tier | The player can | Opens when | Ladder row |
|---|---|---|---|
| 0, from the start | Give one zone order at a time (go for N minutes or until cancelled, fight here, town). Repair now or Farm until repair by hand: the player watches the Gear light and the Needs you row. A new zone order replaces the running one | always | - |
| 1, the Loop | One Loop order: go to an area for X minutes, come home, repair, go back, repeat | the first full manual cycle: a zone order the player gave ended `done` or was running, and a repair the player asked for ended `done` | new row `orderloop` |
| 2, the queue | Add to queue: up to three orders wait behind the running one (the existing sequencer) | three Loop rounds finished, or the present triggers as alternatives | existing rows `orderqueue`, `orderqueue2`, `orderqueue3` |

Automatic repair by policy (`repair_break`, `repair_hour`, ... , tally rows, V in `data/unlocks/ladder.json`) is a separate ladder and stays as it is; a Loop repairs because the player asked for the loop, whatever the policy row says.

### The Loop order (new kind)

| Field | Values | Default |
|---|---|---|
| `area` | an orderable area (the same list as `go`) | - |
| `minutes` | 5 to 240 per round | 20 |
| `repair` | `damaged` (repair in town only when something worn is under the warn line or broken) or `always` | `damaged` |
| `rounds` | 1 to 20, or `forever` | 5 |

One round: (1) `go` to `area` for `minutes` (the normal order: the charter, the route guard and the 5-open limit apply unchanged); (2) when it ends, come home: a repair request when `repair` says so (the repair trip of `docs/repair.md`, including its safe moment and one farm leg when gold is short), otherwise a plain `home` order; the town chores run as on any visit; (3) back out for the next round.

The loop stops, with the reason on its row and in Needs you, when: the rounds are done (`done`); the player cancels it or gives a zone order with Go now (`cancelled`); the hero dies; the guard refuses the area on arrival; the repair fails or stays unpayable after its farm leg; the Pact or low-life halt is on (the same `Halt` hook as the queue); the bot stops or pauses (`paused`, resumes at the start of the round). Safety and player commands keep their place on the ladder: an identify or a repair-now given during a loop runs at the next safe moment and the round goes on.

Row text: "Loop: Cold Plains 20 min, repair, repeat (round 3 of 5)", state `active`, phase `farming`, `going home`, `repairing`, `returning`; one Cancel.

### Where it is built

The loop controller is a sequencer in `internal/watch`, next to `OrderQueue` and on the same pattern: it places one existing order at a time through `Hub.place` and the repair desk, and reacts to their end. That needs no new behaviour in `internal/bot` (U until built: the `home` order's reliability and the repair request answering `undamaged` are the two edges). New: `GET /api/orderloop` (`watch.orderloop/1`: `open`, `teaser`, `state`, `round`, `rounds`, `area`, `minutes`, `repair`, `stoppedWhy`) and `POST /api/orderloop {action: "start" | "stop", area, minutes, repair, rounds | forever}` on the answer listener; state kept in `<state dir>/orderloop.json` so a restart resumes at the start of the round. A loop and waiting queue entries: a finite loop holds the running place and the queue waits; a `forever` loop is refused while entries wait (`queue-busy`), and adding to the queue during a `forever` loop is refused with the same code.

### Ladder rows (`data/unlocks/ladder.json`)

```json
{"id": "orderloop", "group": "ui",
 "trigger": {"kind": "seen", "args": {"what": "manual_repair_cycle"}},
 "reveal": {"title": "Loop order",
            "text": "Send the hero to an area for a set time, home to repair, and out again, as many rounds as you say. One loop at a time; it stops when the hero dies or the repair cannot be paid.",
            "teaser": "Send the hero out, then repair its gear yourself once."}}
```

The `seen` trigger kind exists (V: rows `chronicle_tape`, `risk`, and the chapter events). The event `manual_repair_cycle` is new: the repair wiring emits it when a player-requested repair ends `done` and at least one player zone order was given before it.

For the queue the proposal is to move the first slot behind the Loop and keep today's trigger as the alternative, so nobody who would have had it loses it:

```json
{"id": "orderqueue", "group": "ui",
 "trigger": {"kind": "seen", "args": {"what": "loop_rounds_3"}},
 "alt": [{"kind": "quest_stage", "args": {"quest": "den_of_evil", "min": 2}}],
 "reveal": {"title": "Order queue, first slot", "text": "(unchanged)", "teaser": "Finish three loop rounds, or clear the Den of Evil."}}
```

`orderqueue2` and `orderqueue3` keep their triggers. A row that is already open for a hero stays open (U: assumed from the ladder being a latch; the worker checks it and says so).

**Inconsistency to settle:** today the first queue slot opens at the Den of Evil, which most heroes clear before their first repair. With the alternative kept, the queue can still open before the Loop, which is not the order the user described. Dropping the alternative gives the strict order but takes the slot away from the Den as a reward. The briefs assume the strict order (no `alt` on `orderqueue`); the coordinator can restore the alternative with one line of data.

### UI states

| State | Zone menu | Orders panel |
|---|---|---|
| Loop locked | a muted line under Go now: "Loop: send the hero out, then repair its gear yourself once" (the ladder teaser) | nothing |
| Loop just opened | the reveal the ladder already plays; the Loop button appears with a dot until first used | a Needs you row of kind info: "New: Loop order" |
| Loop open | `Loop...` opens three controls (minutes, repair, rounds) and Start loop | the loop row, round counter, Cancel |
| Queue locked | a muted line: the teaser of `orderqueue` | - |
| Queue open, slots full | Add to queue disabled with "2 of 2 waiting; the next slot: <teaser>" | waiting rows numbered |

A locked control is a line of text, never a dead button.
