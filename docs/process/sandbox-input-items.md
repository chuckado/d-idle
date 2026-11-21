# Item, gold and identify verbs promoted to input (Q13, ADR 0021; 2026-11-20)

> status: candidate. Fork branch `didle/input-items-20261120` (private remote only, never upstream without the owner's go), tip `81195ff04`, binary built from `54b422a9f` (the later commits change tests only). No sandbox image yet; nothing switched. Tags: V = verified in this run (how said), S = read from source, U = unverified.

## What changed in the fork

`item equip|unequip|move|identify|drop` and `gold deposit|withdraw` no longer need `--item-verbs`; `caps` lists `input.item-equip`, `input.item-unequip`, `input.item-move`, `input.item-identify`, `input.item-drop`, `input.gold-stash` always. Still behind `--item-verbs`: `item merc-equip|merc-unequip`, `labels`, and the old `item.*` / `gold.stash` names; behind `--lab-verbs`: every `debug` setup verb. `cmd use` and `cmd pickup` were never behind a flag (S: verbs.hpp), they have differentials only. The read-only debug list (`DebugVerbs::read_only`) is untouched: no mutating verb joined it (V: `lab_verbs_d2d` rc 0, `item_verbs_d2d` rc 0 with merc and cheats still refused).

## The differential (V)

`tests/input_items_diff_d2d.py`, run on the builder's long lane against the binary above with the game data there: two headless d2d (same seed, same setup); the verb on one, the engine's own path on the other (the manual message sequence `cmd hand|body|grid|drop|idwith|stashgold`, and for move, identify, drop, pick-up, use and gold also the pointer path `click|rclick|text|key` through the panels). After each step the digest (`dump player|stats|items`, the ground, the item and gold fields of `state`, the item hover text; not life, mana, stamina, camera, open-panel flags) is compared. A non-refusal step must change the verb run (two idle runs cannot pass); a refusal must move nothing. Result: 19 steps, 0 differences (rc 0).

| verb | scenarios (verb against ...) |
|---|---|
| equip | shield back, nothing comes off (messages); axe over the javelin, the javelin comes off (messages); refusal wrong slot |
| unequip | shield to the pack; axe to the pack (messages) |
| move | scroll 5,2 (messages); scroll 8,3 (pointer); cap pack to stash 2,2 at the chest (messages) |
| identify | the verb's own scroll choice (idwith message); an explicit scroll (right-click scroll, click item) |
| drop | axe (hand, drop); cap (pick up, click the world) |
| gold | deposit 500 and withdraw 150 (stashgold message); deposit 300 (coin, amount box, Enter); refusal too much |
| pick-up (no flag before) | axe, cap (`cmd pickup` against a click on the item after a `move` onto it) |
| use (no flag before) | healing potion (`cmd use` against a right click) |

A fourth run with no flag does equip, unequip, move and drop on the starting items and ends in the same digest as a `--lab-verbs` run (display-only `inv_*` item fields excluded: they exist only with `--item-verbs`). Listed, not hidden: the pointer path needs the panel open, a ground item needs a hover (`move`) before the click or it only walks, the pointer path takes about half a second per click, and the verb picks the first free spot where a player chooses one. Not covered: merc verbs (stay lab), the cube, stacked drops, a joined game.

## Bot side

The bot adopts the new names only with `input.planner` = 1: `internal/invdesk/inputcaps.go` lets `input.item-*` stand for the lab names in the mover and the equipper (`InputCaps`, wired in `cmd/d2bot/invverbs.go`); with the knob off only the lab names count as before. Whether the item verbs run at all is still the knob `invdesk.verbs`. Capability set for this build: `data/input/caps/d2d-fork-input-items.json` (U until an image is loaded).

## Generated hotkey map (I7, knob `input.hotkeyMap`, default 0)

`internal/hotkeymap`: `Generate` orders the planner's skills (main attack, attack when it differs, second element, movement, buffs in upkeep order, aura, summons, town portal), `Apply` writes through a `Binder` and reads the table back, any slot that does not read back is `ErrReadBack`. The real binder (`cmd/d2bot/hotkeymapwire.go`) uses the skill popup (hover, F key) through `Fighter.BindSkill` for F1 to F8 and reads the 16 slots from the `state` line; no save file is read or written. V: unit tests on a fake binder for a Paladin and a Sorceress, including must-fail cases (a write the game ignores, a wrong skill read back). U: not run on a live game; the town portal row name (`Book of Townportal`) and the value the state prints for an item-granted skill are unverified; slots 9 to 16 have no route yet.

## Switch

Not requested. After an image of the fork branch is loaded: `caps` must list the six input names, then the bot with `input.planner=1` and `invdesk.verbs=1`; rollback sets both back to 0 (the click path is unchanged).
