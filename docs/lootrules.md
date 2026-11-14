# Loot rules: one engine for every item decision (TASK-421)

> status: design plus a PROTOTYPE (`internal/lootrules/`, stdlib only, tested offline, NOT wired into the bot, no flag). User request 2026-11-14: "a loot filter the bot follows that the player can configure, maybe an unlockable". The player-facing editor is designed in parallel in `docs/ui/redesign-2026-11-14/loot-rules-ux.md` (branch agent/uidesign2); section 8 maps its section 5 needs to this engine. Tags: V verified (how), S secondary (a doc, the SME, memory of public tools), U unverified, EXP veteran-player advice (SME consult 2026-11-14). Nothing here ran against a live game.

## 0. Summary

Today four tables decide about items and they disagree at the edges (docs/sellskip.md section 1): the pickit `internal/loot/pickit_starter.json` with its poor-mode and empty-slot overrides in `evaluate`, the tidy policy `internal/tidy/policy.json`, the sell chore's own `sellable` test, and the shop and identify chores. The engine replaces the decision part of all of them with ONE ordered, versioned, JSON rule set, compiled once per hero, asked at three places: the ground (pickup), the pack (keep, stash, sell, drop, identify, ask) and the stash. The chores keep doing the clicking; they ask the engine what to do.

The shipped `Starter` preset is today's pickit (ground rules) and today's tidy policy (pack rules) written in the new format, so the default behaviour does not change (V for the pack side: `TestStarterPackEqualsTidyPolicy` compares 432 decisions with `tidy.Decide`, three classes, two pack fill levels, two stash fill levels; the ground side is translated by hand and spot-checked, the full table test against the real pickit needs the game tables and is card L2).

## 1. The item fact model

A rule sees an `Item` (what the thing is), a `Ctx` (the situation) and, for requirement rules, two `Actor`s (the hero and the mercenary). Every field has a source and a place where it is unreadable; an unreadable field makes a condition UNKNOWN, never false (section 2.2).

