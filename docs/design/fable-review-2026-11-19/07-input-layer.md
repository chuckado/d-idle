# 07. The input layer: clicks, hotkeys and direct verbs

> status: REVIEW for the owner (design review 2026-11-19, card fablereview, docs only; added on the owner's note of 2026-11-19: "the bot clicks a lot where it could use hotkeys"). Nothing here is decided. Part of [00-summary.md](00-summary.md); the short version is section 4a of [05](05-architecture-and-overhaul-options.md) and package R17 of [06](06-roadmap.md).
>
> Tags: **[V]** read by a helper of this review with file and line (helper I = input audit of this tree at `989432fcd`; helper H = read of the public bots' clones plus four web fetches) or by me; **[S]** secondary; **[U]** unverified. No key-binding file, save file or credential was opened by anyone in this review; what is said about binding storage comes from public pages and from our own docs.

## 0. Answer in eight lines

1. The owner's observation is right, and the picture differs by backend. On the **product backend (d2d)** fighting already goes through direct verbs (`cmd skill`, `cmd belt`, `cmd pickup`, `cmd move`, `cmd interact`); the clicking is everywhere else: aura and left-skill selection through the skill popup, stat and skill points, town portal through the inventory, vendor tabs and cells, selling, repair, gold, identify, cube, waypoint rows, drinking from the pack, and a "priming" click after every level change [V helper I].
2. On the **frozen Game.exe backend** skills use F-keys the bot binds itself (with read-back) and potions use keys 1 to 4; everything else is pixel clicks.
3. The best public practice is neither clicks nor hotkeys. kolbot and d2bsng send the game's own client messages: select skill, cast, drink from belt, pick up, interact, buy, sell, add stat point, take waypoint. They press no skill hotkey at all [V helper H]. Hotkeys are what bots WITHOUT engine access use (Koolo, Botty, the vision bots), and those bots carry the cost: bindings must exist and match, and nobody verifies them.
4. We own the engine. **Recommendation: one rule for the fork, "the verb set is the game's own client-to-server message set", and one preference order in the bot: direct verb, then key, then verified click.** A verb that equals a message a real client can send is ordinary input (it goes through the same server-side handler as the click that would have produced it), so fidelity and the crutch rule are untouched.
5. That removes almost every click on d2d, including the two that have caused incidents: the priming "kick" (bounces through warps) and the inventory right-click for town portal (a dropped cast that still spent the charge).
6. With verbs there is NO key-binding bootstrap on d2d. A generated hotkey map is still worth writing into the hero (so the save has sensible skill keys if the player opens it in the real game, and for the Game.exe path): it falls out of the planner's slots in [04](04-stat-skill-agency.md).
7. Speed and noise: a hover-first click on d2d costs about 600 ms and an aura switch through the popup two of them; a select message costs one tick. On the watch video the popup and the inventory stop flashing open during fights.
8. Effort: about 10 to 14 bot cards and a fork branch with 6 to 8 small verbs; fidelity check per verb by differential against the packet handler. It fits inside the first month.

## 1. What the bot does today

Architecture [V helper I]: `input.Controller` (12 methods) mixes three levels: pointer and key (`MoveMouse`, `Click`, `Key`, `Hold`, `Release`), geometry (`ClickWorld`, `MoveTo`, `Pos`) and intent (`SelectSkill`, `CastAt`, `Interact`, `CloseAll`). An intent-level interface already exists beside it (`SemanticInput`: Move, Cast, InteractNPC, Pickup, ChooseMenu, UsePortal, Belt) and has ONE production caller (the mercenary potion). On Game.exe a helper in the Wine prefix posts window messages (a click is move, sleep, down, sleep, up; press 60 to 100 ms, settle 50 to 90 ms, gap 60 to 140 ms). On d2d the controller sends control-socket verbs; pixel clicks are possible but have no modifiers, and a click that is not preceded by a hover can be ignored, so town clicks are wrapped in a two-step hover (150 ms + 450 ms).

| interaction | Game.exe (frozen) | d2d (product) | verdict |
|---|---|---|---|
| drink from belt | key 1 to 4 (default bindings assumed, never verified) | `cmd belt n` | fine |
| mercenary potion | none ("potion wanted (no verb)") | `cmd belt n merc` | fine on d2d |
| drink from the pack | key `i`, hover, right-click the potion | the same clicks | **click**; rare, but it is the emergency path |
| fill the belt | shift-click, fallback drag, memory verified | `cmd belt put` | fine on d2d |
| cast a skill | F-key then right-click at a computed pixel | `cmd skill id x y [unit]` | fine on d2d |
| melee swing | left click or held click | `cmd skill ... unit left` (after "a bare click acts on the previous frame's hover unit": 15 clicks, no damage [V docs/probes/ptrclick.md]) | fine now |
| select right skill, switch aura | F-key, 3 tries, read-back | **popup clicks** at fixed cells then read-back, because `cmd select` leaves the right skill unchanged (deviation R08) | **click, in every fight for a Paladin** |
| select left skill | popup clicks at fixed cells | skipped for swings (the popup path lost 100 ticks in 4 minutes) | click |
| bind skills to keys | the bot binds F-keys itself: open the right popup, find the icon by trying cells, hover, press the key, re-check by read-back (verified live for Sorceress and Paladin only) | nothing is bound; pseudo-keys `sk<id>` live in the bot | Game.exe only |
| Teleport | bound key + right click | `cmd skill` | fine |
| town portal | key `i`, right-click the tome at a grid pixel, close | key `i`, wait 700 ms, pixel from `debug itemat`, hover, right-click; "the click is a cast that is dropped while busy and spends the charge" | **click, with an incident** |
| walk | left click at a pixel, re-click on stall | `cmd move`, but ignored after a level change until one real `click` is sent (the "kick"); a kick on arrival bounced the hero back through a warp [V docs/passagebounce.md, docs/bounce3.md] | verb plus a **priming click** |
| run or walk | not handled | key `r`, state not readable, the bot keeps a belief | key, unverifiable |
| pick up | hover offsets until the hover id matches, then click | `cmd pickup unit` | fine on d2d |
| talk to an NPC, open stash or waypoint | hover-confirmed click, slow-hover retries (61 "greeting never gave way" in the NPC doc) | `cmd interact idx` | fine on d2d |
| NPC menu line | click at offsets from the NPC's screen position, with pixel nudges | click at the coordinates the `menu` verb reports, hover-first | **click** |
| vendor: tab, buy, sell | tab click, right-click buy, right-click or drag sell; a tab click once bought the wrong item until a 600 ms repaint wait was added | tab click, right-click cell, sell by pick up and drop | **clicks** |
| repair | click a fixed button, verify durability | same | **click** |
| gold to and from the stash | click the coin, type digits, Enter | hover-click, `text`, `key Return`; lab verbs `gold deposit` and `withdraw` exist behind a knob | **clicks** |
| move items, identify, cube | take and put clicks; right-click scroll then click item; Transmute button | same; `item move` and `item identify` verbs exist but need `invdesk.verbs` = 1 (default 0) | **clicks** |
| waypoint travel | click act tab and row at fixed coordinates | same, two rounds because d2d "eats a row click" | **clicks** |
| enter a portal | hover click | `cmd interact -2000-k` | fine on d2d |
| stat and skill points | key `c` or `t`, click + buttons and icons, read-back; a wrong icon once taught the wrong skill and paused the bot | the same clicks, although `cmd stat` and `cmd skillpt` exist with no production caller | **clicks** |
| panels | keys `i`, `c`, `t`, `w`, Esc | same keys; closing needs case logic because a second Escape opens the game menu | keys |
| leave the game | Esc and a click | `save`, `quit` | fine on d2d |

Call sites of pointer-type methods outside tests (approximate) [V helper I]: town 20 to 22 `Click` and 8 `MoveMouse`, combat 12 and 3, nav 3 plus 11 `ClickWorld`, leveler 3. Key presses: town 11 to 12, combat 8 to 9.

Things nobody checks today: that keys 1 to 4, `i`, `c`, `t`, `w`, `r` are bound as assumed (only the F-keys are verified); the run toggle's state; on d2d, whether the plain clicks outside town (waypoint rows, popups, leveler) are affected by the "click without hover is ignored" rule, since only the town controller is wrapped [U, helper I's own caveat].

## 2. How public bots do it

| interaction | kolbot, D2BS, d2bsng | Koolo [D2R] | Botty [D2R], vision bots | clientless (D2NG) |
|---|---|---|---|---|
| select skill | message 0x3C through the client's own send function, then poll the skill id up to 1 s; **no hotkey is ever pressed** | press the bound key (bindings read from game memory by a library that is not in the clone) | press a key from the bot's config, which must match the game | message |
| cast | the game's internal click-map call (no OS mouse), retried 3 times until the hero is attacking; or message 0x0C when "packet casting" is on | posted mouse messages at computed screen points | key plus OS mouse | message |
| potion | message 0x26 (also for the mercenary) | belt key; Shift plus key for the mercenary | keys | message |
| pick up | message 0x16 first, click as the fallback, poll up to 1 s | hover spiral until the game's hover id is the item, then click; capped attempts | hold Alt, find labels in the picture, click | message |
| NPC, waypoint, portal | interact message 0x13 first, click later in the retry ladder, a re-sync message when stuck | hover then click; newer fork has optional messages | clicks at fixed UI coordinates | messages |
| buy, sell, stash | messages (a config switch chooses clicks instead) | clicks | clicks | messages |
| stat and skill points | messages 0x3A and 0x3B; d2bsng one per point, 500 ms apart, after checking the counter | clicks, or the same messages as an option | clicks | messages |
| panels | none needed; panel state is read | keys from the game's bindings | keys from config | n/a |
| bind a skill to a key | n/a | open the skill popup, hover the icon, press a free key; **no verification**; the newer fork also rewrites the D2R binding file before launch | click the skill button, find the icon by image, press the key | n/a |

All rows [V helper H: d2bsng `src/backends/lod114d/game/Unit.cpp:1375-1432, 1644-1668`, `GameHelpers.cpp:913-958, 1795-1830`; kolbot `core/Skill.js:505-730`, `core/Pickit.js:470-500`, `core/Packet.js`, `core/Auto/AutoStat.js:605-622`; Koolo `internal/health/belt_manager.go:29-42`, `internal/action/leveling_tools.go:185-230`, `internal/action/step/pickup_item.go:85-165`; Botty `config/params.ini:58-72, 150-171`].

Reliability notes found in those sources [V helper H]: d2bsng settles 100 ms before and after synthetic keys and clicks so tight loops do not race the game's input pump, and documents a 1.14d client crash when a belt drink and a pickup into the same belt column land in one server frame (serialise them); kolbot's retry ladders alternate message and click and re-sync when stuck; Koolo and Botty are tied to one window size because UI coordinates are literals; Botty moves "stand still" off Shift because Shift plus a belt key feeds the mercenary; the AutoHotkey library warns its macros go out of sync under lag.

Lesson: **everyone who can talk to the engine does; hotkeys are the fallback of those who cannot; clicks are the fallback of the fallback.** Our Game.exe path is already at the "hotkeys with verification" level, better than Koolo's unverified bind. Our d2d path is half at the top level and half at the bottom.

## 3. Evaluation

| criterion | pixel click | hotkey | direct verb (a client message) |
|---|---|---|---|
| depends on | window size, panel layout, repaint and animation timing, hover state, what is under the pointer | the binding existing and matching; focus; no panel eating the key | the verb existing in the fork |
| failure modes seen here | wrong vendor item after a tab click; wrong skill learned; ignored clicks without hover; previous-frame hover unit; row clicks eaten; dropped town portal cast; warp bounce from the priming click | "key bound to an empty column; a panel open eats the key" [V docs/primitives.md P3] | a verb that silently does less than the click (R08: `cmd select` did not set the right skill) |
| cost per action | Game.exe about 170 to 330 ms plus read-back; d2d hover-first about 600 ms | one key press, about 100 to 200 ms with settle | one command, effect on the next tick |
| verification | read-back needed every time | read-back needed (active skill id) | read-back still needed, but failures are engine errors with a reason, not silence |
| lockstep and determinism | wall-clock hover waits had to be converted to game time [V docs/probes/waits-audit.md] | fine | best |
| what the viewer sees | popups and the inventory flash open in fights; the pointer jumps | nothing | nothing |
| fidelity | the reference behaviour | identical server-side: a hotkey produces the same select message as a popup click [U from the message tables in the public bots; the differential in section 5 would make it V] | identical IF the verb is implemented as the same handler the message would reach; a verb that skips a client-side check (busy, range, panel open) is a deviation to list |

Effect on flailing: every ignored or eaten click today becomes a retry with a wait, and two of the five loop guards of [05](05-architecture-and-overhaul-options.md) 3.1 were written for bounces that began with a click (the priming kick). Fewer clicks is fewer phantom failures feeding those guards.

## 4. Recommendation

### 4.1 One principle for the fork

**The control channel's action verbs are the game's own client-to-server messages, one verb per message, each routed to the handler that message reaches in the engine.** The list is public knowledge from the bots above: walk and run to a point or unit, left and right skill on a point or unit, select skill (with hand and item owner), use belt item (with the mercenary flag), pick up, interact with a unit, NPC buy, sell, repair, identify, drop and pick to cursor, move item between containers and cube, transmute, deposit and withdraw gold, add stat point, add skill point, bind a skill to a hotkey slot, take waypoint, respec at Akara, leave game. Anything a real client can send is INPUT. Anything else (create an item, set a quest bit, set life) stays a lab verb behind `--lab-verbs`.

Why this line is the right one:

- It gives a principled answer to "is this a cheat?" that does not depend on who asks. The crutch rule (ADR 0014) already says a measured run uses the real input channel; this defines the channel for d2d.
- Upstream d2d must implement these handlers anyway to join a game.exe host over the network (its README lists that as checked for Acts 1 to 4) [S helper E], so the verbs are thin and can be offered upstream.
- It ends the split where `item move`, `item identify` and `gold deposit` are "lab verbs" behind a knob although a player can do all three with a mouse.

### 4.2 One preference order in the bot

`SemanticInput` becomes the interface the behaviour packages use. Each backend implements an intent with the best mechanism it has: **verb, else key, else verified click**, and every intent ends with a postcondition read from state (the pattern of [05](05-architecture-and-overhaul-options.md) 3.2). The raw pointer methods stay available to the adapters only.

| intent | d2d should use | Game.exe keeps | change needed |
|---|---|---|---|
| select right or left skill, switch aura | `cmd select` fixed to set the skill (closes R08) | bound key | fork fix; delete the popup path on d2d |
| town portal | `cmd skill` with the tome or scroll as the item-owned skill (the verb already takes an item argument) | the same skill on a bound key instead of the inventory click | 1 card; removes the 700 ms wait and the dropped-cast incident |
| stat and skill points | `cmd stat`, `cmd skillpt` (exist, unused) | clicks | 1 card (already in [04](04-stat-skill-agency.md) P0) |
| drink from the pack | use-item message for the potion | click | fork verb, 1 card |
| vendor buy, sell, repair | buy, sell, repair messages | clicks | fork verbs; removes tab and cell geometry |
| gold, item moves, identify, cube transmute | the existing item and gold verbs, moved from "lab" to "input" | clicks | flip the knob per verb after its differential |
| waypoint travel | take-waypoint message after `cmd interact` | clicks | fork verb |
| NPC menu line | a `menu choose <line>` verb (the fork already lists the lines) | clicks | fork verb |
| walk after a level change | `cmd move` without the priming click | n/a | fork bug fix; delete `Kick` |
| run or walk | readable state plus set-run message | key | fork: expose the flag |
| panels | not opened at all for verbs; keys only when the player asks to see a panel | keys | |
| what stays a pointer action on d2d | nothing in normal play; pixel `click` remains for the lab and for UI drills | movement, pickup, object and NPC clicks stay hover-verified clicks | |

### 4.3 Key bindings

Facts (public sources only) [V/S helper H]: the classic controls page lists skills on F1 to F8 with up to 16 slots in the options, run toggle R, weapon swap W, Shift to stand still, Alt to show items, Tab for the automap; belt 1 to 4 and Shift plus belt for the mercenary are the common defaults but were not on the page read [U]. A save-format document for 1.13 places WHICH SKILL sits in each of the 16 hotkey slots in the character save header (offset 56, 16 entries) with the left and right skills after it, while WHICH KEY triggers slot N is in the per-character binding file next to the save. A client message exists for "bind this skill to slot N" (the clientless bot's message table names it) [V helper H, not followed further].

Consequences:

- **d2d needs no binding bootstrap for the bot** once selection is a verb. The pseudo-keys `sk<id>` can go.
- **Generate the hero's hotkey map anyway**, from the planner's slots ([04](04-stat-skill-agency.md) 3.1): slot order = main attack, second element, movement, buffs in upkeep order, aura or curse, summons, town portal. Write it with the bind-slot message and read the slots back from state. It costs one card, keeps the save pleasant for a player who opens it in the real game, and is the map the Game.exe path's `EnsureHotkeys` should bind instead of the hand-written `layout` in `profiles.json`. With player-chosen skills a hand-written layout cannot exist; a generated one must.
- **Never read or write the binding file.** On Game.exe keep binding through the UI with read-back (it works and is verified live for two classes); verify the assumed default keys once per game by a harmless probe (press, read the UI flag, press again) instead of assuming.

### 4.4 What must stay a click, and where

On Game.exe: movement, pickup, NPC and object interaction, vendor and stash work. That backend is frozen and an oracle ([05](05-architecture-and-overhaul-options.md) section 6); no new work there. On d2d: nothing in normal play.

## 5. Fidelity of each verb (how to keep this honest)

A verb is promoted from "lab" to "input" only with a differential, in the lane's existing style: the same starting state, action A by the UI path (click or key, the way a player does it) and action B by the verb, then compare the full state dump and the journal for N ticks. Expected result: identical. Known places where they may differ and must be listed if they do: client-side refusals (a cast while busy, a click while a panel is open, range and line-of-sight checks done before the message is sent), the one-tick delay of the pointer path, and target choice (the pointer picks what is under it; the verb names the unit). About one differential per verb, 6 to 8 in all.

## 6. Work, in order

| step | content | cards |
|---|---|---|
| I1 | fork: `cmd select` sets the right skill (R08); `cmd move` works without a priming click; run flag readable | fork branch, 2 to 3 |
| I2 | bot: aura and left-skill selection by verb on d2d; delete the popup path and `Kick` there; scenario: a Paladin fight with three aura switches and zero pixel clicks | 2 |
| I3 | town portal as an item-owned skill cast; stat and skill points by verb | 2 |
| I4 | fork: buy, sell, repair, take-waypoint, menu-choose, use-item verbs as message equivalents; differentials | fork branch, 4 to 5 |
| I5 | bot: town chores, waypoint and pack-drink on the verbs; item, gold and identify verbs promoted after their differentials | 4 to 5 |
| I6 | `SemanticInput` as the only interface of behaviour packages; pointer methods adapter-only; a CI check that counts pointer call sites outside adapters (target: 0 on the d2d path) | 2 to 3 |
| I7 | generated hotkey map from the planner slots, written and read back | 1 |

Measure of success, reported by the Proving Ground: pixel clicks per virtual hour on d2d (today: every aura switch, town visit, level-up and level change; target: zero), and the count of "click ignored" retries.

## 7. Question for the owner

Q13 in [06-roadmap.md](06-roadmap.md): adopt "the fork's input verbs are the game's client messages; verb, then key, then verified click" as the input rule, and promote the item, gold and identify verbs from lab to input after a differential each?

## 8. What I could not verify

- The classic binding file's layout and exact folder for 1.14d (not opened by rule; public sources disagree on the folder).
- That a hotkey and a popup click produce the same message in 1.14d (expected; the differential would show it).
- How upstream d2d structures its client-message handlers, and therefore how thin each verb is [U]; helper E did not read that part.
- Whether the plain clicks outside town on d2d are silently ignored sometimes (helper I's caveat).
