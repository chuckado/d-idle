# I. The character screen: hero, mercenary, pack, belt, stash on one canvas

> status: section 3 (the canvas) and the tooltip and action placement of section 4 are SUPERSEDED by [J-second-attempt.md](J-second-attempt.md) section 4 (2026-11-14). Sections 5 to 10 (requirements, comparison, gestures, instant actions, the mercenary, the test table) stand. Earlier status: design proposal (TASK-419, 2026-11-14), not built. Asked for by the user after the first pass: "the current one is too spread out and too discontinuous, and the same goes for the mercenary side of the house"; "functional and readable but most importantly usable". It replaces the layout of `/watch/inventory`; routes, POST bodies and every existing field of `watch.inventory` stay. Tags: V = read in the code or seen in the picture named, S = from a doc or from knowledge of the game named (not re-checked on the web for this page), U = unverified.

Mockup (private link, fake data, placeholder shapes, no game art): see the README. File: [mockups/character-screen.html](mockups/character-screen.html).

## 1. What is wrong today (V: [audit/shots/inventory-1440.jpg](audit/shots/inventory-1440.jpg), fixture viewer)

1. **The doll and the pack are in different cards, 350 px apart**, with an empty "Pick a slot or an item to see it here" box between them. Every equip is a trip across the screen.
2. **A dead zone of about 450 px** under the backpack, while the left column scrolls for two screens (doll, mercenary, gear sets).
3. **The mercenary is a fold under the hero's doll** with a doll frame as large as the hero's, three empty boxes and an "n/a" slot, and a line saying it is read-only.
4. **Gear sets are three text cards 110 px tall each**, further from the doll than anything else.
5. **Stash, belt and gold are not on the first screen** (gold is a line of text under the pack; the stash is further down, the belt too).
6. **Three ways to learn about an item**: the hover tooltip, the detail box, the legend line. None says whether the hero can use the item.
7. **Repair sits alone at the bottom right**, unrelated to the broken weapon shown top left.

## 2. Patterns from other games (S: from knowledge of these games' interfaces; patterns are described, nothing is copied)

| Game | Adopt | Why | Avoid |
|---|---|---|---|
| Diablo II (the original) | The left panel swaps (stash, mercenary, cube) while the inventory stays open on the right. Requirement lines in the tooltip turn red when unmet. Shift-click a potion to the belt, ctrl-click to sell at a vendor | Players of this game already think "left = the other container, right = me". Red requirement text is the convention this audience reads without a legend | A 10 by 4 pack with no search; one modal panel at a time; no comparison |
| Diablo III | Comparison tooltip: the worn item beside the hovered one, with the differences listed. Right-click = the obvious action (equip) | One hover answers "is this better for me". One gesture for the common case | A single green or red "better" arrow from a hidden formula: it makes the gear decision for the player, which this project forbids (`docs/gear-policy.md`) |
| Diablo IV | Tabs filter the pack by kind (equipment, consumables, quest). Mark as junk, then one "sell all junk" | Small packs get findable; the junk mark is exactly our `sell` mark | Full-screen character sheet that hides the game |
| Path of Exile | Ctrl-click moves between pack and stash. A search box that dims everything that does not match instead of hiding it. Stash tabs with names | The fastest stash interaction there is; dimming keeps positions stable so nothing jumps | Its search syntax (regex-like) as the only way; dozens of tab types |
| Grim Dawn | Auto-sort button for the pack. Comparison against both ring or weapon slots at once | Tetris-packing by hand is busywork; two-slot compare matters for rings | Many nested bag tabs |
| Last Epoch | The loot filter editor: ordered rules, each "show / hide / recolour" with conditions picked from lists, and a count of what a rule matches. Stash search with saved filters | The best existing model for a rules editor a non-programmer can use (taken up in `loot-rules-ux.md`) | Affix-tier jargon in every row |
| Torchlight II | The pet has its own small inventory next to the player's, and "send pet to town to sell" | A companion panel that is always beside the player's, small, with one clear errand button: the model for the mercenary and for "Sell now" | Fishing and pet-transform clutter |
| Titan Quest | Drag highlights the cells an item will occupy, green when it fits and red when not, before the drop | The validity preview is the whole point of drag and drop | Tiny cells with unreadable icons |
| Mercenary and companion panels (D2 hireling screen, D3 follower panel, PoE2-style companion lists, S) | Few slots, the companion's own level and life in the header, items checked against the companion's stats, dead or dismissed as a state of the same panel | Matches what the game allows (four slots at most, V `docs/research/merc-gear.md` 2.3) | Giving the companion a full doll with slots it can never use |