| field | ground | pack | stash | source (V = read in code) |
|---|---|---|---|---|
| code, type and ancestors, name | yes (label parse on d2d) | yes | yes | `d2.Unit.Class` -> `data.ItemBase` (V `internal/loot/loot.go base`); ancestors through `data.DB.IsA` (V `isAny`); d2d ground: label text only (V docs/d2d-loot.md) |
| quality | GUESSED from the label (U for magic, rare, unique, set labels; docs/d2d-loot.md "Quality guess") | yes (`Item.Quality`) | yes | V `d2.Item.Quality` 1..8 |
| ilvl | NO on d2d ground (0); yes on the Wine path | yes (`Item.ILvl`, V +0x2C) | yes | flag `HILvl` marks it unreadable |
| identified, ethereal | identified: unreadable on d2d ground; ethereal likewise | yes (`Item.Flags`, V) | yes | `HIdentified` |
| sockets | no on ground | yes (stat 194 in the pickit, V rules) | yes | |
| stat values (life, fcr, resists, skills ...) | NO | only when identified (an unidentified magic or better item shows none, S: game rule; the engine also treats `!identified && quality >= magic` as hidden) | same | `d2.Stats` via `UnitStats` (V `loot.Looter.stats`); the display name table is `stats.go` (ids S from the pickit, others U) |
| sell value | no | `loot.SellValue` (V, a lower bound for identified magic and better: docs say so in `SellValue`) | same | flag `HValue` on ground |
| required level, strength, dexterity | yes (base row) | yes | yes | V `data.ItemBase.LevelReq/ReqStr/ReqDex` (used by `wearsBetter`) |
| class restriction | yes | yes | yes | V `data.ItemType.Class` (`wearsBetter` compares it with the hero's class code) |
| hireling types that can wear it | yes | yes | yes | U: from Hireling.txt plus the item type tree; `internal/merc` reads the merc; no function exists yet, card L2 builds it |
| durability, max durability, broken | no | yes (stats, U which ids) | yes | adapter fills `Durability`, `MaxDur` |
| rune number, gem grade, base tier, max sockets | yes (from the base row) | yes | yes | data tables (U: derived by the adapter; the rune number is the code order) |
| stackable, quest, start item, qty | yes | yes | yes | V `data.ItemBase.Quest`, `FlagStartItem`; qty: gold amount from the label (V docs/d2d-loot.md) or the stack stat 70 |
| mark (keep, sell) | n/a | yes (`invdesk` marks, V docs/inventory-actions.md) | yes | |
| place | the container: ground, pack, stash; equipment, belt, cube, cursor and vendor are NEVER decided |

Ctx: hero class and build (data keys), difficulty, act, clvl, gold, pack fullness percent, stash fullness percent, `Needs(kind)` (the belt, scroll, tome, key, gold, gem, charm-room, runeword and empty-slot needs the looter has today: `loot.needsOf`), `Owned(item)` (how many of the base the hero holds). Actors: `Hero{Class, Level, Str, Dex}` and `Merc{Type, Level, Str, Dex}`, nil merc = none, level 0 = not read yet.

Unidentified rules, explicit: a condition on `stat:*` of an unidentified magic-or-better item is UNKNOWN. A rule that is UNKNOWN is skipped, and the decision carries `Pending: true` ("a rule above could not be judged; identify and ask again"). A rule may say `"onUnknown": "match"` to take the item anyway (the Starter ground rules "charms to id" and "jewelry to id" do: pick it up, identify later). The `identify` action spends a scroll only where a rule says so (docs/inventory-actions.md: identify stays a player command unless a rule asks; the Starter has no identify rule, matching today).

## 2. The rule language

### 2.1 The file (`lootrules/1`, JSON, `internal/lootrules/types.go`)

```
{ "schema": "lootrules/1", "seq": 12, "name": "Starter", "comment": "...",
  "limits":   { "maxActions": 12, "stashMaxPct": 90 },
  "groups":   { "my_rings": { "types": ["ring"], "codes": ["rin*"] } },        // adds to the built-ins
  "defaults": { "ground": "ignore", "pack": "ignore" },                         // the fixed last row of each stage
  "ack":      [ "drop-valuable" ],                                               // the player's explicit acknowledgements
  "rules": [ { "id": "t-runes", "name": "runes", "scope": ["pack"], "classes": ["Amazon"], "builds": [],
               "when": <condition>, "action": "stash", "prio": 90, "tier": "S",
               "reason": "rune", "confirm": false, "onUnknown": "skip", "pickup": true, "off": false } ] }
```

Six lines of language: (1) an ordered list, first match wins, an explicit default per scope; (2) a rule is `when` (a condition tree) plus an `action` and a `reason` shown in the "why" tooltip; (3) conditions are `all`, `any`, `not` or a leaf `{f, op, v}`; (4) `scope` picks the stage (ground, pack, stash; none = ground and pack) and `classes` and `builds` are the per-hero overlays (data, no class literal in Go); (5) three-valued logic: unreadable fields make a rule UNKNOWN, not false; (6) safety floors that no rule can change, and `confirm` plus the set's `ack` for the one thing a player may unlock (dropping or selling valuables).

Strictness: unknown keys are refused (a typo never silently disables a rule), `id` is unique and stable across edits and reorders, `reason` is required.

### 2.2 Conditions

Leaf fields (the editor reads the list from `lootrules.Fields()`, 58 entries with label, unit, allowed values, where hidden; V `TestSchemaPreviewDiffShadow` checks the schema and the parser agree):
- item: `quality`, `type` (with ancestors), `group` (built-in: runes, gems, charms, jewels, jewelry, circlets, potions, supplies, gold, gear, javelins; own groups in `groups`), `code` (globs), `name` (contains), `ilvl`, `reqlvl`, `reqstr`, `reqdex`, `sockets`, `maxsockets`, `value`, `qty`, `runeno`, `grade`, `basetier`, `owned`, `durpct`, `identified`, `ethereal`, `stackable`, `quest`, `broken`, `stat:<name>` (life, mana, str, dex, vit, enr, fire/cold/light/poisres, fcr, fhr, frw, ias, mf, gf, allskills, classskills, skilltab, ed, and the derived `stat:ressum`).
- requirements (user addition 2026-11-14): `usable_by eq hero|merc` with an optional `plus {lvl, str, dex}` ("usable after +5 strength or +2 levels"), `hero_usable_from` and `merc_usable_from` (the lowest level at which the wearer can wear it: 0 now, the required level when only the level stands in the way, 999 when the class, hireling type, strength or dexterity does), `class_restricted in [classes]`, `reqlvl`, `reqstr`, `reqdex` for plain comparisons. "Unusable now but usable at clvl 30" is `hero_usable_from le 30` after a `usable_by` rule. A merc can wear an item only when its `MercTypes` list names his type (data from Hireling.txt, U); an unread actor (level 0) answers UNKNOWN, so a rule never says "cannot wear it" because a number has not arrived yet.
- context: `class`, `build`, `difficulty`, `act`, `clvl`, `gold`, `packfull`, `stashfull`, `need` (the looter's needs).
- ops: numbers `eq ne ge gt le lt`; bools `eq`; enums `eq ne in`; lists `in`; `name contains`. "Rare or better" is expanded by the editor into `quality in [...]`.

Evaluation is Kleene logic (true, false, unknown): `all` is false if any child is false, else unknown if any is unknown; `any` mirrors it; `not` keeps unknown. V `TestUnidentifiedIsExplicit` (an unidentified ring is identified, not sold; `not(unknown)` stays unknown; ground `onUnknown`).

### 2.3 Actions

`pickup` (ground only), `keep`, `keepmerc` ("keep for the mercenary": a keep the sell chore and tidy leave alone; the player equips him), `stash`, `sell`, `drop`, `identify`, `ask` (raises a Needs-you row and leaves the item until answered), `ignore`. There is NO equip or unequip action: the vocabulary itself guarantees the filter never makes an equipment decision (docs/gear-policy.md; V `TestInvariants`: `"equip"` does not parse). A rule with `keep`, `stash` or `keepmerc` on the ground means "pick it up" (the same rule serves both stages, as the pickit does today); `sell` and `drop` on the ground mean "leave it" unless the rule says `pickup: true` (the poor-mode junk-gear rule does). The decision then carries `Fate` (the rule's own action) and `Tier`, so the strictness knob (tier floor) and the sell-pickup in poor mode keep working in card L2.

### 2.4 Order of evaluation, marks and commands

1. Place: anything but ground, pack, stash is `ignore` (equipment, belt, cube, cursor, vendor).
2. Player marks win: keep -> ignore ("Kept: you marked it keep"), sell -> ignore ("Will sell: you marked it"; the sell chore owns marked sales). A player command (equip, trash, identify, stash move from the inventory menu) is not a decision of the engine; it runs before the rules, through the queue of docs/inventory-actions.md, and the engine never overrides it.
3. Start kit -> keep. Quest item -> pickup on the ground, keep in the pack.
4. The rules of the scope, in order, first match. Then the scope default.
5. Floors on the answer: stash action while the stash is at `stashMaxPct` -> keep; `sell` or `drop` of tomes, scrolls or keys -> keep; `drop` of anything but low, normal or superior quality (or of unreadable quality), and `sell` or `drop` of runes, gems, charms or jewels -> ignore unless the winning rule has `confirm: true` AND the set lists `"drop-valuable"` in `ack`. V `TestInvariants` (Paladin context), `TestStarterPackEqualsTidyPolicy` (the same floors as tidy's `dropSafe`, `neverSell`).
6. Plan: `Compiled.Plan` takes the pack, returns identify, stash (best value first), sell, drop (cheapest first), at most `limits.maxActions` (12 = `town.tidyMaxActions`), ties by item id, so one run is the same as the next (V `TestPlanCapOrderAndDeterminism`).

Determinism: no map iteration and no clock in an evaluation. Performance: `Compile(rs, class, build)` validates once, drops disabled rules and other classes' rules and turns every condition into a closure; a decision is a linear walk without parsing. V `BenchmarkStarterGround`: 40 ground items per iteration in about 65 microseconds (about 1.6 microseconds per item, one allocation per item for the why text), Starter, offline box. The ground scan each tick is therefore negligible; the only per-item cost the card L2 adds is building the `Item` from the unit. A decision cache keyed by (unit id, quality, identified flag, `Ctx.Fingerprint()` when `Compiled.UsesActors`) re-evaluates when the hero levels up or stats change (fingerprint changes: V `TestRequirementsHeroMercAndLater`).

### 2.5 Safety invariants (code, not data; `Invariants()` lists them for the editor's grey block)

place; marks; start kit; quest items; tomes, scrolls and keys never sold or dropped; dropping magic or better items (or anything of unreadable quality), and selling or dropping runes, gems, charms and jewels, needs a confirmed rule and the acknowledgement (selling a magic or rare item stays an ordinary rule, as the tidy `sellable` row is today: V `TestInvariants` shows a unique sells under a plain sell-all rule; the preview warns); stash-full floor; the action cap per pass; no equip action exists. Also in the town code, unchanged: abort on a fight, the 120 s pass cap, two failed drops end the pass, the cursor is never left holding an item (docs/tidy.md).

## 3. What exists today and what it becomes

| today | where | becomes |
|---|---|---|
| pickit rules (quest, runes, gems, uniques, potions "needed", charms with stats, jewelry, runeword bases, junk) | `internal/loot/pickit_starter.json`, `rules.go` | Starter ground rules `g-*` (translated 1:1 by name and order) |
| `needed` (potion, scroll, tome, key, gold, gem, runeword, charmroom) | `loot.needsOf` | `need eq <kind>`; `Ctx.Needs` stays the loot package's function |
| stat conditions `{stat id, min, shift}` | pickit | `stat:<name> ge N` (id table S, section 1) |
| tier S/A/B/C and strictness | pickit | `tier` kept on the rule; strictness is a tier floor applied by L2 |
| `fillsEmptySlot` (empty weapon or body slot) | `loot.EvaluateWith` | rule `g-empty-slot`, `need eq emptyslot` (keeps a wearable weapon or armor; picking up is not equipping) |
| poor mode (gold < 300: junk gear worth >= 20 sells) | `EvaluateWith`, `poorSellable` | rule `g-poor-gear` (`gold lt 300`, `pickup: true`) |
| `WearBetter` | `ctxFrom` | stays OFF: gear is the player's (docs/sellskip.md); not translated, the `usable_by` rules replace its purpose ("keep what I could wear") as data |
| tidy rows, per-class overrides, `dropSafe`, marks, start kit | `internal/tidy` | Starter pack rules `t-*` and `amazon-javelins` (`classes: ["Amazon"]`), floors in the engine |
| the sell chore's `sellable` and `needSell` | `internal/town` | asks `Plan(..., Sell)` (L3), ending the "bot counts N, town finds 0" disagreement of docs/sellskip.md |
| `Valuables` (what to stash before a dangerous trip) | `loot/valuables.go` | `stash` decisions plus the `value` field (L3) |
| the drop-pickup loop guard | `tidy.Drops`, `Looter.SetIgnore` | unchanged; the engine only adds that a rule that drops puts the item on the ignore list (section 7) |

Behaviour changes by default: NONE intended. Known deltas, all to be shown in the L2 table test: (a) a unique or set weapon or armor that fills an empty slot: today `fillsEmptySlot` turns Stash into Keep, the translation answers Stash (U, rare); (b) the engine adds invariants that were only in tidy (equipment never touched) to the ground stage where they cannot matter; (c) `WearBetter` stays off as it already is by default.

## 4. Migration and the Starter preset

Starter = the file `internal/lootrules/starter.json`, generated from the two sources (the 12 tidy rows plus the Amazon override (13 pack rules, tidy names as `name`) and 33 ground rules translating the pickit rows plus its empty-slot and poor-mode overrides) and kept as data. Exact translation of today's tidy rows (V by the table test):

| tidy row | rule `when` | action |
|---|---|---|
| tomes and scrolls | `code in [tbk ibk tsc isc key]` | keep |
| potions | `group potions` | keep |
| runes | `group runes` | stash |
| gems | `group gems` (type prefixes gem, skul) | stash |
| charms, jewels | `group charms`, `group jewels` | stash |
| unique and set | `quality in [unique set]` | stash |
| rare jewelry | `type in [ring amul circ] and quality in [rare crafted]` | stash |
| magic jewelry | `type in [ring amul] and quality in [magic]` | stash |
| socketed bases | `sockets ge 4 and quality in [normal superior]` | stash |
| pure junk | `value lt 20 and packfull ge 85 and quality in [low normal superior]` | drop |
| sellable | `value ge 20` | sell |
| Amazon override | `group javelins`, `classes ["Amazon"]`, placed first | keep |

`limits` carry `tidyMaxActions` (12) and `stashMaxPct` (90). The knobs `town.tidyEnabled`, `town.tidyMaxActions`, `town.tidyMinPackPct` stay as the switch and the cap; `loot.*` knobs that are conditions today (`poorGoldBelow` 300, `poorMinSell` 20, `valuableMinGold`) become the numbers inside Starter and the knobs are read once to fill them when a hero has none (L6 does the lift, so the migration is data, not code). The embedded `policy.json` and `pickit_starter.json` stay until L2 and L3 are merged and the table tests pass; then they become one-line stubs pointing here.

### Preset defaults from the SME (EXP, 2026-11-14)
All EXP. Starter does NOT adopt them (it is today's behaviour); they become the other presets in card L6 and the preview shows the difference first.
- Always pick up in Normal Act 1-2: all potions and rejuvenations, town portal and identify scrolls up to a cap (20 and 10), ALL runes, magic-or-better rings and amulets, rare circlets, uniques and sets of any base, all charms, white 3 to 4 socket-capable runeword bases. Gold floor about 20 x clvl per pile (conflicts with docs/goldprio.md, where the user wants gold prioritised: keep today's bar, offer the SME bar as a preset knob). Ignore: plain white no-socket items, magic items under about 40 gold sale value, javelins and arrows for non-ranged builds, keys (a leveling bot has no use), skulls and low gems beyond a stack cap.
- Runes: keep every rune, stash, never sell; grades El to Ort for the starter runewords, Pul (rune 20) and up are the upgrade path. Gems: chipped and flawed are junk unless a recipe needs them; keep regular and flawless of Ruby, Topaz, Skull for sockets and every perfect gem.
- Charms: stash +20 life, +1 class skill, all resists with +3 each, FCR 20 on a grand charm, MF 7+ on a small charm; early (clvl < 25) only life 15+ or a single resist 7+. Rings and amulets: +skills, FCR 10+ on a ring, life 20+, all resists 10+ (sum), MF 20+. Jewels: 15% IAS or good resist and life. Uniques and sets: stash anything with use; keep the first 3 to 4 set pieces; sell low-value uniques with no function.
- Sell: white gear, magic and rare items under about 400 gold at clvl < 20 (threshold rises with level). Stash: keepers above. Drop: only when the pack is full AND the item is white junk or an unneeded potion, in the order arrows, junk potions, cheapest white, magic under the threshold. Never drop: unique, set, rune, gem, jewel, charm, quest items, runeword bases with 3+ sockets, identified rares, rejuvenation. Sockets: keep 3 to 4 socket armor and shields, 4 socket weapons; ethereal only when the base has 3+ sockets.
- Identify: scroll 100 gold; spend one on rings, amulets, jewels, circlets or a usable base, not on plain magic items; unidentified rares: carry rings, amulets, circlets, boots, gloves, belts up to about 6 slots, sell the rest.
- Traps (become tests and preview warnings): picking up everything; dropping unidentified rares and uniques the filter cannot see into (so: the floor blocks dropping anything magic or better, and `Pending` defers a sale, section 7); selling charms, jewels, runes without a stat check; dropping a low gem a cube recipe needs; fetching gold behind a pack (the fight and danger guards of docs/goldprio.md, unchanged); the pickup-drop loop; selling a base you hoped to socket; "sells for little" is not "useless" (Stone of Jordan sells low, is worth a lot: so a sell rule on `quality unique` is never in Starter).
- Safety defaults the SME asks for, all adopted: caps per pass (here 12 stash plus drop; the SME says 10 pickups, 10 sells, 5 drops: L6 sets per-kind caps in `limits`, U), confirmation to drop or sell valuables with default OFF (the `ack`), a never-drop list (the invariants), one log line per decision with item, quality and rule id (L2, L3), keep 2 to 3 free pack cells (today `loot.freeCellsKeep` 2).

## 5. Unlock design (ladder pattern: `data/unlocks/ladder.json`, `docs/unlock-ladder.md`, `internal/watch/unlockorderqueue.go`)

Pattern: `ui` group rows with a trigger, optional `alt` triggers and a `price` (tally) when `tally_spend` is among them; slot rows (`orderqueue`, `orderqueue2`, `orderqueue3`) open features in steps and a small embedded table (`orderqueue.json`) lists them. The same shape fits here: `data/unlocks/lootrules.json`-style table not needed, three ladder rows suffice and the engine reads `Allowed(id)`.

| row | gives | trigger proposal (the user's call) | alt |
|---|---|---|---|
| (no row) | Starter, fixed, read-only view of the active rules with the reasons ("this is what the hero does and why"), the item tooltip's why line | always | |
| `lootrules_presets` | choose among the shipped presets (Starter, Runes and gems, Sell junk, Keep everything, Veteran), with the difference preview | `item_quality` qualities `[rare]` (first rare item found, early in Act 1, reuses the idreveal trigger kind) | `level` min 12 |
| `lootrules` | the full editor: add, edit, reorder, switch off, import, export, undo, own groups | `quest_stage` `search_for_cain` min 3 (Deckard Cain rescued) | `tally_spend`, price 60 |
| `lootrules_ack` (optional) | the `drop-valuable` acknowledgement and `confirm` on rules that drop or sell valuables | never automatic: a separate toggle in the editor with a typed confirmation, no ladder row | |

The UX page suggests "sold twenty items" (`seen` event `sold_20`): no emitter exists (U), so it is listed as an alternative for a later card, not the trigger. Reveal text for `lootrules_presets`: "Your hero follows a loot list. Pick another one that suits how you play." for `lootrules`: "Write your own loot rules: what to pick up, keep, stash, sell or drop, and why."

What "locked" means: the engine always runs the active set (Starter until the player picks otherwise); the API serves `GET /api/lootrules` in every state so the page can show the active rules read-only with the `Why` of every rule; a PUT or a preset change answers 403 `locked` with the teaser. The player learns the language before owning it. No behaviour is held back by the lock.

## 6. API for the UI (additive, versioned: `watch.lootrules/1`)

Same gate as the inventory routes (origin list, token or passcode cookie, JSON only, small body, unknown fields refused). All per hero.

| route | body | answer |
|---|---|---|
| `GET /api/lootrules` | | `{schema, seq, locked: none\|presets\|editor, active: <RuleSet>, protected: Invariants(), presets: [{id, name, description}], stats: {rules, shadowed:[...]}}` |
| `PUT /api/lootrules` | a RuleSet (seq of the one it was based on) | 200 `{seq}` or 400 `{errors:[{path, msg, line, col}]}` (syntax errors carry line and column, field errors the path `rules[3].when.all[1].v`); 409 `seq` when the stored set changed since; 403 `locked` |
| `GET /api/lootrules/schema` | | `{fields: Fields(), actions: {ground, pack, stash}, groups, classes, builds, qualities, itemTypes}`: the editor builds its pickers from it |
| `POST /api/lootrules/preview` | `{rules: <RuleSet, may be unsaved>, scope: ground\|pack, sample: drops\|pack\|stash\|both}` | `{counts:[{rule, count, ids}], decisions:{id:{action, rule, why}}, changes:[{id, from, to}], shadowed:[{rule, by}], warnings:[...]}`: `Preview` over the sample, `Diff` against the saved set, `Shadowed` for "never reached"; a valuable that a `sell` or `drop` rule would take gets a warning |
| `POST /api/lootrules/preset` | `{id}` | like PUT after the preview diff has been shown |
| `POST /api/lootrules/undo` | | restores the previous saved set (one step, kept in the state file) |
| `GET /api/lootrules/export`, `POST /api/lootrules/import` | the canonical JSON (`Marshal`) | import validates and previews, it does not apply |
| `GET /api/lootrules/why?id=<item>` | | the `Decision` and the `Explain` trace of every rule for one item |

The inventory view (`invdesk.ItemView`) keeps its additive `tidy` field and gains `rule` (id) and `why`; `ViewSchema` bump is the invdesk owner's call (as in docs/tidy.md). Versioning: `schema: "lootrules/1"` is checked on read; a newer major is refused with the version named; unknown keys are refused so a file cannot lose rules silently; a migration table `v1 -> v2` is the place for later changes. Storage: `lootrules-<profile>.json` beside the other state files, atomic write, one previous copy for undo, `seq` incremented on every save (the page reads once and on the change event).

### The recent-drops ring buffer (card L7)
Preview over the ground needs the last N items seen on the ground. `loot.Looter.ObserveGround` already sees each ground unit once per tick (docs/lootmemory.md, "Remember"); L7 adds a ring of 100 entries `{id, area, time, Item facts as filled by the adapter, the decision taken and the rule id, what happened (picked, left, failed)}` next to the memory, in memory only plus a JSON dump on shutdown (so a restart does not empty the preview). Unidentified items keep their `Hidden` mask so the preview can say "decided before identification". About 100 bytes of facts per entry plus name text: under 30 KB.

## 7. Risks and what answers them

| risk | answer |
|---|---|
| A bad rule drops or sells good items | floors in code (section 2.5); `confirm` plus `ack` for dropping magic or better and for selling or dropping runes, gems, charms, jewels; the preview warns before saving; the cap per pass; the town log line per decision with rule id; undo of the last save; marks always win |
| Rule UNKNOWN on an unidentified item, then dropped | an UNKNOWN condition never matches, and dropping magic or better is blocked by the floor whatever the rules say; for sell, the town chore (L3) does not sell a `Pending` item that can still be identified, and the plan lists `identify` first |
| Drop-pickup loop (the ground rules pick up what the pack rules drop) | every executed drop goes to `tidy.Drops` and `Looter.SetIgnore` as today; the preview shows a "ground picks it, pack drops it" warning (a static check: a drop rule whose `when` implies a ground pickup rule) in L4; the loop test of docs/tidy.md stays |
| Walking into danger for a pickup | unchanged: the fight and danger guards of docs/goldprio.md sit above the looter; the engine only answers what is wanted |
| Hero and merc numbers change (level up, stat points) | `Ctx.Fingerprint()` changes, decision caches clear when `UsesActors` |
| The editor produces a rule set the engine rejects | one validator (`Validate`) used by PUT, import, preview and `Compile`; errors carry paths |
| Performance | section 2.4, about 1.6 microseconds per item |
| Two readers disagree (sell chore, tooltip, pickup) | one `Compiled` per hero, built by one owner (the bot wiring in L2) and handed to the looter, tidy, the town sell chore and the inventory view |
| Per-class or per-build data goes missing | an unknown class has no overlay rules and the common rules apply (a table lookup, not a Sorceress fallback; AGENTS.md "Build flexibility"); L6 tests presets with Sorceress and one other class and a merc |
| Gear policy | no equip action; `keepmerc` only keeps; `usable_by` reads, never writes |

## 8. How this meets the UX page (loot-rules-ux.md section 5)

| UX need | here |
|---|---|
| schema with types, values, labels, units | `Fields()`, `ActionsFor()`, `GET /api/lootrules/schema` |
| validate with rule and field | `Validate`, `Error{Path, Msg, Line, Col}` |
| evaluate one item, return action, rule id, reason | `Decide`, `GET /why` |
| dry run over snapshots, per-rule counts and ids, difference to saved | `Preview`, `Diff` |
| never-reached analysis | `Shadowed` (conservative: an unconditional or identical earlier rule) |
| protected set as data | `Invariants()` |
| presets as data in the same format | `starter.json` plus L6 files |
| versioned format, forward tolerance, migration | `schema`, unknown keys refused, migration table; from today's tables via Starter |
| stable rule ids | `Rule.ID` required and unique |
| per-hero storage and `seq` | `RuleSet.Seq`, per-profile file |
| two stages, a fixed last row | `scope` plus `defaults.ground` and `defaults.pack`; the editor writes every rule with ONE scope, an unscoped rule (hand-written or imported) shows as "both stages" |
| "X or better" quality | expanded to `quality in [...]` by the editor |
| ask me | action `ask` (Needs-you row; L4 plumbs the question) |
| usable by hero or merc, after +N | `usable_by` with `plus`, `hero_usable_from`, `merc_usable_from`, `class_restricted` |
| quantity already owned | `owned` (`Ctx.Owned`) |
| drop log | L7 |
| ladder row `lootrules` | section 5; the sold-20 trigger needs an emitter (U) |

## 9. Wiring plan (cards, none started; sizes in agent hours, S 1 to 2, M 3 to 5, L 6 to 10)

| card | what | depends on | size |
|---|---|---|---|
| L1 | engine and tests: THIS branch (`internal/lootrules`, Starter, table test, benchmark). To harden: fuzz the parser, golden Starter file | none | done, S to harden |
| L2 | pickup wiring: `internal/loot` builds `Item`/`Ctx` from `d2.Unit` and the d2d ground label (`Enrich`), `Looter` asks the engine behind the knob `loot.rulesEngine` (default 0, then 1 after the table test passes against the real pickit with game tables: `D2_MPQ_DIR=$HOME/d2d-lab/data`, two classes); hireling-type data; `Needs` and strictness tier floor; poor mode and empty slot as rules | L1 | L (about 8) |
| L3 | tidy, sell, stash chores ask `Plan`: `internal/tidy` becomes a thin adapter over the engine, `town` sell chore uses `Plan(Sell)`, `Valuables`, `identify` plan step, `keepmerc` and `ask` plumbing; delete `needSell` disagreement | L1, L2 | L (about 8) |
| L4 | API: store, routes of section 6, locked states, undo, preview with sample sources, loop-warning check | L1, L7 (for ground preview) | M |
| L5 | UI: the Loot rules tab (Fable's slice, loot-rules-ux.md) | L4 | L, theirs |
| L6 | presets and SME defaults as data files (Runes and gems, Sell junk, Keep everything, Veteran), the three ladder rows, knob lift into Starter numbers, per-kind caps in `limits` | L1 (L3 for live check) | M |
| L7 | recent-drops ring buffer in `loot.Looter` (section 6), JSON dump, the sample source for preview | none (parallel with L1) | S |

Order: L1 -> (L2 and L7 in parallel) -> L3 -> L4 -> L6 -> L5. The first live check is L2's: a run with the engine on and the Starter, comparing `loot:` pickups and `tidy:` plans with the old paths over the same hour (zero differences expected except delta (a) of section 3).

## 10. The prototype (what is in `internal/lootrules/`)

`types.go` (rule set, rule, condition, item, ctx, decision), `parse.go` (strict parse, `Validate` with paths, line and column for syntax errors), `compile.go` (three-valued closures), `requirements.go` (usable_by, usable_from, fingerprint), `eval.go` (`Compile`, `Decide`, floors, `Plan`, `Explain`), `preview.go` (`Preview`, `Diff`, `Shadowed`), `schema.go` (`Fields`, `ActionsFor`, `Invariants`), `groups.go`, `stats.go`, `starter.json`/`starter.go`. Stdlib only; the only non-test import of another package is none (the test imports `internal/tidy`). Tests V (offline, `go test -p 1 ./internal/lootrules`): parse errors by path and line, round trip, first match with all/any/not, unidentified and `onUnknown`, ground fate to pickup, invariants (equipment, marks, start kit, quest, tome, rune, unique, drop gate with and without `confirm` and `ack`, no equip), plan cap order and determinism, class and build overlay (Amazon and others, unknown class), requirements (Paladin and Sorceress heroes, a Desert Guard and a Rogue Scout merc, level 22 vs 24, class orb, unread actor stays pending, fingerprint), durability, schema/preview/diff/shadow, ask, owned, rune number, the Starter against tidy (432 decisions), Starter ground spot checks. Benchmark above. Not done (U): the ground half against the real pickit with game tables (L2), the id-to-name stat table against live stat ids, `MercTypes` data, the API.

## Proposed text for the guarded files (not written, per the guard rule)
- `GLOSSARY.md`: "Loot rule set: the ordered, versioned list of rules (docs/lootrules.md) that decides pickup, keep, stash, sell, drop, identify for every item; the Starter preset is today's behaviour."
- `docs/decisions.md` (index line, for the coordinator or the user): "Item decisions move to one rules engine (`internal/lootrules`); the bot never gains an equip action; dropping or selling valuables needs a confirmed rule plus an acknowledgement (2026-11-14, proposal, not a decision yet)."
