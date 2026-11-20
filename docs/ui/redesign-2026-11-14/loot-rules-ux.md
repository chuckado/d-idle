# Loot rules: the player-facing UX and what it needs from the engine

> status: UX proposal (TASK-419, 2026-11-14), not built. The ENGINE (rule language, evaluation, merging `internal/tidy/policy.json` and the `internal/loot` rules into one) is designed by another agent in parallel; this page does not design it. It says what the player sees and lists, in section 5, what the editor needs the engine to provide, so the two designs meet. Tags: V = read in the code named, U = unverified.

## 1. What the player gets

A list of rules the bot follows for what it picks up and what it does with what it carries. The bot still makes no equipment decision: a rule can keep, stash, sell or drop an item; it never puts one on (`docs/gear-policy.md`).

Today the same decisions are two hard-wired tables the player cannot see: the pickup rules (`internal/loot/rules.go`: codes, types, qualities, ethereal, level ranges, stat conditions, V) and the tidy policy (`internal/tidy/policy.json`: ordered rows, first match wins, actions keep, stash, sell and others, per-class overrides, V). The UX shows them as one list in two stages.

## 2. Where it lives

A tab of the right-hand container on the character screen: `Stash | Cube | Loot rules` (I-character-screen.md section 3). Reason: the pack and stash it acts on are beside it, so the preview can light the very items a rule would touch. At 390 px it is the third choice of the upper switch. It is recommended over a page of its own (question 19).

## 3. The editor

```
LOOT RULES   preset: Starter (edited)   [Presets v]  [Import] [Export] [Undo]          rules: on

ON THE GROUND (what to pick up)                                   matched of the last 50 drops
 1  [on]  Runes, gems, jewels, charms          -> pick up                       6
 2  [on]  Unique and set items                 -> pick up                       1
 3  [on]  Rare items, hero level under 30      -> pick up                       4
 4  [on]  Magic items usable by hero           -> pick up                       3
 5  [on]  Gold over 50                         -> pick up                      21
 -        Everything else                      -> leave                        15   (fixed last row)

IN THE PACK (what to do with it in town)                          matched in the pack now
 1  [on]  Runes, gems, jewels, charms          -> stash                         2
 2  [on]  Unique and set items                 -> stash                         0
 3  [on]  Usable by mercenary, rare or better  -> keep                          1
 4  [on]  Sell value over 300                  -> sell                          3
 5  [on]  White and magic, not usable          -> sell                          4
 -        Everything else                      -> keep                          6   (fixed last row)
[+ Add rule]
```

- **Two stages, ordered, first match wins.** Stage one decides pick up or leave; stage two decides keep, stash, sell or drop for what is in the pack. Each stage ends with a fixed catch-all row whose action can be changed but which cannot be deleted.
- **A rule is a sentence built from pickers**, never typed syntax: `[Quality: rare or better] and [Type: body armor] and [Usable by: hero] -> [stash]`. Add a condition from a menu grouped as in section 5; each condition is a chip with its own small editor (a range, a list with search, a yes or no).
- **Reorder** with drag or with up and down buttons; a rule that can never match because an earlier one catches everything it would is marked "never reached" with the earlier rule's number.
- **Protected, shown but not editable** (a grey block above the list): quest items, the start kit, anything the player pinned or marked (a mark beats every rule), worn items. So no rule can sell what the player marked keep.
- **Each rule can be switched off** without deleting it, and has a free-text name.
- **Actions.** Ground: `pick up`, `leave`. Pack: `keep`, `stash`, `sell`, `drop`, and `ask me` (raises a Needs you row "Rare ring found: keep, stash or sell?" and keeps it until answered). Drop asks for one confirmation when the rule is saved, not per item.

## 4. Preview, presets, safety