## 3. The canvas

One frame, three zones that touch, no card gaps. The pack is the hub: everything is one move away from it.

```
1440 px
+ Character ---------------------------------------------------------------------------------------------+
| HERO  Simmy, Sorceress, level 7    | PACK  10 x 4                  search [        ]  | STASH  6 x 8        |
| Str 38  Dex 25  life 86/120        | [chips: usable by hero | by merc | unidentified   | at the chest / in the field: |
| +------+  [helm]   +------+        |         | marked sell | potions]                  | moves wait for town          |
| |weapon| [amulet]  |shield|        | +--+--+--+--+--+--+--+--+--+--+                   | +--+--+--+--+--+--+          |
| |      |  [armor]  |      |        | |  |  |  |  |  |  |  |  |  |  |                   | |  |  |  |  |  |  |          |
| +------+           +------+        | |  |  |  |  |  |  |  |  |  |  |                   | |  |  |  |  |  |  |          |
| [gloves][ring][belt][ring][boots]  | +--+--+--+--+--+--+--+--+--+--+                   | ...                          |
| Sets: (Fire damage)(Town)(+ save)  | BELT [][][][]   GOLD 1,240  [Sell marked (2)]     | gold 5,300 [Deposit][Withdraw]|
|------------------------------------| -----------------------------------------------   | Stash from anywhere: locked  |
| MERCENARY  Desert Guard, lvl 14    | SELECTED  Short Sword                             |   (a convenience that skips  |
| life ====  alive   Str 60 Dex 40   |  Requires level 1 (you 7)  Str 30 (you 38) ok     |   a game rule; unlock hint)  |
| [helm][armor][weapon]              |  [Equip] [Give to mercenary] [Mark sell] [Drop]   | tabs: Stash | Cube | Loot rules|
+----------------------------------------------------------------------------------------------------------+
| Pending: Light Plate -> stash (waits for town)  x     Identify Ring ... done                              |
+----------------------------------------------------------------------------------------------------------+
```

| Zone | Holds | Rule |
|---|---|---|
| Left: who wears | Hero header (name, class, level, Str, Dex, life) and doll; the gear set chips directly under the doll; the mercenary strip under that: header (kind, level, life meter, state), only the slots this mercenary can use (three or four, never an "n/a" box), Revive when dead | The hero's and the mercenary's slots are the only drop targets for wearing |
| Middle: the pack | Search and filter chips, the 10 by 4 grid, the belt, gold on the hero, "Sell marked (n)" (the Sell now command), and the inspector for the selected item | Always visible at every width |
| Right: the other container | Tabs: Stash (grid, stash gold, deposit and withdraw), Cube (when owned), Loot rules. The at-the-chest state is a line on top | D2's own "left panel swaps" pattern, mirrored so the doll stays left |
| Bottom strip | Pending and recent actions: the same rows as the Orders panel, filtered to item actions | One line when empty |

Repair is not a card here: the broken item's slot carries the badge, the inspector of a damaged item has Repair now, and the page's status band has the same Needs you count as `/watch`.

### Widths

| Width | Layout |
|---|---|
| 1440 | three zones side by side as drawn; nothing scrolls at 900 px height |
| 900 | two columns: left (hero, sets, mercenary), right (pack, belt, inspector); the stash tabs become a full-width band under both, opened by default only in town |
| 390 | the pack and belt are always the lower part of the screen; the upper part is one of Hero, Mercenary, Stash (a three-way switch). The inspector is a bottom sheet with the actions as 44 px buttons. This is D2's "inventory stays, the other panel swaps", vertical |

## 4. One tooltip, one inspector

- **Hover** (fine pointers): the tooltip. Top: the game's own lines (`tip`, V in `status.Item`). Then the requirement rows (section 5). Then, for something wearable, the comparison (section 6).
- **Click or tap**: selects the item. The inspector shows the same content and the action buttons. It replaces the empty "Pick a slot" box, the legend line and the per-page detail card. A selected item also lights every place it can go.
- No other place explains an item. Marks (sell, keep), pin and pending state are corner badges on the cell with a word in the tooltip.

