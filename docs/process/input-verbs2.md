# Input verbs, second set: buy, sell, waypoint, portal, pickup (Q13, ADR 0021; 2026-11-21)

> status: candidate. Fork branch `didle/input-verbs2-20261121` (private remote only, never upstream without the owner's go; it starts from `didle/devp17-20261121`). No sandbox image yet; nothing switched. Tags: V = verified (how is said), S = secondary, U = unverified.

ADR 0021: the fork's action verbs are the game's own client-to-server messages, one verb per message, routed to the handler the click reaches; the bot prefers verb, then key, then verified click (07-input-layer.md sections 4 to 6). The first sets were select, repair, menu choose (devp14) and the item, gold and identify verbs (devp15). This is the rest: the vendor, the waypoint, the town portal and the floor.

## The verbs and the message each one is

| verb | message (V: `components/game/protocol.hpp`, `world/commands.cpp`, `world/tick.cpp` of the fork) | handler | refusals named before anything is sent |
|---|---|---|---|
| `buy <stock> [tab] [fill]` | `cmd::Buy`, C->S 0x32 | `World::deal` -> `rules::store_buy` / `store_gamble` | `closed` no vendor window, `stock` no such index, `gold`, `space` no room in the pack, `refused` |
| `sell <item id>` | `cmd::Sell`, C->S 0x33 | `World::deal` -> `rules::sellable`, `rules::store_sell` | `closed`, `item`, `refused` (flag 0x1000, quest item), `belt`, `cursor` (worn item with a full hand) |
| `waypoint <level> [object]` | `cmd::Waypoint`, C->S 0x49 | the `cmd::Waypoint` branch of `commands.cpp` (active waypoint of another level) | `closed` panel not open, `object`, `level`, `here`, `inactive` |
| `portal` / `portal <which>` | `cmd::Interact` with `-2000 - which` (C->S 0x13), the engine walks to the portal and takes it | the town portal branch of `cmd::Interact` | `none`, `elsewhere` (the portal stands in another level) |
| `pickup <ground id> [max]` | `cmd::Pickup`, C->S 0x16, the engine walks to the item and takes it (within 1 cell) | the `pick_item` walk of `tick.cpp` | `item` not on the floor, `far` only when the caller gave a maximum |

Common: single player only (`err joined` on a joined game, the host's world answers there), `err dead` when dead, `err game` outside a game. Reply: one JSON line, then `ok` (`{"verb":"buy","code":"hp1","stock":8,"tab":3,"price":30,"gold_after":49970}`); a refusal is `err <code> <why>` and nothing is sent. No flag: `caps` lists `input.buy`, `input.sell`, `input.waypoint`, `input.portal`, `input.pickup`. Nothing is written except what the game's own handler does: the checks run the handler's own rule functions on copies (V: code read, `apps/d2d/devctl/world_verbs.hpp`).

Not new verbs, on purpose: opening a vendor window is `menu choose Trade` (input.menu-choose, devp14); a town portal CAST from the tome or a scroll is `cmd skill <id> <x> <y> -1 0 <tome id>` / `cmd use <scroll id>` (the portal pair then shows in `portal`); the portal verb only enters one. The waypoint verb works from the OPEN panel, as the client's row click does (open it with `cmd interact <object>`, the panel then closes as after a click). The pickup verb leaves the range to the engine: the hero walks, as the client's 0x16 does; `max` is the bot's own guard.

## The differential (fork `tests/input_verbs2_diff_d2d.py`)

Two headless d2d (same seed, same setup, `--lab-verbs` only for the setup verbs): the verb on one, the engine's own path on the other, a state digest compared after each step (gold, items, held, the player and stats dumps without life, mana, stamina and wear, the ground by id, the vendor's stock; position only where the pointer path nudges it). A step that changed nothing on the verb side fails, a refusal must move nothing and name its cause.

RESULT_PLACEHOLDER

## Bot side (behind the knob `input.planner`, default 0)

| action | seam | with the knob on and the capability listed |
|---|---|---|
| vendor buy | `town.Town.buyUnit` (`internal/town/shop.go`) | `BuyAt(tab, x, y)`: the stock index is the item's place in its tab of `dump vendor`; the gold drop proves it; a `gold` refusal is `ErrNoMoney`; never repeated by a click |
| vendor sell | the sell loop of `shop.go` | `SellItem(id)`; the sale is proved by the item gone and the gold up; a refusal skips the item |
| waypoint | `nav.Navigator.TakeWaypoint` | `TakeWaypoint(level)` after the panel is open and the cooldown waited; an `inactive` refusal is `ErrNoWaypoint`, no click follows |
| portal | `d2dinput.Controller.UsePortal` (used by `Interact`, so also the town's safety portal) | `portal <which>` instead of `cmd interact -2000-which` |
| pickup | `d2dinput.Controller.Pickup` (used by `Interact`, so also `internal/loot`) | `pickup <id>` instead of `cmd pickup <id>`; a named refusal (`item`) is returned |

Rule: a rule refusal is the answer and is never repeated by another path; any other failure (an older build, a joined game, a transport error, an item the stock dump does not list) falls back to the click or `cmd` path that was there before. With the knob off, or without the capability in `caps`, nothing changes byte for byte (tests: `internal/town/worldverbs_unit_test.go`, `internal/nav/wpverb_test.go`, `internal/d2dinput/controller_test.go`, each on a Paladin and a Sorceress where the code has a class). The sim-rig tests (`internal/town/worldverbs_test.go`) need the game tables and skip without them. Capability set for this build: `data/input/caps/d2d-fork-input-verbs2.json` (U until an image is loaded).

## Not done

- No image was built or loaded and nothing was switched (the next image, with these verbs, is a later brief). No live game run: every bot-side claim is offline (fake channel and sim), the fork side ran headless on the builder.
- The town's vendor tab selection still clicks when the buy verb is not used; the sell-by-cursor fallback is unchanged.
- `smoke_d2d` fails at line 512 (`cmd skill didn't hurt zombie`) on every image so far: pre-existing, not touched.