- **Live preview.** Next to every rule: how many of the last N drops (ground stage) or of the items in the pack and stash now (pack stage) it matches. Selecting a rule lights those items in the pack grid and lists the drops ("kept 3 of the last 20 drops: Chipped Topaz, Tir Rune, ..."). While editing, the counts update before saving, and a line states the net change: "With this change: 2 items in your pack would be sold instead of kept."
- **Why, on every item.** The item tooltip gains a line: "Loot rule 4 (Sell value over 300): sell at the next vendor". The pack badge for the planned action exists today as the `tidy` field (V `invdesk.ItemView.Tidy`).
- **Presets** (data, not code): `Starter` (today's behaviour, the default, so nothing changes for a player who never opens the tab), `Runes and gems` (pick up and stash every rune, gem, jewel and charm; leave white items), `Sell junk` (pick up what sells above a value, sell everything not usable), `Keep everything` (for a cautious player: stash or keep, never sell or drop). Choosing a preset shows the difference to the current list before it applies. An edited preset reads "Starter (edited)" and can be reset.
- **Import and export**: one text file (the engine's own versioned format), Copy and Paste buttons, and a validation result that names the rule and the field on an error. An imported list is previewed like a preset before it applies.
- **Safe defaults.** A new rule starts switched on only after its preview has been seen once. `drop` and `sell` rules that would match a unique, set or rune item show an amber line saying so. One-step Undo of the last save. Rules are per hero.
- **Unlock.** The tab is visible from the start as a read-only view of the Starter preset with the reasons ("this is what the hero does and why"). Editing is the unlock: ladder row `lootrules` (group `ui`), proposed trigger the first time the hero has sold twenty items (`seen`, event `sold_20`), with tally as the alternative; presets other than Starter open with the same row. The trigger and price are the user's call (question 18).

## 5. What the editor needs from the engine

### Conditions (each needs: a stable field name, a type, the list of allowed values for pickers, and a human label)

| Group | Condition | Editor control |
|---|---|---|
| Kind | quality (low, normal, superior, magic, rare, set, unique, crafted), "X or better" | ordered list |
| Kind | item type, with the type hierarchy (weapon > sword; armor > helm; ...) and class-specific types | tree with search |
| Kind | base item (code and name), tier (normal, exceptional, elite) | list with search |
| Kind | named unique or set item | list with search |
| Kind | rune (number or name, "X or higher"), gem (kind and grade), jewel, charm (size), potion (kind and tier), key, scroll, tome, quest item, gold (amount) | lists and ranges |
| State | identified, ethereal, sockets (count, and maximum possible), broken, stackable, quantity already owned (pack plus stash) | yes or no, ranges |
| Numbers | item level, required level, required strength, required dexterity, sell value | ranges |
| Properties | any property the tooltip can show, with a minimum: "+N to life", "N% fire resistance", "+N to a skill tree", "N% magic find"; all of / any of | property picker with a number; needs the property catalogue with labels and units |
| Use | **usable by hero now**, **usable by mercenary now**, **usable by hero after +N strength / +N dexterity / +N levels**, class-restricted to the hero's class | yes or no, a number |
| Context | hero level range, difficulty, act or area, pack fullness percent, stash fullness percent, gold on hand | ranges, lists |

The "Use" group is required by the user's requirement constraint: it must come from the same validity function as the equip path (`usable` in I-character-screen.md section 5), so a rule and the doll never disagree.

### Actions

Ground stage: `pickup`, `leave`. Pack stage: `keep`, `stash`, `sell`, `drop`, `ask`. The engine tells the editor which actions are legal for a stage.

### Functions the UI calls (names are placeholders; the engine agent picks them)

| Need | Why |
|---|---|
| The schema: every condition field with its type, allowed values, label and unit; every action per stage | the pickers are generated from it, so a new condition needs no UI change |
| Validate a rule list, with errors that name the rule index and the field | inline errors in the editor and on import |
| Evaluate one item in a context and return the action, the rule id and a one-line reason | the tooltip line and the pack badge |
| Dry run a (possibly unsaved) rule list over a set of item snapshots and return, per rule, the count and the matched item ids; and the difference to the saved list | the live preview and "with this change" |
| "Never reached" analysis, or enough information to compute it | the warning on shadowed rules |
| The protected set, as data | the grey block |
| Presets as data files in the same format, each with a name and description | the preset menu |
| A versioned text format with forward tolerance (unknown fields kept or reported, never silently dropped) and a migration from the present tables | import, export, old saves |
| Stable rule ids across edits and reorders | preview highlighting, Undo, the tooltip line |
| Per-hero storage and a `seq` for change detection | the page polls nothing new: it reads the rules once and on the change event |

### Data the preview needs that is not the engine's

A **drop log**: the last N items seen on the ground with full snapshots (quality, type, base, level, sockets, properties when known, what was done and by which rule). U whether today's loot memory (`docs/lootmemory.md`) keeps enough; if not, a small ring of snapshots is bot work for the loot-rules slice. Unidentified items can only be judged on what is visible; the preview says "decided before identification" for those.

## 6. Not in this UX

- No rule ever equips, and none compares items for the player.
- No free-text expression language in the editor. If the engine has one, it is reachable through Import only.
- No sound or beam-style drop highlighting on the picture: the picture is the game's.