## 5. Requirements are first-class (user constraint, 2026-11-14)

Not every class can wear every item, and the right class still needs the level, strength and dexterity. The mercenary is checked against his own level and stats and his type's item kinds.

### What shows

| Where | Shows |
|---|---|
| Tooltip and inspector | One row per requirement with needed and current: `Level 12 (you 7)`, `Strength 47 (you 38)`, `Dexterity 20 (you 25)`, `Class: Paladin only`. Unmet rows are red with the gap in words (`needs 9 more`), met rows green with a tick. For a class item: `Usable by: Paladin`. A second block "For your Desert Guard" with his numbers when the item is of a kind any mercenary wears |
| Grid cells (pack, stash, cube) | An item the hero cannot use now is dimmed and carries a corner badge with the first reason: `CLS`, `LVL`, `STR`, `DEX`, `BRK` (broken), `UNID`. The badge has the full text as its label. If the mercenary can use it the badge reads `MERC` in the info colour instead of dimming |
| Doll slots while an item is selected or dragged | Green outline = the drop will be accepted; red outline with the reason under the cursor = refused; no outline = not a slot for this kind. The colours come from `canWear` and `noWearSlot` (V, `invdesk.ItemView`), the server's own rule, never from a second copy in the page |
| Filter chips | "Usable by hero", "Usable by mercenary" dim everything else in pack and stash |
| Gear set chip | A set with an item that no longer meets its requirements shows a warn mark; Apply lists them ("skipped: War Staff needs 47 strength, you have 38") and skips them. It never substitutes (the existing set rule, V `docs/gearsets.md`) |
| Instant action | A refused action never shows as done: the page checks `usable` first and does not send; if the server still refuses, the item returns with the reason (section 8) |
| Mercenary strip | His level, Str, Dex in the header; "Give to mercenary" is offered only on items with `usable.merc.ok` |

### API (additive on `watch.inventory`, next schema number)

```json
"req":    { "lvl": 12, "str": 47, "dex": 20, "class": ["Paladin"] },
"usable": {
  "hero": { "ok": false, "reasons": [
      { "code": "requirements", "stat": "str", "need": 47, "have": 38, "text": "needs 47 strength (you have 38)" },
      { "code": "class", "need": "Paladin", "have": "Sorceress", "text": "Paladin only" } ] },
  "merc": { "ok": true, "reasons": [] }
}
```

- `req` is the item's own requirement after its modifiers; `usable` is the verdict for the hero and, when a mercenary is owned, for him. Reason codes are the equip path's own vocabulary (`class`, `requirements`, `identified`, `twohand`, `slot`, `nospace`, `dead`, V `docs/design/item-verbs.md` section 5), extended additively with `stat`, `need`, `have`, and for the mercenary `merctype` ("a Desert Guard carries spears and polearms").
- `noWearSlot` (V, slot to reason text) stays; its texts gain the numbers.
- On equipped items `usable` is still filled: a worn item whose requirement is no longer met (after a stat change) shows the warn badge.

| Fact | Status |
|---|---|
| `canWear`, `noWear`, `noWearSlot` computed server-side for pack items | V `internal/invdesk/desk.go` |
| Base requirements `reqStr`, `reqDex`, `reqLvl` | V `internal/status/inventory.go` |
| The click path's pre-check knows class, level, strength, dexterity, identified | V comment in `internal/leveler/playerequip.go`; the reason texts carry no numbers today (U exact strings) |
| Requirement modifiers (ethereal, "requirements -x%", affix level requirements) | U in the bot. d2d's mercenary rule does not apply ethereal or "requirements -x%" (V note in `docs/research/merc-gear.md` 2.2), so the page must show what the server will do, not what the player expects; a difference is listed under the row as "the game may allow this" only if a probe confirms it |
| Mercenary level, Str, Dex and legal slots in the view | V `MercView` and the fixtures `inventory_merc_*.json` |
| Mercenary legal kinds per type | V decomp for Rogue (bow), Desert Guard (spear, polearm), Iron Wolf (one-handed sword, shield), Barbarian (sword); U javelins, two-handed swords, axes (`docs/research/merc-gear.md` 2.3) |

## 6. Comparison

