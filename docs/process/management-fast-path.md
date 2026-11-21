# Management fast path (audit, lane design, first slice)

> status: current (2026-11-21, branch agent/mgmtfast). Built offline and tested; the live RED and green runs are NOT done (U, checklist in section 6).

User principle (2026-11-21): driving the bot (farm X, do quest Y, pause, orders) may queue; MANAGEMENT (identify, equip, unequip, drop, move items, bank gold, vendor buy and sell, Vault take-out, equip and reveal) must answer in seconds through a direct engine path, not behind the goal loop and the town-visit cadence. When a management action physically needs the town and the hero is in the field, the page offers "Do it now" (a priority walk home) instead of silently queueing.

Tags: V = read in the code of this branch or run in a test, U = not verified live.

## 1. Audit: every management action of the viewer and API

Served today: IMM = answered by the POST itself, LANE = instant lane when `invdesk.instant` (and now `ui.managementFast`) and the backend caps allow it, else QUEUE, QUEUE = a request that waits for the bot (`playerActionStep` at a safe moment, `errandTripStep`, `vaultKeepStep`). Latency is from the code (V, no live measurement, U live): the goal loop ticks every `bot.tickMs` (100 ms, `internal/env/tuning_bot.go`), the player step runs on the next tick at a safe moment (no hostile in the clear radius, no town trip), a postponed attempt waits `playerActionGap` (10 s), a failed errand trip backs off `errandBackoff` (5 min), the Vault keeper retries a stuck queue every `vaultKeepWait` (30 s).