For a wearable item the tooltip puts the worn item of the same slot beside it (both ring slots; both hands for a weapon) and lists the differing lines as facts: `Defense 18 -> 31`, `+12 fire resistance (new)`, `durability 5/24 -> 30/30`. No arrow, no "better": the player decides (`docs/gear-policy.md`). Comparison is offered only for items the wearer could use or nearly use; an unmet requirement is the first line: `would need 9 more strength`.

## 7. Gestures

| Gesture | Does | Same as button |
|---|---|---|
| Click | select, open the inspector, light the valid places | - |
| Double-click or Enter | the primary action for where the item is: pack wearable and usable -> Equip; potion -> Belt; worn -> Take off; stash item at the chest -> Take | first button of the inspector |
| Drag | move; cells and slots preview green or red before the drop | Equip, Take off, Stash, Take, Belt |
| Ctrl-click | pack to stash and back (at the chest: instant; in the field: queued for town) | Stash, Take |
| Alt-click | mark for sale, again to clear | Mark sell |
| Delete | Drop, with the one confirmation the Orders rules keep for trash | Drop |
| `/` | focus search | - |

Every gesture has a button in the inspector; a phone uses only taps and buttons. Sort: one "Tidy pack" button that packs items to the top left by size is recommended for a later slice (it is a batch of moves; question 17).

## 8. Instant actions

The user wants item actions to happen at once instead of waiting for the bot.

### What exists

- Item verbs (TASK-367, TASK-371): `equip`, `unequip`, `move`, `identify`, `drop`, `gold` go through d2d's own server item handlers in one tick, planned on a copy with the server's rule functions; refusals come back as codes. Knob `invdesk.verbs`, default 0; d2d must be started with `--item-verbs` (V `docs/design/item-verbs.md` sections 2 to 5). The stash rule is strict: a stash move is refused unless the hero is at the chest (`place`).
- The queue (`internal/invdesk`): a request waits for the bot's goal step and a safe moment, and errands walk home (V `docs/inventory-actions.md`).

### The instant path

| Step | What happens |
|---|---|
| 1. Page | Checks `usable` and the target. Not allowed: no request, the reason shows at the item. Allowed: the item is drawn at its target in a pending style (a ring on the cell, not the done look) |
| 2. `POST` (the existing routes) | The Desk sees the request kind is verb-capable (`view.instant` says which) and runs it on a fast lane now, instead of leaving it for the next goal step: the verb goes to d2d at the next tick. No safe-moment wait, no walking, no panel |
| 3. Reply | `{ok, request: {state: "done", how: "verb"}}` within the request, or a refusal `{code, error}` with the verb's own code. If the fast lane cannot run (the bot holds the cursor in a click path, the backend has no verbs, a named run), the request is queued as today and the reply says `state: "queued", how: "queue"` |
| 4. Page | done: the pending ring goes. Refused: the item returns to where it was and the reason shows at the item and in the pending strip. Queued: the item stays where it was with a clock badge and the strip says what it waits for |

Target time from click to done is one round trip plus one game tick (U until measured; the verb applies on a server tick, 25 per second).

| Action | In the field | In town | Why |
|---|---|---|---|
| Equip, take off | instant | instant | plain client actions in the real game too |
| Move in the pack, to and from the belt | instant | instant | same |
| Identify with a scroll or tome | instant (`notool` when none) | instant | same |
| Drop | instant, after the confirmation | instant | same |
| Stash, take, deposit or withdraw gold | **queued for the next town visit**: the item keeps its place with a clock badge, a ghost marks the target, the strip reads "Light Plate -> stash, waits for town"; Cancel on the row | instant when the hero is at the chest, else at the chest on this visit | the game rule: the chest must be open (`place`) |
| Sell marked, Cain, buy scrolls, repair, revive | errand (walks home) | at the vendor on this visit | they need an NPC |
| Mercenary equip, take off | not possible yet (section 9) | not possible yet | no verb, no executor |

Rules that stay: safety first (a dead hero refuses with `dead`; the fast lane never runs while a safety action holds the input); the priority ladder for everything that is not a verb; one vocabulary of refusal codes for both paths; the read-back check after a verb (a verb that answered ok but did not move the item is reported failed, not retried, V item-verbs section 6). In a named run (`D2BOT_RUN`) the instant path is off, like the knob (U: the crutch-rule text for verbs is a draft in item-verbs section 10 and is the user's to decide).

API (additive): `view.instant: {equip, unequip, move, identify, drop, stash, take, gold: bool}` (what would be instant right now, from `caps`, the knob and the hero's place); replies gain `request.how` (exists, V) and return the final state when the fast lane ran; the request states are unchanged.

### Stash from anywhere (decided by the user, 2026-11-14)

Not the default. Stash and take stay at the chest with the pending badge. A possible later unlock is tracked as TASK-420 (its note is on another branch). The screen reserves one line under the stash grid: "Stash from anywhere: locked" with the unlock hint and the words "a convenience that skips a game rule". Nothing behind it is built.

## 9. The mercenary until he can be equipped

Today: the view reads him and his gear (V); `equip` and `unequip` with `target: "merc"` are refused with code `merc` (V `invapi.go` comment); there is no verb (`MercItem`, packet 0x61, is not wrapped; item-verbs "Mercenary items") and no click executor (merc-gear stage C).

Until card B9 (fork verbs) and its bot wiring land: the strip shows him, his level, life, state and gear; hovering a pack item shows the "For your Desert Guard" requirement block and the `MERC` badge; dropping on his slot or pressing "Give to mercenary" shows one line at the slot: "Giving items to the mercenary is not built yet." The button is not drawn as a live control. When B9 lands the same gestures work, instant, with his own refusal reasons (`merctype`, `requirements` with his numbers, `dead`: no gear change on a dead mercenary, V). Hiring is not on this screen (the policy hires in town); Revive is (card B3).

## 10. Test plan for the validity display (part of slice C1)

Fixture-driven, both for the Go view (`usable`, `req`, `noWearSlot`) and the node DOM goldens (badge, dimming, tooltip rows, slot outline). One row per case; "have" is the wearer's current number.

| # | Wearer | Item | Stats | Expect |
|---|---|---|---|---|
| 1 | Sorceress 7, Str 38 | a one-handed sword whose strength requirement she meets | met | usable, green rows, Equip offered |
| 2 | Sorceress 7, Str 38 | War Staff, Str above 38 | str unmet | `requirements/str` with need and have, `STR` badge, hand slot red with "needs N strength (you have 38)" |
| 3 | Sorceress 7 | an item with level requirement 12 | lvl unmet | `LVL` badge, row "Level 12 (you 7)" |
| 4 | Sorceress | a Paladin shield | class | `CLS`, "Usable by: Paladin" |
| 5 | Paladin | the same shield | met | usable |
| 6 | Paladin | a Sorceress orb | class | `CLS` |
| 7 | Amazon | a bow with Dex above hers | dex unmet | `DEX` with numbers |
| 8 | Barbarian | a primal helm; the same on a Druid | met; class | usable; `CLS` |
| 9 | Necromancer, Druid, Assassin | wand or head; pelt; claw, each on its class and on another | met; class | usable; `CLS` |
| 10 | any | an unidentified ring | - | `UNID`, Equip not offered, Identify offered |
| 11 | any | a broken weapon (durability 0) | - | `BRK`; worn: warn badge on the slot |
| 12 | any | a two-handed weapon with a shield worn | - | hand slot accepts, the tooltip says the shield goes to the pack; with a full pack `nospace` |
| 13 | any | a ring with both ring slots full | - | both slots green (swap), comparison against both |
| 14 | hero after a stat change | a worn item now under its requirement | - | warn badge on the worn slot, set Apply skips it with the reason |
| 15 | Desert Guard 14, Str 60 | a spear he meets; a polearm above his Str; a bow | - | usable.merc ok; `requirements` with his numbers; `merctype` |
| 16 | Rogue | a bow; a crossbow | - | ok; `merctype` |
| 17 | Iron Wolf | a one-handed sword and a shield; a two-handed sword | - | ok in both hands; refused |
| 18 | Barbarian mercenary | a sword; a primal helm | - | ok |
| 19 | any mercenary, dead | any legal item | - | `dead`, slots read-only |
| 20 | no mercenary | any | - | no merc block, no `MERC` badge |
| 21 | any | an ethereal or "requirements -x%" item | - | the view shows the number the server will enforce; the case is listed as U until a probe says which number that is |

Item names in rows 1 to 21 are placeholders for the worker to bind to real rows of the game tables at test time (table tests run with the game data directory set; nothing from the tables is committed).