| Action (route) | Served today | Latency (estimate) | Precondition |
|---|---|---|---|
| Equip pack item (`POST /api/inventory/equip`) | LANE, else QUEUE | lane under 4 s (`DefaultFastWait`); queue: next tick at a safe moment, minutes in a long fight | anywhere (equip works in the field) |
| Unequip (`/api/inventory/unequip`) | LANE, else QUEUE | same | anywhere |
| Drop / trash (`/api/inventory/trash`) | LANE, else QUEUE | same | anywhere |
| Identify pack or worn item (`/api/inventory/identify`) | LANE (verb), else QUEUE (scroll or tome click path) | same | scroll or tome in the pack |
| Identify stash item | QUEUE, `wait: town` | next town visit | town, stash |
| Stash / take / gold (`/api/inventory/stash`, `take`, `gold`) | LANE in town at the chest, else QUEUE `wait: town` | lane under 4 s in town; field: walk home | town, stash |
| Buy scrolls, Identify all (Cain) (`buyid`, `cain`) | QUEUE (errand trip) | trip at the next safe moment | town, vendor or Cain |
| Sell now (`sell` kind) and marks (`/api/inventory/mark`) | marks IMM (data); sale QUEUE (errand) | trip | town, vendor |
| Gear sets apply (`/api/inventory/sets/apply`) | QUEUE (series of equips) | tick each | anywhere |
| Pin / unpin (`/api/vault/pin`, `unpin`, `/api/inventory/unpin`) | IMM (store write) | under 1 s | none |
| Discard a Vault record (`/api/vault/discard`) | IMM (store archives it) | under 1 s | none |
| Vault Keep (take out into the stash or pack) (`/api/vault/keep`) | QUEUE; `now` sets `tripWant` | in town: next tick (30 s retry when stuck); field with `now`: errand trip; without: next visit | town, stash |
| Vault Equip (`/api/vault/equip`, #1061) | QUEUE: a Keep to the pack, then the equip order | as Keep; one town visit | town (the record is created in game at the stash) |
| Vault Reveal, Reveal gold (`/api/vault/reveal`, `revealgold`, #1071) | QUEUE (scroll or Cain, or a purchase); Ether reveal is data and IMM | as Keep | town, scroll or vendor |
| Vault Sell (`/api/vault/sell`) | QUEUE | as Keep | town, vendor |
| Loot rules apply, bulk identify rules (`PUT /api/lootrules`, #949 #951 #959) | IMM (rule set write, decisions run on the next loot) | under 1 s | none |
| Points: stat and skill spend (`/api/hero/points/*`) | IMM by verb, read back (`internal/points`) | under 2 s | anywhere |
| Repair (`/api/repair`) | QUEUE (repair trip) | trip | town, smith |
| Mercenary revive (`/api/merc/revive`) | QUEUE | trip | town |
| Belt / potion arrangement | none on the page (bot only) | n/a | n/a |
| Farm gold order, goals, orders, Pause (`/api/farmgold`, `/api/orders`, `/api/pause`) | DRIVING: queue is fine (pause freezes at once) | n/a | n/a |

Findings. (1) The Inventory sheet already has an instant lane (TASK-427, `internal/invdesk/instant.go`) but it hid behind THREE knobs (`invdesk.verbs`, `invdesk.instant`, `input.planner`) and had no umbrella; now `ui.managementFast` opens all three (section 3). (2) The Vault sheet had only the queue and a "Recall now" checkbox without status (the page could not say "going home", "in town" or why it waits). (3) The pack-to-pack move verb exists (`ItemMove`) but the desk has no pack move request (`instant.move` says so). (4) Vendor buy, sell, waypoint, portal and pickup have no verb yet (agent/verbs2), so they stay QUEUE.

## 2. The direct path: what serializes input today (V, read)

- The control channel: `invdesk.Channel` is one token; the bot's executors (equip, mover, trasher) and the lane reach the client only through `invdesk.Serial`, so two verbs are never in flight together and a lane verb cannot land inside a bot verb (`instant.go`).
- Click sequences are not covered by that token: they run on the bot goroutine. The lane therefore stays closed (`FastLane.Hold`, `cmd/d2bot/invverbs.go inputHeld`) while the cursor holds an item, a panel is open (the bot is at a vendor, the stash or an NPC menu) or a safety action runs. That is the "safe point" rule; a verb needs no panel, so the lane never opens one.
- Combat: equip, unequip, drop and identify through verbs are allowed in a fight (the verbs are server calls); vendor and stash actions need the town, the lane says `far` and the request falls to the queue with `wait: town`.
- Named runs (`D2BOT_RUN`: gate, soak, drill) keep the queue (crutch rule, the user decides, `item-verbs.md` section 10); the recall desk is not installed there either.

Design: a "player command lane" separate from the bot tick was NOT built, because the existing lane already is one (it runs the verb on the API goroutine under the channel token and returns the final state in the POST answer, within a poll for the rest). The new piece is the priority intent for the town: `Recall now` (`internal/bot/recall.go`), consumed at the START of `runGoal`'s player steps (before `playerActionStep`) and by the guard (`playerDue`, so a roam or travel walk is cancelled, rate limited by `playerCancelGap`). It uses the errand trip's `safeMoment` (no hostile in reach, no trip running), walks with `Town.ToTown`, and the queued town work runs in the same tick on arrival. States `waiting`, `going`, `arrived`, `failed` (retry after 20 s) are the status the page shows. Verbs still to come (buy, sell, waypoint, portal, pickup from agent/verbs2) join the same pattern: add the capability name to `capFor` (`instant.go`) and mark the kind town-only in `needsTown`; the lane, the channel and the recall already serve them. Nothing here writes game memory.

Latency target: equip, unequip, drop, identify in the field under 2 s from the tap (lane bound 4 s worst case; the verb plus the read-back of fresh snapshots, up to 25 polls of 100 ms). U live.

## 3. First slice (built)

- Knob `ui.managementFast` (default 0, `internal/env/tuning_ui.go`): stands for `invdesk.verbs`, `invdesk.instant` and `input.planner` for the desk's lane and executors (`verbKnobs` in `cmd/d2bot/invverbs.go`; V: `TestManagementFastStandsForTheThreeKnobs`). It does not enable anything by itself on a backend without the verbs.
- Equip from the pack, identify, drop: the existing lane, now behind the umbrella. Swap-back: the verb and the click path both leave the replaced item in the pack (the lane never drops anything it was not asked to; U live for the verb's swap, V for the queue path in `equipnow` tests).
- Refusal: when the lane is closed or refuses, the request queues as before and the page names the reason from `view.instant.why` and the request's `code`/`wait`.
- Recall now: `POST /api/recall {}` (answer listener) and `GET /api/recall` (`internal/watch/recallapi.go`), reply `watch.recall/1` with `enabled`, `state`, `in_town`, `reason`, `error` and `ok`; `{"cancel": true}` withdraws a waiting command. Off (error `ui.managementFast`) when the knob is 0. The Vault equip sheet shows the status line and a "Do it now: go to town" button when an order waits and the hero is in the field (`vault.js`, text keys `recall*` in `vault.json`); the old "Recall now" checkbox stays (it sets `now` on the order, the keeper's own trip).
- Scenarios (`tools/provingground/scenarios/manage.json`, all `pending`: the runner cannot yet substitute a given item's id into `trigger.api` nor seed a vault record, same reason as `vault.equip-one-tap`): `manage.equip-immediate`, `manage.identify-immediate`, `manage.recall-now`.

## 4. Tests (V, offline)

- `internal/bot/recall_test.go` (Paladin and Sorceress): field recall walks home, says arrived and the goal resumes; knob off refuses with the reason and nothing walks; a hostile in reach holds the walk; a failed walk reports `failed` with a reason, cancel withdraws.
- `internal/watch/recallapi_test.go`: gates (503 unbound, token, cancel, refusal carries `ok:false` and `recall_refused`).
- `cmd/d2bot/invverbs_test.go`: the knob umbrella.
- Run: `go test -p 1 -race -count=20 ./internal/bot -run RecallNow`, `go test -p 1 ./internal/watch ./cmd/d2bot ./internal/env ./cmd/d2scenario`.

## 5. Not done

- Direct vault equip, reveal and take-out without the town: impossible, the record is created at the stash (town); the answer is Recall now plus the same-visit run. The keeper's 30 s retry (`vaultKeepWait`) applies only when the queue is stuck, not to a fresh order (the queue length changed).
- A pack move request (pack-to-pack) in the desk; the buy and sell verbs (agent/verbs2); a viewer status for the Inventory sheet's recall (only the Vault sheet has the button).
- `ui.managementFast` stays 0 everywhere; the user turns it on in the sandbox.

## 6. Live checklist (lab slot, hidden instance or the builder, nice 19, CHEATED setup = mechanics only)

1. Fork build advertising `input.item-equip`, `input.item-identify`, `input.item-drop` (or the lab names). Start the bot with `ui.managementFast=1`.
2. RED first: knob 0, field fight, tap Equip on a pack upgrade: expect the queue answer (`how: queue`) and a delay above 2 s in a fight.
3. Knob 1: the same tap answers `how: fast` within 2 s, no `totown` in the log, the fight continues, the old helm is still in the pack. Repeat with Identify (scroll in the pack) and Drop.
4. Vault item, hero in the field: tap Equip, then "Do it now": expect `recall trip` in the log, state `going` then `arrived`, the helm worn, the hero back on its goal.
5. Determinism pair on one host; `writes after start: 0`. Paladin and Sorceress.
